package adapter

import (
	"encoding/json"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		length      int
		inputData   string
		expected    string
		expectError bool
	}{
		{
			name:   "truncate simple string field",
			key:    "message",
			length: 5,
			inputData: `{
				"message": "hello world",
				"status": "success"
			}`,
			expected: `{
				"message": "hello",
				"status": "success"
			}`,
			expectError: false,
		},
		{
			name:   "truncate nested field",
			key:    "user.name",
			length: 3,
			inputData: `{
				"user": {
					"name": "John Doe",
					"age": 30
				}
			}`,
			expected: `{
				"user": {
					"name": "Joh",
					"age": 30
				}
			}`,
			expectError: false,
		},
		{
			name:   "truncate with wildcard in object",
			key:    "users.*.name",
			length: 4,
			inputData: `{
				"users": {
					"user1": {
						"name": "Alice Smith",
						"age": 25
					},
					"user2": {
						"name": "Bob Johnson",
						"age": 30
					}
				}
			}`,
			expected: `{
				"users": {
					"user1": {
						"name": "Alic",
						"age": 25
					},
					"user2": {
						"name": "Bob ",
						"age": 30
					}
				}
			}`,
			expectError: false,
		},
		{
			name:   "truncate array element by index",
			key:    "items.0.title",
			length: 6,
			inputData: `{
				"items": [
					{"title": "First item description", "id": 1},
					{"title": "Second item description", "id": 2}
				]
			}`,
			expected: `{
				"items": [
					{"title": "First ", "id": 1},
					{"title": "Second item description", "id": 2}
				]
			}`,
			expectError: false,
		},
		{
			name:   "truncate all array elements with wildcard",
			key:    "items.*.title",
			length: 7,
			inputData: `{
				"items": [
					{"title": "First item description", "id": 1},
					{"title": "Second item description", "id": 2},
					{"title": "Third item description", "id": 3}
				]
			}`,
			expected: `{
				"items": [
					{"title": "First i", "id": 1},
					{"title": "Second ", "id": 2},
					{"title": "Third i", "id": 3}
				]
			}`,
			expectError: false,
		},
		{
			name:   "string shorter than length - no truncation",
			key:    "message",
			length: 20,
			inputData: `{
				"message": "short"
			}`,
			expected: `{
				"message": "short"
			}`,
			expectError: false,
		},
		{
			name:   "deeply nested path",
			key:    "level1.level2.level3.text",
			length: 3,
			inputData: `{
				"level1": {
					"level2": {
						"level3": {
							"text": "deep nested text"
						}
					}
				}
			}`,
			expected: `{
				"level1": {
					"level2": {
						"level3": {
							"text": "dee"
						}
					}
				}
			}`,
			expectError: false,
		},
		{
			name:   "non-existent path - data unchanged",
			key:    "nonexistent.field",
			length: 5,
			inputData: `{
				"message": "hello world"
			}`,
			expected: `{
				"message": "hello world"
			}`,
			expectError: false,
		},
		{
			name:   "truncate non-string value - no effect",
			key:    "count",
			length: 2,
			inputData: `{
				"count": 12345,
				"message": "test"
			}`,
			expected: `{
				"count": 12345,
				"message": "test"
			}`,
			expectError: false,
		},
		{
			name:   "array wildcard at top level",
			key:    "*",
			length: 5,
			inputData: `["hello world", "foo bar", "test"]`,
			expected:    `["hello", "foo b", "test"]`,
			expectError: false,
		},
		{
			name:        "invalid JSON input",
			key:         "message",
			length:      5,
			inputData:   `{invalid json}`,
			expected:    `{invalid json}`,
			expectError: true,
		},
		{
			name:   "empty string field",
			key:    "message",
			length: 5,
			inputData: `{
				"message": ""
			}`,
			expected: `{
				"message": ""
			}`,
			expectError: false,
		},
		{
			name:   "truncate complex nested structure",
			key:    "response.data.*.description",
			length: 10,
			inputData: `{
				"response": {
					"data": {
						"item1": {
							"description": "This is a very long description that needs truncation",
							"id": 1
						},
						"item2": {
							"description": "Another lengthy description here",
							"id": 2
						}
					},
					"status": "ok"
				}
			}`,
			expected: `{
				"response": {
					"data": {
						"item1": {
							"description": "This is a ",
							"id": 1
						},
						"item2": {
							"description": "Another le",
							"id": 2
						}
					},
					"status": "ok"
				}
			}`,
			expectError: false,
		},
		{
			name:   "zero length truncation",
			key:    "message",
			length: 0,
			inputData: `{
				"message": "hello"
			}`,
			expected: `{
				"message": ""
			}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := truncate(tt.key, tt.length, []byte(tt.inputData))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// 比较 JSON 结构而不是字符串，以忽略格式差异
			var resultJSON, expectedJSON interface{}
			if err := json.Unmarshal(result, &resultJSON); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}
			if err := json.Unmarshal([]byte(tt.expected), &expectedJSON); err != nil {
				t.Errorf("Failed to unmarshal expected: %v", err)
				return
			}

			resultStr, _ := json.Marshal(resultJSON)
			expectedStr, _ := json.Marshal(expectedJSON)

			if string(resultStr) != string(expectedStr) {
				t.Errorf("Result mismatch:\nGot:      %s\nExpected: %s", string(resultStr), string(expectedStr))
			}
		})
	}
}

func TestTruncateByPath(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		path     []string
		length   int
		expected interface{}
	}{
		{
			name: "truncate map field",
			data: map[string]any{
				"message": "hello world",
				"status":  "success",
			},
			path:   []string{"message"},
			length: 5,
			expected: map[string]any{
				"message": "hello",
				"status":  "success",
			},
		},
		{
			name: "truncate nested map field",
			data: map[string]any{
				"user": map[string]any{
					"name": "John Doe",
					"age":  30,
				},
			},
			path:   []string{"user", "name"},
			length: 4,
			expected: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
				},
			},
		},
		{
			name: "wildcard in map",
			data: map[string]any{
				"user1": "Alice Smith",
				"user2": "Bob Johnson",
			},
			path:   []string{"*"},
			length: 5,
			expected: map[string]any{
				"user1": "Alice",
				"user2": "Bob J",
			},
		},
		{
			name: "truncate array element by index",
			data: []any{
				"first string",
				"second string",
				"third string",
			},
			path:   []string{"1"},
			length: 6,
			expected: []any{
				"first string",
				"second",
				"third string",
			},
		},
		{
			name: "wildcard in array",
			data: []any{
				"first string",
				"second string",
				"third string",
			},
			path:   []string{"*"},
			length: 5,
			expected: []any{
				"first",
				"secon",
				"third",
			},
		},
		{
			name: "complex nested structure with wildcard",
			data: map[string]any{
				"users": []any{
					map[string]any{"name": "Alice Johnson", "age": 25},
					map[string]any{"name": "Bob Smith", "age": 30},
				},
			},
			path:   []string{"users", "*", "name"},
			length: 5,
			expected: map[string]any{
				"users": []any{
					map[string]any{"name": "Alice", "age": 25},
					map[string]any{"name": "Bob S", "age": 30},
				},
			},
		},
		{
			name:     "empty path returns original data",
			data:     map[string]any{"message": "hello"},
			path:     []string{},
			length:   5,
			expected: map[string]any{"message": "hello"},
		},
		{
			name:     "non-existent key returns original data",
			data:     map[string]any{"message": "hello"},
			path:     []string{"nonexistent"},
			length:   5,
			expected: map[string]any{"message": "hello"},
		},
		{
			name:     "invalid array index returns original array",
			data:     []any{"first", "second"},
			path:     []string{"10"},
			length:   3,
			expected: []any{"first", "second"},
		},
		{
			name:     "non-string value is not truncated",
			data:     map[string]any{"count": 12345},
			path:     []string{"count"},
			length:   2,
			expected: map[string]any{"count": 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateByPath(tt.data, tt.path, tt.length)

			resultJSON, _ := json.Marshal(result)
			expectedJSON, _ := json.Marshal(tt.expected)

			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("Result mismatch:\nGot:      %s\nExpected: %s", string(resultJSON), string(expectedJSON))
			}
		})
	}
}

func BenchmarkTruncate(b *testing.B) {
	data := []byte(`{
		"response": {
			"data": {
				"items": [
					{"description": "This is a very long description that needs truncation", "id": 1},
					{"description": "Another lengthy description here", "id": 2},
					{"description": "Yet another description to truncate", "id": 3}
				]
			},
			"status": "ok"
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = truncate("response.data.items.*.description", 10, data)
	}
}
