package service

import (
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"
)

// Create DTO
type CreateInterfaceRequest struct {
	AppID       int64  `json:"app_id" validate:"required,gt=0"`
	Name        string `json:"name" validate:"required, max=128"`
	Description string `json:"description" validate:"max=1024"`
	Protocol    string `json:"protocol" validate:"required,oneof=sse streamable"`
	URL         string `json:"url" validate:"required, max=1024"`
	AuthType    string `json:"auth_type" validate:"required,oneof=none"`
	Enabled     *bool  `json:"enabled,omitempty"`
	PostProcess string `json:"post_process" validate:"max=1024"`
	Options     string `json:"options" validate:"max=16384"`
}

// Read DTO
type GetInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListInterfacesRequest struct {
	AppID *int64 `json:"app_id,omitempty" validate:"omitempty,gt=0"`
}

// Update DTO (partial)
type UpdateInterfaceRequest struct {
	ID          int64   `json:"id" validate:"required,gt=0"`
	AppID       *int64  `json:"app_id,omitempty" validate:"omitempty,gt=0"`
	Name        *string `json:"name,omitempty" validate:"max=128"`
	Description *string `json:"description,omitempty" validate:"max=1024"`
	Protocol    *string `json:"protocol,omitempty" validate:"required,oneof=sse streamable"`
	URL         *string `json:"url,omitempty" validate:"max=1024"`
	AuthType    *string `json:"auth_type,omitempty" validate:"required,oneof=none"`
	Enabled     *bool   `json:"enabled,omitempty"`
	PostProcess *string `json:"post_process,omitempty"`
	Options     *string `json:"options,omitempty" validate:"max=16384"`
}

// Delete
type DeleteInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

// Response DTOs
type InterfaceDTO struct {
	ID          int64     `json:"id"`
	AppID       int64     `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Protocol    string    `json:"protocol"`
	URL         string    `json:"url"`
	AuthType    string    `json:"auth_type"`
	Enabled     bool      `json:"enabled"`
	PostProcess string    `json:"post_process"`
	Options     string    `json:"options"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type InterfaceResponse struct {
	Interface InterfaceDTO `json:"interface"`
}

type InterfacesResponse struct {
	Interfaces []InterfaceDTO `json:"interfaces"`
}

// mapper
func toInterfaceDTO(m models.Interface) InterfaceDTO {
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
		Options:     m.Options,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// Services
func CreateInterface(req CreateInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, ErrValidation
	}
	db := database.GetDB()
	// check app exists
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return InterfaceResponse{}, ErrNotFound
	}
	// unique name within app
	var count int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ?", req.AppID, req.Name).Count(&count)
	if count > 0 {
		return InterfaceResponse{}, ErrIfaceNameExists
	}
	iface := models.Interface{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Protocol:    req.Protocol,
		URL:         req.URL,
		AuthType:    req.AuthType,
		PostProcess: req.PostProcess,
		Options:     req.Options,
	}
	if req.Enabled != nil {
		iface.Enabled = *req.Enabled
	}
	// validate options
	if opts, err := iface.GetToolOptions(); err != nil {
		return InterfaceResponse{}, ErrInvalidOptions
	} else if err = opts.Validate(); err != nil {
		return InterfaceResponse{}, ErrInvalidOptions
	}
	if err := db.Create(&iface).Error; err != nil {
		return InterfaceResponse{}, err
	}
	return InterfaceResponse{Interface: toInterfaceDTO(iface)}, nil
}

func GetInterface(req GetInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, ErrValidation
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return InterfaceResponse{}, ErrNotFound
	}
	return InterfaceResponse{Interface: toInterfaceDTO(iface)}, nil
}

func ListInterfaces(req ListInterfacesRequest) (InterfacesResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfacesResponse{}, ErrValidation
	}
	db := database.GetDB()
	var ifaces []models.Interface
	query := db
	if req.AppID != nil {
		query = query.Where("app_id = ?", *req.AppID)
	}
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
	if err := validate.Var(req.ID, "required,gt=0"); err != nil {
		return InterfaceResponse{}, ErrValidation
	}
	db := database.GetDB()
	var existing models.Interface
	if err := db.First(&existing, req.ID).Error; err != nil {
		return InterfaceResponse{}, ErrNotFound
	}
	// apply changes
	if req.AppID != nil && *req.AppID != existing.AppID {
		var app models.Application
		if err := db.First(&app, *req.AppID).Error; err != nil {
			return InterfaceResponse{}, ErrNotFound
		}
		existing.AppID = *req.AppID
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
		existing.Options = *req.Options
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	// validate options
	if opts, err := existing.GetToolOptions(); err != nil {
		return InterfaceResponse{}, ErrInvalidOptions
	} else if err = opts.Validate(); err != nil {
		return InterfaceResponse{}, ErrInvalidOptions
	}
	// unique name within app
	var cnt int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ? AND id <> ?", existing.AppID, existing.Name, existing.ID).Count(&cnt)
	if cnt > 0 {
		return InterfaceResponse{}, ErrIfaceNameExists
	}
	if err := db.Save(&existing).Error; err != nil {
		return InterfaceResponse{}, err
	}
	return InterfaceResponse{Interface: toInterfaceDTO(existing)}, nil
}

func DeleteInterface(req DeleteInterfaceRequest) (EmptyResponse, error) {
	if err := validate.Struct(req); err != nil {
		return EmptyResponse{}, ErrValidation
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return EmptyResponse{}, ErrNotFound
	}
	if err := db.Delete(&iface).Error; err != nil {
		return EmptyResponse{}, err
	}
	return EmptyResponse{}, nil
}
