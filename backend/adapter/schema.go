package adapter

import (
	"errors"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
)

func buildMcpSchemaByField(fieldId int64) (map[string]any, error) {
	db := database.GetDB()
	var field models.CustomTypeField
	if err := db.First(&field, fieldId).Error; err != nil {
		return nil, errors.New("custom type field not found")
	}
	schema := make(map[string]any)
	schema["description"] = field.Description
	if field.IsArray {
		schema["type"] = "array"
		if field.Type != "custom" {
			schema["items"] = map[string]any{"type": field.Type}
		} else {
			typeSchema, err := buildMcpSchemaByType(*field.Ref)
			if err != nil {
				return nil, err
			}
			schema["items"] = typeSchema
		}
	} else {
		if field.Type != "custom" {
			schema["type"] = field.Type
		} else {
			schema["type"] = "object"
			typeSchema, err := buildMcpSchemaByType(*field.Ref)
			if err != nil {
				return nil, err
			}
			schema["properties"] = typeSchema
		}
	}
	return schema, nil
}
func buildMcpSchemaByType(customTypeId int64) (map[string]any, error) {
	db := database.GetDB()
	var customType models.CustomType
	if err := db.First(&customType, customTypeId).Error; err != nil {
		return nil, errors.New("custom type not found")
	}
	// 获取字段列表
	var fields []models.CustomTypeField
	if db.Where("custom_type_id = ?", customType.ID).Find(&fields).Error != nil {
		return nil, errors.New("failed to fetch custom type fields")
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
		property, err := buildMcpSchemaByField(field.ID)
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
	db := database.GetDB()
	var iface models.Interface
	if err := db.First(&iface, id).Error; err != nil {
		return nil, errors.New("interface not found")
	}
	// 获取参数列表
	var params []models.InterfaceParameter
	if err := db.Where("interface_id = ?", iface.ID).Find(&params).Error; err != nil {
		return nil, errors.New("failed to fetch interface parameters")
	}
	schema := make(map[string]any)
	schema["type"] = "object"
	required := make([]string, 0)
	properties := make(map[string]any)
	for _, field := range params {
		property := make(map[string]any)
		property["description"] = field.Description
		if field.Required {
			required = append(required, field.Name)
		}
		if field.IsArray {
			property["type"] = "array"
			if field.Type != "custom" {
				property["items"] = map[string]any{"type": field.Type}
			} else {
				if field.Ref == nil {
					return nil, errors.New("custom type field is required")
				}
				typeSchema, err := buildMcpSchemaByType(*field.Ref)
				if err != nil {
					return nil, err
				}
				property["items"] = typeSchema
			}
		} else {
			if field.Type != "custom" {
				property["type"] = field.Type
			} else {
				property["type"] = "object"
				typeSchema, err := buildMcpSchemaByType(*field.Ref)
				if err != nil {
					return nil, err
				}
				property["properties"] = typeSchema
			}
		}
		properties[field.Name] = property
	}
	schema["required"] = required
	schema["properties"] = properties
	return schema, nil
}
