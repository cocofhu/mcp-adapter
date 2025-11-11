package routes

import (
	"log"
	"mcp-adapter/backend/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.New()
	// 日志与恢复中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 启用CORS
	r.Use(corsMiddleware())

	// API路由组
	api := r.Group("/api")
	{
		// 应用相关路由
		api.POST("/applications", handlers.CreateApplication)
		api.GET("/applications", handlers.GetApplications)
		api.GET("/applications/:id", handlers.GetApplication)
		api.PUT("/applications/:id", handlers.UpdateApplication)
		api.DELETE("/applications/:id", handlers.DeleteApplication)

		// 接口相关路由
		api.POST("/interfaces", handlers.CreateInterface)
		api.GET("/interfaces", handlers.GetInterfaces)
		api.GET("/interfaces/:id", handlers.GetInterface)
		api.PUT("/interfaces/:id", handlers.UpdateInterface)
		api.DELETE("/interfaces/:id", handlers.DeleteInterface)

		// 自定义类型相关路由
		api.POST("/custom-types", handlers.CreateCustomType)
		api.GET("/custom-types", handlers.GetCustomTypes) // 需要 app_id 查询参数
		api.GET("/custom-types/:id", handlers.GetCustomType)
		api.PUT("/custom-types/:id", handlers.UpdateCustomType)
		api.DELETE("/custom-types/:id", handlers.DeleteCustomType)
	}

	// 静态文件服务
	r.Static("/static", "./web/static")

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	// MCP-SSE服务
	r.Any("/sse/:path", handlers.ServeSSE)
	r.Any("/message/:path", handlers.ServeSSE)
	// MCP-Streamable服务
	r.Any("/streamable/:path", handlers.ServeStreamable)

	log.Println("Routes initialized with Gin")
	return r
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Mcp-Protocol-Version, Mcp-Session-Id")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusOK)
			c.Abort()
			return
		}
		c.Next()
	}
}
