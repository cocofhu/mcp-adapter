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
    
    grid.innerHTML = filteredApps.map(app => `
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
                
                // 如果是自定义类型，添加引用
                if (fieldType === 'custom' && fieldRef) {
                    field.ref = parseInt(fieldRef);
                }
                
                fields.push(field);
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
    
    // 构建类型选项（包含自定义类型）
    let typeOptions = `
        <option value="string">string</option>
        <option value="number">number</option>
        <option value="boolean">boolean</option>
    `;
    
    // 添加当前应用的自定义类型
    if (state.customTypes && state.customTypes.length > 0) {
        typeOptions += '<optgroup label="自定义类型">';
        state.customTypes.forEach(type => {
            typeOptions += `<option value="custom" data-ref="${type.id}">${type.name}</option>`;
        });
        typeOptions += '</optgroup>';
    }
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="field-name-input" placeholder="字段名">
            <select class="field-type-select" onchange="handleTypeChange(this)">
                ${typeOptions}
            </select>
            <input type="hidden" class="field-ref-input" value="">
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-array-checkbox">
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-required-checkbox">
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <input type="text" class="field-desc-input" placeholder="字段描述（可选）">
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
    const type = state.customTypes.find(t => t.id === id);
    if (!type) return;
    
    // 构建类型选项（包含自定义类型，排除当前编辑的类型）
    const buildTypeOptions = (currentType, currentRef) => {
        let options = `
            <option value="string" ${currentType === 'string' ? 'selected' : ''}>string</option>
            <option value="number" ${currentType === 'number' ? 'selected' : ''}>number</option>
            <option value="boolean" ${currentType === 'boolean' ? 'selected' : ''}>boolean</option>
        `;
        
        if (state.customTypes && state.customTypes.length > 0) {
            options += '<optgroup label="自定义类型">';
            state.customTypes.forEach(ct => {
                // 排除当前正在编辑的类型，避免循环引用
                if (ct.id !== type.id) {
                    const isSelected = currentType === 'custom' && currentRef === ct.id;
                    options += `<option value="custom" data-ref="${ct.id}" ${isSelected ? 'selected' : ''}>${ct.name}</option>`;
                }
            });
            options += '</optgroup>';
        }
        
        return options;
    };
    
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
                            <select class="field-type-select" onchange="handleTypeChange(this)">
                                ${buildTypeOptions(field.type, field.ref)}
                            </select>
                            <input type="hidden" class="field-ref-input" value="${field.ref || ''}">
                            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                                <input type="checkbox" class="field-array-checkbox" ${field.is_array ? 'checked' : ''}>
                                数组
                            </label>
                            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                                <input type="checkbox" class="field-required-checkbox" ${field.required ? 'checked' : ''}>
                                必填
                            </label>
                            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                        <input type="text" class="field-desc-input" value="${field.description || ''}" placeholder="字段描述（可选）">
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
                
                // 如果是自定义类型，添加引用
                if (fieldType === 'custom' && fieldRef) {
                    field.ref = parseInt(fieldRef);
                }
                
                fields.push(field);
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
    
    // 构建类型选项（包含自定义类型）
    let typeOptions = `
        <option value="string">string</option>
        <option value="number">number</option>
        <option value="boolean">boolean</option>
    `;
    
    // 添加当前应用的自定义类型
    if (state.customTypes && state.customTypes.length > 0) {
        typeOptions += '<optgroup label="自定义类型">';
        state.customTypes.forEach(type => {
            typeOptions += `<option value="custom" data-ref="${type.id}">${type.name}</option>`;
        });
        typeOptions += '</optgroup>';
    }
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="field-name-input" placeholder="字段名">
            <select class="field-type-select" onchange="handleTypeChange(this)">
                ${typeOptions}
            </select>
            <input type="hidden" class="field-ref-input" value="">
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-array-checkbox">
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="field-required-checkbox">
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <input type="text" class="field-desc-input" placeholder="字段描述（可选）">
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

// 创建接口
document.getElementById('create-interface-btn').addEventListener('click', async () => {
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
            const paramRef = row.querySelector('.param-ref-input')?.value;
            const paramLocation = row.querySelector('.param-location-select').value;
            const paramIsArray = row.querySelector('.param-array-checkbox')?.checked || false;
            const paramRequired = row.querySelector('.param-required-checkbox').checked;
            
            if (paramName) {
                const param = {
                    name: paramName,
                    type: paramType,
                    location: paramLocation,
                    is_array: paramIsArray,
                    required: paramRequired
                };
                
                // 如果是自定义类型，添加引用
                if (paramType === 'custom' && paramRef) {
                    param.ref = parseInt(paramRef);
                }
                
                parameters.push(param);
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
    
    // 构建类型选项（包含自定义类型）
    let typeOptions = `
        <option value="string">string</option>
        <option value="number">number</option>
        <option value="boolean">boolean</option>
    `;
    
    // 添加当前应用的自定义类型
    if (state.customTypes && state.customTypes.length > 0) {
        typeOptions += '<optgroup label="自定义类型">';
        state.customTypes.forEach(type => {
            typeOptions += `<option value="custom" data-ref="${type.id}">${type.name}</option>`;
        });
        typeOptions += '</optgroup>';
    }
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="param-name-input" placeholder="参数名">
            <select class="param-type-select" onchange="handleTypeChange(this)">
                ${typeOptions}
            </select>
            <input type="hidden" class="param-ref-input" value="">
            <select class="param-location-select">
                <option value="query">query</option>
                <option value="header">header</option>
                <option value="body">body</option>
                <option value="path">path</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-array-checkbox">
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-required-checkbox">
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
    container.appendChild(row);
}

function addEditParamRow() {
    const container = document.getElementById('edit-params-container');
    const row = document.createElement('div');
    row.className = 'param-row mb-2';
    
    // 构建类型选项（包含自定义类型）
    let typeOptions = `
        <option value="string">string</option>
        <option value="number">number</option>
        <option value="boolean">boolean</option>
    `;
    
    // 添加当前应用的自定义类型
    if (state.customTypes && state.customTypes.length > 0) {
        typeOptions += '<optgroup label="自定义类型">';
        state.customTypes.forEach(type => {
            typeOptions += `<option value="custom" data-ref="${type.id}">${type.name}</option>`;
        });
        typeOptions += '</optgroup>';
    }
    
    row.innerHTML = `
        <div class="form-row">
            <input type="text" class="param-name-input" placeholder="参数名">
            <select class="param-type-select" onchange="handleTypeChange(this)">
                ${typeOptions}
            </select>
            <input type="hidden" class="param-ref-input" value="">
            <select class="param-location-select">
                <option value="query">query</option>
                <option value="header">header</option>
                <option value="body">body</option>
                <option value="path">path</option>
            </select>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-array-checkbox">
                数组
            </label>
            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                <input type="checkbox" class="param-required-checkbox">
                必填
            </label>
            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
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

async function editInterface(id) {
    const iface = state.interfaces.find(i => i.id === id);
    if (!iface) return;
    
    // 确保已加载自定义类型
    const appId = document.getElementById('global-app-select').value;
    if (appId && (!state.customTypes || state.customTypes.length === 0)) {
        try {
            const types = await apiRequest(`/custom-types?app_id=${appId}`);
            state.customTypes = types;
        } catch (error) {
            console.error('加载自定义类型失败:', error);
        }
    }
    
    // 构建类型选项（包含自定义类型）
    const buildParamTypeOptions = (currentType, currentRef) => {
        let options = `
            <option value="string" ${currentType === 'string' ? 'selected' : ''}>string</option>
            <option value="number" ${currentType === 'number' ? 'selected' : ''}>number</option>
            <option value="boolean" ${currentType === 'boolean' ? 'selected' : ''}>boolean</option>
        `;
        
        if (state.customTypes && state.customTypes.length > 0) {
            options += '<optgroup label="自定义类型">';
            state.customTypes.forEach(type => {
                const isSelected = currentType === 'custom' && currentRef === type.id;
                options += `<option value="custom" data-ref="${type.id}" ${isSelected ? 'selected' : ''}>${type.name}</option>`;
            });
            options += '</optgroup>';
        }
        
        return options;
    };
    
    showModal('编辑接口', `
        <div class="form-group">
            <label>接口名称</label>
            <input type="text" id="edit-interface-name" value="${iface.name}">
        </div>
        <div class="form-group">
            <label>接口描述</label>
            <textarea id="edit-interface-description" rows="2">${iface.description || ''}</textarea>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label>HTTP 方法</label>
                <select id="edit-interface-method">
                    <option value="GET" ${iface.method === 'GET' ? 'selected' : ''}>GET</option>
                    <option value="POST" ${iface.method === 'POST' ? 'selected' : ''}>POST</option>
                    <option value="PUT" ${iface.method === 'PUT' ? 'selected' : ''}>PUT</option>
                    <option value="DELETE" ${iface.method === 'DELETE' ? 'selected' : ''}>DELETE</option>
                    <option value="PATCH" ${iface.method === 'PATCH' ? 'selected' : ''}>PATCH</option>
                </select>
            </div>
            <div class="form-group">
                <label>协议</label>
                <select id="edit-interface-protocol">
                    <option value="http" ${iface.protocol === 'http' ? 'selected' : ''}>HTTP</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label>URL</label>
            <input type="text" id="edit-interface-url" value="${iface.url}">
        </div>
        <div class="form-group">
            <label>认证类型</label>
            <select id="edit-interface-auth">
                <option value="none" ${iface.auth_type === 'none' ? 'selected' : ''}>无认证</option>
                <option value="bearer" ${iface.auth_type === 'bearer' ? 'selected' : ''}>Bearer Token</option>
                <option value="basic" ${iface.auth_type === 'basic' ? 'selected' : ''}>Basic Auth</option>
            </select>
        </div>
        <div class="form-group">
            <label>参数定义</label>
            <div id="edit-params-container">
                ${iface.parameters?.map(param => `
                    <div class="param-row mb-2">
                        <div class="form-row">
                            <input type="text" class="param-name-input" value="${param.name}" placeholder="参数名">
                            <select class="param-type-select" onchange="handleTypeChange(this)">
                                ${buildParamTypeOptions(param.type, param.ref)}
                            </select>
                            <input type="hidden" class="param-ref-input" value="${param.ref || ''}">
                            <select class="param-location-select">
                                <option value="query" ${param.location === 'query' ? 'selected' : ''}>query</option>
                                <option value="header" ${param.location === 'header' ? 'selected' : ''}>header</option>
                                <option value="body" ${param.location === 'body' ? 'selected' : ''}>body</option>
                                <option value="path" ${param.location === 'path' ? 'selected' : ''}>path</option>
                            </select>
                            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                                <input type="checkbox" class="param-array-checkbox" ${param.is_array ? 'checked' : ''}>
                                数组
                            </label>
                            <label style="display: flex; align-items: center; gap: 4px; white-space: nowrap;">
                                <input type="checkbox" class="param-required-checkbox" ${param.required ? 'checked' : ''}>
                                必填
                            </label>
                            <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()" title="删除">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                    </div>
                `).join('') || ''}
            </div>
            <button type="button" class="btn btn-secondary mt-2" onclick="addEditParamRow()">
                <i class="fas fa-plus"></i> 添加参数
            </button>
        </div>
    `, async () => {
        const name = document.getElementById('edit-interface-name').value;
        const description = document.getElementById('edit-interface-description').value;
        const method = document.getElementById('edit-interface-method').value;
        const protocol = document.getElementById('edit-interface-protocol').value;
        const url = document.getElementById('edit-interface-url').value;
        const auth_type = document.getElementById('edit-interface-auth').value;
        
        if (!name || !url) {
            showToast('请填写必填字段', 'error');
            return;
        }
        
        const parameters = [];
        document.querySelectorAll('#edit-params-container .param-row').forEach(row => {
            const paramName = row.querySelector('.param-name-input').value;
            const paramType = row.querySelector('.param-type-select').value;
            const paramRef = row.querySelector('.param-ref-input')?.value;
            const paramLocation = row.querySelector('.param-location-select').value;
            const paramIsArray = row.querySelector('.param-array-checkbox')?.checked || false;
            const paramRequired = row.querySelector('.param-required-checkbox').checked;
            
            if (paramName) {
                const param = {
                    name: paramName,
                    type: paramType,
                    location: paramLocation,
                    is_array: paramIsArray,
                    required: paramRequired
                };
                
                // 如果是自定义类型，添加引用
                if (paramType === 'custom' && paramRef) {
                    param.ref = parseInt(paramRef);
                }
                
                parameters.push(param);
            }
        });
        
        try {
            await apiRequest(`/interfaces/${id}`, {
                method: 'PUT',
                body: JSON.stringify({
                    name,
                    description,
                    method,
                    protocol,
                    url,
                    auth_type,
                    parameters
                })
            });
            showToast('接口更新成功', 'success');
            loadInterfaces();
        } catch (error) {
            showToast('更新失败: ' + error.message, 'error');
        }
    });
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
