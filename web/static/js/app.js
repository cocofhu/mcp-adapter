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
    
    if (!apps || apps.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-folder-open"></i>
                <h3>还没有应用</h3>
                <p>点击"创建应用"开始</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = apps.map(app => `
        <div class="card" onclick="viewApplication(${app.id})">
            <div class="card-header">
                <div>
                    <div class="card-title">${app.name}</div>
                    <div class="card-subtitle">${app.path || '/'}</div>
                </div>
                <span class="card-badge badge-${app.enabled ? 'success' : 'warning'}">
                    ${app.enabled ? '已启用' : '已禁用'}
                </span>
            </div>
            <div class="card-body">
                <p class="card-subtitle">${app.description || '暂无描述'}</p>
                <div class="card-meta">
                    <div class="card-meta-item">
                        <i class="fas fa-network-wired"></i>
                        ${app.protocol || 'http'}
                    </div>
                </div>
            </div>
            <div class="card-footer">
                <button class="btn btn-sm btn-secondary" onclick="event.stopPropagation(); editApplication(${app.id})">
                    <i class="fas fa-edit"></i> 编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="event.stopPropagation(); deleteApplication(${app.id})">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
    `).join('');
}

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

// 创建应用
document.getElementById('create-app-btn').addEventListener('click', () => {
    showModal('创建应用', `
        <div class="form-group">
            <label class="required">应用名称</label>
            <input type="text" id="app-name" placeholder="例如: 我的API服务">
        </div>
        <div class="form-group">
            <label>应用描述</label>
            <textarea id="app-description" rows="3" placeholder="描述应用的功能"></textarea>
        </div>
        <div class="form-group">
            <label class="required">应用路径</label>
            <input type="text" id="app-path" placeholder="例如: my-api">
        </div>
        <div class="form-group">
            <label class="required">协议类型</label>
            <select id="app-protocol">
                <option value="sse">SSE</option>
                <option value="http">HTTP</option>
            </select>
        </div>
    `, async () => {
        const name = document.getElementById('app-name').value;
        const description = document.getElementById('app-description').value;
        const path = document.getElementById('app-path').value;
        const protocol = document.getElementById('app-protocol').value;
        
        if (!name || !path) {
            showToast('请填写必填字段', 'error');
            return;
        }
        
        try {
            await apiRequest('/applications', {
                method: 'POST',
                body: JSON.stringify({ name, description, path, protocol, enabled: true })
            });
            showToast('应用创建成功', 'success');
            loadApplications();
        } catch (error) {
            showToast('创建失败: ' + error.message, 'error');
        }
    });
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

function editApplication(id) {
    const app = state.applications.find(a => a.id === id);
    if (!app) return;
    
    showModal('编辑应用', `
        <div class="form-group">
            <label>应用名称</label>
            <input type="text" id="edit-app-name" value="${app.name}">
        </div>
        <div class="form-group">
            <label>应用描述</label>
            <textarea id="edit-app-description" rows="3">${app.description || ''}</textarea>
        </div>
        <div class="form-group">
            <label>应用路径</label>
            <input type="text" id="edit-app-path" value="${app.path}">
        </div>
    `, async () => {
        const name = document.getElementById('edit-app-name').value;
        const description = document.getElementById('edit-app-description').value;
        const path = document.getElementById('edit-app-path').value;
        
        try {
            await apiRequest(`/applications/${id}`, {
                method: 'PUT',
                body: JSON.stringify({ name, description, path })
            });
            showToast('应用更新成功', 'success');
            loadApplications();
        } catch (error) {
            showToast('更新失败: ' + error.message, 'error');
        }
    });
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
    
    if (!types || types.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-shapes"></i>
                <h3>还没有自定义类型</h3>
                <p>点击"创建类型"开始</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = types.map(type => `
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
                                        <span>${field.type}${field.is_array ? '[]' : ''}</span>
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

// 创建自定义类型
document.getElementById('create-type-btn').addEventListener('click', () => {
    const appId = document.getElementById('global-app-select').value;
    if (!appId) {
        showToast('请先选择一个应用', 'warning');
        return;
    }
    
    showModal('创建自定义类型', `
        <div class="form-group">
            <label class="required">类型名称</label>
            <input type="text" id="type-name" placeholder="例如: User">
        </div>
        <div class="form-group">
            <label>类型描述</label>
            <textarea id="type-description" rows="2" placeholder="描述这个类型的用途"></textarea>
        </div>
        <div class="form-group">
            <label>字段定义</label>
            <div id="fields-container"></div>
            <button type="button" class="btn btn-secondary mt-2" onclick="addFieldRow()">
                <i class="fas fa-plus"></i> 添加字段
            </button>
        </div>
    `, async () => {
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
            const fieldRequired = row.querySelector('.field-required-checkbox').checked;
            const fieldDesc = row.querySelector('.field-desc-input').value;
            
            if (fieldName) {
                fields.push({
                    name: fieldName,
                    type: fieldType,
                    required: fieldRequired,
                    description: fieldDesc
                });
            }
        });
        
        try {
            await apiRequest('/custom-types', {
                method: 'POST',
                body: JSON.stringify({ app_id: parseInt(appId), name, description, fields })
            });
            showToast('类型创建成功', 'success');
            loadCustomTypes();
        } catch (error) {
            showToast('创建失败: ' + error.message, 'error');
        }
    });
    
    // 添加一个默认字段行
    setTimeout(() => addFieldRow(), 100);
});

function addFieldRow() {
    const container = document.getElementById('fields-container');
    const row = document.createElement('div');
    row.className = 'field-row mb-2';
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="field-name-input" placeholder="字段名">
            <select class="field-type-select">
                <option value="string">string</option>
                <option value="number">number</option>
                <option value="boolean">boolean</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px;">
                <input type="checkbox" class="field-required-checkbox">
                必填
            </label>
            <button type="button" class="btn btn-sm btn-danger" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <input type="text" class="field-desc-input mt-1" placeholder="字段描述（可选）">
    `;
    container.appendChild(row);
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
                                    <span>${field.type}${field.is_array ? '[]' : ''}</span>
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
    const type = state.customTypes.find(t => t.id === id);
    if (!type) return;
    
    showModal('编辑自定义类型', `
        <div class="form-group">
            <label>类型名称</label>
            <input type="text" id="edit-type-name" value="${type.name}">
        </div>
        <div class="form-group">
            <label>类型描述</label>
            <textarea id="edit-type-description" rows="2">${type.description || ''}</textarea>
        </div>
        <div class="form-group">
            <label>字段定义</label>
            <div id="edit-fields-container">
                ${type.fields?.map(field => `
                    <div class="field-row mb-2">
                        <div class="form-row">
                            <input type="text" class="field-name-input" value="${field.name}" placeholder="字段名">
                            <select class="field-type-select">
                                <option value="string" ${field.type === 'string' ? 'selected' : ''}>string</option>
                                <option value="number" ${field.type === 'number' ? 'selected' : ''}>number</option>
                                <option value="boolean" ${field.type === 'boolean' ? 'selected' : ''}>boolean</option>
                            </select>
                            <label style="display: flex; align-items: center; gap: 4px;">
                                <input type="checkbox" class="field-required-checkbox" ${field.required ? 'checked' : ''}>
                                必填
                            </label>
                            <button type="button" class="btn btn-sm btn-danger" onclick="this.parentElement.parentElement.remove()">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                        <input type="text" class="field-desc-input mt-1" value="${field.description || ''}" placeholder="字段描述（可选）">
                    </div>
                `).join('') || ''}
            </div>
            <button type="button" class="btn btn-secondary mt-2" onclick="addEditFieldRow()">
                <i class="fas fa-plus"></i> 添加字段
            </button>
        </div>
    `, async () => {
        const name = document.getElementById('edit-type-name').value;
        const description = document.getElementById('edit-type-description').value;
        
        const fields = [];
        document.querySelectorAll('#edit-fields-container .field-row').forEach(row => {
            const fieldName = row.querySelector('.field-name-input').value;
            const fieldType = row.querySelector('.field-type-select').value;
            const fieldRequired = row.querySelector('.field-required-checkbox').checked;
            const fieldDesc = row.querySelector('.field-desc-input').value;
            
            if (fieldName) {
                fields.push({
                    name: fieldName,
                    type: fieldType,
                    required: fieldRequired,
                    description: fieldDesc
                });
            }
        });
        
        try {
            await apiRequest(`/custom-types/${id}`, {
                method: 'PUT',
                body: JSON.stringify({ name, description, fields })
            });
            showToast('类型更新成功', 'success');
            loadCustomTypes();
        } catch (error) {
            showToast('更新失败: ' + error.message, 'error');
        }
    });
}

function addEditFieldRow() {
    const container = document.getElementById('edit-fields-container');
    const row = document.createElement('div');
    row.className = 'field-row mb-2';
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="field-name-input" placeholder="字段名">
            <select class="field-type-select">
                <option value="string">string</option>
                <option value="number">number</option>
                <option value="boolean">boolean</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px;">
                <input type="checkbox" class="field-required-checkbox">
                必填
            </label>
            <button type="button" class="btn btn-sm btn-danger" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <input type="text" class="field-desc-input mt-1" placeholder="字段描述（可选）">
    `;
    container.appendChild(row);
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
    
    if (!interfaces || interfaces.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-plug"></i>
                <h3>还没有接口</h3>
                <p>点击"创建接口"开始</p>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = interfaces.map(iface => `
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

// 创建接口
document.getElementById('create-interface-btn').addEventListener('click', () => {
    const appId = document.getElementById('global-app-select').value;
    if (!appId) {
        showToast('请先选择一个应用', 'warning');
        return;
    }
    
    showModal('创建接口', `
        <div class="form-group">
            <label class="required">接口名称</label>
            <input type="text" id="interface-name" placeholder="例如: GetUser">
        </div>
        <div class="form-group">
            <label>接口描述</label>
            <textarea id="interface-description" rows="2" placeholder="描述接口的功能"></textarea>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label class="required">HTTP 方法</label>
                <select id="interface-method">
                    <option value="GET">GET</option>
                    <option value="POST">POST</option>
                    <option value="PUT">PUT</option>
                    <option value="DELETE">DELETE</option>
                    <option value="PATCH">PATCH</option>
                </select>
            </div>
            <div class="form-group">
                <label class="required">协议</label>
                <select id="interface-protocol">
                    <option value="http">HTTP</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label class="required">URL</label>
            <input type="text" id="interface-url" placeholder="https://api.example.com/users">
        </div>
        <div class="form-group">
            <label>认证类型</label>
            <select id="interface-auth">
                <option value="none">无认证</option>
                <option value="bearer">Bearer Token</option>
                <option value="basic">Basic Auth</option>
            </select>
        </div>
        <div class="form-group">
            <label>参数定义</label>
            <div id="params-container"></div>
            <button type="button" class="btn btn-secondary mt-2" onclick="addParamRow()">
                <i class="fas fa-plus"></i> 添加参数
            </button>
        </div>
    `, async () => {
        const name = document.getElementById('interface-name').value;
        const description = document.getElementById('interface-description').value;
        const method = document.getElementById('interface-method').value;
        const protocol = document.getElementById('interface-protocol').value;
        const url = document.getElementById('interface-url').value;
        const auth_type = document.getElementById('interface-auth').value;
        
        if (!name || !url) {
            showToast('请填写必填字段', 'error');
            return;
        }
        
        const parameters = [];
        document.querySelectorAll('.param-row').forEach(row => {
            const paramName = row.querySelector('.param-name-input').value;
            const paramType = row.querySelector('.param-type-select').value;
            const paramLocation = row.querySelector('.param-location-select').value;
            const paramRequired = row.querySelector('.param-required-checkbox').checked;
            
            if (paramName) {
                parameters.push({
                    name: paramName,
                    type: paramType,
                    location: paramLocation,
                    required: paramRequired
                });
            }
        });
        
        try {
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
                    parameters
                })
            });
            showToast('接口创建成功', 'success');
            loadInterfaces();
        } catch (error) {
            showToast('创建失败: ' + error.message, 'error');
        }
    });
});

function addParamRow() {
    const container = document.getElementById('params-container');
    const row = document.createElement('div');
    row.className = 'param-row mb-2';
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="param-name-input" placeholder="参数名">
            <select class="param-type-select">
                <option value="string">string</option>
                <option value="number">number</option>
                <option value="boolean">boolean</option>
            </select>
            <select class="param-location-select">
                <option value="query">query</option>
                <option value="header">header</option>
                <option value="body">body</option>
                <option value="path">path</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px;">
                <input type="checkbox" class="param-required-checkbox">
                必填
            </label>
            <button type="button" class="btn btn-sm btn-danger" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
    container.appendChild(row);
}

function viewInterface(id) {
    const iface = state.interfaces.find(i => i.id === id);
    if (!iface) return;
    
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
                                    <span>${param.type}</span>
                                    <span>${param.location}</span>
                                    ${param.required ? '<span class="badge-danger">必填</span>' : ''}
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            ` : '<p>暂无参数</p>'}
        </div>
    `, null, false);
}

function editInterface(id) {
    // 简化版编辑，完整版需要更复杂的表单
    showToast('编辑功能开发中...', 'info');
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
