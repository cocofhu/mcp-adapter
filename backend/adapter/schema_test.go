package adapter

import (
	"mcp-adapter/backend/models"
	"testing"
)

// TestSatisfySchema_BasicTypes 测试基础类型的schema验证
func TestSatisfySchema_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "string type - valid",
			schema: map[string]any{
				"type": "string",
			},
			data: map[string]any{
				"value": "hello",
			},
			expected: false, // 因为schema不是object类型，直接传string会失败
		},
		{
			name: "number type - valid float64",
			schema: map[string]any{
				"type": "number",
			},
			data: map[string]any{
				"value": 42.5,
			},
			expected: false,
		},
		{
			name: "boolean type - valid",
			schema: map[string]any{
				"type": "boolean",
			},
			data: map[string]any{
				"value": true,
			},
			expected: false,
		},
		{
			name: "object with string property - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"name": "John",
			},
			expected: true,
		},
		{
			name: "object with number property - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"age": map[string]any{
						"type": "number",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"age": 25.0,
			},
			expected: true,
		},
		{
			name: "object with boolean property - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"active": map[string]any{
						"type": "boolean",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"active": true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_RequiredFields 测试必填字段验证
func TestSatisfySchema_RequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "required field present",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{"name"},
			},
			data: map[string]any{
				"name": "John",
			},
			expected: true,
		},
		{
			name: "required field missing",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{"name"},
			},
			data:     map[string]any{},
			expected: false,
		},
		{
			name: "multiple required fields - all present",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"age": map[string]any{
						"type": "number",
					},
				},
				"required": []any{"name", "age"},
			},
			data: map[string]any{
				"name": "John",
				"age":  25.0,
			},
			expected: true,
		},
		{
			name: "multiple required fields - one missing",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"age": map[string]any{
						"type": "number",
					},
				},
				"required": []any{"name", "age"},
			},
			data: map[string]any{
				"name": "John",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_ArrayTypes 测试数组类型验证
func TestSatisfySchema_ArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "array of strings - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": []any{"tag1", "tag2", "tag3"},
			},
			expected: true,
		},
		{
			name: "array of numbers - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"scores": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"scores": []any{90.5, 85.0, 92.3},
			},
			expected: true,
		},
		{
			name: "array of objects - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"users": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type": "string",
								},
								"age": map[string]any{
									"type": "number",
								},
							},
							"required": []any{},
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"users": []any{
					map[string]any{"name": "John", "age": 25.0},
					map[string]any{"name": "Jane", "age": 30.0},
				},
			},
			expected: true,
		},
		{
			name: "array with wrong item type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": []any{"tag1", 123, "tag3"}, // 包含number
			},
			expected: false,
		},
		{
			name: "empty array - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": []any{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_NestedObjects 测试嵌套对象验证
func TestSatisfySchema_NestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "nested object - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type": "string",
							},
							"address": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"city": map[string]any{
										"type": "string",
									},
									"zipcode": map[string]any{
										"type": "string",
									},
								},
								"required": []any{},
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"user": map[string]any{
					"name": "John",
					"address": map[string]any{
						"city":    "Beijing",
						"zipcode": "100000",
					},
				},
			},
			expected: true,
		},
		{
			name: "deeply nested object - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"level1": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"level2": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"level3": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"value": map[string]any{
												"type": "string",
											},
										},
										"required": []any{},
									},
								},
								"required": []any{},
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"value": "deep",
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_TypeMismatch 测试类型不匹配的情况
func TestSatisfySchema_TypeMismatch(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "string expected, number provided",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"name": 123,
			},
			expected: false,
		},
		{
			name: "number expected, string provided",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"age": map[string]any{
						"type": "number",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"age": "25",
			},
			expected: false,
		},
		{
			name: "boolean expected, string provided",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"active": map[string]any{
						"type": "boolean",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"active": "true",
			},
			expected: false,
		},
		{
			name: "array expected, object provided",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": map[string]any{"tag": "value"},
			},
			expected: false,
		},
		{
			name: "object expected, array provided",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type": "string",
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"user": []any{"John"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_EdgeCases 测试边界情况
func TestSatisfySchema_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name:     "nil schema",
			schema:   nil,
			data:     map[string]any{"key": "value"},
			expected: true,
		},
		{
			name: "nil data",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []any{},
			},
			data:     nil,
			expected: true,
		},
		{
			name:     "both nil",
			schema:   nil,
			data:     nil,
			expected: true,
		},
		{
			name: "empty object schema with empty data",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []any{},
			},
			data:     map[string]any{},
			expected: true,
		},
		{
			name: "extra fields in data (should be ignored)",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"name":  "John",
				"age":   25,
				"email": "john@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFilterDataBySchema_BasicTypes 测试基础类型的数据过滤
func TestFilterDataBySchema_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected map[string]any
	}{
		{
			name: "filter string field",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
			},
			data: map[string]any{
				"name": "John",
			},
			expected: map[string]any{
				"name": "John",
			},
		},
		{
			name: "filter number field",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"age": map[string]any{
						"type": "number",
					},
				},
			},
			data: map[string]any{
				"age": 25.0,
			},
			expected: map[string]any{
				"age": 25.0,
			},
		},
		{
			name: "filter boolean field",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"active": map[string]any{
						"type": "boolean",
					},
				},
			},
			data: map[string]any{
				"active": true,
			},
			expected: map[string]any{
				"active": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataBySchema(tt.schema, tt.data)
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("FilterDataBySchema() returned non-map type")
			}

			if len(resultMap) != len(tt.expected) {
				t.Errorf("FilterDataBySchema() returned %d fields, want %d", len(resultMap), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if resultMap[key] != expectedValue {
					t.Errorf("FilterDataBySchema()[%s] = %v, want %v", key, resultMap[key], expectedValue)
				}
			}
		})
	}
}

// TestFilterDataBySchema_RemoveExtraFields 测试移除额外字段
func TestFilterDataBySchema_RemoveExtraFields(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"age": map[string]any{
				"type": "number",
			},
		},
	}

	data := map[string]any{
		"name":    "John",
		"age":     25.0,
		"email":   "john@example.com", // 额外字段
		"address": "Beijing",          // 额外字段
		"phone":   "123456789",        // 额外字段
	}

	result := FilterDataBySchema(schema, data)
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("FilterDataBySchema() returned non-map type")
	}

	// 应该只保留name和age
	if len(resultMap) != 2 {
		t.Errorf("FilterDataBySchema() returned %d fields, want 2", len(resultMap))
	}

	if resultMap["name"] != "John" {
		t.Errorf("FilterDataBySchema()[name] = %v, want John", resultMap["name"])
	}

	if resultMap["age"] != 25.0 {
		t.Errorf("FilterDataBySchema()[age] = %v, want 25.0", resultMap["age"])
	}

	// 确保额外字段被移除
	if _, exists := resultMap["email"]; exists {
		t.Error("FilterDataBySchema() should remove email field")
	}
	if _, exists := resultMap["address"]; exists {
		t.Error("FilterDataBySchema() should remove address field")
	}
	if _, exists := resultMap["phone"]; exists {
		t.Error("FilterDataBySchema() should remove phone field")
	}
}

// TestFilterDataBySchema_ArrayTypes 测试数组类型的过滤
func TestFilterDataBySchema_ArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		validate func(t *testing.T, result any)
	}{
		{
			name: "filter array of strings",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
			data: map[string]any{
				"tags": []any{"tag1", "tag2", "tag3"},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				tags := resultMap["tags"].([]any)
				if len(tags) != 3 {
					t.Errorf("Expected 3 tags, got %d", len(tags))
				}
			},
		},
		{
			name: "filter array with mixed types - remove invalid items",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
			data: map[string]any{
				"tags": []any{"tag1", 123, "tag3", true}, // 包含非string类型
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				tags := resultMap["tags"].([]any)
				// 应该只保留string类型的元素
				if len(tags) != 2 {
					t.Errorf("Expected 2 valid tags, got %d", len(tags))
				}
			},
		},
		{
			name: "filter array of objects",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"users": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			},
			data: map[string]any{
				"users": []any{
					map[string]any{"name": "John", "age": 25},   // age应该被过滤
					map[string]any{"name": "Jane", "email": ""}, // email应该被过滤
				},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				users := resultMap["users"].([]any)
				if len(users) != 2 {
					t.Errorf("Expected 2 users, got %d", len(users))
				}
				// 检查第一个用户只有name字段
				user1 := users[0].(map[string]any)
				if len(user1) != 1 {
					t.Errorf("Expected user to have 1 field, got %d", len(user1))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataBySchema(tt.schema, tt.data)
			tt.validate(t, result)
		})
	}
}

// TestFilterDataBySchema_NestedObjects 测试嵌套对象的过滤
func TestFilterDataBySchema_NestedObjects(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"address": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"city": map[string]any{
								"type": "string",
							},
						},
					},
				},
			},
		},
	}

	data := map[string]any{
		"user": map[string]any{
			"name":  "John",
			"age":   25,    // 应该被过滤
			"email": "...", // 应该被过滤
			"address": map[string]any{
				"city":    "Beijing",
				"zipcode": "100000", // 应该被过滤
				"country": "China",  // 应该被过滤
			},
		},
		"extra": "field", // 应该被过滤
	}

	result := FilterDataBySchema(schema, data)
	resultMap := result.(map[string]any)

	// 检查顶层只有user字段
	if len(resultMap) != 1 {
		t.Errorf("Expected 1 top-level field, got %d", len(resultMap))
	}

	user := resultMap["user"].(map[string]any)
	// user应该只有name和address
	if len(user) != 2 {
		t.Errorf("Expected user to have 2 fields, got %d", len(user))
	}

	address := user["address"].(map[string]any)
	// address应该只有city
	if len(address) != 1 {
		t.Errorf("Expected address to have 1 field, got %d", len(address))
	}

	if address["city"] != "Beijing" {
		t.Errorf("Expected city to be Beijing, got %v", address["city"])
	}
}

// TestFilterDataBySchema_EdgeCases 测试边界情况
func TestFilterDataBySchema_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected any
	}{
		{
			name:     "nil schema",
			schema:   nil,
			data:     map[string]any{"key": "value"},
			expected: map[string]any{"key": "value"},
		},
		{
			name: "nil data",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			data:     nil,
			expected: nil,
		},
		{
			name: "empty properties",
			schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			data: map[string]any{
				"field1": "value1",
				"field2": "value2",
			},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataBySchema(tt.schema, tt.data)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("FilterDataBySchema() = %v, want nil", result)
				}
				return
			}

			expectedMap, ok1 := tt.expected.(map[string]any)
			resultMap, ok2 := result.(map[string]any)

			if ok1 && ok2 {
				if len(resultMap) != len(expectedMap) {
					t.Errorf("FilterDataBySchema() returned %d fields, want %d", len(resultMap), len(expectedMap))
				}
			}
		})
	}
}

// TestSatisfySchema_NumberTypes 测试各种数字类型
func TestSatisfySchema_NumberTypes(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"value": map[string]any{
				"type": "number",
			},
		},
		"required": []any{},
	}

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"float64", 42.5, true},
		{"float32", float32(42.5), true},
		{"int", int(42), true},
		{"int64", int64(42), true},
		{"int32", int32(42), true},
		{"uint", uint(42), true},
		{"uint64", uint64(42), true},
		{"string", "42", false},
		{"boolean", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{
				"value": tt.value,
			}
			result := SatisfySchema(schema, data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() with %s = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_NestedRequired 测试嵌套对象的required字段验证
func TestSatisfySchema_NestedRequired(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "nested object with required field - missing",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"Response": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"TotalCount1": map[string]any{
								"type":        "number",
								"description": "",
							},
						},
						"required":    []any{"TotalCount1"},
						"description": "",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"Response": map[string]any{
					"TotalCount": 1,
				},
			},
			expected: false, // TotalCount1是必填的但缺失了
		},
		{
			name: "nested object with required field - present",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"Response": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"TotalCount1": map[string]any{
								"type":        "number",
								"description": "",
							},
						},
						"required":    []any{"TotalCount1"},
						"description": "",
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"Response": map[string]any{
					"TotalCount1": 100.0,
				},
			},
			expected: true,
		},
		{
			name: "deeply nested required field - missing",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"Level1": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"Level2": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"RequiredField": map[string]any{
										"type": "string",
									},
								},
								"required": []any{"RequiredField"},
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"Level1": map[string]any{
					"Level2": map[string]any{},
				},
			},
			expected: false,
		},
		{
			name: "multiple nested required fields - all present",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"User": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"Name": map[string]any{
								"type": "string",
							},
							"Age": map[string]any{
								"type": "number",
							},
						},
						"required": []any{"Name", "Age"},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"User": map[string]any{
					"Name": "John",
					"Age":  25.0,
				},
			},
			expected: true,
		},
		{
			name: "multiple nested required fields - one missing",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"User": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"Name": map[string]any{
								"type": "string",
							},
							"Age": map[string]any{
								"type": "number",
							},
						},
						"required": []any{"Name", "Age"},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"User": map[string]any{
					"Name": "John",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v\nSchema: %+v\nData: %+v",
					result, tt.expected, tt.schema, tt.data)
			}
		})
	}
}

// TestSatisfySchema_ReflectionArrayTypes 测试反射实现对各种数组类型的支持
func TestSatisfySchema_ReflectionArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "[]string type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
			expected: true,
		},
		{
			name: "[]int type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"scores": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"scores": []int{90, 85, 92},
			},
			expected: true,
		},
		{
			name: "[]int64 type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ids": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"ids": []int64{1001, 1002, 1003},
			},
			expected: true,
		},
		{
			name: "[]float64 type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prices": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"prices": []float64{19.99, 29.99, 39.99},
			},
			expected: true,
		},
		{
			name: "[]bool type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"flags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "boolean",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"flags": []bool{true, false, true},
			},
			expected: true,
		},
		{
			name: "array type - not a slice",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"tags": "not an array",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSatisfySchema_ReflectionMapTypes 测试反射实现对各种map类型的支持
func TestSatisfySchema_ReflectionMapTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "map[string]string type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"metadata": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"key": map[string]any{
								"type": "string",
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"metadata": map[string]string{
					"key":    "value",
					"author": "John",
				},
			},
			expected: true,
		},
		{
			name: "map[string]int type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"counts": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"total": map[string]any{
								"type": "number",
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"counts": map[string]int{
					"total":   100,
					"pending": 20,
				},
			},
			expected: true,
		},
		{
			name: "map[string]interface{} type - valid",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"config": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"enabled": map[string]any{
								"type": "boolean",
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"config": map[string]interface{}{
					"enabled": true,
					"timeout": 30,
				},
			},
			expected: true,
		},
		{
			name: "object type - not a map",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type": "string",
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"user": "not a map",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFilterDataBySchema_ReflectionArrayTypes 测试反射实现对各种数组类型的过滤
func TestFilterDataBySchema_ReflectionArrayTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		validate func(t *testing.T, result any)
	}{
		{
			name: "filter []string type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
			data: map[string]any{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				tags := resultMap["tags"].([]any)
				if len(tags) != 3 {
					t.Errorf("Expected 3 tags, got %d", len(tags))
				}
				if tags[0] != "tag1" || tags[1] != "tag2" || tags[2] != "tag3" {
					t.Errorf("Tag values don't match expected")
				}
			},
		},
		{
			name: "filter []int type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"scores": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
			},
			data: map[string]any{
				"scores": []int{90, 85, 92},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				scores := resultMap["scores"].([]any)
				if len(scores) != 3 {
					t.Errorf("Expected 3 scores, got %d", len(scores))
				}
			},
		},
		{
			name: "filter []float64 type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prices": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
			},
			data: map[string]any{
				"prices": []float64{19.99, 29.99, 39.99},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				prices := resultMap["prices"].([]any)
				if len(prices) != 3 {
					t.Errorf("Expected 3 prices, got %d", len(prices))
				}
			},
		},
		{
			name: "filter mixed type array - keep only valid items",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"id": map[string]any{
									"type": "number",
								},
								"name": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			},
			data: map[string]any{
				"items": []map[string]any{
					{"id": 1, "name": "Item1", "extra": "field"},
					{"id": 2, "name": "Item2", "description": "desc"},
				},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				items := resultMap["items"].([]any)
				if len(items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(items))
				}
				// 检查额外字段是否被过滤
				item1 := items[0].(map[string]any)
				if _, exists := item1["extra"]; exists {
					t.Error("Extra field should be filtered out")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataBySchema(tt.schema, tt.data)
			tt.validate(t, result)
		})
	}
}

// TestFilterDataBySchema_ReflectionMapTypes 测试反射实现对各种map类型的过滤
func TestFilterDataBySchema_ReflectionMapTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		validate func(t *testing.T, result any)
	}{
		{
			name: "filter map[string]string type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"metadata": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"author": map[string]any{
								"type": "string",
							},
						},
					},
				},
			},
			data: map[string]any{
				"metadata": map[string]string{
					"author":  "John",
					"version": "1.0",
					"status":  "active",
				},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				metadata := resultMap["metadata"].(map[string]any)
				// 应该只保留schema中定义的author字段
				if len(metadata) != 1 {
					t.Errorf("Expected 1 field in metadata, got %d", len(metadata))
				}
				if metadata["author"] != "John" {
					t.Errorf("Expected author to be John, got %v", metadata["author"])
				}
			},
		},
		{
			name: "filter map[string]int type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"counts": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"total": map[string]any{
								"type": "number",
							},
						},
					},
				},
			},
			data: map[string]any{
				"counts": map[string]int{
					"total":     100,
					"pending":   20,
					"completed": 80,
				},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				counts := resultMap["counts"].(map[string]any)
				// 应该只保留total字段
				if len(counts) != 1 {
					t.Errorf("Expected 1 field in counts, got %d", len(counts))
				}
				if counts["total"] != 100 {
					t.Errorf("Expected total to be 100, got %v", counts["total"])
				}
			},
		},
		{
			name: "filter nested map[string]interface{} type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"config": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"database": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"host": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			data: map[string]any{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"host":     "localhost",
						"port":     3306,
						"username": "admin",
					},
					"cache": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			validate: func(t *testing.T, result any) {
				resultMap := result.(map[string]any)
				config := resultMap["config"].(map[string]any)
				// cache应该被过滤掉
				if _, exists := config["cache"]; exists {
					t.Error("cache field should be filtered out")
				}
				database := config["database"].(map[string]any)
				// 只保留host字段
				if len(database) != 1 {
					t.Errorf("Expected 1 field in database, got %d", len(database))
				}
				if database["host"] != "localhost" {
					t.Errorf("Expected host to be localhost, got %v", database["host"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataBySchema(tt.schema, tt.data)
			tt.validate(t, result)
		})
	}
}

// TestSatisfySchema_ComplexReflectionTypes 测试复杂的反射类型组合
func TestSatisfySchema_ComplexReflectionTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		data     map[string]any
		expected bool
	}{
		{
			name: "array of maps with concrete types",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"users": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type": "string",
								},
								"age": map[string]any{
									"type": "number",
								},
							},
							"required": []any{"name"},
						},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"users": []map[string]interface{}{
					{"name": "John", "age": 25},
					{"name": "Jane", "age": 30},
				},
			},
			expected: true,
		},
		{
			name: "map containing arrays of concrete types",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"groups": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"admins": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "string",
								},
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"groups": map[string][]string{
					"admins": {"user1", "user2"},
					"users":  {"user3", "user4"},
				},
			},
			expected: true,
		},
		{
			name: "deeply nested with mixed concrete types",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"data": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"items": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"tags": map[string]any{
											"type": "array",
											"items": map[string]any{
												"type": "string",
											},
										},
									},
									"required": []any{},
								},
							},
						},
						"required": []any{},
					},
				},
				"required": []any{},
			},
			data: map[string]any{
				"data": map[string]interface{}{
					"items": []map[string][]string{
						{"tags": []string{"tag1", "tag2"}},
						{"tags": []string{"tag3", "tag4"}},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SatisfySchema(tt.schema, tt.data)
			if result != tt.expected {
				t.Errorf("SatisfySchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// BenchmarkSatisfySchema 性能基准测试
func BenchmarkSatisfySchema(b *testing.B) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"age": map[string]any{
				"type": "number",
			},
			"tags": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []any{"name"},
	}

	data := map[string]any{
		"name": "John",
		"age":  25.0,
		"tags": []any{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SatisfySchema(schema, data)
	}
}

// BenchmarkFilterDataBySchema 性能基准测试
func BenchmarkFilterDataBySchema(b *testing.B) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"age": map[string]any{
				"type": "number",
			},
		},
	}

	data := map[string]any{
		"name":    "John",
		"age":     25.0,
		"email":   "john@example.com",
		"address": "Beijing",
		"phone":   "123456789",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterDataBySchema(schema, data)
	}
}

// TestSchemaBuilder_BuildFieldTypeSchema 测试构建字段类型schema
func TestSchemaBuilder_BuildFieldTypeSchema(t *testing.T) {
	// 创建测试用的schemaBuilder
	refId2 := int64(2)

	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "User",
				Description: "User type",
			},
			2: {
				ID:          2,
				Name:        "Address",
				Description: "Address type",
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:          1,
					Name:        "name",
					Type:        "string",
					Description: "User name",
					Required:    true,
					IsArray:     false,
				},
				{
					ID:          2,
					Name:        "age",
					Type:        "number",
					Description: "User age",
					Required:    false,
					IsArray:     false,
				},
			},
			2: {
				{
					ID:          3,
					Name:        "city",
					Type:        "string",
					Description: "City name",
					Required:    true,
					IsArray:     false,
				},
			},
		},
	}

	tests := []struct {
		name      string
		field     *models.CustomTypeField
		expected  map[string]any
		expectErr bool
	}{
		{
			name: "basic string type",
			field: &models.CustomTypeField{
				Name:        "username",
				Type:        "string",
				Description: "Username field",
			},
			expected: map[string]any{
				"type":        "string",
				"description": "Username field",
			},
			expectErr: false,
		},
		{
			name: "basic number type",
			field: &models.CustomTypeField{
				Name:        "count",
				Type:        "number",
				Description: "Count field",
			},
			expected: map[string]any{
				"type":        "number",
				"description": "Count field",
			},
			expectErr: false,
		},
		{
			name: "basic boolean type",
			field: &models.CustomTypeField{
				Name:        "enabled",
				Type:        "boolean",
				Description: "Enabled flag",
			},
			expected: map[string]any{
				"type":        "boolean",
				"description": "Enabled flag",
			},
			expectErr: false,
		},
		{
			name: "custom type with valid ref",
			field: &models.CustomTypeField{
				Name:        "address",
				Type:        "custom",
				Description: "Address field",
				Ref:         &refId2,
			},
			expected: map[string]any{
				"type":        "object",
				"description": "Address type",
				"properties": map[string]any{
					"city": map[string]any{
						"type":        "string",
						"description": "City name",
					},
				},
				"required": []string{"city"},
			},
			expectErr: false,
		},
		{
			name: "custom type without ref - should error",
			field: &models.CustomTypeField{
				Name:        "invalid",
				Type:        "custom",
				Description: "Invalid custom field",
				Ref:         nil,
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newBuildContext()
			result, err := builder.buildFieldTypeSchema(tt.field, ctx)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// 验证结果
			if !compareSchemas(result, tt.expected) {
				t.Errorf("Schema mismatch.\nGot: %+v\nWant: %+v", result, tt.expected)
			}
		})
	}
}

// TestSchemaBuilder_BuildSchemaByField 测试构建字段schema（包含数组处理）
func TestSchemaBuilder_BuildSchemaByField(t *testing.T) {
	refId := int64(1)

	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "Tag",
				Description: "Tag type",
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:          1,
					Name:        "name",
					Type:        "string",
					Description: "Tag name",
					Required:    true,
					IsArray:     false,
				},
			},
		},
	}

	tests := []struct {
		name      string
		field     *models.CustomTypeField
		expected  map[string]any
		expectErr bool
	}{
		{
			name: "non-array string field",
			field: &models.CustomTypeField{
				Name:        "title",
				Type:        "string",
				Description: "Title field",
				IsArray:     false,
			},
			expected: map[string]any{
				"type":        "string",
				"description": "Title field",
			},
			expectErr: false,
		},
		{
			name: "array of strings",
			field: &models.CustomTypeField{
				Name:        "tags",
				Type:        "string",
				Description: "Tags array",
				IsArray:     true,
			},
			expected: map[string]any{
				"type":        "array",
				"description": "Tags array",
				"items": map[string]any{
					"type":        "string",
					"description": "Tags array",
				},
			},
			expectErr: false,
		},
		{
			name: "array of numbers",
			field: &models.CustomTypeField{
				Name:        "scores",
				Type:        "number",
				Description: "Score array",
				IsArray:     true,
			},
			expected: map[string]any{
				"type":        "array",
				"description": "Score array",
				"items": map[string]any{
					"type":        "number",
					"description": "Score array",
				},
			},
			expectErr: false,
		},
		{
			name: "array of custom type",
			field: &models.CustomTypeField{
				Name:        "tags",
				Type:        "custom",
				Description: "Tag objects",
				IsArray:     true,
				Ref:         &refId,
			},
			expected: map[string]any{
				"type":        "array",
				"description": "Tag objects",
				"items": map[string]any{
					"type":        "object",
					"description": "Tag type",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Tag name",
						},
					},
					"required": []string{"name"},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newBuildContext()
			result, err := builder.buildSchemaByField(tt.field, ctx)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !compareSchemas(result, tt.expected) {
				t.Errorf("Schema mismatch.\nGot: %+v\nWant: %+v", result, tt.expected)
			}
		})
	}
}

// TestSchemaBuilder_BuildSchemaByType 测试构建完整的对象schema
func TestSchemaBuilder_BuildSchemaByType(t *testing.T) {
	refId := int64(2)

	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "User",
				Description: "User information",
			},
			2: {
				ID:          2,
				Name:        "Address",
				Description: "Address information",
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:          1,
					Name:        "name",
					Type:        "string",
					Description: "User name",
					Required:    true,
					IsArray:     false,
				},
				{
					ID:          2,
					Name:        "age",
					Type:        "number",
					Description: "User age",
					Required:    false,
					IsArray:     false,
				},
				{
					ID:          3,
					Name:        "email",
					Type:        "string",
					Description: "User email",
					Required:    true,
					IsArray:     false,
				},
				{
					ID:          4,
					Name:        "address",
					Type:        "custom",
					Description: "User address",
					Required:    false,
					IsArray:     false,
					Ref:         &refId,
				},
			},
			2: {
				{
					ID:          5,
					Name:        "city",
					Type:        "string",
					Description: "City name",
					Required:    true,
					IsArray:     false,
				},
				{
					ID:          6,
					Name:        "zipcode",
					Type:        "string",
					Description: "Zip code",
					Required:    false,
					IsArray:     false,
				},
			},
		},
	}

	tests := []struct {
		name      string
		typeId    int64
		expected  map[string]any
		expectErr bool
	}{
		{
			name:   "simple type with basic fields",
			typeId: 2,
			expected: map[string]any{
				"type":        "object",
				"description": "Address information",
				"properties": map[string]any{
					"city": map[string]any{
						"type":        "string",
						"description": "City name",
					},
					"zipcode": map[string]any{
						"type":        "string",
						"description": "Zip code",
					},
				},
				"required": []string{"city"},
			},
			expectErr: false,
		},
		{
			name:   "complex type with nested custom type",
			typeId: 1,
			expected: map[string]any{
				"type":        "object",
				"description": "User information",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "User name",
					},
					"age": map[string]any{
						"type":        "number",
						"description": "User age",
					},
					"email": map[string]any{
						"type":        "string",
						"description": "User email",
					},
					"address": map[string]any{
						"type":        "object",
						"description": "Address information",
						"properties": map[string]any{
							"city": map[string]any{
								"type":        "string",
								"description": "City name",
							},
							"zipcode": map[string]any{
								"type":        "string",
								"description": "Zip code",
							},
						},
						"required": []string{"city"},
					},
				},
				"required": []string{"name", "email"},
			},
			expectErr: false,
		},
		{
			name:      "non-existent type - should error",
			typeId:    999,
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newBuildContext()
			result, err := builder.buildSchemaByType(tt.typeId, ctx)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !compareSchemas(result, tt.expected) {
				t.Errorf("Schema mismatch.\nGot: %+v\nWant: %+v", result, tt.expected)
			}
		})
	}
}

// TestSchemaBuilder_ComplexNestedTypes 测试复杂嵌套类型
func TestSchemaBuilder_ComplexNestedTypes(t *testing.T) {
	addressRefId := int64(2)
	contactRefId := int64(3)

	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "Company",
				Description: "Company information",
			},
			2: {
				ID:          2,
				Name:        "Address",
				Description: "Address details",
			},
			3: {
				ID:          3,
				Name:        "Contact",
				Description: "Contact information",
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					Name:        "name",
					Type:        "string",
					Description: "Company name",
					Required:    true,
					IsArray:     false,
				},
				{
					Name:        "addresses",
					Type:        "custom",
					Description: "Company addresses",
					Required:    false,
					IsArray:     true,
					Ref:         &addressRefId,
				},
				{
					Name:        "contact",
					Type:        "custom",
					Description: "Primary contact",
					Required:    true,
					IsArray:     false,
					Ref:         &contactRefId,
				},
			},
			2: {
				{
					Name:        "street",
					Type:        "string",
					Description: "Street address",
					Required:    true,
					IsArray:     false,
				},
				{
					Name:        "city",
					Type:        "string",
					Description: "City",
					Required:    true,
					IsArray:     false,
				},
			},
			3: {
				{
					Name:        "email",
					Type:        "string",
					Description: "Email address",
					Required:    true,
					IsArray:     false,
				},
				{
					Name:        "phones",
					Type:        "string",
					Description: "Phone numbers",
					Required:    false,
					IsArray:     true,
				},
			},
		},
	}

	ctx := newBuildContext()
	result, err := builder.buildSchemaByType(1, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证顶层结构
	if result["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", result["type"])
	}

	properties := result["properties"].(map[string]any)

	// 验证数组类型的嵌套对象
	addresses := properties["addresses"].(map[string]any)
	if addresses["type"] != "array" {
		t.Errorf("Expected addresses type 'array', got %v", addresses["type"])
	}

	addressItems := addresses["items"].(map[string]any)
	if addressItems["type"] != "object" {
		t.Errorf("Expected address items type 'object', got %v", addressItems["type"])
	}

	// 验证嵌套对象
	contact := properties["contact"].(map[string]any)
	if contact["type"] != "object" {
		t.Errorf("Expected contact type 'object', got %v", contact["type"])
	}

	contactProps := contact["properties"].(map[string]any)
	phones := contactProps["phones"].(map[string]any)
	if phones["type"] != "array" {
		t.Errorf("Expected phones type 'array', got %v", phones["type"])
	}

	// 验证required字段
	required := result["required"].([]string)
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

// TestSchemaBuilder_GetCustomType 测试获取自定义类型
func TestSchemaBuilder_GetCustomType(t *testing.T) {
	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "TestType",
				Description: "Test description",
			},
		},
		fields: map[int64][]models.CustomTypeField{},
	}

	tests := []struct {
		name      string
		typeId    int64
		expectErr bool
	}{
		{
			name:      "existing type",
			typeId:    1,
			expectErr: false,
		},
		{
			name:      "non-existing type",
			typeId:    999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.getCustomType(tt.typeId)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

// TestSchemaBuilder_GetCustomTypeFields 测试获取自定义类型字段
func TestSchemaBuilder_GetCustomTypeFields(t *testing.T) {
	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:   1,
					Name: "field1",
					Type: "string",
				},
				{
					ID:   2,
					Name: "field2",
					Type: "number",
				},
			},
		},
	}

	tests := []struct {
		name          string
		typeId        int64
		expectedCount int
	}{
		{
			name:          "type with fields",
			typeId:        1,
			expectedCount: 2,
		},
		{
			name:          "type without fields",
			typeId:        999,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.getCustomTypeFields(tt.typeId)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d fields, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

// compareSchemas 辅助函数：深度比较两个schema是否相等
func compareSchemas(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}

		// 处理不同类型的值
		switch aTyped := aVal.(type) {
		case map[string]any:
			bTyped, ok := bVal.(map[string]any)
			if !ok {
				return false
			}
			if !compareSchemas(aTyped, bTyped) {
				return false
			}
		case []string:
			bTyped, ok := bVal.([]string)
			if !ok {
				return false
			}
			if len(aTyped) != len(bTyped) {
				return false
			}
			for i := range aTyped {
				if aTyped[i] != bTyped[i] {
					return false
				}
			}
		case []any:
			bTyped, ok := bVal.([]any)
			if !ok {
				return false
			}
			if len(aTyped) != len(bTyped) {
				return false
			}
			// 简化处理：只比较长度
		default:
			if aVal != bVal {
				return false
			}
		}
	}

	return true
}

// TestBuildContext_DepthCheck 测试递归深度检查
func TestBuildContext_DepthCheck(t *testing.T) {
	tests := []struct {
		name        string
		depth       int
		shouldError bool
	}{
		{
			name:        "depth 0 - should pass",
			depth:       0,
			shouldError: false,
		},
		{
			name:        "depth 5 - should pass",
			depth:       5,
			shouldError: false,
		},
		{
			name:        "depth at max - should pass",
			depth:       maxRecursionDepth - 1,
			shouldError: false,
		},
		{
			name:        "depth exceeds max - should error",
			depth:       maxRecursionDepth,
			shouldError: true,
		},
		{
			name:        "depth far exceeds max - should error",
			depth:       maxRecursionDepth + 10,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &buildContext{
				depth: tt.depth,
			}

			err := ctx.checkDepth()

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for depth %d, but got nil", tt.depth)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for depth %d, but got: %v", tt.depth, err)
			}
		})
	}
}

// TestBuildContext_Next 测试上下文的 next 方法
func TestBuildContext_Next(t *testing.T) {
	tests := []struct {
		name          string
		initialDepth  int
		expectedDepth int
	}{
		{
			name:          "depth 0 to 1",
			initialDepth:  0,
			expectedDepth: 1,
		},
		{
			name:          "depth 5 to 6",
			initialDepth:  5,
			expectedDepth: 6,
		},
		{
			name:          "depth 10 to 11",
			initialDepth:  10,
			expectedDepth: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &buildContext{
				depth: tt.initialDepth,
			}

			newCtx := ctx.next()

			if newCtx.depth != tt.expectedDepth {
				t.Errorf("Expected depth %d, got %d", tt.expectedDepth, newCtx.depth)
			}

			// 确保原上下文未被修改
			if ctx.depth != tt.initialDepth {
				t.Errorf("Original context depth was modified from %d to %d", tt.initialDepth, ctx.depth)
			}
		})
	}
}

// TestNewBuildContext 测试新建上下文
func TestNewBuildContext(t *testing.T) {
	ctx := newBuildContext()

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}

	if ctx.depth != 0 {
		t.Errorf("Expected initial depth 0, got %d", ctx.depth)
	}
}

// TestSchemaBuilder_RecursionDepthLimit 测试递归深度限制
func TestSchemaBuilder_RecursionDepthLimit(t *testing.T) {
	// 创建一个深度嵌套的类型结构
	ref1 := int64(2)
	ref2 := int64(1)
	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "Type1",
				Description: "First type",
				AppID:       100,
			},
			2: {
				ID:          2,
				Name:        "Type2",
				Description: "Second type",
				AppID:       100,
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:       1,
					Name:     "nested",
					Type:     "custom",
					Ref:      &ref1,
					Required: false,
					AppID:    100,
				},
			},
			2: {
				{
					ID:       2,
					Name:     "nested",
					Type:     "custom",
					Ref:      &ref2, // 循环引用
					Required: false,
					AppID:    100,
				},
			},
		},
	}

	// 创建一个接近最大深度的上下文
	ctx := &buildContext{
		depth: maxRecursionDepth - 1,
	}

	// 尝试构建 schema，应该在下一层递归时失败
	_, err := builder.buildSchemaByType(1, ctx)

	if err == nil {
		t.Error("Expected error when exceeding max recursion depth, but got nil")
		return
	}

	if err.Error() != "maximum recursion depth exceeded" {
		t.Errorf("Expected 'maximum recursion depth exceeded' error, got: %v", err)
	}
}

// TestSchemaBuilder_AppIdFiltering 测试 AppId 过滤功能
func TestSchemaBuilder_AppIdFiltering(t *testing.T) {
	// 注意：这个测试需要实际的数据库连接
	// 这里我们测试 schemaBuilder 的内存数据结构是否正确过滤

	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "App1Type",
				Description: "Type for app 1",
				AppID:       1,
			},
			2: {
				ID:          2,
				Name:        "App2Type",
				Description: "Type for app 2",
				AppID:       2,
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:       1,
					Name:     "field1",
					Type:     "string",
					Required: true,
					AppID:    1,
				},
			},
			2: {
				{
					ID:       2,
					Name:     "field2",
					Type:     "number",
					Required: false,
					AppID:    2,
				},
			},
		},
	}

	tests := []struct {
		name           string
		typeId         int64
		expectedAppId  int64
		shouldHaveType bool
	}{
		{
			name:           "get type from app 1",
			typeId:         1,
			expectedAppId:  1,
			shouldHaveType: true,
		},
		{
			name:           "get type from app 2",
			typeId:         2,
			expectedAppId:  2,
			shouldHaveType: true,
		},
		{
			name:           "get non-existent type",
			typeId:         999,
			expectedAppId:  0,
			shouldHaveType: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customType, err := builder.getCustomType(tt.typeId)

			if tt.shouldHaveType {
				if err != nil {
					t.Errorf("Expected to find type, but got error: %v", err)
				}
				if customType.AppID != tt.expectedAppId {
					t.Errorf("Expected AppID %d, got %d", tt.expectedAppId, customType.AppID)
				}
			} else {
				if err == nil {
					t.Error("Expected error for non-existent type, but got nil")
				}
			}
		})
	}
}

// TestSchemaBuilder_CircularReferenceWithDepthLimit 测试循环引用在深度限制下的行为
func TestSchemaBuilder_CircularReferenceWithDepthLimit(t *testing.T) {
	// 创建一个简单的循环引用：A -> B -> A
	refToB := int64(2)
	refToA := int64(1)
	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {
				ID:          1,
				Name:        "TypeA",
				Description: "Type A",
				AppID:       100,
			},
			2: {
				ID:          2,
				Name:        "TypeB",
				Description: "Type B",
				AppID:       100,
			},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{
					ID:       1,
					Name:     "toB",
					Type:     "custom",
					Ref:      &refToB,
					Required: false,
					AppID:    100,
				},
			},
			2: {
				{
					ID:       2,
					Name:     "toA",
					Type:     "custom",
					Ref:      &refToA,
					Required: false,
					AppID:    100,
				},
			},
		},
	}

	ctx := newBuildContext()

	// 循环引用会导致递归深度不断增加，最终达到深度限制
	schema, err := builder.buildSchemaByType(1, ctx)

	// 由于循环引用，最终会因为深度限制而返回错误
	if err == nil {
		t.Error("Expected error due to circular reference exceeding depth limit, but got nil")
	}

	if err != nil && err.Error() != "maximum recursion depth exceeded" {
		t.Errorf("Expected 'maximum recursion depth exceeded' error, got: %v", err)
	}

	// schema 应该为 nil，因为构建失败
	if schema != nil {
		t.Error("Expected nil schema due to error")
	}
}

// TestSchemaBuilder_DeepNesting 测试深度嵌套但无循环的情况
func TestSchemaBuilder_DeepNesting(t *testing.T) {
	// 创建一个深度嵌套的结构：A -> B -> C -> D
	refToB := int64(2)
	refToC := int64(3)
	refToD := int64(4)
	builder := &schemaBuilder{
		types: map[int64]*models.CustomType{
			1: {ID: 1, Name: "TypeA", Description: "Type A", AppID: 100},
			2: {ID: 2, Name: "TypeB", Description: "Type B", AppID: 100},
			3: {ID: 3, Name: "TypeC", Description: "Type C", AppID: 100},
			4: {ID: 4, Name: "TypeD", Description: "Type D", AppID: 100},
		},
		fields: map[int64][]models.CustomTypeField{
			1: {
				{ID: 1, Name: "toB", Type: "custom", Ref: &refToB, AppID: 100},
			},
			2: {
				{ID: 2, Name: "toC", Type: "custom", Ref: &refToC, AppID: 100},
			},
			3: {
				{ID: 3, Name: "toD", Type: "custom", Ref: &refToD, AppID: 100},
			},
			4: {
				{ID: 4, Name: "value", Type: "string", AppID: 100},
			},
		},
	}

	ctx := newBuildContext()

	schema, err := builder.buildSchemaByType(1, ctx)

	if err != nil {
		t.Errorf("Unexpected error for deep nesting: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	// 验证嵌套结构
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// 检查第一层
	toBProp, exists := properties["toB"]
	if !exists {
		t.Fatal("Expected 'toB' property")
	}

	toBMap, ok := toBProp.(map[string]any)
	if !ok {
		t.Fatal("Expected 'toB' to be a map")
	}

	// 检查第二层
	toBProperties, ok := toBMap["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected nested properties")
	}

	if _, exists := toBProperties["toC"]; !exists {
		t.Error("Expected 'toC' property in nested structure")
	}
}
