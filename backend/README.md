# MCP Adapter Backend

HTTP接口管理系统后端服务，使用Go语言开发，基于Gin框架和GORM。

## 功能特性

- 🚀 **应用管理**: 创建、编辑、删除API应用
- 🔌 **接口管理**: 管理HTTP API接口，支持多种认证方式
- 📝 **参数配置**: 支持请求参数、默认参数和默认请求头配置
- 🧪 **接口测试**: 内置HTTP客户端，支持接口测试
- 📊 **统计信息**: 应用和接口的统计数据
- 🗄️ **多数据库支持**: 支持MySQL、PostgreSQL、SQLite
- 📚 **API文档**: 自动生成Swagger文档
- 🐳 **Docker支持**: 容器化部署

## 技术栈

- **框架**: Gin Web Framework
- **ORM**: GORM
- **数据库**: MySQL/PostgreSQL/SQLite
- **文档**: Swagger/OpenAPI
- **配置**: 环境变量
- **日志**: 结构化日志
- **容器**: Docker

## 项目结构

```
backend/
├── config/          # 配置管理
├── database/        # 数据库连接和迁移
├── dto/            # 数据传输对象
├── handlers/       # HTTP处理器
├── middleware/     # 中间件
├── models/         # 数据模型
├── repositories/   # 数据访问层
├── routes/         # 路由配置
├── services/       # 业务逻辑层
├── main.go         # 程序入口
├── Dockerfile      # Docker配置
├── Makefile        # 构建脚本
└── README.md       # 项目文档
```

## 快速开始

### 环境要求

- Go 1.21+
- 数据库 (MySQL/PostgreSQL/SQLite)

### 安装依赖

```bash
# 下载依赖
go mod download
go mod tidy

# 或使用 Makefile
make deps
```

### 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
vim .env
```

### 运行应用

```bash
# 直接运行
go run main.go

# 或使用 Makefile
make run

# 开发模式 (需要安装 air)
make dev
```

### 构建应用

```bash
# 构建
make build

# 跨平台构建
make build-linux
make build-windows
make build-mac
```

## API文档

启动服务后，访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

## API 端点

### 应用管理

- `POST /api/applications` - 创建应用
- `GET /api/applications` - 获取应用列表
- `GET /api/applications/{id}` - 获取应用详情
- `PUT /api/applications/{id}` - 更新应用
- `DELETE /api/applications/{id}` - 删除应用
- `GET /api/applications/{id}/stats` - 获取应用统计

### 接口管理

- `POST /api/interfaces` - 创建接口
- `GET /api/interfaces` - 获取接口列表
- `GET /api/interfaces/{id}` - 获取接口详情
- `PUT /api/interfaces/{id}` - 更新接口
- `DELETE /api/interfaces/{id}` - 删除接口
- `PATCH /api/interfaces/{id}/toggle` - 切换接口状态
- `POST /api/interfaces/{id}/test` - 测试接口

### 健康检查

- `GET /health` - 健康检查

## 数据库配置

### SQLite (默认)

```env
DB_DRIVER=sqlite
DB_DATABASE=mcp_adapter.db
```

### MySQL

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=password
DB_DATABASE=mcp_adapter
DB_CHARSET=utf8mb4
```

### PostgreSQL

```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=password
DB_DATABASE=mcp_adapter
DB_SSL_MODE=disable
```

## Docker 部署

### 构建镜像

```bash
# 使用 Dockerfile
docker build -t mcp-adapter .

# 或使用 Makefile
make docker-build
```

### 运行容器

```bash
# 直接运行
docker run -p 8080:8080 mcp-adapter

# 或使用 Makefile
make docker-run

# 使用环境变量
docker run -p 8080:8080 \
  -e DB_DRIVER=mysql \
  -e DB_HOST=host.docker.internal \
  -e DB_USERNAME=root \
  -e DB_PASSWORD=password \
  -e DB_DATABASE=mcp_adapter \
  mcp-adapter
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_DRIVER=mysql
      - DB_HOST=db
      - DB_USERNAME=root
      - DB_PASSWORD=password
      - DB_DATABASE=mcp_adapter
    depends_on:
      - db
  
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
      - MYSQL_DATABASE=mcp_adapter
    ports:
      - "3306:3306"
```

## 开发工具

### 安装开发工具

```bash
make install-tools
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 安全检查

```bash
make security
```

### 生成 Swagger 文档

```bash
make swagger
```

### 运行测试

```bash
make test
```

## 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_HOST | 服务器主机 | 0.0.0.0 |
| SERVER_PORT | 服务器端口 | 8080 |
| GIN_MODE | Gin模式 | debug |
| DB_DRIVER | 数据库驱动 | sqlite |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USERNAME | 数据库用户名 | - |
| DB_PASSWORD | 数据库密码 | - |
| DB_DATABASE | 数据库名称 | mcp_adapter.db |
| LOG_LEVEL | 日志级别 | info |

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。