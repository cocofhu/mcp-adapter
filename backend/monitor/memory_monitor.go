package monitor

import (
	"log"
	"runtime"
	"time"
)

// MemoryStats 内存统计信息
type MemoryStats struct {
	Timestamp      time.Time
	AllocMB        float64 // 当前分配的内存（MB）
	TotalAllocMB   float64 // 累计分配的内存（MB）
	SysMB          float64 // 从系统获取的内存（MB）
	NumGC          uint32  // GC 次数
	NumGoroutine   int     // Goroutine 数量
	HeapObjects    uint64  // 堆对象数量
	HeapAllocMB    float64 // 堆分配内存（MB）
	HeapIdleMB     float64 // 堆空闲内存（MB）
	HeapReleasedMB float64 // 已释放给系统的堆内存（MB）
}

// GetMemoryStats 获取当前内存统计
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Timestamp:      time.Now(),
		AllocMB:        float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB:   float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:          float64(m.Sys) / 1024 / 1024,
		NumGC:          m.NumGC,
		NumGoroutine:   runtime.NumGoroutine(),
		HeapObjects:    m.HeapObjects,
		HeapAllocMB:    float64(m.HeapAlloc) / 1024 / 1024,
		HeapIdleMB:     float64(m.HeapIdle) / 1024 / 1024,
		HeapReleasedMB: float64(m.HeapReleased) / 1024 / 1024,
	}
}

// Log 记录内存统计信息
func (ms MemoryStats) Log() {
	log.Printf("=== Memory Stats ===")
	log.Printf("Time:           %s", ms.Timestamp.Format("2006-01-02 15:04:05"))
	log.Printf("Alloc:          %.2f MB", ms.AllocMB)
	log.Printf("TotalAlloc:     %.2f MB", ms.TotalAllocMB)
	log.Printf("Sys:            %.2f MB", ms.SysMB)
	log.Printf("HeapAlloc:      %.2f MB", ms.HeapAllocMB)
	log.Printf("HeapIdle:       %.2f MB", ms.HeapIdleMB)
	log.Printf("HeapReleased:   %.2f MB", ms.HeapReleasedMB)
	log.Printf("HeapObjects:    %d", ms.HeapObjects)
	log.Printf("NumGC:          %d", ms.NumGC)
	log.Printf("NumGoroutine:   %d", ms.NumGoroutine)
	log.Printf("===================")
}

// StartMemoryMonitor 启动内存监控（定期输出内存统计）
func StartMemoryMonitor(interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Memory monitor started (interval: %v)", interval)

		for {
			select {
			case <-ticker.C:
				stats := GetMemoryStats()
				stats.Log()
			case <-stop:
				log.Println("Memory monitor stopped")
				return
			}
		}
	}()

	return stop
}

// MemoryComparison 内存对比
type MemoryComparison struct {
	Before MemoryStats
	After  MemoryStats
}

// Log 对比两次内存统计
func (mc MemoryComparison) Log() {
	allocDiff := mc.After.AllocMB - mc.Before.AllocMB
	sysDiff := mc.After.SysMB - mc.Before.SysMB
	goroutineDiff := mc.After.NumGoroutine - mc.Before.NumGoroutine
	gcDiff := mc.After.NumGC - mc.Before.NumGC
	objectsDiff := int64(mc.After.HeapObjects) - int64(mc.Before.HeapObjects)

	log.Printf("=== Memory Comparison ===")
	log.Printf("Alloc:        %.2f MB (%.2f MB)", mc.After.AllocMB, allocDiff)
	log.Printf("Sys:          %.2f MB (%.2f MB)", mc.After.SysMB, sysDiff)
	log.Printf("Goroutines:   %d (%+d)", mc.After.NumGoroutine, goroutineDiff)
	log.Printf("HeapObjects:  %d (%+d)", mc.After.HeapObjects, objectsDiff)
	log.Printf("GC Runs:      %d (%+d)", mc.After.NumGC, gcDiff)

	// 警告检测
	if allocDiff > 100 {
		log.Printf("⚠️  WARNING: Memory increased by %.2f MB", allocDiff)
	}
	if goroutineDiff > 10 {
		log.Printf("⚠️  WARNING: Goroutines increased by %d", goroutineDiff)
	}
	if objectsDiff > 100000 {
		log.Printf("⚠️  WARNING: Heap objects increased by %d", objectsDiff)
	}

	log.Printf("========================")
}

// ForceGC 强制执行垃圾回收
func ForceGC() {
	log.Println("Forcing garbage collection...")
	runtime.GC()
	time.Sleep(5000 * time.Millisecond) // 等待 GC 完成
	log.Println("Garbage collection completed")
}

// PrintGoroutineStack 打印所有 Goroutine 的堆栈信息
func PrintGoroutineStack() {
	buf := make([]byte, 1<<20) // 1MB buffer
	stackSize := runtime.Stack(buf, true)
	log.Printf("=== Goroutine Stacks (%d bytes) ===\n%s\n", stackSize, buf[:stackSize])
}
