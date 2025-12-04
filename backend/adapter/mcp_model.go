package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ServerManager 管理所有 MCP 服务器的生命周期
type ServerManager struct {
	sseServers sync.Map           // path -> *Server
	ctx        context.Context    // 控制 goroutine 生命周期
	cancel     context.CancelFunc // 取消函数
	wg         sync.WaitGroup     // 等待 goroutine 完成
	mu         sync.RWMutex       // 保护并发操作
	maxEventID int64              // 记录已处理的最大事件ID
	handles    []RequestHandle    // 处理器列表
}

var serverManager *ServerManager
var initOnce sync.Once

type EventCode int

const (
	AddToolEvent           EventCode = iota // 工具添加事件
	RemoveToolEvent                         // 工具移除事件
	AddApplicationEvent                     // 应用添加事件
	RemoveApplicationEvent                  // 应用移除事件
	ToolListChanged                         // 工具列表变更事件
)

type Server struct {
	protocol   string
	path       string
	server     *server.MCPServer
	impl       http.Handler
	cleanupFns []func() // 清理函数列表
	mu         sync.Mutex
}

type Event struct {
	Interface *models.Interface
	App       *models.Application
	Code      EventCode
}

type Parameters struct {
	HeaderParams map[string]any
	QueryParams  map[string]any
	PathParams   map[string]any
	BodyParams   map[string]any
}

type RequestMeta struct {
	URL      string
	Method   string
	AuthType string
	Protocol string
	Ext      map[string]string
}

type PostProcessMeta struct {
	TruncateFields   map[string]int `json:"truncate_fields"`
	StructuredOutput bool           `json:"structured_output"`
}

// AddCleanup 添加清理函数
func (s *Server) AddCleanup(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupFns = append(s.cleanupFns, fn)
}

// Cleanup 执行所有清理操作
func (s *Server) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Cleaning up server: %s", s.path)

	// 逆序执行清理函数
	for i := len(s.cleanupFns) - 1; i >= 0; i-- {
		if s.cleanupFns[i] != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Cleanup function panic: %v", r)
					}
				}()
				s.cleanupFns[i]()
			}()
		}
	}

	s.cleanupFns = nil
	log.Printf("Server cleanup completed: %s", s.path)
}

// SendEvent 发送事件到处理队列
func SendEvent(evt Event) {
	if serverManager == nil {
		log.Printf("Warning: ServerManager not initialized, event dropped: %v", evt.Code)
		return
	}

	// 将事件写入数据库
	db := database.GetDB()
	eventLog := models.EventLog{
		EventCode: int(evt.Code),
	}

	// 序列化 Interface 为 JSON
	if evt.Interface != nil {
		interfaceJSON, err := json.Marshal(evt.Interface)
		if err != nil {
			log.Printf("Error marshaling interface: %v", err)
			return
		}
		interfaceStr := string(interfaceJSON)
		eventLog.InterfaceData = &interfaceStr
	}

	// 序列化 Application 为 JSON
	if evt.App != nil {
		appJSON, err := json.Marshal(evt.App)
		if err != nil {
			log.Printf("Error marshaling application: %v", err)
			return
		}
		appStr := string(appJSON)
		eventLog.ApplicationData = &appStr
	}

	if err := db.Create(&eventLog).Error; err != nil {
		log.Printf("Error saving event to database: %v", err)
		return
	}

	log.Printf("Event saved to database: ID=%d, Code=%v", eventLog.ID, evt.Code)
}

// InitServer 初始化服务器管理器
func InitServer() {
	initOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		serverManager = &ServerManager{
			ctx:     ctx,
			cancel:  cancel,
			handles: make([]RequestHandle, 0),
		}

		// 添加处理器
		serverManager.handles = append(serverManager.handles, HTTPSimpleAdapter{})
		serverManager.handles = append(serverManager.handles, HTTPCAPIAdapter{})

		// 加载现有应用
		serverManager.loadExistingApplications()

		// 初始化 maxEventID：从数据库获取当前最大事件ID
		db := database.GetDB()
		var maxID int64
		if err := db.Model(&models.EventLog{}).Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error; err != nil {
			log.Printf("Warning: Failed to get max event ID, starting from 0: %v", err)
			maxID = 0
		}
		serverManager.maxEventID = maxID
		log.Printf("Initialized maxEventID: %d (will only process events with ID > %d)", maxID, maxID)

		// 启动事件处理循环
		serverManager.wg.Add(1)
		go serverManager.eventLoop()

		log.Println("ServerManager initialized successfully")
	})
}

// eventLoop 事件处理循环
func (sm *ServerManager) eventLoop() {
	defer sm.wg.Done()
	log.Printf("Event loop started, maxEventID initialized to: %d", sm.maxEventID)

	// 定时轮询间隔
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			log.Println("Event loop shutting down...")
			return

		case <-ticker.C:
			// 从数据库轮询未处理的事件
			sm.pollAndProcessEvents()
		}
	}
}

// pollAndProcessEvents 轮询并处理事件
func (sm *ServerManager) pollAndProcessEvents() {
	db := database.GetDB()

	// 查询 ID > maxEventID 且未处理的事件，按 ID 升序排列
	var eventLogs []models.EventLog
	if err := db.Where("id > ?", sm.maxEventID).
		Order("id ASC").
		Find(&eventLogs).Error; err != nil {
		log.Printf("Error polling events: %v", err)
		return
	}

	// 没有新事件
	if len(eventLogs) == 0 {
		return
	}

	log.Printf("Found %d new events to process", len(eventLogs))

	// 处理每个事件
	for i := range eventLogs {
		eventLog := &eventLogs[i]

		// 构造 Event 对象
		evt := Event{
			Code: EventCode(eventLog.EventCode),
		}

		// 反序列化 Interface JSON
		if eventLog.InterfaceData != nil && *eventLog.InterfaceData != "" {
			var iface models.Interface
			if err := json.Unmarshal([]byte(*eventLog.InterfaceData), &iface); err != nil {
				log.Printf("Error unmarshaling interface for event %d: %v", eventLog.ID, err)
			} else {
				evt.Interface = &iface
			}
		}

		// 反序列化 Application JSON
		if eventLog.ApplicationData != nil && *eventLog.ApplicationData != "" {
			var app models.Application
			if err := json.Unmarshal([]byte(*eventLog.ApplicationData), &app); err != nil {
				log.Printf("Error unmarshaling application for event %d: %v", eventLog.ID, err)
			} else {
				evt.App = &app
			}
		}

		// 处理事件
		sm.handleEvent(evt)

		// 更新 maxEventID
		if eventLog.ID > sm.maxEventID {
			sm.maxEventID = eventLog.ID
		}
	}
}

// handleEvent 处理单个事件
func (sm *ServerManager) handleEvent(evt Event) {
	var err error

	switch evt.Code {
	case AddToolEvent:
		err = sm.addTool(evt.Interface, evt.App)
	case RemoveToolEvent:
		err = sm.removeTool(evt.Interface, evt.App)
	case ToolListChanged:
		err = sm.toolChanged(evt.Interface, evt.App)
	case AddApplicationEvent:
		err = sm.addApplication(evt.App)
	case RemoveApplicationEvent:
		err = sm.removeApplication(evt.App)
	default:
		log.Printf("Unknown event code: %v", evt.Code)
		return
	}

	if err != nil {
		log.Printf("Error handling event %v: %v", evt.Code, err)
	}
}

// loadExistingApplications 加载现有应用
func (sm *ServerManager) loadExistingApplications() {
	db := database.GetDB()
	var apps []models.Application

	if err := db.Find(&apps).Error; err != nil {
		log.Printf("Error loading applications: %v", err)
		return
	}

	for i := range apps {
		if err := sm.addApplication(&apps[i]); err != nil {
			log.Printf("Error adding application %s: %v", apps[i].Name, err)
		}
	}

	log.Printf("Loaded %d applications", len(apps))
}

// Shutdown 优雅关闭服务器管理器
func Shutdown() {
	if serverManager == nil {
		return
	}

	log.Println("Shutting down ServerManager...")

	// 取消 context，停止事件循环
	serverManager.cancel()

	// 等待事件循环结束
	serverManager.wg.Wait()

	// 清理所有服务器
	serverManager.cleanupAllServers()

	log.Println("ServerManager shutdown completed")
}

// cleanupAllServers 清理所有服务器资源
func (sm *ServerManager) cleanupAllServers() {
	log.Println("Cleaning up all servers...")

	count := 0
	sm.sseServers.Range(func(key, value interface{}) bool {
		if srv, ok := value.(*Server); ok {
			srv.Cleanup()
			count++
		}
		sm.sseServers.Delete(key)
		return true
	})

	log.Printf("Cleaned up %d servers", count)
}

// GetServerImpl 获取服务器实现
func GetServerImpl(path, protocol string) http.Handler {
	if serverManager == nil {
		return nil
	}

	if s, ok := serverManager.sseServers.Load(path); ok && protocol == s.(*Server).protocol {
		return s.(*Server).impl
	}
	return nil
}

// addTool 添加工具到指定应用
func (sm *ServerManager) addTool(iface *models.Interface, app *models.Application) error {
	if s, ok := sm.sseServers.Load(app.Path); ok {
		srv := s.(*Server)
		tool := srv.server.GetTool(iface.Name)
		if tool != nil {
			return fmt.Errorf("tool %s in %s already exists, skipped", iface.Name, app.Name)
		}

		// 从数据库获取接口参数
		db := database.GetDB()
		var params []models.InterfaceParameter
		if db.Where("interface_id = ? and `group` <> 'output'", iface.ID).Find(&params).Error != nil {
			return fmt.Errorf("error getting interface input parameters for tool %s", iface.Name)
		}

		var outputs []models.InterfaceParameter
		if db.Where("interface_id = ? and `group` = 'output'", iface.ID).Find(&outputs).Error != nil {
			return fmt.Errorf("error getting interface output parameters for tool %s", iface.Name)
		}

		schema, err := BuildMcpInputSchemaByInterface(iface.ID)
		if err != nil {
			return err
		}

		marshal, err := json.Marshal(schema)
		if err != nil {
			return err
		}

		log.Printf("Input schema for tool %s: %s", iface.Name, string(marshal))
		newTool := mcp.NewToolWithRawSchema(iface.Name, iface.Description, marshal)

		postProcessMeta := PostProcessMeta{
			TruncateFields:   make(map[string]int),
			StructuredOutput: false,
		}
		if iface.PostProcess != "" {
			if err := json.Unmarshal([]byte(iface.PostProcess), &postProcessMeta); err != nil {
				log.Printf("Error unmarshalling post process meta: %v, tool id %d", err, iface.ID)
			}
			log.Printf("Post process meta for tool %s: %+v", iface.Name, postProcessMeta)
		}

		var outputSchema map[string]any
		if postProcessMeta.StructuredOutput {
			if len(outputs) > 0 {
				outputSchema, err = BuildMcpOutputSchemaByInterface(iface.ID)
				if err != nil {
					return err
				}
				marshal, err = json.Marshal(outputSchema)
				if err != nil {
					return err
				}
				newTool.RawOutputSchema = marshal
				log.Printf("Output schema for tool %s: %s", iface.Name, string(marshal))
			} else {
				log.Printf("Disabling structured output for tool %s due to no output parameters defined", iface.Name)
				postProcessMeta.StructuredOutput = false
			}
		}
		// 创建参数副本，缓存参数信息避免在调用时查库
		paramsCopy := make([]models.InterfaceParameter, len(params))
		copy(paramsCopy, params)
		// 创建 outputSchema 的副本
		outputSchemaCopy := outputSchema
		inputSchemaCopy := schema
		meta := RequestMeta{
			URL:      iface.URL,
			Method:   iface.Method,
			AuthType: iface.AuthType,
			Protocol: iface.Protocol,
			Ext:      make(map[string]string),
		}
		srv.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

			if ok := SatisfySchema(inputSchemaCopy, req.GetArguments()); !ok {
				return mcp.NewToolResultError(fmt.Sprintf("invalid input schema: %v", err)), nil
			}

			finalParams, err := rearrangeParametersAndValidate(req.GetArguments(), paramsCopy)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			var handle RequestHandle = nil
			for _, h := range sm.handles {
				if h.Compatible(meta) {
					handle = h
					break
				}
			}
			if handle == nil {
				return mcp.NewToolResultError(fmt.Sprintf("no compatible handle found for tool %s", iface.Name)), nil
			}
			data, err := handle.DoRequest(ctx, req, *finalParams, meta)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// 后处理：截取字段
			for key, length := range postProcessMeta.TruncateFields {
				bytes, err := truncate(key, length, data)
				if err != nil {
					log.Printf("Error truncating field %s for tool %s: %v", key, iface.Name, err)
				} else {
					data = bytes
				}
			}

			if postProcessMeta.StructuredOutput {
				// 解析 JSON 数据
				var result any
				if err := json.Unmarshal(data, &result); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to parse JSON: %v", err)), nil
				}
				// 使用 SatisfySchema 进行验证
				if !SatisfySchema(outputSchemaCopy, result) {
					return mcp.NewToolResultError("output does not satisfy schema"), nil
				}
				// 过滤数据以匹配 schema
				filtered := FilterDataBySchema(outputSchemaCopy, result)
				bytes, err := json.Marshal(filtered)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to marshal filtered output: %v", err)), nil
				}
				return mcp.NewToolResultStructured(filtered, string(bytes)), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		})

		log.Printf("Added tool: %s", iface.Name)
		return nil
	}

	return fmt.Errorf("application %s not found for tool %s", app.Name, iface.Name)
}

func truncateByPath(data any, path []string, length int) any {
	if len(path) == 0 {
		return data
	}
	// 处理 map 类型
	if m, ok := data.(map[string]any); ok {
		result := make(map[string]any)
		if path[0] == "*" {
			// 通配符：处理所有 key
			for k, v := range m {
				if len(path) == 1 {
					// 最后一层，截断字符串
					if str, ok := v.(string); ok && len(str) > length {
						result[k] = str[:length]
					} else {
						result[k] = v
					}
				} else {
					// 继续递归
					result[k] = truncateByPath(v, path[1:], length)
				}
			}
		} else {
			// 指定 key
			targetKey := path[0]
			for k, v := range m {
				if k == targetKey {
					if len(path) == 1 {
						// 最后一层，截断字符串
						if str, ok := v.(string); ok && len(str) > length {
							result[k] = str[:length]
						} else {
							result[k] = v
						}
					} else {
						// 继续递归
						result[k] = truncateByPath(v, path[1:], length)
					}
				} else {
					result[k] = v
				}
			}
		}
		return result
	}

	// 处理 slice 类型
	if arr, ok := data.([]any); ok {
		result := make([]any, len(arr))
		if path[0] == "*" {
			for i, v := range arr {
				if len(path) == 1 {
					if str, ok := v.(string); ok && len(str) > length {
						result[i] = str[:length]
					} else {
						result[i] = v
					}
				} else {
					result[i] = truncateByPath(v, path[1:], length)
				}
			}
		} else {
			index, err := strconv.Atoi(path[0])
			if err == nil && index >= 0 && index < len(arr) {
				for i, v := range arr {
					if i == index {
						if len(path) == 1 {
							if str, ok := v.(string); ok && len(str) > length {
								result[i] = str[:length]
							} else {
								result[i] = v
							}
						} else {
							result[i] = truncateByPath(v, path[1:], length)
						}
					} else {
						result[i] = v
					}
				}
			} else {
				copy(result, arr)
			}
		}
		return result
	}
	// 其他类型直接返回
	return data
}

func truncate(key string, length int, data []byte) ([]byte, error) {
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return data, fmt.Errorf("failed to parse JSON: %v", err)
	}
	keys := strings.Split(key, ".")
	result = truncateByPath(result, keys, length)
	newData, err := json.Marshal(result)
	if err != nil {
		return data, fmt.Errorf("failed to marshal JSON: %v", err)
	}
	return newData, nil
}

// rearrangeParametersAndValidate 应用默认值并验证参数
func rearrangeParametersAndValidate(rawParams map[string]any, params []models.InterfaceParameter) (*Parameters, error) {

	headerParams := make(map[string]any)
	bodyParams := make(map[string]any)
	queryParams := make(map[string]any)
	pathParams := make(map[string]any)
	setVal := func(p models.InterfaceParameter, val any) {
		switch strings.ToLower(p.Location) {
		case "query":
			queryParams[p.Name] = val
		case "header":
			headerParams[p.Name] = val
		case "path":
			pathParams[p.Name] = val
		default: // body
			bodyParams[p.Name] = val
		}
	}

	// 先处理input组的参数
	for _, p := range params {
		// 只处理 input 组的参数
		if p.Group != "input" {
			continue
		}
		param, provided := rawParams[p.Name]
		// 如果参数已提供，应用提供的值
		if provided {
			setVal(p, param)
			continue
		}
		// 如果参数未提供且有默认值，应用默认值
		if p.DefaultValue != nil && *p.DefaultValue != "" {
			// 根据参数类型转换默认值
			convertedVal, err := ConvertDefaultValue(*p.DefaultValue, p.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to convert default value for parameter %s: %w", p.Name, err)
			}
			setVal(p, convertedVal)
			log.Printf("Applied default value for input parameter %s: %v", p.Name, convertedVal)
			continue
		}
		// 参数未提供且没有默认值，如果参数是必需的，返回错误
		if p.Required {
			return nil, fmt.Errorf("missing required parameter: %s", p.Name)
		}
	}
	// fixed 组的参数最终覆盖上去
	for _, p := range params {
		if p.Group != "fixed" {
			continue
		}
		if p.DefaultValue == nil || *p.DefaultValue == "" {
			log.Printf("Warning: fixed parameter %s has no default value", p.Name)
			continue
		}
		convertedVal, err := ConvertDefaultValue(*p.DefaultValue, p.Type)
		if err != nil {
			log.Printf("Warning: failed to convert fixed parameter %s: %v", p.Name, err)
			continue
		}
		setVal(p, convertedVal)
	}

	return &Parameters{
		HeaderParams: headerParams,
		QueryParams:  queryParams,
		PathParams:   pathParams,
		BodyParams:   bodyParams,
	}, nil
}

// ConvertDefaultValue 转换逻辑需要保持和MCP接口的一致, 这个接口暴露出去
func ConvertDefaultValue(defaultValue string, paramType string) (any, error) {
	switch paramType {
	case "number":
		// 尝试转换为 float64
		if val, err := strconv.ParseFloat(defaultValue, 64); err == nil {
			return val, nil
		}
		// 如果失败，尝试转换为 int
		if val, err := strconv.ParseInt(defaultValue, 10, 64); err == nil {
			return val, nil
		}
		return nil, fmt.Errorf("invalid number format: %s", defaultValue)
	case "boolean":
		val, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean format: %s", defaultValue)
		}
		return val, nil
	case "string":
		return defaultValue, nil
	default:
		// 自定义类型不应该有默认值，但如果有就返回字符串
		return defaultValue, nil
	}
}

// removeTool 从指定应用移除工具
func (sm *ServerManager) removeTool(iface *models.Interface, app *models.Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}
	if iface == nil {
		return fmt.Errorf("interface is nil")
	}

	if s, ok := sm.sseServers.Load(app.Path); ok {
		s.(*Server).server.DeleteTools(iface.Name)
		log.Printf("Removed tool: %s", iface.Name)
	}
	return nil
}

// toolChanged 通知工具列表变更
func (sm *ServerManager) toolChanged(_ *models.Interface, app *models.Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	if s, ok := sm.sseServers.Load(app.Path); ok {
		s.(*Server).server.SendNotificationToAllClients(mcp.MethodNotificationToolsListChanged, nil)
		log.Printf("Tool changed notification sent: %s", app.Name)
	}
	return nil
}

// addApplication 添加应用
func (sm *ServerManager) addApplication(app *models.Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}
	if app.Protocol != "sse" && app.Protocol != "streamable" {
		return fmt.Errorf("unsupported protocol: %s", app.Protocol)
	}

	// 检查是否已存在
	if _, exists := sm.sseServers.Load(app.Path); exists {
		log.Printf("Application %s already exists, skipping", app.Name)
		return nil
	}
	var interfaces []models.Interface
	db := database.GetDB()
	query := db.Where("app_id = ?", app.ID)
	if err := query.Find(&interfaces).Error; err != nil {
		return fmt.Errorf("error getting interfaces: %v", err)
	}
	mcpServer := server.NewMCPServer(app.Name, "1.0.0")
	var srv *Server = nil
	if app.Protocol == "sse" {
		srv = &Server{
			protocol: app.Protocol,
			path:     app.Path,
			server:   mcpServer,
			impl: server.NewSSEServer(
				mcpServer,
				server.WithSSEEndpoint(fmt.Sprintf("/sse/%s", app.Path)),
				server.WithMessageEndpoint(fmt.Sprintf("/message/%s", app.Path)),
			),
			cleanupFns: make([]func(), 0),
		}
	} else {
		srv = &Server{
			protocol: app.Protocol,
			path:     app.Path,
			server:   mcpServer,
			impl: server.NewStreamableHTTPServer(
				mcpServer,
				server.WithEndpointPath(fmt.Sprintf("/streamable/%s", app.Path)),
				server.WithStateLess(true),
			),
			cleanupFns: make([]func(), 0),
		}
	}
	// 添加清理函数：清理所有工具
	srv.AddCleanup(func() {
		log.Printf("Cleaning up tools for application: %s", app.Name)
		// 如果 MCPServer 有 Close/Shutdown 方法，在此调用
		// mcpServer.Close()
	})
	// 存储服务器
	sm.sseServers.Store(app.Path, srv)
	// 添加所有接口作为工具
	for i := range interfaces {
		if err := sm.addTool(&interfaces[i], app); err != nil {
			log.Printf("Error adding tool %s: %v", interfaces[i].Name, err)
			continue
		}
	}
	log.Printf("Added MCP server: %s, protocol: %s, tools: %d", app.Name, app.Protocol, len(interfaces))
	return nil
}

// removeApplication 移除应用并清理资源
func (sm *ServerManager) removeApplication(app *models.Application) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}
	if s, ok := sm.sseServers.Load(app.Path); ok {
		srv := s.(*Server)
		// 执行清理
		srv.Cleanup()
		// 从 map 中删除
		sm.sseServers.Delete(app.Path)
		log.Printf("Removed application and cleaned up resources: %s", app.Name)
	} else {
		log.Printf("Application not found for removal: %s", app.Name)
	}
	return nil
}
