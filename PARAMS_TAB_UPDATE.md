# 参数定义 Tab 功能更新

## 更新概述

在接口编辑页面的参数定义部分增加了 Tab 页面功能，将参数分为"默认参数"和"普通参数"两类，只有默认参数才能设置参数默认值。

## 主要改动

### 1. 前端界面更新 (`web/static/index.html`)

- 将原有的单一参数容器改为带 Tab 的双容器结构
- 添加"默认参数"和"普通参数"两个 Tab 页签
- 每个 Tab 对应独立的参数容器：
  - `default-params-container` - 默认参数容器
  - `regular-params-container` - 普通参数容器

### 2. 样式更新 (`web/static/css/style.css`)

新增以下样式类：
- `.params-tabs` - Tab 容器样式
- `.params-tab-header` - Tab 头部样式
- `.params-tab-btn` - Tab 按钮样式（包含激活状态）
- `.params-tab-content` - Tab 内容区域样式
- `.params-tab-pane` - 单个 Tab 面板样式

### 3. JavaScript 逻辑更新 (`web/static/js/app.js`)

#### 新增变量
```javascript
let currentParamsTab = 'default'; // 跟踪当前激活的 Tab
```

#### 新增函数

**`switchParamsTab(tab)`**
- 切换参数 Tab 的显示
- 更新 Tab 按钮的激活状态
- 控制参数容器的显示/隐藏

**`collectParamFromRow(row, allowDefaultValue)`**
- 从参数行收集数据的辅助函数
- 只有在 `allowDefaultValue=true` 时才保存默认值
- 用于区分默认参数和普通参数

#### 修改函数

**`showInterfaceForm(interfaceId)`**
- 初始化时重置 Tab 为"默认参数"
- 清空两个参数容器
- 加载参数时根据是否有默认值分配到不同 Tab
  - 有默认值 → 默认参数 Tab
  - 无默认值 → 普通参数 Tab

**`addParamRow(paramData, isDefaultParam)`**
- 新增 `isDefaultParam` 参数，指定参数添加到哪个 Tab
- 只有在默认参数 Tab 中才显示"默认值"输入框
- 普通参数 Tab 只显示参数描述输入框

**接口表单提交逻辑**
- 分别从两个容器收集参数
- 默认参数：调用 `collectParamFromRow(row, true)` - 可保存默认值
- 普通参数：调用 `collectParamFromRow(row, false)` - 不保存默认值

## 使用说明

### 创建/编辑接口时

1. **添加默认参数**：
   - 切换到"默认参数" Tab
   - 点击"添加参数"按钮
   - 填写参数信息，可以设置默认值

2. **添加普通参数**：
   - 切换到"普通参数" Tab
   - 点击"添加参数"按钮
   - 填写参数信息（不显示默认值输入框）

3. **编辑已有接口**：
   - 系统自动根据参数是否有默认值分配到对应 Tab
   - 有默认值的参数 → 默认参数 Tab
   - 无默认值的参数 → 普通参数 Tab

### 特性说明

- **Tab 切换**：点击 Tab 按钮即可切换，当前激活的 Tab 高亮显示
- **独立管理**：两个 Tab 的参数互相独立，可以分别添加和删除
- **自动分类**：编辑接口时，系统根据参数的默认值自动分配到对应 Tab
- **约束条件**：只有默认参数 Tab 中的参数才能设置默认值，普通参数不显示默认值输入框

## 后端兼容性

此更新为纯前端改动，不需要修改后端代码：
- 后端 API 保持不变
- 数据模型保持不变（`InterfaceParameter.DefaultValue` 字段已存在）
- 只是前端限制了哪些参数可以设置默认值

## 测试建议

1. 创建新接口，分别在两个 Tab 中添加参数
2. 保存后刷新页面，验证参数是否正确分配到对应 Tab
3. 编辑接口，修改参数的默认值，保存后验证
4. 确认普通参数 Tab 中不显示默认值输入框
5. 确认只有默认参数 Tab 中的默认值能被保存

## 浏览器兼容性

- 支持所有现代浏览器（Chrome, Firefox, Safari, Edge）
- 使用 CSS Flexbox 布局，IE11 可能需要 polyfill
