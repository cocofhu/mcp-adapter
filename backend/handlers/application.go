package handlers

import (
	"mcp-adapter/backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateApplication 创建应用
func CreateApplication(c *gin.Context) {
	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	resp, err := service.CreateApplication(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Application)
}

// GetApplications 获取所有应用
func GetApplications(c *gin.Context) {
	resp, err := service.ListApplications(service.ListApplicationsRequest{})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to fetch applications")
		return
	}
	c.JSON(http.StatusOK, resp.Applications)
}

// GetApplicationDetail 获取单个应用详情 这个接口不要暴露MCP比较好
func GetApplicationDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid application ID")
		return
	}
	resp, err := service.GetApplication(service.GetApplicationRequest{ID: id, ShowDetail: true})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetApplication 获取单个应用
func GetApplication(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid application ID")
		return
	}
	resp, err := service.GetApplication(service.GetApplicationRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Application)
}

// UpdateApplication 更新应用（部分字段）
func UpdateApplication(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid application ID")
		return
	}

	var body service.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON format")
		return
	}
	body.ID = id

	resp, err := service.UpdateApplication(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp.Application)
}

// DeleteApplication 删除应用
func DeleteApplication(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid application ID")
		return
	}
	_, err = service.DeleteApplication(service.DeleteApplicationRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
