package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Mode         string `json:"mode"` // debug, release, test
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `json:"driver"`   // mysql, postgres, sqlite
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
	SSLMode  string `json:"ssl_mode"`
	MaxIdle  int    `json:"max_idle"`
	MaxOpen  int    `json:"max_open"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"` // debug, info, warn, error
	Format string `json:"format"` // json, text
	Output string `json:"output"` // stdout, file
	File   string `json:"file"`
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 60),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 60),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			Username: getEnv("DB_USERNAME", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "mcp_adapter.db"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 10),
			MaxOpen:  getEnvInt("DB_MAX_OPEN", 100),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
			File:   getEnv("LOG_FILE", "logs/app.log"),
		},
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "sqlite":
		return c.Database
	default:
		return ""
	}
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}