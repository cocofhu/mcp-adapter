// MCP Adapter 前端应用
// API 基础 URL
const API_BASE = '/api';

// 全局状态
const state = {
    currentApp: null,
    applications: [],
    customTypes: [],
    interfaces: []
};

// ========== 工具函数 ==========

// HTML 转义函数，防止 XSS
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 日期格式化函数
function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) return '今天';
    if (diffDays === 1) return '昨天';
    if (diffDays < 7) return `${diffDays}天前`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)}周前`;
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}

// 显示 Toast 通知
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
        <span>${message}</span>
    `;
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease-in';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// API 请求封装
async function apiRequest(url, options = {}) {
    try {
        const response = await fetch(`${API_BASE}${url}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || `HTTP ${response.status}`);
        }
        
        if (response.status === 204) {
            return null;
        }
        
        return await response.json();
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}

// ========== 模态框管理 ==========

function showModal(title, bodyHTML, onConfirm, showFooter = true) {
    const modal = document.getElementById('modal');
    const modalTitle = document.getElementById('modal-title');
    const modalBody = document.getElementById('modal-body');
    const modalFooter = document.getElementById('modal-footer');
    const confirmBtn = document.getElementById('modal-confirm');
    
    modalTitle.textContent = title;
    modalBody.innerHTML = bodyHTML;
    modalFooter.style.display = showFooter ? 'flex' : 'none';
    
    modal.classList.add('active');
    
    // 移除旧的事件监听器
    const newConfirmBtn = confirmBtn.cloneNode(true);
    confirmBtn.parentNode.replaceChild(newConfirmBtn, confirmBtn);
    
    if (onConfirm) {
        newConfirmBtn.addEventListener('click', () => {
            onConfirm();
            closeModal();
        });
    }
}

function closeModal() {
    const modal = document.getElementById('modal');
    modal.classList.remove('active');
}

// 关闭模态框事件
document.getElementById('modal-close').addEventListener('click', closeModal);
document.getElementById('modal-cancel').addEventListener('click', closeModal);
document.getElementById('modal').addEventListener('click', (e) => {
    if (e.target.id === 'modal') closeModal();
});

// ========== 标签页切换 ==========

document.querySelectorAll('.nav-item').forEach(item => {
    item.addEventListener('click', () => {
        const tab = item.dataset.tab;
        
        // 更新导航激活状态
        document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
        item.classList.add('active');
        
        // 更新内容显示
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tab}-tab`).classList.add('active');
        
        // 加载对应数据
        if (tab === 'applications') {
            loadApplications();
        } else if (tab === 'custom-types') {
            loadCustomTypes();
        } else if (tab === 'interfaces') {
            loadInterfaces();
        }
    });
});

// ========== 应用管理 ==========

async function loadApplications() {
    try {
        const apps = await apiRequest('/applications');
        state.applications = apps;
        renderApplications(apps);
        updateAppSelectors(apps);
    } catch (error) {
        showToast('加载应用失败: ' + error.message, 'error');
    }
}

function renderApplications(apps) {
    const grid = document.getElementById('applications-grid');
    
    // 获取搜索关键词
    const searchTerm = document.getElementById('app-search')?.value.toLowerCase() || '';
    
    // 过滤应用
    const filteredApps = apps.filter(app => {
        if (!searchTerm) return true;
        return app.name.toLowerCase().includes(searchTerm) ||
               (app.description && app.description.toLowerCase().includes(searchTerm)) ||
               (app.path && app.path.toLowerCase().includes(searchTerm));
    });
    
    if (!filteredApps || filteredApps.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-folder-open"></i>
                <h3>${searchTerm ? '未找到匹配的应用' : '还没有应用'}</h3>
                <p>${searchTerm ? '尝试其他搜索关键词' : '点击"创建应用"开始'}</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = filteredApps.map(app => {
        // 处理描述文本，限制长度
        const description = app.description || '暂无描述';
        const displayDesc = description.length > 120 ? description.substring(0, 120) + '...' : description;
        
        return `
        <div class="card" onclick="viewApplication(${app.id})">
            <div class="card-header">
                <div style="flex: 1; min-width: 0;">
                    <div class="card-title" title="${escapeHtml(app.name)}">${escapeHtml(app.name)}</div>
                    <div class="card-subtitle">路径: /${app.path || ''}</div>
                </div>
                <span class="card-badge badge-${app.enabled ? 'success' : 'warning'}">
                    ${app.enabled ? '已启用' : '已禁用'}
                </span>
            </div>
            <div class="card-body">
                <div class="card-subtitle" title="${escapeHtml(description)}">${escapeHtml(displayDesc)}</div>
                <div class="card-meta">
                    <div class="card-meta-item">
                        <i class="fas fa-network-wired"></i>
                        ${app.protocol?.toUpperCase() || 'HTTP'}
                    </div>
                    <div class="card-meta-item">
                        <i class="fas fa-calendar"></i>
                        ${formatDate(app.created_at)}
                    </div>
                </div>
            </div>
            <div class="card-footer">
                <button class="btn btn-sm btn-primary" onclick="event.stopPropagation(); viewApplicationEndpoint(${app.id})" title="查看 MCP 接入链接">
                    <i class="fas fa-link"></i> 接入
                </button>
                <button class="btn btn-sm btn-info" onclick="event.stopPropagation(); manageApplicationInterfaces(${app.id})" title="管理此应用的接口">
                    <i class="fas fa-plug"></i> 接口
                </button>
                <button class="btn btn-sm btn-success" onclick="event.stopPropagation(); viewApplicationJSON(${app.id})" title="查看应用 JSON 数据">
                    <i class="fas fa-code"></i> JSON
                </button>
                <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editApplication(${app.id})" title="编辑应用信息">
                    <i class="fas fa-edit"></i> 编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteApplication(${app.id})" title="删除此应用">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
        `;
    }).join('');
}

// 应用搜索事件监听
document.getElementById('app-search')?.addEventListener('input', () => {
    renderApplications(state.applications);
});

function updateAppSelectors(apps) {
    const selectors = [document.getElementById('global-app-select')];
    
    selectors.forEach(select => {
        if (!select) return;
        select.innerHTML = '<option value="">请选择应用</option>' +
            apps.map(app => `<option value="${app.id}">${app.name}</option>`).join('');
        
        if (state.currentApp) {
            select.value = state.currentApp.id;
        }
    });
}

// 应用表单状态
let currentAppId = null;

// 显示应用表单
function showAppForm(appId = null) {
    currentAppId = appId;
    document.getElementById('app-list-view').style.display = 'none';
    document.getElementById('app-form-view').style.display = 'block';
    
    if (appId) {
        // 编辑模式
        const app = state.applications.find(a => a.id === appId);
        if (app) {
            document.getElementById('app-form-title').textContent = '编辑应用';
            document.getElementById('app-form-subtitle').textContent = '修改应用信息';
            document.getElementById('app-name').value = app.name;
            document.getElementById('app-description').value = app.description || '';
            document.getElementById('app-path').value = app.path;
            document.getElementById('app-protocol').value = app.protocol;
        }
    } else {
        // 创建模式
        document.getElementById('app-form-title').textContent = '创建应用';
        document.getElementById('app-form-subtitle').textContent = '填写应用信息';
        document.getElementById('app-name').value = '';
        document.getElementById('app-description').value = '';
        document.getElementById('app-path').value = '';
        document.getElementById('app-protocol').value = 'sse';
    }
}

// 隐藏应用表单
function hideAppForm() {
    document.getElementById('app-list-view').style.display = 'block';
    document.getElementById('app-form-view').style.display = 'none';
    currentAppId = null;
}

// 创建应用按钮
document.getElementById('create-app-btn').addEventListener('click', () => {
    showAppForm();
});

// 应用表单返回按钮
document.getElementById('app-form-back').addEventListener('click', hideAppForm);
document.getElementById('app-form-cancel').addEventListener('click', hideAppForm);

// 应用表单提交
document.getElementById('app-form-submit').addEventListener('click', async () => {
    const name = document.getElementById('app-name').value;
    const description = document.getElementById('app-description').value;
    const path = document.getElementById('app-path').value;
    const protocol = document.getElementById('app-protocol').value;
    
    if (!name || !path) {
        showToast('请填写必填字段', 'error');
        return;
    }
    
    try {
        if (currentAppId) {
            // 更新
            await apiRequest(`/applications/${currentAppId}`, {
                method: 'PUT',
                body: JSON.stringify({ name, description, path, protocol })
            });
            showToast('应用更新成功', 'success');
        } else {
            // 创建
            await apiRequest('/applications', {
                method: 'POST',
                body: JSON.stringify({ name, description, path, protocol, enabled: true })
            });
            showToast('应用创建成功', 'success');
        }
        hideAppForm();
        loadApplications();
    } catch (error) {
        showToast('操作失败: ' + error.message, 'error');
    }
});

function viewApplication(id) {
    const app = state.applications.find(a => a.id === id);
    if (!app) return;
    
    showModal(app.name, `
        <div class="doc-section">
            <h3>基本信息</h3>
            <p><strong>路径:</strong> ${app.path}</p>
            <p><strong>协议:</strong> ${app.protocol}</p>
            <p><strong>状态:</strong> ${app.enabled ? '已启用' : '已禁用'}</p>
            <p><strong>描述:</strong> ${app.description || '无'}</p>
        </div>
    `, null, false);
}

function viewApplicationEndpoint(id) {
    const app = state.applications.find(a => a.id === id);
    if (!app) return;
    
    const protocol = app.protocol || 'sse';
    const path = app.path;
    const baseUrl = window.location.origin;
    
    let endpointInfo = '';
    if (protocol === 'sse') {
        const sseEndpoint = `${baseUrl}/sse/${path}`;
        const messageEndpoint = `${baseUrl}/message/${path}`;
        endpointInfo = `
            <div class="doc-section">
                <h3>SSE 端点信息</h3>
                <p><strong>SSE 端点:</strong></p>
                <pre><code>${sseEndpoint}</code></pre>
                <p><strong>消息端点:</strong></p>
                <pre><code>${messageEndpoint}</code></pre>
                <p class="text-muted" style="margin-top: 12px;">
                    <i class="fas fa-info-circle"></i> SSE 端点用于建立服务器发送事件连接，消息端点用于发送客户端消息。
                </p>
            </div>
        `;
    } else if (protocol === 'streamable') {
        const streamableEndpoint = `${baseUrl}/streamable/${path}`;
        endpointInfo = `
            <div class="doc-section">
                <h3>Streamable 端点信息</h3>
                <p><strong>Streamable 端点:</strong></p>
                <pre><code>${streamableEndpoint}</code></pre>
                <p class="text-muted" style="margin-top: 12px;">
                    <i class="fas fa-info-circle"></i> Streamable 端点支持基于 HTTP 的流式传输协议。
                </p>
            </div>
        `;
    }
    
    showModal(`${app.name} - 接口信息`, endpointInfo, null, false);
}

// 查看应用 JSON 数据
async function viewApplicationJSON(id) {
    try {
        const data = await apiRequest(`/applications-detail/${id}`);
        if (!data) return;
        
        const jsonString = JSON.stringify(data, null, 2);
        const jsonId = 'app-json-' + id;
        
        // 统计工具定义数量
        const toolCount = data.tool_definitions ? data.tool_definitions.length : 0;
        const appName = data.application ? data.application.name : '应用';
        
        const jsonContent = `
            <div class="doc-section">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
                    <h3 style="margin: 0;">应用详细信息 (包含 ${toolCount} 个工具定义)</h3>
                    <button class="btn btn-sm btn-primary" onclick="copyToClipboard('${jsonId}')" title="复制 JSON">
                        <i class="fas fa-copy"></i> 复制
                    </button>
                </div>
                <pre style="background: #f5f5f5; padding: 16px; border-radius: 4px; max-height: 500px; overflow: auto; user-select: text;"><code id="${jsonId}" style="user-select: text;">${escapeHtml(jsonString)}</code></pre>
                <p class="text-muted" style="margin-top: 12px;">
                    <i class="fas fa-info-circle"></i> 这是从 GetApplication 接口返回的完整 JSON 数据，包含应用信息和工具定义 (ToolDefinitions)。
                </p>
            </div>
        `;
        
        showModal(`${appName} - JSON 数据`, jsonContent, null, false);
    } catch (error) {
        showToast('获取应用数据失败: ' + error.message, 'error');
    }
}

// 复制到剪贴板
function copyToClipboard(elementId) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const text = element.textContent;
    
    // 使用现代 Clipboard API
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            showToast('已复制到剪贴板', 'success');
        }).catch(err => {
            // 降级到旧方法
            fallbackCopyToClipboard(text);
        });
    } else {
        // 降级到旧方法
        fallbackCopyToClipboard(text);
    }
}

// 降级复制方法
function fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        document.execCommand('copy');
        showToast('已复制到剪贴板', 'success');
    } catch (err) {
        showToast('复制失败，请手动复制', 'error');
    }
    
    document.body.removeChild(textArea);
}

function editApplication(id) {
    showAppForm(id);
}

function manageApplicationInterfaces(id) {
    const app = state.applications.find(a => a.id === id);
    if (!app) return;
    
    // 在全局应用选择器中选中该应用
    const globalAppSelect = document.getElementById('global-app-select');
    globalAppSelect.value = id;
    state.currentApp = app;
    
    // 切换到接口管理标签页
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
        if (item.dataset.tab === 'interfaces') {
            item.classList.add('active');
        }
    });
    
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById('interfaces-tab').classList.add('active');
    
    // 加载该应用的接口列表
    loadInterfaces();
    
    // 提示用户
    showToast(`已切换到"${app.name}"的接口管理`, 'success');
}

async function deleteApplication(id) {
    if (!confirm('确定要删除此应用吗？这将删除所有相关的接口和类型。')) return;
    
    try {
        await apiRequest(`/applications/${id}`, { method: 'DELETE' });
        showToast('应用删除成功', 'success');
        loadApplications();
    } catch (error) {
        showToast('删除失败: ' + error.message, 'error');
    }
}

// ========== 自定义类型管理 ==========

// 类型表单状态
let currentTypeId = null;

// 显示类型表单
function showTypeForm(typeId = null) {
    const appId = document.getElementById('global-app-select').value;
    if (!appId) {
        showToast('请先选择一个应用', 'warning');
        return;
    }
    
    currentTypeId = typeId;
    document.getElementById('type-list-view').style.display = 'none';
    document.getElementById('type-form-view').style.display = 'block';
    
    // 清空字段容器
    document.getElementById('fields-container').innerHTML = '';
    
    if (typeId) {
        // 编辑模式
        const type = state.customTypes.find(t => t.id === typeId);
        if (type) {
            document.getElementById('type-form-title').textContent = '编辑类型';
            document.getElementById('type-form-subtitle').textContent = '修改类型信息';
            document.getElementById('type-name').value = type.name;
            document.getElementById('type-description').value = type.description || '';
            
            // 加载字段
            if (type.fields && type.fields.length > 0) {
                type.fields.forEach(field => {
                    addFieldRow(field);
                });
            } else {
                addFieldRow();
            }
        }
    } else {
        // 创建模式
        document.getElementById('type-form-title').textContent = '创建类型';
        document.getElementById('type-form-subtitle').textContent = '定义数据类型和字段';
        document.getElementById('type-name').value = '';
        document.getElementById('type-description').value = '';
        addFieldRow();
    }
}

// 隐藏类型表单
function hideTypeForm() {
    document.getElementById('type-list-view').style.display = 'block';
    document.getElementById('type-form-view').style.display = 'none';
    currentTypeId = null;
}

// 创建类型按钮
document.getElementById('create-type-btn').addEventListener('click', () => {
    showTypeForm();
});

// 类型表单返回按钮
document.getElementById('type-form-back').addEventListener('click', hideTypeForm);
document.getElementById('type-form-cancel').addEventListener('click', hideTypeForm);

// 类型表单提交
document.getElementById('type-form-submit').addEventListener('click', async () => {
    const appId = document.getElementById('global-app-select').value;
    const name = document.getElementById('type-name').value;
    const description = document.getElementById('type-description').value;
    
    if (!name) {
        showToast('请填写类型名称', 'error');
        return;
    }
    
    const fields = [];
    document.querySelectorAll('.field-row').forEach(row => {
        const fieldName = row.querySelector('.field-name-input').value;
        const fieldType = row.querySelector('.field-type-select').value;
        const fieldRef = row.querySelector('.field-ref-input')?.value;
        const fieldIsArray = row.querySelector('.field-array-checkbox')?.checked || false;
        const fieldRequired = row.querySelector('.field-required-checkbox').checked;
        const fieldDesc = row.querySelector('.field-desc-input').value;
        
        if (fieldName) {
            const field = {
                name: fieldName,
                type: fieldType,
                is_array: fieldIsArray,
                required: fieldRequired,
                description: fieldDesc
            };
            
            if (fieldType === 'custom' && fieldRef) {
                field.ref = parseInt(fieldRef);
            }
            
            fields.push(field);
        }
    });
    
    try {
        if (currentTypeId) {
            // 更新
            await apiRequest(`/custom-types/${currentTypeId}`, {
                method: 'PUT',
                body: JSON.stringify({ name, description, fields })
            });
            showToast('类型更新成功', 'success');
        } else {
            // 创建
            await apiRequest('/custom-types', {
                method: 'POST',
                body: JSON.stringify({ app_id: parseInt(appId), name, description, fields })
            });
            showToast('类型创建成功', 'success');
        }
        hideTypeForm();
        loadCustomTypes();
    } catch (error) {
        showToast('操作失败: ' + error.message, 'error');
    }
});

// 获取字段类型的显示名称
function getFieldTypeDisplay(field) {
    if (field.type === 'custom' && field.ref) {
        const refType = state.customTypes.find(t => t.id === field.ref);
        return refType ? refType.name : 'custom';
    }
    return field.type;
}

async function loadCustomTypes() {
    const appId = document.getElementById('global-app-select').value;
    
    if (!appId) {
        document.getElementById('no-app-alert').style.display = 'flex';
        document.getElementById('types-grid').innerHTML = '';
        return;
    }
    
    document.getElementById('no-app-alert').style.display = 'none';
    
    try {
        const types = await apiRequest(`/custom-types?app_id=${appId}`);
        state.customTypes = types;
        renderCustomTypes(types);
    } catch (error) {
        showToast('加载类型失败: ' + error.message, 'error');
    }
}

function renderCustomTypes(types) {
    const grid = document.getElementById('types-grid');
    
    // 获取搜索关键词
    const searchTerm = document.getElementById('type-search')?.value.toLowerCase() || '';
    
    // 过滤类型
    const filteredTypes = types.filter(type => {
        if (!searchTerm) return true;
        return type.name.toLowerCase().includes(searchTerm) ||
               (type.description && type.description.toLowerCase().includes(searchTerm));
    });
    
    if (!filteredTypes || filteredTypes.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-shapes"></i>
                <h3>${searchTerm ? '未找到匹配的类型' : '还没有自定义类型'}</h3>
                <p>${searchTerm ? '尝试其他搜索关键词' : '点击"创建类型"开始'}</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = filteredTypes.map(type => `
        <div class="card" onclick="viewCustomType(${type.id})">
            <div class="card-header">
                <div>
                    <div class="card-title">${type.name}</div>
                    <div class="card-subtitle">${type.fields?.length || 0} 个字段</div>
                </div>
                <span class="card-badge badge-primary">
                    <i class="fas fa-shapes"></i>
                </span>
            </div>
            <div class="card-body">
                <p class="card-subtitle">${type.description || '暂无描述'}</p>
                ${type.fields && type.fields.length > 0 ? `
                    <div class="fields-list">
                        ${type.fields.slice(0, 3).map(field => `
                            <div class="field-item">
                                <div class="field-info">
                                    <div class="field-name">${field.name}</div>
                                    <div class="field-meta">
                                        <span>${getFieldTypeDisplay(field)}${field.is_array ? '[]' : ''}</span>
                                        ${field.required ? '<span class="badge-danger">必填</span>' : ''}
                                    </div>
                                </div>
                            </div>
                        `).join('')}
                        ${type.fields.length > 3 ? `<div class="text-muted">还有 ${type.fields.length - 3} 个字段...</div>` : ''}
                    </div>
                ` : ''}
            </div>
            <div class="card-footer">
                <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editCustomType(${type.id})">
                    <i class="fas fa-edit"></i> 编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteCustomType(${type.id})">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
    `).join('');
}

// 类型搜索事件监听
document.getElementById('type-search')?.addEventListener('input', () => {
    renderCustomTypes(state.customTypes);
});

function addFieldRow(fieldData = null) {
    const container = document.getElementById('fields-container');
    const row = document.createElement('div');
    row.className = 'field-row mb-2';
    
    // 构建类型选项（包含自定义类型）
    const buildTypeOptions = () => {
        let options = `
            <option value="string" ${fieldData && fieldData.type === 'string' ? 'selected' : ''}>string</option>
            <option value="number" ${fieldData && fieldData.type === 'number' ? 'selected' : ''}>number</option>
            <option value="boolean" ${fieldData && fieldData.type === 'boolean' ? 'selected' : ''}>boolean</option>
        `;
        
        // 添加当前应用的自定义类型
        if (state.customTypes && state.customTypes.length > 0) {
            options += '<optgroup label="自定义类型">';
            state.customTypes.forEach(type => {
                const isSelected = fieldData && fieldData.type === 'custom' && fieldData.ref === type.id;
                options += `<option value="custom" data-ref="${type.id}" ${isSelected ? 'selected' : ''}>${type.name}</option>`;
            });
            options += '</optgroup>';
        }
        
        return options;
    };
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="field-name-input" placeholder="字段名" value="${fieldData ? fieldData.name : ''}">
            <select class="field-type-select" onchange="handleTypeChange(this)">
                ${buildTypeOptions()}
            </select>
            <input type="hidden" class="field-ref-input" value="${fieldData && fieldData.ref ? fieldData.ref : ''}">
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-array-checkbox" ${fieldData && fieldData.is_array ? 'checked' : ''}>
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-required-checkbox" ${fieldData && fieldData.required ? 'checked' : ''}>
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <input type="text" class="field-desc-input" placeholder="字段描述（可选）" value="${fieldData && fieldData.description ? fieldData.description : ''}">
    `;
    container.appendChild(row);
}

// 处理类型选择变化
function handleTypeChange(selectElement) {
    const row = selectElement.closest('.field-row, .param-row');
    const refInput = row.querySelector('.field-ref-input, .param-ref-input');
    const selectedOption = selectElement.options[selectElement.selectedIndex];
    
    if (selectedOption.value === 'custom' && selectedOption.dataset.ref) {
        refInput.value = selectedOption.dataset.ref;
    } else {
        refInput.value = '';
    }
}

// 处理参数类型变化（包括默认值输入框的显示/隐藏）
function handleParamTypeChange(selectElement) {
    // 首先处理引用值的更新
    handleTypeChange(selectElement);
    
    const row = selectElement.closest('.param-row');
    const defaultInput = row.querySelector('.param-default-input');
    
    // 如果存在默认值输入框（即在默认参数Tab中）
    if (defaultInput) {
        const selectedOption = selectElement.options[selectElement.selectedIndex];
        const isBasicType = ['string', 'number', 'boolean'].includes(selectedOption.value);
        
        // 根据类型显示/隐藏默认值输入框
        if (isBasicType) {
            defaultInput.style.display = '';
        } else {
            defaultInput.style.display = 'none';
            defaultInput.value = ''; // 清空默认值
        }
    }
    
    // 更新必填复选框状态（默认参数中数组不能必填）
    updateRequiredCheckboxState(row);
}

// 处理参数数组复选框变化
function handleParamArrayChange(checkbox) {
    const row = checkbox.closest('.param-row');
    updateRequiredCheckboxState(row);
}

// 更新必填复选框状态（默认参数中数组不能必填）
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

function viewCustomType(id) {
    const type = state.customTypes.find(t => t.id === id);
    if (!type) return;
    
    showModal(type.name, `
        <div class="doc-section">
            <h3>类型信息</h3>
            <p><strong>描述:</strong> ${type.description || '无'}</p>
            <h3>字段列表</h3>
            ${type.fields && type.fields.length > 0 ? `
                <div class="fields-list">
                    ${type.fields.map(field => `
                        <div class="field-item">
                            <div class="field-info">
                                <div class="field-name">${field.name}</div>
                                <div class="field-meta">
                                    <span>${getFieldTypeDisplay(field)}${field.is_array ? '[]' : ''}</span>
                                    ${field.required ? '<span class="badge-danger">必填</span>' : ''}
                                    ${field.description ? `<span>${field.description}</span>` : ''}
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            ` : '<p>暂无字段</p>'}
        </div>
    `, null, false);
}

function editCustomType(id) {
    showTypeForm(id);
}

async function deleteCustomType(id) {
    if (!confirm('确定要删除此类型吗？')) return;
    
    try {
        await apiRequest(`/custom-types/${id}`, { method: 'DELETE' });
        showToast('类型删除成功', 'success');
        loadCustomTypes();
    } catch (error) {
        showToast('删除失败: ' + error.message, 'error');
    }
}

// ========== 接口管理 ==========

// 接口表单状态
let currentInterfaceId = null;
let currentParamsTab = 'input'; // 当前激活的参数Tab

// 切换参数Tab
function switchParamsTab(tab) {
    currentParamsTab = tab;
    
    // 更新Tab按钮状态
    document.querySelectorAll('.params-tab-btn').forEach(btn => {
        if (btn.dataset.tab === tab) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });
    
    // 更新Tab描述显示
    document.querySelectorAll('.tab-desc').forEach(desc => {
        desc.style.display = 'none';
    });
    const descId = `${tab}-tab-desc`;
    const descElem = document.getElementById(descId);
    if (descElem) {
        descElem.style.display = 'block';
    }
    
    // 更新Tab内容显示
    const containers = ['input-params-container', 'output-params-container', 'fixed-params-container'];
    containers.forEach(containerId => {
        const container = document.getElementById(containerId);
        if (container) {
            if (containerId === `${tab}-params-container`) {
                container.style.display = 'block';
                container.classList.add('active');
            } else {
                container.style.display = 'none';
                container.classList.remove('active');
            }
        }
    });
}

// 显示接口表单
async function showInterfaceForm(interfaceId = null) {
    const appId = document.getElementById('global-app-select').value;
    if (!appId) {
        showToast('请先选择一个应用', 'warning');
        return;
    }
    
    // 确保已加载自定义类型
    if (!state.customTypes || state.customTypes.length === 0) {
        try {
            const types = await apiRequest(`/custom-types?app_id=${appId}`);
            state.customTypes = types;
        } catch (error) {
            console.error('加载自定义类型失败:', error);
        }
    }
    
    currentInterfaceId = interfaceId;
    currentParamsTab = 'input'; // 重置为 input Tab
    document.getElementById('interface-list-view').style.display = 'none';
    document.getElementById('interface-form-view').style.display = 'block';
    
    // 清空参数容器
    document.getElementById('input-params-container').innerHTML = '';
    document.getElementById('output-params-container').innerHTML = '';
    document.getElementById('fixed-params-container').innerHTML = '';
    
    // 重置Tab状态
    switchParamsTab('input');
    
    if (interfaceId) {
        // 编辑模式
        const iface = state.interfaces.find(i => i.id === interfaceId);
        if (iface) {
            document.getElementById('interface-form-title').textContent = '编辑接口';
            document.getElementById('interface-form-subtitle').textContent = '修改接口信息';
            document.getElementById('interface-name').value = iface.name;
            document.getElementById('interface-description').value = iface.description || '';
            document.getElementById('interface-method').value = iface.method;
            document.getElementById('interface-protocol').value = iface.protocol;
            document.getElementById('interface-url').value = iface.url;
            document.getElementById('interface-auth').value = iface.auth_type;
            document.getElementById('interface-post-process').value = iface.post_process || '';
            
            // 加载参数，根据 group 分配到不同Tab
            if (iface.parameters && iface.parameters.length > 0) {
                iface.parameters.forEach(param => {
                    addParamRow(param);
                });
            }
        }
    } else {
        // 创建模式
        document.getElementById('interface-form-title').textContent = '创建接口';
        document.getElementById('interface-form-subtitle').textContent = '配置接口信息和参数';
        document.getElementById('interface-name').value = '';
        document.getElementById('interface-description').value = '';
        document.getElementById('interface-method').value = 'GET';
        document.getElementById('interface-protocol').value = 'http';
        document.getElementById('interface-url').value = '';
        document.getElementById('interface-auth').value = 'none';
        document.getElementById('interface-post-process').value = '';
    }
}

// 隐藏接口表单
function hideInterfaceForm() {
    document.getElementById('interface-list-view').style.display = 'block';
    document.getElementById('interface-form-view').style.display = 'none';
    currentInterfaceId = null;
}

// 创建接口按钮
document.getElementById('create-interface-btn').addEventListener('click', () => {
    showInterfaceForm();
});

// 接口表单返回按钮
document.getElementById('interface-form-back').addEventListener('click', hideInterfaceForm);
document.getElementById('interface-form-cancel').addEventListener('click', hideInterfaceForm);

// 接口表单提交
document.getElementById('interface-form-submit').addEventListener('click', async () => {
    const appId = document.getElementById('global-app-select').value;
    const name = document.getElementById('interface-name').value;
    const description = document.getElementById('interface-description').value;
    const method = document.getElementById('interface-method').value;
    const protocol = document.getElementById('interface-protocol').value;
    const url = document.getElementById('interface-url').value;
    const auth_type = document.getElementById('interface-auth').value;
    const post_process = document.getElementById('interface-post-process').value;
    
    if (!name || !url) {
        showToast('请填写必填字段', 'error');
        return;
    }
    
    const parameters = [];
    
    // 收集所有三个 Tab 的参数
    let validationError = null;
    ['input', 'output', 'fixed'].forEach(tabName => {
        const containerId = `${tabName}-params-container`;
        document.querySelectorAll(`#${containerId} .param-row`).forEach(row => {
            const param = collectParamFromRow(row);
            if (param) {
                // 验证 fixed 参数必须有默认值且不能是数组
                if (param.group === 'fixed') {
                    if (!param.default_value || param.default_value.trim() === '') {
                        validationError = `Fixed 参数 "${param.name}" 必须设置默认值`;
                    }
                    if (param.is_array) {
                        validationError = `Fixed 参数 "${param.name}" 不能是数组类型`;
                    }
                }
                parameters.push(param);
            }
        });
    });
    
    if (validationError) {
        showToast(validationError, 'error');
        return;
    }
    
    try {
        if (currentInterfaceId) {
            // 更新
            await apiRequest(`/interfaces/${currentInterfaceId}`, {
                method: 'PUT',
                body: JSON.stringify({
                    name,
                    description,
                    method,
                    protocol,
                    url,
                    auth_type,
                    post_process,
                    parameters
                })
            });
            showToast('接口更新成功', 'success');
        } else {
            // 创建
            await apiRequest('/interfaces', {
                method: 'POST',
                body: JSON.stringify({
                    app_id: parseInt(appId),
                    name,
                    description,
                    method,
                    protocol,
                    url,
                    auth_type,
                    enabled: true,
                    post_process,
                    parameters
                })
            });
            showToast('接口创建成功', 'success');
        }
        hideInterfaceForm();
        loadInterfaces();
    } catch (error) {
        showToast('操作失败: ' + error.message, 'error');
    }
});

// 从参数行收集数据的辅助函数
function collectParamFromRow(row) {
    const paramName = row.querySelector('.param-name-input').value;
    const paramType = row.querySelector('.param-type-select').value;
    const paramRef = row.querySelector('.param-ref-input')?.value;
    const paramLocation = row.querySelector('.param-location-select').value;
    const paramGroup = row.querySelector('.param-group-input')?.value || 'input';
    const paramIsArray = row.querySelector('.param-array-checkbox')?.checked || false;
    const paramRequired = row.querySelector('.param-required-checkbox').checked;
    const paramDefaultValue = row.querySelector('.param-default-input')?.value;
    const paramDescription = row.querySelector('.param-desc-input')?.value;
    
    if (!paramName) return null;
    
    const param = {
        name: paramName,
        type: paramType,
        location: paramLocation,
        group: paramGroup,
        is_array: paramIsArray,
        required: paramRequired
    };
    
    if (paramType === 'custom' && paramRef) {
        param.ref = parseInt(paramRef);
    }
    
    // 保存默认值
    // Fixed 参数必须有默认值
    // Input 参数不再支持默认值（已移除该功能）
    // Output 参数不能有默认值
    if (paramGroup === 'fixed') {
        param.default_value = paramDefaultValue || '';
    }
    
    if (paramDescription) {
        param.description = paramDescription;
    }
    
    return param;
}

async function loadInterfaces() {
    const appId = document.getElementById('global-app-select').value;
    
    if (!appId) {
        document.getElementById('no-app-alert-interface').style.display = 'flex';
        document.getElementById('interfaces-grid').innerHTML = '';
        return;
    }
    
    document.getElementById('no-app-alert-interface').style.display = 'none';
    
    try {
        const interfaces = await apiRequest(`/interfaces?app_id=${appId}`);
        state.interfaces = interfaces;
        renderInterfaces(interfaces);
    } catch (error) {
        showToast('加载接口失败: ' + error.message, 'error');
    }
}

function renderInterfaces(interfaces) {
    const grid = document.getElementById('interfaces-grid');
    
    // 获取搜索关键词和过滤条件
    const searchTerm = document.getElementById('interface-search')?.value.toLowerCase() || '';
    const methodFilter = document.getElementById('method-filter')?.value || '';
    
    // 过滤接口
    const filteredInterfaces = interfaces.filter(iface => {
        const matchesSearch = !searchTerm || 
            iface.name.toLowerCase().includes(searchTerm) ||
            (iface.description && iface.description.toLowerCase().includes(searchTerm)) ||
            (iface.url && iface.url.toLowerCase().includes(searchTerm));
        
        const matchesMethod = !methodFilter || iface.method === methodFilter;
        
        return matchesSearch && matchesMethod;
    });
    
    if (!filteredInterfaces || filteredInterfaces.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-plug"></i>
                <h3>${searchTerm || methodFilter ? '未找到匹配的接口' : '还没有接口'}</h3>
                <p>${searchTerm || methodFilter ? '尝试其他搜索条件' : '点击"创建接口"开始'}</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = filteredInterfaces.map(iface => `
        <div class="card" onclick="viewInterface(${iface.id})">
            <div class="card-header">
                <div>
                    <div class="card-title">${iface.name}</div>
                    <div class="card-subtitle">${iface.url}</div>
                </div>
                <span class="card-badge badge-${getMethodColor(iface.method)}">
                    ${iface.method}
                </span>
            </div>
            <div class="card-body">
                <p class="card-subtitle">${iface.description || '暂无描述'}</p>
                <div class="card-meta">
                    <div class="card-meta-item">
                        <i class="fas fa-cog"></i>
                        ${iface.parameters?.length || 0} 个参数
                    </div>
                    <div class="card-meta-item">
                        <i class="fas fa-${iface.enabled ? 'check-circle' : 'times-circle'}"></i>
                        ${iface.enabled ? '已启用' : '已禁用'}
                    </div>
                </div>
            </div>
            <div class="card-footer">
                <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editInterface(${iface.id})">
                    <i class="fas fa-edit"></i> 编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteInterface(${iface.id})">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
    `).join('');
}

// 接口搜索和过滤事件监听
document.getElementById('interface-search')?.addEventListener('input', () => {
    renderInterfaces(state.interfaces);
});

document.getElementById('method-filter')?.addEventListener('change', () => {
    renderInterfaces(state.interfaces);
});

function getMethodColor(method) {
    const colors = {
        'GET': 'success',
        'POST': 'primary',
        'PUT': 'warning',
        'DELETE': 'danger',
        'PATCH': 'info'
    };
    return colors[method] || 'secondary';
}

function addParamRow(paramData = null) {
    // 确定参数组类型：从参数数据获取，或使用当前激活的 Tab
    const group = paramData && paramData.group ? paramData.group : currentParamsTab;
    
    // 根据 group 选择对应的容器
    const containerId = `${group}-params-container`;
    const container = document.getElementById(containerId);
    
    if (!container) {
        console.error(`Container not found for group: ${group}`);
        return;
    }
    
    const row = document.createElement('div');
    row.className = 'param-row mb-2';
    
    // 构建类型选项（包含自定义类型）
    const buildTypeOptions = () => {
        let options = `
            <option value="string" ${paramData && paramData.type === 'string' ? 'selected' : ''}>string</option>
            <option value="number" ${paramData && paramData.type === 'number' ? 'selected' : ''}>number</option>
            <option value="boolean" ${paramData && paramData.type === 'boolean' ? 'selected' : ''}>boolean</option>
        `;
        
        // 添加当前应用的自定义类型
        if (state.customTypes && state.customTypes.length > 0) {
            options += '<optgroup label="自定义类型">';
            state.customTypes.forEach(type => {
                const isSelected = paramData && paramData.type === 'custom' && paramData.ref === type.id;
                options += `<option value="custom" data-ref="${type.id}" ${isSelected ? 'selected' : ''}>${type.name}</option>`;
            });
            options += '</optgroup>';
        }
        
        return options;
    };
    
    // 判断当前参数类型是否为基本类型
    const currentType = paramData ? paramData.type : 'string';
    const isBasicType = ['string', 'number', 'boolean'].includes(currentType);
    
    // Fixed 参数需要显示默认值输入框（必填）
    // Input 参数不显示默认值（已移除该功能）
    // Output 参数不需要默认值
    const showDefaultValue = (group === 'fixed');
    const defaultValueRequired = (group === 'fixed') ? '（必填）' : '';
    
    const defaultValueHTML = showDefaultValue ? `
        <div class="form-row param-extra-row" style="margin-top: 4px;">
            <input type="text" class="param-default-input" placeholder="默认值${defaultValueRequired}" value="${paramData && paramData.default_value ? paramData.default_value : ''}" style="flex: 1;" required>
            <input type="text" class="param-desc-input" placeholder="参数描述（可选）" value="${paramData && paramData.description ? paramData.description : ''}" style="flex: 1;">
        </div>
    ` : `
        <div class="form-row param-extra-row" style="margin-top: 4px;">
            <input type="text" class="param-desc-input" placeholder="参数描述（可选）" value="${paramData && paramData.description ? paramData.description : ''}" style="flex: 1;">
        </div>
    `;
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="param-name-input" placeholder="参数名" value="${paramData ? paramData.name : ''}">
            <select class="param-type-select" onchange="handleParamTypeChange(this)">
                ${buildTypeOptions()}
            </select>
            <input type="hidden" class="param-ref-input" value="${paramData && paramData.ref ? paramData.ref : ''}">
            <input type="hidden" class="param-group-input" value="${group}">
            <select class="param-location-select">
                <option value="query" ${paramData && paramData.location === 'query' ? 'selected' : ''}>query</option>
                <option value="header" ${paramData && paramData.location === 'header' ? 'selected' : ''}>header</option>
                <option value="body" ${paramData && paramData.location === 'body' ? 'selected' : ''}>body</option>
                <option value="path" ${paramData && paramData.location === 'path' ? 'selected' : ''}>path</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-array-checkbox" ${paramData && paramData.is_array ? 'checked' : ''} ${group === 'fixed' ? 'disabled' : ''} onchange="handleParamArrayChange(this)">
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-required-checkbox" ${paramData && paramData.required ? 'checked' : ''} ${group === 'fixed' ? 'checked disabled' : ''}>
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
        ${defaultValueHTML}
    `;
    
    // Fixed 参数特殊处理：必填且不可修改，数组选项禁用
    if (group === 'fixed') {
        const requiredCheckbox = row.querySelector('.param-required-checkbox');
        if (requiredCheckbox) {
            requiredCheckbox.checked = true;
            requiredCheckbox.disabled = true;
        }
        
        const arrayCheckbox = row.querySelector('.param-array-checkbox');
        if (arrayCheckbox) {
            arrayCheckbox.checked = false;
            arrayCheckbox.disabled = true;
        }
    }
    container.appendChild(row);
}

function viewInterface(id) {
    const iface = state.interfaces.find(i => i.id === id);
    if (!iface) return;
    
    // 获取参数类型的显示名称
    const getParamTypeDisplay = (param) => {
        if (param.type === 'custom' && param.ref) {
            const refType = state.customTypes.find(t => t.id === param.ref);
            return refType ? refType.name : 'custom';
        }
        return param.type;
    };
    
    showModal(iface.name, `
        <div class="doc-section">
            <h3>基本信息</h3>
            <p><strong>方法:</strong> ${iface.method}</p>
            <p><strong>URL:</strong> ${iface.url}</p>
            <p><strong>协议:</strong> ${iface.protocol}</p>
            <p><strong>认证:</strong> ${iface.auth_type}</p>
            <p><strong>描述:</strong> ${iface.description || '无'}</p>
            
            <h3>参数列表</h3>
            ${iface.parameters && iface.parameters.length > 0 ? `
                <div class="fields-list">
                    ${iface.parameters.map(param => `
                        <div class="field-item">
                            <div class="field-info">
                                <div class="field-name">${param.name}</div>
                                <div class="field-meta">
                                    <span>${getParamTypeDisplay(param)}${param.is_array ? '[]' : ''}</span>
                                    <span>${param.location}</span>
                                    <span class="badge-${param.group === 'input' ? 'primary' : param.group === 'output' ? 'success' : 'warning'}">${param.group || 'input'}</span>
                                    ${param.required ? '<span class="badge-danger">必填</span>' : ''}
                                    ${param.default_value ? `<span>默认值: ${param.default_value}</span>` : ''}
                                </div>
                                ${param.description ? `<div style="margin-top: 4px; color: #666;">${param.description}</div>` : ''}
                            </div>
                        </div>
                    `).join('')}
                </div>
            ` : '<p>暂无参数</p>'}
        </div>
    `, null, false);
}

async function editInterface(id) {
    await showInterfaceForm(id);
}

async function deleteInterface(id) {
    if (!confirm('确定要删除此接口吗？')) return;
    
    try {
        await apiRequest(`/interfaces/${id}`, { method: 'DELETE' });
        showToast('接口删除成功', 'success');
        loadInterfaces();
    } catch (error) {
        showToast('删除失败: ' + error.message, 'error');
    }
}

// ========== 全局应用选择器 ==========

document.getElementById('global-app-select').addEventListener('change', (e) => {
    const appId = e.target.value;
    if (appId) {
        state.currentApp = state.applications.find(a => a.id == appId);
        
        // 重新加载当前标签页的数据
        const activeTab = document.querySelector('.nav-item.active').dataset.tab;
        if (activeTab === 'custom-types') {
            loadCustomTypes();
        } else if (activeTab === 'interfaces') {
            loadInterfaces();
        }
    }
});

// ========== 初始化 ==========

document.addEventListener('DOMContentLoaded', () => {
    loadApplications();
});
