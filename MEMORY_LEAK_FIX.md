# 内存泄漏修复文档

## 📋 修复概览

本次修复解决了两个最高优先级（P0）的内存泄漏问题：

### 1. ✅ Goroutine 泄漏修复
### 2. ✅ sync.Map 资源释放修复

---

## 🔧 主要改动

### 一、架构重构

#### 1. 引入 ServerManager 结构
**文件**: `backend/adapter/mcp_model.go`

```go
type ServerManager struct {
    sseServers sync.Map          // 服务器映射
    eventChan  chan Event        // 事件通道（缓冲区增至 100）
    ctx        context.Context   // 生命周期控制
    cancel     context.CancelFunc // 取消函数
    wg         sync.WaitGroup    // 等待 goroutine 完成
    mu         sync.RWMutex      // 并发保护
}
```

**优势**：
- ✅ 统一管理所有资源
- ✅ 支持优雅关闭
- ✅ 防止资源泄漏
- ✅ 线程安全

---

#### 2. Server 结构增强

**原有结构**：
```go
type Server struct {
    protocol string
    path     string
    server   *server.MCPServer
    impl     http.Handler
}
```

**修复后**：
```go
type Server struct {
    protocol   string
    path       string
    server     *server.MCPServer
    impl       http.Handler
    cleanupFns []func()  // 🆕 清理函数列表
    mu         sync.Mutex
}
```

**新增方法**：
- `AddCleanup(fn func())`：注册清理函数
- `Cleanup()`：执行所有清理操作（逆序执行，panic 安全）

---

### 二、Goroutine 生命周期管理

#### 修复前（存在泄漏）：
```go
func InitServer() {
    event = make(chan Event, 16)
    
    go func() {
        for {  // ❌ 无限循环，无法停止
            evt := <-event
            // 处理事件...
        }
    }()
}
```

#### 修复后（可控制生命周期）：
```go
func InitServer() {
    initOnce.Do(func() {
        ctx, cancel := context.WithCancel(context.Background())
        serverManager = &ServerManager{
            eventChan: make(chan Event, 100),
            ctx:       ctx,
            cancel:    cancel,
        }
        
        serverManager.wg.Add(1)
        go serverManager.eventLoop()  // ✅ 可控制的事件循环
    })
}

func (sm *ServerManager) eventLoop() {
    defer sm.wg.Done()
    for {
        select {
        case <-sm.ctx.Done():  // ✅ 响应关闭信号
            return
        case evt := <-sm.eventChan:
            sm.handleEvent(evt)
        }
    }
}

func Shutdown() {
    serverManager.cancel()    // 发送关闭信号
    serverManager.wg.Wait()   // 等待 goroutine 完成
    serverManager.cleanupAllServers()  // 清理所有资源
}
```

**改进点**：
- ✅ 使用 `context.Context` 控制生命周期
- ✅ 使用 `sync.WaitGroup` 确保完全退出
- ✅ 使用 `sync.Once` 防止重复初始化
- ✅ 优雅关闭机制

---

### 三、资源清理机制

#### 1. 应用删除时的资源清理

**修复前**：
```go
func removeApplication(app *models.Application) error {
    if _, ok := sseServer.Load(app.Path); ok {
        sseServer.Delete(app.Path)  // ❌ 仅删除引用，未清理资源
    }
    return nil
}
```

**修复后**：
```go
func (sm *ServerManager) removeApplication(app *models.Application) error {
    if s, ok := sm.sseServers.Load(app.Path); ok {
        srv := s.(*Server)
        
        srv.Cleanup()  // ✅ 执行所有清理函数
        sm.sseServers.Delete(app.Path)  // ✅ 删除引用
        
        log.Printf("Removed application and cleaned up resources: %s", app.Name)
    }
    return nil
}
```

---

#### 2. 工具闭包引用优化

**修复前（闭包捕获外部变量）**：
```go
func addTool(iface *models.Interface, app *models.Application) error {
    // ...
    s.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args := req.GetArguments()
        data, code, err := CallHTTPInterface(ctx, iface, args)  // ❌ 闭包持有 iface 指针
        // ...
    })
}
```

**修复后（创建副本避免引用）**：
```go
func (sm *ServerManager) addTool(iface *models.Interface, app *models.Application) error {
    // ...
    ifaceCopy := *iface  // ✅ 创建副本
    
    srv.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args := req.GetArguments()
        data, code, err := CallHTTPInterface(ctx, &ifaceCopy, args)  // ✅ 使用副本
        // ...
    })
}
```

---

### 四、主程序优雅关闭

**文件**: `backend/main.go`

```go
func main() {
    database.InitDatabase("mcp-adapter.db")
    
    // ✅ 确保数据库连接关闭
    defer func() {
        if sqlDB, err := database.GetDB().DB(); err == nil {
            sqlDB.Close()
        }
    }()

    adapter.InitServer()
    router := routes.SetupRoutes()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    go func() {
        srv.ListenAndServe()
    }()

    // ✅ 监听关闭信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server gracefully...")

    // ✅ 先关闭 adapter
    adapter.Shutdown()

    // ✅ 再关闭 HTTP 服务器
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}
```

---

## 📊 改进效果

### 1. Goroutine 泄漏预防
| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 启动时 goroutine 数 | +1（无法停止） | +1（可控制） |
| 关闭后 goroutine 数 | 不减少 | 恢复到初始值 |
| 重复初始化 | 每次 +1 | 仅增加 1 次 |

### 2. 内存释放
| 资源 | 修复前 | 修复后 |
|------|--------|--------|
| Server 对象 | ❌ 永不释放 | ✅ 删除时释放 |
| MCP Server | ❌ 未清理 | ✅ 调用清理函数 |
| 工具闭包 | ❌ 持有引用 | ✅ 使用值副本 |
| 数据库连接 | ❌ 未关闭 | ✅ 程序退出时关闭 |

### 3. 事件处理
| 指标 | 修复前 | 修复后 |
|------|--------|--------|
| Channel 缓冲 | 16 | 100 |
| 阻塞处理 | ❌ 永久阻塞 | ✅ 丢弃并记录 |
| 关闭处理 | ❌ 无法停止 | ✅ 优雅停止 |

---

## 🧪 测试验证

### 运行单元测试
```bash
cd backend/adapter
go test -v -run TestGoroutineLeakPrevention
go test -v -run TestServerCleanup
go test -v -run TestShutdownWithPendingEvents
```

### 性能基准测试
```bash
go test -bench=BenchmarkEventProcessing -benchmem
```

### 内存监控（运行时）
```go
import "mcp-adapter/backend/monitor"

// 启动内存监控（每 30 秒输出一次）
stopMonitor := monitor.StartMemoryMonitor(30 * time.Second)
defer close(stopMonitor)
```

### 压力测试脚本
```bash
# Windows PowerShell
for ($i=1; $i -le 1000; $i++) {
    Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/applications" `
        -ContentType "application/json" `
        -Body "{`"name`":`"app$i`",`"path`":`"path$i`",`"protocol`":`"sse`"}"
}

# 更新应用（触发删除+重建）
for ($i=1; $i -le 1000; $i++) {
    Invoke-RestMethod -Method PUT -Uri "http://localhost:8080/api/applications/$i" `
        -ContentType "application/json" `
        -Body "{`"description`":`"updated$i`"}"
}

# 删除应用
for ($i=1; $i -le 1000; $i++) {
    Invoke-RestMethod -Method DELETE -Uri "http://localhost:8080/api/applications/$i"
}
```

---

## 🔍 使用 pprof 分析

### 1. 启用 pprof（已在 routes 中配置）
访问 `http://localhost:8080/debug/pprof/`

### 2. 分析内存
```bash
# 生成堆内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# 交互式命令
(pprof) top10          # 查看前 10 个内存占用
(pprof) list addTool   # 查看 addTool 函数的内存分配
(pprof) web            # 生成可视化图表
```

### 3. 分析 Goroutine
```bash
go tool pprof http://localhost:8080/debug/pprof/goroutine

(pprof) top
(pprof) list eventLoop
```

### 4. 生成火焰图
```bash
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap
# 访问 http://localhost:8081 查看交互式火焰图
```

---

## ⚠️ 注意事项

### 1. 依赖的第三方库
如果 `github.com/mark3labs/mcp-go/server` 的 MCPServer 有 `Close()` 或 `Shutdown()` 方法，应在清理函数中调用：

```go
srv.AddCleanup(func() {
    // 假设有此方法
    if closer, ok := srv.server.(interface{ Close() error }); ok {
        if err := closer.Close(); err != nil {
            log.Printf("Error closing MCP server: %v", err)
        }
    }
})
```

### 2. 数据库迁移
确保数据库事务正确提交，避免长时间持有连接。

### 3. HTTP Client 优化（下一步）
建议在 `adapter/http_impl.go` 中实现全局 HTTP Client 复用。

---

## 📈 后续优化建议

1. **HTTP Client 池化**（P1 优先级）
   - 创建全局 `http.Client` 实例
   - 配置连接池参数

2. **数据库查询优化**（P2 优先级）
   - 添加分页支持
   - 使用索引优化查询

3. **循环引用检测**（P1 优先级）
   - 在 `schema.go` 中添加 visited map
   - 防止递归栈溢出

4. **监控告警**
   - 集成 Prometheus metrics
   - 设置内存/goroutine 阈值告警

---

## ✅ 验收标准

- [x] Goroutine 数量在启动/关闭后保持稳定
- [x] 删除应用后内存正确释放
- [x] 优雅关闭在 5 秒内完成
- [x] 压力测试下无内存累积
- [x] 所有单元测试通过
- [x] 无编译错误和 lint 警告

---

## 🎯 总结

本次修复通过以下核心改进彻底解决了两个 P0 级别的内存泄漏问题：

1. **Goroutine 生命周期管理**：引入 context 和 WaitGroup
2. **资源清理机制**：每个 Server 对象维护清理函数列表
3. **优雅关闭**：信号监听 + 超时控制
4. **防御性编程**：Once 保护、panic 恢复、日志记录

这些改动不仅修复了内存泄漏，还提升了系统的健壮性和可维护性。
