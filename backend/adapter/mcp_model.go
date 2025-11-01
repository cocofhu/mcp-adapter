package adapter

import (
	"context"
	"fmt"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var event chan Event
var sseServer map[string]*Server

const (
	AddInterface int = iota
	RemoveInterface
	AddApplication
	RemoveApplication
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
	Code      int
}

func InitServer() {
	sseServer = make(map[string]*Server)
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
		mcpServer.AddTool(mcp.NewTool("echo"), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(fmt.Sprintf("Echo: %v", req.GetArguments()["message"])), nil
		})

		sseServer[app.Path] = &Server{
			protocol: app.Protocol,
			path:     app.Path,
			server:   mcpServer,
			impl: server.NewSSEServer(
				mcpServer,
				server.WithSSEEndpoint(fmt.Sprintf("/sse/%s", app.Path)),
			),
		}
		log.Printf("Added SSE server: %s, path : %s", app.Name, fmt.Sprintf("/sse/%s", app.Path))

	}

	for {
		evt := <-event
		if evt.Code == AddInterface {

		}
	}
}

func GetServerImpl(path string) http.Handler {
	if s, ok := sseServer[path]; ok {
		return s.impl
	}
	return nil
}

func addTool(iface *models.Interface, app *models.Application) {
	if s, ok := sseServer[app.Path]; ok {
		tool := s.server.GetTool(iface.Name)
		if tool != nil {
			log.Printf("tool %s in %s already exists, skipped!", iface.Name, app.Name)
			return
		}

	}
}
