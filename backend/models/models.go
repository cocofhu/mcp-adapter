package models

import (
	"time"
	"gorm.io/gorm"
)

// Application 应用实体
type Application struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;size:100" validate:"required"`
	Description string    `json:"description" gorm:"size:500"`
	Version     string    `json:"version" gorm:"size:20;default:v1.0.0"`
	BaseURL     string    `json:"base_url" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联关系
	Interfaces []Interface `json:"interfaces,omitempty" gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"`
}

// Interface 接口实体
type Interface struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	AppID               uint      `json:"app_id" gorm:"not null;index" validate:"required"`
	Name                string    `json:"name" gorm:"not null;size:100" validate:"required"`
	Description         string    `json:"description" gorm:"size:500"`
	Protocol            string    `json:"protocol" gorm:"not null;size:20;default:http" validate:"required,oneof=http https"`
	Method              string    `json:"method" gorm:"not null;size:10;default:GET" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	URL                 string    `json:"url" gorm:"not null;size:500" validate:"required,url"`
	AuthType            string    `json:"auth_type" gorm:"size:20;default:none" validate:"oneof=none bearer basic api-key"`
	AuthValue           string    `json:"auth_value" gorm:"size:500"`
	Status              string    `json:"status" gorm:"size:20;default:active" validate:"oneof=active inactive error"`
	Enabled             bool      `json:"enabled" gorm:"default:true"`
	HTTPParamLocation   string    `json:"http_param_location" gorm:"size:20;default:query" validate:"oneof=query body"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	
	// 关联关系
	Application     Application     `json:"application,omitempty" gorm:"foreignKey:AppID"`
	Parameters      []Parameter     `json:"parameters,omitempty" gorm:"foreignKey:InterfaceID;constraint:OnDelete:CASCADE"`
	DefaultParams   []DefaultParam  `json:"default_params,omitempty" gorm:"foreignKey:InterfaceID;constraint:OnDelete:CASCADE"`
	DefaultHeaders  []DefaultHeader `json:"default_headers,omitempty" gorm:"foreignKey:InterfaceID;constraint:OnDelete:CASCADE"`
}

// Parameter 请求参数实体
type Parameter struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	InterfaceID uint   `json:"interface_id" gorm:"not null;index" validate:"required"`
	Name        string `json:"name" gorm:"not null;size:100" validate:"required"`
	Type        string `json:"type" gorm:"not null;size:20;default:string" validate:"required,oneof=string integer number boolean object array"`
	Location    string `json:"location" gorm:"not null;size:20;default:query" validate:"required,oneof=query body header path"`
	Required    bool   `json:"required" gorm:"default:false"`
	Description string `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联关系
	Interface Interface `json:"interface,omitempty" gorm:"foreignKey:InterfaceID"`
}

// DefaultParam 默认参数实体
type DefaultParam struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	InterfaceID uint   `json:"interface_id" gorm:"not null;index" validate:"required"`
	Name        string `json:"name" gorm:"not null;size:100" validate:"required"`
	Value       string `json:"value" gorm:"not null;size:1000" validate:"required"`
	Description string `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联关系
	Interface Interface `json:"interface,omitempty" gorm:"foreignKey:InterfaceID"`
}

// DefaultHeader 默认请求头实体
type DefaultHeader struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	InterfaceID uint   `json:"interface_id" gorm:"not null;index" validate:"required"`
	Name        string `json:"name" gorm:"not null;size:100" validate:"required"`
	Value       string `json:"value" gorm:"not null;size:1000" validate:"required"`
	Description string `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联关系
	Interface Interface `json:"interface,omitempty" gorm:"foreignKey:InterfaceID"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

func (Interface) TableName() string {
	return "interfaces"
}

func (Parameter) TableName() string {
	return "parameters"
}

func (DefaultParam) TableName() string {
	return "default_params"
}

func (DefaultHeader) TableName() string {
	return "default_headers"
}

// BeforeCreate 创建前钩子
func (app *Application) BeforeCreate(tx *gorm.DB) error {
	if app.Version == "" {
		app.Version = "v1.0.0"
	}
	return nil
}

func (iface *Interface) BeforeCreate(tx *gorm.DB) error {
	if iface.Protocol == "" {
		iface.Protocol = "http"
	}
	if iface.Method == "" {
		iface.Method = "GET"
	}
	if iface.Status == "" {
		iface.Status = "active"
	}
	if iface.HTTPParamLocation == "" {
		iface.HTTPParamLocation = "query"
	}
	return nil
}

func (param *Parameter) BeforeCreate(tx *gorm.DB) error {
	if param.Type == "" {
		param.Type = "string"
	}
	if param.Location == "" {
		param.Location = "query"
	}
	return nil
}