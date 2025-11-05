# 数据库迁移指南

## 重构说明

本次重构将 `Interface` 从使用 `Options` 字符串字段改为使用 `InterfaceParameter` 关联表来定义参数。

## 主要变更

### 1. 模型变更

**Interface 模型**：
- ❌ 移除：`Options string` (JSON 字符串)
- ✅ 新增：`Method string` (HTTP 方法)
- ✅ 关联：通过 `InterfaceParameter` 表管理参数

**新增模型**：
- `CustomType` - 自定义类型定义
- `CustomTypeField` - 自定义类型字段
- `InterfaceParameter` - 接口参数

### 2. 数据迁移步骤

#### 方式一：删除旧数据（推荐用于开发环境）

```bash
# 删除数据库文件（如果使用 SQLite）
rm mcp-adapter.db

# 重新运行程序，GORM 会自动创建新表结构
go run main.go
```

#### 方式二：手动迁移（生产环境）

如果你有重要的现有数据需要保留，需要编写迁移脚本：

```sql
-- 1. 备份现有数据
CREATE TABLE interfaces_backup AS SELECT * FROM interfaces;

-- 2. 添加新字段
ALTER TABLE interfaces ADD COLUMN method VARCHAR(50);

-- 3. 从 options JSON 中提取 method 并更新
-- 注意：这需要根据你的数据库类型（MySQL/PostgreSQL/SQLite）调整
-- SQLite 示例：
UPDATE interfaces 
SET method = json_extract(options, '$.method')
WHERE options IS NOT NULL AND options != '';

-- 4. 创建新表
CREATE TABLE custom_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE custom_type_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    custom_type_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    ref INTEGER,
    is_array BOOLEAN DEFAULT 0,
    required BOOLEAN DEFAULT 0,
    description TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE interface_parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    interface_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    ref INTEGER,
    location VARCHAR(50) NOT NULL,
    is_array BOOLEAN DEFAULT 0,
    required BOOLEAN DEFAULT 0,
    description TEXT,
    default_value TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

-- 5. 迁移参数数据（需要自定义脚本解析 JSON）
-- 这部分建议使用 Go 程序来完成

-- 6. 删除旧字段（确认迁移成功后）
-- ALTER TABLE interfaces DROP COLUMN options;
```

#### 方式三：使用 GORM AutoMigrate（推荐）

在 `database/database.go` 中添加迁移代码：

```go
// 自动迁移
db.AutoMigrate(
    &models.Application{},
    &models.Interface{},
    &models.CustomType{},
    &models.CustomTypeField{},
    &models.InterfaceParameter{},
)
```

GORM 会自动：
- 创建缺失的表
- 添加缺失的列
- 添加缺失的索引
- **不会删除未使用的列**（需要手动删除 `options` 列）

### 3. 手动清理（可选）

如果需要删除旧的 `options` 列：

```sql
-- SQLite 不支持直接 DROP COLUMN，需要重建表
-- 其他数据库：
ALTER TABLE interfaces DROP COLUMN options;
```

## API 变更

### 旧的接口创建请求

```json
{
  "app_id": 1,
  "name": "GetUser",
  "protocol": "http",
  "url": "https://api.example.com/users",
  "auth_type": "none",
  "options": {
    "method": "GET",
    "parameters": [
      {
        "name": "id",
        "type": "string",
        "required": true,
        "location": "query",
        "description": "User ID"
      }
    ]
  }
}
```

### 新的接口创建请求

```json
{
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
      "required": true,
      "location": "query",
      "description": "User ID"
    }
  ]
}
```

## 新增功能

### 1. 自定义类型管理

```bash
# 创建自定义类型
POST /api/custom-types
{
  "app_id": 1,
  "name": "User",
  "description": "用户信息",
  "fields": [
    {"name": "id", "type": "number", "required": true},
    {"name": "name", "type": "string", "required": true},
    {"name": "email", "type": "string", "required": false}
  ]
}

# 获取应用的所有自定义类型
GET /api/custom-types?app_id=1

# 获取单个自定义类型
GET /api/custom-types/1

# 更新自定义类型
PUT /api/custom-types/1

# 删除自定义类型
DELETE /api/custom-types/1
```

### 2. 接口参数支持自定义类型

```json
{
  "app_id": 1,
  "name": "CreateUser",
  "method": "POST",
  "url": "https://api.example.com/users",
  "parameters": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,
      "location": "body",
      "required": true,
      "description": "用户信息"
    }
  ]
}
```

## 验证迁移

运行以下测试确保迁移成功：

```bash
# 1. 启动服务
go run main.go

# 2. 创建应用
curl -X POST http://localhost:8080/api/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test App",
    "path": "test",
    "protocol": "sse"
  }'

# 3. 创建自定义类型
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "User",
    "fields": [
      {"name": "id", "type": "number", "required": true},
      {"name": "name", "type": "string", "required": true}
    ]
  }'

# 4. 创建接口
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
        "required": true,
        "location": "query"
      }
    ]
  }'

# 5. 验证接口列表
curl http://localhost:8080/api/interfaces?app_id=1
```

## 回滚方案

如果迁移出现问题，可以：

1. 恢复数据库备份
2. 回退代码到重构前的版本
3. 重新启动服务

```bash
# 恢复备份（SQLite 示例）
cp mcp-adapter.db.backup mcp-adapter.db

# 回退代码
git checkout <previous-commit>

# 重启服务
go run main.go
```
