// API 基础配置
const API_BASE_URL = 'http://192.168.2.17:8080/api';

// 全局状态管理
class MCPAdapterApp {
    constructor() {
        this.applications = [];
        this.interfaces = [];
        this.currentTab = 'applications';
        this.currentAppId = null;
        this.editingInterfaceId = null;
        this.init();
    }

    async init() {
        await this.loadApplications();
        await this.loadInterfaces();
        this.bindEvents();
        this.renderApplications();
        this.updateAppSelectors();
        
        // 初始化参数信息显示和UI状态
        const protocolType = document.getElementById('protocol-type');
        if (protocolType && protocolType.value) {
            protocolType.dispatchEvent(new Event('change'));
        }
    }

    // API 调用方法
    async apiCall(endpoint, method = 'GET', data = null) {
        const config = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };

        if (data) {
            config.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // 如果是 DELETE 请求且返回 204，直接返回 true
            if (method === 'DELETE' && response.status === 204) {
                return true;
            }

            return await response.json();
        } catch (error) {
            console.error('API call failed:', error);
            this.showNotification('操作失败: ' + error.message, 'error');
            throw error;
        }
    }

    // 加载应用数据
    async loadApplications() {
        try {
            this.applications = await this.apiCall('/applications');
        } catch (error) {
            console.error('Failed to load applications:', error);
            this.applications = [];
        }
    }

    // 加载接口数据
    async loadInterfaces(appId = null) {
        try {
            const endpoint = appId ? `/interfaces?app_id=${appId}` : '/interfaces';
            this.interfaces = await this.apiCall(endpoint);
        } catch (error) {
            console.error('Failed to load interfaces:', error);
            this.interfaces = [];
        }
    }

    // 创建应用
    async createApplication(appData) {
        try {
            const newApp = await this.apiCall('/applications', 'POST', {
                name: appData.name,
                description: appData.description,
                path: appData.path || '',
                protocol: appData.protocol || 'http',
                post_process: appData.postProcess || '',
                environment: appData.environment || '{}',
                enabled: appData.enabled !== false
            });
            
            await this.loadApplications();
            this.renderApplications();
            this.updateAppSelectors();
            this.showNotification('应用创建成功', 'success');
            return newApp;
        } catch (error) {
            console.error('Failed to create application:', error);
            throw error;
        }
    }

    // 更新应用
    async updateApplication(id, appData) {
        try {
            const updatedApp = await this.apiCall(`/applications/${id}`, 'PUT', {
                name: appData.name,
                description: appData.description,
                path: appData.path || '',
                protocol: appData.protocol || 'http',
                post_process: appData.postProcess || '',
                environment: appData.environment || '{}',
                enabled: appData.enabled !== false
            });
            
            await this.loadApplications();
            this.renderApplications();
            this.updateAppSelectors();
            this.showNotification('应用更新成功', 'success');
            return updatedApp;
        } catch (error) {
            console.error('Failed to update application:', error);
            throw error;
        }
    }

    // 删除应用
    async deleteApplication(id) {
        try {
            await this.apiCall(`/applications/${id}`, 'DELETE');
            await this.loadApplications();
            await this.loadInterfaces(); // 重新加载接口，因为关联接口可能被删除
            this.renderApplications();
            this.updateAppSelectors();
            this.showNotification('应用删除成功', 'success');
        } catch (error) {
            console.error('Failed to delete application:', error);
            throw error;
        }
    }

    // 创建接口
    async createInterface(interfaceData) {
        try {
            const newInterface = await this.apiCall('/interfaces', 'POST', {
                app_id: parseInt(interfaceData.appId),
                name: interfaceData.name,
                description: interfaceData.description || '',
                protocol: interfaceData.protocol || 'HTTP',
                url: interfaceData.url,
                auth_type: interfaceData.authType || 'none',
                enabled: interfaceData.enabled !== false,
                post_process: interfaceData.postProcess || '',
                options: JSON.stringify({
                    method: interfaceData.method || 'GET',
                    parameters: interfaceData.parameters || [],
                    defaultParams: interfaceData.defaultParams || [],
                    defaultHeaders: interfaceData.defaultHeaders || []
                })
            });
            
            await this.loadInterfaces();
            this.renderInterfaces();
            this.showNotification('接口创建成功', 'success');
            return newInterface;
        } catch (error) {
            console.error('Failed to create interface:', error);
            throw error;
        }
    }

    // 更新接口
    async updateInterface(id, interfaceData) {
        try {
            const updatedInterface = await this.apiCall(`/interfaces/${id}`, 'PUT', {
                app_id: parseInt(interfaceData.appId),
                name: interfaceData.name,
                description: interfaceData.description || '',
                protocol: interfaceData.protocol || 'HTTP',
                url: interfaceData.url,
                auth_type: interfaceData.authType || 'none',
                enabled: interfaceData.enabled !== false,
                post_process: interfaceData.postProcess || '',
                options: JSON.stringify({
                    method: interfaceData.method || 'GET',
                    parameters: interfaceData.parameters || [],
                    defaultParams: interfaceData.defaultParams || [],
                    defaultHeaders: interfaceData.defaultHeaders || []
                })
            });
            
            await this.loadInterfaces();
            this.renderInterfaces();
            this.showNotification('接口更新成功', 'success');
            return updatedInterface;
        } catch (error) {
            console.error('Failed to update interface:', error);
            throw error;
        }
    }

    // 删除接口
    async deleteInterface(id) {
        try {
            await this.apiCall(`/interfaces/${id}`, 'DELETE');
            await this.loadInterfaces();
            this.renderInterfaces();
            this.showNotification('接口删除成功', 'success');
        } catch (error) {
            console.error('Failed to delete interface:', error);
            throw error;
        }
    }

    // 显示通知
    showNotification(message, type = 'info') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        
        // 创建消息文本节点
        const messageSpan = document.createElement('span');
        messageSpan.textContent = message;
        notification.appendChild(messageSpan);
        
        // 添加到页面
        document.body.appendChild(notification);
        
        // 3秒后开始淡出动画
        setTimeout(() => {
            if (notification.parentNode) {
                notification.style.animation = 'slideOut 0.3s ease-in forwards';
                setTimeout(() => {
                    if (notification.parentNode) {
                        notification.parentNode.removeChild(notification);
                    }
                }, 300);
            }
        }, 3000);
    }

    // 绑定事件
    bindEvents() {
        // 导航切换
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                const tab = e.currentTarget.dataset.tab;
                this.switchTab(tab);
            });
        });

        // 创建应用按钮
        const createAppBtn = document.getElementById('create-app-btn');
        if (createAppBtn) {
            createAppBtn.addEventListener('click', () => this.showCreateAppModal());
        }

        // 创建应用模态框事件
        const createAppConfirmBtn = document.getElementById('create-app-confirm-btn');
        if (createAppConfirmBtn) {
            createAppConfirmBtn.addEventListener('click', () => this.handleCreateApp());
        }

        const createAppCancelBtn = document.getElementById('create-app-cancel-btn');
        if (createAppCancelBtn) {
            createAppCancelBtn.addEventListener('click', () => this.hideCreateAppModal());
        }

        const createAppCloseBtn = document.getElementById('create-app-close-btn');
        if (createAppCloseBtn) {
            createAppCloseBtn.addEventListener('click', () => this.hideCreateAppModal());
        }

        // 接口表单提交
        const addInterfaceForm = document.getElementById('add-interface-form');
        if (addInterfaceForm) {
            addInterfaceForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleSaveInterface();
            });
        }

        // 应用选择器变化
        const currentAppSelect = document.getElementById('current-app-select');
        if (currentAppSelect) {
            currentAppSelect.addEventListener('change', (e) => {
                this.currentAppId = e.target.value;
                this.loadInterfaces(this.currentAppId);
                this.renderInterfaces();
            });
        }

        // 搜索功能
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            searchInput.addEventListener('input', () => this.renderInterfaces());
        }

        const appSearchInput = document.getElementById('app-search-input');
        if (appSearchInput) {
            appSearchInput.addEventListener('input', () => this.renderApplications());
        }

        // 过滤器
        const protocolFilter = document.getElementById('protocol-filter');
        if (protocolFilter) {
            protocolFilter.addEventListener('change', () => this.renderInterfaces());
        }

        const statusFilter = document.getElementById('status-filter');
        if (statusFilter) {
            statusFilter.addEventListener('change', () => this.renderInterfaces());
        }

        // 参数管理按钮
        const addParameterBtn = document.getElementById('add-parameter-btn');
        if (addParameterBtn) {
            addParameterBtn.addEventListener('click', () => this.addParameter());
        }

        const addDefaultParamBtn = document.getElementById('add-default-param-btn');
        if (addDefaultParamBtn) {
            addDefaultParamBtn.addEventListener('click', () => this.addDefaultParameter());
        }

        const addDefaultHeaderBtn = document.getElementById('add-default-header-btn');
        if (addDefaultHeaderBtn) {
            addDefaultHeaderBtn.addEventListener('click', () => this.addDefaultHeader());
        }

        // 协议类型变化
        const protocolType = document.getElementById('protocol-type');
        if (protocolType) {
            protocolType.addEventListener('change', () => this.updateParameterLocationInfo());
        }
    }

    // 切换标签页
    switchTab(tabName) {
        // 更新导航状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // 更新内容区域
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');

        // 更新页面标题
        const titles = {
            'applications': '应用管理',
            'interfaces': '接口列表',
            'add-interface': '添加接口'
        };
        document.getElementById('page-title').textContent = titles[tabName];

        this.currentTab = tabName;

        // 根据标签页加载相应数据
        if (tabName === 'interfaces') {
            this.renderInterfaces();
        } else if (tabName === 'applications') {
            this.renderApplications();
        }
    }

    // 渲染应用列表
    renderApplications() {
        const container = document.getElementById('applications-grid');
        if (!container) return;

        const searchTerm = document.getElementById('app-search-input')?.value.toLowerCase() || '';
        
        const filteredApps = this.applications.filter(app => 
            app.name.toLowerCase().includes(searchTerm) ||
            (app.description && app.description.toLowerCase().includes(searchTerm))
        );

        if (filteredApps.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-folder-open"></i>
                    <h3>暂无应用</h3>
                    <p>点击"创建应用"按钮开始添加您的第一个应用</p>
                </div>
            `;
            return;
        }

        container.innerHTML = filteredApps.map(app => `
            <div class="application-card" data-app-id="${app.id}">
                <div class="app-header">
                    <div class="app-info">
                        <h3 class="app-name">${app.name}</h3>
                        <span class="app-version">${app.path || 'v1.0.0'}</span>
                    </div>
                    <div class="app-actions">
                        <button class="btn-icon" onclick="app.editApplication(${app.id})" title="编辑">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn-icon" onclick="app.deleteApplicationConfirm(${app.id})" title="删除">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
                <div class="app-description">
                    ${app.description || '暂无描述'}
                </div>
                <div class="app-stats">
                    <div class="stat-item">
                        <i class="fas fa-list"></i>
                        <span>接口数量: ${this.getInterfaceCountForApp(app.id)}</span>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-calendar"></i>
                        <span>创建时间: ${new Date(app.created_at).toLocaleDateString()}</span>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-${app.enabled ? 'check-circle' : 'times-circle'}"></i>
                        <span>${app.enabled ? '已启用' : '已禁用'}</span>
                    </div>
                </div>
                <div class="app-footer">
                    <button class="btn btn-primary btn-sm" onclick="app.viewAppInterfaces(${app.id})">
                        查看接口
                    </button>
                </div>
            </div>
        `).join('');
    }

    // 获取应用的接口数量
    getInterfaceCountForApp(appId) {
        return this.interfaces.filter(iface => iface.app_id === appId).length;
    }

    // 查看应用的接口
    viewAppInterfaces(appId) {
        this.currentAppId = appId;
        const appSelect = document.getElementById('current-app-select');
        if (appSelect) {
            appSelect.value = appId;
        }
        this.loadInterfaces(appId);
        this.switchTab('interfaces');
    }

    // 渲染接口列表
    renderInterfaces() {
        const container = document.getElementById('interfaces-grid');
        if (!container) return;

        const searchTerm = document.getElementById('search-input')?.value.toLowerCase() || '';
        const protocolFilter = document.getElementById('protocol-filter')?.value || '';
        const statusFilter = document.getElementById('status-filter')?.value || '';
        
        let filteredInterfaces = this.interfaces.filter(iface => {
            const matchesSearch = iface.name.toLowerCase().includes(searchTerm) ||
                                (iface.description && iface.description.toLowerCase().includes(searchTerm));
            
            const matchesProtocol = !protocolFilter || iface.protocol.toLowerCase() === protocolFilter;
            
            let matchesStatus = true;
            if (statusFilter) {
                switch (statusFilter) {
                    case 'enabled':
                        matchesStatus = iface.enabled;
                        break;
                    case 'disabled':
                        matchesStatus = !iface.enabled;
                        break;
                    default:
                        matchesStatus = true;
                }
            }
            
            return matchesSearch && matchesProtocol && matchesStatus;
        });

        if (filteredInterfaces.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-list"></i>
                    <h3>暂无接口</h3>
                    <p>点击"添加接口"开始创建您的第一个接口</p>
                </div>
            `;
            return;
        }

        container.innerHTML = filteredInterfaces.map(iface => {
            const app = this.applications.find(a => a.id === iface.app_id);
            const options = iface.options ? JSON.parse(iface.options) : {};
            
            return `
                <div class="interface-card" data-interface-id="${iface.id}">
                    <div class="interface-header">
                        <div class="interface-info">
                            <h3 class="interface-name">${iface.name}</h3>
                            <span class="interface-protocol">${iface.protocol}</span>
                            <span class="interface-method">${options.method || 'GET'}</span>
                        </div>
                        <div class="interface-status ${iface.enabled ? 'enabled' : 'disabled'}">
                            <i class="fas fa-${iface.enabled ? 'check-circle' : 'times-circle'}"></i>
                            ${iface.enabled ? '已启用' : '已禁用'}
                        </div>
                    </div>
                    <div class="interface-description">
                        ${iface.description || '暂无描述'}
                    </div>
                    <div class="interface-url">
                        <i class="fas fa-link"></i>
                        <span>${iface.url}</span>
                    </div>
                    <div class="interface-app">
                        <i class="fas fa-folder"></i>
                        <span>所属应用: ${app ? app.name : '未知应用'}</span>
                    </div>
                    <div class="interface-actions">
                        <button class="btn btn-secondary btn-sm" onclick="app.editInterface(${iface.id})">
                            <i class="fas fa-edit"></i> 编辑
                        </button>
                        <button class="btn btn-danger btn-sm" onclick="app.deleteInterfaceConfirm(${iface.id})">
                            <i class="fas fa-trash"></i> 删除
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    }

    // 更新应用选择器
    updateAppSelectors() {
        const selectors = ['current-app-select', 'interface-app-id'];
        
        selectors.forEach(selectorId => {
            const selector = document.getElementById(selectorId);
            if (selector) {
                const currentValue = selector.value;
                selector.innerHTML = '<option value="">请选择应用</option>' +
                    this.applications.map(app => 
                        `<option value="${app.id}">${app.name}</option>`
                    ).join('');
                
                // 恢复之前的选择
                if (currentValue) {
                    selector.value = currentValue;
                }
            }
        });
    }

    // 显示创建应用模态框
    showCreateAppModal() {
        const modal = document.getElementById('create-app-modal');
        if (modal) {
            modal.style.display = 'flex';
            // 清空表单
            document.getElementById('create-app-form').reset();
        }
    }

    // 隐藏创建应用模态框
    hideCreateAppModal() {
        const modal = document.getElementById('create-app-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    }

    // 处理创建应用
    async handleCreateApp() {
        const form = document.getElementById('create-app-form');
        const formData = new FormData(form);
        
        const appData = {
            name: document.getElementById('app-name').value,
            description: document.getElementById('app-description').value,
            path: document.getElementById('app-version').value,
            protocol: 'http',
            enabled: true
        };

        if (!appData.name) {
            this.showNotification('请输入应用名称', 'error');
            return;
        }

        try {
            await this.createApplication(appData);
            this.hideCreateAppModal();
        } catch (error) {
            // 错误已在 createApplication 中处理
        }
    }

    // 编辑应用
    editApplication(id) {
        const app = this.applications.find(a => a.id === id);
        if (!app) return;

        // 填充表单数据
        document.getElementById('app-name').value = app.name;
        document.getElementById('app-description').value = app.description || '';
        document.getElementById('app-version').value = app.path || '';
        
        // 更新模态框标题和按钮
        document.getElementById('create-app-modal-title').textContent = '编辑应用';
        const confirmBtn = document.getElementById('create-app-confirm-btn');
        confirmBtn.textContent = '更新应用';
        confirmBtn.onclick = () => this.handleUpdateApp(id);
        
        this.showCreateAppModal();
    }

    // 处理更新应用
    async handleUpdateApp(id) {
        const appData = {
            name: document.getElementById('app-name').value,
            description: document.getElementById('app-description').value,
            path: document.getElementById('app-version').value,
            protocol: 'http',
            enabled: true
        };

        if (!appData.name) {
            this.showNotification('请输入应用名称', 'error');
            return;
        }

        try {
            await this.updateApplication(id, appData);
            this.hideCreateAppModal();
            // 重置模态框
            this.resetCreateAppModal();
        } catch (error) {
            // 错误已在 updateApplication 中处理
        }
    }

    // 重置创建应用模态框
    resetCreateAppModal() {
        document.getElementById('create-app-modal-title').textContent = '创建应用';
        const confirmBtn = document.getElementById('create-app-confirm-btn');
        confirmBtn.textContent = '创建应用';
        confirmBtn.onclick = () => this.handleCreateApp();
    }

    // 删除应用确认
    deleteApplicationConfirm(id) {
        const app = this.applications.find(a => a.id === id);
        if (!app) return;

        if (confirm(`确定要删除应用"${app.name}"吗？这将同时删除该应用下的所有接口。`)) {
            this.deleteApplication(id);
        }
    }

    // 处理保存接口
    async handleSaveInterface() {
        const interfaceData = this.collectInterfaceFormData();
        
        if (!this.validateInterfaceData(interfaceData)) {
            return;
        }

        try {
            if (this.editingInterfaceId) {
                await this.updateInterface(this.editingInterfaceId, interfaceData);
                this.editingInterfaceId = null;
            } else {
                await this.createInterface(interfaceData);
            }
            
            // 清空表单
            document.getElementById('add-interface-form').reset();
            this.clearParameters();
            
            // 切换到接口列表
            this.switchTab('interfaces');
        } catch (error) {
            // 错误已在相应方法中处理
        }
    }

    // 收集接口表单数据
    collectInterfaceFormData() {
        return {
            appId: document.getElementById('interface-app-id').value,
            name: document.getElementById('interface-name').value,
            description: document.getElementById('interface-description').value,
            protocol: document.getElementById('protocol-type').value,
            method: document.getElementById('http-method').value,
            url: document.getElementById('endpoint-url').value,
            authType: document.getElementById('auth-type').value,
            enabled: true,
            parameters: this.collectParameters(),
            defaultParams: this.collectDefaultParams(),
            defaultHeaders: this.collectDefaultHeaders()
        };
    }

    // 验证接口数据
    validateInterfaceData(data) {
        if (!data.appId) {
            this.showNotification('请选择所属应用', 'error');
            return false;
        }
        
        if (!data.name) {
            this.showNotification('请输入接口名称', 'error');
            return false;
        }
        
        if (!data.url) {
            this.showNotification('请输入端点URL', 'error');
            return false;
        }
        
        return true;
    }

    // 编辑接口
    editInterface(id) {
        const iface = this.interfaces.find(i => i.id === id);
        if (!iface) return;

        this.editingInterfaceId = id;
        
        // 填充基本信息
        document.getElementById('interface-app-id').value = iface.app_id;
        document.getElementById('interface-name').value = iface.name;
        document.getElementById('interface-description').value = iface.description || '';
        document.getElementById('protocol-type').value = iface.protocol;
        document.getElementById('endpoint-url').value = iface.url;
        document.getElementById('auth-type').value = iface.auth_type || 'none';

        // 解析选项
        const options = iface.options ? JSON.parse(iface.options) : {};
        document.getElementById('http-method').value = options.method || 'GET';

        // 填充参数
        this.loadParametersToForm(options.parameters || []);
        this.loadDefaultParamsToForm(options.defaultParams || []);
        this.loadDefaultHeadersToForm(options.defaultHeaders || []);

        // 切换到添加接口页面
        this.switchTab('add-interface');
        
        // 更新页面标题
        document.getElementById('page-title').textContent = '编辑接口';
    }

    // 删除接口确认
    deleteInterfaceConfirm(id) {
        const iface = this.interfaces.find(i => i.id === id);
        if (!iface) return;

        if (confirm(`确定要删除接口"${iface.name}"吗？`)) {
            this.deleteInterface(id);
        }
    }

    // 收集参数
    collectParameters() {
        const parameters = [];
        document.querySelectorAll('#parameters-list .parameter-item').forEach(item => {
            const name = item.querySelector('.param-name').value;
            const type = item.querySelector('.param-type').value;
            const required = item.querySelector('.param-required').checked;
            const description = item.querySelector('.param-description').value;
            
            if (name) {
                parameters.push({ name, type, required, description });
            }
        });
        return parameters;
    }

    // 收集默认参数
    collectDefaultParams() {
        const params = [];
        document.querySelectorAll('#default-params-list .parameter-item').forEach(item => {
            const name = item.querySelector('.param-name').value;
            const value = item.querySelector('.param-value').value;
            const description = item.querySelector('.param-description').value;
            
            if (name) {
                params.push({ name, value, description });
            }
        });
        return params;
    }

    // 收集默认请求头
    collectDefaultHeaders() {
        const headers = [];
        document.querySelectorAll('#default-headers-list .parameter-item').forEach(item => {
            const name = item.querySelector('.param-name').value;
            const value = item.querySelector('.param-value').value;
            const description = item.querySelector('.param-description').value;
            
            if (name) {
                headers.push({ name, value, description });
            }
        });
        return headers;
    }

    // 添加参数
    addParameter() {
        const container = document.getElementById('parameters-list');
        const parameterHtml = `
            <div class="parameter-item">
                <div class="parameter-row">
                    <input type="text" class="param-name" placeholder="参数名称" required>
                    <select class="param-type">
                        <option value="string">字符串</option>
                        <option value="integer">整数</option>
                        <option value="number">数字</option>
                        <option value="boolean">布尔值</option>
                    </select>
                    <label class="param-required-label">
                        <input type="checkbox" class="param-required"> 必需
                    </label>
                    <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <input type="text" class="param-description" placeholder="参数描述">
            </div>
        `;
        container.insertAdjacentHTML('beforeend', parameterHtml);
    }

    // 添加默认参数
    addDefaultParameter() {
        const container = document.getElementById('default-params-list');
        const parameterHtml = `
            <div class="parameter-item">
                <div class="parameter-row">
                    <input type="text" class="param-name" placeholder="参数名称" required>
                    <input type="text" class="param-value" placeholder="参数值" required>
                    <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <input type="text" class="param-description" placeholder="参数描述">
            </div>
        `;
        container.insertAdjacentHTML('beforeend', parameterHtml);
    }

    // 添加默认请求头
    addDefaultHeader() {
        const container = document.getElementById('default-headers-list');
        const headerHtml = `
            <div class="parameter-item">
                <div class="parameter-row">
                    <input type="text" class="param-name" placeholder="请求头名称" required>
                    <input type="text" class="param-value" placeholder="请求头值" required>
                    <button type="button" class="btn-remove" onclick="this.parentElement.parentElement.remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <input type="text" class="param-description" placeholder="请求头描述">
            </div>
        `;
        container.insertAdjacentHTML('beforeend', headerHtml);
    }

    // 清空参数
    clearParameters() {
        document.getElementById('parameters-list').innerHTML = '';
        document.getElementById('default-params-list').innerHTML = '';
        document.getElementById('default-headers-list').innerHTML = '';
    }

    // 加载参数到表单
    loadParametersToForm(parameters) {
        const container = document.getElementById('parameters-list');
        container.innerHTML = '';
        
        parameters.forEach(param => {
            this.addParameter();
            const item = container.lastElementChild;
            item.querySelector('.param-name').value = param.name;
            item.querySelector('.param-type').value = param.type;
            item.querySelector('.param-required').checked = param.required;
            item.querySelector('.param-description').value = param.description || '';
        });
    }

    // 加载默认参数到表单
    loadDefaultParamsToForm(params) {
        const container = document.getElementById('default-params-list');
        container.innerHTML = '';
        
        params.forEach(param => {
            this.addDefaultParameter();
            const item = container.lastElementChild;
            item.querySelector('.param-name').value = param.name;
            item.querySelector('.param-value').value = param.value;
            item.querySelector('.param-description').value = param.description || '';
        });
    }

    // 加载默认请求头到表单
    loadDefaultHeadersToForm(headers) {
        const container = document.getElementById('default-headers-list');
        container.innerHTML = '';
        
        headers.forEach(header => {
            this.addDefaultHeader();
            const item = container.lastElementChild;
            item.querySelector('.param-name').value = header.name;
            item.querySelector('.param-value').value = header.value;
            item.querySelector('.param-description').value = header.description || '';
        });
    }

    // 切换参数标签页
    switchParameterTab(tabName) {
        // 更新标签按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // 更新内容区域
        document.querySelectorAll('.parameters-container').forEach(container => {
            container.classList.remove('active');
        });
        document.getElementById(tabName).classList.add('active');
    }

    // 更新参数位置信息
    updateParameterLocationInfo() {
        const protocolType = document.getElementById('protocol-type').value;
        const locationConfig = document.getElementById('parameter-location-config');
        const paramInfo = document.getElementById('input-param-info');
        
        if (protocolType === 'http') {
            locationConfig.style.display = 'block';
            paramInfo.innerHTML = '<small>HTTP协议支持URL查询参数和请求体参数</small>';
        } else {
            locationConfig.style.display = 'none';
            paramInfo.innerHTML = '<small>参数位置将根据协议类型自动确定</small>';
        }
    }
}

// 初始化应用
const app = new MCPAdapterApp();

// 添加通知样式
const notificationStyles = `
    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 16px 20px;
        border-radius: 8px;
        color: white;
        font-weight: 500;
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        display: flex;
        align-items: center;
        gap: 8px;
        min-width: 300px;
        max-width: 500px;
    }
    
    .notification-success {
        background: linear-gradient(135deg, #10b981, #059669);
    }
    
    .notification-error {
        background: linear-gradient(135deg, #ef4444, #dc2626);
    }
    
    .notification-info {
        background: linear-gradient(135deg, #3b82f6, #1d4ed8);
    }
    
    .notification::before {
        content: "✓";
        font-weight: bold;
        font-size: 16px;
    }
    
    .notification-error::before {
        content: "✕";
    }
    
    .notification-info::before {
        content: "ℹ";
    }
    
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;

// 添加样式到页面
const styleSheet = document.createElement('style');
styleSheet.textContent = notificationStyles;
document.head.appendChild(styleSheet);