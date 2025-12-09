package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCustomType(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "CustomTypeTestApp",
		Path:     "custom-type-test-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     CreateCustomTypeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功创建自定义类型",
			req: CreateCustomTypeRequest{
				AppID:       app.Application.ID,
				Name:        "User",
				Description: "User type",
				Fields: []CreateCustomTypeFieldReq{
					{
						Name:        "name",
						Type:        "string",
						Required:    true,
						Description: "User name",
					},
					{
						Name:        "age",
						Type:        "number",
						Required:    false,
						Description: "User age",
					},
					{
						Name:        "active",
						Type:        "boolean",
						Required:    true,
						Description: "Is active",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段name",
			req: CreateCustomTypeRequest{
				AppID: app.Application.ID,
			},
			wantErr: true,
		},
		{
			name: "无效的应用ID",
			req: CreateCustomTypeRequest{
				AppID: 99999,
				Name:  "InvalidApp",
			},
			wantErr: true,
			errMsg:  "application not found",
		},
		{
			name: "重复的类型名称",
			req: CreateCustomTypeRequest{
				AppID: app.Application.ID,
				Name:  "User",
			},
			wantErr: true,
			errMsg:  "duplicate custom type name",
		},
		{
			name: "创建带有数组字段的类型",
			req: CreateCustomTypeRequest{
				AppID:       app.Application.ID,
				Name:        "Post",
				Description: "Post type",
				Fields: []CreateCustomTypeFieldReq{
					{
						Name:        "title",
						Type:        "string",
						Required:    true,
						Description: "Post title",
					},
					{
						Name:        "tags",
						Type:        "string",
						IsArray:     true,
						Required:    false,
						Description: "Post tags",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := CreateCustomType(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotZero(t, resp.CustomType.ID)
				assert.Equal(t, tt.req.Name, resp.CustomType.Name)
				assert.Equal(t, tt.req.AppID, resp.CustomType.AppID)
				assert.Len(t, resp.CustomType.Fields, len(tt.req.Fields))
			}
		})
	}
}

func TestCreateCustomTypeWithReference(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "RefTypeApp",
		Path:     "ref-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建基础类型
	addressType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Address",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "street",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "city",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建引用其他类型的类型
	userType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "UserWithAddress",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "address",
				Type:     "custom",
				Ref:      int64Ptr(addressType.CustomType.ID),
				Required: true,
			},
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, userType.CustomType.ID)
	assert.Len(t, userType.CustomType.Fields, 2)

	// 验证引用字段
	addressField := userType.CustomType.Fields[1]
	assert.Equal(t, "address", addressField.Name)
	assert.Equal(t, "custom", addressField.Type)
	assert.NotNil(t, addressField.Ref)
	assert.Equal(t, addressType.CustomType.ID, *addressField.Ref)
}

func TestCreateCustomTypeInvalidReference(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "InvalidRefApp",
		Path:     "invalid-ref-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		fields  []CreateCustomTypeFieldReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "custom类型缺少引用",
			fields: []CreateCustomTypeFieldReq{
				{
					Name:     "field1",
					Type:     "custom",
					Required: true,
				},
			},
			wantErr: true,
			errMsg:  "field reference must be provided",
		},
		{
			name: "引用不存在的类型",
			fields: []CreateCustomTypeFieldReq{
				{
					Name:     "field1",
					Type:     "custom",
					Ref:      int64Ptr(99999),
					Required: true,
				},
			},
			wantErr: true,
			errMsg:  "custom type not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateCustomType(CreateCustomTypeRequest{
				AppID:  app.Application.ID,
				Name:   "TestType_" + tt.name,
				Fields: tt.fields,
			})
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateCustomTypeCrossAppReference(t *testing.T) {
	setupTestDB(t)

	// 创建两个应用
	app1, err := CreateApplication(CreateApplicationRequest{
		Name:     "App1",
		Path:     "app1",
		Protocol: "sse",
	})
	require.NoError(t, err)

	app2, err := CreateApplication(CreateApplicationRequest{
		Name:     "App2",
		Path:     "app2",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 在app1中创建类型
	type1, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app1.Application.ID,
		Name:  "Type1",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 尝试在app2中引用app1的类型（应该失败）
	_, err = CreateCustomType(CreateCustomTypeRequest{
		AppID: app2.Application.ID,
		Name:  "Type2",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "custom",
				Ref:      int64Ptr(type1.CustomType.ID),
				Required: true,
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field reference must belong to the same application")
}

func TestCreateCustomTypeCircularReference(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "CircularRefApp",
		Path:     "circular-ref-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建类型A
	typeA, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "TypeA",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建类型B，引用类型A
	typeB, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "TypeB",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "typeA",
				Type:     "custom",
				Ref:      int64Ptr(typeA.CustomType.ID),
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 尝试更新类型A，让它引用类型B（形成循环）
	_, err = UpdateCustomType(UpdateCustomTypeRequest{
		ID: typeA.CustomType.ID,
		Fields: &[]UpdateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "typeB",
				Type:     "custom",
				Ref:      int64Ptr(typeB.CustomType.ID),
				Required: true,
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular reference detected")
}

func TestGetCustomType(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用和类型
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "GetTypeApp",
		Path:     "get-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "GetTestType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     GetCustomTypeRequest
		wantErr bool
	}{
		{
			name: "成功获取类型",
			req: GetCustomTypeRequest{
				ID: customType.CustomType.ID,
			},
			wantErr: false,
		},
		{
			name: "获取不存在的类型",
			req: GetCustomTypeRequest{
				ID: 99999,
			},
			wantErr: true,
		},
		{
			name: "无效的ID",
			req: GetCustomTypeRequest{
				ID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GetCustomType(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.CustomType.ID)
				assert.Equal(t, "GetTestType", resp.CustomType.Name)
			}
		})
	}
}

func TestListCustomTypes(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "ListTypeApp",
		Path:     "list-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建多个类型
	types := []string{"Type1", "Type2", "Type3"}
	for _, name := range types {
		_, err := CreateCustomType(CreateCustomTypeRequest{
			AppID: app.Application.ID,
			Name:  name,
			Fields: []CreateCustomTypeFieldReq{
				{
					Name:     "field1",
					Type:     "string",
					Required: true,
				},
			},
		})
		require.NoError(t, err)
	}

	// 列出所有类型
	resp, err := ListCustomTypes(ListCustomTypesRequest{
		AppID: app.Application.ID,
	})
	require.NoError(t, err)
	assert.Len(t, resp.CustomTypes, 3)

	// 验证类型名称
	names := make(map[string]bool)
	for _, ct := range resp.CustomTypes {
		names[ct.Name] = true
	}
	for _, name := range types {
		assert.True(t, names[name])
	}
}

func TestListCustomTypesInvalidApp(t *testing.T) {
	setupTestDB(t)

	_, err := ListCustomTypes(ListCustomTypesRequest{
		AppID: 99999,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application not found")
}

func TestUpdateCustomType(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用和类型
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "UpdateTypeApp",
		Path:     "update-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID:       app.Application.ID,
		Name:        "UpdateTestType",
		Description: "Original description",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "old_field",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     UpdateCustomTypeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功更新名称",
			req: UpdateCustomTypeRequest{
				ID:   customType.CustomType.ID,
				Name: stringPtr("UpdatedType"),
			},
			wantErr: false,
		},
		{
			name: "成功更新描述",
			req: UpdateCustomTypeRequest{
				ID:          customType.CustomType.ID,
				Description: stringPtr("Updated description"),
			},
			wantErr: false,
		},
		{
			name: "更新不存在的类型",
			req: UpdateCustomTypeRequest{
				ID:   99999,
				Name: stringPtr("NonExistent"),
			},
			wantErr: true,
			errMsg:  "custom type not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := UpdateCustomType(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.CustomType.ID)
				if tt.req.Name != nil {
					assert.Equal(t, *tt.req.Name, resp.CustomType.Name)
				}
				if tt.req.Description != nil {
					assert.Equal(t, *tt.req.Description, resp.CustomType.Description)
				}
			}
		})
	}
}

func TestUpdateCustomTypeDuplicateName(t *testing.T) {
	setupTestDB(t)

	// 创建应用和两个类型
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "DuplicateTypeApp",
		Path:     "duplicate-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	type1, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Type1",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	_, err = CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Type2",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 尝试将Type1的名称改为Type2（重复）
	_, err = UpdateCustomType(UpdateCustomTypeRequest{
		ID:   type1.CustomType.ID,
		Name: stringPtr("Type2"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate custom type name")
}

func TestUpdateCustomTypeFields(t *testing.T) {
	setupTestDB(t)

	// 创建应用和类型
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "FieldUpdateApp",
		Path:     "field-update-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "FieldTestType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "old_field",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 更新字段列表
	newFields := []UpdateCustomTypeFieldReq{
		{
			Name:     "new_field1",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "new_field2",
			Type:     "number",
			Required: false,
		},
	}

	resp, err := UpdateCustomType(UpdateCustomTypeRequest{
		ID:     customType.CustomType.ID,
		Fields: &newFields,
	})
	require.NoError(t, err)
	assert.Len(t, resp.CustomType.Fields, 2)
	assert.Equal(t, "new_field1", resp.CustomType.Fields[0].Name)
	assert.Equal(t, "new_field2", resp.CustomType.Fields[1].Name)
}

func TestDeleteCustomType(t *testing.T) {
	setupTestDB(t)

	// 创建应用和类型
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "DeleteTypeApp",
		Path:     "delete-type-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "DeleteTestType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     DeleteCustomTypeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功删除类型",
			req: DeleteCustomTypeRequest{
				ID: customType.CustomType.ID,
			},
			wantErr: false,
		},
		{
			name: "删除不存在的类型",
			req: DeleteCustomTypeRequest{
				ID: 99999,
			},
			wantErr: true,
			errMsg:  "custom type not found",
		},
		{
			name: "无效的ID",
			req: DeleteCustomTypeRequest{
				ID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeleteCustomType(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				// 验证类型已被删除
				_, err := GetCustomType(GetCustomTypeRequest{ID: tt.req.ID})
				assert.Error(t, err)
			}
		})
	}
}

func TestDeleteCustomTypeReferencedByOtherType(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "RefDeleteApp",
		Path:     "ref-delete-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建基础类型
	baseType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "BaseType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建引用基础类型的类型
	_, err = CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "RefType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "base",
				Type:     "custom",
				Ref:      int64Ptr(baseType.CustomType.ID),
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 尝试删除被引用的类型
	_, err = DeleteCustomType(DeleteCustomTypeRequest{
		ID: baseType.CustomType.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "referenced by other type fields")
}

func TestDeleteCustomTypeReferencedByInterface(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "InterfaceRefApp",
		Path:     "interface-ref-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建自定义类型
	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "CustomType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "field1",
				Type:     "string",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建使用该类型的接口
	_, err = CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "InterfaceWithCustomType",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "POST",
		AuthType: "none",
		Parameters: []CreateInterfaceParameterReq{
			{
				Name:     "custom_param",
				Type:     "custom",
				Ref:      int64Ptr(customType.CustomType.ID),
				Location: "body",
				Group:    "input",
			},
		},
	})
	require.NoError(t, err)

	// 尝试删除被接口引用的类型
	_, err = DeleteCustomType(DeleteCustomTypeRequest{
		ID: customType.CustomType.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "referenced by interface parameters")
}

func TestComplexCustomTypeHierarchy(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "HierarchyApp",
		Path:     "hierarchy-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建地址类型
	addressType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Address",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "street",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "city",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "zipCode",
				Type:     "string",
				Required: false,
			},
		},
	})
	require.NoError(t, err)

	// 创建公司类型
	companyType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Company",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "address",
				Type:     "custom",
				Ref:      int64Ptr(addressType.CustomType.ID),
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建员工类型
	employeeType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Employee",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "age",
				Type:     "number",
				Required: false,
			},
			{
				Name:     "homeAddress",
				Type:     "custom",
				Ref:      int64Ptr(addressType.CustomType.ID),
				Required: true,
			},
			{
				Name:     "company",
				Type:     "custom",
				Ref:      int64Ptr(companyType.CustomType.ID),
				Required: false,
			},
		},
	})
	require.NoError(t, err)

	// 验证层级结构
	assert.NotZero(t, employeeType.CustomType.ID)
	assert.Len(t, employeeType.CustomType.Fields, 4)

	// 获取员工类型并验证
	resp, err := GetCustomType(GetCustomTypeRequest{
		ID: employeeType.CustomType.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Employee", resp.CustomType.Name)
	assert.Len(t, resp.CustomType.Fields, 4)
}

func TestCustomTypeWithArrayFields(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "ArrayFieldApp",
		Path:     "array-field-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建带有数组字段的类型
	customType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "ArrayType",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "tags",
				Type:     "string",
				IsArray:  true,
				Required: false,
			},
			{
				Name:     "scores",
				Type:     "number",
				IsArray:  true,
				Required: false,
			},
			{
				Name:     "flags",
				Type:     "boolean",
				IsArray:  true,
				Required: false,
			},
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, customType.CustomType.ID)
	assert.Len(t, customType.CustomType.Fields, 3)

	// 验证数组字段
	for _, field := range customType.CustomType.Fields {
		assert.True(t, field.IsArray)
	}
}

func TestCustomTypeWithNestedArrays(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "NestedArrayApp",
		Path:     "nested-array-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建基础类型
	itemType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Item",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "id",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "value",
				Type:     "number",
				Required: true,
			},
		},
	})
	require.NoError(t, err)

	// 创建包含自定义类型数组的类型
	containerType, err := CreateCustomType(CreateCustomTypeRequest{
		AppID: app.Application.ID,
		Name:  "Container",
		Fields: []CreateCustomTypeFieldReq{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "items",
				Type:     "custom",
				Ref:      int64Ptr(itemType.CustomType.ID),
				IsArray:  true,
				Required: false,
			},
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, containerType.CustomType.ID)
	assert.Len(t, containerType.CustomType.Fields, 2)

	// 验证数组字段
	itemsField := containerType.CustomType.Fields[1]
	assert.Equal(t, "items", itemsField.Name)
	assert.Equal(t, "custom", itemsField.Type)
	assert.True(t, itemsField.IsArray)
	assert.NotNil(t, itemsField.Ref)
	assert.Equal(t, itemType.CustomType.ID, *itemsField.Ref)
}
