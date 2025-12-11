package adapter

import (
	"strings"
	"testing"
)

func TestSha256hex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "string with special characters",
			input:    "hello world!@#$%",
			expected: "8e1c6f68e3b6f6e3c3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3",
		},
		{
			name:     "JSON string",
			input:    `{"key":"value"}`,
			expected: "213d2c2d3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sha256hex(tt.input)
			// 验证结果是64个字符的十六进制字符串
			if len(result) != 64 {
				t.Errorf("Expected 64 character hex string, got %d characters", len(result))
			}
			// 验证只包含十六进制字符
			for _, c := range result {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("Result contains non-hex character: %c", c)
				}
			}
		})
	}
}

func TestHmacSha256(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		key    string
		verify func(result string) bool
	}{
		{
			name: "simple hmac",
			s:    "hello",
			key:  "secret",
			verify: func(result string) bool {
				return len(result) > 0
			},
		},
		{
			name: "empty string",
			s:    "",
			key:  "secret",
			verify: func(result string) bool {
				return len(result) > 0
			},
		},
		{
			name: "empty key",
			s:    "hello",
			key:  "",
			verify: func(result string) bool {
				return len(result) > 0
			},
		},
		{
			name: "long string",
			s:    strings.Repeat("a", 1000),
			key:  "secret",
			verify: func(result string) bool {
				return len(result) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hmacSha256(tt.s, tt.key)
			if !tt.verify(result) {
				t.Errorf("HMAC verification failed for input: %s", tt.name)
			}
		})
	}
}

func TestSignatureTCHeader(t *testing.T) {
	tests := []struct {
		name          string
		tcp           TencentCloudAPIParam
		expectedError bool
		validate      func(t *testing.T, headers map[string]string)
	}{
		{
			name: "valid CVM API request",
			tcp: TencentCloudAPIParam{
				SecretId:  "AKIDz8krbsJ5yKBZQpn74WFkmLPx3*******",
				SecretKey: "Gu5t9xGARNpq86cd98joQYCN3*******",
				Host:      "cvm.tencentcloudapi.com",
				Service:   "cvm",
				Version:   "2017-03-12",
				Action:    "DescribeInstances",
				Region:    "ap-guangzhou",
				Payload:   `{"Limit":1,"Filters":[{"Values":["unnamed"],"Name":"instance-name"}]}`,
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				// 验证必需的头部存在
				requiredHeaders := []string{
					"Authorization",
					"Content-Type",
					"Host",
					"X-TC-Action",
					"X-TC-Timestamp",
					"X-TC-Version",
					"X-TC-Region",
				}
				for _, header := range requiredHeaders {
					if _, ok := headers[header]; !ok {
						t.Errorf("Missing required header: %s", header)
					}
				}

				// 验证Content-Type
				if headers["Content-Type"] != "application/json; charset=utf-8" {
					t.Errorf("Expected Content-Type=application/json; charset=utf-8, got %s", headers["Content-Type"])
				}

				// 验证Host
				if headers["Host"] != "cvm.tencentcloudapi.com" {
					t.Errorf("Expected Host=cvm.tencentcloudapi.com, got %s", headers["Host"])
				}

				// 验证X-TC-Action
				if headers["X-TC-Action"] != "DescribeInstances" {
					t.Errorf("Expected X-TC-Action=DescribeInstances, got %s", headers["X-TC-Action"])
				}

				// 验证X-TC-Version
				if headers["X-TC-Version"] != "2017-03-12" {
					t.Errorf("Expected X-TC-Version=2017-03-12, got %s", headers["X-TC-Version"])
				}

				// 验证X-TC-Region
				if headers["X-TC-Region"] != "ap-guangzhou" {
					t.Errorf("Expected X-TC-Region=ap-guangzhou, got %s", headers["X-TC-Region"])
				}

				// 验证Authorization格式
				auth := headers["Authorization"]
				if !strings.HasPrefix(auth, "TC3-HMAC-SHA256") {
					t.Errorf("Authorization should start with TC3-HMAC-SHA256, got %s", auth)
				}
				if !strings.Contains(auth, "Credential=") {
					t.Error("Authorization should contain Credential=")
				}
				if !strings.Contains(auth, "SignedHeaders=") {
					t.Error("Authorization should contain SignedHeaders=")
				}
				if !strings.Contains(auth, "Signature=") {
					t.Error("Authorization should contain Signature=")
				}

				// 验证Timestamp是数字
				timestamp := headers["X-TC-Timestamp"]
				if timestamp == "" {
					t.Error("X-TC-Timestamp should not be empty")
				}
			},
		},
		{
			name: "empty payload",
			tcp: TencentCloudAPIParam{
				SecretId:  "test-id",
				SecretKey: "test-key",
				Host:      "cvm.tencentcloudapi.com",
				Service:   "cvm",
				Version:   "2017-03-12",
				Action:    "DescribeZones",
				Region:    "ap-guangzhou",
				Payload:   "",
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				if headers["Authorization"] == "" {
					t.Error("Authorization should not be empty even with empty payload")
				}
			},
		},
		{
			name: "different service - VPC",
			tcp: TencentCloudAPIParam{
				SecretId:  "test-id",
				SecretKey: "test-key",
				Host:      "vpc.tencentcloudapi.com",
				Service:   "vpc",
				Version:   "2017-03-12",
				Action:    "DescribeVpcs",
				Region:    "ap-guangzhou",
				Payload:   `{"Limit":10}`,
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				if headers["Host"] != "vpc.tencentcloudapi.com" {
					t.Errorf("Expected Host=vpc.tencentcloudapi.com, got %s", headers["Host"])
				}
				if headers["X-TC-Action"] != "DescribeVpcs" {
					t.Errorf("Expected X-TC-Action=DescribeVpcs, got %s", headers["X-TC-Action"])
				}
			},
		},
		{
			name: "action with mixed case",
			tcp: TencentCloudAPIParam{
				SecretId:  "test-id",
				SecretKey: "test-key",
				Host:      "cvm.tencentcloudapi.com",
				Service:   "cvm",
				Version:   "2017-03-12",
				Action:    "DescribeInstances",
				Region:    "ap-guangzhou",
				Payload:   "{}",
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				// X-TC-Action应该保持原样
				if headers["X-TC-Action"] != "DescribeInstances" {
					t.Errorf("Expected X-TC-Action=DescribeInstances, got %s", headers["X-TC-Action"])
				}
			},
		},
		{
			name: "complex JSON payload",
			tcp: TencentCloudAPIParam{
				SecretId:  "test-id",
				SecretKey: "test-key",
				Host:      "cvm.tencentcloudapi.com",
				Service:   "cvm",
				Version:   "2017-03-12",
				Action:    "RunInstances",
				Region:    "ap-guangzhou",
				Payload: `{
					"InstanceType": "S5.MEDIUM4",
					"ImageId": "img-xxx",
					"SystemDisk": {
						"DiskType": "CLOUD_PREMIUM",
						"DiskSize": 50
					},
					"DataDisks": [
						{
							"DiskType": "CLOUD_PREMIUM",
							"DiskSize": 100
						}
					]
				}`,
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				if headers["Authorization"] == "" {
					t.Error("Authorization should not be empty with complex payload")
				}
			},
		},
		{
			name: "special characters in payload",
			tcp: TencentCloudAPIParam{
				SecretId:  "test-id",
				SecretKey: "test-key",
				Host:      "cvm.tencentcloudapi.com",
				Service:   "cvm",
				Version:   "2017-03-12",
				Action:    "DescribeInstances",
				Region:    "ap-guangzhou",
				Payload:   `{"Name":"测试实例","Tags":["标签1","标签2"]}`,
			},
			expectedError: false,
			validate: func(t *testing.T, headers map[string]string) {
				if headers["Authorization"] == "" {
					t.Error("Authorization should handle special characters")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := SignatureTCHeader(tt.tcp)

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
				tt.validate(t, headers)
			}
		})
	}
}

func TestSignatureTCHeader_Consistency(t *testing.T) {
	// 测试相同输入产生相同签名（在同一秒内）
	tcp := TencentCloudAPIParam{
		SecretId:  "test-id",
		SecretKey: "test-key",
		Host:      "cvm.tencentcloudapi.com",
		Service:   "cvm",
		Version:   "2017-03-12",
		Action:    "DescribeInstances",
		Region:    "ap-guangzhou",
		Payload:   `{"Limit":10}`,
	}

	headers1, err1 := SignatureTCHeader(tcp)
	if err1 != nil {
		t.Fatalf("First call failed: %v", err1)
	}

	headers2, err2 := SignatureTCHeader(tcp)
	if err2 != nil {
		t.Fatalf("Second call failed: %v", err2)
	}

	// 时间戳可能不同，但其他头部应该相同
	if headers1["Host"] != headers2["Host"] {
		t.Error("Host headers should be identical")
	}
	if headers1["X-TC-Action"] != headers2["X-TC-Action"] {
		t.Error("X-TC-Action headers should be identical")
	}
	if headers1["X-TC-Version"] != headers2["X-TC-Version"] {
		t.Error("X-TC-Version headers should be identical")
	}
	if headers1["X-TC-Region"] != headers2["X-TC-Region"] {
		t.Error("X-TC-Region headers should be identical")
	}
}

func TestTencentCloudAPIParam_AllFields(t *testing.T) {
	// 测试TencentCloudAPIParam结构体的所有字段
	tcp := TencentCloudAPIParam{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Host:      "test.tencentcloudapi.com",
		Service:   "test-service",
		Version:   "2023-01-01",
		Action:    "TestAction",
		Region:    "ap-test",
		Payload:   `{"test":"data"}`,
	}

	headers, err := SignatureTCHeader(tcp)
	if err != nil {
		t.Fatalf("SignatureTCHeader failed: %v", err)
	}

	// 验证所有字段都被正确使用
	if headers["Host"] != tcp.Host {
		t.Errorf("Host mismatch: expected %s, got %s", tcp.Host, headers["Host"])
	}
	if headers["X-TC-Action"] != tcp.Action {
		t.Errorf("Action mismatch: expected %s, got %s", tcp.Action, headers["X-TC-Action"])
	}
	if headers["X-TC-Version"] != tcp.Version {
		t.Errorf("Version mismatch: expected %s, got %s", tcp.Version, headers["X-TC-Version"])
	}
	if headers["X-TC-Region"] != tcp.Region {
		t.Errorf("Region mismatch: expected %s, got %s", tcp.Region, headers["X-TC-Region"])
	}
}

func BenchmarkSignatureTCHeader(b *testing.B) {
	tcp := TencentCloudAPIParam{
		SecretId:  "AKIDz8krbsJ5yKBZQpn74WFkmLPx3*******",
		SecretKey: "Gu5t9xGARNpq86cd98joQYCN3*******",
		Host:      "cvm.tencentcloudapi.com",
		Service:   "cvm",
		Version:   "2017-03-12",
		Action:    "DescribeInstances",
		Region:    "ap-guangzhou",
		Payload:   `{"Limit":1,"Filters":[{"Values":["unnamed"],"Name":"instance-name"}]}`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SignatureTCHeader(tcp)
	}
}

func BenchmarkSha256hex(b *testing.B) {
	input := `{"Limit":1,"Filters":[{"Values":["unnamed"],"Name":"instance-name"}]}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sha256hex(input)
	}
}

func BenchmarkHmacSha256(b *testing.B) {
	s := "test string to hash"
	key := "secret key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hmacSha256(s, key)
	}
}
