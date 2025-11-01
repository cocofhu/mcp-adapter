package services

import (
	"context"
	"fmt"
	"math"

	"mcp-adapter/backend/dto"
	"mcp-adapter/backend/models"
)

// applicationService 应用服务实现
type applicationService struct {
	appRepo ApplicationRepository
}

// NewApplicationService 创建应用服务
func NewApplicationService(appRepo ApplicationRepository) ApplicationService {
	return &applicationService{
		appRepo: appRepo,
	}
}

// CreateApplication 创建应用
func (s *applicationService) CreateApplication(ctx context.Context, req *dto.CreateApplicationRequest) (*dto.ApplicationResponse, error) {
	// 创建应用模型
	app := &models.Application{
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		BaseURL:     req.BaseURL,
	}

	// 设置默认版本
	if app.Version == "" {
		app.Version = "v1.0.0"
	}

	// 保存到数据库
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// 转换为响应DTO
	return s.toApplicationResponse(ctx, app), nil
}

// GetApplications 获取应用列表
func (s *applicationService) GetApplications(ctx context.Context, pagination *dto.PaginationRequest) (*dto.PaginationResponse, error) {
	// 从仓储获取数据
	apps, total, err := s.appRepo.GetAll(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}

	// 转换为响应DTO
	var responses []dto.ApplicationResponse
	for _, app := range apps {
		responses = append(responses, *s.toApplicationResponse(ctx, app))
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))

	return &dto.PaginationResponse{
		Data:       responses,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetApplicationByID 根据ID获取应用
func (s *applicationService) GetApplicationByID(ctx context.Context, id uint) (*dto.ApplicationResponse, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return s.toApplicationResponse(ctx, app), nil
}

// UpdateApplication 更新应用
func (s *applicationService) UpdateApplication(ctx context.Context, id uint, req *dto.UpdateApplicationRequest) (*dto.ApplicationResponse, error) {
	// 获取现有应用
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// 更新字段
	app.Name = req.Name
	app.Description = req.Description
	app.Version = req.Version
	app.BaseURL = req.BaseURL

	// 保存更新
	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return s.toApplicationResponse(ctx, app), nil
}

// DeleteApplication 删除应用
func (s *applicationService) DeleteApplication(ctx context.Context, id uint) error {
	// 检查应用是否存在
	_, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("application not found: %w", err)
	}

	// 删除应用（级联删除相关接口）
	if err := s.appRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}

// GetApplicationStats 获取应用统计信息
func (s *applicationService) GetApplicationStats(ctx context.Context, id uint) (map[string]interface{}, error) {
	// 检查应用是否存在
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// 获取接口数量
	interfaceCount, err := s.appRepo.GetInterfaceCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface count: %w", err)
	}

	// TODO: 可以添加更多统计信息，如：
	// - 活跃接口数量
	// - 最近调用次数
	// - 错误率等

	stats := map[string]interface{}{
		"application_id":    app.ID,
		"application_name":  app.Name,
		"interface_count":   interfaceCount,
		"created_at":        app.CreatedAt,
		"updated_at":        app.UpdatedAt,
	}

	return stats, nil
}

// toApplicationResponse 转换为应用响应DTO
func (s *applicationService) toApplicationResponse(ctx context.Context, app *models.Application) *dto.ApplicationResponse {
	// 获取接口数量
	interfaceCount, _ := s.appRepo.GetInterfaceCount(ctx, app.ID)

	return &dto.ApplicationResponse{
		ID:             app.ID,
		Name:           app.Name,
		Description:    app.Description,
		Version:        app.Version,
		BaseURL:        app.BaseURL,
		InterfaceCount: interfaceCount,
		CreatedAt:      app.CreatedAt,
		UpdatedAt:      app.UpdatedAt,
	}
}