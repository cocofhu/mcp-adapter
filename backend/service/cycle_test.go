package service

import (
	"testing"
)

// 测试拓扑排序判环的基本逻辑
func TestTopologicalSortCycleDetection(t *testing.T) {
	tests := []struct {
		name      string
		graph     map[int64][]int64
		hasCycle  bool
	}{
		{
			name: "无环图",
			graph: map[int64][]int64{
				1: {2, 3},
				2: {4},
				3: {4},
				4: {},
			},
			hasCycle: false,
		},
		{
			name: "简单环 A->B->A",
			graph: map[int64][]int64{
				1: {2},
				2: {1},
			},
			hasCycle: true,
		},
		{
			name: "自环 A->A",
			graph: map[int64][]int64{
				1: {1},
			},
			hasCycle: true,
		},
		{
			name: "复杂环 A->B->C->A",
			graph: map[int64][]int64{
				1: {2},
				2: {3},
				3: {1},
			},
			hasCycle: true,
		},
		{
			name: "部分环",
			graph: map[int64][]int64{
				1: {2},
				2: {3},
				3: {4},
				4: {2}, // 2->3->4->2 形成环
				5: {6},
				6: {},
			},
			hasCycle: true,
		},
		{
			name: "空图",
			graph: map[int64][]int64{},
			hasCycle: false,
		},
		{
			name: "单节点无边",
			graph: map[int64][]int64{
				1: {},
			},
			hasCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 计算入度
			inDegree := make(map[int64]int)
			for node := range tt.graph {
				if _, exists := inDegree[node]; !exists {
					inDegree[node] = 0
				}
			}
			for _, neighbors := range tt.graph {
				for _, neighbor := range neighbors {
					inDegree[neighbor]++
				}
			}

			// Kahn 算法
			queue := []int64{}
			for node := range tt.graph {
				if inDegree[node] == 0 {
					queue = append(queue, node)
				}
			}

			processedCount := 0
			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				processedCount++

				for _, neighbor := range tt.graph[current] {
					inDegree[neighbor]--
					if inDegree[neighbor] == 0 {
						queue = append(queue, neighbor)
					}
				}
			}

			hasCycle := processedCount < len(tt.graph)
			if hasCycle != tt.hasCycle {
				t.Errorf("期望 hasCycle=%v, 实际 hasCycle=%v (处理了 %d/%d 个节点)",
					tt.hasCycle, hasCycle, processedCount, len(tt.graph))
			}
		})
	}
}

// 性能基准测试
func BenchmarkTopologicalSort(b *testing.B) {
	// 构建一个100个节点的图,每个节点平均引用3个其他节点
	graph := make(map[int64][]int64)
	for i := int64(1); i <= 100; i++ {
		graph[i] = []int64{}
		for j := 0; j < 3; j++ {
			target := (i + int64(j) + 1) % 100
			if target == 0 {
				target = 100
			}
			if target != i {
				graph[i] = append(graph[i], target)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 计算入度
		inDegree := make(map[int64]int)
		for node := range graph {
			inDegree[node] = 0
		}
		for _, neighbors := range graph {
			for _, neighbor := range neighbors {
				inDegree[neighbor]++
			}
		}

		// Kahn 算法
		queue := []int64{}
		for node := range graph {
			if inDegree[node] == 0 {
				queue = append(queue, node)
			}
		}

		processedCount := 0
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			processedCount++

			for _, neighbor := range graph[current] {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					queue = append(queue, neighbor)
				}
			}
		}
	}
}
