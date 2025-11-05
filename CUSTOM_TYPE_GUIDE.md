# 自定义类型使用指南

## 功能概述

MCP Adapter 支持强大的自定义类型系统，允许您：
1. **定义复杂类型**：创建包含多个字段的自定义类型
2. **类型引用**：字段可以引用其他自定义类型，构建复杂的数据结构
3. **数组支持**：任何类型都可以声明为数组类型

## 基本类型

系统内置三种基本类型：
- `string` - 字符串
- `number` - 数字
- `boolean` - 布尔值

## 创建自定义类型

### 示例 1：简单用户类型

```json
{
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
      "name": "active",
      "type": "boolean",
      "required": true,
      "description": "是否激活"
    }
  ]
}
```

### 示例 2：地址类型

```json
{
  "name": "Address",
  "description": "地址信息",
  "fields": [
    {
      "name": "street",
      "type": "string",
      "required": true,
      "description": "街道地址"
    },
    {
      "name": "city",
      "type": "string",
      "required": true,
      "description": "城市"
    },
    {
      "name": "zipCode",
      "type": "string",
      "required": false,
      "description": "邮政编码"
    }
  ]
}
```

## 类型引用

创建了基础类型后，可以在其他类型中引用它们。

### 示例 3：引用其他类型

```json
{
  "name": "UserProfile",
  "description": "用户详细信息",
  "fields": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,  // 引用 User 类型的 ID
      "required": true,
      "description": "用户基本信息"
    },
    {
      "name": "address",
      "type": "custom",
      "ref": 2,  // 引用 Address 类型的 ID
      "required": false,
      "description": "用户地址"
    },
    {
      "name": "bio",
      "type": "string",
      "required": false,
      "description": "个人简介"
    }
  ]
}
```

## 数组类型

任何字段都可以声明为数组类型，通过设置 `is_array: true`。

### 示例 4：包含数组的类型

```json
{
  "name": "Company",
  "description": "公司信息",
  "fields": [
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "公司名称"
    },
    {
      "name": "employees",
      "type": "custom",
      "ref": 1,  // 引用 User 类型
      "is_array": true,  // 声明为数组
      "required": false,
      "description": "员工列表"
    },
    {
      "name": "tags",
      "type": "string",
      "is_array": true,  // 字符串数组
      "required": false,
      "description": "标签列表"
    }
  ]
}
```

在前端显示时，数组类型会显示为 `User[]` 或 `string[]`。

## 在接口中使用自定义类型

创建接口时，参数也可以使用自定义类型。

### 示例 5：接口参数使用自定义类型

```json
{
  "name": "CreateUser",
  "method": "POST",
  "url": "https://api.example.com/users",
  "parameters": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,  // 引用 User 类型
      "location": "body",
      "required": true
    },
    {
      "name": "addresses",
      "type": "custom",
      "ref": 2,  // 引用 Address 类型
      "is_array": true,  // 地址数组
      "location": "body",
      "required": false
    }
  ]
}
```

## 前端操作指南

### 创建自定义类型

1. 选择应用
2. 切换到"自定义类型"标签
3. 点击"创建类型"按钮
4. 填写类型名称和描述
5. 添加字段：
   - **字段名**：字段的名称
   - **类型**：选择基本类型或已创建的自定义类型
   - **数组**：勾选此项使字段成为数组类型
   - **必填**：勾选此项使字段成为必填
   - **描述**：字段的说明文字

### 编辑自定义类型

1. 在类型列表中点击"编辑"按钮
2. 修改类型信息或字段
3. 点击"确认"保存

**注意**：
- 编辑类型时，不能引用自己（避免循环引用）
- 字段列表会完全替换，请确保包含所有需要的字段

### 删除自定义类型

删除类型前，系统会检查：
- 是否被其他类型的字段引用
- 是否被接口参数引用

如果存在引用，删除会失败，需要先删除引用。

## API 示例

### 创建自定义类型

```bash
POST /api/custom-types
Content-Type: application/json

{
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
      "name": "tags",
      "type": "string",
      "is_array": true,
      "required": false,
      "description": "标签列表"
    }
  ]
}
```

### 创建引用其他类型的类型

```bash
POST /api/custom-types
Content-Type: application/json

{
  "app_id": 1,
  "name": "UserProfile",
  "description": "用户详细信息",
  "fields": [
    {
      "name": "user",
      "type": "custom",
      "ref": 1,
      "required": true,
      "description": "用户基本信息"
    },
    {
      "name": "friends",
      "type": "custom",
      "ref": 1,
      "is_array": true,
      "required": false,
      "description": "好友列表"
    }
  ]
}
```

### 创建使用自定义类型的接口

```bash
POST /api/interfaces
Content-Type: application/json

{
  "app_id": 1,
  "name": "GetUserProfile",
  "method": "GET",
  "url": "https://api.example.com/users/{id}/profile",
  "parameters": [
    {
      "name": "id",
      "type": "number",
      "location": "path",
      "required": true
    },
    {
      "name": "include",
      "type": "string",
      "is_array": true,
      "location": "query",
      "required": false
    }
  ]
}
```

## 最佳实践

1. **命名规范**
   - 类型名使用 PascalCase（如 `UserProfile`）
   - 字段名使用 camelCase（如 `firstName`）

2. **类型设计**
   - 保持类型单一职责
   - 避免过深的嵌套（建议不超过 3 层）
   - 合理使用类型引用，提高复用性

3. **数组使用**
   - 明确数组元素的类型
   - 考虑是否需要限制数组长度（在描述中说明）

4. **引用管理**
   - 先创建基础类型，再创建引用它们的类型
   - 删除类型前检查引用关系
   - 避免循环引用

## 类型系统的优势

1. **类型安全**：明确定义数据结构，减少错误
2. **代码复用**：一次定义，多处使用
3. **文档化**：类型定义即文档
4. **可维护性**：集中管理类型定义，便于修改
5. **工具支持**：为 MCP 工具提供准确的类型信息

## 常见问题

### Q: 可以创建循环引用吗？
A: 不建议。虽然系统不会阻止，但可能导致序列化问题。编辑类型时会自动排除自引用。

### Q: 数组可以嵌套吗？
A: 目前不支持多维数组（如 `string[][]`）。如需复杂结构，请创建新的自定义类型。

### Q: 如何表示可选字段？
A: 将 `required` 设置为 `false`。

### Q: 类型可以跨应用使用吗？
A: 不可以。每个类型都属于特定应用，只能在该应用内使用。

### Q: 如何查看类型被哪些地方引用？
A: 目前需要手动检查。删除类型时系统会提示是否存在引用。

## 更新日志

- **v1.0** - 初始版本，支持基本类型、自定义类型、类型引用和数组
