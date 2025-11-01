package dto

import "time"

// CreateApplicationRequest 创建应用请求
type CreateApplicationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Version     string `json:"version" validate:"max=20"`
	BaseURL     string `json:"base_url" validate:"omitempty,url,max=255"`
}

// UpdateApplicationRequest 更新应用请求
type UpdateApplicationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Version     string `json:"version" validate:"max=20"`
	BaseURL     string `json:"base_url" validate:"omitempty,url,max=255"`
}

// ApplicationResponse 应用响应
type ApplicationResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Version        string    `json:"version"`
	BaseURL        string    `json:"base_url"`
	InterfaceCount int64     `json:"interface_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateInterfaceRequest 创建接口请求
type CreateInterfaceRequest struct {
	AppID               uint                      `json:"app_id" validate:"required"`
	Name                string                    `json:"name" validate:"required,min=1,max=100"`
	Description         string                    `json:"description" validate:"max=500"`
	Protocol            string                    `json:"protocol" validate:"required,oneof=http https"`
	Method              string                    `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	URL                 string                    `json:"url" validate:"required,url,max=500"`
	AuthType            string                    `json:"auth_type" validate:"oneof=none bearer basic api-key"`
	AuthValue           string                    `json:"auth_value" validate:"max=500"`
	HTTPParamLocation   string                    `json:"http_param_location" validate:"oneof=query body"`
	Parameters          []CreateParameterRequest  `json:"parameters"`
	DefaultParams       []CreateDefaultParamRequest `json:"default_params"`
	DefaultHeaders      []CreateDefaultHeaderRequest `json:"default_headers"`
}

// UpdateInterfaceRequest 更新接口请求
type UpdateInterfaceRequest struct {
	Name                string                      `json:"name" validate:"required,min=1,max=100"`
	Description         string                      `json:"description" validate:"max=500"`
	Protocol            string                      `json:"protocol" validate:"required,oneof=http https"`
	Method              string                      `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	URL                 string                      `json:"url" validate:"required,url,max=500"`
	AuthType            string                      `json:"auth_type" validate:"oneof=none bearer basic api-key"`
	AuthValue           string                      `json:"auth_value" validate:"max=500"`
	Status              string                      `json:"status" validate:"oneof=active inactive error"`
	Enabled             *bool                       `json:"enabled"`
	HTTPParamLocation   string                      `json:"http_param_location" validate:"oneof=query body"`
	Parameters          []CreateParameterRequest    `json:"parameters"`
	DefaultParams       []CreateDefaultParamRequest `json:"default_params"`
	DefaultHeaders      []CreateDefaultHeaderRequest `json:"default_headers"`
}

// InterfaceResponse 接口响应
type InterfaceResponse struct {
	ID                  uint                    `json:"id"`
	AppID               uint                    `json:"app_id"`
	Name                string                  `json:"name"`
	Description         string                  `json:"description"`
	Protocol            string                  `json:"protocol"`
	Method              string                  `json:"method"`
	URL                 string                  `json:"url"`
	AuthType            string                  `json:"auth_type"`
	AuthValue           string                  `json:"auth_value,omitempty"` // 敏感信息可选择不返回
	Status              string                  `json:"status"`
	Enabled             bool                    `json:"enabled"`
	HTTPParamLocation   string                  `json:"http_param_location"`
	Parameters          []ParameterResponse     `json:"parameters"`
	DefaultParams       []DefaultParamResponse  `json:"default_params"`
	DefaultHeaders      []DefaultHeaderResponse `json:"default_headers"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

// CreateParameterRequest 创建参数请求
type CreateParameterRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Type        string `json:"type" validate:"required,oneof=string integer number boolean object array"`
	Location    string `json:"location" validate:"required,oneof=query body header path"`
	Required    bool   `json:"required"`
	Description string `json:"description" validate:"max=500"`
}

// ParameterResponse 参数响应
type ParameterResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Location    string `json:"location"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// CreateDefaultParamRequest 创建默认参数请求
type CreateDefaultParamRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Value       string `json:"value" validate:"required,max=1000"`
	Description string `json:"description" validate:"max=500"`
}

// DefaultParamResponse 默认参数响应
type DefaultParamResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// CreateDefaultHeaderRequest 创建默认请求头请求
type CreateDefaultHeaderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Value       string `json:"value" validate:"required,max=1000"`
	Description string `json:"description" validate:"max=500"`
}

// DefaultHeaderResponse 默认请求头响应
type DefaultHeaderResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// ToggleInterfaceRequest 切换接口状态请求
type ToggleInterfaceRequest struct {
	Enabled bool `json:"enabled"`
}

// TestInterfaceRequest 测试接口请求
type TestInterfaceRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
	Headers    map[string]string      `json:"headers"`
}

// TestInterfaceResponse 测试接口响应
type TestInterfaceResponse struct {
	Success      bool                   `json:"success"`
	StatusCode   int                    `json:"status_code"`
	ResponseTime int64                  `json:"response_time"` // 毫秒
	Headers      map[string]string      `json:"headers"`
	Body         interface{}            `json:"body"`
	Error        string                 `json:"error,omitempty"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int    `json:"page" form:"page" validate:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" validate:"min=1,max=100"`
	Search   string `json:"search" form:"search"`
	Status   string `json:"status" form:"status" validate:"omitempty,oneof=active inactive error enabled disabled"`
	Protocol string `json:"protocol" form:"protocol" validate:"omitempty,oneof=http https"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}