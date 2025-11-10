package adapter

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestGoroutineLeakPrevention 测试 Goroutine 泄漏预防
func TestGoroutineLeakPrevention(t *testing.T) {
	// 记录初始 goroutine 数量
	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// 初始化服务器
	InitServer()
	time.Sleep(100 * time.Millisecond) // 等待 goroutine 启动

	afterInitGoroutines := runtime.NumGoroutine()
	t.Logf("After init goroutines: %d", afterInitGoroutines)

	// 应该增加 1 个 goroutine（事件循环）
	if afterInitGoroutines <= initialGoroutines {
		t.Errorf("Expected goroutine count to increase after init")
	}

	// 关闭服务器
	Shutdown()
	time.Sleep(200 * time.Millisecond) // 等待清理完成

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d", finalGoroutines)

	// goroutine 应该恢复到初始水平（允许少量误差）
	if finalGoroutines > initialGoroutines+2 {
		t.Errorf("Goroutine leak detected: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	}
}

// TestMultipleInitializationProtection 测试多次初始化保护
func TestMultipleInitializationProtection(t *testing.T) {
	// 重置 initOnce（仅用于测试）
	initOnce = sync.Once{}
	serverManager = nil

	initialGoroutines := runtime.NumGoroutine()

	// 多次调用 InitServer
	for i := 0; i < 5; i++ {
		InitServer()
	}

	time.Sleep(100 * time.Millisecond)
	afterGoroutines := runtime.NumGoroutine()

	// 即使调用多次，也只应该启动一个事件循环
	expectedIncrease := 1
	actualIncrease := afterGoroutines - initialGoroutines

	if actualIncrease > expectedIncrease+1 {
		t.Errorf("Multiple initialization created extra goroutines: expected ~%d, got %d", expectedIncrease, actualIncrease)
	}

	Shutdown()
}

// TestEventChannelNonBlocking 测试事件通道不阻塞
func TestEventChannelNonBlocking(t *testing.T) {
	initOnce = sync.Once{}
	serverManager = nil

	InitServer()
	defer Shutdown()

	// 发送大量事件
	done := make(chan bool)
	go func() {
		for i := 0; i < 200; i++ {
			SendEvent(Event{
				Code: AddToolEvent,
			})
		}
		done <- true
	}()

	// 确保在合理时间内完成（不应该阻塞）
	select {
	case <-done:
		t.Log("Event sending completed without blocking")
	case <-time.After(2 * time.Second):
		t.Error("Event sending blocked - possible deadlock")
	}
}

// TestServerCleanup 测试服务器清理机制
func TestServerCleanup(t *testing.T) {
	srv := &Server{
		path:       "test",
		cleanupFns: make([]func(), 0),
	}

	cleanupCalled := 0
	var mu sync.Mutex

	// 添加多个清理函数
	for i := 0; i < 5; i++ {
		srv.AddCleanup(func() {
			mu.Lock()
			cleanupCalled++
			mu.Unlock()
		})
	}

	// 执行清理
	srv.Cleanup()

	mu.Lock()
	count := cleanupCalled
	mu.Unlock()

	if count != 5 {
		t.Errorf("Expected 5 cleanup calls, got %d", count)
	}

	// 再次清理应该不会出错
	srv.Cleanup()
}

// TestShutdownWithPendingEvents 测试有待处理事件时的关闭
func TestShutdownWithPendingEvents(t *testing.T) {
	initOnce = sync.Once{}
	serverManager = nil

	InitServer()

	// 发送事件但不等待处理完成
	for i := 0; i < 50; i++ {
		SendEvent(Event{
			Code: ToolListChanged,
		})
	}

	// 立即关闭
	done := make(chan bool)
	go func() {
		Shutdown()
		done <- true
	}()

	// 关闭应该在合理时间内完成
	select {
	case <-done:
		t.Log("Shutdown completed gracefully with pending events")
	case <-time.After(3 * time.Second):
		t.Error("Shutdown hung with pending events")
	}
}

// BenchmarkEventProcessing 性能基准测试
func BenchmarkEventProcessing(b *testing.B) {
	initOnce = sync.Once{}
	serverManager = nil

	InitServer()
	defer Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SendEvent(Event{
			Code: ToolListChanged,
		})
	}
	b.StopTimer()

	// 等待所有事件处理完成
	time.Sleep(100 * time.Millisecond)
}
