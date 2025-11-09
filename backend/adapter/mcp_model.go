package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var sseServer sync.Map
var event chan Event

type EventCode int

const (
	AddToolEvent           EventCode = iota // 工具添加事件
	RemoveToolEvent                         // 工具移除事件
	AddApplicationEvent                     // 应用添加事件
	RemoveApplicationEvent                  // 应用移除事件
	ToolListChanged                         // 工具列表变更事件
)

type Server struct {
	protocol string
	path     string
	server   *server.MCPServer
	impl     http.Handler
}
type Event struct {
	Interface *models.Interface
	App       *models.Application
	Code      EventCode
}

func SendEvent(evt Event) {
	log.Printf("Sent event: %v", evt)
	event <- evt
}

func InitServer() {
	event = make(chan Event, 16)
	var apps []models.Application
	db := database.GetDB()
	if err := db.Find(&apps).Error; err != nil {
		log.Fatalf("Error getting applications: %v", err)
		return
	}
	for _, app := range apps {
		if app.Protocol != "sse" {
			log.Printf("Skipping %s protocol", app.Protocol)
		}
		var interfaces []models.Interface
		db := database.GetDB()
		query := db
		query = query.Where("app_id = ?", app.ID)
		if err := query.Find(&interfaces).Error; err != nil {
			log.Fatalf("Error getting interfaces: %v", err)
			return
		}
		mcpServer := server.NewMCPServer(app.Name, "1.0.0")
		sseServer.Store(app.Path, &Server{
			protocol: app.Protocol,
			path:     app.Path,
			server:   mcpServer,
			impl: server.NewSSEServer(
				mcpServer,
				server.WithSSEEndpoint(fmt.Sprintf("/sse/%s", app.Path)),
				server.WithMessageEndpoint(fmt.Sprintf("/message/%s", app.Path)),
			),
		})
		for _, iface := range interfaces {
			addTool(&iface, &app)
		}
		log.Printf("Added SSE server: %s, path : %s", app.Name, fmt.Sprintf("/sse/%s", app.Path))
	}

	go func() {
		for {
			evt := <-event
			if evt.Code == AddToolEvent {
				addTool(evt.Interface, evt.App)
			} else if evt.Code == RemoveToolEvent {
				removeTool(evt.Interface, evt.App)
			} else if evt.Code == ToolListChanged {
				toolChanged(evt.Interface, evt.App)
			}
		}
	}()
}

func GetServerImpl(path string) http.Handler {
	if s, ok := sseServer.Load(path); ok {
		return s.(*Server).impl
	}
	return nil
}

func addTool(iface *models.Interface, app *models.Application) {
	if s, ok := sseServer.Load(app.Path); ok {
		s := s.(*Server)
		tool := s.server.GetTool(iface.Name)
		if tool != nil {
			log.Printf("tool %s in %s already exists, skipped!", iface.Name, app.Name)
			return
		}
		// 从数据库获取接口参数
		db := database.GetDB()
		var params []models.InterfaceParameter
		if db.Where("interface_id = ?", iface.ID).Find(&params).Error != nil {
			log.Printf("Error getting interface parameters for tool %s", iface.Name)
			return
		}
		schema, err := BuildMcpInputSchemaByInterface(iface.ID)
		if err != nil {
			log.Printf("Error building input schema for tool %s: %v", iface.Name, err)
			return
		}
		marshal, err := json.Marshal(schema)
		if err != nil {
			log.Printf("Error marshaling input schema for tool %s: %v", iface.Name, err)
			return
		}
		log.Printf("Input schema for tool %s: %s", iface.Name, string(marshal))
		newTool := mcp.NewToolWithRawSchema(iface.Name, iface.Description, marshal)
		s.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			data, code, err := CallHTTPInterface(ctx, iface, args)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if code != http.StatusOK {
				log.Printf("Error calling tool %s, code: %d", iface.Name, code)
			}
			return mcp.NewToolResultText(string(data)), nil
		})
		log.Printf("Added tool: %s", iface.Name)
	} else {
		log.Printf("Warning Skipping add tool %s in %s not found!", iface.Name, app.Name)
	}
}

func removeTool(iface *models.Interface, app *models.Application) {
	if s, ok := sseServer.Load(app.Path); ok {
		s.(*Server).server.DeleteTools(iface.Name)
		log.Printf("Removed tool: %s", iface.Name)
	} else {
		log.Printf("Warning Skipping remove tool %s in %s not found!", iface.Name, app.Name)
	}
}

func toolChanged(_ *models.Interface, app *models.Application) {
	if s, ok := sseServer.Load(app.Path); ok {
		s.(*Server).server.SendNotificationToAllClients(mcp.MethodNotificationToolsListChanged, nil)
		log.Printf("Tool changed notification sent: %s", app.Name)
	} else {
		log.Printf("Warning Skipping toolChanged in %s not found!", app.Name)
	}
}
