package handlers

import (
	"mcp-adapter/backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateCustomType 创建自定义类型
func CreateCustomType(c *gin.Context) {
	var req service.CreateCustomTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	resp, err := service.CreateCustomType(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.CustomType)
}

// GetCustomTypes 获取应用下的所有自定义类型
func GetCustomTypes(c *gin.Context) {
	appID, err := strconv.ParseInt(c.Query("app_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid app_id parameter")
		return
	}
	resp, err := service.ListCustomTypes(service.ListCustomTypesRequest{AppID: appID})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCustomType 获取单个自定义类型
func GetCustomType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid custom type ID")
		return
	}
	resp, err := service.GetCustomType(service.GetCustomTypeRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.CustomType)
}

// UpdateCustomType 更新自定义类型
func UpdateCustomType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid custom type ID")
		return
	}

	var body service.UpdateCustomTypeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	body.ID = id

	resp, err := service.UpdateCustomType(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.CustomType)
}

// DeleteCustomType 删除自定义类型
func DeleteCustomType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid custom type ID")
		return
	}
	_, err = service.DeleteCustomType(service.DeleteCustomTypeRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
