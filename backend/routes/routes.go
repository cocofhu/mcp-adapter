package routes

import (
	"mcp-adapter/backend/handlers"
	"mcp-adapter/backend/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(
	appHandler *handlers.ApplicationHandler,
	interfaceHandler *handlers.InterfaceHandler,
) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "MCP Adapter API is running",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 应用管理路由
		applications := api.Group("/applications")
		{
			applications.POST("", appHandler.CreateApplication)
			applications.GET("", appHandler.GetApplications)
			applications.GET("/:id", appHandler.GetApplication)
			applications.PUT("/:id", appHandler.UpdateApplication)
			applications.DELETE("/:id", appHandler.DeleteApplication)
			applications.GET("/:id/stats", appHandler.GetApplicationStats)
		}

		// 接口管理路由
		interfaces := api.Group("/interfaces")
		{
			interfaces.POST("", interfaceHandler.CreateInterface)
			interfaces.GET("", interfaceHandler.GetInterfaces)
			interfaces.GET("/:id", interfaceHandler.GetInterface)
			interfaces.PUT("/:id", interfaceHandler.UpdateInterface)
			interfaces.DELETE("/:id", interfaceHandler.DeleteInterface)
			interfaces.PATCH("/:id/toggle", interfaceHandler.ToggleInterface)
			interfaces.POST("/:id/test", interfaceHandler.TestInterface)
		}

		// 批量操作路由
		batch := api.Group("/batch")
		{
			batch.PATCH("/interfaces/toggle", func(c *gin.Context) {
				// TODO: 实现批量切换接口状态
				c.JSON(200, gin.H{"message": "批量切换接口状态"})
			})
			batch.DELETE("/interfaces", func(c *gin.Context) {
				// TODO: 实现批量删除接口
				c.JSON(200, gin.H{"message": "批量删除接口"})
			})
		}
	}

	// 静态文件服务（如果需要）
	r.Static("/static", "./static")

	return r
}

// SetupTestRoutes 设置测试路由（用于单元测试）
func SetupTestRoutes(
	appHandler *handlers.ApplicationHandler,
	interfaceHandler *handlers.InterfaceHandler,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return SetupRoutes(appHandler, interfaceHandler)
}