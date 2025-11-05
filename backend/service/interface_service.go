package service

import (
	"errors"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"

	"gorm.io/gorm"
)

// ========== 循环引用检测 ==========

// checkInterfaceParameterCycle 使用拓扑排序(Kahn算法)检测接口参数是否存在循环引用
// 接口参数可能引用自定义类型,而自定义类型的字段也可能引用其他自定义类型
func checkInterfaceParameterCycle(db *gorm.DB, appID int64, params []CreateInterfaceParameterReq) error {
	// 构建引用图: typeID -> []refTypeID (邻接表)
	graph := make(map[int64][]int64)
	// 入度表: typeID -> inDegree
	inDegree := make(map[int64]int)
	
	// 获取应用下所有自定义类型
	var existingTypes []models.CustomType
	db.Where("app_id = ?", appID).Find(&existingTypes)
	
	// 初始化所有节点
	for _, t := range existingTypes {
		graph[t.ID] = []int64{}
		inDegree[t.ID] = 0
	}
	
	// 获取所有字段的引用关系
	var existingFields []models.CustomTypeField
	for tid := range graph {
		var fields []models.CustomTypeField
		db.Where("custom_type_id = ?", tid).Find(&fields)
		existingFields = append(existingFields, fields...)
	}
	
	// 构建引用关系和入度
	for _, field := range existingFields {
		if field.Type == "custom" && field.Ref != nil {
			// 添加边: field.CustomTypeID -> *field.Ref
			graph[field.CustomTypeID] = append(graph[field.CustomTypeID], *field.Ref)
			inDegree[*field.Ref]++
		}
	}
	
	// Kahn 算法: 拓扑排序检测环
	// 1. 找出所有入度为0的节点
	queue := []int64{}
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
		return errors.New("circular reference detected in interface parameters")
	}
	
	return nil
}

// checkInterfaceParameterCycleForUpdate 检测更新时的循环引用
func checkInterfaceParameterCycleForUpdate(db *gorm.DB, appID int64, params []UpdateInterfaceParameterReq) error {
	// 转换为 CreateInterfaceParameterReq 格式
	createParams := make([]CreateInterfaceParameterReq, len(params))
	for i, p := range params {
		createParams[i] = CreateInterfaceParameterReq{
			Name:         p.Name,
			Type:         p.Type,
			Ref:          p.Ref,
			Location:     p.Location,
			IsArray:      p.IsArray,
			Required:     p.Required,
			Description:  p.Description,
			DefaultValue: p.DefaultValue,
		}
	}
	return checkInterfaceParameterCycle(db, appID, createParams)
}

// ========== Request/Response 结构体 ==========

type CreateInterfaceRequest struct {
	AppID       int64                           `json:"app_id" validate:"required,gt=0"`
	Name        string                          `json:"name" validate:"required,max=255"`
	Description string                          `json:"description" validate:"max=16384"`
	Protocol    string                          `json:"protocol" validate:"required,oneof=http"`
	URL         string                          `json:"url" validate:"required,max=1024"`
	Method      string                          `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	AuthType    string                          `json:"auth_type" validate:"required,oneof=none"`
	Enabled     bool                            `json:"enabled"`
	PostProcess string                          `json:"post_process" validate:"max=1048576"`
	Parameters  []CreateInterfaceParameterReq   `json:"parameters"` // 接口参数列表
}

type CreateInterfaceParameterReq struct {
	Name         string `json:"name" validate:"required,max=255"`
	Type         string `json:"type" validate:"required,oneof=number string boolean custom"`
	Ref          *int64 `json:"ref"`                                        // 如果 type=custom，引用 CustomType.ID
	Location     string `json:"location" validate:"required,oneof=query header body path"`
	IsArray      bool   `json:"is_array"`
	Required     bool   `json:"required"`
	Description  string `json:"description" validate:"max=16384"`
	DefaultValue *string `json:"default_value"`
}

type GetInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type ListInterfacesRequest struct {
	AppID int64 `json:"app_id" validate:"required,gt=0"`
}
type UpdateInterfaceRequest struct {
	ID          int64                            `json:"id" validate:"required,gt=0"`
	Name        *string                          `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string                          `json:"description,omitempty" validate:"omitempty,max=16384"`
	Protocol    *string                          `json:"protocol,omitempty" validate:"omitempty,oneof=http"`
	URL         *string                          `json:"url,omitempty" validate:"omitempty,max=1024"`
	Method      *string                          `json:"method,omitempty" validate:"omitempty,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	AuthType    *string                          `json:"auth_type,omitempty" validate:"omitempty,oneof=none"`
	Enabled     *bool                            `json:"enabled,omitempty"`
	PostProcess *string                          `json:"post_process,omitempty" validate:"omitempty,max=1048576"`
	Parameters  *[]UpdateInterfaceParameterReq   `json:"parameters,omitempty"` // 如果提供，则完全替换参数列表
}

type UpdateInterfaceParameterReq struct {
	Name         string  `json:"name" validate:"required,max=255"`
	Type         string  `json:"type" validate:"required,oneof=number string boolean custom"`
	Ref          *int64  `json:"ref"`
	Location     string  `json:"location" validate:"required,oneof=query header body path"`
	IsArray      bool    `json:"is_array"`
	Required     bool    `json:"required"`
	Description  string  `json:"description" validate:"max=16384"`
	DefaultValue *string `json:"default_value"`
}

type DeleteInterfaceRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type InterfaceParameterDTO struct {
	ID           int64   `json:"id"`
	InterfaceID  int64   `json:"interface_id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Ref          *int64  `json:"ref"`
	Location     string  `json:"location"`
	IsArray      bool    `json:"is_array"`
	Required     bool    `json:"required"`
	Description  string  `json:"description"`
	DefaultValue *string `json:"default_value"`
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
	Parameters  []InterfaceParameterDTO `json:"parameters"` // 包含参数列表
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
	
	// 检测循环引用
	if err := checkInterfaceParameterCycle(db, req.AppID, req.Parameters); err != nil {
		return InterfaceResponse{}, err
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
			AppID:        req.AppID,
			InterfaceID:  iface.ID,
			Name:         paramReq.Name,
			Type:         paramReq.Type,
			Ref:          paramReq.Ref,
			Location:     paramReq.Location,
			IsArray:      paramReq.IsArray,
			Required:     paramReq.Required,
			Description:  paramReq.Description,
			DefaultValue: paramReq.DefaultValue,
		}
		if err := tx.Create(&param).Error; err != nil {
			tx.Rollback()
			return InterfaceResponse{}, err
		}
		params = append(params, param)
	}
	
	tx.Commit()
	
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
		
		// 检测循环引用
		if err := checkInterfaceParameterCycleForUpdate(tx, existing.AppID, *req.Parameters); err != nil {
			tx.Rollback()
			return InterfaceResponse{}, err
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
	
	return EmptyResponse{}, nil
}
