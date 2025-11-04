package service

import (
	"encoding/json"
	"errors"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"
)

type CreateInterfaceRequest struct {
	AppID       int64              `json:"app_id" validate:"required,gt=0"`
	Name        string             `json:"name,omitempty" validate:"required,max=255" `
	Description string             `json:"description,omitempty" validate:"max=16384"`
	Protocol    string             `json:"protocol,omitempty" validate:"required,oneof=http"`
	URL         string             `json:"url" validate:"required,max=1024"`
	AuthType    string             `json:"auth_type" validate:"required,oneof=none"`
	Enabled     bool               `json:"enabled,omitempty"`
	PostProcess string             `json:"post_process" validate:"max=1048576"`
	Options     models.ToolOptions `json:"options" validate:"required"`
}

type GetInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListInterfacesRequest struct {
	AppID int64 `json:"app_id" validate:"required,gt=0"`
}
type UpdateInterfaceRequest struct {
	ID          int64               `json:"id" validate:"required,gt=0"`
	Name        *string             `json:"name,omitempty" validate:"required,max=255" `
	Description *string             `json:"description,omitempty" validate:"max=16384"`
	Protocol    *string             `json:"protocol,omitempty" validate:"required,oneof=http"`
	URL         *string             `json:"url" validate:"required,max=1024"`
	AuthType    *string             `json:"auth_type" validate:"required,oneof=none"`
	Enabled     *bool               `json:"enabled,omitempty"`
	PostProcess *string             `json:"post_process" validate:"max=1048576"`
	Options     *models.ToolOptions `json:"options" validate:"required"`
}

type DeleteInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type InterfaceDTO struct {
	ID          int64              `json:"id"`
	AppID       int64              `json:"app_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Protocol    string             `json:"protocol"`
	URL         string             `json:"url"`
	AuthType    string             `json:"auth_type"`
	Enabled     bool               `json:"enabled"`
	PostProcess string             `json:"post_process"`
	Options     models.ToolOptions `json:"options"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type InterfaceResponse struct {
	Interface InterfaceDTO `json:"interface"`
}

type InterfacesResponse struct {
	Interfaces []InterfaceDTO `json:"interfaces"`
}

func toInterfaceDTO(m models.Interface) InterfaceDTO {
	// 此处不应该出错，上游需要保证options合法
	options, _ := m.GetToolOptions()
	return InterfaceDTO{
		ID:          m.ID,
		AppID:       m.AppID,
		Name:        m.Name,
		Description: m.Description,
		Protocol:    m.Protocol,
		URL:         m.URL,
		AuthType:    m.AuthType,
		Enabled:     m.Enabled,
		PostProcess: m.PostProcess,
		Options:     options,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func CreateInterface(req CreateInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, err
	}
	db := database.GetDB()
	// 检查应用是否存在
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return InterfaceResponse{}, errors.New("no such application")
	}
	// 接口名字在应用内唯一
	var count int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ?", req.AppID, req.Name).Count(&count)
	if count > 0 {
		return InterfaceResponse{}, errors.New("interface name already exists in the application")
	}
	// 将option转换JSONString
	opt, err := json.Marshal(req.Options)
	if err != nil {
		return InterfaceResponse{}, errors.New("interface options is invalid")
	}
	if len(opt) > 1048576 {
		return InterfaceResponse{}, errors.New("interface options exceed max size")
	}
	iface := models.Interface{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Protocol:    req.Protocol,
		URL:         req.URL,
		AuthType:    req.AuthType,
		PostProcess: req.PostProcess,
		Options:     string(opt),
		Enabled:     req.Enabled,
	}
	if err := db.Create(&iface).Error; err != nil {
		return InterfaceResponse{}, err
	}
	return InterfaceResponse{Interface: toInterfaceDTO(iface)}, nil
}

func GetInterface(req GetInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, err
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return InterfaceResponse{}, errors.New("no such interface")
	}
	return InterfaceResponse{Interface: toInterfaceDTO(iface)}, nil
}

func ListInterfaces(req ListInterfacesRequest) (InterfacesResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfacesResponse{}, err
	}
	db := database.GetDB()
	// 查数据库判断应用是否存在
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return InterfacesResponse{}, errors.New("no such application")
	}
	var ifaces []models.Interface
	query := db.Where("app_id = ?", req.AppID)
	if err := query.Find(&ifaces).Error; err != nil {
		return InterfacesResponse{}, err
	}
	res := make([]InterfaceDTO, 0, len(ifaces))
	for _, it := range ifaces {
		res = append(res, toInterfaceDTO(it))
	}
	return InterfacesResponse{Interfaces: res}, nil
}

func UpdateInterface(req UpdateInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, err
	}
	db := database.GetDB()
	var existing models.Interface
	if err := db.First(&existing, req.ID).Error; err != nil {
		return InterfaceResponse{}, errors.New("no such interface")
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Protocol != nil {
		existing.Protocol = *req.Protocol
	}
	if req.URL != nil {
		existing.URL = *req.URL
	}
	if req.AuthType != nil {
		existing.AuthType = *req.AuthType
	}
	if req.PostProcess != nil {
		existing.PostProcess = *req.PostProcess
	}
	if req.Options != nil {
		opt, err := json.Marshal(req.Options)
		if err != nil {
			return InterfaceResponse{}, errors.New("interface options is invalid")
		}
		if len(opt) > 1048576 {
			return InterfaceResponse{}, errors.New("interface options exceed max size")
		}
		existing.Options = string(opt)
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	var cnt int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ? AND id <> ?", existing.AppID, existing.Name, existing.ID).Count(&cnt)
	if cnt > 0 {
		return InterfaceResponse{}, errors.New("interface name already exists in the application")
	}
	if err := db.Save(&existing).Error; err != nil {
		return InterfaceResponse{}, err
	}
	return InterfaceResponse{Interface: toInterfaceDTO(existing)}, nil
}

func DeleteInterface(req DeleteInterfaceRequest) (EmptyResponse, error) {
	if err := validate.Struct(req); err != nil {
		return EmptyResponse{}, err
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return EmptyResponse{}, errors.New("no such interface")
	}
	if err := db.Delete(&iface).Error; err != nil {
		return EmptyResponse{}, err
	}
	return EmptyResponse{}, nil
}
