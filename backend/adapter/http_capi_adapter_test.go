package adapter

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHTTPCAPIAdapter_Compatible(t *testing.T) {
	adapter := HTTPCAPIAdapter{}

	tests := []struct {
		name     string
		meta     RequestMeta
		expected bool
	}{
		{
			name: "compatible - http with capi auth",
			meta: RequestMeta{
				Protocol: "http",
				AuthType: "capi",
			},
			expected: true,
		},
		{
			name: "incompatible - http with none auth",
			meta: RequestMeta{
				Protocol: "http",
				AuthType: "none",
			},
			expected: false,
		},
		{
			name: "incompatible - grpc protocol",
			meta: RequestMeta{
				Protocol: "grpc",
				AuthType: "capi",
			},
			expected: false,
		},
		{
			name: "incompatible - https protocol",
			meta: RequestMeta{
				Protocol: "https",
				AuthType: "capi",
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

func TestHTTPCAPIAdapter_DoRequest(t *testing.T) {
	tests := []struct {
		name           string
		parameters     Parameters
		meta           RequestMeta
		setupRequest   func(req *mcp.CallToolRequest)
		serverHandler  http.HandlerFunc
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, data []byte)
	}{
		{
			name: "successful request with credentials in header",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
				BodyParams: map[string]any{
					"Limit": 10,
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-secret-id")
				req.Header.Set("TC-API-SecretKey", "test-secret-key")
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// 验证必需的腾讯云API头部
				if r.Header.Get("Authorization") == "" {
					t.Error("Missing Authorization header")
				}
				if r.Header.Get("X-TC-Action") != "DescribeInstances" {
					t.Errorf("Expected X-TC-Action=DescribeInstances, got %s", r.Header.Get("X-TC-Action"))
				}
				if r.Header.Get("X-TC-Version") != "2017-03-12" {
					t.Errorf("Expected X-TC-Version=2017-03-12, got %s", r.Header.Get("X-TC-Version"))
				}
				if r.Header.Get("X-TC-Region") != "ap-guangzhou" {
					t.Errorf("Expected X-TC-Region=ap-guangzhou, got %s", r.Header.Get("X-TC-Region"))
				}
				if r.Header.Get("X-TC-Timestamp") == "" {
					t.Error("Missing X-TC-Timestamp header")
				}

				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"Response":{"TotalCount":1}}`))
				if err != nil {
					log.Printf("Error writing response: %v", err)
					return
				}
			},
			expectedError: false,
			validateResult: func(t *testing.T, data []byte) {
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				response, ok := result["Response"].(map[string]any)
				if !ok {
					t.Error("Expected Response field in result")
				}
				if response["TotalCount"] != float64(1) {
					t.Errorf("Expected TotalCount=1, got %v", response["TotalCount"])
				}
			},
		},
		{
			name: "credentials in parameters",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":      "cvm.tencentcloudapi.com",
					"Service":   "cvm",
					"Version":   "2017-03-12",
					"Action":    "DescribeInstances",
					"Region":    "ap-guangzhou",
					"SecretId":  "param-secret-id",
					"SecretKey": "param-secret-key",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"Response":{"RequestId":"test-id"}}`))
				if err != nil {
					log.Printf("Error writing response: %v", err)
					return
				}
			},
			expectedError: false,
		},
		{
			name: "missing required parameter - Host",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-id")
				req.Header.Set("TC-API-SecretKey", "test-key")
			},
			expectedError: true,
			errorContains: "missing required parameter: Host",
		},
		{
			name: "missing required parameter - Service",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-id")
				req.Header.Set("TC-API-SecretKey", "test-key")
			},
			expectedError: true,
			errorContains: "missing required parameter: Service",
		},
		{
			name: "missing SecretId",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretKey", "test-key")
			},
			expectedError: true,
			errorContains: "missing SecretId or SecretKey",
		},
		{
			name: "missing SecretKey",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-id")
			},
			expectedError: true,
			errorContains: "missing SecretId or SecretKey",
		},
		{
			name: "server returns error",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "DescribeInstances",
					"Region":  "ap-guangzhou",
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-id")
				req.Header.Set("TC-API-SecretKey", "test-key")
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(`{"Response":{"Error":{"Code":"InvalidParameter"}}}`))
				if err != nil {
					log.Printf("Error writing response: %v", err)
					return
				}
			},
			expectedError: true,
			errorContains: "http request failed",
		},
		{
			name: "complex body parameters",
			parameters: Parameters{
				HeaderParams: map[string]any{
					"Host":    "cvm.tencentcloudapi.com",
					"Service": "cvm",
					"Version": "2017-03-12",
					"Action":  "RunInstances",
					"Region":  "ap-guangzhou",
				},
				BodyParams: map[string]any{
					"InstanceType": "S5.MEDIUM4",
					"ImageId":      "img-xxx",
					"SystemDisk": map[string]any{
						"DiskType": "CLOUD_PREMIUM",
						"DiskSize": 50,
					},
					"DataDisks": []map[string]any{
						{
							"DiskType": "CLOUD_PREMIUM",
							"DiskSize": 100,
						},
					},
				},
			},
			meta: RequestMeta{
				Method: http.MethodPost,
			},
			setupRequest: func(req *mcp.CallToolRequest) {
				req.Header = http.Header{}
				req.Header.Set("TC-API-SecretId", "test-id")
				req.Header.Set("TC-API-SecretKey", "test-key")
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var data map[string]any
				err := json.Unmarshal(body, &data)
				if err != nil {
					log.Printf("Error unmarshalling body: %v", err)
					return
				}

				if data["InstanceType"] != "S5.MEDIUM4" {
					t.Errorf("Expected InstanceType=S5.MEDIUM4, got %v", data["InstanceType"])
				}

				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(`{"Response":{"InstanceIdSet":["ins-xxx"]}}`))
				if err != nil {
					log.Printf("Error writing response: %v", err)
					return
				}
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务器
			var server *httptest.Server
			if tt.serverHandler != nil {
				server = httptest.NewServer(tt.serverHandler)
				defer server.Close()
			}

			// 设置URL
			if server != nil {
				tt.meta.URL = server.URL
			} else {
				tt.meta.URL = "http://example.com"
			}

			adapter := HTTPCAPIAdapter{}
			req := mcp.CallToolRequest{}
			if tt.setupRequest != nil {
				tt.setupRequest(&req)
			}

			data, err := adapter.DoRequest(context.Background(), req, tt.parameters, tt.meta)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
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

func TestCheckCommonParam(t *testing.T) {
	tests := []struct {
		name          string
		names         []string
		params        map[string]any
		expectedError bool
		errorContains string
	}{
		{
			name:  "all parameters present",
			names: []string{"Host", "Service", "Version"},
			params: map[string]any{
				"Host":    "example.com",
				"Service": "test",
				"Version": "1.0",
			},
			expectedError: false,
		},
		{
			name:  "missing one parameter",
			names: []string{"Host", "Service", "Version"},
			params: map[string]any{
				"Host":    "example.com",
				"Service": "test",
			},
			expectedError: true,
			errorContains: "missing required parameter: Version",
		},
		{
			name:  "missing multiple parameters",
			names: []string{"Host", "Service", "Version"},
			params: map[string]any{
				"Host": "example.com",
			},
			expectedError: true,
			errorContains: "missing required parameter: Service",
		},
		{
			name:          "empty parameter list",
			names:         []string{},
			params:        map[string]any{},
			expectedError: false,
		},
		{
			name:  "parameter with nil value",
			names: []string{"Host"},
			params: map[string]any{
				"Host": nil,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkCommonParam(tt.names, tt.params)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func BenchmarkHTTPCAPIAdapter_DoRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"Response":{"RequestId":"test"}}`))
		if err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}
	}))
	defer server.Close()

	adapter := HTTPCAPIAdapter{}
	parameters := Parameters{
		HeaderParams: map[string]any{
			"Host":    "cvm.tencentcloudapi.com",
			"Service": "cvm",
			"Version": "2017-03-12",
			"Action":  "DescribeInstances",
			"Region":  "ap-guangzhou",
		},
		BodyParams: map[string]any{
			"Limit": 10,
		},
	}
	meta := RequestMeta{
		URL:    server.URL,
		Method: http.MethodPost,
	}
	req := mcp.CallToolRequest{
		Header: http.Header{},
	}
	req.Header.Set("TC-API-SecretId", "test-id")
	req.Header.Set("TC-API-SecretKey", "test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.DoRequest(context.Background(), req, parameters, meta)
	}
}
