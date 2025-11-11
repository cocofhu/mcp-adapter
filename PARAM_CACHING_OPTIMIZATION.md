# 参数缓存与默认值优化

## 概述

优化了 MCP 工具调用流程，将参数信息（包括默认值）在工具注册时缓存，避免在每次调用时重复查询数据库，提升性能并简化逻辑。

## 问题背景

### 原有流程

在原有实现中，每次调用 MCP 工具时都需要查询数据库：

```go
// 在 addTool 中注册工具
srv.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) {
    args := req.GetArguments()
    // 直接调用 CallHTTPInterface
    data, code, err := CallHTTPInterface(ctx, &ifaceCopy, args)
    // ...
})

// 在 BuildHTTPRequest 中查询数据库
func BuildHTTPRequest(ctx context.Context, iface *models.Interface, args map[string]any) {
    // 每次调用都查库
    db := database.GetDB()
    var params []models.InterfaceParameter
    db.Where("interface_id = ?", iface.ID).Find(&params)
    
    // 应用默认值
    for _, p := range params {
        if p.DefaultValue != nil && args[p.Name] == nil {
            args[p.Name] = *p.DefaultValue
        }
    }
    // ...
}
```

### 存在的问题

1. **性能问题**：每次工具调用都要查询数据库，增加延迟
2. **类型问题**：默认值以字符串形式存储，但没有根据参数类型转换
3. **验证滞后**：必填参数验证在构建请求时才进行，错误反馈较晚
4. **重复查询**：同一个工具可能被多次调用，每次都重复查询相同的参数信息

## 优化方案

### 1. 参数缓存

在工具注册时（`addTool`）就查询并缓存参数信息：

```go
// 从数据库获取接口参数（只查询一次）
var params []models.InterfaceParameter
if db.Where("interface_id = ?", iface.ID).Find(&params).Error != nil {
    return fmt.Errorf("error getting interface parameters for tool %s", iface.Name)
}

// 创建参数副本，缓存参数信息避免在调用时查库
paramsCopy := make([]models.InterfaceParameter, len(params))
copy(paramsCopy, params)

// 在闭包中使用缓存的参数
srv.server.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) {
    args := req.GetArguments()
    
    // 使用缓存的参数，不需要查库
    processedArgs, err := applyDefaultsAndValidate(args, paramsCopy)
    // ...
})
```

### 2. 默认值应用与类型转换

新增 `applyDefaultsAndValidate` 函数，在工具调用时立即处理：

```go
func applyDefaultsAndValidate(args map[string]any, params []models.InterfaceParameter) (map[string]any, error) {
    processedArgs := make(map[string]any)
    
    // 复制提供的参数
    for k, v := range args {
        processedArgs[k] = v
    }
    
    // 应用默认值并验证
    for _, p := range params {
        _, provided := processedArgs[p.Name]
        
        // 应用默认值（只对基本类型）
        if !provided && p.DefaultValue != nil && *p.DefaultValue != "" {
            convertedVal, err := convertDefaultValue(*p.DefaultValue, p.Type)
            if err != nil {
                return nil, fmt.Errorf("failed to convert default value for parameter %s: %w", p.Name, err)
            }
            processedArgs[p.Name] = convertedVal
            log.Printf("Applied default value for parameter %s: %v", p.Name, convertedVal)
        }
        
        // 验证必填参数
        if p.Required {
            finalVal, exists := processedArgs[p.Name]
            if !exists || finalVal == nil {
                return nil, fmt.Errorf("missing required parameter: %s", p.Name)
            }
        }
    }
    
    return processedArgs, nil
}
```

### 3. 类型转换

新增 `convertDefaultValue` 函数，根据参数类型正确转换默认值：

```go
func convertDefaultValue(defaultValue string, paramType string) (any, error) {
    switch paramType {
    case "number":
        // 尝试转换为 float64
        if val, err := strconv.ParseFloat(defaultValue, 64); err == nil {
            return val, nil
        }
        // 如果失败，尝试转换为 int
        if val, err := strconv.ParseInt(defaultValue, 10, 64); err == nil {
            return val, nil
        }
        return nil, fmt.Errorf("invalid number format: %s", defaultValue)
    case "boolean":
        val, err := strconv.ParseBool(defaultValue)
        if err != nil {
            return nil, fmt.Errorf("invalid boolean format: %s", defaultValue)
        }
        return val, nil
    case "string":
        return defaultValue, nil
    default:
        // 自定义类型不应该有默认值
        return defaultValue, nil
    }
}
```

### 4. 无查库的 HTTP 调用

新增 `CallHTTPInterfaceWithParams` 和 `BuildHTTPRequestWithParams` 函数：

```go
// 使用提供的参数列表，不查询数据库
func CallHTTPInterfaceWithParams(ctx context.Context, iface *models.Interface, args map[string]any, params []models.InterfaceParameter) ([]byte, int, error) {
    req, err := BuildHTTPRequestWithParams(ctx, iface, args, params)
    if err != nil {
        return nil, 0, err
    }
    // ... 执行请求
}

func BuildHTTPRequestWithParams(ctx context.Context, iface *models.Interface, args map[string]any, params []models.InterfaceParameter) (*http.Request, error) {
    // 直接使用传入的参数列表，不查询数据库
    paramIndex := make(map[string]models.InterfaceParameter)
    for _, p := range params {
        paramIndex[p.Name] = p
    }
    // ... 构建请求
}
```

## 优化效果

### 性能提升

| 场景 | 原方案 | 优化后 | 提升 |
|-----|-------|-------|-----|
| 单次工具调用 | 1次数据库查询 | 0次查询 | ✅ 100% |
| 100次工具调用 | 100次查询 | 0次查询 | ✅ 节省100次查询 |
| 工具注册时 | 1次查询 | 1次查询 | - 相同 |

**总结**：
- ✅ 工具调用时完全避免数据库查询
- ✅ 只在工具注册时查询一次数据库
- ✅ 高频调用场景下性能显著提升

### 功能增强

1. **类型安全**
   - ✅ 默认值根据参数类型正确转换
   - ✅ number 类型转换为数值
   - ✅ boolean 类型转换为布尔值
   - ✅ string 类型保持字符串

2. **快速失败**
   - ✅ 在工具调用开始时就验证参数
   - ✅ 缺少必填参数立即返回错误
   - ✅ 默认值格式错误立即返回错误

3. **日志改进**
   - ✅ 记录默认值应用情况
   - ✅ 便于调试和问题排查

## 使用示例

### 示例 1：带默认值的分页接口

**接口定义**：
```json
{
  "name": "GetUsers",
  "url": "https://api.example.com/users",
  "method": "GET",
  "parameters": [
    {"name": "page", "type": "number", "location": "query", "default_value": "1"},
    {"name": "pageSize", "type": "number", "location": "query", "default_value": "10"}
  ]
}
```

**调用场景**：

1. **不提供任何参数**：
   ```javascript
   // MCP 调用
   callTool("GetUsers", {})
   
   // 日志输出
   // Applied default value for parameter page: 1
   // Applied default value for parameter pageSize: 10
   
   // 实际请求
   // GET https://api.example.com/users?page=1&pageSize=10
   ```

2. **只提供部分参数**：
   ```javascript
   // MCP 调用
   callTool("GetUsers", {"page": 2})
   
   // 日志输出
   // Applied default value for parameter pageSize: 10
   
   // 实际请求
   // GET https://api.example.com/users?page=2&pageSize=10
   ```

3. **提供所有参数**：
   ```javascript
   // MCP 调用
   callTool("GetUsers", {"page": 3, "pageSize": 20})
   
   // 日志输出
   // (无默认值应用日志)
   
   // 实际请求
   // GET https://api.example.com/users?page=3&pageSize=20
   ```

### 示例 2：必填参数验证

**接口定义**：
```json
{
  "name": "CreateUser",
  "url": "https://api.example.com/users",
  "method": "POST",
  "parameters": [
    {"name": "name", "type": "string", "location": "body", "required": true},
    {"name": "email", "type": "string", "location": "body", "required": true},
    {"name": "enabled", "type": "boolean", "location": "body", "default_value": "true"}
  ]
}
```

**调用场景**：

1. **缺少必填参数**：
   ```javascript
   // MCP 调用
   callTool("CreateUser", {"name": "Alice"})
   
   // 返回错误
   // Error: missing required parameter: email
   ```

2. **提供必填参数**：
   ```javascript
   // MCP 调用
   callTool("CreateUser", {"name": "Alice", "email": "alice@example.com"})
   
   // 日志输出
   // Applied default value for parameter enabled: true
   
   // 实际请求
   // POST https://api.example.com/users
   // Body: {"name": "Alice", "email": "alice@example.com", "enabled": true}
   ```

### 示例 3：类型转换

**接口定义**：
```json
{
  "parameters": [
    {"name": "count", "type": "number", "default_value": "10"},
    {"name": "active", "type": "boolean", "default_value": "true"},
    {"name": "keyword", "type": "string", "default_value": "search"}
  ]
}
```

**默认值转换**：
```go
// "10" -> float64(10) 或 int64(10)
// "true" -> bool(true)
// "search" -> string("search")
```

## 代码变更总结

### 新增文件
无

### 修改文件

#### 1. `backend/adapter/mcp_model.go`

**新增函数**：
- `applyDefaultsAndValidate(args, params)` - 应用默认值并验证参数
- `convertDefaultValue(defaultValue, paramType)` - 类型转换

**修改函数**：
- `addTool()` - 缓存参数信息，使用新的调用流程

**新增导入**：
- `strconv` - 用于类型转换

#### 2. `backend/adapter/http_impl.go`

**新增函数**：
- `CallHTTPInterfaceWithParams()` - 使用缓存参数的调用版本
- `BuildHTTPRequestWithParams()` - 使用缓存参数的构建版本

**保留函数**：
- `CallHTTPInterface()` - 保留用于兼容性
- `BuildHTTPRequest()` - 保留用于其他场景

## 兼容性说明

1. **向后兼容**：
   - ✅ 保留了原有的 `CallHTTPInterface` 函数
   - ✅ 保留了原有的 `BuildHTTPRequest` 函数
   - ✅ 其他模块可以继续使用原有函数

2. **新旧对比**：
   - MCP 工具调用：使用新的优化流程（缓存参数）
   - 其他场景：可以继续使用原有流程（查询数据库）

## 测试建议

### 单元测试

1. **默认值应用测试**：
   - 测试 number 类型默认值转换
   - 测试 boolean 类型默认值转换
   - 测试 string 类型默认值保持
   - 测试无效默认值格式的错误处理

2. **参数验证测试**：
   - 测试必填参数缺失的错误
   - 测试必填参数存在且有默认值
   - 测试必填参数存在且无默认值

3. **类型转换测试**：
   ```go
   // number: "10" -> 10
   // number: "3.14" -> 3.14
   // boolean: "true" -> true
   // boolean: "false" -> false
   // string: "hello" -> "hello"
   ```

### 集成测试

1. **创建带默认值的接口**
2. **调用 MCP 工具，不提供参数**
3. **验证默认值被正确应用**
4. **验证请求发送正确**

### 性能测试

1. **对比测试**：
   - 测试原方案：100次工具调用的数据库查询次数
   - 测试优化后：100次工具调用的数据库查询次数
   - 预期：优化后为0次查询

2. **压力测试**：
   - 并发调用同一个工具1000次
   - 验证无数据库查询
   - 验证响应时间稳定

## 最佳实践

1. **设置合理的默认值**：
   - 分页参数：`page=1`, `pageSize=10`
   - 排序参数：`order=asc`, `sortBy=created_at`
   - 过滤参数：`enabled=true`, `status=active`

2. **类型匹配**：
   - number 类型参数使用数字格式的默认值：`"10"`, `"3.14"`
   - boolean 类型参数使用布尔格式的默认值：`"true"`, `"false"`
   - string 类型参数使用字符串默认值：`"default"`, `""`

3. **避免的做法**：
   - ❌ 为自定义类型设置默认值（前端已限制）
   - ❌ 使用无效格式的默认值（会在运行时报错）
   - ❌ 过度使用默认值（应该让必填参数保持必填）

## 总结

这次优化实现了：
- ✅ **性能提升**：消除工具调用时的数据库查询
- ✅ **类型安全**：默认值根据参数类型正确转换
- ✅ **快速验证**：在调用开始时就验证参数完整性
- ✅ **更好的日志**：记录默认值应用情况
- ✅ **向后兼容**：保留原有函数供其他场景使用

这是一个重要的性能优化，特别适合高频调用的 MCP 工具场景。
