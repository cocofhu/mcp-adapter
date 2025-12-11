package adapter

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHTTPSimpleAdapter_Compatible(t *testing.T) {
	adapter := HTTPSimpleAdapter{}

	tests := []struct {
		name     string
		meta     RequestMeta
		expected bool
	}{
		{
			name: "compatible - http with none auth",
			meta: RequestMeta{
				Protocol: "http",
				AuthType: "none",
			},
			expected: true,
		},
		{
			name: "incompatible - https protocol",
			meta: RequestMeta{
				Protocol: "https",
				AuthType: "none",
			},
			expected: false,
		},
		{
			name: "incompatible - capi auth",
			meta: RequestMeta{
				Protocol: "http",
				AuthType: "capi",
			},
			expected: false,
		},
		{
			name: "incompatible - grpc protocol",
			meta: RequestMeta{
				Protocol: "grpc",
				AuthType: "none",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.Compatible(tt.meta)
			if result != tt.expected {
				t.Errorf("Compatible() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHTTPSimpleAdapter_DoRequest(t *testing.T) {
	tests := []struct {
		name           string
		parameters     Parameters
		meta           RequestMeta
		serverHandler  http.HandlerFunc
		expectedError  bool
		expectedStatus int
		validateResult func(t *testing.T, data []byte)
	}{
		{
			name: "successful GET request",
			parameters: Parameters{
				QueryParams: map[string]any{
					"key": "value",
				},
				HeaderParams: map[string]any{
					"X-Custom-Header": "test",
				},
			},
			meta: RequestMeta{
				Method: http.MethodGet,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if r.URL.Query().Get("key") != "value" {
					t.Errorf("Expected query param key=value, got %s", r.URL.Query().Get("key"))
				}
				if r.Header.Get("X-Custom-Header") != "test" {
					t.Errorf("Expected header X-Custom-Header=test, got %s", r.Header.Get("X-Custom-Header"))
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"status":"success"}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, data []byte) {
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if result["status"] != "success" {
					t.Errorf("Expected status=success, got %v", result["status"])
				}
			},
		},
		{
			name: "successful POST request with body",
			parameters: Parameters{
				BodyParams: map[string]any{
					"name": "test",
					"age":  30,
				},
				HeaderParams: map[string]any{
					"Content-Type": "application/json",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				body, _ := io.ReadAll(r.Body)
				var data map[string]any
				err := json.Unmarshal(body, &data)
				if err != nil {
					t.Errorf("Failed to unmarshal body: %v", err)
					return
				}
				if data["name"] != "test" {
					t.Errorf("Expected name=test, got %v", data["name"])
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(`{"created":true}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, data []byte) {
				var result map[string]any
				err := json.Unmarshal(data, &result)
				if err != nil {
					log.Printf("Unmarshal response error: %v", err)
					return
				}
				if result["created"] != true {
					t.Errorf("Expected created=true, got %v", result["created"])
				}
			},
		},
		{
			name: "path parameters replacement",
			parameters: Parameters{
				PathParams: map[string]any{
					"id":   "123",
					"name": "test",
				},
			},
			meta: RequestMeta{
				URL:    "http://example.com/users/{id}/profile/{name}",
				Method: http.MethodGet,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/users/123/profile/test" {
					t.Errorf("Expected path /users/123/profile/test, got %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"ok":true}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name: "GET request moves body params to query",
			parameters: Parameters{
				BodyParams: map[string]any{
					"filter": "active",
					"limit":  10,
				},
			},
			meta: RequestMeta{
				Method: http.MethodGet,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("filter") != "active" {
					t.Errorf("Expected query param filter=active, got %s", r.URL.Query().Get("filter"))
				}
				if r.URL.Query().Get("limit") != "10" {
					t.Errorf("Expected query param limit=10, got %s", r.URL.Query().Get("limit"))
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"ok":true}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name: "array query parameters",
			parameters: Parameters{
				QueryParams: map[string]any{
					"tags": []string{"tag1", "tag2", "tag3"},
				},
			},
			meta: RequestMeta{
				Method: http.MethodGet,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				tags := r.URL.Query()["tags"]
				if len(tags) != 3 {
					t.Errorf("Expected 3 tags, got %d", len(tags))
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"ok":true}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:       "server returns error status",
			parameters: Parameters{},
			meta: RequestMeta{
				Method: http.MethodGet,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte(`{"error":"internal server error"}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError: true,
		},
		{
			name: "default content-type header",
			parameters: Parameters{
				BodyParams: map[string]any{
					"data": "test",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("Expected default Content-Type, got %s", contentType)
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"ok":true}`))
				if err != nil {
					log.Printf("Write response error: %v", err)
					return
				}
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务器
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			// 如果meta中没有URL，使用测试服务器的URL
			if tt.meta.URL == "" {
				tt.meta.URL = server.URL
			} else {
				// 替换URL中的host为测试服务器的host
				tt.meta.URL = server.URL + tt.meta.URL[len("http://example.com"):]
			}

			adapter := HTTPSimpleAdapter{}
			req := mcp.CallToolRequest{}

			data, err := adapter.DoRequest(context.Background(), req, tt.parameters, tt.meta)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, data)
			}
		})
	}
}

func TestBuildCommonHttpRequest(t *testing.T) {
	tests := []struct {
		name          string
		parameters    Parameters
		meta          RequestMeta
		expectedError bool
		validate      func(t *testing.T, req *http.Request, payload []byte)
	}{
		{
			name: "build request with all parameter types",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Authorization": "Bearer token",
				},
				QueryParams: map[string]any{
					"page": 1,
				},
				PathParams: map[string]any{
					"id": "123",
				},
				BodyParams: map[string]any{
					"name": "test",
				},
			},
			meta: RequestMeta{
				URL:    "http://example.com/users/{id}",
				Method: http.MethodPost,
			},
			expectedError: false,
			validate: func(t *testing.T, req *http.Request, payload []byte) {
				if req.URL.Path != "/users/123" {
					t.Errorf("Expected path /users/123, got %s", req.URL.Path)
				}
				if req.URL.Query().Get("page") != "1" {
					t.Errorf("Expected query page=1, got %s", req.URL.Query().Get("page"))
				}
				if req.Header.Get("Authorization") != "Bearer token" {
					t.Errorf("Expected Authorization header, got %s", req.Header.Get("Authorization"))
				}
				var body map[string]any
				err := json.Unmarshal(payload, &body)
				if err != nil {
					log.Printf("Unmarshal body error: %v", err)
					return
				}
				if body["name"] != "test" {
					t.Errorf("Expected body name=test, got %v", body["name"])
				}
			},
		},
		{
			name: "colon-style path parameters",
			parameters: Parameters{
				PathParams: map[string]any{
					"userId": "456",
				},
			},
			meta: RequestMeta{
				URL:    "http://example.com/api/:userId/profile",
				Method: http.MethodGet,
			},
			expectedError: false,
			validate: func(t *testing.T, req *http.Request, payload []byte) {
				if req.URL.Path != "/api/456/profile" {
					t.Errorf("Expected path /api/456/profile, got %s", req.URL.Path)
				}
			},
		},
		{
			name: "merge existing query parameters",
			parameters: Parameters{
				QueryParams: map[string]any{
					"new": "param",
				},
			},
			meta: RequestMeta{
				URL:    "http://example.com/api?existing=value",
				Method: http.MethodGet,
			},
			expectedError: false,
			validate: func(t *testing.T, req *http.Request, payload []byte) {
				if req.URL.Query().Get("existing") != "value" {
					t.Errorf("Expected existing query param, got %s", req.URL.Query().Get("existing"))
				}
				if req.URL.Query().Get("new") != "param" {
					t.Errorf("Expected new query param, got %s", req.URL.Query().Get("new"))
				}
			},
		},
		{
			name:       "invalid URL",
			parameters: Parameters{},
			meta: RequestMeta{
				URL:    "://invalid-url",
				Method: http.MethodGet,
			},
			expectedError: true,
		},
		{
			name: "empty body for GET request",
			parameters: Parameters{
				BodyParams: map[string]any{},
			},
			meta: RequestMeta{
				URL:    "http://example.com/api",
				Method: http.MethodGet,
			},
			expectedError: false,
			validate: func(t *testing.T, req *http.Request, payload []byte) {
				if len(payload) != 0 {
					t.Errorf("Expected empty payload for GET request, got %d bytes", len(payload))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, payload, err := BuildCommonHttpRequest(context.Background(), tt.parameters, tt.meta)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, req, payload)
			}
		})
	}
}

func BenchmarkHTTPSimpleAdapter_DoRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"success"}`))
		if err != nil {
			log.Printf("write response error: %v", err)
			return
		}
	}))
	defer server.Close()

	adapter := HTTPSimpleAdapter{}
	parameters := Parameters{
		QueryParams: map[string]any{
			"key": "value",
		},
		BodyParams: map[string]any{
			"data": "test",
		},
	}
	meta := RequestMeta{
		URL:    server.URL,
		Method: http.MethodPost,
	}
	req := mcp.CallToolRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.DoRequest(context.Background(), req, parameters, meta)
	}
}
