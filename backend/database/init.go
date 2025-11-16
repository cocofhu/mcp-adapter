package database

import (
	"log"
	"mcp-adapter/backend/models"
)

// InterfaceParam 接口参数定义
type InterfaceParam struct {
	Name         string
	Type         string
	Location     string
	IsArray      bool
	Required     bool
	Description  string
	DefaultValue *string // 默认值
	Group        string  // 参数组: input-输入参数, output-输出参数, fixed-固定参数
	Ref          *int64  // 引用自定义类型ID
}

// InterfaceDefinition 接口定义
type InterfaceDefinition struct {
	Name        string
	Description string
	URL         string
	Method      string
	Parameters  []InterfaceParam
}

// CustomTypeDefinition 自定义类型定义
type CustomTypeDefinition struct {
	Name        string
	Description string
	Fields      []CustomTypeFieldDef
}

// CustomTypeFieldDef 自定义类型字段定义
type CustomTypeFieldDef struct {
	Name        string
	Type        string
	IsArray     bool
	Required    bool
	Description string
}

// InitDefaultData 初始化默认数据
func InitDefaultData() {
	db := GetDB()

	// 检查是否已存在 MCP-Adapter 应用
	var count int64
	db.Model(&models.Application{}).Where("name = ?", "MCP-Adapter").Count(&count)
	if count > 0 {
		log.Println("Default application 'MCP-Adapter' already exists, skipping initialization")
		return
	}

	// 创建默认应用
	app := models.Application{
		Name:        "MCP-Adapter",
		Description: "MCP-Adapter 系统管理控制台 - 这是一个完整的MCP(Model Context Protocol)适配器管理系统，专门为大模型提供HTTP接口转换为MCP协议的能力。该应用包含三大核心功能模块：\n\n1. 【应用管理】- 管理MCP应用容器，每个应用代表一个独立的MCP服务端点，通过唯一的path标识对外提供服务。应用是接口的逻辑分组，支持独立的环境配置和后处理脚本。\n\n2. 【接口管理】- 管理HTTP API接口定义，将外部HTTP服务包装为MCP工具。支持完整的RESTful接口定义，包括URL路径参数、查询参数、请求头、请求体等。接口定义后可通过MCP协议被大模型调用。\n\n3. 【自定义类型管理】- 管理复杂数据结构定义，支持嵌套类型和数组类型。当接口参数为复杂对象时，可通过自定义类型进行结构化定义，确保大模型能正确理解和使用参数格式。\n\n核心价值：让大模型能够通过标准化的MCP协议调用任意HTTP接口，实现AI与外部系统的无缝集成。系统自动处理协议转换、参数验证、类型检查等复杂逻辑，大模型只需要按照MCP规范调用即可。\n\n使用场景：AI Agent集成第三方API、企业内部系统对接、微服务调用、数据查询等任何需要AI调用HTTP接口的场景。",
		Path:        "adapter",
		Protocol:    "sse",
		Enabled:     true,
	}

	if err := db.Create(&app).Error; err != nil {
		log.Printf("Failed to create default application: %v", err)
		return
	}

	log.Printf("Created default application 'MCP-Adapter' with ID: %d", app.ID)

	// 定义自定义类型
	customTypes := []CustomTypeDefinition{
		{
			Name:        "InterfaceParameter",
			Description: "接口参数定义",
			Fields: []CustomTypeFieldDef{
				{Name: "name", Type: "string", Required: true, Description: "参数名称"},
				{Name: "type", Type: "string", Required: true, Description: "参数类型: number, string, boolean, custom"},
				{Name: "ref", Type: "number", Required: false, Description: "如果type=custom，引用CustomType的ID"},
				{Name: "location", Type: "string", Required: true, Description: "参数位置: query, header, body, path"},
				{Name: "is_array", Type: "boolean", Required: false, Description: "是否为数组类型"},
				{Name: "required", Type: "boolean", Required: false, Description: "是否必填"},
				{Name: "description", Type: "string", Required: false, Description: "参数描述"},
				{Name: "default_value", Type: "string", Required: false, Description: "默认值"},
				{Name: "group", Type: "string", Required: true, Description: "参数组: input-输入参数(用户调用时提供), output-输出参数(从响应提取), fixed-固定参数(使用默认值,不可修改)"},
			},
		},
		{
			Name:        "CustomTypeField",
			Description: "自定义类型字段定义",
			Fields: []CustomTypeFieldDef{
				{Name: "name", Type: "string", Required: true, Description: "字段名称"},
				{Name: "type", Type: "string", Required: true, Description: "字段类型: number, string, boolean, custom"},
				{Name: "ref", Type: "number", Required: false, Description: "如果type=custom，引用CustomType的ID"},
				{Name: "is_array", Type: "boolean", Required: false, Description: "是否为数组类型"},
				{Name: "required", Type: "boolean", Required: false, Description: "是否必填"},
				{Name: "description", Type: "string", Required: false, Description: "字段描述"},
			},
		},
		// 响应类型定义
		{
			Name:        "ApplicationResponse",
			Description: "应用响应数据",
			Fields: []CustomTypeFieldDef{
				{Name: "id", Type: "number", Required: true, Description: "应用ID"},
				{Name: "name", Type: "string", Required: true, Description: "应用名称"},
				{Name: "description", Type: "string", Required: false, Description: "应用描述"},
				{Name: "path", Type: "string", Required: true, Description: "应用路径"},
				{Name: "protocol", Type: "string", Required: true, Description: "应用协议"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本"},
				{Name: "environment", Type: "string", Required: false, Description: "环境变量"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间"},
			},
		},
		{
			Name:        "InterfaceResponse",
			Description: "接口响应数据",
			Fields: []CustomTypeFieldDef{
				{Name: "id", Type: "number", Required: true, Description: "接口ID"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID"},
				{Name: "name", Type: "string", Required: true, Description: "接口名称"},
				{Name: "description", Type: "string", Required: false, Description: "接口描述"},
				{Name: "protocol", Type: "string", Required: true, Description: "协议类型"},
				{Name: "url", Type: "string", Required: true, Description: "接口URL"},
				{Name: "method", Type: "string", Required: true, Description: "HTTP方法"},
				{Name: "auth_type", Type: "string", Required: true, Description: "鉴权类型"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间"},
			},
		},
		{
			Name:        "CustomTypeResponse",
			Description: "自定义类型响应数据（包含字段列表）",
			Fields: []CustomTypeFieldDef{
				{Name: "id", Type: "number", Required: true, Description: "类型ID"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID"},
				{Name: "name", Type: "string", Required: true, Description: "类型名称"},
				{Name: "description", Type: "string", Required: false, Description: "类型描述"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间"},
			},
		},
		{
			Name:        "OperationResult",
			Description: "操作结果",
			Fields: []CustomTypeFieldDef{
				{Name: "success", Type: "boolean", Required: true, Description: "操作是否成功"},
				{Name: "message", Type: "string", Required: false, Description: "结果消息"},
			},
		},
	}

	// 创建自定义类型并保存ID映射
	customTypeIDs := make(map[string]int64)
	for _, ctDef := range customTypes {
		ct := models.CustomType{
			AppID:       app.ID,
			Name:        ctDef.Name,
			Description: ctDef.Description,
		}

		if err := db.Create(&ct).Error; err != nil {
			log.Printf("Failed to create custom type '%s': %v", ctDef.Name, err)
			continue
		}

		customTypeIDs[ctDef.Name] = ct.ID
		log.Printf("Created custom type '%s' with ID: %d", ctDef.Name, ct.ID)

		// 创建字段
		for _, fieldDef := range ctDef.Fields {
			field := models.CustomTypeField{
				CustomTypeID: ct.ID,
				Name:         fieldDef.Name,
				Type:         fieldDef.Type,
				IsArray:      fieldDef.IsArray,
				Required:     fieldDef.Required,
				Description:  fieldDef.Description,
			}

			if err := db.Create(&field).Error; err != nil {
				log.Printf("Failed to create field '%s' for custom type '%s': %v", fieldDef.Name, ctDef.Name, err)
				continue
			}
		}

		log.Printf("Created %d fields for custom type '%s'", len(ctDef.Fields), ctDef.Name)
	}

	// 获取自定义类型ID的指针（用于参数引用）
	interfaceParamTypeID := customTypeIDs["InterfaceParameter"]
	customTypeFieldTypeID := customTypeIDs["CustomTypeField"]
	applicationResponseTypeID := customTypeIDs["ApplicationResponse"]
	interfaceResponseTypeID := customTypeIDs["InterfaceResponse"]
	customTypeResponseTypeID := customTypeIDs["CustomTypeResponse"]

	// 为 CustomTypeResponse 添加 fields 字段（引用 CustomTypeField）
	fieldsField := models.CustomTypeField{
		CustomTypeID: customTypeResponseTypeID,
		Name:         "fields",
		Type:         "custom",
		Ref:          &customTypeFieldTypeID,
		IsArray:      true,
		Required:     false,
		Description:  "字段列表",
	}
	if err := db.Create(&fieldsField).Error; err != nil {
		log.Printf("Failed to add fields field to CustomTypeResponse: %v", err)
	} else {
		log.Printf("Added fields field to CustomTypeResponse")
	}

	// 定义系统接口列表及其参数
	interfaces := []InterfaceDefinition{
		// 应用管理接口
		{
			Name:        "CreateApplication",
			Description: "创建新应用。应用是接口的容器，每个应用可以包含多个接口。应用通过path字段对外暴露MCP服务，例如path='myapp'时，protocol='sse'可通过/sse/myapp访问，protocol='streamable'可通过/streamable/myapp访问。",
			URL:         "http://localhost:8080/api/applications",
			Method:      "POST",
			Parameters: []InterfaceParam{
				{Name: "name", Type: "string", Location: "body", Required: true, Description: "应用名称，必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "应用描述，用于说明应用的用途", Group: "input"},
				{Name: "path", Type: "string", Location: "body", Required: true, Description: "应用路径标识，必须唯一，用于构建访问URL", Group: "input"},
				{Name: "protocol", Type: "string", Location: "body", Required: true, Description: "应用对外协议，支持'sse'和'streamable'", Group: "input"},
				{Name: "post_process", Type: "string", Location: "body", Required: false, Description: "后处理脚本，用于处理接口返回结果", Group: "input"},
				{Name: "environment", Type: "string", Location: "body", Required: false, Description: "环境变量，JSON字符串格式", Group: "input"},
				{Name: "enabled", Type: "boolean", Location: "body", Required: false, Description: "是否启用应用，默认为true", Group: "input"},

				{Name: "id", Type: "number", Required: true, Description: "应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "应用名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "应用描述", Group: "output"},
				{Name: "path", Type: "string", Required: true, Description: "应用路径", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "应用协议", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "environment", Type: "string", Required: false, Description: "环境变量", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "GetApplications",
			Description: "获取所有应用列表。返回系统中所有已创建的应用，包括已启用和未启用的应用。",
			URL:         "http://localhost:8080/api/applications",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "applications", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "应用列表数组", Group: "output", Ref: &applicationResponseTypeID},
			},
		},
		{
			Name:        "GetApplication",
			Description: "根据ID获取单个应用的详细信息，包括应用的所有配置项。",
			URL:         "http://localhost:8080/api/applications/{id}",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "应用ID", Group: "input"},

				{Name: "id", Type: "number", Required: true, Description: "应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "应用名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "应用描述", Group: "output"},
				{Name: "path", Type: "string", Required: true, Description: "应用路径", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "应用协议", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "environment", Type: "string", Required: false, Description: "环境变量", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "UpdateApplication",
			Description: "更新应用信息。可以更新应用的任意字段，未提供的字段保持不变。注意：如果应用下有接口正在使用，修改path可能影响访问。",
			URL:         "http://localhost:8080/api/applications/{id}",
			Method:      "PUT",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要更新的应用ID", Group: "input"},
				{Name: "name", Type: "string", Location: "body", Required: false, Description: "应用名称，必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "应用描述", Group: "input"},
				{Name: "path", Type: "string", Location: "body", Required: false, Description: "应用路径标识，必须唯一", Group: "input"},
				{Name: "protocol", Type: "string", Location: "body", Required: false, Description: "应用协议，支持'sse'和'streamable'", Group: "input"},
				{Name: "post_process", Type: "string", Location: "body", Required: false, Description: "后处理脚本", Group: "input"},
				{Name: "environment", Type: "string", Location: "body", Required: false, Description: "环境变量，JSON字符串格式", Group: "input"},
				{Name: "enabled", Type: "boolean", Location: "body", Required: false, Description: "是否启用应用", Group: "input"},

				{Name: "id", Type: "number", Required: true, Description: "应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "应用名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "应用描述", Group: "output"},
				{Name: "path", Type: "string", Required: true, Description: "应用路径", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "应用协议", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "environment", Type: "string", Required: false, Description: "环境变量", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "DeleteApplication",
			Description: "删除应用。注意：如果应用下存在接口，删除操作将失败。需要先删除应用下的所有接口才能删除应用。",
			URL:         "http://localhost:8080/api/applications/{id}",
			Method:      "DELETE",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要删除的应用ID", Group: "input"},

				{Name: "success", Type: "boolean", Required: true, Description: "操作是否成功", Group: "output"},
				{Name: "message", Type: "string", Required: false, Description: "结果消息", Group: "output"},
			},
		},
		// 接口管理接口
		{
			Name:        "CreateInterface",
			Description: "创建新接口。接口定义了一个HTTP API的完整信息，包括URL、方法、参数等。接口名称在同一应用内必须唯一。parameters字段是InterfaceParameter类型的数组，用于定义接口的输入参数。URL中可以使用{name}格式的占位符表示path参数。",
			URL:         "http://localhost:8080/api/interfaces",
			Method:      "POST",
			Parameters: []InterfaceParam{
				{Name: "app_id", Type: "number", Location: "body", Required: true, Description: "所属应用ID，接口必须归属于某个应用", Group: "input"},
				{Name: "name", Type: "string", Location: "body", Required: true, Description: "接口名称，在同一应用内必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "接口描述，详细说明接口的功能和用法", Group: "input"},
				{Name: "protocol", Type: "string", Location: "body", Required: true, Description: "协议类型，目前仅支持'http'", Group: "input"},
				{Name: "url", Type: "string", Location: "body", Required: true, Description: "接口URL，支持{name}格式的path参数占位符", Group: "input"},
				{Name: "method", Type: "string", Location: "body", Required: true, Description: "HTTP方法: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS", Group: "input"},
				{Name: "auth_type", Type: "string", Location: "body", Required: true, Description: "鉴权类型，目前仅支持'none和capi'", Group: "input"},
				{Name: "enabled", Type: "boolean", Location: "body", Required: false, Description: "是否启用接口，默认为true", Group: "input"},
				{Name: "post_process", Type: "string", Location: "body", Required: false, Description: "后处理脚本，用于处理接口返回结果", Group: "input"},
				{Name: "parameters", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "接口参数列表，InterfaceParameter类型的数组", Group: "input", Ref: &interfaceParamTypeID},

				{Name: "id", Type: "number", Required: true, Description: "接口ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "接口名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "接口描述", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "协议类型", Group: "output"},
				{Name: "url", Type: "string", Required: true, Description: "接口URL", Group: "output"},
				{Name: "method", Type: "string", Required: true, Description: "HTTP方法", Group: "output"},
				{Name: "auth_type", Type: "string", Required: true, Description: "鉴权类型", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "GetInterfaces",
			Description: "获取指定应用下的所有接口列表。返回接口的完整信息，包括参数定义。",
			URL:         "http://localhost:8080/api/interfaces",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "app_id", Type: "number", Location: "query", Required: true, Description: "应用ID，查询该应用下的所有接口", Group: "input"},
				{Name: "interfaces", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "接口列表数组", Group: "output", Ref: &interfaceResponseTypeID},
			},
		},
		{
			Name:        "GetInterface",
			Description: "根据ID获取单个接口的详细信息，包括接口的所有参数定义。",
			URL:         "http://localhost:8080/api/interfaces/{id}",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "接口ID", Group: "input"},

				{Name: "id", Type: "number", Required: true, Description: "接口ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "接口名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "接口描述", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "协议类型", Group: "output"},
				{Name: "url", Type: "string", Required: true, Description: "接口URL", Group: "output"},
				{Name: "method", Type: "string", Required: true, Description: "HTTP方法", Group: "output"},
				{Name: "auth_type", Type: "string", Required: true, Description: "鉴权类型", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "UpdateInterface",
			Description: "更新接口信息。可以更新接口的任意字段，未提供的字段保持不变。如果提供了parameters字段，将完全替换原有的参数列表。",
			URL:         "http://localhost:8080/api/interfaces/{id}",
			Method:      "PUT",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要更新的接口ID", Group: "input"},
				{Name: "name", Type: "string", Location: "body", Required: false, Description: "接口名称，在同一应用内必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "接口描述", Group: "input"},
				{Name: "protocol", Type: "string", Location: "body", Required: false, Description: "协议类型，目前仅支持'http'", Group: "input"},
				{Name: "url", Type: "string", Location: "body", Required: false, Description: "接口URL，支持{name}格式的path参数占位符", Group: "input"},
				{Name: "method", Type: "string", Location: "body", Required: false, Description: "HTTP方法", Group: "input"},
				{Name: "auth_type", Type: "string", Location: "body", Required: false, Description: "鉴权类型", Group: "input"},
				{Name: "enabled", Type: "boolean", Location: "body", Required: false, Description: "是否启用接口", Group: "input"},
				{Name: "post_process", Type: "string", Location: "body", Required: false, Description: "后处理脚本", Group: "input"},
				{Name: "parameters", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "接口参数列表，如果提供则完全替换原有参数", Group: "input", Ref: &interfaceParamTypeID},

				{Name: "id", Type: "number", Required: true, Description: "接口ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "接口名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "接口描述", Group: "output"},
				{Name: "protocol", Type: "string", Required: true, Description: "协议类型", Group: "output"},
				{Name: "url", Type: "string", Required: true, Description: "接口URL", Group: "output"},
				{Name: "method", Type: "string", Required: true, Description: "HTTP方法", Group: "output"},
				{Name: "auth_type", Type: "string", Required: true, Description: "鉴权类型", Group: "output"},
				{Name: "enabled", Type: "boolean", Required: true, Description: "是否启用", Group: "output"},
				{Name: "post_process", Type: "string", Required: false, Description: "后处理脚本", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "DeleteInterface",
			Description: "删除接口。删除后，该接口将不再通过MCP协议对外提供服务。",
			URL:         "http://localhost:8080/api/interfaces/{id}",
			Method:      "DELETE",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要删除的接口ID", Group: "input"},

				{Name: "success", Type: "boolean", Required: true, Description: "操作是否成功", Group: "output"},
				{Name: "message", Type: "string", Required: false, Description: "结果消息", Group: "output"},
			},
		},
		// 自定义类型管理接口
		{
			Name:        "CreateCustomType",
			Description: "创建自定义类型。自定义类型用于定义复杂的数据结构，可以在接口参数中引用。类型名称在同一应用内必须唯一。fields字段是CustomTypeField类型的数组，定义了类型包含的字段。字段可以是基础类型(number/string/boolean)或引用其他自定义类型(custom)。",
			URL:         "http://localhost:8080/api/custom-types",
			Method:      "POST",
			Parameters: []InterfaceParam{
				{Name: "app_id", Type: "number", Location: "body", Required: true, Description: "所属应用ID，自定义类型必须归属于某个应用", Group: "input"},
				{Name: "name", Type: "string", Location: "body", Required: true, Description: "类型名称，在同一应用内必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "类型描述，说明类型的用途和结构", Group: "input"},
				{Name: "fields", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "字段列表，CustomTypeField类型的数组", Group: "input", Ref: &customTypeFieldTypeID},

				{Name: "id", Type: "number", Required: true, Description: "类型ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "类型名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "类型描述", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "GetCustomTypes",
			Description: "获取指定应用下的所有自定义类型列表。返回类型的完整信息，包括字段定义。",
			URL:         "http://localhost:8080/api/custom-types",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "app_id", Type: "number", Location: "query", Required: true, Description: "应用ID，查询该应用下的所有自定义类型", Group: "input"},
				{Name: "custom_types", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "自定义类型列表数组", Group: "output", Ref: &customTypeResponseTypeID},
			},
		},
		{
			Name:        "GetCustomType",
			Description: "根据ID获取单个自定义类型的详细信息，包括类型的所有字段定义。",
			URL:         "http://localhost:8080/api/custom-types/{id}",
			Method:      "GET",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "类型ID", Group: "input"},

				{Name: "id", Type: "number", Required: true, Description: "类型ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "类型名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "类型描述", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "UpdateCustomType",
			Description: "更新自定义类型。可以更新类型的任意字段，未提供的字段保持不变。如果提供了fields字段，将完全替换原有的字段列表。注意：如果该类型正在被接口参数引用，修改可能影响接口的使用。",
			URL:         "http://localhost:8080/api/custom-types/{id}",
			Method:      "PUT",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要更新的类型ID", Group: "input"},
				{Name: "name", Type: "string", Location: "body", Required: false, Description: "类型名称，在同一应用内必须唯一", Group: "input"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "类型描述", Group: "input"},
				{Name: "fields", Type: "custom", Location: "body", IsArray: true, Required: false, Description: "字段列表，如果提供则完全替换原有字段", Group: "input", Ref: &customTypeFieldTypeID},

				{Name: "id", Type: "number", Required: true, Description: "类型ID", Group: "output"},
				{Name: "app_id", Type: "number", Required: true, Description: "所属应用ID", Group: "output"},
				{Name: "name", Type: "string", Required: true, Description: "类型名称", Group: "output"},
				{Name: "description", Type: "string", Required: false, Description: "类型描述", Group: "output"},
				{Name: "created_at", Type: "string", Required: true, Description: "创建时间", Group: "output"},
				{Name: "updated_at", Type: "string", Required: true, Description: "更新时间", Group: "output"},
			},
		},
		{
			Name:        "DeleteCustomType",
			Description: "删除自定义类型。注意：如果该类型正在被接口参数或其他自定义类型引用，删除操作将失败。需要先解除所有引用才能删除。",
			URL:         "http://localhost:8080/api/custom-types/{id}",
			Method:      "DELETE",
			Parameters: []InterfaceParam{
				{Name: "id", Type: "number", Location: "path", Required: true, Description: "要删除的类型ID", Group: "input"},

				{Name: "success", Type: "boolean", Required: true, Description: "操作是否成功", Group: "output"},
				{Name: "message", Type: "string", Required: false, Description: "结果消息", Group: "output"},
			},
		},
	}

	// 批量创建接口及其参数
	for _, ifaceData := range interfaces {
		// 创建接口
		iface := models.Interface{
			AppID:       app.ID,
			Name:        ifaceData.Name,
			Description: ifaceData.Description,
			Protocol:    "http",
			URL:         ifaceData.URL,
			Method:      ifaceData.Method,
			AuthType:    "none",
			Enabled:     true,
		}

		if err := db.Create(&iface).Error; err != nil {
			log.Printf("Failed to create interface '%s': %v", ifaceData.Name, err)
			continue
		}

		log.Printf("Created interface '%s' with ID: %d", ifaceData.Name, iface.ID)

		// 创建接口参数
		for _, paramData := range ifaceData.Parameters {
			param := models.InterfaceParameter{
				AppID:        app.ID,
				InterfaceID:  iface.ID,
				Name:         paramData.Name,
				Type:         paramData.Type,
				Location:     paramData.Location,
				IsArray:      paramData.IsArray,
				Required:     paramData.Required,
				Description:  paramData.Description,
				DefaultValue: paramData.DefaultValue,
				Group:        paramData.Group,
				Ref:          paramData.Ref,
			}

			if err := db.Create(&param).Error; err != nil {
				log.Printf("Failed to create parameter '%s' for interface '%s': %v", paramData.Name, ifaceData.Name, err)
				continue
			}
		}

		if len(ifaceData.Parameters) > 0 {
			log.Printf("Created %d parameters for interface '%s'", len(ifaceData.Parameters), ifaceData.Name)
		}
	}

	log.Println("Default data initialization completed")
}
