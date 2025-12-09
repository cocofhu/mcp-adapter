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

	// 执行数据迁移
	migrateCustomTypeFieldAppID(db)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return dbInstance
}

// migrateCustomTypeFieldAppID 为现有的 CustomTypeField 记录填充 AppID
func migrateCustomTypeFieldAppID(db *gorm.DB) {
	log.Println("Starting migration: filling AppID for CustomTypeField records...")

	// 检查是否有需要迁移的数据（AppID 为 0 或 NULL 的记录）
	var count int64
	db.Model(&models.CustomTypeField{}).Where("app_id = 0 OR app_id IS NULL").Count(&count)

	if count == 0 {
		log.Println("No CustomTypeField records need migration")
		return
	}

	log.Printf("Found %d CustomTypeField records that need AppID migration", count)

	// 获取所有需要迁移的字段
	var fields []models.CustomTypeField
	if err := db.Where("app_id = 0 OR app_id IS NULL").Find(&fields).Error; err != nil {
		log.Printf("Failed to fetch CustomTypeField records for migration: %v", err)
		return
	}

	// 批量更新：通过 CustomType 获取 AppID
	successCount := 0
	failCount := 0

	for _, field := range fields {
		var customType models.CustomType
		if err := db.First(&customType, field.CustomTypeID).Error; err != nil {
			log.Printf("Failed to find CustomType (ID: %d) for CustomTypeField (ID: %d): %v",
				field.CustomTypeID, field.ID, err)
			failCount++
			continue
		}

		// 更新字段的 AppID
		if err := db.Model(&field).Update("app_id", customType.AppID).Error; err != nil {
			log.Printf("Failed to update AppID for CustomTypeField (ID: %d): %v", field.ID, err)
			failCount++
			continue
		}

		successCount++
	}

	log.Printf("Migration completed: %d records updated successfully, %d failed", successCount, failCount)
}
