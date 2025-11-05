package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
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
	AuthType    string         `json:"auth_type"`                                         // 鉴权类型
	Enabled     bool           `json:"enabled" gorm:"default:true"`                       // 是否启用
	PostProcess string         `json:"post_process" gorm:"size:1048576"`                  // 后处理脚本
	Options     string         `json:"options" gorm:"size:500"`                           // 选择配置(JSON String), 包含Method, Parameter等参数
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// CustomType 自定义类型实体
type CustomType struct {
	ID          int64          `json:"id" gorm:"primaryKey"`                               //类型ID
	AppID       int64          `json:"app_id" gorm:"not null;index" validate:"required"`   // 应用ID 一个应用对应多个CustomType
	Name        string         `json:"name" validate:"required"`                           // 类型名称
	Type        string         `json:"type" validate:"oneof=number string boolean custom"` // 类型 custom表示复杂类型
	Description string         `json:"description" gorm:"size:16384"`                      // 类型说明 该参数会传入LLM
	IsArray     bool           `json:"isArray"`                                            // 是否是数组
	Required    bool           `json:"required"`                                           // 是否必须参数
	Ref         *int64         `json:"ref"`                                                // 如果是复杂类型, 则引用CustomType.ID
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ToolParameter struct {
	Name     string     `json:"name" validate:"required"`
	Type     CustomType `json:"type"`
	Location string     `json:"location" validate:"oneof=query header body"`
	Value    *string    `json:"value"`
}
type ToolOptions struct {
	Method            string          `json:"method" validate:"oneof=GET POST PUT PATCH DELETE"`
	Parameters        []ToolParameter `json:"parameters"`
	DefaultParameters []ToolParameter `json:"defaultParams"`
}

func (iface *Interface) GetToolOptions() (ToolOptions, error) {
	var spec ToolOptions
	err := json.Unmarshal([]byte(iface.Options), &spec)
	if err != nil {
		return spec, err
	}
	return spec, nil
}

func (to *ToolOptions) Validate() error {
	if to == nil {
		return errors.New("tool options is nil")
	}
	validate := validator.New()
	return validate.Struct(to)
}
