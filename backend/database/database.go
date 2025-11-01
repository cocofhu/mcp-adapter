package database

import (
	"fmt"
	"log"
	"time"

	"mcp-adapter/backend/config"
	"mcp-adapter/backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector

	// 根据驱动类型选择方言
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.GetDSN())
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
	case "sqlite":
		dialector = sqlite.Open(cfg.GetDSN())
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// 连接数据库
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to %s database", cfg.Driver)
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 按依赖顺序迁移表
	err := DB.AutoMigrate(
		&models.Application{},
		&models.Interface{},
		&models.Parameter{},
		&models.DefaultParam{},
		&models.DefaultHeader{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// SeedData 初始化种子数据
func SeedData() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 检查是否已有数据
	var count int64
	DB.Model(&models.Application{}).Count(&count)
	if count > 0 {
		log.Println("Database already has data, skipping seed")
		return nil
	}

	// 创建示例应用
	apps := []models.Application{
		{
			Name:        "天气服务API",
			Description: "提供全球天气信息查询服务，包括实时天气、天气预报等功能",
			Version:     "v1.0.0",
			BaseURL:     "https://api.openweathermap.org",
		},
		{
			Name:        "用户管理系统",
			Description: "用户注册、登录、信息管理等核心功能接口",
			Version:     "v2.1.0",
			BaseURL:     "https://api.example.com",
		},
	}

	for i := range apps {
		if err := DB.Create(&apps[i]).Error; err != nil {
			return fmt.Errorf("failed to create application: %w", err)
		}
	}

	// 创建示例接口
	interfaces := []models.Interface{
		{
			AppID:               apps[0].ID,
			Name:                "weather_api",
			Description:         "获取天气信息的API接口",
			Protocol:            "http",
			Method:              "GET",
			URL:                 "https://api.openweathermap.org/data/2.5/weather",
			AuthType:            "api-key",
			AuthValue:           "your-api-key",
			Status:              "active",
			Enabled:             true,
			HTTPParamLocation:   "query",
		},
		{
			AppID:               apps[0].ID,
			Name:                "forecast_api",
			Description:         "获取天气预报的API接口",
			Protocol:            "http",
			Method:              "GET",
			URL:                 "https://api.openweathermap.org/data/2.5/forecast",
			AuthType:            "api-key",
			AuthValue:           "your-api-key",
			Status:              "active",
			Enabled:             true,
			HTTPParamLocation:   "query",
		},
		{
			AppID:               apps[1].ID,
			Name:                "user_api",
			Description:         "用户信息API接口",
			Protocol:            "http",
			Method:              "POST",
			URL:                 "https://api.example.com/users",
			AuthType:            "bearer",
			AuthValue:           "eyJhbGciOiJIUzI1NiIs...",
			Status:              "inactive",
			Enabled:             false,
			HTTPParamLocation:   "body",
		},
	}

	for i := range interfaces {
		if err := DB.Create(&interfaces[i]).Error; err != nil {
			return fmt.Errorf("failed to create interface: %w", err)
		}
	}

	// 创建示例参数
	parameters := []models.Parameter{
		{
			InterfaceID: interfaces[0].ID,
			Name:        "q",
			Type:        "string",
			Location:    "query",
			Required:    true,
			Description: "城市名称",
		},
		{
			InterfaceID: interfaces[0].ID,
			Name:        "appid",
			Type:        "string",
			Location:    "query",
			Required:    true,
			Description: "API密钥",
		},
		{
			InterfaceID: interfaces[0].ID,
			Name:        "units",
			Type:        "string",
			Location:    "query",
			Required:    false,
			Description: "单位制",
		},
	}

	for i := range parameters {
		if err := DB.Create(&parameters[i]).Error; err != nil {
			return fmt.Errorf("failed to create parameter: %w", err)
		}
	}

	// 创建示例默认参数
	defaultParams := []models.DefaultParam{
		{
			InterfaceID: interfaces[0].ID,
			Name:        "lang",
			Value:       "zh_cn",
			Description: "语言设置",
		},
		{
			InterfaceID: interfaces[0].ID,
			Name:        "mode",
			Value:       "json",
			Description: "响应格式",
		},
	}

	for i := range defaultParams {
		if err := DB.Create(&defaultParams[i]).Error; err != nil {
			return fmt.Errorf("failed to create default param: %w", err)
		}
	}

	// 创建示例默认请求头
	defaultHeaders := []models.DefaultHeader{
		{
			InterfaceID: interfaces[0].ID,
			Name:        "User-Agent",
			Value:       "WeatherApp/1.0",
			Description: "用户代理",
		},
		{
			InterfaceID: interfaces[0].ID,
			Name:        "Accept",
			Value:       "application/json",
			Description: "接受的内容类型",
		},
	}

	for i := range defaultHeaders {
		if err := DB.Create(&defaultHeaders[i]).Error; err != nil {
			return fmt.Errorf("failed to create default header: %w", err)
		}
	}

	log.Println("Database seed data created successfully")
	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}