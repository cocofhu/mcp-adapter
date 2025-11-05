# 自定义类型功能测试脚本（PowerShell 版本）
# 测试类型引用和数组功能

$BaseUrl = "http://localhost:8080/api"

Write-Host "========== 自定义类型功能测试 ==========" -ForegroundColor Cyan
Write-Host ""

# 1. 创建应用
Write-Host "1. 创建测试应用..." -ForegroundColor Yellow
$appBody = @{
    name = "测试应用"
    description = "用于测试自定义类型功能"
    path = "test-app"
    protocol = "http"
    enabled = $true
} | ConvertTo-Json

$appResponse = Invoke-RestMethod -Uri "$BaseUrl/applications" -Method Post -Body $appBody -ContentType "application/json"
$appId = $appResponse.id
Write-Host "✓ 应用创建成功，ID: $appId" -ForegroundColor Green
Write-Host ""

# 2. 创建基础类型 - Address
Write-Host "2. 创建 Address 类型..." -ForegroundColor Yellow
$addressBody = @{
    app_id = $appId
    name = "Address"
    description = "地址信息"
    fields = @(
        @{
            name = "street"
            type = "string"
            required = $true
            description = "街道地址"
        },
        @{
            name = "city"
            type = "string"
            required = $true
            description = "城市"
        },
        @{
            name = "zipCode"
            type = "string"
            required = $false
            description = "邮政编码"
        }
    )
} | ConvertTo-Json -Depth 10

$addressResponse = Invoke-RestMethod -Uri "$BaseUrl/custom-types" -Method Post -Body $addressBody -ContentType "application/json"
$addressId = $addressResponse.id
Write-Host "✓ Address 类型创建成功，ID: $addressId" -ForegroundColor Green
Write-Host ""

# 3. 创建基础类型 - User（包含数组字段）
Write-Host "3. 创建 User 类型（包含字符串数组）..." -ForegroundColor Yellow
$userBody = @{
    app_id = $appId
    name = "User"
    description = "用户信息"
    fields = @(
        @{
            name = "id"
            type = "number"
            required = $true
            description = "用户ID"
        },
        @{
            name = "username"
            type = "string"
            required = $true
            description = "用户名"
        },
        @{
            name = "email"
            type = "string"
            required = $false
            description = "邮箱地址"
        },
        @{
            name = "tags"
            type = "string"
            is_array = $true
            required = $false
            description = "标签列表（字符串数组）"
        },
        @{
            name = "active"
            type = "boolean"
            required = $true
            description = "是否激活"
        }
    )
} | ConvertTo-Json -Depth 10

$userResponse = Invoke-RestMethod -Uri "$BaseUrl/custom-types" -Method Post -Body $userBody -ContentType "application/json"
$userId = $userResponse.id
Write-Host "✓ User 类型创建成功，ID: $userId" -ForegroundColor Green
Write-Host ""

# 4. 创建复杂类型 - UserProfile（引用其他类型）
Write-Host "4. 创建 UserProfile 类型（引用 User 和 Address）..." -ForegroundColor Yellow
$profileBody = @{
    app_id = $appId
    name = "UserProfile"
    description = "用户详细信息"
    fields = @(
        @{
            name = "user"
            type = "custom"
            ref = $userId
            required = $true
            description = "用户基本信息"
        },
        @{
            name = "address"
            type = "custom"
            ref = $addressId
            required = $false
            description = "用户地址"
        },
        @{
            name = "bio"
            type = "string"
            required = $false
            description = "个人简介"
        }
    )
} | ConvertTo-Json -Depth 10

$profileResponse = Invoke-RestMethod -Uri "$BaseUrl/custom-types" -Method Post -Body $profileBody -ContentType "application/json"
$profileId = $profileResponse.id
Write-Host "✓ UserProfile 类型创建成功，ID: $profileId" -ForegroundColor Green
Write-Host ""

# 5. 创建包含数组引用的类型 - Company
Write-Host "5. 创建 Company 类型（包含 User 数组）..." -ForegroundColor Yellow
$companyBody = @{
    app_id = $appId
    name = "Company"
    description = "公司信息"
    fields = @(
        @{
            name = "name"
            type = "string"
            required = $true
            description = "公司名称"
        },
        @{
            name = "employees"
            type = "custom"
            ref = $userId
            is_array = $true
            required = $false
            description = "员工列表（User 数组）"
        },
        @{
            name = "tags"
            type = "string"
            is_array = $true
            required = $false
            description = "公司标签"
        }
    )
} | ConvertTo-Json -Depth 10

$companyResponse = Invoke-RestMethod -Uri "$BaseUrl/custom-types" -Method Post -Body $companyBody -ContentType "application/json"
$companyId = $companyResponse.id
Write-Host "✓ Company 类型创建成功，ID: $companyId" -ForegroundColor Green
Write-Host ""

# 6. 查询所有类型
Write-Host "6. 查询应用下的所有类型..." -ForegroundColor Yellow
$types = Invoke-RestMethod -Uri "$BaseUrl/custom-types?app_id=$appId" -Method Get
$types | ConvertTo-Json -Depth 10
Write-Host ""

# 7. 创建使用自定义类型的接口
Write-Host "7. 创建接口（使用自定义类型参数）..." -ForegroundColor Yellow
$interfaceBody = @{
    app_id = $appId
    name = "CreateUserProfile"
    description = "创建用户档案"
    method = "POST"
    protocol = "http"
    url = "https://api.example.com/profiles"
    auth_type = "bearer"
    enabled = $true
    parameters = @(
        @{
            name = "profile"
            type = "custom"
            ref = $profileId
            location = "body"
            required = $true
            description = "用户档案信息"
        },
        @{
            name = "addresses"
            type = "custom"
            ref = $addressId
            is_array = $true
            location = "body"
            required = $false
            description = "地址列表"
        }
    )
} | ConvertTo-Json -Depth 10

$interfaceResponse = Invoke-RestMethod -Uri "$BaseUrl/interfaces" -Method Post -Body $interfaceBody -ContentType "application/json"
$interfaceId = $interfaceResponse.id
Write-Host "✓ 接口创建成功，ID: $interfaceId" -ForegroundColor Green
Write-Host ""

# 8. 测试删除被引用的类型（应该失败）
Write-Host "8. 测试删除被引用的 User 类型（应该失败）..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BaseUrl/custom-types/$userId" -Method Delete
    Write-Host "✗ 删除成功（不应该）" -ForegroundColor Red
} catch {
    Write-Host "✓ 删除失败（符合预期）：$($_.Exception.Message)" -ForegroundColor Green
}
Write-Host ""

# 9. 获取单个类型详情
Write-Host "9. 获取 Company 类型详情..." -ForegroundColor Yellow
$companyDetail = Invoke-RestMethod -Uri "$BaseUrl/custom-types/$companyId" -Method Get
$companyDetail | ConvertTo-Json -Depth 10
Write-Host ""

Write-Host "========== 测试完成 ==========" -ForegroundColor Cyan
Write-Host ""
Write-Host "总结：" -ForegroundColor White
Write-Host "- 创建了 4 个自定义类型（Address, User, UserProfile, Company）" -ForegroundColor White
Write-Host "- 测试了类型引用功能（UserProfile 引用 User 和 Address）" -ForegroundColor White
Write-Host "- 测试了数组功能（User.tags[], Company.employees[], Company.tags[]）" -ForegroundColor White
Write-Host "- 测试了接口参数使用自定义类型" -ForegroundColor White
Write-Host "- 测试了引用完整性检查（删除被引用的类型会失败）" -ForegroundColor White
Write-Host ""
Write-Host "请访问 http://localhost:8080 查看前端界面" -ForegroundColor Cyan
