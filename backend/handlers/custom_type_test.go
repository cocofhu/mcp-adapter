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

func TestCreateCustomType(t *testing.T) {
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
	router.POST("/custom-types", CreateCustomType)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: service.CreateCustomTypeRequest{
				AppID:       app.ID,
				Name:        "TestType",
				Description: "Test custom type",
				Fields: []service.CreateCustomTypeFieldReq{
					{
						Name:        "field1",
						Type:        "string",
						Required:    true,
						Description: "Test field",
						IsArray:     false,
					},
					{
						Name:        "field2",
						Type:        "number",
						Required:    false,
						Description: "Numeric field",
						IsArray:     false,
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var customType models.CustomType
				err := json.Unmarshal(resp.Body.Bytes(), &customType)
				assert.NoError(t, err)
				assert.Equal(t, "TestType", customType.Name)
				assert.Equal(t, "Test custom type", customType.Description)
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
			requestBody: service.CreateCustomTypeRequest{
				AppID: app.ID,
				// Missing Name
				Fields: []service.CreateCustomTypeFieldReq{},
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
		{
			name: "invalid field type",
			requestBody: service.CreateCustomTypeRequest{
				AppID:       app.ID,
				Name:        "TestType",
				Description: "Test custom type",
				Fields: []service.CreateCustomTypeFieldReq{
					{
						Name:     "field1",
						Type:     "invalid_type",
						Required: true,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return validation error
			},
		},
		{
			name: "nested custom type",
			requestBody: service.CreateCustomTypeRequest{
				AppID:       app.ID,
				Name:        "NestedType",
				Description: "Type with nested custom type",
				Fields: []service.CreateCustomTypeFieldReq{
					{
						Name:     "field1",
						Type:     "string",
						Required: true,
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var customType models.CustomType
				err := json.Unmarshal(resp.Body.Bytes(), &customType)
				assert.NoError(t, err)
				assert.Equal(t, "NestedType", customType.Name)
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

			req, _ := http.NewRequest(http.MethodPost, "/custom-types", bytes.NewBuffer(body))
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

func TestGetCustomTypes(t *testing.T) {
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

	// 创建测试自定义类型
	customType1 := models.CustomType{
		AppID:       app.ID,
		Name:        "Type1",
		Description: "Description 1",
	}
	customType2 := models.CustomType{
		AppID:       app.ID,
		Name:        "Type2",
		Description: "Description 2",
	}
	db.Create(&customType1)
	db.Create(&customType2)

	router := setupTestRouter()
	router.GET("/custom-types", GetCustomTypes)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "get all custom types for app",
			queryParams:    "?app_id=1",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result service.CustomTypesResponse
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.CustomTypes), 2)
			},
		},
		{
			name:           "missing app_id parameter",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid app_id parameter")
			},
		},
		{
			name:           "invalid app_id format",
			queryParams:    "?app_id=invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid app_id parameter")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/custom-types"+tt.queryParams, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestGetCustomType(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用和自定义类型
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	customType := models.CustomType{
		AppID:       app.ID,
		Name:        "TestType",
		Description: "Test Description",
	}
	db.Create(&customType)

	router := setupTestRouter()
	router.GET("/custom-types/:id", GetCustomType)

	tests := []struct {
		name           string
		customTypeID   string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "successful retrieval",
			customTypeID:   "1",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.CustomType
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "TestType", result.Name)
				assert.Equal(t, "Test Description", result.Description)
			},
		},
		{
			name:           "invalid ID format",
			customTypeID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid custom type ID")
			},
		},
		{
			name:           "non-existent custom type",
			customTypeID:   "9999",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Should return error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/custom-types/"+tt.customTypeID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestUpdateCustomType(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	// 创建测试应用和自定义类型
	app := models.Application{
		Name:     "Test App",
		Path:     "test-app",
		Protocol: "sse",
		Enabled:  true,
	}
	db := database.GetDB()
	db.Create(&app)

	customType := models.CustomType{
		AppID:       app.ID,
		Name:        "OriginalType",
		Description: "Original Description",
	}
	db.Create(&customType)

	router := setupTestRouter()
	router.PUT("/custom-types/:id", UpdateCustomType)

	tests := []struct {
		name           string
		customTypeID   string
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:         "successful update",
			customTypeID: "1",
			requestBody: service.UpdateCustomTypeRequest{
				Name:        stringPtr("UpdatedType"),
				Description: stringPtr("Updated Description"),
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.CustomType
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "UpdatedType", result.Name)
				assert.Equal(t, "Updated Description", result.Description)
			},
		},
		{
			name:           "invalid ID format",
			customTypeID:   "invalid",
			requestBody:    service.UpdateCustomTypeRequest{},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid custom type ID")
			},
		},
		{
			name:           "invalid JSON format",
			customTypeID:   "1",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid JSON format")
			},
		},
		{
			name:         "update with new fields",
			customTypeID: "1",
			requestBody: service.UpdateCustomTypeRequest{
				Name: stringPtr("UpdatedType"),
				Fields: &[]service.UpdateCustomTypeFieldReq{
					{
						Name:        "newField",
						Type:        "string",
						Required:    true,
						Description: "New field",
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result models.CustomType
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "UpdatedType", result.Name)
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

			req, _ := http.NewRequest(http.MethodPut, "/custom-types/"+tt.customTypeID, bytes.NewBuffer(body))
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

func TestDeleteCustomType(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	router := setupTestRouter()
	router.DELETE("/custom-types/:id", DeleteCustomType)

	tests := []struct {
		name           string
		setupFunc      func() int64
		customTypeID   string
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

				customType := models.CustomType{
					AppID:       app.ID,
					Name:        "ToDelete",
					Description: "To be deleted",
				}
				db.Create(&customType)
				return customType.ID
			},
			customTypeID:   "1",
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Verify deletion
				db := database.GetDB()
				var customType models.CustomType
				err := db.First(&customType, 1).Error
				assert.Error(t, err) // Should not find the deleted custom type
			},
		},
		{
			name:           "invalid ID format",
			setupFunc:      func() int64 { return 0 },
			customTypeID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Body.String(), "Invalid custom type ID")
			},
		},
		{
			name:           "non-existent custom type",
			setupFunc:      func() int64 { return 0 },
			customTypeID:   "9999",
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

			req, _ := http.NewRequest(http.MethodDelete, "/custom-types/"+tt.customTypeID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}
