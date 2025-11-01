package database

import (
	"log"
	"mcp-adapter/backend/models"

	"gorm.io/driver/sqlite"
	_ "modernc.org/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	var err error
	DB, err = gorm.Open(sqlite.Open("mcp_adapter.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(&models.Application{}, &models.Interface{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}