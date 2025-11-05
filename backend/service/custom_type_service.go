package service

import (
	"errors"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"
)

// ========== Request/Response 结构体 ==========

type CreateCustomTypeRequest struct {
	AppID       int64                      `json:"app_id" validate:"required,gt=0"`
	Name        string                     `json:"name" validate:"required,max=255"`
	Description string                     `json:"description" validate:"max=16384"`
	Fields      []CreateCustomTypeFieldReq `json:"fields"` // 字段列表
}

type CreateCustomTypeFieldReq struct {
	Name        string `json:"name" validate:"required,max=255"`
	Type        string `json:"type" validate:"required,oneof=number string boolean custom"`
	Ref         *int64 `json:"ref"`         // 如果 type=custom，引用其他 CustomType.ID
	IsArray     bool   `json:"is_array"`    // 是否数组
	Required    bool   `json:"required"`    // 是否必填
	Description string `json:"description" validate:"max=16384"`
}

type GetCustomTypeRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListCustomTypesRequest struct {
	AppID int64 `json:"app_id" validate:"required,gt=0"`
}

type UpdateCustomTypeRequest struct {
	ID          int64                       `json:"id" validate:"required,gt=0"`
	Name        *string                     `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string                     `json:"description,omitempty" validate:"omitempty,max=16384"`
	Fields      *[]UpdateCustomTypeFieldReq `json:"fields,omitempty"` // 如果提供，则完全替换字段列表
}

type UpdateCustomTypeFieldReq struct {
	ID          *int64  `json:"id,omitempty"` // 如果有 ID，则更新；否则新建
	Name        string  `json:"name" validate:"required,max=255"`
	Type        string  `json:"type" validate:"required,oneof=number string boolean custom"`
	Ref         *int64  `json:"ref"`
	IsArray     bool    `json:"is_array"`
	Required    bool    `json:"required"`
	Description string  `json:"description" validate:"max=16384"`
}

type DeleteCustomTypeRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

// ========== DTO 结构体 ==========

type CustomTypeFieldDTO struct {
	ID           int64     `json:"id"`
	CustomTypeID int64     `json:"custom_type_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Ref          *int64    `json:"ref"`
	IsArray      bool      `json:"is_array"`
	Required     bool      `json:"required"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CustomTypeDTO struct {
	ID          int64                `json:"id"`
	AppID       int64                `json:"app_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Fields      []CustomTypeFieldDTO `json:"fields"` // 包含字段列表
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type CustomTypeResponse struct {
	CustomType CustomTypeDTO `json:"custom_type"`
}

type CustomTypesResponse struct {
	CustomTypes []CustomTypeDTO `json:"custom_types"`
}

// ========== Mapper 函数 ==========

func toCustomTypeFieldDTO(m models.CustomTypeField) CustomTypeFieldDTO {
	return CustomTypeFieldDTO{
		ID:           m.ID,
		CustomTypeID: m.CustomTypeID,
		Name:         m.Name,
		Type:         m.Type,
		Ref:          m.Ref,
		IsArray:      m.IsArray,
		Required:     m.Required,
		Description:  m.Description,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toCustomTypeDTO(m models.CustomType, fields []models.CustomTypeField) CustomTypeDTO {
	fieldDTOs := make([]CustomTypeFieldDTO, 0, len(fields))
	for _, f := range fields {
		fieldDTOs = append(fieldDTOs, toCustomTypeFieldDTO(f))
	}
	return CustomTypeDTO{
		ID:          m.ID,
		AppID:       m.AppID,
		Name:        m.Name,
		Description: m.Description,
		Fields:      fieldDTOs,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ========== Service 函数 ==========

// CreateCustomType 创建自定义类型（包含字段）
func CreateCustomType(req CreateCustomTypeRequest) (CustomTypeResponse, error) {
	if err := validate.Struct(req); err != nil {
		return CustomTypeResponse{}, err
	}

	db := database.GetDB()

	// 检查应用是否存在
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return CustomTypeResponse{}, errors.New("application not found")
	}

	// 检查类型名称在该应用下是否唯一
	var count int64
	db.Model(&models.CustomType{}).Where("app_id = ? AND name = ?", req.AppID, req.Name).Count(&count)
	if count > 0 {
		return CustomTypeResponse{}, errors.New("duplicate custom type name in this application")
	}

	// 验证字段的 Ref 引用是否有效
	for _, field := range req.Fields {
		if field.Type == "custom" && field.Ref != nil {
			var refType models.CustomType
			if err := db.First(&refType, *field.Ref).Error; err != nil {
				return CustomTypeResponse{}, errors.New("invalid field reference: custom type not found")
			}
			// 确保引用的类型属于同一个应用
			if refType.AppID != req.AppID {
				return CustomTypeResponse{}, errors.New("field reference must belong to the same application")
			}
		}
	}

	// 创建自定义类型
	customType := models.CustomType{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
	}

	// 使用事务
	tx := db.Begin()
	if err := tx.Create(&customType).Error; err != nil {
		tx.Rollback()
		return CustomTypeResponse{}, err
	}

	// 创建字段
	fields := make([]models.CustomTypeField, 0, len(req.Fields))
	for _, fieldReq := range req.Fields {
		field := models.CustomTypeField{
			CustomTypeID: customType.ID,
			Name:         fieldReq.Name,
			Type:         fieldReq.Type,
			Ref:          fieldReq.Ref,
			IsArray:      fieldReq.IsArray,
			Required:     fieldReq.Required,
			Description:  fieldReq.Description,
		}
		if err := tx.Create(&field).Error; err != nil {
			tx.Rollback()
			return CustomTypeResponse{}, err
		}
		fields = append(fields, field)
	}

	tx.Commit()

	return CustomTypeResponse{CustomType: toCustomTypeDTO(customType, fields)}, nil
}

// GetCustomType 获取单个自定义类型（包含字段）
func GetCustomType(req GetCustomTypeRequest) (CustomTypeResponse, error) {
	if err := validate.Struct(req); err != nil {
		return CustomTypeResponse{}, err
	}

	db := database.GetDB()

	var customType models.CustomType
	if err := db.First(&customType, req.ID).Error; err != nil {
		return CustomTypeResponse{}, errors.New("custom type not found")
	}

	// 获取字段列表
	var fields []models.CustomTypeField
	db.Where("custom_type_id = ?", customType.ID).Find(&fields)

	return CustomTypeResponse{CustomType: toCustomTypeDTO(customType, fields)}, nil
}

// ListCustomTypes 获取应用下的所有自定义类型
func ListCustomTypes(req ListCustomTypesRequest) (CustomTypesResponse, error) {
	if err := validate.Struct(req); err != nil {
		return CustomTypesResponse{}, err
	}

	db := database.GetDB()

	// 检查应用是否存在
	var app models.Application
	if err := db.First(&app, req.AppID).Error; err != nil {
		return CustomTypesResponse{}, errors.New("application not found")
	}

	var customTypes []models.CustomType
	if err := db.Where("app_id = ?", req.AppID).Find(&customTypes).Error; err != nil {
		return CustomTypesResponse{}, err
	}

	// 批量获取所有字段
	typeIDs := make([]int64, 0, len(customTypes))
	for _, ct := range customTypes {
		typeIDs = append(typeIDs, ct.ID)
	}

	var allFields []models.CustomTypeField
	if len(typeIDs) > 0 {
		db.Where("custom_type_id IN ?", typeIDs).Find(&allFields)
	}

	// 按 CustomTypeID 分组
	fieldsByTypeID := make(map[int64][]models.CustomTypeField)
	for _, field := range allFields {
		fieldsByTypeID[field.CustomTypeID] = append(fieldsByTypeID[field.CustomTypeID], field)
	}

	// 构建 DTO
	dtos := make([]CustomTypeDTO, 0, len(customTypes))
	for _, ct := range customTypes {
		fields := fieldsByTypeID[ct.ID]
		if fields == nil {
			fields = []models.CustomTypeField{}
		}
		dtos = append(dtos, toCustomTypeDTO(ct, fields))
	}

	return CustomTypesResponse{CustomTypes: dtos}, nil
}

// UpdateCustomType 更新自定义类型
func UpdateCustomType(req UpdateCustomTypeRequest) (CustomTypeResponse, error) {
	if err := validate.Struct(req); err != nil {
		return CustomTypeResponse{}, err
	}

	db := database.GetDB()

	var existing models.CustomType
	if err := db.First(&existing, req.ID).Error; err != nil {
		return CustomTypeResponse{}, errors.New("custom type not found")
	}

	// 使用事务
	tx := db.Begin()

	// 更新基本信息
	if req.Name != nil {
		// 检查名称唯一性
		var count int64
		tx.Model(&models.CustomType{}).Where("app_id = ? AND name = ? AND id <> ?", existing.AppID, *req.Name, existing.ID).Count(&count)
		if count > 0 {
			tx.Rollback()
			return CustomTypeResponse{}, errors.New("duplicate custom type name in this application")
		}
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}

	if err := tx.Save(&existing).Error; err != nil {
		tx.Rollback()
		return CustomTypeResponse{}, err
	}

	// 如果提供了字段列表，则完全替换
	var fields []models.CustomTypeField
	if req.Fields != nil {
		// 验证字段的 Ref 引用
		for _, fieldReq := range *req.Fields {
			if fieldReq.Type == "custom" && fieldReq.Ref != nil {
				var refType models.CustomType
				if err := tx.First(&refType, *fieldReq.Ref).Error; err != nil {
					tx.Rollback()
					return CustomTypeResponse{}, errors.New("invalid field reference: custom type not found")
				}
				if refType.AppID != existing.AppID {
					tx.Rollback()
					return CustomTypeResponse{}, errors.New("field reference must belong to the same application")
				}
			}
		}

		// 删除旧字段
		tx.Where("custom_type_id = ?", existing.ID).Delete(&models.CustomTypeField{})

		// 创建新字段
		for _, fieldReq := range *req.Fields {
			field := models.CustomTypeField{
				CustomTypeID: existing.ID,
				Name:         fieldReq.Name,
				Type:         fieldReq.Type,
				Ref:          fieldReq.Ref,
				IsArray:      fieldReq.IsArray,
				Required:     fieldReq.Required,
				Description:  fieldReq.Description,
			}
			if err := tx.Create(&field).Error; err != nil {
				tx.Rollback()
				return CustomTypeResponse{}, err
			}
			fields = append(fields, field)
		}
	} else {
		// 如果没有提供字段列表，保持原有字段
		tx.Where("custom_type_id = ?", existing.ID).Find(&fields)
	}

	tx.Commit()

	return CustomTypeResponse{CustomType: toCustomTypeDTO(existing, fields)}, nil
}

// DeleteCustomType 删除自定义类型
func DeleteCustomType(req DeleteCustomTypeRequest) (EmptyResponse, error) {
	if err := validate.Struct(req); err != nil {
		return EmptyResponse{}, err
	}

	db := database.GetDB()

	var customType models.CustomType
	if err := db.First(&customType, req.ID).Error; err != nil {
		return EmptyResponse{}, errors.New("custom type not found")
	}

	// 检查是否被其他类型的字段引用
	var count int64
	db.Model(&models.CustomTypeField{}).Where("ref = ?", customType.ID).Count(&count)
	if count > 0 {
		return EmptyResponse{}, errors.New("cannot delete custom type: referenced by other type fields")
	}

	// 检查是否被接口参数引用
	db.Model(&models.InterfaceParameter{}).Where("ref = ?", customType.ID).Count(&count)
	if count > 0 {
		return EmptyResponse{}, errors.New("cannot delete custom type: referenced by interface parameters")
	}

	// 使用事务删除
	tx := db.Begin()

	// 删除字段
	if err := tx.Where("custom_type_id = ?", customType.ID).Delete(&models.CustomTypeField{}).Error; err != nil {
		tx.Rollback()
		return EmptyResponse{}, err
	}

	// 删除类型
	if err := tx.Delete(&customType).Error; err != nil {
		tx.Rollback()
		return EmptyResponse{}, err
	}

	tx.Commit()

	return EmptyResponse{}, nil
}
