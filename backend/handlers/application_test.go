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

func TestCreateApplication(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	router := setupTestRouter()
	router.POST("/applications", CreateApplication)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: service.CreateApplicationRequest{
				Name:        "Test App",
				Description: "Test Description",
				Path:        "test-app",
				Protocol:    "sse",
				PostProcess: "",
				Environment: "{}",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var app models.Application
				err := json.Unmarshal(resp.Body.Bytes(), &app)
				assert.NoError(t, err)
				assert.Equal(t, "Test App", app.Name)
				assert.Equal(t, "test-app", app.Path)
				assert.Equal(t, "sse", app.Protocol)
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
			requestBody: service.CreateApplicationRequest{
				Name: "Test App",
				// Missing Path and Protocol
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
		{
			name: "invalid protocol",
			requestBody: service.CreateApplicationRequest{
				Name:     "Test App",
				Path:     "test-app",
				Protocol: "invalid",
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

			req, _ := http.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
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

func TestGetApplications(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试数据
	app1 := models.Application{
		Name:        "App 1",
		Path:        "app1",
		Protocol:    "sse",
		Description: "Description 1",
		Enabled:     true,
	}
	app2 := models.Application{
		Name:        "App 2",
		Path:        "app2",
		Protocol:    "streamable",
		Description: "Description 2",
		Enabled:     true,
	}

	db := database.GetDB()
	db.Create(&app1)
	db.Create(&app2)

	router := setupTestRouter()
	router.GET("/applications", GetApplications)

	req, _ := http.NewRequest(http.MethodGet, "/applications", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result service.ApplicationsResponse
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Applications), 2)
}

func TestGetApplication(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试数据
	app := models.Application{
		Name:        "Test App",
		Path:        "test-app",
		Protocol:    "sse",
		Description: "Test Description",
		Enabled:     true,
	}

	db := database.GetDB()
	db.Create(&app)

	router := setupTestRouter()
	router.GET("/applications/:id", GetApplication)

	tests := []struct {
		name           string
		appID          string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "successful retrieval",
			appID:          "1",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.Application
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Test App", result.Name)
			},
		},
		{
			name:           "invalid ID format",
			appID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid application ID")
			},
		},
		{
			name:           "non-existent application",
			appID:          "9999",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/applications/"+tt.appID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestGetApplicationDetail(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试数据
	app := models.Application{
		Name:        "Test App",
		Path:        "test-app",
		Protocol:    "sse",
		Description: "Test Description",
		Enabled:     true,
	}

	db := database.GetDB()
	db.Create(&app)

	router := setupTestRouter()
	router.GET("/applications/:id/detail", GetApplicationDetail)

	req, _ := http.NewRequest(http.MethodGet, "/applications/1/detail", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result service.ApplicationDetailResponse
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result.Application)
}

func TestUpdateApplication(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试数据
	app := models.Application{
		Name:        "Original App",
		Path:        "original-app",
		Protocol:    "sse",
		Description: "Original Description",
		Enabled:     true,
	}

	db := database.GetDB()
	db.Create(&app)

	router := setupTestRouter()
	router.PUT("/applications/:id", UpdateApplication)

	tests := []struct {
		name           string
		appID          string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:  "successful update",
			appID: "1",
			requestBody: service.UpdateApplicationRequest{
				Name:        stringPtr("Updated App"),
				Description: stringPtr("Updated Description"),
				Path:        stringPtr("updated-app"),
				Protocol:    stringPtr("sse"),
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.Application
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Updated App", result.Name)
				assert.Equal(t, "Updated Description", result.Description)
				assert.Equal(t, "updated-app", result.Path)
			},
		},
		{
			name:           "invalid ID format",
			appID:          "invalid",
			requestBody:    service.UpdateApplicationRequest{},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid application ID")
			},
		},
		{
			name:           "invalid JSON format",
			appID:          "1",
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

			req, _ := http.NewRequest(http.MethodPut, "/applications/"+tt.appID, bytes.NewBuffer(body))
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

func TestDeleteApplication(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	router := setupTestRouter()
	router.DELETE("/applications/:id", DeleteApplication)

	tests := []struct {
		name           string
		setupFunc      func() int64
		appID          string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful deletion",
			setupFunc: func() int64 {
				app := models.Application{
					Name:     "To Delete",
					Path:     "to-delete",
					Protocol: "sse",
					Enabled:  true,
				}
				db := database.GetDB()
				db.Create(&app)
				return app.ID
			},
			appID:          "1",
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Verify deletion
				db := database.GetDB()
				var app models.Application
				err := db.First(&app, 1).Error
				assert.Error(t, err) // Should not find the deleted app
			},
		},
		{
			name:           "invalid ID format",
			setupFunc:      func() int64 { return 0 },
			appID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid application ID")
			},
		},
		{
			name:           "non-existent application",
			setupFunc:      func() int64 { return 0 },
			appID:          "9999",
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

			req, _ := http.NewRequest(http.MethodDelete, "/applications/"+tt.appID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}
