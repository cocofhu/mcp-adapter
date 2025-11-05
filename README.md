# MCP Adapter - HTTP 接口管理系统

一个现代化的 HTTP/REST API 接口管理系统，支持自定义类型定义和 MCP (Model Context Protocol) 协议适配。

## ✨ 功能特性

### 🎨 自定义类型系统（新增）
- **类型定义** - 创建可复用的自定义类型（类似 TypeScript interface）
- **嵌套类型** - 支持类型之间的引用和嵌套
- **数组支持** - 支持数组类型定义
- **类型复用** - 在多个接口间共享类型定义
- **引用完整性** - 自动检查类型引用的有效性

### 🔌 接口管理
- **多种 HTTP 方法** - 支持 GET、POST、PUT、DELETE、PATCH、HEAD、OPTIONS
- **灵活参数配置** - 支持 query、header、body、path 四种参数位置
- **参数类型** - 支持基本类型（number, string, boolean）和自定义类型
- **默认值支持** - 为参数设置默认值
- **必填验证** - 自动验证必填参数

### 📋 应用管理
- **多应用支持** - 管理多个独立的应用
- **MCP 协议** - 支持 SSE (Server-Sent Events) 协议
- **应用隔离** - 每个应用有独立的接口和类型定义

### ⚙️ 高级特性
- **事务支持** - 保证数据一致性
- **批量查询优化** - 避免 N+1 查询问题
- **引用检查** - 防止删除被引用的类型
- **数据验证** - 完整的输入验证

## 🚀 快速开始

### 安装依赖

```bash
go mod download
```

### 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 访问前端界面

打开浏览器访问 `http://localhost:8080`，即可使用 Web 界面管理应用、类型和接口。

详细使用说明请参考 [前端使用指南](./FRONTEND_GUIDE.md)。

### 运行测试

**Linux/Mac**:
```bash
chmod +x test_api.sh
./test_api.sh
```

**Windows**:
```powershell
.\test_api.ps1
```

## 📖 使用指南

### 1. 创建应用

```bash
curl -X POST http://localhost:8080/api/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My App",
    "path": "myapp",
    "protocol": "sse",
    "enabled": true
  }'
```

### 2. 创建自定义类型

```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "User",
    "description": "用户信息",
    "fields": [
      {"name": "id", "type": "number", "required": true},
      {"name": "name", "type": "string", "required": true},
      {"name": "email", "type": "string", "required": false}
    ]
  }'
```

### 3. 创建接口

```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "GetUser",
    "protocol": "http",
    "url": "https://api.example.com/users",
    "method": "GET",
    "auth_type": "none",
    "parameters": [
      {
        "name": "id",
        "type": "string",
        "location": "query",
        "required": true
      }
    ]
  }'
```

更多示例请参考 [API_EXAMPLES.md](./API_EXAMPLES.md)。

## 📚 文档

- [前端使用指南](./FRONTEND_GUIDE.md) - 前端界面使用说明
- [API 使用示例](./API_EXAMPLES.md) - 完整的 API 使用示例
- [数据库迁移指南](./MIGRATION.md) - 从旧版本迁移的指南
- [重构总结](./REFACTORING_SUMMARY.md) - 项目重构的详细说明
- [API 文档](./README-API.md) - 详细的 API 文档

## 🏗️ 项目结构

```
mcp-adapter/
├── backend/
│   ├── adapter/          # MCP 协议适配器
│   │   ├── mcp_model.go  # MCP 工具注册
│   │   └── http_impl.go  # HTTP 请求构建
│   ├── database/         # 数据库配置
│   ├── handlers/         # HTTP 处理器
│   │   ├── application.go
│   │   ├── interface.go
│   │   └── custom_type.go
│   ├── models/           # 数据模型
│   │   └── models.go
│   ├── routes/           # 路由配置
│   └── service/          # 业务逻辑
│       ├── application_service.go
│       ├── interface_service.go
│       └── custom_type_service.go
├── web/                  # 前端文件
├── test_api.sh          # Linux/Mac 测试脚本
├── test_api.ps1         # Windows 测试脚本
└── main.go              # 入口文件
```

## 🔧 技术栈

- **后端**: Go 1.21+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **协议**: MCP (Model Context Protocol)

## 📊 数据模型

### Application (应用)
- 管理多个独立的应用
- 每个应用有独立的接口和类型定义

### CustomType (自定义类型)
- 定义可复用的复杂类型
- 支持嵌套和引用

### CustomTypeField (类型字段)
- 定义类型包含的字段
- 支持基本类型和自定义类型引用

### Interface (接口)
- HTTP 接口定义
- 关联参数定义

### InterfaceParameter (接口参数)
- 接口的参数定义
- 支持基本类型和自定义类型引用

## 🎯 API 端点

### 应用管理
- `POST /api/applications` - 创建应用
- `GET /api/applications` - 获取应用列表
- `GET /api/applications/:id` - 获取单个应用
- `PUT /api/applications/:id` - 更新应用
- `DELETE /api/applications/:id` - 删除应用

### 自定义类型
- `POST /api/custom-types` - 创建自定义类型
- `GET /api/custom-types?app_id=1` - 获取应用的类型列表
- `GET /api/custom-types/:id` - 获取单个类型
- `PUT /api/custom-types/:id` - 更新类型
- `DELETE /api/custom-types/:id` - 删除类型

### 接口管理
- `POST /api/interfaces` - 创建接口
- `GET /api/interfaces?app_id=1` - 获取应用的接口列表
- `GET /api/interfaces/:id` - 获取单个接口
- `PUT /api/interfaces/:id` - 更新接口
- `DELETE /api/interfaces/:id` - 删除接口

## 🔄 从旧版本迁移

如果你正在从旧版本（使用 `Options` JSON 字段）迁移，请参考 [MIGRATION.md](./MIGRATION.md)。

**快速迁移（开发环境）**:
```bash
# 删除旧数据库
rm mcp-adapter.db

# 重新启动，自动创建新表结构
go run main.go
```

## 🧪 测试

项目包含完整的 API 测试脚本：

```bash
# Linux/Mac
./test_api.sh

# Windows
.\test_api.ps1
```

测试覆盖：
- ✅ 应用 CRUD
- ✅ 自定义类型 CRUD
- ✅ 接口 CRUD
- ✅ 嵌套类型引用
- ✅ 错误处理
- ✅ 数据验证

## 🛠️ 开发计划

- [x] 自定义类型系统
- [x] 接口参数关联表
- [x] 事务支持
- [x] 引用完整性检查
- [ ] 自定义类型递归展开（MCP Schema）
- [ ] Path 参数支持
- [ ] 更多认证方式
- [ ] 接口版本管理
- [ ] GraphQL 支持
- [ ] 代码生成器

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**注意**: 本项目正在积极开发中，API 可能会有变化。