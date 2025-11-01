package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-adapter/backend/config"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/handlers"
	"mcp-adapter/backend/repositories"
	"mcp-adapter/backend/routes"
	"mcp-adapter/backend/services"

	"github.com/gin-gonic/gin"
)

// @title MCP Adapter API
// @version 1.0
// @description HTTP接口管理系统API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.InitDatabase(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化种子数据
	if err := database.SeedData(); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// 初始化依赖注入
	db := database.GetDB()

	// 初始化仓储层
	appRepo := repositories.NewApplicationRepository(db)
	interfaceRepo := repositories.NewInterfaceRepository(db)

	// 初始化服务层
	appService := services.NewApplicationService(appRepo)
	// TODO: 实现其他服务
	// interfaceService := services.NewInterfaceService(interfaceRepo, ...)

	// 初始化处理器
	appHandler := handlers.NewApplicationHandler(appService)
	// TODO: 实现接口处理器
	// interfaceHandler := handlers.NewInterfaceHandler(interfaceService)

	// 临时创建一个空的接口处理器用于路由设置
	var interfaceHandler *handlers.InterfaceHandler

	// 设置路由
	router := routes.SetupRoutes(appHandler, interfaceHandler)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		log.Printf("Swagger documentation available at: http://%s/swagger/index.html", server.Addr)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 设置5秒的超时时间来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// 健康检查处理器
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "MCP Adapter API is running",
		"timestamp": time.Now().Unix(),
	})
}

// 版本信息处理器
func version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "1.0.0",
		"build_time":  "2024-01-01T00:00:00Z", // 可以在编译时注入
		"git_commit":  "unknown",               // 可以在编译时注入
		"go_version":  "go1.21",
	})
}