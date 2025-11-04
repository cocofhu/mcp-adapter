package service

import (
	"errors"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CreateApplicationRequest struct {
	Name        string `json:"name" validate:"required, max=128"`                        // 应用名称 不允许重复
	Description string `json:"description" validate:"max=16384"`                         // 应用描述
	Path        string `json:"path" validate:"required,regexp=^[a-zA-Z0-9_]+$, max=128"` // 应用路由标识
	Protocol    string `json:"protocol" validate:"required,oneof=sse"`                   // 应用暴露协议
	PostProcess string `json:"post_process" validate:"max=1048576"`                      // 应用后处理脚本
	Environment string `json:"environment" validate:"max=1048576"`                       // 应用环境变量
	Enabled     *bool  `json:"enabled,omitempty"`                                        // 是否启用应用
}

type GetApplicationRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListApplicationsRequest struct{}

type UpdateApplicationRequest struct {
	ID          int64   `json:"id" validate:"required,gt=0"`
	Name        *string `json:"name,omitempty" validate:"max=128"`
	Description *string `json:"description,omitempty" validate:"max=16384"`
	Path        *string `json:"path,omitempty" validate:"required,regexp=^[a-zA-Z0-9_]+$, max=128"`
	Protocol    *string `json:"protocol,omitempty" validate:"required,oneof=http"`
	PostProcess *string `json:"post_process,omitempty" validate:"max=1048576"`
	Environment *string `json:"environment,omitempty" validate:"max=1048576"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type DeleteApplicationRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ApplicationDTO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Protocol    string    `json:"protocol"`
	PostProcess string    `json:"post_process"`
	Environment string    `json:"environment"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApplicationResponse struct {
	Application ApplicationDTO `json:"application"`
}

type ApplicationsResponse struct {
	Applications []ApplicationDTO `json:"applications"`
}

type EmptyResponse struct{}

// mapper
func toApplicationDTO(m models.Application) ApplicationDTO {
	return ApplicationDTO{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Path:        m.Path,
		Protocol:    m.Protocol,
		PostProcess: m.PostProcess,
		Environment: m.Environment,
		Enabled:     m.Enabled,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// CreateApplication 创建应用
func CreateApplication(req CreateApplicationRequest) (ApplicationResponse, error) {
	if err := validate.Struct(req); err != nil {
		return ApplicationResponse{}, err
	}
	db := database.GetDB()
	app := models.Application{
		Name:        req.Name,
		Description: req.Description,
		Path:        req.Path,
		Protocol:    req.Protocol,
		PostProcess: req.PostProcess,
		Environment: req.Environment,
	}
	if req.Enabled != nil {
		app.Enabled = *req.Enabled
	}
	var count int64
	db.Model(&models.Application{}).Where("name = ?", app.Name).Count(&count)
	if count > 0 {
		return ApplicationResponse{}, errors.New("duplicate application name")
	}
	if err := db.Create(&app).Error; err != nil {
		return ApplicationResponse{}, err
	}
	return ApplicationResponse{Application: toApplicationDTO(app)}, nil
}

// GetApplication 获取单个应用
func GetApplication(req GetApplicationRequest) (ApplicationResponse, error) {
	if err := validate.Struct(req); err != nil {
		return ApplicationResponse{}, err
	}
	db := database.GetDB()
	var app models.Application
	if err := db.First(&app, req.ID).Error; err != nil {
		return ApplicationResponse{}, errors.New("no such application")
	}
	return ApplicationResponse{Application: toApplicationDTO(app)}, nil
}

// ListApplications 获取应用列表
func ListApplications(req ListApplicationsRequest) (ApplicationsResponse, error) {
	if err := validate.Struct(req); err != nil {
		return ApplicationsResponse{}, err
	}
	db := database.GetDB()
	var apps []models.Application
	if err := db.Find(&apps).Error; err != nil {
		return ApplicationsResponse{}, err
	}
	res := make([]ApplicationDTO, 0, len(apps))
	for _, a := range apps {
		res = append(res, toApplicationDTO(a))
	}
	return ApplicationsResponse{Applications: res}, nil
}

func UpdateApplication(req UpdateApplicationRequest) (ApplicationResponse, error) {
	if err := validate.Struct(req); err != nil {
		return ApplicationResponse{}, err
	}
	db := database.GetDB()
	var existing models.Application
	if err := db.First(&existing, req.ID).Error; err != nil {
		return ApplicationResponse{}, errors.New("no such application")
	}
	// apply changes
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Path != nil {
		existing.Path = *req.Path
	}
	if req.Protocol != nil {
		existing.Protocol = *req.Protocol
	}
	if req.PostProcess != nil {
		existing.PostProcess = *req.PostProcess
	}
	if req.Environment != nil {
		existing.Environment = *req.Environment
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	// unique name check
	newName := existing.Name
	var cnt int64
	db.Model(&models.Application{}).Where("name = ? AND id <> ?", newName, existing.ID).Count(&cnt)
	if cnt > 0 {
		return ApplicationResponse{}, errors.New("duplicate application name")
	}
	if err := db.Save(&existing).Error; err != nil {
		return ApplicationResponse{}, err
	}
	return ApplicationResponse{Application: toApplicationDTO(existing)}, nil
}

func DeleteApplication(req DeleteApplicationRequest) (EmptyResponse, error) {
	if err := validate.Struct(req); err != nil {
		return EmptyResponse{}, err
	}
	db := database.GetDB()
	var app models.Application
	if err := db.First(&app, req.ID).Error; err != nil {
		return EmptyResponse{}, errors.New("no such application")
	}
	// 如果存在关联接口，拒绝删除
	var count int64
	db.Model(&models.Interface{}).Where("app_id = ?", app.ID).Count(&count)
	if count > 0 {
		return EmptyResponse{}, errors.New("cannot delete application with associated interfaces")
	}
	if err := db.Delete(&app).Error; err != nil {
		return EmptyResponse{}, err
	}
	return EmptyResponse{}, nil
}
