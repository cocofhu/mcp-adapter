package models

import (
	"time"

	"gorm.io/gorm"
)

// Application 应用实体
type Application struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:255" validate:"required"` // 应用名称
	Description string         `json:"description" gorm:"size:500"`                       // 应用描述
	Path        string         `json:"path" gorm:"size:255"`                              // 应用路径标识
	Protocol    string         `json:"protocol" gorm:"size:255"`                          // 应用对外协议 sse, streamable
	PostProcess string         `json:"post_process" gorm:"size:1048576"`                  // 后处理脚本
	Environment string         `json:"environment" gorm:"size:1048576"`                   // 环境变量 (JSON String)
	Enabled     bool           `json:"enabled" gorm:"default:true"`                       // 是否启用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Interface 接口实体
type Interface struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	AppID       int64          `json:"app_id" gorm:"not null;index" validate:"required"`  // 应用ID 一个应用对应多个Interface
	Name        string         `json:"name" gorm:"not null;size:255" validate:"required"` // 接口名称
	Description string         `json:"description" gorm:"size:16384"`                     // 接口描述
	Protocol    string         `json:"protocol"`                                          // 接口协议: HTTP
	URL         string         `json:"url"`                                               // 接口地址
	Method      string         `json:"method" gorm:"size:50"`                             // HTTP方法: GET, POST, PUT, DELETE等
	AuthType    string         `json:"auth_type"`                                         // 鉴权类型
	Enabled     bool           `json:"enabled" gorm:"default:true"`                       // 是否启用
	PostProcess string         `json:"post_process" gorm:"size:1048576"`                  // 后处理脚本
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// CustomType 自定义类型定义（纯类型定义，不包含使用属性）
type CustomType struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	AppID       int64          `json:"app_id" gorm:"not null;index" validate:"required"`
	Name        string         `json:"name" gorm:"not null;size:255" validate:"required"` // 类型名称，如 "User", "Address"
	Description string         `json:"description" gorm:"size:16384"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// CustomTypeField 自定义类型的字段定义
type CustomTypeField struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	CustomTypeID int64          `json:"custom_type_id" gorm:"not null;index"`               // 所属类型ID
	Name         string         `json:"name" gorm:"not null;size:255" validate:"required"`  // 字段名
	Type         string         `json:"type" validate:"oneof=number string boolean custom"` // 字段类型
	Ref          *int64         `json:"ref"`                                                // 如果是 custom 类型，引用 CustomType.ID
	IsArray      bool           `json:"is_array"`                                           // 是否数组
	Required     bool           `json:"required"`                                           // 该字段是否必填
	Description  string         `json:"description" gorm:"size:16384"`                      // 字段描述
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// InterfaceParameter 接口参数（使用类型）
type InterfaceParameter struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	AppID        int64          `json:"app_id" gorm:"not null;index" validate:"required"`   // 应用ID 用于后续查询
	InterfaceID  int64          `json:"interface_id" gorm:"not null;index"`                 // 接口ID
	Name         string         `json:"name" gorm:"not null;size:255" validate:"required"`  // 类型名称
	Type         string         `json:"type" validate:"oneof=number string boolean custom"` // 类型
	Ref          *int64         `json:"ref"`                                                // 如果是 custom 类型，引用 CustomType.ID
	Location     string         `json:"location" validate:"oneof=query header body path"`   // 参数位置
	IsArray      bool           `json:"is_array"`                                           // 是否为数组类型
	Required     bool           `json:"required"`                                           // 添加必填标识
	Description  string         `json:"description" gorm:"size:16384"`
	DefaultValue *string        `json:"default_value"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
