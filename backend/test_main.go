package main

import (
	"log"
	"net/http"
	"os"

	"mcp-adapter/backend/config"
	"mcp-adapter/backend/handlers"
	"mcp-adapter/backend/routes"

	"github.com/gin-gonic/gin"
)

// 简化版本的main函数，用于测试API结构
func testMain() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建一个模拟的应用处理器
	appHandler := &handlers.ApplicationHandler{}
	
	// 创建一个模拟的接口处理器
	var interfaceHandler *handlers.InterfaceHandler

	// 设置路由
	router := routes.SetupRoutes(appHandler, interfaceHandler)

	// 添加一个简单的测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "MCP Adapter Backend API is working!",
			"version": "1.0.0",
			"status":  "ok",
		})
	})

	// 启动服务器
	log.Printf("Test server starting on %s", cfg.Server.GetServerAddr())
	log.Printf("Test endpoint: http://%s/test", cfg.Server.GetServerAddr())
	
	if err := router.Run(cfg.Server.GetServerAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func init() {
	// 如果是测试模式，运行测试版本
	if len(os.Args) > 1 && os.Args[1] == "test" {
		testMain()
	}
}