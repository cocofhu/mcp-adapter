package service

import (
	"errors"
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"

	"gorm.io/gorm"
)

type CreateCustomTypeRequest struct {
	AppID       int64                      `json:"app_id" validate:"required,gt=0"`  // 所属应用 ID
	Name        string                     `json:"name" validate:"required,max=255"` // 类型名称
	Description string                     `json:"description" validate:"max=16384"` // 类型描述
	Fields      []CreateCustomTypeFieldReq `json:"fields"`                           // 字段列表
}

type CreateCustomTypeFieldReq struct {
	Name        string `json:"name" validate:"required,max=255"`                            // 字段名称
	Type        string `json:"type" validate:"required,oneof=number string boolean custom"` // 字段类型
	Ref         *int64 `json:"ref"`                                                         // 如果 type=custom，引用其他 CustomType.ID
	IsArray     bool   `json:"is_array"`                                                    // 是否数组
	Required    bool   `json:"required"`                                                    // 是否必填
	Description string `json:"description" validate:"max=16384"`                            // 字段描述
}

type GetCustomTypeRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListCustomTypesRequest struct {
	AppID int64 `json:"app_id" validate:"required,gt=0"`
}

type UpdateCustomTypeRequest struct {
	ID          int64                       `json:"id" validate:"required,gt=0"`                          // 要更新的自定义类型 ID
	Name        *string                     `json:"name,omitempty" validate:"omitempty,max=255"`          // 如果提供，则更新名称
	Description *string                     `json:"description,omitempty" validate:"omitempty,max=16384"` // 如果提供，则更新描述
	Fields      *[]UpdateCustomTypeFieldReq `json:"fields,omitempty"`                                     // 如果提供，则完全替换字段列表
}

type UpdateCustomTypeFieldReq struct {
	ID          *int64 `json:"id,omitempty"`                                                // 增加字段的时候没有ID，更新字段时有ID(目前是先删除再增加的逻辑 该参数并未使用) 在修改时候需要清理历史数据
	Name        string `json:"name" validate:"required,max=255"`                            // 字段名称
	Type        string `json:"type" validate:"required,oneof=number string boolean custom"` // 字段类型
	Ref         *int64 `json:"ref"`                                                         // 如果 type=custom，引用其他 CustomType.ID
	IsArray     bool   `json:"is_array"`                                                    // 是否数组
	Required    bool   `json:"required"`                                                    // 是否必填
	Description string `json:"description" validate:"max=16384"`                            // 字段描述
}

type DeleteCustomTypeRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

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
	Fields      []CustomTypeFieldDTO `json:"fields"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type CustomTypeResponse struct {
	CustomType CustomTypeDTO `json:"custom_type"`
}

type CustomTypesResponse struct {
	CustomTypes []CustomTypeDTO `json:"custom_types"`
}

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

// checkCustomTypeCycle 检测自定义类型字段的循环引用 看起来会有并发问题
func checkCustomTypeCycle(db *gorm.DB, typeID int64, appID int64, newFields []CreateCustomTypeFieldReq) error {
	// 构建引用图: typeID -> []refTypeID (邻接表)
	graph := make(map[int64][]int64)
	// 入度表: typeID -> inDegree
	inDegree := make(map[int64]int)
	// 获取应用下所有现有的自定义类型
	var existingTypes []models.CustomType
	db.Where("app_id = ?", appID).Find(&existingTypes)

	// 初始化所有节点
	for _, t := range existingTypes {
		graph[t.ID] = []int64{}
		inDegree[t.ID] = 0
	}
	// 如果是创建新类型,添加到图中
	if typeID == 0 {
		// 数据库不应该出现ID为0的类型
		graph[typeID] = []int64{}
		inDegree[typeID] = 0
	}
	// 获取所有现有字段的引用关系
	// 这里应该可以修改数据库表结构,添加一个AppID字段,批量查询会更高效
	var existingFields []models.CustomTypeField
	for tid := range graph {
		if tid > 0 { // 跳过临时ID
			var fields []models.CustomTypeField
			db.Where("custom_type_id = ?", tid).Find(&fields)
			existingFields = append(existingFields, fields...)
		}
	}
	// 构建现有的引用关系和入度
	for _, field := range existingFields {
		if field.Type == "custom" && field.Ref != nil {
			// 如果当前更新的类型,跳过其旧字段(稍后会用新字段替换)
			if typeID > 0 && field.CustomTypeID == typeID {
				continue
			}
			// 添加边: field.CustomTypeID -> *field.Ref
			graph[field.CustomTypeID] = append(graph[field.CustomTypeID], *field.Ref)
			inDegree[*field.Ref]++
		}
	}
	// 添加新字段的引用关系和入度
	for _, field := range newFields {
		if field.Type == "custom" && field.Ref != nil {
			// 添加边: typeID -> *field.Ref
			graph[typeID] = append(graph[typeID], *field.Ref)
			inDegree[*field.Ref]++
		}
	}
	// Kahn 算法: 拓扑排序检测环
	// 1. 找出所有入度为0的节点
	var queue []int64
	for node := range graph {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}
	// 2. BFS 处理
	processedCount := 0
	for len(queue) > 0 {
		// 取出队首节点
		current := queue[0]
		queue = queue[1:]
		processedCount++
		// 遍历所有邻接节点
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			// 如果入度变为0,加入队列
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	// 3. 如果处理的节点数小于总节点数,说明存在环
	if processedCount < len(graph) {
		return errors.New("circular reference detected in custom type fields")
	}
	return nil
}

// checkCustomTypeCycleForUpdate 检测更新时的循环引用
func checkCustomTypeCycleForUpdate(db *gorm.DB, typeID int64, appID int64, newFields []UpdateCustomTypeFieldReq) error {
	// 转换为 CreateCustomTypeFieldReq 格式
	createFields := make([]CreateCustomTypeFieldReq, len(newFields))
	for i, f := range newFields {
		createFields[i] = CreateCustomTypeFieldReq{
			Name:        f.Name,
			Type:        f.Type,
			Ref:         f.Ref,
			IsArray:     f.IsArray,
			Required:    f.Required,
			Description: f.Description,
		}
	}
	return checkCustomTypeCycle(db, typeID, appID, createFields)
}

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
		if field.Type == "custom" {
			if field.Ref == nil {
				return CustomTypeResponse{}, errors.New("field reference must be provided for custom type field")
			}
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
	// 检测循环引用
	if err := checkCustomTypeCycle(db, 0, req.AppID, req.Fields); err != nil {
		return CustomTypeResponse{}, err
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
	// 批量获取所有字段 增加AppId同样可以优化 避免where in 走不到索引
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
		// 检测循环引用
		if err := checkCustomTypeCycleForUpdate(tx, existing.ID, existing.AppID, *req.Fields); err != nil {
			tx.Rollback()
			return CustomTypeResponse{}, err
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
	// 这里还需要向上递归寻找间接引用找到所有的依赖改类型的应用发送 ToolListChanged 通知
	// 举例来说 Type1 里有一个 Type2 的字段，Type2 里有一个 Type3 的字段
	// 如果现在更新 Type3，则 Type2 和 Type1 也都需要收到通知，目前这里处理了接口的直接引用
	var refs []int64
	if db.Model(&models.InterfaceParameter{}).Where("ref = ?", existing.ID).Distinct("app_id").Pluck("app_id", &refs).Error == nil {
		var refApps []models.Application
		db.Model(&models.Application{}).Where("id IN ?", refs).Find(&refApps)
		for i := range refApps {
			if !refApps[i].Enabled {
				continue
			}
			adapter.SendEvent(adapter.Event{
				App:       &refApps[i],
				Interface: nil,
				Code:      adapter.ToolListChanged,
			})
		}
	}
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
