# 更新日志

## [2.0.0] - 2024-01-XX

### 🎉 重大更新

#### 新增功能

**自定义类型系统**
- ✨ 新增 `CustomType` 模型 - 支持定义可复用的复杂类型
- ✨ 新增 `CustomTypeField` 模型 - 定义类型的字段结构
- ✨ 支持类型嵌套 - 自定义类型可以引用其他自定义类型
- ✨ 支持数组类型 - 通过 `is_array` 字段定义数组
- ✨ 完整的类型 CRUD API - 创建、读取、更新、删除自定义类型

**接口参数管理**
- ✨ 新增 `InterfaceParameter` 模型 - 使用关联表管理接口参数
- ✨ 支持参数位置 - query, header, body, path
- ✨ 支持默认值 - 为参数设置默认值
- ✨ 支持自定义类型引用 - 参数可以引用自定义类型

**数据完整性**
- ✨ 引用完整性检查 - 防止删除被引用的类型
- ✨ 事务支持 - 保证数据的原子性操作
- ✨ 批量查询优化 - 避免 N+1 查询问题

#### 重大变更

**Interface 模型重构**
- ⚠️ **破坏性变更**: 移除 `Options` 字段（JSON 字符串）
- ✅ 新增 `Method` 字段 - 直接存储 HTTP 方法
- ✅ 使用 `InterfaceParameter` 关联表管理参数

**API 变更**
- ⚠️ 接口创建/更新 API 的请求格式已变更
- ✅ 新增自定义类型管理 API
- ✅ 接口参数现在作为数组直接传递，而非嵌套在 `options` 中

#### 改进

**代码质量**
- 🔧 重构 `interface_service.go` - 使用关联表替代 JSON
- 🔧 重构 `http_impl.go` - 简化 HTTP 请求构建逻辑
- 🔧 重构 `mcp_model.go` - 从数据库读取参数定义
- 🔧 移除不再使用的结构体 - `ToolOptions`, `HTTPOptions` 等

**性能优化**
- ⚡ 批量查询参数 - 列表查询时一次性获取所有参数
- ⚡ 减少 JSON 解析 - 直接使用关系型数据
- ⚡ 索引优化 - 为关联字段添加索引

**文档**
- 📝 新增 `API_EXAMPLES.md` - 完整的 API 使用示例
- 📝 新增 `MIGRATION.md` - 数据库迁移指南
- 📝 新增 `REFACTORING_SUMMARY.md` - 重构详细说明
- 📝 更新 `README.md` - 反映新功能和架构
- 📝 新增测试脚本 - `test_api.sh` 和 `test_api.ps1`

#### 数据库变更

**新增表**
- `custom_types` - 自定义类型定义
- `custom_type_fields` - 自定义类型字段
- `interface_parameters` - 接口参数

**修改表**
- `interfaces` - 移除 `options` 列，新增 `method` 列

### 迁移指南

#### 从 1.x 迁移到 2.0

**开发环境（推荐）**:
```bash
# 删除旧数据库
rm mcp-adapter.db

# 重新启动，自动创建新表结构
go run main.go
```

**生产环境**:
请参考 [MIGRATION.md](./MIGRATION.md) 中的详细步骤。

#### API 变更示例

**旧版本 (1.x)**:
```json
{
  "name": "GetUser",
  "protocol": "http",
  "url": "https://api.example.com/users",
  "options": {
    "method": "GET",
    "parameters": [
      {
        "name": "id",
        "type": "string",
        "required": true,
        "location": "query"
      }
    ]
  }
}
```

**新版本 (2.0)**:
```json
{
  "name": "GetUser",
  "protocol": "http",
  "url": "https://api.example.com/users",
  "method": "GET",
  "parameters": [
    {
      "name": "id",
      "type": "string",
      "required": true,
      "location": "query"
    }
  ]
}
```

### 新增 API 端点

#### 自定义类型管理
- `POST /api/custom-types` - 创建自定义类型
- `GET /api/custom-types?app_id=1` - 获取应用的类型列表
- `GET /api/custom-types/:id` - 获取单个类型
- `PUT /api/custom-types/:id` - 更新类型
- `DELETE /api/custom-types/:id` - 删除类型

### 技术债务

#### 已解决
- ✅ 参数管理使用 JSON 字符串 → 使用关联表
- ✅ 缺少类型复用机制 → 自定义类型系统
- ✅ 难以查询和过滤参数 → 关系型数据库设计
- ✅ 缺少引用完整性检查 → 外键约束和业务逻辑验证

#### 待解决
- ⏳ 自定义类型递归展开（MCP Schema）
- ⏳ Path 参数替换实现
- ⏳ 更多认证方式支持
- ⏳ 接口版本管理

### 已知问题

- 自定义类型在 MCP 工具注册时暂时作为 string 处理，后续版本将支持递归展开
- Path 参数（如 `/users/{id}`）暂未实现占位符替换

### 贡献者

感谢所有为本次重构做出贡献的开发者！

---

## [1.0.0] - 2024-01-XX

### 初始版本

- ✨ 应用管理功能
- ✨ 接口管理功能
- ✨ HTTP 协议支持
- ✨ MCP SSE 协议支持
- ✨ 基于 JSON 的参数配置
