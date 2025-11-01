package repositories

import (
	"context"
	"fmt"
	"strings"

	"mcp-adapter/backend/dto"
	"mcp-adapter/backend/models"
	"mcp-adapter/backend/services"

	"gorm.io/gorm"
)

// applicationRepository 应用仓储实现
type applicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 创建应用仓储
func NewApplicationRepository(db *gorm.DB) services.ApplicationRepository {
	return &applicationRepository{db: db}
}

// Create 创建应用
func (r *applicationRepository) Create(ctx context.Context, app *models.Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

// GetByID 根据ID获取应用
func (r *applicationRepository) GetByID(ctx context.Context, id uint) (*models.Application, error) {
	var app models.Application
	err := r.db.WithContext(ctx).First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetAll 获取所有应用（分页）
func (r *applicationRepository) GetAll(ctx context.Context, pagination *dto.PaginationRequest) ([]*models.Application, int64, error) {
	var apps []*models.Application
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Application{})

	// 搜索过滤
	if pagination.Search != "" {
		searchTerm := "%" + strings.ToLower(pagination.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pagination.Page - 1) * pagination.PageSize
	err := query.Offset(offset).Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&apps).Error

	return apps, total, err
}

// Update 更新应用
func (r *applicationRepository) Update(ctx context.Context, app *models.Application) error {
	return r.db.WithContext(ctx).Save(app).Error
}

// Delete 删除应用
func (r *applicationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除相关的接口数据
		if err := tx.Where("app_id = ?", id).Delete(&models.Interface{}).Error; err != nil {
			return fmt.Errorf("failed to delete related interfaces: %w", err)
		}

		// 删除应用
		if err := tx.Delete(&models.Application{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete application: %w", err)
		}

		return nil
	})
}

// GetInterfaceCount 获取应用下的接口数量
func (r *applicationRepository) GetInterfaceCount(ctx context.Context, appID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Interface{}).Where("app_id = ?", appID).Count(&count).Error
	return count, err
}