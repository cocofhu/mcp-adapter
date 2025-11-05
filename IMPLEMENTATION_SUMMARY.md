# 循环引用检测功能实现总结

## 实现内容

为 MCP Adapter 项目添加了**自定义类型和接口参数的循环引用检测**功能,使用 **DFS (深度优先搜索)** 算法进行判环。

## 修改的文件

### 1. `backend/service/custom_type_service.go`

**新增函数:**
- `checkCustomTypeCycle()`: 检测自定义类型创建时的循环引用
- `checkCustomTypeCycleForUpdate()`: 检测自定义类型更新时的循环引用

**修改函数:**
- `CreateCustomType()`: 添加循环引用检测调用
- `UpdateCustomType()`: 添加循环引用检测调用

**导入包:**
- 添加 `"gorm.io/gorm"` 导入

### 2. `backend/service/interface_service.go`

**新增函数:**
- `checkInterfaceParameterCycle()`: 检测接口参数创建时的循环引用
- `checkInterfaceParameterCycleForUpdate()`: 检测接口参数更新时的循环引用

**修改函数:**
- `CreateInterface()`: 添加循环引用检测调用
- `UpdateInterface()`: 添加循环引用检测调用

**导入包:**
- 添加 `"gorm.io/gorm"` 导入

## 新增文件

### 1. `CYCLE_DETECTION.md`
循环引用检测功能的详细文档,包括:
- 功能概述
- 检测场景说明
- 实现细节
- 错误信息
- 性能分析
- 注意事项

### 2. `test_cycle_detection.ps1`
Windows PowerShell 测试脚本,用于验证循环引用检测功能。

### 3. `test_cycle_detection.sh`
Linux/Mac Bash 测试脚本,用于验证循环引用检测功能。

### 4. `IMPLEMENTATION_SUMMARY.md`
本文档,实现总结。

## 核心算法

### DFS 判环算法

```go
// 伪代码
visited := make(map[int64]bool)      // 已访问节点
recStack := make(map[int64]bool)     // 递归栈中的节点

func dfs(node int64) bool {
    visited[node] = true
    recStack[node] = true
    
    for _, neighbor := range graph[node] {
        if !visited[neighbor] {
            if dfs(neighbor) {
                return true  // 发现环
            }
        } else if recStack[neighbor] {
            return true  // 发现环
        }
    }
    
    recStack[node] = false
    return false
}
```

### 检测流程

1. **构建引用图**: 遍历所有自定义类型及其字段,构建 `typeID -> []refTypeID` 的引用关系图
2. **DFS 遍历**: 从当前操作的类型开始,使用 DFS 检测是否存在环
3. **返回结果**: 如果发现环,返回错误;否则允许操作继续

## 检测场景

### 自定义类型

✅ **允许的场景:**
```
TypeA { name: string }
TypeB { refA: TypeA }
TypeC { refA: TypeA }
```

❌ **禁止的场景:**
```
// 循环引用
TypeA { refB: TypeB }
TypeB { refA: TypeA }

// 自引用
TypeA { self: TypeA }

// 间接循环
TypeA { refB: TypeB }
TypeB { refC: TypeC }
TypeC { refA: TypeA }
```

### 接口参数

接口参数引用的自定义类型也会被检测循环引用。

✅ **允许的场景:**
```
Interface {
  parameter: TypeB  // TypeB 引用 TypeA,但不形成环
}
```

❌ **禁止的场景:**
```
Interface {
  parameter: TypeA  // TypeA 和 TypeB 形成环
}
```

## 错误信息

- 自定义类型: `"circular reference detected in custom type fields"`
- 接口参数: `"circular reference detected in interface parameters"`

## 性能分析

- **时间复杂度**: O(V + E)
  - V: 类型数量
  - E: 引用关系数量
  
- **空间复杂度**: O(V)
  - 用于存储访问状态和递归栈

对于典型应用场景(类型数量 < 1000),性能影响可忽略不计。

## 测试方法

### 启动服务
```bash
cd backend
go run main.go
```

### 运行测试

**Windows:**
```powershell
.\test_cycle_detection.ps1
```

**Linux/Mac:**
```bash
chmod +x test_cycle_detection.sh
./test_cycle_detection.sh
```

### 测试覆盖

测试脚本验证以下场景:
1. ✅ 创建正常的类型引用链 (A -> B)
2. ❌ 创建循环引用 (A -> B -> A)
3. ❌ 创建自引用 (A -> A)
4. ✅ 创建不形成环的多重引用
5. ✅ 接口参数使用正常类型
6. ❌ 接口参数使用有循环的类型

## 前端集成

前端无需修改,已有的错误处理机制会自动显示循环引用错误:

```javascript
// app.js 中的错误处理
catch (error) {
    showToast(error.message, 'error');
}
```

当后端返回 `"circular reference detected..."` 错误时,前端会通过 Toast 通知用户。

## 注意事项

1. **事务安全**: 循环引用检测在事务提交前进行,检测失败会自动回滚
2. **应用隔离**: 检测范围限定在同一应用内
3. **数组支持**: 数组类型 (`is_array: true`) 也会被检测
4. **更新策略**: 更新类型时,使用新字段列表进行检测

## 后续优化建议

1. **缓存优化**: 对于频繁访问的类型关系图,可以考虑缓存
2. **批量检测**: 如果需要批量创建类型,可以优化为一次性检测
3. **可视化**: 前端可以添加类型关系图可视化,帮助用户理解引用关系
4. **详细错误**: 可以在错误信息中包含具体的循环路径,如 "A -> B -> C -> A"

## 总结

本次实现为 MCP Adapter 添加了完善的循环引用检测机制,确保数据模型的完整性和一致性。使用经典的 DFS 算法,性能优秀,易于维护。
