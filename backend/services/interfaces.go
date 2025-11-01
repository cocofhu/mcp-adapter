package services

import (
	"context"
	"mcp-adapter/backend/dto"
	"mcp-adapter/backend/models"
)

// ApplicationService 应用服务接口
type ApplicationService interface {
	// 创建应用
	CreateApplication(ctx context.Context, req *dto.CreateApplicationRequest) (*dto.ApplicationResponse, error)
	
	// 获取应用列表
	GetApplications(ctx context.Context, pagination *dto.PaginationRequest) (*dto.PaginationResponse, error)
	
	// 根据ID获取应用
	GetApplicationByID(ctx context.Context, id uint) (*dto.ApplicationResponse, error)
	
	// 更新应用
	UpdateApplication(ctx context.Context, id uint, req *dto.UpdateApplicationRequest) (*dto.ApplicationResponse, error)
	
	// 删除应用
	DeleteApplication(ctx context.Context, id uint) error
	
	// 获取应用统计信息
	GetApplicationStats(ctx context.Context, id uint) (map[string]interface{}, error)
}

// InterfaceService 接口服务接口
type InterfaceService interface {
	// 创建接口
	CreateInterface(ctx context.Context, req *dto.CreateInterfaceRequest) (*dto.InterfaceResponse, error)
	
	// 获取接口列表（按应用ID过滤）
	GetInterfaces(ctx context.Context, appID uint, pagination *dto.PaginationRequest) (*dto.PaginationResponse, error)
	
	// 根据ID获取接口
	GetInterfaceByID(ctx context.Context, id uint) (*dto.InterfaceResponse, error)
	
	// 更新接口
	UpdateInterface(ctx context.Context, id uint, req *dto.UpdateInterfaceRequest) (*dto.InterfaceResponse, error)
	
	// 删除接口
	DeleteInterface(ctx context.Context, id uint) error
	
	// 切换接口启用状态
	ToggleInterface(ctx context.Context, id uint, req *dto.ToggleInterfaceRequest) (*dto.InterfaceResponse, error)
	
	// 测试接口
	TestInterface(ctx context.Context, id uint, req *dto.TestInterfaceRequest) (*dto.TestInterfaceResponse, error)
	
	// 批量操作接口
	BatchToggleInterfaces(ctx context.Context, ids []uint, enabled bool) error
	BatchDeleteInterfaces(ctx context.Context, ids []uint) error
}

// ParameterService 参数服务接口
type ParameterService interface {
	// 创建参数
	CreateParameter(ctx context.Context, interfaceID uint, req *dto.CreateParameterRequest) (*dto.ParameterResponse, error)
	
	// 获取接口的所有参数
	GetParametersByInterfaceID(ctx context.Context, interfaceID uint) ([]dto.ParameterResponse, error)
	
	// 更新参数
	UpdateParameter(ctx context.Context, id uint, req *dto.CreateParameterRequest) (*dto.ParameterResponse, error)
	
	// 删除参数
	DeleteParameter(ctx context.Context, id uint) error
	
	// 批量创建参数
	BatchCreateParameters(ctx context.Context, interfaceID uint, params []dto.CreateParameterRequest) ([]dto.ParameterResponse, error)
	
	// 批量更新参数（先删除旧的，再创建新的）
	BatchUpdateParameters(ctx context.Context, interfaceID uint, params []dto.CreateParameterRequest) ([]dto.ParameterResponse, error)
}

// DefaultParamService 默认参数服务接口
type DefaultParamService interface {
	// 创建默认参数
	CreateDefaultParam(ctx context.Context, interfaceID uint, req *dto.CreateDefaultParamRequest) (*dto.DefaultParamResponse, error)
	
	// 获取接口的所有默认参数
	GetDefaultParamsByInterfaceID(ctx context.Context, interfaceID uint) ([]dto.DefaultParamResponse, error)
	
	// 更新默认参数
	UpdateDefaultParam(ctx context.Context, id uint, req *dto.CreateDefaultParamRequest) (*dto.DefaultParamResponse, error)
	
	// 删除默认参数
	DeleteDefaultParam(ctx context.Context, id uint) error
	
	// 批量创建默认参数
	BatchCreateDefaultParams(ctx context.Context, interfaceID uint, params []dto.CreateDefaultParamRequest) ([]dto.DefaultParamResponse, error)
	
	// 批量更新默认参数
	BatchUpdateDefaultParams(ctx context.Context, interfaceID uint, params []dto.CreateDefaultParamRequest) ([]dto.DefaultParamResponse, error)
}

// DefaultHeaderService 默认请求头服务接口
type DefaultHeaderService interface {
	// 创建默认请求头
	CreateDefaultHeader(ctx context.Context, interfaceID uint, req *dto.CreateDefaultHeaderRequest) (*dto.DefaultHeaderResponse, error)
	
	// 获取接口的所有默认请求头
	GetDefaultHeadersByInterfaceID(ctx context.Context, interfaceID uint) ([]dto.DefaultHeaderResponse, error)
	
	// 更新默认请求头
	UpdateDefaultHeader(ctx context.Context, id uint, req *dto.CreateDefaultHeaderRequest) (*dto.DefaultHeaderResponse, error)
	
	// 删除默认请求头
	DeleteDefaultHeader(ctx context.Context, id uint) error
	
	// 批量创建默认请求头
	BatchCreateDefaultHeaders(ctx context.Context, interfaceID uint, headers []dto.CreateDefaultHeaderRequest) ([]dto.DefaultHeaderResponse, error)
	
	// 批量更新默认请求头
	BatchUpdateDefaultHeaders(ctx context.Context, interfaceID uint, headers []dto.CreateDefaultHeaderRequest) ([]dto.DefaultHeaderResponse, error)
}

// Repository 仓储层接口
type ApplicationRepository interface {
	Create(ctx context.Context, app *models.Application) error
	GetByID(ctx context.Context, id uint) (*models.Application, error)
	GetAll(ctx context.Context, pagination *dto.PaginationRequest) ([]*models.Application, int64, error)
	Update(ctx context.Context, app *models.Application) error
	Delete(ctx context.Context, id uint) error
	GetInterfaceCount(ctx context.Context, appID uint) (int64, error)
}

type InterfaceRepository interface {
	Create(ctx context.Context, iface *models.Interface) error
	GetByID(ctx context.Context, id uint) (*models.Interface, error)
	GetByAppID(ctx context.Context, appID uint, pagination *dto.PaginationRequest) ([]*models.Interface, int64, error)
	Update(ctx context.Context, iface *models.Interface) error
	Delete(ctx context.Context, id uint) error
	BatchToggle(ctx context.Context, ids []uint, enabled bool) error
	BatchDelete(ctx context.Context, ids []uint) error
	GetWithRelations(ctx context.Context, id uint) (*models.Interface, error)
}

type ParameterRepository interface {
	Create(ctx context.Context, param *models.Parameter) error
	GetByID(ctx context.Context, id uint) (*models.Parameter, error)
	GetByInterfaceID(ctx context.Context, interfaceID uint) ([]*models.Parameter, error)
	Update(ctx context.Context, param *models.Parameter) error
	Delete(ctx context.Context, id uint) error
	DeleteByInterfaceID(ctx context.Context, interfaceID uint) error
	BatchCreate(ctx context.Context, params []*models.Parameter) error
}

type DefaultParamRepository interface {
	Create(ctx context.Context, param *models.DefaultParam) error
	GetByID(ctx context.Context, id uint) (*models.DefaultParam, error)
	GetByInterfaceID(ctx context.Context, interfaceID uint) ([]*models.DefaultParam, error)
	Update(ctx context.Context, param *models.DefaultParam) error
	Delete(ctx context.Context, id uint) error
	DeleteByInterfaceID(ctx context.Context, interfaceID uint) error
	BatchCreate(ctx context.Context, params []*models.DefaultParam) error
}

type DefaultHeaderRepository interface {
	Create(ctx context.Context, header *models.DefaultHeader) error
	GetByID(ctx context.Context, id uint) (*models.DefaultHeader, error)
	GetByInterfaceID(ctx context.Context, interfaceID uint) ([]*models.DefaultHeader, error)
	Update(ctx context.Context, header *models.DefaultHeader) error
	Delete(ctx context.Context, id uint) error
	DeleteByInterfaceID(ctx context.Context, interfaceID uint) error
	BatchCreate(ctx context.Context, headers []*models.DefaultHeader) error
}

// HTTPClient HTTP客户端接口，用于测试接口
type HTTPClient interface {
	// 发送HTTP请求
	SendRequest(ctx context.Context, method, url string, headers map[string]string, body interface{}) (*dto.TestInterfaceResponse, error)
	
	// 构建请求URL（处理查询参数）
	BuildURL(baseURL string, params map[string]interface{}) (string, error)
	
	// 构建请求体
	BuildRequestBody(params map[string]interface{}, contentType string) (interface{}, error)
}