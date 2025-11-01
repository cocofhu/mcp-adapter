package main

import (
	"log"
	"strconv"

	"mcp-adapter/backend/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 简单的内存存储
var (
	applications = make(map[uint]*dto.ApplicationResponse)
	interfaces   = make(map[uint]*dto.InterfaceResponse)
	nextAppID    uint = 1
	nextIfaceID  uint = 1
	validator_   = validator.New()
)

func main() {
	// 初始化一些示例数据
	initSampleData()

	// 创建Gin路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "MCP Adapter API is running",
		})
	})

	// API路由
	api := r.Group("/api")
	{
		// 应用管理
		api.POST("/applications", createApplication)
		api.GET("/applications", getApplications)
		api.GET("/applications/:id", getApplication)
		api.PUT("/applications/:id", updateApplication)
		api.DELETE("/applications/:id", deleteApplication)

		// 接口管理
		api.POST("/interfaces", createInterface)
		api.GET("/interfaces", getInterfaces)
		api.GET("/interfaces/:id", getInterface)
		api.PUT("/interfaces/:id", updateInterface)
		api.DELETE("/interfaces/:id", deleteInterface)
		api.PATCH("/interfaces/:id/toggle", toggleInterface)
	}

	log.Println("Server starting on :8080")
	log.Println("Health check: http://localhost:8080/health")
	log.Println("API base URL: http://localhost:8080/api")
	
	r.Run(":8080")
}

func initSampleData() {
	// 示例应用
	applications[1] = &dto.ApplicationResponse{
		ID:             1,
		Name:           "天气服务API",
		Description:    "提供全球天气信息查询服务",
		Version:        "v1.0.0",
		BaseURL:        "https://api.openweathermap.org",
		InterfaceCount: 2,
	}
	
	applications[2] = &dto.ApplicationResponse{
		ID:             2,
		Name:           "用户管理系统",
		Description:    "用户注册、登录、信息管理等功能",
		Version:        "v2.1.0",
		BaseURL:        "https://api.example.com",
		InterfaceCount: 1,
	}

	nextAppID = 3

	// 示例接口
	interfaces[1] = &dto.InterfaceResponse{
		ID:                1,
		AppID:             1,
		Name:              "weather_api",
		Description:       "获取天气信息的API接口",
		Protocol:          "http",
		Method:            "GET",
		URL:               "https://api.openweathermap.org/data/2.5/weather",
		AuthType:          "api-key",
		Status:            "active",
		Enabled:           true,
		HTTPParamLocation: "query",
	}

	interfaces[2] = &dto.InterfaceResponse{
		ID:                2,
		AppID:             1,
		Name:              "forecast_api",
		Description:       "获取天气预报的API接口",
		Protocol:          "http",
		Method:            "GET",
		URL:               "https://api.openweathermap.org/data/2.5/forecast",
		AuthType:          "api-key",
		Status:            "active",
		Enabled:           true,
		HTTPParamLocation: "query",
	}

	nextIfaceID = 3
}

// 应用管理API
func createApplication(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if err := validator_.Struct(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "参数验证失败",
			Error:   err.Error(),
		})
		return
	}

	app := &dto.ApplicationResponse{
		ID:             nextAppID,
		Name:           req.Name,
		Description:    req.Description,
		Version:        req.Version,
		BaseURL:        req.BaseURL,
		InterfaceCount: 0,
	}

	if app.Version == "" {
		app.Version = "v1.0.0"
	}

	applications[nextAppID] = app
	nextAppID++

	c.JSON(201, dto.APIResponse{
		Success: true,
		Message: "应用创建成功",
		Data:    app,
	})
}

func getApplications(c *gin.Context) {
	var apps []dto.ApplicationResponse
	for _, app := range applications {
		apps = append(apps, *app)
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "获取应用列表成功",
		Data: dto.PaginationResponse{
			Data:       apps,
			Total:      int64(len(apps)),
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	})
}

func getApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	app, exists := applications[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "应用不存在",
			Error:   "application not found",
		})
		return
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "获取应用详情成功",
		Data:    app,
	})
}

func updateApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	app, exists := applications[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "应用不存在",
			Error:   "application not found",
		})
		return
	}

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	app.Name = req.Name
	app.Description = req.Description
	app.Version = req.Version
	app.BaseURL = req.BaseURL

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "应用更新成功",
		Data:    app,
	})
}

func deleteApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	_, exists := applications[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "应用不存在",
			Error:   "application not found",
		})
		return
	}

	delete(applications, uint(id))

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "应用删除成功",
	})
}

// 接口管理API
func createInterface(c *gin.Context) {
	var req dto.CreateInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	iface := &dto.InterfaceResponse{
		ID:                nextIfaceID,
		AppID:             req.AppID,
		Name:              req.Name,
		Description:       req.Description,
		Protocol:          req.Protocol,
		Method:            req.Method,
		URL:               req.URL,
		AuthType:          req.AuthType,
		Status:            "active",
		Enabled:           true,
		HTTPParamLocation: req.HTTPParamLocation,
	}

	interfaces[nextIfaceID] = iface
	nextIfaceID++

	// 更新应用的接口数量
	if app, exists := applications[req.AppID]; exists {
		app.InterfaceCount++
	}

	c.JSON(201, dto.APIResponse{
		Success: true,
		Message: "接口创建成功",
		Data:    iface,
	})
}

func getInterfaces(c *gin.Context) {
	appIDStr := c.Query("app_id")
	if appIDStr == "" {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "应用ID不能为空",
			Error:   "app_id is required",
		})
		return
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	var ifaces []dto.InterfaceResponse
	for _, iface := range interfaces {
		if iface.AppID == uint(appID) {
			ifaces = append(ifaces, *iface)
		}
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "获取接口列表成功",
		Data: dto.PaginationResponse{
			Data:       ifaces,
			Total:      int64(len(ifaces)),
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	})
}

func getInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	iface, exists := interfaces[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "接口不存在",
			Error:   "interface not found",
		})
		return
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "获取接口详情成功",
		Data:    iface,
	})
}

func updateInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	iface, exists := interfaces[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "接口不存在",
			Error:   "interface not found",
		})
		return
	}

	var req dto.UpdateInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	iface.Name = req.Name
	iface.Description = req.Description
	iface.Protocol = req.Protocol
	iface.Method = req.Method
	iface.URL = req.URL
	iface.AuthType = req.AuthType
	iface.Status = req.Status
	if req.Enabled != nil {
		iface.Enabled = *req.Enabled
	}
	iface.HTTPParamLocation = req.HTTPParamLocation

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "接口更新成功",
		Data:    iface,
	})
}

func deleteInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	iface, exists := interfaces[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "接口不存在",
			Error:   "interface not found",
		})
		return
	}

	delete(interfaces, uint(id))

	// 更新应用的接口数量
	if app, exists := applications[iface.AppID]; exists {
		app.InterfaceCount--
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "接口删除成功",
	})
}

func toggleInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	iface, exists := interfaces[uint(id)]
	if !exists {
		c.JSON(404, dto.ErrorResponse{
			Success: false,
			Message: "接口不存在",
			Error:   "interface not found",
		})
		return
	}

	var req dto.ToggleInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	iface.Enabled = req.Enabled
	if req.Enabled {
		iface.Status = "active"
	} else {
		iface.Status = "inactive"
	}

	c.JSON(200, dto.APIResponse{
		Success: true,
		Message: "接口状态切换成功",
		Data:    iface,
	})
}