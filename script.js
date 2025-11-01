// 全局状态管理
class MCPAdapterApp {
    constructor() {
        this.applications = [];
        this.interfaces = [];
        this.currentTab = 'applications';
        this.currentAppId = null; // 当前选中的应用ID
        this.editingInterfaceId = null; // 当前编辑的接口ID
        this.init();
    }

    init() {
        this.loadSampleData();
        this.bindEvents();
        this.renderApplications();
        this.updateAppSelectors();
        
        // 初始化参数信息显示和UI状态
        const protocolType = document.getElementById('protocol-type');
        if (protocolType && protocolType.value) {
            // 触发协议类型变化事件来初始化UI
            protocolType.dispatchEvent(new Event('change'));
        }
    }

    // 加载示例数据
    loadSampleData() {
        // 示例应用数据
        this.applications = [
            {
                id: 1,
                name: '天气服务API',
                description: '提供全球天气信息查询服务，包括实时天气、天气预报等功能',
                version: 'v1.0.0',
                baseUrl: 'https://api.openweathermap.org',
                createdAt: new Date('2024-01-15'),
                interfaceCount: 2
            },
            {
                id: 2,
                name: '用户管理系统',
                description: '用户注册、登录、信息管理等核心功能接口',
                version: 'v2.1.0',
                baseUrl: 'https://api.example.com',
                createdAt: new Date('2024-01-10'),
                interfaceCount: 1
            }
        ];

        // 示例接口数据
        this.interfaces = [
            {
                id: 1,
                appId: 1, // 所属应用ID
                name: 'weather_api',
                description: '获取天气信息的API接口',
                protocol: 'http',
                method: 'GET',
                url: 'https://api.openweathermap.org/data/2.5/weather',
                authType: 'api-key',
                authValue: 'your-api-key',
                status: 'active',
                enabled: true,
                httpParamLocation: 'query',
                parameters: [
                    { name: 'q', type: 'string', required: true, location: 'query', description: '城市名称' },
                    { name: 'appid', type: 'string', required: true, location: 'query', description: 'API密钥' },
                    { name: 'units', type: 'string', required: false, location: 'query', description: '单位制' }
                ],
                defaultParams: [
                    { name: 'lang', value: 'zh_cn', description: '语言设置' },
                    { name: 'mode', value: 'json', description: '响应格式' }
                ],
                defaultHeaders: [
                    { name: 'User-Agent', value: 'WeatherApp/1.0', description: '用户代理' },
                    { name: 'Accept', value: 'application/json', description: '接受的内容类型' }
                ],
                createdAt: new Date('2024-01-15')
            },
            {
                id: 2,
                appId: 1, // 所属应用ID
                name: 'forecast_api',
                description: '获取天气预报的API接口',
                protocol: 'http',
                method: 'GET',
                url: 'https://api.openweathermap.org/data/2.5/forecast',
                authType: 'api-key',
                authValue: 'your-api-key',
                status: 'active',
                enabled: true,
                httpParamLocation: 'query',
                parameters: [
                    { name: 'q', type: 'string', required: true, location: 'query', description: '城市名称' },
                    { name: 'appid', type: 'string', required: true, location: 'query', description: 'API密钥' },
                    { name: 'cnt', type: 'integer', required: false, location: 'query', description: '预报天数' }
                ],
                defaultParams: [
                    { name: 'lang', value: 'zh_cn', description: '语言设置' },
                    { name: 'units', value: 'metric', description: '单位制' }
                ],
                defaultHeaders: [
                    { name: 'User-Agent', value: 'WeatherApp/1.0', description: '用户代理' }
                ],
                createdAt: new Date('2024-01-16')
            },
            {
                id: 3,
                appId: 2, // 所属应用ID
                name: 'user_api',
                description: '用户信息API接口',
                protocol: 'http',
                method: 'POST',
                url: 'https://api.example.com/users',
                authType: 'bearer',
                authValue: 'eyJhbGciOiJIUzI1NiIs...',
                status: 'inactive',
                enabled: false,
                httpParamLocation: 'body',
                parameters: [
                    { name: 'user_id', type: 'string', required: true, location: 'body', description: '用户ID' },
                    { name: 'include_profile', type: 'boolean', required: false, location: 'body', description: '是否包含详细信息' }
                ],
                defaultParams: [
                    { name: 'version', value: 'v1', description: 'API版本' }
                ],
                defaultHeaders: [
                    { name: 'Content-Type', value: 'application/json', description: '内容类型' },
                    { name: 'X-Client-Version', value: '1.0.0', description: '客户端版本' }
                ],
                createdAt: new Date('2024-01-10')
            }
        ];


    }

    // 绑定事件
    bindEvents() {
        // 导航菜单切换
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                const tab = e.currentTarget.dataset.tab;
                if (tab === 'add-interface') {
                    if (this.applications.length === 0) {
                        this.showNotification('请先创建应用', 'warning');
                        this.switchTab('applications');
                        return;
                    }
                    this.editingInterfaceId = null; // 清除编辑状态
                    this.resetForm(); // 重置表单
                }
                this.switchTab(tab);
            });
        });

        // 快速添加按钮
        document.getElementById('quick-add-btn').addEventListener('click', () => {
            if (this.applications.length === 0) {
                this.showNotification('请先创建应用', 'warning');
                this.switchTab('applications');
                return;
            }
            this.editingInterfaceId = null; // 清除编辑状态
            this.resetForm(); // 重置表单
            this.switchTab('add-interface');
        });

        // 搜索功能
        document.getElementById('search-input').addEventListener('input', (e) => {
            this.filterInterfaces(e.target.value);
        });

        // 协议过滤
        document.getElementById('protocol-filter').addEventListener('change', (e) => {
            this.filterByProtocol(e.target.value);
        });

        // 状态过滤
        document.getElementById('status-filter').addEventListener('change', (e) => {
            this.filterByStatus(e.target.value);
        });

        // 添加接口表单
        document.getElementById('add-interface-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addInterface();
        });

        // 添加参数按钮
        document.getElementById('add-parameter-btn').addEventListener('click', () => {
            this.addParameter();
        });

        // 添加默认参数按钮
        document.getElementById('add-default-param-btn').addEventListener('click', () => {
            this.addDefaultParameter();
        });

        // 添加默认请求头按钮
        document.getElementById('add-default-header-btn').addEventListener('click', () => {
            this.addDefaultHeader();
        });

        // 创建应用按钮
        document.getElementById('create-app-btn').addEventListener('click', () => {
            this.showCreateAppModal();
        });

        // 应用搜索
        document.getElementById('app-search-input').addEventListener('input', (e) => {
            this.filterApplications(e.target.value);
        });

        // 当前应用选择
        document.getElementById('current-app-select').addEventListener('change', (e) => {
            this.currentAppId = e.target.value ? parseInt(e.target.value) : null;
            this.renderInterfaces();
        });

        // 创建应用模态框事件
        document.getElementById('create-app-close-btn').addEventListener('click', () => {
            this.closeCreateAppModal();
        });

        document.getElementById('create-app-cancel-btn').addEventListener('click', () => {
            this.closeCreateAppModal();
        });

        document.getElementById('create-app-confirm-btn').addEventListener('click', () => {
            this.createApplication();
        });

        // 测试接口按钮
        document.getElementById('test-interface-btn').addEventListener('click', () => {
            this.testInterface();
        });





        // 模态框关闭
        document.getElementById('modal-close-btn').addEventListener('click', () => {
            this.closeModal();
        });

        document.getElementById('modal-cancel-btn').addEventListener('click', () => {
            this.closeModal();
        });

        // 协议类型变化时显示/隐藏HTTP方法并更新参数位置
        document.getElementById('protocol-type').addEventListener('change', (e) => {
            const httpMethodGroup = document.getElementById('http-method').closest('.form-group');
            const paramLocationConfig = document.getElementById('parameter-location-config');
            
            if (e.target.value === 'http') {
                httpMethodGroup.style.display = 'block';
                paramLocationConfig.style.display = 'block';
            } else {
                httpMethodGroup.style.display = 'none';
                paramLocationConfig.style.display = 'none';
            }
            
            // 更新参数信息显示
            this.updateParameterInfo(e.target.value);
        });

        // HTTP参数位置变化时更新参数信息
        document.getElementById('http-param-location').addEventListener('change', (e) => {
            const protocolType = document.getElementById('protocol-type').value;
            this.updateParameterInfo(protocolType);
        });
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
            'applications': { title: '应用管理', subtitle: '创建和管理API应用' },
            'interfaces': { title: '接口列表', subtitle: '管理HTTP API接口' },
            'add-interface': { 
                title: this.editingInterfaceId ? '编辑接口' : '添加接口', 
                subtitle: this.editingInterfaceId ? '修改现有HTTP API接口' : '注册新的HTTP API接口' 
            }
        };

        const pageInfo = titles[tabName];
        document.getElementById('page-title').textContent = pageInfo.title;
        document.getElementById('page-subtitle').textContent = pageInfo.subtitle;

        // 更新保存按钮文本
        if (tabName === 'add-interface') {
            this.updateSaveButtonText();
        }

        this.currentTab = tabName;
    }

    // 渲染应用列表
    renderApplications(filteredApps = null) {
        const applications = filteredApps || this.applications;
        const grid = document.getElementById('applications-grid');
        
        if (applications.length === 0) {
            grid.innerHTML = `
                <div class="empty-applications">
                    <i class="fas fa-folder-open"></i>
                    <h3>暂无应用</h3>
                    <p>创建您的第一个应用来开始管理API接口</p>
                    <button class="btn btn-primary" onclick="app.showCreateAppModal()">
                        <i class="fas fa-plus"></i> 创建应用
                    </button>
                </div>
            `;
            return;
        }

        grid.innerHTML = applications.map(app => {
            const interfaceCount = this.interfaces.filter(i => i.appId === app.id).length;
            return `
                <div class="application-card" data-id="${app.id}" onclick="app.selectApplication(${app.id})">
                    <div class="application-header">
                        <div>
                            <div class="application-title">${app.name}</div>
                            <div class="application-version">${app.version}</div>
                        </div>
                        <div class="application-actions" onclick="event.stopPropagation()">
                            <button class="action-btn" onclick="app.editApplication(${app.id})" title="编辑">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button class="action-btn" onclick="app.deleteApplication(${app.id})" title="删除">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    <div class="application-description">${app.description}</div>
                    <div class="application-stats">
                        <div class="stat-item">
                            <i class="fas fa-plug"></i>
                            <span>${interfaceCount} 个接口</span>
                        </div>
                        <div class="stat-item">
                            <i class="fas fa-calendar"></i>
                            <span>${app.createdAt.toLocaleDateString()}</span>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    // 选择应用
    selectApplication(appId) {
        this.currentAppId = appId;
        this.switchTab('interfaces');
        this.updateAppSelectors();
        this.renderInterfaces();
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
                        `<option value="${app.id}" ${app.id == this.currentAppId ? 'selected' : ''}>${app.name}</option>`
                    ).join('');
                
                // 如果之前有选中值，尝试恢复
                if (currentValue && !this.currentAppId) {
                    selector.value = currentValue;
                }
            }
        });
    }

    // 过滤应用
    filterApplications(searchTerm) {
        const filtered = this.applications.filter(app => 
            app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            app.description.toLowerCase().includes(searchTerm.toLowerCase())
        );
        this.renderApplications(filtered);
    }

    // 显示创建应用模态框
    showCreateAppModal() {
        document.getElementById('create-app-modal').classList.add('active');
        document.getElementById('app-name').focus();
    }

    // 关闭创建应用模态框
    closeCreateAppModal() {
        document.getElementById('create-app-modal').classList.remove('active');
        document.getElementById('create-app-form').reset();
    }

    // 创建应用
    createApplication() {
        const name = document.getElementById('app-name').value.trim();
        const description = document.getElementById('app-description').value.trim();
        const version = document.getElementById('app-version').value.trim() || 'v1.0.0';
        const baseUrl = document.getElementById('app-base-url').value.trim();

        if (!name) {
            this.showNotification('请输入应用名称', 'error');
            return;
        }

        const newApp = {
            id: Date.now(),
            name,
            description,
            version,
            baseUrl,
            createdAt: new Date(),
            interfaceCount: 0
        };

        this.applications.push(newApp);
        this.renderApplications();
        this.updateAppSelectors();
        this.closeCreateAppModal();
        this.showNotification('应用创建成功！', 'success');
    }

    // 编辑应用
    editApplication(appId) {
        const app = this.applications.find(a => a.id === appId);
        if (!app) return;

        // 填充表单数据
        document.getElementById('app-name').value = app.name;
        document.getElementById('app-description').value = app.description;
        document.getElementById('app-version').value = app.version;
        document.getElementById('app-base-url').value = app.baseUrl;

        // 更新模态框标题
        document.getElementById('create-app-modal-title').textContent = '编辑应用';
        document.getElementById('create-app-confirm-btn').textContent = '更新应用';

        // 显示模态框
        this.showCreateAppModal();

        // 临时存储编辑的应用ID
        this.editingAppId = appId;
    }

    // 删除应用
    deleteApplication(appId) {
        const app = this.applications.find(a => a.id === appId);
        if (!app) return;

        const interfaceCount = this.interfaces.filter(i => i.appId === appId).length;
        const message = interfaceCount > 0 
            ? `确定要删除应用"${app.name}"吗？这将同时删除该应用下的 ${interfaceCount} 个接口。`
            : `确定要删除应用"${app.name}"吗？`;

        if (confirm(message)) {
            // 删除应用
            this.applications = this.applications.filter(a => a.id !== appId);
            // 删除应用下的所有接口
            this.interfaces = this.interfaces.filter(i => i.appId !== appId);
            
            // 如果删除的是当前选中的应用，清除选中状态
            if (this.currentAppId === appId) {
                this.currentAppId = null;
            }

            this.renderApplications();
            this.updateAppSelectors();
            this.renderInterfaces();
            this.showNotification('应用已删除', 'success');
        }
    }

    // 渲染接口列表
    renderInterfaces(filteredInterfaces = null) {
        // 如果没有选择应用，显示提示
        if (!this.currentAppId) {
            const grid = document.getElementById('interfaces-grid');
            grid.innerHTML = `
                <div class="empty-state" style="grid-column: 1 / -1; text-align: center; padding: 3rem;">
                    <i class="fas fa-folder" style="font-size: 3rem; color: #cbd5e1; margin-bottom: 1rem;"></i>
                    <h3 style="color: #64748b; margin-bottom: 0.5rem;">请先选择应用</h3>
                    <p style="color: #94a3b8;">选择一个应用来查看和管理其接口</p>
                </div>
            `;
            return;
        }

        // 过滤当前应用的接口
        const appInterfaces = this.interfaces.filter(i => i.appId === this.currentAppId);
        const interfaces = filteredInterfaces || appInterfaces;
        const grid = document.getElementById('interfaces-grid');
        
        if (interfaces.length === 0) {
            grid.innerHTML = `
                <div class="empty-state" style="grid-column: 1 / -1; text-align: center; padding: 3rem;">
                    <i class="fas fa-inbox" style="font-size: 3rem; color: #cbd5e1; margin-bottom: 1rem;"></i>
                    <h3 style="color: #64748b; margin-bottom: 0.5rem;">暂无接口</h3>
                    <p style="color: #94a3b8;">点击"添加接口"开始创建您的第一个接口</p>
                </div>
            `;
            return;
        }

        grid.innerHTML = interfaces.map(apiInterface => `
            <div class="interface-card ${apiInterface.enabled === false ? 'disabled' : ''}" data-id="${apiInterface.id}">
                <div class="interface-header">
                    <div>
                        <div class="interface-title">${apiInterface.name}</div>
                        <div class="interface-description">${apiInterface.description}</div>
                    </div>
                    <div class="interface-header-right">
                        <span class="protocol-badge protocol-${apiInterface.protocol}">${apiInterface.protocol.toUpperCase()}</span>
                        <div class="interface-switch">
                            <label class="switch">
                                <input type="checkbox" ${apiInterface.enabled !== false ? 'checked' : ''} 
                                       onchange="app.toggleInterface(${apiInterface.id}, this.checked)"
                                       onclick="event.stopPropagation()">
                                <span class="slider"></span>
                            </label>
                        </div>
                    </div>
                </div>
                <div class="interface-details">
                    <div style="font-size: 0.875rem; color: #64748b; margin-bottom: 0.5rem;">
                        <strong>端点:</strong> ${apiInterface.url}
                    </div>
                    <div style="font-size: 0.875rem; color: #64748b;">
                        <strong>方法:</strong> ${apiInterface.method}
                    </div>
                </div>
                <div class="interface-meta">
                    <div class="status-indicator status-${apiInterface.status}">
                        <span class="status-dot"></span>
                        <span>${this.getStatusText(apiInterface.status)}</span>
                    </div>
                    <div class="interface-actions">
                        <button class="action-btn" onclick="app.editInterface(${apiInterface.id})" title="编辑">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="action-btn" onclick="app.testInterface(${apiInterface.id})" title="测试">
                            <i class="fas fa-play"></i>
                        </button>
                        <button class="action-btn" onclick="app.deleteInterface(${apiInterface.id})" title="删除">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        `).join('');

        // 绑定卡片点击事件
        grid.querySelectorAll('.interface-card').forEach(card => {
            card.addEventListener('click', (e) => {
                if (!e.target.closest('.interface-actions') && !e.target.closest('.interface-switch')) {
                    const id = parseInt(card.dataset.id);
                    this.showInterfaceDetails(id);
                }
            });
        });
    }



    // 过滤接口
    filterInterfaces(searchTerm) {
        if (!this.currentAppId) {
            this.renderInterfaces();
            return;
        }

        const appInterfaces = this.interfaces.filter(i => i.appId === this.currentAppId);
        const filtered = appInterfaces.filter(apiInterface => 
            apiInterface.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            apiInterface.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
            apiInterface.protocol.toLowerCase().includes(searchTerm.toLowerCase())
        );
        this.renderInterfaces(filtered);
    }

    // 按协议过滤
    filterByProtocol(protocol) {
        if (!this.currentAppId) {
            this.renderInterfaces();
            return;
        }

        const appInterfaces = this.interfaces.filter(i => i.appId === this.currentAppId);
        if (!protocol) {
            this.renderInterfaces(appInterfaces);
            return;
        }
        const filtered = appInterfaces.filter(apiInterface => apiInterface.protocol === protocol);
        this.renderInterfaces(filtered);
    }

    // 按状态过滤
    filterByStatus(status) {
        if (!this.currentAppId) {
            this.renderInterfaces();
            return;
        }

        const appInterfaces = this.interfaces.filter(i => i.appId === this.currentAppId);
        if (!status) {
            this.renderInterfaces(appInterfaces);
            return;
        }
        
        let filtered;
        if (status === 'enabled') {
            filtered = appInterfaces.filter(apiInterface => apiInterface.enabled !== false);
        } else if (status === 'disabled') {
            filtered = appInterfaces.filter(apiInterface => apiInterface.enabled === false);
        } else {
            filtered = appInterfaces.filter(apiInterface => apiInterface.status === status);
        }
        
        this.renderInterfaces(filtered);
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

    // 根据协议类型更新参数位置


    // 根据协议类型获取入参位置
    getInputLocationByProtocol(protocol) {
        if (protocol === 'http') {
            // HTTP协议根据用户选择确定参数位置
            const httpParamLocation = document.getElementById('http-param-location');
            return httpParamLocation ? httpParamLocation.value : 'query';
        }
        
        return 'query';  // 默认使用query
    }



    // 更新参数信息显示
    updateParameterInfo(protocol) {
        const inputInfo = document.getElementById('input-param-info');
        
        if (!inputInfo) return;
        
        if (protocol === 'http') {
            const inputLocation = this.getInputLocationByProtocol(protocol);
            const locationNames = {
                'query': 'URL查询参数 (?param=value)',
                'body': '请求体 (Request Body)'
            };
            inputInfo.innerHTML = `<small>参数位置：<strong>${locationNames[inputLocation]}</strong> - 所有参数将统一放在此位置</small>`;
        } else {
            inputInfo.innerHTML = `<small>请选择HTTP协议</small>`;
        }
    }

    // 添加参数
    addParameter() {
        const parametersList = document.getElementById('parameters-list');
        const parameterItem = document.createElement('div');
        parameterItem.className = 'parameter-item';

        parameterItem.innerHTML = `
            <div class="form-group">
                <label>参数名</label>
                <input type="text" name="param-name" placeholder="参数名称">
            </div>
            <div class="form-group">
                <label>类型</label>
                <select name="param-type">
                    <option value="string">String</option>
                    <option value="integer">Integer</option>
                    <option value="number">Number</option>
                    <option value="boolean">Boolean</option>
                    <option value="object">Object</option>
                    <option value="array">Array</option>
                </select>
            </div>
            <div class="form-group">
                <label>必填</label>
                <select name="param-required">
                    <option value="false">否</option>
                    <option value="true">是</option>
                </select>
            </div>
            <div class="form-group">
                <label>描述</label>
                <input type="text" name="param-description" placeholder="参数描述">
            </div>
            <button type="button" class="remove-parameter-btn" onclick="this.parentElement.remove()">
                <i class="fas fa-trash"></i>
            </button>
        `;
        parametersList.appendChild(parameterItem);
    }

    // 添加默认参数
    addDefaultParameter() {
        const defaultParamsList = document.getElementById('default-params-list');
        const parameterItem = document.createElement('div');
        parameterItem.className = 'default-param-item';

        parameterItem.innerHTML = `
            <div class="form-group">
                <label>参数名</label>
                <input type="text" name="default-param-name" placeholder="例如: api_key, version">
            </div>
            <div class="form-group">
                <label>默认值</label>
                <input type="text" name="default-param-value" placeholder="参数的默认值">
            </div>
            <div class="form-group">
                <label>描述</label>
                <input type="text" name="default-param-description" placeholder="参数说明">
            </div>
            <button type="button" class="remove-default-btn" onclick="this.parentElement.remove()">
                <i class="fas fa-trash"></i>
            </button>
        `;
        defaultParamsList.appendChild(parameterItem);
    }

    // 添加默认请求头
    addDefaultHeader() {
        const defaultHeadersList = document.getElementById('default-headers-list');
        const headerItem = document.createElement('div');
        headerItem.className = 'default-header-item';

        headerItem.innerHTML = `
            <div class="form-group">
                <label>请求头名称</label>
                <input type="text" name="default-header-name" placeholder="例如: Content-Type, User-Agent">
            </div>
            <div class="form-group">
                <label>请求头值</label>
                <input type="text" name="default-header-value" placeholder="请求头的值">
            </div>
            <div class="form-group">
                <label>描述</label>
                <input type="text" name="default-header-description" placeholder="请求头说明">
            </div>
            <button type="button" class="remove-default-btn" onclick="this.parentElement.remove()">
                <i class="fas fa-trash"></i>
            </button>
        `;
        defaultHeadersList.appendChild(headerItem);
    }

    // 添加或更新接口
    addInterface() {
        const form = document.getElementById('add-interface-form');
        const formData = new FormData(form);
        
        const protocolType = document.getElementById('protocol-type').value;
        const inputLocation = this.getInputLocationByProtocol(protocolType);

        // 收集参数
        const parameters = [];
        const parameterItems = document.querySelectorAll('#parameters-list .parameter-item');
        parameterItems.forEach(item => {
            const name = item.querySelector('[name="param-name"]').value;
            const type = item.querySelector('[name="param-type"]').value;
            const requiredSelect = item.querySelector('[name="param-required"]');
            const required = requiredSelect ? requiredSelect.value === 'true' : false;
            const description = item.querySelector('[name="param-description"]').value;
            
            if (name) {
                parameters.push({ 
                    name, 
                    type, 
                    location: inputLocation, // 根据协议类型自动设置
                    required, 
                    description 
                });
            }
        });

        // 收集默认参数
        const defaultParams = [];
        const defaultParamItems = document.querySelectorAll('#default-params-list .default-param-item');
        defaultParamItems.forEach(item => {
            const name = item.querySelector('[name="default-param-name"]').value;
            const value = item.querySelector('[name="default-param-value"]').value;
            const description = item.querySelector('[name="default-param-description"]').value;
            
            if (name && value) {
                defaultParams.push({ name, value, description });
            }
        });

        // 收集默认请求头
        const defaultHeaders = [];
        const defaultHeaderItems = document.querySelectorAll('#default-headers-list .default-header-item');
        defaultHeaderItems.forEach(item => {
            const name = item.querySelector('[name="default-header-name"]').value;
            const value = item.querySelector('[name="default-header-value"]').value;
            const description = item.querySelector('[name="default-header-description"]').value;
            
            if (name && value) {
                defaultHeaders.push({ name, value, description });
            }
        });

        if (this.editingInterfaceId) {
            // 编辑模式：更新现有接口
            const interfaceIndex = this.interfaces.findIndex(i => i.id === this.editingInterfaceId);
            if (interfaceIndex !== -1) {
                const updatedInterface = {
                    ...this.interfaces[interfaceIndex], // 保留原有的id和createdAt
                    name: document.getElementById('interface-name').value,
                    description: document.getElementById('interface-description').value,
                    protocol: document.getElementById('protocol-type').value,
                    method: document.getElementById('http-method').value,
                    url: document.getElementById('endpoint-url').value,
                    authType: document.getElementById('auth-type').value,
                    authValue: document.getElementById('auth-value').value,
                    parameters: parameters,
                    defaultParams: defaultParams,
                    defaultHeaders: defaultHeaders,
                    httpParamLocation: protocolType === 'http' ? document.getElementById('http-param-location').value : null,
                    updatedAt: new Date()
                };
                
                this.interfaces[interfaceIndex] = updatedInterface;
                this.showNotification('接口更新成功！', 'success');
            }
        } else {
            // 添加模式：创建新接口
            const appId = parseInt(document.getElementById('interface-app-id').value);
            if (!appId) {
                this.showNotification('请选择所属应用', 'error');
                return;
            }

            const newInterface = {
                id: Date.now(),
                appId: appId,
                name: document.getElementById('interface-name').value,
                description: document.getElementById('interface-description').value,
                protocol: document.getElementById('protocol-type').value,
                method: document.getElementById('http-method').value,
                url: document.getElementById('endpoint-url').value,
                authType: document.getElementById('auth-type').value,
                authValue: document.getElementById('auth-value').value,
                status: 'active',
                enabled: true,
                parameters: parameters,
                defaultParams: defaultParams,
                defaultHeaders: defaultHeaders,
                httpParamLocation: protocolType === 'http' ? document.getElementById('http-param-location').value : null,
                createdAt: new Date()
            };

            this.interfaces.push(newInterface);
            this.showNotification('接口添加成功！', 'success');
        }

        this.renderInterfaces();
        
        // 清空表单和编辑状态
        this.resetForm();
        
        // 切换到接口列表
        this.switchTab('interfaces');
    }

    // 重置表单
    resetForm() {
        const form = document.getElementById('add-interface-form');
        form.reset();
        document.getElementById('parameters-list').innerHTML = '';
        document.getElementById('default-params-list').innerHTML = '';
        document.getElementById('default-headers-list').innerHTML = '';
        this.editingInterfaceId = null;
        this.updateSaveButtonText();
    }

    // 更新保存按钮文本
    updateSaveButtonText() {
        const saveBtn = document.getElementById('save-interface-btn');
        if (saveBtn) {
            if (this.editingInterfaceId) {
                saveBtn.innerHTML = '<i class="fas fa-save"></i> 更新接口';
            } else {
                saveBtn.innerHTML = '<i class="fas fa-save"></i> 保存接口';
            }
        }
    }

    // 测试接口
    testInterface(id = null) {
        if (id) {
            const apiInterface = this.interfaces.find(i => i.id === id);
            this.showNotification(`正在测试接口: ${apiInterface.name}`, 'info');
        } else {
            this.showNotification('正在测试当前配置的接口...', 'info');
        }
        
        // 模拟测试过程
        setTimeout(() => {
            this.showNotification('接口测试成功！', 'success');
        }, 2000);
    }

    // 编辑接口
    editInterface(id) {
        const apiInterface = this.interfaces.find(i => i.id === id);
        if (!apiInterface) return;

        // 设置编辑状态
        this.editingInterfaceId = id;

        // 切换到添加接口页面并填充数据
        this.switchTab('add-interface');
        
        document.getElementById('interface-app-id').value = apiInterface.appId;
        document.getElementById('interface-name').value = apiInterface.name;
        document.getElementById('interface-description').value = apiInterface.description;
        document.getElementById('protocol-type').value = apiInterface.protocol;
        document.getElementById('http-method').value = apiInterface.method;
        document.getElementById('endpoint-url').value = apiInterface.url;
        document.getElementById('auth-type').value = apiInterface.authType;
        document.getElementById('auth-value').value = apiInterface.authValue;
        
        // 处理HTTP参数位置
        if (apiInterface.protocol === 'http' && apiInterface.httpParamLocation) {
            document.getElementById('http-param-location').value = apiInterface.httpParamLocation;
        }
        
        // 触发协议类型变化事件来更新UI显示
        document.getElementById('protocol-type').dispatchEvent(new Event('change'));

        // 更新保存按钮文本
        this.updateSaveButtonText();

        // 填充参数
        const parametersList = document.getElementById('parameters-list');
        parametersList.innerHTML = '';
        const params = apiInterface.parameters || apiInterface.inputParameters || [];
        params.forEach(param => {
            this.addParameter();
            const lastItem = parametersList.lastElementChild;
            lastItem.querySelector('[name="param-name"]').value = param.name;
            lastItem.querySelector('[name="param-type"]').value = param.type;
            if (lastItem.querySelector('[name="param-required"]')) {
                lastItem.querySelector('[name="param-required"]').value = (param.required || false).toString();
            }
            lastItem.querySelector('[name="param-description"]').value = param.description;
        });

        // 填充默认参数
        const defaultParamsList = document.getElementById('default-params-list');
        defaultParamsList.innerHTML = '';
        const defaultParams = apiInterface.defaultParams || [];
        defaultParams.forEach(param => {
            this.addDefaultParameter();
            const lastItem = defaultParamsList.lastElementChild;
            lastItem.querySelector('[name="default-param-name"]').value = param.name;
            lastItem.querySelector('[name="default-param-value"]').value = param.value;
            lastItem.querySelector('[name="default-param-description"]').value = param.description || '';
        });

        // 填充默认请求头
        const defaultHeadersList = document.getElementById('default-headers-list');
        defaultHeadersList.innerHTML = '';
        const defaultHeaders = apiInterface.defaultHeaders || [];
        defaultHeaders.forEach(header => {
            this.addDefaultHeader();
            const lastItem = defaultHeadersList.lastElementChild;
            lastItem.querySelector('[name="default-header-name"]').value = header.name;
            lastItem.querySelector('[name="default-header-value"]').value = header.value;
            lastItem.querySelector('[name="default-header-description"]').value = header.description || '';
        });
    }

    // 删除接口
    deleteInterface(id) {
        if (confirm('确定要删除这个接口吗？')) {
            this.interfaces = this.interfaces.filter(i => i.id !== id);
            this.renderInterfaces();
            this.showNotification('接口已删除', 'success');
        }
    }

    // 切换接口启用/禁用状态
    toggleInterface(id, enabled) {
        const apiInterface = this.interfaces.find(i => i.id === id);
        if (apiInterface) {
            apiInterface.enabled = enabled;
            apiInterface.status = enabled ? 'active' : 'inactive';
            this.renderInterfaces();
            this.showNotification(
                `接口已${enabled ? '启用' : '禁用'}`, 
                enabled ? 'success' : 'info'
            );
        }
    }

    // 显示接口详情
    showInterfaceDetails(id) {
        const apiInterface = this.interfaces.find(i => i.id === id);
        if (!apiInterface) return;

        const modal = document.getElementById('interface-modal');
        const modalTitle = document.getElementById('modal-title');
        const modalBody = document.getElementById('modal-body');

        modalTitle.textContent = `接口详情 - ${apiInterface.name}`;
        modalBody.innerHTML = `
            <div class="interface-details-content">
                <div class="detail-section">
                    <h4>基本信息</h4>
                    <div class="detail-grid">
                        <div><strong>名称:</strong> ${apiInterface.name}</div>
                        <div><strong>协议:</strong> ${apiInterface.protocol.toUpperCase()}</div>
                        <div><strong>方法:</strong> ${apiInterface.method}</div>
                        <div><strong>状态:</strong> ${this.getStatusText(apiInterface.status)}</div>
                    </div>
                </div>
                <div class="detail-section">
                    <h4>连接信息</h4>
                    <div><strong>端点:</strong> ${apiInterface.url}</div>
                    <div><strong>认证:</strong> ${apiInterface.authType}</div>
                    <div><strong>响应格式:</strong> ${apiInterface.responseFormat || 'json'}</div>
                    ${apiInterface.contentType ? `<div><strong>Content-Type:</strong> ${apiInterface.contentType}</div>` : ''}
                </div>
                <div class="detail-section">
                    <h4>请求参数</h4>
                    ${(apiInterface.parameters || apiInterface.inputParameters) && (apiInterface.parameters || apiInterface.inputParameters).length > 0 ? `
                        <table style="width: 100%; border-collapse: collapse;">
                            <thead>
                                <tr style="background: #f8fafc;">
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">参数名</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">类型</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">位置</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">必填</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${(apiInterface.parameters || apiInterface.inputParameters).map(param => `
                                    <tr>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.name}</td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.type}</td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.location || 'query'}</td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.required ? '是' : '否'}</td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.description}</td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    ` : '<p style="color: #64748b;">暂无参数</p>'}
                </div>
                <div class="detail-section">
                    <h4>默认参数</h4>
                    ${apiInterface.defaultParams && apiInterface.defaultParams.length > 0 ? `
                        <table style="width: 100%; border-collapse: collapse;">
                            <thead>
                                <tr style="background: #f0f9ff;">
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">参数名</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">默认值</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${apiInterface.defaultParams.map(param => `
                                    <tr>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;"><code>${param.name}</code></td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;"><code>${param.value}</code></td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${param.description || '-'}</td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    ` : '<p style="color: #64748b;">暂无默认参数</p>'}
                </div>
                <div class="detail-section">
                    <h4>默认请求头</h4>
                    ${apiInterface.defaultHeaders && apiInterface.defaultHeaders.length > 0 ? `
                        <table style="width: 100%; border-collapse: collapse;">
                            <thead>
                                <tr style="background: #f0fdf4;">
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">请求头名称</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">值</th>
                                    <th style="padding: 0.5rem; text-align: left; border: 1px solid #e2e8f0;">描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${apiInterface.defaultHeaders.map(header => `
                                    <tr>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;"><code>${header.name}</code></td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;"><code>${header.value}</code></td>
                                        <td style="padding: 0.5rem; border: 1px solid #e2e8f0;">${header.description || '-'}</td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    ` : '<p style="color: #64748b;">暂无默认请求头</p>'}
                </div>
            </div>
        `;

        modal.classList.add('active');
    }





    // 关闭模态框
    closeModal() {
        document.getElementById('interface-modal').classList.remove('active');
    }

    // 获取状态文本
    getStatusText(status) {
        const statusMap = {
            'active': '已激活',
            'inactive': '未激活',
            'error': '错误'
        };
        return statusMap[status] || status;
    }

    // 显示通知
    showNotification(message, type = 'info') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem 1.5rem;
            border-radius: 0.5rem;
            color: white;
            font-weight: 500;
            z-index: 3000;
            animation: slideIn 0.3s ease;
        `;

        // 设置背景色
        const colors = {
            'success': '#10b981',
            'error': '#ef4444',
            'warning': '#f59e0b',
            'info': '#3b82f6'
        };
        notification.style.backgroundColor = colors[type] || colors.info;

        notification.textContent = message;
        document.body.appendChild(notification);

        // 3秒后自动移除
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }
}

// 添加动画样式
const style = document.createElement('style');
style.textContent = `
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
    
    .detail-section {
        margin-bottom: 1.5rem;
    }
    
    .detail-section h4 {
        font-size: 1rem;
        font-weight: 600;
        color: #1e293b;
        margin-bottom: 0.75rem;
        padding-bottom: 0.25rem;
        border-bottom: 1px solid #e2e8f0;
    }
    
    .detail-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 0.5rem;
        font-size: 0.875rem;
    }
    
    .detail-grid > div {
        padding: 0.25rem 0;
    }
`;
document.head.appendChild(style);

// 初始化应用
const app = new MCPAdapterApp();