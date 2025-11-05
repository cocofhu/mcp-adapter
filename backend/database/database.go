package database

import (
	"log"
	"mcp-adapter/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(dbPath string) {
	// 使用纯 Go SQLite 驱动
	dsn := dbPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite", // 使用 modernc.org/sqlite 驱动
		DSN:        dsn,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Application{},
		&models.Interface{},
		&models.CustomType{},
		&models.CustomTypeField{},
		&models.InterfaceParameter{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	DB = db
	log.Println("Database connected and migrated successfully")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
