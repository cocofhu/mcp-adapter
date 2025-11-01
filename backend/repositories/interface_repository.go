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

// interfaceRepository 接口仓储实现
type interfaceRepository struct {
	db *gorm.DB
}

// NewInterfaceRepository 创建接口仓储
func NewInterfaceRepository(db *gorm.DB) services.InterfaceRepository {
	return &interfaceRepository{db: db}
}

// Create 创建接口
func (r *interfaceRepository) Create(ctx context.Context, iface *models.Interface) error {
	return r.db.WithContext(ctx).Create(iface).Error
}

// GetByID 根据ID获取接口
func (r *interfaceRepository) GetByID(ctx context.Context, id uint) (*models.Interface, error) {
	var iface models.Interface
	err := r.db.WithContext(ctx).First(&iface, id).Error
	if err != nil {
		return nil, err
	}
	return &iface, nil
}

// GetWithRelations 获取接口及其关联数据
func (r *interfaceRepository) GetWithRelations(ctx context.Context, id uint) (*models.Interface, error) {
	var iface models.Interface
	err := r.db.WithContext(ctx).
		Preload("Parameters").
		Preload("DefaultParams").
		Preload("DefaultHeaders").
		First(&iface, id).Error
	if err != nil {
		return nil, err
	}
	return &iface, nil
}

// GetByAppID 根据应用ID获取接口列表（分页）
func (r *interfaceRepository) GetByAppID(ctx context.Context, appID uint, pagination *dto.PaginationRequest) ([]*models.Interface, int64, error) {
	var interfaces []*models.Interface
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Interface{}).Where("app_id = ?", appID)

	// 搜索过滤
	if pagination.Search != "" {
		searchTerm := "%" + strings.ToLower(pagination.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(protocol) LIKE ?", 
			searchTerm, searchTerm, searchTerm)
	}

	// 协议过滤
	if pagination.Protocol != "" {
		query = query.Where("protocol = ?", pagination.Protocol)
	}

	// 状态过滤
	if pagination.Status != "" {
		switch pagination.Status {
		case "enabled":
			query = query.Where("enabled = ?", true)
		case "disabled":
			query = query.Where("enabled = ?", false)
		default:
			query = query.Where("status = ?", pagination.Status)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pagination.Page - 1) * pagination.PageSize
	err := query.Offset(offset).Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&interfaces).Error

	return interfaces, total, err
}

// Update 更新接口
func (r *interfaceRepository) Update(ctx context.Context, iface *models.Interface) error {
	return r.db.WithContext(ctx).Save(iface).Error
}

// Delete 删除接口
func (r *interfaceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的参数数据
		if err := tx.Where("interface_id = ?", id).Delete(&models.Parameter{}).Error; err != nil {
			return fmt.Errorf("failed to delete related parameters: %w", err)
		}

		// 删除相关的默认参数数据
		if err := tx.Where("interface_id = ?", id).Delete(&models.DefaultParam{}).Error; err != nil {
			return fmt.Errorf("failed to delete related default params: %w", err)
		}

		// 删除相关的默认请求头数据
		if err := tx.Where("interface_id = ?", id).Delete(&models.DefaultHeader{}).Error; err != nil {
			return fmt.Errorf("failed to delete related default headers: %w", err)
		}

		// 删除接口
		if err := tx.Delete(&models.Interface{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete interface: %w", err)
		}

		return nil
	})
}

// BatchToggle 批量切换接口状态
func (r *interfaceRepository) BatchToggle(ctx context.Context, ids []uint, enabled bool) error {
	status := "active"
	if !enabled {
		status = "inactive"
	}

	return r.db.WithContext(ctx).
		Model(&models.Interface{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"enabled": enabled,
			"status":  status,
		}).Error
}

// BatchDelete 批量删除接口
func (r *interfaceRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的参数数据
		if err := tx.Where("interface_id IN ?", ids).Delete(&models.Parameter{}).Error; err != nil {
			return fmt.Errorf("failed to delete related parameters: %w", err)
		}

		// 删除相关的默认参数数据
		if err := tx.Where("interface_id IN ?", ids).Delete(&models.DefaultParam{}).Error; err != nil {
			return fmt.Errorf("failed to delete related default params: %w", err)
		}

		// 删除相关的默认请求头数据
		if err := tx.Where("interface_id IN ?", ids).Delete(&models.DefaultHeader{}).Error; err != nil {
			return fmt.Errorf("failed to delete related default headers: %w", err)
		}

		// 删除接口
		if err := tx.Where("id IN ?", ids).Delete(&models.Interface{}).Error; err != nil {
			return fmt.Errorf("failed to delete interfaces: %w", err)
		}

		return nil
	})
}