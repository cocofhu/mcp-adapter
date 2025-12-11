package adapter

import (
	"testing"
)

func BenchmarkSatisfySchema_DeeplyNestedWithRequired(b *testing.B) {
	schema := buildDeeplyNestedSchema(10) // 10层嵌套
	data := buildDeeplyNestedData(10)     // 构造对应的数据

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_WideNestedWithRequired 测试宽度嵌套schema的性能
func BenchmarkSatisfySchema_WideNestedWithRequired(b *testing.B) {
	schema := buildWideNestedSchema(5, 5) // 5层嵌套，每层5个子对象
	data := buildWideNestedData(5, 5)     // 构造对应的数据

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// buildDeeplyNestedSchema 构建深度嵌套的schema
// depth: 嵌套深度
func buildDeeplyNestedSchema(depth int) map[string]any {
	if depth == 0 {
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"leaf": map[string]any{
					"type": "string",
				},
			},
			"required": []any{"leaf"},
		}
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"nested": buildDeeplyNestedSchema(depth - 1),
		},
		"required": []any{},
	}
}

func buildDeeplyNestedSchemaWithNoRequired(depth int) map[string]any {
	if depth == 0 {
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"leaf": map[string]any{
					"type": "string",
				},
			},
			"required": []any{},
		}
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"nested": buildDeeplyNestedSchemaWithNoRequired(depth - 1),
		},
		"required": []any{},
	}
}

// buildWideNestedSchema 构建宽度嵌套的schema
// depth: 嵌套深度
// width: 每层的子对象数量
func buildWideNestedSchema(depth, width int) map[string]any {
	if depth == 0 {
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"leaf": map[string]any{
					"type": "string",
				},
			},
			"required": []any{"leaf"},
		}
	}

	properties := make(map[string]any)
	for i := 0; i < width; i++ {
		properties[string(rune('a'+i))] = buildWideNestedSchema(depth-1, width)
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   []any{},
	}
}

// buildDeeplyNestedData 构建深度嵌套的测试数据
// depth: 嵌套深度，需要与schema的深度匹配
func buildDeeplyNestedData(depth int) map[string]any {
	if depth == 0 {
		return map[string]any{
			"leaf": "test-value",
		}
	}

	return map[string]any{
		"nested": buildDeeplyNestedData(depth - 1),
	}
}

// buildWideNestedData 构建宽度嵌套的测试数据
// depth: 嵌套深度
// width: 每层的子对象数量
func buildWideNestedData(depth, width int) map[string]any {
	if depth == 0 {
		return map[string]any{
			"leaf": "test-value",
		}
	}

	data := make(map[string]any)
	for i := 0; i < width; i++ {
		data[string(rune('a'+i))] = buildWideNestedData(depth-1, width)
	}

	return data
}

// BenchmarkSatisfySchema_ComplexValidData 测试复杂schema验证有效数据的性能
func BenchmarkSatisfySchema_ComplexValidData(b *testing.B) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"profile": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{"type": "string"},
							"age":  map[string]any{"type": "number"},
						},
						"required": []any{"name"},
					},
					"settings": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"theme": map[string]any{"type": "string"},
						},
						"required": []any{},
					},
				},
				"required": []any{"profile"},
			},
		},
		"required": []any{"user"},
	}

	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"name": "John",
				"age":  30,
			},
			"settings": map[string]any{
				"theme": "dark",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_SharedSchemaReferences 测试共享schema引用的性能
// 这个测试场景中，多个字段共享同一个schema定义，会触发缓存优化
func BenchmarkSatisfySchema_SharedSchemaReferences(b *testing.B) {
	// 创建一个共享的子schema（同一个map实例）
	sharedUserSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "number"},
			"address": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"street": map[string]any{"type": "string"},
					"city":   map[string]any{"type": "string"},
				},
				"required": []any{"city"},
			},
		},
		"required": []any{"name"},
	}

	// 多个字段引用同一个schema实例
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"user1":  sharedUserSchema,
			"user2":  sharedUserSchema,
			"user3":  sharedUserSchema,
			"user4":  sharedUserSchema,
			"user5":  sharedUserSchema,
			"user6":  sharedUserSchema,
			"user7":  sharedUserSchema,
			"user8":  sharedUserSchema,
			"user9":  sharedUserSchema,
			"user10": sharedUserSchema,
		},
		"required": []any{},
	}

	// 构造对应的数据（部分字段为nil，触发hasAnyRequiredFields检查）
	data := map[string]any{
		"user1": map[string]any{
			"name": "Alice",
			"age":  25,
		},
		"user2": nil, // nil值会触发hasAnyRequiredFields检查
		"user3": map[string]any{
			"name": "Bob",
		},
		"user4": nil, // nil值会触发hasAnyRequiredFields检查
		"user5": map[string]any{
			"name": "Charlie",
			"age":  35,
		},
		// user6-user10 不提供数据，也会触发检查
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayWithMixedNilElements 测试数组中混合nil元素的性能
// 这个测试验证数组中第一个元素为非nil，后续为nil的场景
// 场景: [{}, nil, nil, nil] - 第一个元素是空对象，后续都是nil
func BenchmarkSatisfySchema_ArrayWithMixedNilElements(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 第一个元素是空对象，后续都是nil
	data := []any{
		map[string]any{},
		nil,
		nil,
		nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayWithManyNilElements 测试数组中大量nil元素的性能
// 这个测试验证大量nil元素时的性能表现
func BenchmarkSatisfySchema_ArrayWithManyNilElements(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 构造一个包含100个元素的数组，第一个是空对象，其余都是nil
	data := make([]any, 100)
	data[0] = map[string]any{}
	for i := 1; i < 100; i++ {
		data[i] = nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayAllNil 测试数组全部为nil的性能
func BenchmarkSatisfySchema_ArrayAllNil(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 全部都是nil
	data := []any{nil, nil, nil, nil}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayAllValid 测试数组全部为有效对象的性能
func BenchmarkSatisfySchema_ArrayAllValid(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 全部都是有效对象
	data := []any{
		map[string]any{"name": "Alice", "age": 25},
		map[string]any{"name": "Bob", "age": 30},
		map[string]any{"name": "Charlie", "age": 35},
		map[string]any{"name": "David", "age": 40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayNilFirst 测试数组第一个元素为nil的性能
func BenchmarkSatisfySchema_ArrayNilFirst(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 第一个元素是nil，后续是有效对象
	data := []any{
		nil,
		map[string]any{"name": "Bob", "age": 30},
		map[string]any{"name": "Charlie", "age": 35},
		map[string]any{"name": "David", "age": 40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayMixedNilPattern 测试数组中混合非nil和多个nil元素的性能
// 场景: [{}, nil, nil, {}] - 验证checked优化逻辑
// 这个测试验证：第一个nil会被检查，后续的nil会被跳过，最后的非nil对象会被检查
func BenchmarkSatisfySchema_ArrayMixedNilPattern(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 第一个是空对象，第二、三个是nil，第四个是空对象
	data := []any{
		map[string]any{},
		nil,
		nil,
		map[string]any{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayMultipleNilGroups 测试数组中多组nil元素的性能
// 场景: [{}, nil, nil, {}, nil, nil, nil] - 验证多组nil的优化效果
func BenchmarkSatisfySchema_ArrayMultipleNilGroups(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 多组nil元素，中间穿插非nil对象
	data := []any{
		map[string]any{},
		nil,
		nil,
		map[string]any{"name": "Alice"},
		nil,
		nil,
		nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayLargeWithSparseNil 测试大数组中稀疏nil元素的性能
// 场景: 大数组中间隔分布nil元素，验证checked优化在大规模数据下的效果
func BenchmarkSatisfySchema_ArrayLargeWithSparseNil(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
			"required": []any{"name"},
		},
	}

	// 构造一个包含1000个元素的数组，每10个元素中有3个nil
	data := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		if i%10 < 3 {
			data[i] = nil
		} else {
			data[i] = map[string]any{"name": "User" + string(rune('A'+i%26))}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

// BenchmarkSatisfySchema_ArrayConsecutiveNilsVsScattered 对比连续nil和分散nil的性能
// 这个benchmark对比两种场景：
// 1. 连续的nil元素（checked优化效果明显）
// 2. 分散的nil元素（checked优化效果有限）
func BenchmarkSatisfySchema_ArrayConsecutiveNils(b *testing.B) {
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
			"required": []any{"name"},
		},
	}

	// 前面是有效对象，后面是大量连续的nil
	data := make([]any, 1000)
	data[0] = map[string]any{"name": "Alice"}
	data[1] = map[string]any{"name": "Bob"}
	for i := 2; i < 1000; i++ {
		data[i] = nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}

func BenchmarkSatisfySchema_ArrayScatteredNils(b *testing.B) {
	schema := map[string]any{
		"type":  "array",
		"items": buildDeeplyNestedSchemaWithNoRequired(100),
	}

	// nil和非nil元素交替出现
	data := make([]any, 10000)
	for i := 0; i < 10000; i++ {
		if i%100 == 0 {
			data[i] = map[string]any{"leaf": "User"}
		} else {
			data[i] = nil
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SatisfySchema(schema, data)
	}
}
