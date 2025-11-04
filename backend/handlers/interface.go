package handlers

import (
	"mcp-adapter/backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateInterface 创建接口
func CreateInterface(c *gin.Context) {
	var req service.CreateInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	resp, err := service.CreateInterface(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Interface)
}

// GetInterfaces 获取所有接口
func GetInterfaces(c *gin.Context) {
	var req service.ListInterfacesRequest
	if appIDStr := c.Query("app_id"); appIDStr != "" {
		if id, err := strconv.ParseInt(appIDStr, 10, 64); err == nil {
			req.AppID = id
		}
	}
	resp, err := service.ListInterfaces(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Interfaces)
}

// GetInterface 获取单个接口
func GetInterface(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid interface ID")
		return
	}

	resp, err := service.GetInterface(service.GetInterfaceRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Interface)
}

// UpdateInterface 更新接口（部分字段）
func UpdateInterface(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid interface ID")
		return
	}

	var body service.UpdateInterfaceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	body.ID = id

	resp, err := service.UpdateInterface(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Interface)
}

// DeleteInterface 删除接口
func DeleteInterface(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid interface ID")
		return
	}
	_, err = service.DeleteInterface(service.DeleteInterfaceRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
