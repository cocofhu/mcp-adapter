package handlers

import (
	"net/http"
	"strconv"

	"mcp-adapter/backend/dto"
	"mcp-adapter/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ApplicationHandler 应用处理器
type ApplicationHandler struct {
	appService services.ApplicationService
	validator  *validator.Validate
}

// NewApplicationHandler 创建应用处理器
func NewApplicationHandler(appService services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
		validator:  validator.New(),
	}
}

// CreateApplication 创建应用
// @Summary 创建应用
// @Description 创建新的API应用
// @Tags applications
// @Accept json
// @Produce json
// @Param request body dto.CreateApplicationRequest true "创建应用请求"
// @Success 201 {object} dto.APIResponse{data=dto.ApplicationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications [post]
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "参数验证失败",
			Error:   err.Error(),
		})
		return
	}

	app, err := h.appService.CreateApplication(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "创建应用失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "应用创建成功",
		Data:    app,
	})
}

// GetApplications 获取应用列表
// @Summary 获取应用列表
// @Description 分页获取应用列表
// @Tags applications
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Success 200 {object} dto.APIResponse{data=dto.PaginationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications [get]
func (h *ApplicationHandler) GetApplications(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "查询参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}

	result, err := h.appService.GetApplications(c.Request.Context(), &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "获取应用列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "获取应用列表成功",
		Data:    result,
	})
}

// GetApplication 获取单个应用
// @Summary 获取应用详情
// @Description 根据ID获取应用详情
// @Tags applications
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} dto.APIResponse{data=dto.ApplicationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications/{id} [get]
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	app, err := h.appService.GetApplicationByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Success: false,
			Message: "应用不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "获取应用详情成功",
		Data:    app,
	})
}

// UpdateApplication 更新应用
// @Summary 更新应用
// @Description 更新应用信息
// @Tags applications
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Param request body dto.UpdateApplicationRequest true "更新应用请求"
// @Success 200 {object} dto.APIResponse{data=dto.ApplicationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications/{id} [put]
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "参数验证失败",
			Error:   err.Error(),
		})
		return
	}

	app, err := h.appService.UpdateApplication(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "更新应用失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "应用更新成功",
		Data:    app,
	})
}

// DeleteApplication 删除应用
// @Summary 删除应用
// @Description 删除应用及其所有接口
// @Tags applications
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications/{id} [delete]
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	err = h.appService.DeleteApplication(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "删除应用失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "应用删除成功",
	})
}

// GetApplicationStats 获取应用统计信息
// @Summary 获取应用统计信息
// @Description 获取应用的统计信息，如接口数量、状态分布等
// @Tags applications
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} dto.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/applications/{id}/stats [get]
func (h *ApplicationHandler) GetApplicationStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	stats, err := h.appService.GetApplicationStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "获取应用统计信息失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "获取应用统计信息成功",
		Data:    stats,
	})
}

// InterfaceHandler 接口处理器
type InterfaceHandler struct {
	interfaceService services.InterfaceService
	validator        *validator.Validate
}

// NewInterfaceHandler 创建接口处理器
func NewInterfaceHandler(interfaceService services.InterfaceService) *InterfaceHandler {
	return &InterfaceHandler{
		interfaceService: interfaceService,
		validator:        validator.New(),
	}
}

// CreateInterface 创建接口
// @Summary 创建接口
// @Description 在指定应用下创建新的API接口
// @Tags interfaces
// @Accept json
// @Produce json
// @Param request body dto.CreateInterfaceRequest true "创建接口请求"
// @Success 201 {object} dto.APIResponse{data=dto.InterfaceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces [post]
func (h *InterfaceHandler) CreateInterface(c *gin.Context) {
	var req dto.CreateInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "参数验证失败",
			Error:   err.Error(),
		})
		return
	}

	iface, err := h.interfaceService.CreateInterface(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "创建接口失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "接口创建成功",
		Data:    iface,
	})
}

// GetInterfaces 获取接口列表
// @Summary 获取接口列表
// @Description 获取指定应用下的接口列表
// @Tags interfaces
// @Accept json
// @Produce json
// @Param app_id query int true "应用ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Param status query string false "状态过滤" Enums(active,inactive,error,enabled,disabled)
// @Param protocol query string false "协议过滤" Enums(http,https)
// @Success 200 {object} dto.APIResponse{data=dto.PaginationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces [get]
func (h *InterfaceHandler) GetInterfaces(c *gin.Context) {
	appIDStr := c.Query("app_id")
	if appIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "应用ID不能为空",
			Error:   "app_id is required",
		})
		return
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的应用ID",
			Error:   err.Error(),
		})
		return
	}

	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "查询参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}

	result, err := h.interfaceService.GetInterfaces(c.Request.Context(), uint(appID), &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "获取接口列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "获取接口列表成功",
		Data:    result,
	})
}

// GetInterface 获取单个接口
// @Summary 获取接口详情
// @Description 根据ID获取接口详情
// @Tags interfaces
// @Accept json
// @Produce json
// @Param id path int true "接口ID"
// @Success 200 {object} dto.APIResponse{data=dto.InterfaceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces/{id} [get]
func (h *InterfaceHandler) GetInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	iface, err := h.interfaceService.GetInterfaceByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Success: false,
			Message: "接口不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "获取接口详情成功",
		Data:    iface,
	})
}

// UpdateInterface 更新接口
// @Summary 更新接口
// @Description 更新接口信息
// @Tags interfaces
// @Accept json
// @Produce json
// @Param id path int true "接口ID"
// @Param request body dto.UpdateInterfaceRequest true "更新接口请求"
// @Success 200 {object} dto.APIResponse{data=dto.InterfaceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces/{id} [put]
func (h *InterfaceHandler) UpdateInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "参数验证失败",
			Error:   err.Error(),
		})
		return
	}

	iface, err := h.interfaceService.UpdateInterface(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "更新接口失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "接口更新成功",
		Data:    iface,
	})
}

// DeleteInterface 删除接口
// @Summary 删除接口
// @Description 删除接口及其相关数据
// @Tags interfaces
// @Accept json
// @Produce json
// @Param id path int true "接口ID"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces/{id} [delete]
func (h *InterfaceHandler) DeleteInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	err = h.interfaceService.DeleteInterface(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "删除接口失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "接口删除成功",
	})
}

// ToggleInterface 切换接口状态
// @Summary 切换接口启用状态
// @Description 启用或禁用接口
// @Tags interfaces
// @Accept json
// @Produce json
// @Param id path int true "接口ID"
// @Param request body dto.ToggleInterfaceRequest true "切换状态请求"
// @Success 200 {object} dto.APIResponse{data=dto.InterfaceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces/{id}/toggle [patch]
func (h *InterfaceHandler) ToggleInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.ToggleInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	iface, err := h.interfaceService.ToggleInterface(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "切换接口状态失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "接口状态切换成功",
		Data:    iface,
	})
}

// TestInterface 测试接口
// @Summary 测试接口
// @Description 发送HTTP请求测试接口
// @Tags interfaces
// @Accept json
// @Produce json
// @Param id path int true "接口ID"
// @Param request body dto.TestInterfaceRequest true "测试请求"
// @Success 200 {object} dto.APIResponse{data=dto.TestInterfaceResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/interfaces/{id}/test [post]
func (h *InterfaceHandler) TestInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "无效的接口ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.TestInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.interfaceService.TestInterface(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success: false,
			Message: "测试接口失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "接口测试完成",
		Data:    result,
	})
}