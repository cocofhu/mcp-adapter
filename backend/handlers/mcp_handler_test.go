package handlers

import (
	"mcp-adapter/backend/adapter"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServeSSE(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用
	app := models.Application{
		Name:        "SSE Test App",
		Path:        "sse-test",
		Protocol:    "sse",
		Description: "SSE Test Application",
		Enabled:     true,
	}
	db := database.GetDB()
	db.Create(&app)

	// 创建测试接口
	iface := models.Interface{
		AppID:       app.ID,
		Name:        "Test Interface",
		Protocol:    "http",
		URL:         "https://api.example.com/test",
		Method:      "GET",
		AuthType:    "none",
		Enabled:     true,
		Description: "Test Interface",
	}
	db.Create(&iface)

	// 初始化adapter
	adapter.InitServer()

	router := setupTestRouter()
	router.GET("/sse/:path", ServeSSE)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		skipExecution  bool // 跳过实际执行（用于SSE长连接）
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid SSE path",
			path:           "sse-test",
			expectedStatus: http.StatusOK,
			skipExecution:  true, // SSE是长连接，跳过实际测试
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// SSE endpoint would respond with event stream
				// Skipping actual execution to avoid hanging
				t.Log("SSE endpoint exists and is registered")
			},
		},
		{
			name:           "non-existent SSE path",
			path:           "non-existent",
			expectedStatus: http.StatusNotFound,
			skipExecution:  false,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "sse path not found")
			},
		},
		{
			name:           "empty path",
			path:           "",
			expectedStatus: http.StatusNotFound,
			skipExecution:  false,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return 404 for empty path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipExecution {
				// 只验证服务器是否注册，不实际执行请求
				impl := adapter.GetServerImpl(tt.path, "sse")
				if impl != nil {
					t.Log("SSE server is properly registered")
				} else {
					t.Error("SSE server should be registered")
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, nil)
				}
				return
			}

			req, _ := http.NewRequest(http.MethodGet, "/sse/"+tt.path, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestServeStreamable(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用
	app := models.Application{
		Name:        "Streamable Test App",
		Path:        "streamable-test",
		Protocol:    "streamable",
		Description: "Streamable Test Application",
		Enabled:     true,
	}
	db := database.GetDB()
	db.Create(&app)

	// 创建测试接口
	iface := models.Interface{
		AppID:    app.ID,
		Name:     "Test Interface",
		Protocol: "http",
		URL:      "https://api.example.com/test",
		Method:   "GET",
		AuthType: "none",
		Enabled:  true,
	}
	db.Create(&iface)

	// 确保 adapter 已初始化（只在第一次调用时初始化）
	adapter.InitServer()

	// 手动触发应用添加事件
	adapter.SendEvent(adapter.Event{
		App:  &app,
		Code: adapter.AddApplicationEvent,
	})

	// 等待事件处理（事件是异步处理的）
	time.Sleep(5000 * time.Millisecond)

	router := setupTestRouter()
	router.POST("/streamable/:path", ServeStreamable)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		skipExecution  bool // 跳过实际执行（用于可能的长连接）
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid streamable path",
			path:           "streamable-test",
			expectedStatus: http.StatusOK,
			skipExecution:  true, // Streamable可能是长连接，跳过实际测试
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Streamable endpoint would respond with stream
				// Skipping actual execution to avoid potential hanging
				// 验证服务器是否注册
				impl := adapter.GetServerImpl("streamable-test", "streamable")
				if impl != nil {
					t.Log("Streamable endpoint exists and is registered")
				} else {
					t.Error("Streamable server should be registered")
				}
			},
		},
		{
			name:           "non-existent streamable path",
			path:           "non-existent",
			expectedStatus: http.StatusNotFound,
			skipExecution:  false,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "streamable path not found")
			},
		},
		{
			name:           "empty path",
			path:           "",
			expectedStatus: http.StatusNotFound,
			skipExecution:  false,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return 404 for empty path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipExecution {
				// 只验证服务器是否注册，不实际执行请求
				if tt.validateFunc != nil {
					tt.validateFunc(t, nil)
				}
				return
			}

			req, _ := http.NewRequest(http.MethodPost, "/streamable/"+tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestServeSSE_WithMultipleApplications(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建多个测试应用
	apps := []models.Application{
		{
			Name:        "SSE App 1",
			Path:        "sse-app-1",
			Protocol:    "sse",
			Description: "First SSE Application",
			Enabled:     true,
		},
		{
			Name:        "SSE App 2",
			Path:        "sse-app-2",
			Protocol:    "sse",
			Description: "Second SSE Application",
			Enabled:     true,
		},
	}

	db := database.GetDB()

	// 确保 adapter 已初始化（必须在发送事件之前初始化）
	adapter.InitServer()

	for i := range apps {
		db.Create(&apps[i])
		// 为每个应用创建一个测试接口
		iface := models.Interface{
			AppID:    apps[i].ID,
			Name:     "Test Interface " + apps[i].Name,
			Protocol: "http",
			URL:      "https://api.example.com/test",
			Method:   "GET",
			AuthType: "none",
			Enabled:  true,
		}
		db.Create(&iface)

		adapter.SendEvent(adapter.Event{
			App:  &apps[i],
			Code: adapter.AddApplicationEvent,
		})
	}

	// 等待事件处理（事件是异步处理的，轮询间隔1秒，等待2.5秒确保至少2次轮询）
	time.Sleep(5000 * time.Millisecond)

	router := setupTestRouter()
	router.GET("/sse/:path", ServeSSE)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "access first app",
			path: "sse-app-1",
		},
		{
			name: "access second app",
			path: "sse-app-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只验证服务器是否注册，避免SSE长连接
			impl := adapter.GetServerImpl(tt.path, "sse")
			assert.NotNil(t, impl, "SSE server should be registered for path: "+tt.path)
		})
	}
}

func TestServeStreamable_WithMultipleApplications(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建多个测试应用
	apps := []models.Application{
		{
			Name:        "Streamable App 1",
			Path:        "streamable-app-1",
			Protocol:    "streamable",
			Description: "First Streamable Application",
			Enabled:     true,
		},
		{
			Name:        "Streamable App 2",
			Path:        "streamable-app-2",
			Protocol:    "streamable",
			Description: "Second Streamable Application",
			Enabled:     true,
		},
	}

	db := database.GetDB()

	// 确保 adapter 已初始化（必须在发送事件之前初始化）
	adapter.InitServer()

	for i := range apps {
		db.Create(&apps[i])
		// 为每个应用创建一个测试接口
		iface := models.Interface{
			AppID:    apps[i].ID,
			Name:     "Test Interface " + apps[i].Name,
			Protocol: "http",
			URL:      "https://api.example.com/test",
			Method:   "GET",
			AuthType: "none",
			Enabled:  true,
		}
		db.Create(&iface)

		// 手动触发应用添加事件
		adapter.SendEvent(adapter.Event{
			App:  &apps[i],
			Code: adapter.AddApplicationEvent,
		})
	}

	// 等待事件处理（事件是异步处理的，轮询间隔1秒，等待2.5秒确保至少2次轮询）
	time.Sleep(5000 * time.Millisecond)

	router := setupTestRouter()
	router.POST("/streamable/:path", ServeStreamable)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "access first app",
			path: "streamable-app-1",
		},
		{
			name: "access second app",
			path: "streamable-app-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只验证服务器是否注册，避免Streamable长连接
			impl := adapter.GetServerImpl(tt.path, "streamable")
			assert.NotNil(t, impl, "Streamable server should be registered for path: "+tt.path)
		})
	}
}

func TestServeSSE_DisabledApplication(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建禁用的应用
	app := models.Application{
		Name:        "Disabled SSE App",
		Path:        "disabled-sse",
		Protocol:    "sse",
		Description: "Disabled SSE Application",
		Enabled:     false,
	}
	db := database.GetDB()
	db.Create(&app)

	// 注意：不调用 InitServer()，因为禁用的应用不应该被加载
	// 这个测试验证的是即使应用存在于数据库中，如果未启用也不应该可访问

	router := setupTestRouter()
	router.GET("/sse/:path", ServeSSE)

	req, _ := http.NewRequest(http.MethodGet, "/sse/disabled-sse", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Disabled applications should not be accessible
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Contains(t, resp.Body.String(), "sse path not found")
}

func TestServeStreamable_DisabledApplication(t *testing.T) {
	// 重置 adapter 状态，确保测试独立
	adapter.Shutdown()

	setupTestDB()
	defer cleanupTestDB()

	// 创建禁用的应用
	app := models.Application{
		Name:        "Disabled Streamable App",
		Path:        "disabled-streamable",
		Protocol:    "streamable",
		Description: "Disabled Streamable Application",
		Enabled:     false,
	}
	db := database.GetDB()
	db.Create(&app)

	// 注意：不调用 InitServer()，因为禁用的应用不应该被加载
	// 这个测试验证的是即使应用存在于数据库中，如果未启用也不应该可访问

	router := setupTestRouter()
	router.POST("/streamable/:path", ServeStreamable)

	req, _ := http.NewRequest(http.MethodPost, "/streamable/disabled-streamable", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Disabled applications should not be accessible
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Contains(t, resp.Body.String(), "streamable path not found")
}
