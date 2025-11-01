# MCP Adapter - 后端 CRUD API

基于 Go + GORM + Gorilla Mux 实现的完整后端 CRUD 功能。

## 功能特性

- ✅ 应用管理 (Application CRUD)
- ✅ 接口管理 (Interface CRUD)  
- ✅ SQLite 数据库存储
- ✅ RESTful API 设计
- ✅ CORS 支持
- ✅ 前后端分离架构

## 项目结构

```
mcp-adapter/
├── backend/
│   ├── database/
│   │   └── database.go          # 数据库配置和初始化
│   ├── handlers/
│   │   ├── application.go       # 应用 CRUD 处理器
│   │   └── interface.go         # 接口 CRUD 处理器
│   ├── models/
│   │   └── models.go           # 数据模型定义
│   ├── routes/
│   │   └── routes.go           # 路由配置
│   └── main.go                 # 主服务器文件
├── index.html                  # 前端页面
├── script-api.js              # 前端 API 交互逻辑
├── style.css                  # 样式文件
├── go.mod                     # Go 模块配置
└── run.bat                    # 启动脚本
```

## API 接口

### 应用管理 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/applications` | 创建应用 |
| GET | `/api/applications` | 获取所有应用 |
| GET | `/api/applications/{id}` | 获取单个应用 |
| PUT | `/api/applications/{id}` | 更新应用 |
| DELETE | `/api/applications/{id}` | 删除应用 |

### 接口管理 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/interfaces` | 创建接口 |
| GET | `/api/interfaces` | 获取所有接口 |
| GET | `/api/interfaces?app_id={id}` | 获取指定应用的接口 |
| GET | `/api/interfaces/{id}` | 获取单个接口 |
| PUT | `/api/interfaces/{id}` | 更新接口 |
| DELETE | `/api/interfaces/{id}` | 删除接口 |

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务器

**方式一：使用启动脚本**
```bash
# Windows
run.bat
```

**方式二：手动启动**
```bash
# 设置环境变量以使用纯 Go SQLite 驱动
set CGO_ENABLED=0
go run ./backend/main.go
```

**方式三：编译后运行**
```bash
set CGO_ENABLED=0
go build -o mcp-adapter.exe ./backend
./mcp-adapter.exe
```

### 3. 访问应用

打开浏览器访问：`http://localhost:8080`

服务器将在 8080 端口启动，同时提供：
- API 服务：`http://localhost:8080/api/*`
- 静态文件服务：`http://localhost:8080/`

### 4. 测试 API

运行测试脚本验证 API 功能：
```powershell
# 确保服务器正在运行，然后执行
./test-api.ps1
```

## 数据模型

### Application (应用)

```go
type Application struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`        // 应用名称
    Description string    `json:"description"` // 应用描述
    Path        string    `json:"path"`        // 应用路径标识
    Protocol    string    `json:"protocol"`    // 应用协议
    PostProcess string    `json:"post_process"` // 后处理脚本
    Environment string    `json:"environment"` // 环境变量
    Enabled     bool      `json:"enabled"`     // 是否启用
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Interface (接口)

```go
type Interface struct {
    ID          int64     `json:"id"`
    AppID       int64     `json:"app_id"`      // 所属应用ID
    Name        string    `json:"name"`        // 接口名称
    Description string    `json:"description"` // 接口描述
    Protocol    string    `json:"protocol"`    // 接口协议
    URL         string    `json:"url"`         // 接口地址
    AuthType    string    `json:"auth_type"`   // 鉴权类型
    Enabled     bool      `json:"enabled"`     // 是否启用
    PostProcess string    `json:"post_process"` // 后处理脚本
    Options     string    `json:"options"`     // 配置选项(JSON)
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## 技术栈

- **后端**: Go 1.25
- **数据库**: SQLite (GORM)
- **路由**: Gorilla Mux
- **前端**: HTML5 + CSS3 + JavaScript (ES6+)
- **UI**: Font Awesome 图标

## 开发说明

1. 数据库文件 `mcp_adapter.db` 会在首次运行时自动创建
2. 支持自动数据库迁移，无需手动创建表结构
3. 前端通过 Fetch API 与后端交互
4. 支持 CORS，可用于前后端分离开发
5. 所有 API 返回 JSON 格式数据

## 注意事项

- 确保 Go 版本 >= 1.25
- 首次运行需要网络连接下载依赖
- SQLite 数据库文件会保存在项目根目录
- 删除应用时会级联删除其下所有接口