package adapter

import (
	"errors"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"reflect"
)

// Schema validation result constants
const (
	validationMatched = iota
	validationSchemaNotExist
	validationDataNotExist
	validationTypeNotMatch
)

// schemaBuilder 用于构建schema的辅助结构，所有数据在构建前一次性加载到内存
type schemaBuilder struct {
	types  map[int64]*models.CustomType
	fields map[int64][]models.CustomTypeField
}

// newSchemaBuilder 创建新的schema构建器并加载所有数据
func newSchemaBuilder() (*schemaBuilder, error) {
	db := database.GetDB()

	// 一次性加载所有自定义类型
	var allTypes []models.CustomType
	if err := db.Find(&allTypes).Error; err != nil {
		return nil, errors.New("failed to load custom types")
	}

	// 一次性加载所有自定义类型字段
	var allFields []models.CustomTypeField
	if err := db.Find(&allFields).Error; err != nil {
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
func (sb *schemaBuilder) buildSchemaByField(field *models.CustomTypeField) (map[string]any, error) {
	schema := make(map[string]any)
	schema["description"] = field.Description

	if field.IsArray {
		schema["type"] = "array"
		items, err := sb.buildFieldTypeSchema(field)
		if err != nil {
			return nil, err
		}
		schema["items"] = items
	} else {
		// 非数组类型，直接返回类型schema
		return sb.buildFieldTypeSchema(field)
	}

	return schema, nil
}

// buildFieldTypeSchema 构建字段的类型schema（不包含数组包装）
func (sb *schemaBuilder) buildFieldTypeSchema(field *models.CustomTypeField) (map[string]any, error) {
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
	return sb.buildSchemaByType(*field.Ref)
}

// buildSchemaByType 根据自定义类型ID构建完整的object schema
func (sb *schemaBuilder) buildSchemaByType(customTypeId int64) (map[string]any, error) {
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

	for _, field := range fields {
		if field.Required {
			required = append(required, field.Name)
		}

		property, err := sb.buildSchemaByField(&field)
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

	// 创建schema构建器，一次性加载所有数据到内存
	builder, err := newSchemaBuilder()
	if err != nil {
		return nil, err
	}

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

		property, err := builder.buildSchemaByParameter(&param)
		if err != nil {
			return nil, err
		}
		properties[param.Name] = property
	}

	schema["required"] = required
	schema["properties"] = properties
	return schema, nil
}

// buildSchemaByParameter 根据接口参数构建schema
func (sb *schemaBuilder) buildSchemaByParameter(param *models.InterfaceParameter) (map[string]any, error) {
	schema := make(map[string]any)
	schema["description"] = param.Description

	if param.IsArray {
		schema["type"] = "array"
		items, err := sb.buildParameterTypeSchema(param)
		if err != nil {
			return nil, err
		}
		schema["items"] = items
	} else {
		// 非数组类型，直接返回类型schema
		return sb.buildParameterTypeSchema(param)
	}

	return schema, nil
}

// buildParameterTypeSchema 构建参数的类型schema（不包含数组包装）
func (sb *schemaBuilder) buildParameterTypeSchema(param *models.InterfaceParameter) (map[string]any, error) {
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
	return sb.buildSchemaByType(*param.Ref)
}

func SatisfySchema(schema map[string]any, data any) bool {
	if schema == nil {
		return true
	}
	if isNil(data) {
		data = make(map[string]any)
	}
	var dfs func(schema, data any, depth int) int
	dfs = func(schema, data any, depth int) int {
		// 防止无限递归
		const maxDepth = 100
		if depth > maxDepth {
			log.Printf("Warning: schema validation exceeded max depth %d", maxDepth)
			return validationSchemaNotExist
		}

		if isNil(schema) && isNil(data) {
			return validationMatched
		}
		if !isNil(schema) && isNil(data) {
			return validationDataNotExist
		}
		if isNil(schema) && !isNil(data) {
			return validationSchemaNotExist
		}

		left, converted := schema.(map[string]any)
		if !converted {
			return validationSchemaNotExist
		}

		schemaType, ok := left["type"].(string)
		if !ok {
			return validationSchemaNotExist
		}

		switch schemaType {
		case "string":
			if _, ok := data.(string); !ok {
				return validationTypeNotMatch
			}
			return validationMatched

		case "number":
			switch data.(type) {
			case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
				return validationMatched
			default:
				return validationTypeNotMatch
			}

		case "boolean":
			if _, ok := data.(bool); !ok {
				return validationTypeNotMatch
			}
			return validationMatched

		case "array":
			// 使用反射支持所有类型的切片和数组
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Slice && dataValue.Kind() != reflect.Array {
				return validationTypeNotMatch
			}
			items, converted := left["items"].(map[string]any)
			if !converted || items == nil {
				return validationSchemaNotExist
			}

			result := validationMatched
			for i := 0; i < dataValue.Len(); i++ {
				item := dataValue.Index(i).Interface()
				itemResult := dfs(items, item, depth+1)
				if itemResult > result {
					result = itemResult
				}
			}
			return result

		case "object":
			properties, converted := left["properties"].(map[string]any)
			if !converted || properties == nil {
				return validationSchemaNotExist
			}

			// 使用反射支持所有类型的map
			dataValue := reflect.ValueOf(data)
			if dataValue.Kind() != reflect.Map {
				return validationTypeNotMatch
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
					return validationTypeNotMatch
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
				var fieldResult int
				if mapValue.IsValid() {
					fieldResult = dfs(prop, mapValue.Interface(), depth+1)
				} else {
					fieldResult = dfs(prop, nil, depth+1)
				}

				// 任何类型不匹配都应该返回错误
				if fieldResult == validationTypeNotMatch {
					return validationTypeNotMatch
				}

				// 如果字段存在但数据为nil，且该字段是必填的
				if fieldResult == validationDataNotExist && requiredMap[key] {
					return validationTypeNotMatch
				}
			}
			return validationMatched

		default:
			return validationTypeNotMatch
		}
	}

	result := dfs(schema, data, 0)
	// 使用debug级别日志,生产环境可以通过日志级别控制
	if result != validationMatched {
		log.Printf("SatisfySchema validation failed: result=%d, schema=%+v, data=%+v", result, schema, data)
	}
	return result == validationMatched
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
