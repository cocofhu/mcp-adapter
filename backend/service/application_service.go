package service

import (
	"encoding/json"
	"errors"
	"log"
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CreateApplicationRequest struct {
	Name        string `json:"name" validate:"required,max=128"`                  // 应用名称 不允许重复
	Description string `json:"description" validate:"max=16384"`                  // 应用描述
	Path        string `json:"path" validate:"required,max=128"`                  // 应用路由标识
	Protocol    string `json:"protocol" validate:"required,oneof=sse streamable"` // 应用暴露协议
	PostProcess string `json:"post_process" validate:"max=1048576"`               // 应用后处理脚本
	Environment string `json:"environment" validate:"max=1048576"`                // 应用环境变量
	Enabled     *bool  `json:"enabled,omitempty"`                                 // 是否启用应用
}

type GetApplicationRequest struct {
	ID         int64 `json:"id" validate:"required,gt=0"`
	ShowDetail bool  `json:"show_detail,omitempty"`
}

type ListApplicationsRequest struct{}

type UpdateApplicationRequest struct {
	ID          int64   `json:"id" validate:"required,gt=0"`
	Name        *string `json:"name" validate:"required,max=128"`                  // 应用名称 不允许重复
	Description *string `json:"description" validate:"max=16384"`                  // 应用描述
	Path        *string `json:"path" validate:"required,max=128"`                  // 应用路径标识
	Protocol    *string `json:"protocol" validate:"required,oneof=sse streamable"` // 应用暴露协议
	PostProcess *string `json:"post_process" validate:"omitempty,max=1048576"`     // 应用后处理脚本
	Environment *string `json:"environment" validate:"omitempty,max=1048576"`      // 应用环境变量
	Enabled     *bool   `json:"enabled,omitempty"`                                 // 是否启用应用
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

type FixedInputDTO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Value       any    `json:"value"`
}

type ToolMetaDTO struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	AuthType string `json:"auth_type"`
}

type MCPToolDefinitionDTO struct {
	Name         string                  `json:"name"`
	FixedInput   []FixedInputDTO         `json:"fixed_input"`
	InputSchema  map[string]any          `json:"input_schema"`
	OutputSchema map[string]any          `json:"output_schema"`
	PostProcess  adapter.PostProcessMeta `json:"post_process"`
	ToolMeta     ToolMetaDTO             `json:"tool_meta"`
}

type ApplicationDetailResponse struct {
	Application     ApplicationDTO         `json:"application"`
	ToolDefinitions []MCPToolDefinitionDTO `json:"tool_definitions"`
}

type ApplicationResponse struct {
	Application ApplicationDTO `json:"application"`
}

type ApplicationsResponse struct {
	Applications []ApplicationDTO `json:"applications"`
}

type EmptyResponse struct{}

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
	// Name 唯一性检查
	var count int64
	db.Model(&models.Application{}).Where("name = ?", app.Name).Count(&count)
	if count > 0 {
		return ApplicationResponse{}, errors.New("duplicate application name")
	}
	// Path 唯一性检查
	db.Model(&models.Application{}).Where("path = ?", app.Path).Count(&count)
	if count > 0 {
		return ApplicationResponse{}, errors.New("duplicate application path")
	}
	if err := db.Create(&app).Error; err != nil {
		return ApplicationResponse{}, err
	}
	adapter.SendEvent(adapter.Event{
		App:       &app,
		Interface: nil,
		Code:      adapter.AddApplicationEvent,
	})
	return ApplicationResponse{Application: toApplicationDTO(app)}, nil
}

// GetApplication 获取单个应用
func GetApplication(req GetApplicationRequest) (ApplicationDetailResponse, error) {
	if err := validate.Struct(req); err != nil {
		return ApplicationDetailResponse{}, err
	}
	db := database.GetDB()
	var app models.Application
	if err := db.First(&app, req.ID).Error; err != nil {
		return ApplicationDetailResponse{}, errors.New("no such application")
	}
	if !req.ShowDetail {
		return ApplicationDetailResponse{Application: toApplicationDTO(app)}, nil
	}
	interfaces, err := ListInterfaces(ListInterfacesRequest{AppID: app.ID})
	if err != nil {
		return ApplicationDetailResponse{}, err
	}
	toolDefinitions := make([]MCPToolDefinitionDTO, 0)
	for _, iface := range interfaces.Interfaces {
		inputSchema, err := adapter.BuildMcpInputSchemaByInterface(iface.ID)
		if err != nil {
			return ApplicationDetailResponse{}, err
		}
		outputSchema, err := adapter.BuildMcpOutputSchemaByInterface(iface.ID)
		if err != nil {
			return ApplicationDetailResponse{}, err
		}
		fixedInputs := make([]FixedInputDTO, 0)
		for _, param := range iface.Parameters {
			if param.Group != "fixed" {
				continue
			}
			if param.DefaultValue == nil {
				log.Printf("Warning: interface %s has a fixed input %s without default value", iface.Name, param.Name)
				continue
			}
			value, err := adapter.ConvertDefaultValue(*param.DefaultValue, param.Type)
			if err != nil {
				log.Printf("Warning: failed to convert default value for parameter %s: %v", param.Name, err)
				continue
			}
			fixedInputs = append(fixedInputs, FixedInputDTO{
				Name:        param.Name,
				Value:       value,
				Type:        param.Type,
				Description: param.Description,
				Location:    param.Location,
			})
		}
		postProcessMeta := adapter.PostProcessMeta{
			TruncateFields:   make(map[string]int),
			StructuredOutput: false,
		}
		if iface.PostProcess != "" {
			if err := json.Unmarshal([]byte(iface.PostProcess), &postProcessMeta); err != nil {
				log.Printf("Error unmarshalling post process meta: %v, tool id %d", err, iface.ID)
			}
			log.Printf("Post process meta for tool %s: %+v", iface.Name, postProcessMeta)
		}

		toolDefinitions = append(toolDefinitions, MCPToolDefinitionDTO{
			Name:         iface.Name,
			FixedInput:   fixedInputs,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
			PostProcess:  postProcessMeta,
			ToolMeta: ToolMetaDTO{
				URL:      iface.URL,
				Method:   iface.Method,
				AuthType: iface.AuthType,
			},
		})
	}
	return ApplicationDetailResponse{Application: toApplicationDTO(app), ToolDefinitions: toolDefinitions}, nil
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
	oldName := existing.Name
	oldPath := existing.Path
	if req.Name != nil {
		// Name 唯一性检查
		var count int64
		db.Model(&models.Application{}).Where("name = ? AND id <> ?", *req.Name, existing.ID).Count(&count)
		if count > 0 {
			return ApplicationResponse{}, errors.New("duplicate application name")
		}
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Path != nil {
		// Path 唯一性检查
		var count int64
		db.Model(&models.Application{}).Where("path = ? AND id <> ?", *req.Path, existing.ID).Count(&count)
		if count > 0 {
			return ApplicationResponse{}, errors.New("duplicate application path")
		}
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
	adapter.SendEvent(adapter.Event{
		App:       &models.Application{Name: oldName, Path: oldPath},
		Interface: nil,
		Code:      adapter.RemoveApplicationEvent,
	})
	adapter.SendEvent(adapter.Event{
		App:       &existing,
		Interface: nil,
		Code:      adapter.AddApplicationEvent,
	})
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
	adapter.SendEvent(adapter.Event{
		App:       &app,
		Interface: nil,
		Code:      adapter.RemoveApplicationEvent,
	})
	return EmptyResponse{}, nil
}
