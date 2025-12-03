package database

import (
	"log"
	"mcp-adapter/backend/models"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var dbInstance *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(dbPath string) {
	var db *gorm.DB
	var err error

	// 检查是否设置了 MYSQL_DSN 环境变量
	mysqlDSN := os.Getenv("MYSQL_DSN")

	if mysqlDSN != "" {
		// 使用 MySQL
		log.Println("Using MySQL database")
		db, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err != nil {
			log.Fatal("Failed to connect to MySQL database:", err)
		}
	} else {
		// 使用纯 Go SQLite 驱动
		log.Println("Using SQLite database")
		dsn := dbPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"

		db, err = gorm.Open(sqlite.Dialector{
			DriverName: "sqlite", // 使用 modernc.org/sqlite 驱动
			DSN:        dsn,
		}, &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err != nil {
			log.Fatal("Failed to connect to SQLite database:", err)
		}
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Application{},
		&models.Interface{},
		&models.CustomType{},
		&models.CustomTypeField{},
		&models.InterfaceParameter{},
		&models.EventLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	dbInstance = db
	log.Println("Database connected and migrated successfully")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return dbInstance
}
