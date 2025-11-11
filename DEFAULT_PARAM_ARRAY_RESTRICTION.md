# 默认参数数组必填限制

## 功能说明

在前端接口编辑页面的**默认参数** Tab 中，**数组类型的参数不能设置为必填**。

## 限制规则

### 1. 默认参数 Tab
- ✅ 可以设置默认值（仅限 `string`、`number`、`boolean` 三种基本类型）
- ❌ **如果勾选了"数组"复选框，则"必填"复选框会被自动禁用**
- 当取消勾选"数组"时，"必填"复选框恢复可用状态

### 2. 普通参数 Tab
- ❌ 不能设置默认值
- ✅ 数组类型可以设置为必填（无限制）

## 技术实现

### 前端逻辑 (`web/static/js/app.js`)

#### 1. 数组复选框变化处理
```javascript
function handleParamArrayChange(checkbox) {
    const row = checkbox.closest('.param-row');
    updateRequiredCheckboxState(row);
}
```

#### 2. 必填复选框状态更新
```javascript
function updateRequiredCheckboxState(row) {
    const arrayCheckbox = row.querySelector('.param-array-checkbox');
    const requiredCheckbox = row.querySelector('.param-required-checkbox');
    const defaultInput = row.querySelector('.param-default-input');
    
    // 只有在默认参数Tab中（存在默认值输入框）才需要此限制
    if (defaultInput && arrayCheckbox && requiredCheckbox) {
        const isArray = arrayCheckbox.checked;
        
        if (isArray) {
            // 数组类型时禁用必填
            requiredCheckbox.disabled = true;
            requiredCheckbox.checked = false;
        } else {
            // 非数组类型时启用必填
            requiredCheckbox.disabled = false;
        }
    }
}
```

#### 3. 初始化状态
在 `addParamRow` 函数中，创建参数行后会自动调用：
```javascript
if (isDefaultParam) {
    updateRequiredCheckboxState(row);
}
```

### CSS 样式 (`web/static/css/style.css`)

```css
/* 复选框禁用样式 */
.param-row input[type="checkbox"]:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.param-row label:has(input[type="checkbox"]:disabled) {
    opacity: 0.5;
    cursor: not-allowed;
}
```

## 使用场景

### 场景1: 创建新的默认参数
1. 切换到"默认参数" Tab
2. 点击"添加参数"按钮
3. 填写参数名称和类型
4. 勾选"数组"复选框 → "必填"复选框自动禁用且取消勾选
5. 取消勾选"数组" → "必填"复选框恢复可用

### 场景2: 编辑已有参数
1. 打开接口编辑页面
2. 在"默认参数" Tab 中修改已有参数
3. 勾选"数组"复选框 → 如果之前勾选了"必填"，会自动取消
4. 禁用状态会通过灰色样式提示用户

### 场景3: 数据加载
1. 从数据库加载参数时，如果参数同时满足：
   - 在默认参数 Tab 中
   - `is_array = true`
2. 则"必填"复选框会自动设置为 `disabled` 状态

## 设计原因

### 为什么数组类型的默认参数不能必填？

1. **默认值的限制**：默认参数的默认值只支持基本类型（`string`、`number`、`boolean`），不支持数组类型
2. **逻辑一致性**：既然数组类型无法设置默认值，那么将其设为必填会导致：
   - 用户必须提供参数值
   - 但无法设置默认值作为后备
   - 这违背了"默认参数"的设计初衷
3. **避免混淆**：将数组类型的默认参数设为必填会让用户困惑于"默认"和"必填"的矛盾

## 相关文档

- [参数 Tab 功能说明](PARAMS_TAB_UPDATE.md)
- [默认值类型限制](DEFAULT_VALUE_TYPE_RESTRICTION.md)
- [参数缓存优化](PARAM_CACHING_OPTIMIZATION.md)

## 更新日期

2025-11-11
