# 默认值类型限制说明

## 概述

在接口参数定义中，**默认值功能仅支持基本数据类型（number、string、boolean）**。自定义类型参数不支持设置默认值。

## 支持的类型

### ✅ 支持设置默认值的类型

#### 1. String 类型
```json
{
  "name": "username",
  "type": "string",
  "location": "query",
  "default_value": "guest"
}
```

**适用场景**：
- 用户名、邮箱等文本字段
- 搜索关键词
- 过滤条件

**示例默认值**：
- `"guest"` - 默认用户名
- `""` - 空字符串
- `"default"` - 默认标签

---

#### 2. Number 类型
```json
{
  "name": "page",
  "type": "number",
  "location": "query",
  "default_value": "1"
}
```

**适用场景**：
- 分页参数（页码、每页数量）
- 数值限制（最大值、最小值）
- 计数器

**示例默认值**：
- `"1"` - 第一页
- `"10"` - 每页10条
- `"0"` - 初始计数

**注意**：后端会将字符串转换为数值类型

---

#### 3. Boolean 类型
```json
{
  "name": "enabled",
  "type": "boolean",
  "location": "query",
  "default_value": "true"
}
```

**适用场景**：
- 开关标志
- 启用/禁用状态
- 过滤条件

**示例默认值**：
- `"true"` - 启用
- `"false"` - 禁用

**注意**：后端会将字符串 "true"/"false" 转换为布尔值

---

### ❌ 不支持设置默认值的类型

#### Custom（自定义类型）

自定义类型是复杂对象结构，不支持通过简单字符串设置默认值。

```json
{
  "name": "user",
  "type": "custom",
  "ref": 1,
  "location": "body"
  // ❌ 不能设置 default_value
}
```

**原因**：
1. **复杂性**：自定义类型包含多个字段，无法用单一字符串表示
2. **验证困难**：需要完整的 JSON Schema 验证，容易出错
3. **用户体验**：避免用户输入错误格式的复杂对象
4. **语义不清**：默认值的含义对复杂对象不明确

**替代方案**：
- 在后端代码中处理自定义类型的默认值
- 使用必填字段验证确保数据完整性
- 在接口文档中说明预期的对象结构

---

## 前端行为

### 默认参数 Tab 中

1. **选择基本类型（string/number/boolean）**
   ```
   参数名: [username_____] 类型: [string ▼] 位置: [query ▼] □数组 ☑必填 [×]
   默认值: [guest_______] 描述: [默认用户名_______________]
   ```
   ✅ 显示默认值输入框

2. **选择自定义类型**
   ```
   参数名: [user________] 类型: [User ▼] 位置: [body ▼] □数组 ☑必填 [×]
   描述: [用户信息对象_______________]
   ```
   ❌ 默认值输入框自动隐藏

3. **类型切换行为**
   - 从基本类型 → 自定义类型：默认值输入框隐藏，已输入的值被清空
   - 从自定义类型 → 基本类型：默认值输入框显示，可以重新输入

### 普通参数 Tab 中

无论选择什么类型，都不显示默认值输入框。

---

## 使用示例

### 示例 1：分页查询接口

**接口信息**：
- 名称：GetUserList
- 方法：GET
- URL：/api/users

**默认参数**（在"默认参数"Tab中）：
```
参数1: page     | number | query | 默认值: 1  | 描述: 页码
参数2: pageSize | number | query | 默认值: 10 | 描述: 每页数量
参数3: keyword  | string | query | 默认值: "" | 描述: 搜索关键词
```

**普通参数**（在"普通参数"Tab中）：
```
参数4: userId   | string | query | 描述: 用户ID（必填）
```

**生成的 MCP Schema**：
```json
{
  "page": {
    "type": "number",
    "description": "页码",
    "default": 1
  },
  "pageSize": {
    "type": "number",
    "description": "每页数量",
    "default": 10
  },
  "keyword": {
    "type": "string",
    "description": "搜索关键词",
    "default": ""
  },
  "userId": {
    "type": "string",
    "description": "用户ID（必填）"
  }
}
```

---

### 示例 2：创建用户接口

**接口信息**：
- 名称：CreateUser
- 方法：POST
- URL：/api/users

**默认参数**（在"默认参数"Tab中）：
```
参数1: enabled  | boolean | body | 默认值: true | 描述: 是否启用
参数2: role     | string  | body | 默认值: user | 描述: 用户角色
```

**普通参数**（在"普通参数"Tab中）：
```
参数3: name     | string  | body | 描述: 用户名（必填）
参数4: email    | string  | body | 描述: 邮箱（必填）
参数5: profile  | User    | body | 描述: 用户详细信息（自定义类型，不能有默认值）
```

**关键点**：
- ✅ `enabled` 和 `role` 有默认值，在"默认参数"Tab
- ❌ `profile` 是自定义类型，即使在"默认参数"Tab也不能设置默认值

---

## 常见问题

### Q1: 为什么自定义类型不能设置默认值？

**A**: 自定义类型是复杂对象，包含多个字段。如果要设置默认值，需要输入完整的 JSON 对象，这样会：
- 增加用户输入错误的风险
- 需要复杂的格式验证
- 难以维护和理解
- 与类型定义耦合过紧

建议在后端代码中处理复杂对象的默认值。

### Q2: 数组类型可以设置默认值吗？

**A**: 虽然 UI 上会显示默认值输入框（如果是基本类型数组），但建议谨慎使用：
- `string[]` 类型的默认值应该如何表示？是空数组 `[]` 还是包含默认字符串的数组 `["default"]`？
- 后端需要正确解析数组格式的默认值

**建议**：数组参数最好不设置默认值，或在后端处理。

### Q3: 如果我需要为自定义类型设置默认值怎么办？

**A**: 有以下几种方案：
1. **后端处理**：在后端代码中检查参数，如果未提供则使用默认对象
2. **拆分字段**：将自定义类型的字段拆分为多个基本类型参数，分别设置默认值
3. **文档说明**：在接口文档中说明如果不传该参数的默认行为

### Q4: 切换参数类型时，为什么默认值会被清空？

**A**: 这是为了数据一致性：
- 从基本类型切换到自定义类型时，之前输入的默认值对新类型可能无效
- 自动清空避免保存无效数据
- 用户如果需要，可以在切换回基本类型后重新输入

---

## 技术实现细节

### 前端验证
```javascript
// 判断是否为基本类型
const isBasicType = ['string', 'number', 'boolean'].includes(paramType);

// 只有基本类型才保存默认值
if (allowDefaultValue && isBasicType && paramDefaultValue) {
    param.default_value = paramDefaultValue;
}
```

### 动态 UI 控制
```javascript
// 类型改变时，动态显示/隐藏默认值输入框
function handleParamTypeChange(selectElement) {
    const isBasicType = ['string', 'number', 'boolean'].includes(selectedOption.value);
    
    if (isBasicType) {
        defaultInput.style.display = '';  // 显示
    } else {
        defaultInput.style.display = 'none';  // 隐藏
        defaultInput.value = '';  // 清空
    }
}
```

---

## 总结

| 特性 | 基本类型 | 自定义类型 |
|-----|---------|-----------|
| 可设置默认值 | ✅ 是 | ❌ 否 |
| 默认值输入框 | ✅ 显示 | ❌ 隐藏 |
| 默认值格式 | 简单字符串 | - |
| 验证复杂度 | 简单 | 复杂 |
| 推荐使用场景 | 分页、开关、简单配置 | 复杂对象、嵌套结构 |

**最佳实践**：
- ✅ 为分页、过滤等参数设置合理的默认值
- ✅ 使用基本类型参数承载简单配置
- ❌ 避免为复杂对象设置默认值
- ❌ 避免在前端处理复杂的默认值逻辑
