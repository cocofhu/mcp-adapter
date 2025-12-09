package handlers

import (
	"bytes"
	"encoding/json"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"mcp-adapter/backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateInterface(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	router := setupTestRouter()
	router.POST("/interfaces", CreateInterface)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: service.CreateInterfaceRequest{
				AppID:       app.ID,
				Name:        "Test Interface",
				Description: "Test Description",
				Protocol:    "http",
				URL:         "https://api.example.com/test",
				Method:      "GET",
				AuthType:    "none",
				Enabled:     true,
				Parameters: []service.CreateInterfaceParameterReq{
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
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var iface models.Interface
				err := json.Unmarshal(resp.Body.Bytes(), &iface)
				assert.NoError(t, err)
				assert.Equal(t, "Test Interface", iface.Name)
				assert.Equal(t, "GET", iface.Method)
				assert.Equal(t, "none", iface.AuthType)
			},
		},
		{
			name:           "invalid JSON format",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid JSON format")
			},
		},
		{
			name: "missing required fields",
			requestBody: service.CreateInterfaceRequest{
				AppID: app.ID,
				Name:  "Test Interface",
				// Missing Protocol, URL, Method, AuthType
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
		{
			name: "invalid HTTP method",
			requestBody: service.CreateInterfaceRequest{
				AppID:    app.ID,
				Name:     "Test Interface",
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "INVALID",
				AuthType: "none",
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
		{
			name: "invalid parameter location",
			requestBody: service.CreateInterfaceRequest{
				AppID:    app.ID,
				Name:     "Test Interface",
				Protocol: "http",
				URL:      "https://api.example.com/test",
				Method:   "POST",
				AuthType: "none",
				Parameters: []service.CreateInterfaceParameterReq{
					{
						Name:     "param1",
						Type:     "string",
						Location: "invalid",
						Group:    "input",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest(http.MethodPost, "/interfaces", bytes.NewBuffer(body))
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

func TestGetInterfaces(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	// 创建测试接口
	iface1 := models.Interface{
		AppID:       app.ID,
		Name:        "Interface 1",
		Protocol:    "http",
		URL:         "https://api.example.com/1",
		Method:      "GET",
		AuthType:    "none",
		Enabled:     true,
		Description: "Description 1",
	}
	iface2 := models.Interface{
		AppID:       app.ID,
		Name:        "Interface 2",
		Protocol:    "http",
		URL:         "https://api.example.com/2",
		Method:      "POST",
		AuthType:    "none",
		Enabled:     true,
		Description: "Description 2",
	}
	db.Create(&iface1)
	db.Create(&iface2)

	router := setupTestRouter()
	router.GET("/interfaces", GetInterfaces)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "get all interfaces for app",
			queryParams:    "?app_id=1",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result service.InterfacesResponse
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.Interfaces), 2)
			},
		},
		{
			name:           "get interfaces without app_id",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error because app_id is required
				assert.Contains(t, resp.Body.String(), "AppID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/interfaces"+tt.queryParams, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestGetInterface(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用和接口
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	iface := models.Interface{
		AppID:       app.ID,
		Name:        "Test Interface",
		Protocol:    "http",
		URL:         "https://api.example.com/test",
		Method:      "GET",
		AuthType:    "none",
		Enabled:     true,
		Description: "Test Description",
	}
	db.Create(&iface)

	router := setupTestRouter()
	router.GET("/interfaces/:id", GetInterface)

	tests := []struct {
		name           string
		interfaceID    string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "successful retrieval",
			interfaceID:    "1",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.Interface
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Test Interface", result.Name)
				assert.Equal(t, "GET", result.Method)
			},
		},
		{
			name:           "invalid ID format",
			interfaceID:    "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid interface ID")
			},
		},
		{
			name:           "non-existent interface",
			interfaceID:    "9999",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/interfaces/"+tt.interfaceID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestUpdateInterface(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用和接口
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	iface := models.Interface{
		AppID:       app.ID,
		Name:        "Original Interface",
		Protocol:    "http",
		URL:         "https://api.example.com/original",
		Method:      "GET",
		AuthType:    "none",
		Enabled:     true,
		Description: "Original Description",
	}
	db.Create(&iface)

	router := setupTestRouter()
	router.PUT("/interfaces/:id", UpdateInterface)

	tests := []struct {
		name           string
		interfaceID    string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:        "successful update",
			interfaceID: "1",
			requestBody: service.UpdateInterfaceRequest{
				Name:        stringPtr("Updated Interface"),
				Description: stringPtr("Updated Description"),
				Method:      stringPtr("POST"),
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.Interface
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Updated Interface", result.Name)
				assert.Equal(t, "Updated Description", result.Description)
				assert.Equal(t, "POST", result.Method)
			},
		},
		{
			name:           "invalid ID format",
			interfaceID:    "invalid",
			requestBody:    service.UpdateInterfaceRequest{},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid interface ID")
			},
		},
		{
			name:           "invalid JSON format",
			interfaceID:    "1",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid JSON format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest(http.MethodPut, "/interfaces/"+tt.interfaceID, bytes.NewBuffer(body))
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

func TestDeleteInterface(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	router := setupTestRouter()
	router.DELETE("/interfaces/:id", DeleteInterface)

	tests := []struct {
		name           string
		setupFunc      func() int64
		interfaceID    string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful deletion",
			setupFunc: func() int64 {
				app := models.Application{
					Name:     "Test App",
					Path:     "test-app",
					Protocol: "sse",
					Enabled:  true,
				}
				db := database.GetDB()
				db.Create(&app)

				iface := models.Interface{
					AppID:    app.ID,
					Name:     "To Delete",
					Protocol: "http",
					URL:      "https://api.example.com/delete",
					Method:   "GET",
					AuthType: "none",
					Enabled:  true,
				}
				db.Create(&iface)
				return iface.ID
			},
			interfaceID:    "1",
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Verify deletion
				db := database.GetDB()
				var iface models.Interface
				err := db.First(&iface, 1).Error
				assert.Error(t, err) // Should not find the deleted interface
			},
		},
		{
			name:           "invalid ID format",
			setupFunc:      func() int64 { return 0 },
			interfaceID:    "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid interface ID")
			},
		},
		{
			name:           "non-existent interface",
			setupFunc:      func() int64 { return 0 },
			interfaceID:    "9999",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			req, _ := http.NewRequest(http.MethodDelete, "/interfaces/"+tt.interfaceID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}
