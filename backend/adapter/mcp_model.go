package adapter

import (
	"context"
	"encoding/json"
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
	AddTool int = iota
	RemoveTool
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
		sseServer[app.Path] = &Server{
			protocol: app.Protocol,
			path:     app.Path,
			server:   mcpServer,
			impl: server.NewSSEServer(
				mcpServer,
				server.WithSSEEndpoint(fmt.Sprintf("/sse/%s", app.Path)),
				server.WithMessageEndpoint(fmt.Sprintf("/message/%s", app.Path)),
			),
		}
		for _, iface := range interfaces {
			go addTool(&iface, &app)
		}
		log.Printf("Added SSE server: %s, path : %s", app.Name, fmt.Sprintf("/sse/%s", app.Path))
	}

	go func() {
		for {
			evt := <-event
			if evt.Code == AddTool {
				addTool(evt.Interface, evt.App)
			} else if evt.Code == RemoveTool {
				removeTool(evt.Interface, evt.App)
			}
		}
	}()
}

func GetServerImpl(path string) http.Handler {
	if s, ok := sseServer[path]; ok {
		return s.impl
	}
	return nil
}

type ToolParameter struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"oneof=sse streamable"`
	Required    bool   `json:"required"`
	Location    string `json:"location" validate:"oneof=query header body"`
	Description string `json:"description"`
}
type ToolOptions struct {
	Method            string          `json:"method"`
	Parameters        []ToolParameter `json:"parameters"`
	DefaultParameters []ToolParameter `json:"defaultParams"`
	DefaultHeaders    []ToolParameter `json:"defaultHeaders"`
}

func addTool(iface *models.Interface, app *models.Application) {

	if s, ok := sseServer[app.Path]; ok {
		tool := s.server.GetTool(iface.Name)
		if tool != nil {
			log.Printf("tool %s in %s already exists, skipped!", iface.Name, app.Name)
			return
		}
		var spec ToolOptions
		err := json.Unmarshal([]byte(iface.Options), &spec)
		if err != nil {
			log.Fatalf("Error unmarshalling options: %v", err)
			return
		}
		options := make([]mcp.ToolOption, 0)
		options = append(options, mcp.WithDescription(iface.Description))
		for _, p := range spec.Parameters {
			pos := make([]mcp.PropertyOption, 0)
			pos = append(pos, mcp.Description(p.Description))
			if p.Required {
				pos = append(pos, mcp.Required())
			}
			if p.Type == "string" {
				options = append(options, mcp.WithString(p.Name, pos...))
			} else if p.Type == "int64" {
				options = append(options, mcp.WithNumber(p.Name, pos...))
			} else if p.Type == "bool" {
				options = append(options, mcp.WithBoolean(p.Name, pos...))
			} else {
				log.Printf("Unknown option type: %s, using string as default", p.Type)
				options = append(options, mcp.WithString(p.Name, pos...))
			}
		}
		newTool := mcp.NewTool(iface.Name, options...)

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
	}
}

func removeTool(iface *models.Interface, app *models.Application) {
	if s, ok := sseServer[app.Path]; ok {
		s.server.DeleteTools(iface.Name)
	}
}
