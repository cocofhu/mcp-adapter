package service

import (
	"mcp-adapter/backend/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// int64Ptr 返回int64指针
func int64Ptr(i int64) *int64 {
	return &i
}

func TestCreateInterface(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "InterfaceTestApp",
		Path:     "interface-test-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     CreateInterfaceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功创建接口",
			req: CreateInterfaceRequest{
				AppID:       app.Application.ID,
				Name:        "TestInterface",
				Description: "Test interface",
				Protocol:    "http",
				URL:         "https://api.example.com/test",
				Method:      "GET",
				AuthType:    "none",
				Enabled:     true,
				Parameters: []CreateInterfaceParameterReq{
					{
						Name:        "param1",
						Type:        "string",
						Location:    "query",
						Required:    true,
						Description: "Test parameter",
						Group:       "input",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段name",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "GET",
				AuthType: "none",
			},
			wantErr: true,
		},
		{
			name: "无效的应用ID",
			req: CreateInterfaceRequest{
				AppID:    99999,
				Name:     "InvalidAppInterface",
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "GET",
				AuthType: "none",
			},
			wantErr: true,
			errMsg:  "application not found",
		},
		{
			name: "重复的接口名称",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "TestInterface",
				Protocol: "http",
				URL:      "https://api.example.com/test2",
				Method:   "POST",
				AuthType: "none",
			},
			wantErr: true,
			errMsg:  "interface name already exists",
		},
		{
			name: "无效的HTTP方法",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "InvalidMethodInterface",
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "INVALID",
				AuthType: "none",
			},
			wantErr: true,
		},
		{
			name: "无效的鉴权类型",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "InvalidAuthInterface",
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "GET",
				AuthType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "创建带有多个参数的接口",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "MultiParamInterface",
				Protocol: "http",
				URL:      "https://api.example.com/multi",
				Method:   "POST",
				AuthType: "none",
				Parameters: []CreateInterfaceParameterReq{
					{
						Name:        "query_param",
						Type:        "string",
						Location:    "query",
						Required:    true,
						Description: "Query parameter",
						Group:       "input",
					},
					{
						Name:        "header_param",
						Type:        "string",
						Location:    "header",
						Required:    false,
						Description: "Header parameter",
						Group:       "input",
					},
					{
						Name:        "body_param",
						Type:        "number",
						Location:    "body",
						Required:    true,
						Description: "Body parameter",
						Group:       "input",
					},
					{
						Name:        "output_param",
						Type:        "string",
						Location:    "body",
						Required:    false,
						Description: "Output parameter",
						Group:       "output",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "创建带有fixed参数的接口",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "FixedParamInterface",
				Protocol: "http",
				URL:      "https://api.example.com/fixed",
				Method:   "GET",
				AuthType: "none",
				Parameters: []CreateInterfaceParameterReq{
					{
						Name:         "fixed_param",
						Type:         "string",
						Location:     "query",
						Required:     true,
						Description:  "Fixed parameter",
						Group:        "fixed",
						DefaultValue: stringPtr("fixed_value"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "创建CAPI鉴权接口",
			req: CreateInterfaceRequest{
				AppID:    app.Application.ID,
				Name:     "CAPIInterface",
				Protocol: "http",
				URL:      "https://api.example.com/capi",
				Method:   "POST",
				AuthType: "capi",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := CreateInterface(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotZero(t, resp.Interface.ID)
				assert.Equal(t, tt.req.Name, resp.Interface.Name)
				assert.Equal(t, tt.req.AppID, resp.Interface.AppID)
				assert.Equal(t, tt.req.Protocol, resp.Interface.Protocol)
				assert.Equal(t, tt.req.URL, resp.Interface.URL)
				assert.Equal(t, tt.req.Method, resp.Interface.Method)
				assert.Equal(t, tt.req.AuthType, resp.Interface.AuthType)
				assert.Len(t, resp.Interface.Parameters, len(tt.req.Parameters))
			}
		})
	}
}

func TestCreateInterfaceWithInvalidParameters(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "ParamTestApp",
		Path:     "param-test-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		params  []CreateInterfaceParameterReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "output参数不能有默认值",
			params: []CreateInterfaceParameterReq{
				{
					Name:         "output_param",
					Type:         "string",
					Location:     "body",
					Group:        "output",
					DefaultValue: stringPtr("invalid"),
				},
			},
			wantErr: true,
			errMsg:  "output parameters cannot have default values",
		},
		{
			name: "数组类型不能有默认值",
			params: []CreateInterfaceParameterReq{
				{
					Name:         "array_param",
					Type:         "string",
					Location:     "body",
					Group:        "input",
					IsArray:      true,
					DefaultValue: stringPtr("invalid"),
				},
			},
			wantErr: true,
			errMsg:  "array type parameters cannot have default values",
		},
		{
			name: "fixed参数必须有默认值",
			params: []CreateInterfaceParameterReq{
				{
					Name:     "fixed_param",
					Type:     "string",
					Location: "query",
					Group:    "fixed",
				},
			},
			wantErr: true,
			errMsg:  "fixed parameter must have a default value",
		},
		{
			name: "fixed参数不能是数组（有默认值）",
			params: []CreateInterfaceParameterReq{
				{
					Name:         "fixed_array",
					Type:         "string",
					Location:     "query",
					Group:        "fixed",
					IsArray:      true,
					DefaultValue: stringPtr("value"),
				},
			},
			wantErr: true,
			errMsg:  "array type parameters cannot have default values",
		},
		{
			name: "fixed参数不能是数组（无默认值）",
			params: []CreateInterfaceParameterReq{
				{
					Name:     "fixed_array_no_default",
					Type:     "string",
					Location: "query",
					Group:    "fixed",
					IsArray:  true,
				},
			},
			wantErr: true,
			errMsg:  "fixed parameter must have a default value",
		},
		{
			name: "number类型默认值不匹配",
			params: []CreateInterfaceParameterReq{
				{
					Name:         "number_param",
					Type:         "number",
					Location:     "query",
					Group:        "input",
					DefaultValue: stringPtr("not_a_number"),
				},
			},
			wantErr: true,
			errMsg:  "default value does not match parameter type 'number'",
		},
		{
			name: "boolean类型默认值不匹配",
			params: []CreateInterfaceParameterReq{
				{
					Name:         "bool_param",
					Type:         "boolean",
					Location:     "query",
					Group:        "input",
					DefaultValue: stringPtr("invalid_bool"),
				},
			},
			wantErr: true,
			errMsg:  "default value does not match parameter type 'boolean'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateInterface(CreateInterfaceRequest{
				AppID:      app.Application.ID,
				Name:       "TestInterface_" + tt.name,
				Protocol:   "http",
				URL:        "https://api.example.com/test",
				Method:     "GET",
				AuthType:   "none",
				Parameters: tt.params,
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

func TestGetInterface(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用和接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "GetInterfaceApp",
		Path:     "get-interface-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "GetTestInterface",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     GetInterfaceRequest
		wantErr bool
	}{
		{
			name: "成功获取接口",
			req: GetInterfaceRequest{
				ID: iface.Interface.ID,
			},
			wantErr: false,
		},
		{
			name: "获取不存在的接口",
			req: GetInterfaceRequest{
				ID: 99999,
			},
			wantErr: true,
		},
		{
			name: "无效的ID",
			req: GetInterfaceRequest{
				ID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GetInterface(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.Interface.ID)
				assert.Equal(t, "GetTestInterface", resp.Interface.Name)
			}
		})
	}
}

func TestListInterfaces(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "ListInterfaceApp",
		Path:     "list-interface-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建多个接口
	interfaces := []string{"Interface1", "Interface2", "Interface3"}
	for _, name := range interfaces {
		_, err := CreateInterface(CreateInterfaceRequest{
			AppID:    app.Application.ID,
			Name:     name,
			Protocol: "http",
			URL:      "https://api.example.com/" + name,
			Method:   "GET",
			AuthType: "none",
		})
		require.NoError(t, err)
	}

	// 列出所有接口
	resp, err := ListInterfaces(ListInterfacesRequest{
		AppID: app.Application.ID,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Interfaces, 3)

	// 验证接口名称
	names := make(map[string]bool)
	for _, iface := range resp.Interfaces {
		names[iface.Name] = true
	}
	for _, name := range interfaces {
		assert.True(t, names[name])
	}
}

func TestListInterfacesInvalidApp(t *testing.T) {
	setupTestDB(t)

	_, err := ListInterfaces(ListInterfacesRequest{
		AppID: 99999,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application not found")
}

func TestUpdateInterface(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用和接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "UpdateInterfaceApp",
		Path:     "update-interface-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:       app.Application.ID,
		Name:        "UpdateTestInterface",
		Description: "Original description",
		Protocol:    "http",
		URL:         "https://api.example.com/test",
		Method:      "GET",
		AuthType:    "none",
		Enabled:     true,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     UpdateInterfaceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功更新名称",
			req: UpdateInterfaceRequest{
				ID:   iface.Interface.ID,
				Name: stringPtr("UpdatedInterface"),
			},
			wantErr: false,
		},
		{
			name: "成功更新描述",
			req: UpdateInterfaceRequest{
				ID:          iface.Interface.ID,
				Description: stringPtr("Updated description"),
			},
			wantErr: false,
		},
		{
			name: "成功更新URL",
			req: UpdateInterfaceRequest{
				ID:  iface.Interface.ID,
				URL: stringPtr("https://api.example.com/updated"),
			},
			wantErr: false,
		},
		{
			name: "成功更新方法",
			req: UpdateInterfaceRequest{
				ID:     iface.Interface.ID,
				Method: stringPtr("POST"),
			},
			wantErr: false,
		},
		{
			name: "成功更新鉴权类型",
			req: UpdateInterfaceRequest{
				ID:       iface.Interface.ID,
				AuthType: stringPtr("capi"),
			},
			wantErr: false,
		},
		{
			name: "成功禁用接口",
			req: UpdateInterfaceRequest{
				ID:      iface.Interface.ID,
				Enabled: boolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "更新不存在的接口",
			req: UpdateInterfaceRequest{
				ID:   99999,
				Name: stringPtr("NonExistent"),
			},
			wantErr: true,
			errMsg:  "interface not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := UpdateInterface(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.Interface.ID)
				if tt.req.Name != nil {
					assert.Equal(t, *tt.req.Name, resp.Interface.Name)
				}
				if tt.req.Description != nil {
					assert.Equal(t, *tt.req.Description, resp.Interface.Description)
				}
				if tt.req.URL != nil {
					assert.Equal(t, *tt.req.URL, resp.Interface.URL)
				}
				if tt.req.Method != nil {
					assert.Equal(t, *tt.req.Method, resp.Interface.Method)
				}
				if tt.req.AuthType != nil {
					assert.Equal(t, *tt.req.AuthType, resp.Interface.AuthType)
				}
				if tt.req.Enabled != nil {
					assert.Equal(t, *tt.req.Enabled, resp.Interface.Enabled)
				}
			}
		})
	}
}

func TestUpdateInterfaceDuplicateName(t *testing.T) {
	setupTestDB(t)

	// 创建应用和两个接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "DuplicateNameApp",
		Path:     "duplicate-name-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface1, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "Interface1",
		Protocol: "http",
		URL:      "https://api.example.com/1",
		Method:   "GET",
		AuthType: "none",
	})
	require.NoError(t, err)

	_, err = CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "Interface2",
		Protocol: "http",
		URL:      "https://api.example.com/2",
		Method:   "GET",
		AuthType: "none",
	})
	require.NoError(t, err)

	// 尝试将Interface1的名称改为Interface2（重复）
	_, err = UpdateInterface(UpdateInterfaceRequest{
		ID:   iface1.Interface.ID,
		Name: stringPtr("Interface2"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interface name already exists")
}

func TestUpdateInterfaceParameters(t *testing.T) {
	setupTestDB(t)

	// 创建应用和接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "ParamUpdateApp",
		Path:     "param-update-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "ParamInterface",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
		Parameters: []CreateInterfaceParameterReq{
			{
				Name:     "old_param",
				Type:     "string",
				Location: "query",
				Group:    "input",
			},
		},
	})
	require.NoError(t, err)

	// 更新参数列表
	newParams := []CreateInterfaceParameterReq{
		{
			Name:     "new_param1",
			Type:     "string",
			Location: "query",
			Group:    "input",
		},
		{
			Name:     "new_param2",
			Type:     "number",
			Location: "body",
			Group:    "input",
		},
	}

	resp, err := UpdateInterface(UpdateInterfaceRequest{
		ID:         iface.Interface.ID,
		Parameters: &newParams,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Interface.Parameters, 2)
	assert.Equal(t, "new_param1", resp.Interface.Parameters[0].Name)
	assert.Equal(t, "new_param2", resp.Interface.Parameters[1].Name)
}

func TestDeleteInterface(t *testing.T) {
	setupTestDB(t)

	// 创建应用和接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "DeleteInterfaceApp",
		Path:     "delete-interface-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "DeleteTestInterface",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     DeleteInterfaceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功删除接口",
			req: DeleteInterfaceRequest{
				ID: iface.Interface.ID,
			},
			wantErr: false,
		},
		{
			name: "删除不存在的接口",
			req: DeleteInterfaceRequest{
				ID: 99999,
			},
			wantErr: true,
			errMsg:  "interface not found",
		},
		{
			name: "无效的ID",
			req: DeleteInterfaceRequest{
				ID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeleteInterface(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				// 验证接口已被删除
				_, err := GetInterface(GetInterfaceRequest{ID: tt.req.ID})
				assert.Error(t, err)
			}
		})
	}
}

func TestDeleteInterfaceWithParameters(t *testing.T) {
	setupTestDB(t)

	// 创建应用和带参数的接口
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "DeleteParamApp",
		Path:     "delete-param-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "InterfaceWithParams",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
		Parameters: []CreateInterfaceParameterReq{
			{
				Name:     "param1",
				Type:     "string",
				Location: "query",
				Group:    "input",
			},
			{
				Name:     "param2",
				Type:     "number",
				Location: "body",
				Group:    "input",
			},
		},
	})
	require.NoError(t, err)

	// 删除接口
	_, err = DeleteInterface(DeleteInterfaceRequest{
		ID: iface.Interface.ID,
	})
	require.NoError(t, err)

	// 验证参数也被删除
	db := database.GetDB()
	var count int64
	db.Model(&InterfaceParameterDTO{}).Where("interface_id = ?", iface.Interface.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestInterfaceWithPathParameters(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "PathParamApp",
		Path:     "path-param-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建带有path参数的接口
	iface, err := CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "PathParamInterface",
		Protocol: "http",
		URL:      "https://api.example.com/users/{userId}/posts/{postId}",
		Method:   "GET",
		AuthType: "none",
		Parameters: []CreateInterfaceParameterReq{
			{
				Name:     "userId",
				Type:     "string",
				Location: "path",
				Required: true,
				Group:    "input",
			},
			{
				Name:     "postId",
				Type:     "string",
				Location: "path",
				Required: true,
				Group:    "input",
			},
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, iface.Interface.ID)
	assert.Len(t, iface.Interface.Parameters, 2)
}
