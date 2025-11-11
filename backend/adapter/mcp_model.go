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
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ServerManager 管理所有 MCP 服务器的生命周期
type ServerManager struct {
	sseServers sync.Map           // path -> *Server
	eventChan  chan Event         // 事件通道
	ctx        context.Context    // 控制 goroutine 生命周期
	cancel     context.CancelFunc // 取消函数
	wg         sync.WaitGroup     // 等待 goroutine 完成
	mu         sync.RWMutex       // 保护并发操作
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

	select {
	case serverManager.eventChan <- evt:
		log.Printf("Event sent: %v", evt.Code)
	case <-serverManager.ctx.Done():
		log.Printf("Warning: ServerManager shutting down, event dropped: %v", evt.Code)
	default:
		log.Printf("Warning: Event channel full, dropping event: %v", evt.Code)
	}
}

// InitServer 初始化服务器管理器
func InitServer() {
	initOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		serverManager = &ServerManager{
			eventChan: make(chan Event, 100), // 增加缓冲区大小
			ctx:       ctx,
			cancel:    cancel,
		}

		// 启动事件处理循环
		serverManager.wg.Add(1)
		go serverManager.eventLoop()

		// 加载现有应用
		serverManager.loadExistingApplications()

		log.Println("ServerManager initialized successfully")
	})
}

// eventLoop 事件处理循环
func (sm *ServerManager) eventLoop() {
	defer sm.wg.Done()
	log.Println("Event loop started")

	for {
		select {
		case <-sm.ctx.Done():
			log.Println("Event loop shutting down...")
			return

		case evt := <-sm.eventChan:
			sm.handleEvent(evt)
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

		var outputSchema map[string]any
		shouldStructuredOutput := false
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
			shouldStructuredOutput = true
		}

		// 创建工具的副本以避免闭包捕获
		ifaceCopy := *iface
		// 创建参数副本，缓存参数信息避免在调用时查库
		paramsCopy := make([]models.InterfaceParameter, len(params))
		copy(paramsCopy, params)
		// 创建 outputSchema 的副本
		outputSchemaCopy := outputSchema

		srv.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			// 应用默认值并验证参数
			processedArgs, err := applyDefaultsAndValidate(args, paramsCopy)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, code, err := CallHTTPInterfaceWithParams(ctx, &ifaceCopy, processedArgs, paramsCopy)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if code != http.StatusOK {
				log.Printf("Error calling tool %s, code: %d", ifaceCopy.Name, code)
			}
			if shouldStructuredOutput {
				out, text, err := parseStructuredOutput(data, outputSchemaCopy)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to parse structured output: %v", err)), nil
				}
				filtered := filterOutputBySchema(out, outputSchemaCopy)

				return mcp.NewToolResultStructured(filtered, text), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		})

		log.Printf("Added tool: %s", iface.Name)
		return nil
	}

	return fmt.Errorf("application %s not found for tool %s", app.Name, iface.Name)
}

// parseStructuredOutput 解析结构化输出，支持对象、数组和双重编码的 JSON 字符串
func parseStructuredOutput(data []byte, outputSchema map[string]any) (map[string]any, string, error) {
	// 首先尝试解析为通用类型
	var genericData any
	if err := json.Unmarshal(data, &genericData); err != nil {
		return nil, "", fmt.Errorf("invalid JSON: %v", err)
	}

	// 检查解析结果的类型
	switch v := genericData.(type) {
	case map[string]any:
		// 直接是对象，返回
		return v, string(data), nil

	case []any:
		// 是数组，需要包装成对象
		// 检查 schema 的顶层 type，如果是 array，则包装；否则报错
		schemaType, _ := outputSchema["type"].(string)
		if schemaType == "array" {
			// schema 定义的就是数组类型，包装为 { "items": [...] }
			wrapped := map[string]any{"items": v}
			return wrapped, string(data), nil
		}
		// 如果 schema 不是 array，尝试查找第一个是 array 类型的属性
		if properties, ok := outputSchema["properties"].(map[string]any); ok {
			for key, propSchema := range properties {
				if propSchemaMap, ok := propSchema.(map[string]any); ok {
					if propType, _ := propSchemaMap["type"].(string); propType == "array" {
						// 找到了数组类型的属性，使用该属性名包装
						wrapped := map[string]any{key: v}
						return wrapped, string(data), nil
					}
				}
			}
		}
		// 都不是，使用默认的 "items" 包装
		wrapped := map[string]any{"items": v}
		return wrapped, string(data), nil

	case string:
		// 可能是双重编码的 JSON 字符串
		var out map[string]any
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			// 尝试解析为数组
			var arrData []any
			if err2 := json.Unmarshal([]byte(v), &arrData); err2 == nil {
				wrapped := map[string]any{"items": arrData}
				return wrapped, v, nil
			}
			return nil, "", fmt.Errorf("failed to parse inner JSON string: %v", err)
		}
		return out, v, nil

	default:
		return nil, "", fmt.Errorf("unsupported JSON type: %T", genericData)
	}
}

// filterOutputBySchema 根据 outputSchema 过滤输出，只保留 schema 中定义的字段（递归处理）
func filterOutputBySchema(out map[string]any, outputSchema map[string]any) map[string]any {
	if outputSchema == nil {
		return out
	}

	// 获取 schema 中的 properties
	properties, ok := outputSchema["properties"].(map[string]any)
	if !ok || properties == nil {
		return out
	}

	// 创建过滤后的结果
	filtered := make(map[string]any)
	for key, propSchema := range properties {
		if value, exists := out[key]; exists {
			// 递归处理复杂类型
			filtered[key] = filterValueBySchema(value, propSchema)
		}
	}

	return filtered
}

// filterValueBySchema 根据 schema 过滤单个值（递归处理对象和数组）
func filterValueBySchema(value any, schema any) any {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return value
	}

	// 检查 schema 是否有嵌套的 properties（不依赖 type 字段）
	if properties, ok := schemaMap["properties"].(map[string]any); ok {
		// 如果 properties 本身又有 properties，说明有多层嵌套
		if _, hasNested := properties["properties"].(map[string]any); hasNested {
			// 递归处理，使用内层的 properties 作为真正的 schema
			return filterValueBySchema(value, properties)
		}

		// 正常的对象类型，使用 properties 过滤
		if valueMap, ok := value.(map[string]any); ok {
			filtered := make(map[string]any)
			for key, propSchema := range properties {
				if v, exists := valueMap[key]; exists {
					filtered[key] = filterValueBySchema(v, propSchema)
				}
			}
			return filtered
		}
		return value
	}

	schemaType, _ := schemaMap["type"].(string)
	switch schemaType {
	case "array":
		// 处理数组
		if valueSlice, ok := value.([]any); ok {
			items, hasItems := schemaMap["items"]
			if hasItems {
				filtered := make([]any, len(valueSlice))
				for i, item := range valueSlice {
					filtered[i] = filterValueBySchema(item, items)
				}
				return filtered
			}
		}
		return value

	default:
		// 基础类型或无法识别的类型直接返回
		return value
	}
}

// applyDefaultsAndValidate 应用默认值并验证参数
func applyDefaultsAndValidate(args map[string]any, params []models.InterfaceParameter) (map[string]any, error) {
	processedArgs := make(map[string]any)

	// 首先复制所有提供的参数
	for k, v := range args {
		processedArgs[k] = v
	}

	// 构建输入参数索引（只处理 input 组的参数）
	inputParams := make(map[string]models.InterfaceParameter)
	for _, p := range params {
		if p.Group == "input" {
			inputParams[p.Name] = p
		}
	}

	// 应用默认值并验证（只对 input 参数）
	for _, p := range params {
		// 只处理 input 组的参数
		if p.Group != "input" {
			continue
		}

		_, provided := processedArgs[p.Name]
		// 如果参数未提供且有默认值，应用默认值
		if !provided && p.DefaultValue != nil && *p.DefaultValue != "" {
			// 根据参数类型转换默认值
			convertedVal, err := convertDefaultValue(*p.DefaultValue, p.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to convert default value for parameter %s: %w", p.Name, err)
			}
			processedArgs[p.Name] = convertedVal
			log.Printf("Applied default value for input parameter %s: %v", p.Name, convertedVal)
		}
		// 验证必填参数
		if p.Required {
			finalVal, exists := processedArgs[p.Name]
			if !exists || finalVal == nil {
				return nil, fmt.Errorf("missing required parameter: %s", p.Name)
			}
		}
	}

	return processedArgs, nil
}

// convertDefaultValue 根据参数类型转换默认值字符串
func convertDefaultValue(defaultValue string, paramType string) (any, error) {
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
