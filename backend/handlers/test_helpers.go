package handlers

import (
	"mcp-adapter/backend/database"

	"github.com/gin-gonic/gin"
)

// setupTestRouter 创建测试用的Gin路由器
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// setupTestDB 初始化测试数据库
func setupTestDB() {
	// 使用内存数据库进行测试
	// 如果设置了 MYSQL_DSN 环境变量，将使用 MySQL，否则使用 SQLite
	database.InitDatabase(":memory:")
}

// cleanupTestDB 清理测试数据
func cleanupTestDB() {
	// 清理测试数据
	db := database.GetDB()
	db.Exec("DELETE FROM interface_parameters")
	db.Exec("DELETE FROM interfaces")
	db.Exec("DELETE FROM custom_type_fields")
	db.Exec("DELETE FROM custom_types")
	db.Exec("DELETE FROM applications")
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}
