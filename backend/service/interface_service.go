package service

import (
	"errors"
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"
)

type CreateInterfaceRequest struct {
	AppID       int64                         `json:"app_id" validate:"required,gt=0"`                                         // 所属应用 ID
	Name        string                        `json:"name" validate:"required,max=255"`                                        // 接口名称
	Description string                        `json:"description" validate:"max=16384"`                                        // 接口描述
	Protocol    string                        `json:"protocol" validate:"required,oneof=http"`                                 // 协议类型
	URL         string                        `json:"url" validate:"required,max=1024"`                                        // 接口 URL
	Method      string                        `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"` // HTTP 方法
	AuthType    string                        `json:"auth_type" validate:"required,oneof=none"`                                // 鉴权类型
	Enabled     bool                          `json:"enabled"`                                                                 // 是否启用
	PostProcess string                        `json:"post_process" validate:"max=1048576"`                                     // 后置处理脚本
	Parameters  []CreateInterfaceParameterReq `json:"parameters"`                                                              // 接口参数列表
}

type CreateInterfaceParameterReq struct {
	Name         string  `json:"name" validate:"required,max=255"`                            // 参数名称
	Type         string  `json:"type" validate:"required,oneof=number string boolean custom"` // 参数类型
	Ref          *int64  `json:"ref"`                                                         // 如果 type=custom，引用 CustomType.ID
	Location     string  `json:"location" validate:"required,oneof=query header body path"`   // 参数位置
	IsArray      bool    `json:"is_array"`                                                    // 是否为数组
	Required     bool    `json:"required"`                                                    // 是否必填
	Description  string  `json:"description" validate:"max=16384"`                            // 参数描述
	DefaultValue *string `json:"default_value"`                                               // 默认值
	Group        string  `json:"group" validate:"required,oneof=input output fixed"`          // 参数组: input-输入参数, output-输出参数, fixed-固定参数
}

type GetInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListInterfacesRequest struct {
	AppID int64 `json:"app_id" validate:"required,gt=0"`
}
type UpdateInterfaceRequest struct {
	ID          int64                          `json:"id" validate:"required,gt=0"`                                                        // 要更新的接口 ID
	Name        *string                        `json:"name,omitempty" validate:"omitempty,max=255"`                                        // 接口名称
	Description *string                        `json:"description,omitempty" validate:"omitempty,max=16384"`                               // 接口描述
	Protocol    *string                        `json:"protocol,omitempty" validate:"omitempty,oneof=http"`                                 // 协议类型
	URL         *string                        `json:"url,omitempty" validate:"omitempty,max=1024"`                                        // 接口 URL
	Method      *string                        `json:"method,omitempty" validate:"omitempty,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"` // HTTP 方法
	AuthType    *string                        `json:"auth_type,omitempty" validate:"omitempty,oneof=none"`                                // 鉴权类型
	Enabled     *bool                          `json:"enabled,omitempty"`                                                                  // 是否启用
	PostProcess *string                        `json:"post_process,omitempty" validate:"omitempty,max=1048576"`                            // 后置处理脚本
	Parameters  *[]CreateInterfaceParameterReq `json:"parameters,omitempty"`                                                               // 如果提供，则完全替换参数列表
}

type DeleteInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type InterfaceParameterDTO struct {
	ID           int64     `json:"id"`
	InterfaceID  int64     `json:"interface_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Ref          *int64    `json:"ref"`
	Location     string    `json:"location"`
	IsArray      bool      `json:"is_array"`
	Required     bool      `json:"required"`
	Description  string    `json:"description"`
	DefaultValue *string   `json:"default_value"`
	Group        string    `json:"group"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InterfaceDTO struct {
	ID          int64                   `json:"id"`
	AppID       int64                   `json:"app_id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Protocol    string                  `json:"protocol"`
	URL         string                  `json:"url"`
	Method      string                  `json:"method"`
	AuthType    string                  `json:"auth_type"`
	Enabled     bool                    `json:"enabled"`
	PostProcess string                  `json:"post_process"`
	Parameters  []InterfaceParameterDTO `json:"parameters"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type InterfaceResponse struct {
	Interface InterfaceDTO `json:"interface"`
}

type InterfacesResponse struct {
	Interfaces []InterfaceDTO `json:"interfaces"`
}

func toInterfaceParameterDTO(m models.InterfaceParameter) InterfaceParameterDTO {
	return InterfaceParameterDTO{
		ID:           m.ID,
		InterfaceID:  m.InterfaceID,
		Name:         m.Name,
		Type:         m.Type,
		Ref:          m.Ref,
		Location:     m.Location,
		IsArray:      m.IsArray,
		Required:     m.Required,
		Description:  m.Description,
		DefaultValue: m.DefaultValue,
		Group:        m.Group,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toInterfaceDTO(m models.Interface, params []models.InterfaceParameter) InterfaceDTO {
	paramDTOs := make([]InterfaceParameterDTO, 0, len(params))
	for _, p := range params {
		paramDTOs = append(paramDTOs, toInterfaceParameterDTO(p))
	}
	return InterfaceDTO{
		ID:          m.ID,
		AppID:       m.AppID,
		Name:        m.Name,
		Description: m.Description,
		Protocol:    m.Protocol,
		URL:         m.URL,
		Method:      m.Method,
		AuthType:    m.AuthType,
		Enabled:     m.Enabled,
		PostProcess: m.PostProcess,
		Parameters:  paramDTOs,
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
		return InterfaceResponse{}, errors.New("application not found")
	}

	// 接口名字在应用内唯一
	var count int64
	db.Model(&models.Interface{}).Where("app_id = ? AND name = ?", req.AppID, req.Name).Count(&count)
	if count > 0 {
		return InterfaceResponse{}, errors.New("interface name already exists in the application")
	}

	// 验证参数的 Ref 引用
	for _, param := range req.Parameters {
		if param.Type == "custom" && param.Ref != nil {
			var refType models.CustomType
			if err := db.First(&refType, *param.Ref).Error; err != nil {
				return InterfaceResponse{}, errors.New("invalid parameter reference: custom type not found")
			}
			if refType.AppID != req.AppID {
				return InterfaceResponse{}, errors.New("parameter reference must belong to the same application")
			}
		}
	}

	// 创建接口
	iface := models.Interface{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Protocol:    req.Protocol,
		URL:         req.URL,
		Method:      req.Method,
		AuthType:    req.AuthType,
		PostProcess: req.PostProcess,
		Enabled:     req.Enabled,
	}

	// 使用事务
	tx := db.Begin()
	if err := tx.Create(&iface).Error; err != nil {
		tx.Rollback()
		return InterfaceResponse{}, err
	}

	// 创建参数
	params := make([]models.InterfaceParameter, 0, len(req.Parameters))
	for _, paramReq := range req.Parameters {
		param := models.InterfaceParameter{
			AppID:       req.AppID,
			InterfaceID: iface.ID,
			Name:        paramReq.Name,
			Type:        paramReq.Type,
			Ref:         paramReq.Ref,
			Location:    paramReq.Location,
			IsArray:     paramReq.IsArray,
			Required:    paramReq.Required,
			Description: paramReq.Description,
			// 需要确保 DefaultValue 可以匹配参数类型
			DefaultValue: paramReq.DefaultValue,
			Group:        paramReq.Group,
		}
		if err := tx.Create(&param).Error; err != nil {
			tx.Rollback()
			return InterfaceResponse{}, err
		}
		params = append(params, param)
	}
	tx.Commit()
	// 发送创建事件
	adapter.SendEvent(adapter.Event{
		Interface: &iface,
		App:       &app,
		Code:      adapter.AddToolEvent,
	})
	return InterfaceResponse{Interface: toInterfaceDTO(iface, params)}, nil
}

func GetInterface(req GetInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, err
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return InterfaceResponse{}, errors.New("interface not found")
	}
	// 获取参数列表
	var params []models.InterfaceParameter
	db.Where("interface_id = ?", iface.ID).Find(&params)
	return InterfaceResponse{Interface: toInterfaceDTO(iface, params)}, nil
}

func ListInterfaces(req ListInterfacesRequest) (InterfacesResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfacesResponse{}, err
	}
	db := database.GetDB()
	// 检查应用是否存在
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return InterfacesResponse{}, errors.New("application not found")
	}
	var ifaces []models.Interface
	if err := db.Where("app_id = ?", req.AppID).Find(&ifaces).Error; err != nil {
		return InterfacesResponse{}, err
	}
	// 批量获取所有参数
	ifaceIDs := make([]int64, 0, len(ifaces))
	for _, iface := range ifaces {
		ifaceIDs = append(ifaceIDs, iface.ID)
	}
	var allParams []models.InterfaceParameter
	if len(ifaceIDs) > 0 {
		db.Where("interface_id IN ?", ifaceIDs).Find(&allParams)
	}
	// 按 InterfaceID 分组
	paramsByIfaceID := make(map[int64][]models.InterfaceParameter)
	for _, param := range allParams {
		paramsByIfaceID[param.InterfaceID] = append(paramsByIfaceID[param.InterfaceID], param)
	}
	// 构建 DTO
	dtos := make([]InterfaceDTO, 0, len(ifaces))
	for _, iface := range ifaces {
		params := paramsByIfaceID[iface.ID]
		if params == nil {
			params = []models.InterfaceParameter{}
		}
		dtos = append(dtos, toInterfaceDTO(iface, params))
	}
	return InterfacesResponse{Interfaces: dtos}, nil
}

func UpdateInterface(req UpdateInterfaceRequest) (InterfaceResponse, error) {
	if err := validate.Struct(req); err != nil {
		return InterfaceResponse{}, err
	}
	db := database.GetDB()
	var existing models.Interface
	if err := db.First(&existing, req.ID).Error; err != nil {
		return InterfaceResponse{}, errors.New("interface not found")
	}
	var app models.Application
	if err := db.First(&app, existing.AppID).Error; err != nil {
		return InterfaceResponse{}, errors.New("application not found")
	}
	oldName := existing.Name
	// 使用事务
	tx := db.Begin()
	// 更新基本信息
	if req.Name != nil {
		// 检查名称唯一性
		var count int64
		tx.Model(&models.Interface{}).Where("app_id = ? AND name = ? AND id <> ?", existing.AppID, *req.Name, existing.ID).Count(&count)
		if count > 0 {
			tx.Rollback()
			return InterfaceResponse{}, errors.New("interface name already exists in the application")
		}
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
	if req.Method != nil {
		existing.Method = *req.Method
	}
	if req.AuthType != nil {
		existing.AuthType = *req.AuthType
	}
	if req.PostProcess != nil {
		existing.PostProcess = *req.PostProcess
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := tx.Save(&existing).Error; err != nil {
		tx.Rollback()
		return InterfaceResponse{}, err
	}

	// 如果提供了参数列表，则完全替换
	var params []models.InterfaceParameter
	if req.Parameters != nil {
		// 验证参数的 Ref 引用
		for _, paramReq := range *req.Parameters {
			if paramReq.Type == "custom" && paramReq.Ref != nil {
				var refType models.CustomType
				if err := tx.First(&refType, *paramReq.Ref).Error; err != nil {
					tx.Rollback()
					return InterfaceResponse{}, errors.New("invalid parameter reference: custom type not found")
				}
				if refType.AppID != existing.AppID {
					tx.Rollback()
					return InterfaceResponse{}, errors.New("parameter reference must belong to the same application")
				}
			}
		}
		// 删除旧参数
		tx.Where("interface_id = ?", existing.ID).Delete(&models.InterfaceParameter{})
		// 创建新参数
		for _, paramReq := range *req.Parameters {
			param := models.InterfaceParameter{
				AppID:        existing.AppID,
				InterfaceID:  existing.ID,
				Name:         paramReq.Name,
				Type:         paramReq.Type,
				Ref:          paramReq.Ref,
				Location:     paramReq.Location,
				IsArray:      paramReq.IsArray,
				Required:     paramReq.Required,
				Description:  paramReq.Description,
				DefaultValue: paramReq.DefaultValue,
				Group:        paramReq.Group,
			}
			if err := tx.Create(&param).Error; err != nil {
				tx.Rollback()
				return InterfaceResponse{}, err
			}
			params = append(params, param)
		}
	} else {
		// 如果没有提供参数列表，保持原有参数
		tx.Where("interface_id = ?", existing.ID).Find(&params)
	}
	tx.Commit()
	// 发送更新事件 删除根据名字删除就好了
	adapter.SendEvent(adapter.Event{
		Interface: &models.Interface{Name: oldName},
		App:       &app,
		Code:      adapter.RemoveToolEvent,
	})
	adapter.SendEvent(adapter.Event{
		Interface: &existing,
		App:       &app,
		Code:      adapter.AddToolEvent,
	})
	return InterfaceResponse{Interface: toInterfaceDTO(existing, params)}, nil
}

func DeleteInterface(req DeleteInterfaceRequest) (EmptyResponse, error) {
	if err := validate.Struct(req); err != nil {
		return EmptyResponse{}, err
	}
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, req.ID).Error; err != nil {
		return EmptyResponse{}, errors.New("interface not found")
	}
	var app models.Application
	if err := db.First(&app, iface.AppID).Error; err != nil {
		return EmptyResponse{}, errors.New("application not found")
	}
	// 使用事务删除
	tx := db.Begin()
	// 删除参数
	if err := tx.Where("interface_id = ?", iface.ID).Delete(&models.InterfaceParameter{}).Error; err != nil {
		tx.Rollback()
		return EmptyResponse{}, err
	}
	// 删除接口
	if err := tx.Delete(&iface).Error; err != nil {
		tx.Rollback()
		return EmptyResponse{}, err
	}
	tx.Commit()
	adapter.SendEvent(adapter.Event{
		Interface: &iface,
		App:       &app,
		Code:      adapter.RemoveToolEvent,
	})
	return EmptyResponse{}, nil
}
