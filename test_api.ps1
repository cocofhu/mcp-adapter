# PowerShell API 测试脚本
# 用于验证重构后的 API 功能

$BaseUrl = "http://localhost:8080/api"
$Total = 0
$Passed = 0
$Failed = 0

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "MCP Adapter API 测试脚本" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 测试函数
function Test-Api {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Endpoint,
        [string]$Data = $null
    )
    
    $script:Total++
    Write-Host "测试 $script:Total : $Name" -ForegroundColor Yellow
    
    try {
        $uri = "$BaseUrl$Endpoint"
        
        if ($Data) {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Body $Data -ContentType "application/json" -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -ErrorAction Stop
        }
        
        Write-Host "✓ 通过" -ForegroundColor Green
        $response | ConvertTo-Json -Depth 10
        $script:Passed++
    }
    catch {
        Write-Host "✗ 失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host $_.ErrorDetails.Message -ForegroundColor Red
        }
        $script:Failed++
    }
    Write-Host ""
}

# 检查服务是否运行
Write-Host "检查服务是否运行..." -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "$BaseUrl/applications" -Method GET -ErrorAction Stop | Out-Null
    Write-Host "✓ 服务正在运行" -ForegroundColor Green
}
catch {
    Write-Host "错误: 服务未运行，请先启动服务: go run main.go" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 1. 创建应用
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "1. 应用管理测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Test-Api -Name "创建应用" -Method "POST" -Endpoint "/applications" -Data @'
{
  "name": "Test Application",
  "description": "测试应用",
  "path": "test-app",
  "protocol": "sse",
  "enabled": true
}
'@

Test-Api -Name "获取应用列表" -Method "GET" -Endpoint "/applications"

Test-Api -Name "获取单个应用" -Method "GET" -Endpoint "/applications/1"

# 2. 自定义类型测试
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "2. 自定义类型测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Test-Api -Name "创建 User 类型" -Method "POST" -Endpoint "/custom-types" -Data @'
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
      "name": "username",
      "type": "string",
      "required": true,
      "description": "用户名"
    },
    {
      "name": "email",
      "type": "string",
      "required": false,
      "description": "邮箱"
    }
  ]
}
'@

Test-Api -Name "创建 Address 类型" -Method "POST" -Endpoint "/custom-types" -Data @'
{
  "app_id": 1,
  "name": "Address",
  "description": "地址信息",
  "fields": [
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
    }
  ]
}
'@

Test-Api -Name "创建嵌套类型 UserProfile" -Method "POST" -Endpoint "/custom-types" -Data @'
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
      "name": "address",
      "type": "custom",
      "ref": 2,
      "required": false,
      "description": "地址"
    },
    {
      "name": "tags",
      "type": "string",
      "is_array": true,
      "required": false,
      "description": "标签"
    }
  ]
}
'@

Test-Api -Name "获取自定义类型列表" -Method "GET" -Endpoint "/custom-types?app_id=1"

Test-Api -Name "获取单个自定义类型" -Method "GET" -Endpoint "/custom-types/1"

# 3. 接口测试
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "3. 接口管理测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Test-Api -Name "创建 GET 接口" -Method "POST" -Endpoint "/interfaces" -Data @'
{
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
}
'@

Test-Api -Name "创建 POST 接口（使用自定义类型）" -Method "POST" -Endpoint "/interfaces" -Data @'
{
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
}
'@

Test-Api -Name "获取接口列表" -Method "GET" -Endpoint "/interfaces?app_id=1"

Test-Api -Name "获取单个接口" -Method "GET" -Endpoint "/interfaces/1"

Test-Api -Name "更新接口" -Method "PUT" -Endpoint "/interfaces/1" -Data @'
{
  "name": "GetUserById",
  "description": "根据ID获取用户（已更新）"
}
'@

# 4. 更新和删除测试
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "4. 更新和删除测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Test-Api -Name "更新自定义类型" -Method "PUT" -Endpoint "/custom-types/1" -Data @'
{
  "name": "UserInfo",
  "description": "用户基本信息（已更新）",
  "fields": [
    {
      "name": "id",
      "type": "number",
      "required": true
    },
    {
      "name": "username",
      "type": "string",
      "required": true
    },
    {
      "name": "nickname",
      "type": "string",
      "required": false,
      "description": "昵称（新增字段）"
    }
  ]
}
'@

Test-Api -Name "更新应用" -Method "PUT" -Endpoint "/applications/1" -Data @'
{
  "description": "测试应用（已更新）"
}
'@

# 5. 错误处理测试
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "5. 错误处理测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Test-Api -Name "创建重复名称的类型（应失败）" -Method "POST" -Endpoint "/custom-types" -Data @'
{
  "app_id": 1,
  "name": "UserInfo",
  "fields": []
}
'@

Test-Api -Name "引用不存在的类型（应失败）" -Method "POST" -Endpoint "/custom-types" -Data @'
{
  "app_id": 1,
  "name": "InvalidType",
  "fields": [
    {
      "name": "invalid",
      "type": "custom",
      "ref": 9999,
      "required": true
    }
  ]
}
'@

Test-Api -Name "删除被引用的类型（应失败）" -Method "DELETE" -Endpoint "/custom-types/1"

# 6. 清理测试（可选）
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "6. 清理测试数据（可选）" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

$cleanup = Read-Host "是否删除测试数据？(y/N)"
if ($cleanup -eq 'y' -or $cleanup -eq 'Y') {
    Test-Api -Name "删除接口 1" -Method "DELETE" -Endpoint "/interfaces/1"
    Test-Api -Name "删除接口 2" -Method "DELETE" -Endpoint "/interfaces/2"
    Test-Api -Name "删除类型 3" -Method "DELETE" -Endpoint "/custom-types/3"
    Test-Api -Name "删除类型 2" -Method "DELETE" -Endpoint "/custom-types/2"
    Test-Api -Name "删除类型 1" -Method "DELETE" -Endpoint "/custom-types/1"
    Test-Api -Name "删除应用" -Method "DELETE" -Endpoint "/applications/1"
}

# 测试总结
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "测试总结" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "总计: $Total"
Write-Host "通过: $Passed" -ForegroundColor Green
Write-Host "失败: $Failed" -ForegroundColor Red
Write-Host ""

if ($Failed -eq 0) {
    Write-Host "✓ 所有测试通过！" -ForegroundColor Green
    exit 0
} else {
    Write-Host "✗ 部分测试失败" -ForegroundColor Red
    exit 1
}
