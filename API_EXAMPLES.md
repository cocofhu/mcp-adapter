# API 使用示例

本文档提供了重构后的 API 使用示例。

## 基础流程

### 1. 创建应用

```bash
curl -X POST http://localhost:8080/api/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Gateway",
    "description": "API 网关应用",
    "path": "gateway",
    "protocol": "sse",
    "enabled": true
  }'
```

响应：
```json
{
  "id": 1,
  "name": "My API Gateway",
  "description": "API 网关应用",
  "path": "gateway",
  "protocol": "sse",
  "enabled": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## 自定义类型管理

### 2. 创建简单自定义类型

```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "User",
    "description": "用户信息",
    "fields": [
      {
        "name": "id",
        "type": "number",
        "required": true,
        "description": "用户ID"
      },
      {
        "name": "username",
        "type": "string",
        "required": true,
        "description": "用户名"
      },
      {
        "name": "email",
        "type": "string",
        "required": false,
        "description": "邮箱地址"
      },
      {
        "name": "age",
        "type": "number",
        "required": false,
        "description": "年龄"
      },
      {
        "name": "is_active",
        "type": "boolean",
        "required": false,
        "description": "是否激活"
      }
    ]
  }'
```

### 3. 创建嵌套自定义类型

先创建 Address 类型：
```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "Address",
    "description": "地址信息",
    "fields": [
      {
        "name": "country",
        "type": "string",
        "required": true,
        "description": "国家"
      },
      {
        "name": "city",
        "type": "string",
        "required": true,
        "description": "城市"
      },
      {
        "name": "street",
        "type": "string",
        "required": false,
        "description": "街道"
      },
      {
        "name": "zipcode",
        "type": "string",
        "required": false,
        "description": "邮编"
      }
    ]
  }'
```

然后创建包含 Address 的 UserProfile 类型：
```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "UserProfile",
    "description": "用户详细信息",
    "fields": [
      {
        "name": "user",
        "type": "custom",
        "ref": 1,
        "required": true,
        "description": "基本用户信息"
      },
      {
        "name": "address",
        "type": "custom",
        "ref": 2,
        "required": false,
        "description": "地址信息"
      },
      {
        "name": "tags",
        "type": "string",
        "is_array": true,
        "required": false,
        "description": "标签列表"
      }
    ]
  }'
```

### 4. 获取应用的所有自定义类型

```bash
curl http://localhost:8080/api/custom-types?app_id=1
```

响应：
```json
[
  {
    "id": 1,
    "app_id": 1,
    "name": "User",
    "description": "用户信息",
    "fields": [
      {
        "id": 1,
        "custom_type_id": 1,
        "name": "id",
        "type": "number",
        "required": true,
        "description": "用户ID"
      },
      {
        "id": 2,
        "custom_type_id": 1,
        "name": "username",
        "type": "string",
        "required": true,
        "description": "用户名"
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### 5. 更新自定义类型

```bash
curl -X PUT http://localhost:8080/api/custom-types/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "UserInfo",
    "description": "用户基本信息（已更新）",
    "fields": [
      {
        "name": "id",
        "type": "number",
        "required": true,
        "description": "用户ID"
      },
      {
        "name": "username",
        "type": "string",
        "required": true,
        "description": "用户名"
      },
      {
        "name": "nickname",
        "type": "string",
        "required": false,
        "description": "昵称（新增字段）"
      }
    ]
  }'
```

---

## 接口管理

### 6. 创建简单接口（基本类型参数）

```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "GetUser",
    "description": "获取用户信息",
    "protocol": "http",
    "url": "https://api.example.com/users",
    "method": "GET",
    "auth_type": "none",
    "enabled": true,
    "parameters": [
      {
        "name": "id",
        "type": "string",
        "location": "query",
        "required": true,
        "description": "用户ID"
      },
      {
        "name": "include_details",
        "type": "boolean",
        "location": "query",
        "required": false,
        "description": "是否包含详细信息",
        "default_value": "false"
      }
    ]
  }'
```

### 7. 创建使用自定义类型的接口

```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "CreateUser",
    "description": "创建用户",
    "protocol": "http",
    "url": "https://api.example.com/users",
    "method": "POST",
    "auth_type": "none",
    "enabled": true,
    "parameters": [
      {
        "name": "user",
        "type": "custom",
        "ref": 1,
        "location": "body",
        "required": true,
        "description": "用户信息"
      },
      {
        "name": "Authorization",
        "type": "string",
        "location": "header",
        "required": true,
        "description": "认证令牌"
      }
    ]
  }'
```

### 8. 创建复杂接口（多种参数位置）

```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "UpdateUserProfile",
    "description": "更新用户资料",
    "protocol": "http",
    "url": "https://api.example.com/users/{id}/profile",
    "method": "PUT",
    "auth_type": "none",
    "enabled": true,
    "parameters": [
      {
        "name": "id",
        "type": "string",
        "location": "path",
        "required": true,
        "description": "用户ID（路径参数）"
      },
      {
        "name": "profile",
        "type": "custom",
        "ref": 3,
        "location": "body",
        "required": true,
        "description": "用户资料"
      },
      {
        "name": "Authorization",
        "type": "string",
        "location": "header",
        "required": true,
        "description": "认证令牌"
      },
      {
        "name": "version",
        "type": "string",
        "location": "query",
        "required": false,
        "description": "API版本",
        "default_value": "v1"
      }
    ]
  }'
```

### 9. 获取应用的所有接口

```bash
curl http://localhost:8080/api/interfaces?app_id=1
```

响应：
```json
[
  {
    "id": 1,
    "app_id": 1,
    "name": "GetUser",
    "description": "获取用户信息",
    "protocol": "http",
    "url": "https://api.example.com/users",
    "method": "GET",
    "auth_type": "none",
    "enabled": true,
    "parameters": [
      {
        "id": 1,
        "interface_id": 1,
        "name": "id",
        "type": "string",
        "location": "query",
        "required": true,
        "description": "用户ID"
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### 10. 获取单个接口详情

```bash
curl http://localhost:8080/api/interfaces/1
```

### 11. 更新接口

```bash
curl -X PUT http://localhost:8080/api/interfaces/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GetUserById",
    "description": "根据ID获取用户信息（已更新）",
    "url": "https://api.example.com/v2/users",
    "parameters": [
      {
        "name": "user_id",
        "type": "string",
        "location": "query",
        "required": true,
        "description": "用户ID（参数名已更改）"
      },
      {
        "name": "fields",
        "type": "string",
        "is_array": true,
        "location": "query",
        "required": false,
        "description": "需要返回的字段列表（新增）"
      }
    ]
  }'
```

### 12. 删除接口

```bash
curl -X DELETE http://localhost:8080/api/interfaces/1
```

---

## 完整示例：创建博客系统 API

### Step 1: 创建应用

```bash
curl -X POST http://localhost:8080/api/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Blog System",
    "description": "博客系统 API",
    "path": "blog",
    "protocol": "sse",
    "enabled": true
  }'
```

### Step 2: 创建自定义类型

**Author 类型**：
```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "Author",
    "description": "作者信息",
    "fields": [
      {"name": "id", "type": "number", "required": true},
      {"name": "name", "type": "string", "required": true},
      {"name": "email", "type": "string", "required": true}
    ]
  }'
```

**Article 类型**：
```bash
curl -X POST http://localhost:8080/api/custom-types \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "Article",
    "description": "文章信息",
    "fields": [
      {"name": "id", "type": "number", "required": true},
      {"name": "title", "type": "string", "required": true},
      {"name": "content", "type": "string", "required": true},
      {"name": "author", "type": "custom", "ref": 1, "required": true},
      {"name": "tags", "type": "string", "is_array": true, "required": false},
      {"name": "published", "type": "boolean", "required": false}
    ]
  }'
```

### Step 3: 创建接口

**获取文章列表**：
```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "ListArticles",
    "description": "获取文章列表",
    "protocol": "http",
    "url": "https://blog.example.com/api/articles",
    "method": "GET",
    "auth_type": "none",
    "parameters": [
      {"name": "page", "type": "number", "location": "query", "required": false, "default_value": "1"},
      {"name": "limit", "type": "number", "location": "query", "required": false, "default_value": "10"},
      {"name": "tag", "type": "string", "location": "query", "required": false}
    ]
  }'
```

**创建文章**：
```bash
curl -X POST http://localhost:8080/api/interfaces \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "name": "CreateArticle",
    "description": "创建新文章",
    "protocol": "http",
    "url": "https://blog.example.com/api/articles",
    "method": "POST",
    "auth_type": "none",
    "parameters": [
      {"name": "article", "type": "custom", "ref": 2, "location": "body", "required": true},
      {"name": "Authorization", "type": "string", "location": "header", "required": true}
    ]
  }'
```

---

## 参数类型说明

### 基本类型
- `number` - 数字类型
- `string` - 字符串类型
- `boolean` - 布尔类型

### 自定义类型
- `custom` - 引用自定义类型，需要提供 `ref` 字段指向 CustomType.ID

### 参数位置
- `query` - URL 查询参数
- `header` - HTTP 请求头
- `body` - 请求体（JSON）
- `path` - URL 路径参数（如 `/users/{id}`）

### 数组支持
设置 `is_array: true` 可以将参数定义为数组类型。

---

## 错误处理

### 常见错误响应

**应用不存在**：
```json
{
  "error": "application not found"
}
```

**类型名称重复**：
```json
{
  "error": "duplicate custom type name in this application"
}
```

**引用的类型不存在**：
```json
{
  "error": "invalid parameter reference: custom type not found"
}
```

**缺少必填参数**：
```json
{
  "error": "missing required parameter: user"
}
```

**删除被引用的类型**：
```json
{
  "error": "cannot delete custom type: referenced by other type fields"
}
```
