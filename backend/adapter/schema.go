package adapter

import (
	"errors"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"reflect"
)

// 递归深度限制，防止无限递归
const maxRecursionDepth = 4096

// schemaBuilder 用于构建schema的辅助结构，所有数据在构建前一次性加载到内存
type schemaBuilder struct {
	types  map[int64]*models.CustomType
	fields map[int64][]models.CustomTypeField
}

// buildContext 构建上下文，用于追踪递归深度
type buildContext struct {
	depth int // 当前递归深度
}

// newBuildContext 创建新的构建上下文
func newBuildContext() *buildContext {
	return &buildContext{
		depth: 0,
	}
}

// checkDepth 检查递归深度是否超限
func (ctx *buildContext) checkDepth() error {
	if ctx.depth >= maxRecursionDepth {
		return errors.New("maximum recursion depth exceeded")
	}
	return nil
}

// next 创建下一层递归的上下文
func (ctx *buildContext) next() *buildContext {
	return &buildContext{
		depth: ctx.depth + 1,
	}
}

// newSchemaBuilder 创建新的schema构建器并加载指定应用的数据
func newSchemaBuilder(appId int64) (*schemaBuilder, error) {
	db := database.GetDB()

	// 一次性加载指定应用的自定义类型
	var allTypes []models.CustomType
	if err := db.Where("app_id = ?", appId).Find(&allTypes).Error; err != nil {
		return nil, errors.New("failed to load custom types")
	}

	// 一次性加载指定应用的自定义类型字段
	var allFields []models.CustomTypeField
	if err := db.Where("app_id = ?", appId).Find(&allFields).Error; err != nil {
		return nil, errors.New("failed to load custom type fields")
	}

	// 构建内存索引
	types := make(map[int64]*models.CustomType)
	for i := range allTypes {
		types[allTypes[i].ID] = &allTypes[i]
	}

	fields := make(map[int64][]models.CustomTypeField)
	for _, field := range allFields {
		fields[field.CustomTypeID] = append(fields[field.CustomTypeID], field)
	}

	return &schemaBuilder{
		types:  types,
		fields: fields,
	}, nil
}

// getCustomType 从内存获取自定义类型
func (sb *schemaBuilder) getCustomType(typeId int64) (*models.CustomType, error) {
	customType, ok := sb.types[typeId]
	if !ok {
		return nil, errors.New("custom type not found")
	}
	return customType, nil
}

// getCustomTypeFields 从内存获取自定义类型的所有字段
func (sb *schemaBuilder) getCustomTypeFields(typeId int64) ([]models.CustomTypeField, error) {
	fields, ok := sb.fields[typeId]
	if !ok {
		return []models.CustomTypeField{}, nil
	}
	return fields, nil
}

// buildSchemaByField 根据字段构建schema
func (sb *schemaBuilder) buildSchemaByField(field *models.CustomTypeField, ctx *buildContext) (map[string]any, error) {
	schema := make(map[string]any)
	schema["description"] = field.Description

	if field.IsArray {
		schema["type"] = "array"
		items, err := sb.buildFieldTypeSchema(field, ctx)
		if err != nil {
			return nil, err
		}
		schema["items"] = items
	} else {
		// 非数组类型，直接返回类型schema
		return sb.buildFieldTypeSchema(field, ctx)
	}

	return schema, nil
}

// buildFieldTypeSchema 构建字段的类型schema（不包含数组包装）
func (sb *schemaBuilder) buildFieldTypeSchema(field *models.CustomTypeField, ctx *buildContext) (map[string]any, error) {
	if field.Type != "custom" {
		// 基础类型
		return map[string]any{
			"type":        field.Type,
			"description": field.Description,
		}, nil
	}

	// 自定义类型
	if field.Ref == nil {
		return nil, errors.New("custom type field ref is required")
	}

	// 直接返回自定义类型的完整schema
	return sb.buildSchemaByType(*field.Ref, ctx)
}

// buildSchemaByType 根据自定义类型ID构建完整的schema
func (sb *schemaBuilder) buildSchemaByType(customTypeId int64, ctx *buildContext) (map[string]any, error) {
	// 检查递归深度
	if err := ctx.checkDepth(); err != nil {
		return nil, err
	}

	customType, err := sb.getCustomType(customTypeId)
	if err != nil {
		return nil, err
	}

	fields, err := sb.getCustomTypeFields(customType.ID)
	if err != nil {
		return nil, err
	}

	schema := make(map[string]any)
	schema["type"] = "object"
	schema["description"] = customType.Description

	required := make([]string, 0)
	properties := make(map[string]any)

	// 创建下一层递归的上下文
	newCtx := ctx.next()

	for _, field := range fields {
		if field.Required {
			required = append(required, field.Name)
		}

		property, err := sb.buildSchemaByField(&field, newCtx)
		if err != nil {
			return nil, err
		}
		properties[field.Name] = property
	}

	schema["required"] = required
	schema["properties"] = properties
	return schema, nil
}

func BuildMcpInputSchemaByInterface(id int64) (map[string]any, error) {
	return buildMcpSchemaByInterface(id, "input")
}

func BuildMcpOutputSchemaByInterface(id int64) (map[string]any, error) {
	return buildMcpSchemaByInterface(id, "output")
}

func buildMcpSchemaByInterface(id int64, group string) (map[string]any, error) {
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, id).Error; err != nil {
		return nil, errors.New("interface not found")
	}

	// 获取参数列表，只包含指定组的参数
	var params []models.InterfaceParameter
	if err := db.Where("interface_id = ? AND `group` = ?", iface.ID, group).Find(&params).Error; err != nil {
		return nil, errors.New("failed to fetch interface parameters")
	}

	// 创建schema构建器，一次性加载指定应用的数据到内存
	builder, err := newSchemaBuilder(iface.AppID)
	if err != nil {
		return nil, err
	}

	// 创建构建上下文，用于追踪递归深度和环形引用
	ctx := newBuildContext()

	schema := make(map[string]any)
	schema["type"] = "object"
	required := make([]string, 0)
	properties := make(map[string]any)

	for _, param := range params {
		// 有默认值的非数组基础类型参数不需要用户输入，跳过
		if param.DefaultValue != nil &&
			*param.DefaultValue != "" &&
			!param.IsArray && param.Type != "custom" {
			continue
		}

		if param.Required {
			required = append(required, param.Name)
		}

		property, err := builder.buildSchemaByParameter(&param, ctx)
		if err != nil {
			return nil, err
		}
		properties[param.Name] = property
	}

	schema["required"] = required
	schema["properties"] = properties
	return schema, nil
}

// buildSchemaByParameter 根据参数构建schema
func (sb *schemaBuilder) buildSchemaByParameter(param *models.InterfaceParameter, ctx *buildContext) (map[string]any, error) {
	schema := make(map[string]any)
	schema["description"] = param.Description

	if param.IsArray {
		schema["type"] = "array"
		items, err := sb.buildParameterTypeSchema(param, ctx)
		if err != nil {
			return nil, err
		}
		schema["items"] = items
	} else {
		// 非数组类型，直接返回类型schema
		return sb.buildParameterTypeSchema(param, ctx)
	}

	return schema, nil
}

// buildParameterTypeSchema 构建参数的类型schema（不包含数组包装）
func (sb *schemaBuilder) buildParameterTypeSchema(param *models.InterfaceParameter, ctx *buildContext) (map[string]any, error) {
	if param.Type != "custom" {
		// 基础类型
		return map[string]any{
			"type":        param.Type,
			"description": param.Description,
		}, nil
	}

	// 自定义类型
	if param.Ref == nil {
		return nil, errors.New("custom type parameter ref is required")
	}

	// 直接返回自定义类型的完整schema
	return sb.buildSchemaByType(*param.Ref, ctx)
}

// SatisfySchema 验证数据是否满足schema定义
func SatisfySchema(schema map[string]any, data any) bool {
	if schema == nil {
		return true
	}

	// 检查schema树中是否存在任何required字段
	var hasAnyRequired func(s map[string]any, depth int) bool
	hasAnyRequired = func(s map[string]any, depth int) bool {
		if s == nil || depth > maxRecursionDepth {
			return false
		}

		// 检查当前层级的required字段
		if required, ok := s["required"].([]any); ok && len(required) > 0 {
			return true
		}
		if required, ok := s["required"].([]string); ok && len(required) > 0 {
			return true
		}

		// 递归检查properties中的嵌套对象
		if properties, ok := s["properties"].(map[string]any); ok {
			for _, propSchema := range properties {
				if propSchemaMap, ok := propSchema.(map[string]any); ok {
					if hasAnyRequired(propSchemaMap, depth+1) {
						return true
					}
				}
			}
		}

		// 检查array的items
		if items, ok := s["items"].(map[string]any); ok {
			if hasAnyRequired(items, depth+1) {
				return true
			}
		}

		return false
	}

	// 如果数据是nil，检查schema中是否有任何required字段
	if isNil(data) {
		if hasAnyRequired(schema, 0) {
			return false
		}
		// 如果没有required字段，nil数据是有效的
		return true
	}

	var dfs func(schema, data any, depth int, checkRequired bool) bool
	dfs = func(schema, data any, depth int, checkRequired bool) bool {

		if depth > maxRecursionDepth {
			log.Printf("Warning: schema validation exceeded max depth %d", maxRecursionDepth)
			return false
		}

		if isNil(schema) && isNil(data) {
			return true
		}
		if isNil(schema) && !isNil(data) {
			return false
		}

		left, converted := schema.(map[string]any)
		if !converted {
			return false
		}

		schemaType, ok := left["type"].(string)
		if !ok {
			return false
		}

		switch schemaType {
		case "string":
			// 基本类型不接受nil
			if isNil(data) {
				return false
			}
			if _, ok := data.(string); !ok {
				return false
			}
			return true

		case "number":
			// 基本类型不接受nil
			if isNil(data) {
				return false
			}
			switch data.(type) {
			case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
				return true
			default:
				return false
			}

		case "boolean":
			// 基本类型不接受nil
			if isNil(data) {
				return false
			}
			if _, ok := data.(bool); !ok {
				return false
			}
			return true

		case "array":
			// array类型可以接受nil（表示空数组）
			if isNil(data) {
				return true
			}

			// 使用反射支持所有类型的切片和数组
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Slice && dataValue.Kind() != reflect.Array {
				return false
			}
			items, converted := left["items"].(map[string]any)
			if !converted || items == nil {
				return false
			}

			checked := false
			for i := 0; i < dataValue.Len(); i++ {
				item := dataValue.Index(i).Interface()
				if isNil(item) && checked {
					continue
				}
				if !dfs(items, item, depth+1, checkRequired) {
					return false
				}
				if isNil(item) {
					checked = true
				}
			}
			return true

		case "object":
			properties, converted := left["properties"].(map[string]any)
			if !converted || properties == nil {
				return false
			}

			if isNil(data) {
				if checkRequired && hasAnyRequired(left, depth) {
					return false
				}
				return true
			}

			// 使用反射支持所有类型的map
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Map {
				return false
			}
			// 检查map是否为nil
			if dataValue.IsNil() {
				// nil map无法包含任何字段，递归检查整个schema树中是否存在任何required字段
				if checkRequired && hasAnyRequired(left, depth) {
					return false
				}
				return true
			}

			// 处理required字段
			requiredMap := make(map[string]bool)
			if required, ok := left["required"].([]any); ok {
				for _, req := range required {
					if reqStr, ok := req.(string); ok {
						requiredMap[reqStr] = true
					}
				}
			}
			if required, ok := left["required"].([]string); ok {
				for _, req := range required {
					requiredMap[req] = true
				}
			}
			// 首先检查所有必填字段是否存在
			for requiredKey := range requiredMap {
				mapKey := reflect.ValueOf(requiredKey)
				mapValue := dataValue.MapIndex(mapKey)
				if !mapValue.IsValid() {
					// 必填字段不存在
					return false
				}
			}

			// 验证每个属性
			for key, prop := range properties {
				// 使用反射获取map中的值
				mapKey := reflect.ValueOf(key)
				mapValue := dataValue.MapIndex(mapKey)

				// 如果字段不存在
				if !mapValue.IsValid() {
					// 如果是必填字段，已经在上面检查过了
					// 如果不是必填字段，跳过验证
					if !requiredMap[key] {
						continue
					}
				}

				// 递归验证字段
				var fieldResult bool
				if mapValue.IsValid() {
					// 检查值是否为nil（针对map、slice、pointer等可为nil的类型）
					fieldValue := mapValue.Interface()
					fieldResult = dfs(prop, fieldValue, depth+1, checkRequired)
				} else {
					fieldResult = dfs(prop, nil, depth+1, checkRequired)
				}

				// 如果验证失败，直接返回false
				if !fieldResult {
					return false
				}
			}
			return true

		default:
			return false
		}
	}

	result := dfs(schema, data, 0, true)
	return result
}

func FilterDataBySchema(schema map[string]any, data any) any {
	if schema == nil {
		return data
	}
	if isNil(data) {
		return nil
	}

	// 递归过滤函数
	var filter func(schema, data any) any
	filter = func(schema, data any) any {
		if isNil(schema) || isNil(data) {
			return data
		}

		// schema必须是map[string]any类型
		schemaMap, ok := schema.(map[string]any)
		if !ok {
			return data
		}

		schemaType, ok := schemaMap["type"].(string)
		if !ok {
			return data
		}

		switch schemaType {
		case "string":
			if _, ok := data.(string); !ok {
				return nil
			}
			return data
		case "number":
			// 支持多种数字类型
			switch data.(type) {
			case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
				return data
			default:
				return nil
			}
		case "boolean":
			if _, ok := data.(bool); !ok {
				return nil
			}
			return data
		case "array":
			// 处理数组类型 - 使用反射支持所有类型的切片
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Slice && dataValue.Kind() != reflect.Array {
				return nil
			}
			items, ok := schemaMap["items"]
			if !ok {
				return data
			}
			// 递归过滤数组中的每个元素
			filtered := make([]any, 0, dataValue.Len())
			for i := 0; i < dataValue.Len(); i++ {
				item := dataValue.Index(i).Interface()
				t := filter(items, item)
				if t != nil {
					filtered = append(filtered, t)
				}
			}
			return filtered

		case "object":
			// 处理对象类型 - 使用反射支持所有类型的map
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Map {
				return nil
			}
			properties, ok := schemaMap["properties"].(map[string]any)
			if !ok {
				return data
			}
			// 只保留schema中定义的字段
			filtered := make(map[string]any)
			for key, propSchema := range properties {
				// 使用反射获取map中的值
				mapKey := reflect.ValueOf(key)
				mapValue := dataValue.MapIndex(mapKey)
				if mapValue.IsValid() {
					value := mapValue.Interface()
					t := filter(propSchema, value)
					if !isNil(t) {
						filtered[key] = t
					}
				}
			}
			return filtered
		default:
			return data
		}
	}

	return filter(schema, data)
}

func isNil(data any) bool {
	if data == nil {
		return true
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Chan, reflect.Func,
		reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		// default case, not a nil-able type
	}
	return false
}
