package service

import (
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 初始化测试数据库
func setupTestDB(t *testing.T) {
	database.InitDatabase(":memory:")
	t.Cleanup(func() {
		cleanupTestDB()
	})
}

// cleanupTestDB 清理测试数据
func cleanupTestDB() {
	db := database.GetDB()
	db.Exec("DELETE FROM interface_parameters")
	db.Exec("DELETE FROM interfaces")
	db.Exec("DELETE FROM custom_type_fields")
	db.Exec("DELETE FROM custom_types")
	db.Exec("DELETE FROM applications")
}

// boolPtr 返回布尔指针
func boolPtr(b bool) *bool {
	return &b
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

func TestCreateApplication(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name    string
		req     CreateApplicationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功创建应用",
			req: CreateApplicationRequest{
				Name:        "TestApp",
				Description: "Test application",
				Path:        "test-app",
				Protocol:    "sse",
				PostProcess: "",
				Environment: `{"key":"value"}`,
				Enabled:     boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段name",
			req: CreateApplicationRequest{
				Path:     "test-app2",
				Protocol: "sse",
			},
			wantErr: true,
		},
		{
			name: "缺少必填字段path",
			req: CreateApplicationRequest{
				Name:     "TestApp2",
				Protocol: "sse",
			},
			wantErr: true,
		},
		{
			name: "无效的protocol",
			req: CreateApplicationRequest{
				Name:     "TestApp3",
				Path:     "test-app3",
				Protocol: "invalid",
			},
			wantErr: true,
		},
		{
			name: "重复的应用名称",
			req: CreateApplicationRequest{
				Name:     "TestApp",
				Path:     "test-app-duplicate",
				Protocol: "sse",
			},
			wantErr: true,
			errMsg:  "duplicate application name",
		},
		{
			name: "重复的应用路径",
			req: CreateApplicationRequest{
				Name:     "TestAppDuplicate",
				Path:     "test-app",
				Protocol: "sse",
			},
			wantErr: true,
			errMsg:  "duplicate application path",
		},
		{
			name: "使用streamable协议",
			req: CreateApplicationRequest{
				Name:     "StreamableApp",
				Path:     "streamable-app",
				Protocol: "streamable",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := CreateApplication(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotZero(t, resp.Application.ID)
				assert.Equal(t, tt.req.Name, resp.Application.Name)
				assert.Equal(t, tt.req.Path, resp.Application.Path)
				assert.Equal(t, tt.req.Protocol, resp.Application.Protocol)
			}
		})
	}
}

func TestGetApplication(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	createResp, err := CreateApplication(CreateApplicationRequest{
		Name:        "GetTestApp",
		Description: "Test get application",
		Path:        "get-test-app",
		Protocol:    "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     GetApplicationRequest
		wantErr bool
	}{
		{
			name: "成功获取应用",
			req: GetApplicationRequest{
				ID:         createResp.Application.ID,
				ShowDetail: false,
			},
			wantErr: false,
		},
		{
			name: "获取不存在的应用",
			req: GetApplicationRequest{
				ID:         99999,
				ShowDetail: false,
			},
			wantErr: true,
		},
		{
			name: "无效的ID",
			req: GetApplicationRequest{
				ID:         0,
				ShowDetail: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := GetApplication(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.Application.ID)
				assert.Equal(t, "GetTestApp", resp.Application.Name)
			}
		})
	}
}

func TestGetApplicationWithDetail(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	createResp, err := CreateApplication(CreateApplicationRequest{
		Name:     "DetailTestApp",
		Path:     "detail-test-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建测试接口
	_, err = CreateInterface(CreateInterfaceRequest{
		AppID:       createResp.Application.ID,
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
	})
	require.NoError(t, err)

	// 获取应用详情
	resp, err := GetApplication(GetApplicationRequest{
		ID:         createResp.Application.ID,
		ShowDetail: true,
	})
	require.NoError(t, err)
	assert.Equal(t, createResp.Application.ID, resp.Application.ID)
	assert.Len(t, resp.ToolDefinitions, 1)
	assert.Equal(t, "TestInterface", resp.ToolDefinitions[0].Name)
}

func TestListApplications(t *testing.T) {
	setupTestDB(t)

	// 创建多个测试应用
	apps := []CreateApplicationRequest{
		{Name: "App1", Path: "app1", Protocol: "sse"},
		{Name: "App2", Path: "app2", Protocol: "streamable"},
		{Name: "App3", Path: "app3", Protocol: "sse"},
	}

	for _, app := range apps {
		_, err := CreateApplication(app)
		require.NoError(t, err)
	}

	// 列出所有应用
	resp, err := ListApplications(ListApplicationsRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Applications), 3)

	// 验证应用名称
	names := make(map[string]bool)
	for _, app := range resp.Applications {
		names[app.Name] = true
	}
	assert.True(t, names["App1"])
	assert.True(t, names["App2"])
	assert.True(t, names["App3"])
}

func TestUpdateApplication(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	createResp, err := CreateApplication(CreateApplicationRequest{
		Name:        "UpdateTestApp",
		Description: "Original description",
		Path:        "update-test-app",
		Protocol:    "sse",
		Enabled:     boolPtr(true),
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     UpdateApplicationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功更新名称",
			req: UpdateApplicationRequest{
				ID:   createResp.Application.ID,
				Name: stringPtr("UpdatedApp"),
			},
			wantErr: false,
		},
		{
			name: "成功更新描述",
			req: UpdateApplicationRequest{
				ID:          createResp.Application.ID,
				Description: stringPtr("Updated description"),
			},
			wantErr: false,
		},
		{
			name: "成功更新路径",
			req: UpdateApplicationRequest{
				ID:   createResp.Application.ID,
				Path: stringPtr("updated-path"),
			},
			wantErr: false,
		},
		{
			name: "成功更新协议",
			req: UpdateApplicationRequest{
				ID:       createResp.Application.ID,
				Protocol: stringPtr("streamable"),
			},
			wantErr: false,
		},
		{
			name: "成功禁用应用",
			req: UpdateApplicationRequest{
				ID:      createResp.Application.ID,
				Enabled: boolPtr(false),
			},
			wantErr: false,
		},
		{
			name: "更新不存在的应用",
			req: UpdateApplicationRequest{
				ID:   99999,
				Name: stringPtr("NonExistent"),
			},
			wantErr: true,
			errMsg:  "no such application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := UpdateApplication(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.ID, resp.Application.ID)
				if tt.req.Name != nil {
					assert.Equal(t, *tt.req.Name, resp.Application.Name)
				}
				if tt.req.Description != nil {
					assert.Equal(t, *tt.req.Description, resp.Application.Description)
				}
				if tt.req.Path != nil {
					assert.Equal(t, *tt.req.Path, resp.Application.Path)
				}
				if tt.req.Protocol != nil {
					assert.Equal(t, *tt.req.Protocol, resp.Application.Protocol)
				}
				if tt.req.Enabled != nil {
					assert.Equal(t, *tt.req.Enabled, resp.Application.Enabled)
				}
			}
		})
	}
}

func TestUpdateApplicationDuplicateName(t *testing.T) {
	setupTestDB(t)

	// 创建两个应用
	app1, err := CreateApplication(CreateApplicationRequest{
		Name:     "App1",
		Path:     "app1",
		Protocol: "sse",
	})
	require.NoError(t, err)

	_, err = CreateApplication(CreateApplicationRequest{
		Name:     "App2",
		Path:     "app2",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 尝试将App1的名称改为App2（重复）
	_, err = UpdateApplication(UpdateApplicationRequest{
		ID:   app1.Application.ID,
		Name: stringPtr("App2"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate application name")
}

func TestUpdateApplicationDuplicatePath(t *testing.T) {
	setupTestDB(t)

	// 创建两个应用
	app1, err := CreateApplication(CreateApplicationRequest{
		Name:     "PathApp1",
		Path:     "path1",
		Protocol: "sse",
	})
	require.NoError(t, err)

	_, err = CreateApplication(CreateApplicationRequest{
		Name:     "PathApp2",
		Path:     "path2",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 尝试将PathApp1的路径改为path2（重复）
	_, err = UpdateApplication(UpdateApplicationRequest{
		ID:   app1.Application.ID,
		Path: stringPtr("path2"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate application path")
}

func TestDeleteApplication(t *testing.T) {
	setupTestDB(t)

	// 创建测试应用
	createResp, err := CreateApplication(CreateApplicationRequest{
		Name:     "DeleteTestApp",
		Path:     "delete-test-app",
		Protocol: "sse",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     DeleteApplicationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功删除应用",
			req: DeleteApplicationRequest{
				ID: createResp.Application.ID,
			},
			wantErr: false,
		},
		{
			name: "删除不存在的应用",
			req: DeleteApplicationRequest{
				ID: 99999,
			},
			wantErr: true,
			errMsg:  "no such application",
		},
		{
			name: "无效的ID",
			req: DeleteApplicationRequest{
				ID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeleteApplication(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				// 验证应用已被删除
				_, err := GetApplication(GetApplicationRequest{ID: tt.req.ID})
				assert.Error(t, err)
			}
		})
	}
}

func TestDeleteApplicationWithInterfaces(t *testing.T) {
	setupTestDB(t)

	// 创建应用
	app, err := CreateApplication(CreateApplicationRequest{
		Name:     "AppWithInterfaces",
		Path:     "app-with-interfaces",
		Protocol: "sse",
	})
	require.NoError(t, err)

	// 创建接口
	_, err = CreateInterface(CreateInterfaceRequest{
		AppID:    app.Application.ID,
		Name:     "TestInterface",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
	})
	require.NoError(t, err)

	// 尝试删除有接口的应用
	_, err = DeleteApplication(DeleteApplicationRequest{
		ID: app.Application.ID,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete application with associated interfaces")
}

func TestToApplicationDTO(t *testing.T) {
	now := models.Application{
		ID:          1,
		Name:        "TestApp",
		Description: "Test description",
		Path:        "test-path",
		Protocol:    "sse",
		PostProcess: "script",
		Environment: "env",
		Enabled:     true,
	}

	dto := toApplicationDTO(now)
	assert.Equal(t, now.ID, dto.ID)
	assert.Equal(t, now.Name, dto.Name)
	assert.Equal(t, now.Description, dto.Description)
	assert.Equal(t, now.Path, dto.Path)
	assert.Equal(t, now.Protocol, dto.Protocol)
	assert.Equal(t, now.PostProcess, dto.PostProcess)
	assert.Equal(t, now.Environment, dto.Environment)
	assert.Equal(t, now.Enabled, dto.Enabled)
}
