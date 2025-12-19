# MCP Adapter

> 🚀 将任何 HTTP API 转换为 MCP (Model Context Protocol) 工具，让 AI 助手能够调用你的 API

一个轻量级的 HTTP API 管理和适配系统，通过可视化界面配置 API，自动生成 MCP 工具定义，让 Claude Desktop 等 AI 助手能够直接调用你的 HTTP 接口。

## ✨ 核心特性

- 🎯 **零代码配置** - 通过 Web 界面配置 API，无需编写代码
- 🔌 **MCP 协议支持** - 自动将 HTTP API 转换为 MCP 工具
- 🎨 **自定义类型系统** - 类似 TypeScript，定义可复用的复杂数据结构
- 📦 **多应用管理** - 支持管理多个独立的 API 应用
- 🌐 **现代化 UI** - 响应式设计，操作简单直观

## 🚀 Quick Start

### 使用 Docker（推荐）

一键启动，无需安装任何依赖：

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --name mcp-adapter \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

Windows PowerShell:
```powershell
docker run -d -p 8080:8080 -v ${PWD}/data:/app/data --name mcp-adapter ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

启动后访问：**http://localhost:8080**

### 从源码运行

```bash
# 克隆项目
git clone https://github.com/yourusername/mcp-adapter.git
cd mcp-adapter

# 安装依赖
go mod download

# 启动服务
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## 📖 使用流程

### 1️⃣ 创建应用

在 Web 界面中创建一个新应用，例如 "天气 API"。

### 2️⃣ 定义自定义类型（可选）

如果你的 API 使用复杂的数据结构，可以先定义自定义类型。

### 3️⃣ 配置 API 接口

添加你的 HTTP API 接口配置：

- **接口名称**: GetWeather
- **URL**: https://api.weather.com/current
- **方法**: GET
- **参数**: 
  - city (string, query, 必填)
  - units (string, query, 可选)

### 4️⃣ 连接到 AI 助手

配置 Claude Desktop 或其他 MCP 客户端，连接到：
```
http://localhost:8080/mcp/your-app-path
```

现在 AI 助手就可以调用你配置的 API 了！

## 🎯 使用场景

- 🤖 **AI 助手增强** - 让 Claude 等 AI 助手能够调用你的内部 API
- 🔗 **API 聚合** - 将多个 API 统一管理和调用
- 📝 **API 文档** - 可视化管理和展示 API 定义
- 🧪 **快速原型** - 快速配置和测试 API 集成


## 🔧 配置说明

### 环境变量

- `PORT` - 服务端口（默认: 8080）
- `DB_TYPE` - 数据库类型：`sqlite` 或 `mysql`（默认: sqlite）
- `DB_PATH` - SQLite 数据库文件路径（默认: ./data/mcp-adapter.db）
- `DB_DSN` - MySQL 连接字符串（例如: `user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True`）

### 数据库支持

支持 **SQLite** 和 **MySQL** 两种数据库：

#### 🗄️ SQLite（默认）

零配置，开箱即用，适合中小规模使用：

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**特点**：
- ✅ 零配置，开箱即用
- ✅ 轻量级，适合个人和小团队
- ✅ 数据持久化，重启不丢失
- ✅ 支持完整的 SQL 功能

#### 🐬 MySQL

适合生产环境和大规模使用：

```bash
docker run -d \
  -p 8080:8080 \
  -e DB_TYPE=mysql \
  -e DB_DSN="user:password@tcp(mysql-host:3306)/mcp_adapter?charset=utf8mb4&parseTime=True" \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**特点**：
- ✅ 高性能，支持大规模并发
- ✅ 适合生产环境和集群部署
- ✅ 支持主从复制和高可用
- ✅ 更好的数据安全性和备份能力

### Docker 数据持久化

**SQLite 模式**：使用 volume 挂载保存数据

```bash
docker run -d \
  -p 8080:8080 \
  -v /your/local/path:/app/data \
  ccr.ccs.tencentyun.com/cocofhu/mcp-adapter
```

**MySQL 模式**：数据存储在 MySQL 服务器中，无需挂载本地目录

## 🛠️ 技术栈

- Go + Gin - 后端服务
- SQLite - 数据存储
- 原生 JavaScript - 前端界面
- MCP Protocol - AI 助手协议

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📝 许可证

MIT License

