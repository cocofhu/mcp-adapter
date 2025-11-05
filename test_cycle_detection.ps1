# 测试循环引用检测功能
$baseUrl = "http://localhost:8080"

Write-Host "=== 测试自定义类型循环引用检测 ===" -ForegroundColor Cyan

# 1. 创建应用
Write-Host "`n1. 创建测试应用..." -ForegroundColor Yellow
$appResponse = Invoke-RestMethod -Uri "$baseUrl/api/applications" -Method Post -ContentType "application/json" -Body (@{
    name = "循环引用测试应用"
    description = "测试DFS判环功能"
    path = "cycle-test"
    protocol = "sse"
    enabled = $true
} | ConvertTo-Json)

$appId = $appResponse.application.id
Write-Host "应用创建成功, ID: $appId" -ForegroundColor Green

# 2. 创建类型A
Write-Host "`n2. 创建类型A..." -ForegroundColor Yellow
$typeAResponse = Invoke-RestMethod -Uri "$baseUrl/api/custom-types" -Method Post -ContentType "application/json" -Body (@{
    app_id = $appId
    name = "TypeA"
    description = "类型A"
    fields = @(
        @{
            name = "name"
            type = "string"
            required = $true
            description = "名称"
        }
    )
} | ConvertTo-Json -Depth 10)

$typeAId = $typeAResponse.custom_type.id
Write-Host "类型A创建成功, ID: $typeAId" -ForegroundColor Green

# 3. 创建类型B
Write-Host "`n3. 创建类型B..." -ForegroundColor Yellow
$typeBResponse = Invoke-RestMethod -Uri "$baseUrl/api/custom-types" -Method Post -ContentType "application/json" -Body (@{
    app_id = $appId
    name = "TypeB"
    description = "类型B"
    fields = @(
        @{
            name = "refA"
            type = "custom"
            ref = $typeAId
            required = $false
            description = "引用类型A"
        }
    )
} | ConvertTo-Json -Depth 10)

$typeBId = $typeBResponse.custom_type.id
Write-Host "类型B创建成功, ID: $typeBId" -ForegroundColor Green

# 4. 尝试更新类型A,让它引用类型B (形成环: A -> B -> A)
Write-Host "`n4. 尝试更新类型A引用类型B (应该失败)..." -ForegroundColor Yellow
try {
    $updateResponse = Invoke-RestMethod -Uri "$baseUrl/api/custom-types/$typeAId" -Method Put -ContentType "application/json" -Body (@{
        fields = @(
            @{
                name = "name"
                type = "string"
                required = $true
                description = "名称"
            },
            @{
                name = "refB"
                type = "custom"
                ref = $typeBId
                required = $false
                description = "引用类型B"
            }
        )
    } | ConvertTo-Json -Depth 10)
    Write-Host "错误: 应该检测到循环引用但没有!" -ForegroundColor Red
} catch {
    $errorMessage = $_.Exception.Message
    if ($errorMessage -match "circular reference") {
        Write-Host "成功: 检测到循环引用!" -ForegroundColor Green
        Write-Host "错误信息: $errorMessage" -ForegroundColor Gray
    } else {
        Write-Host "错误: 失败原因不是循环引用: $errorMessage" -ForegroundColor Red
    }
}

# 5. 创建类型C (不形成环)
Write-Host "`n5. 创建类型C引用类型A (不形成环,应该成功)..." -ForegroundColor Yellow
try {
    $typeCResponse = Invoke-RestMethod -Uri "$baseUrl/api/custom-types" -Method Post -ContentType "application/json" -Body (@{
        app_id = $appId
        name = "TypeC"
        description = "类型C"
        fields = @(
            @{
                name = "refA"
                type = "custom"
                ref = $typeAId
                required = $false
                description = "引用类型A"
            }
        )
    } | ConvertTo-Json -Depth 10)
    $typeCId = $typeCResponse.custom_type.id
    Write-Host "成功: 类型C创建成功, ID: $typeCId" -ForegroundColor Green
} catch {
    Write-Host "错误: 类型C创建失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 6. 测试自引用 (A -> A)
Write-Host "`n6. 测试自引用 (应该失败)..." -ForegroundColor Yellow
try {
    $selfRefResponse = Invoke-RestMethod -Uri "$baseUrl/api/custom-types" -Method Post -ContentType "application/json" -Body (@{
        app_id = $appId
        name = "SelfRef"
        description = "自引用类型"
        fields = @(
            @{
                name = "self"
                type = "custom"
                ref = $typeAId
                required = $false
                description = "自引用"
            }
        )
    } | ConvertTo-Json -Depth 10)
    
    $selfRefId = $selfRefResponse.custom_type.id
    
    # 尝试更新为自引用
    $updateSelfRef = Invoke-RestMethod -Uri "$baseUrl/api/custom-types/$selfRefId" -Method Put -ContentType "application/json" -Body (@{
        fields = @(
            @{
                name = "self"
                type = "custom"
                ref = $selfRefId
                required = $false
                description = "自引用"
            }
        )
    } | ConvertTo-Json -Depth 10)
    
    Write-Host "错误: 应该检测到自引用循环但没有!" -ForegroundColor Red
} catch {
    $errorMessage = $_.Exception.Message
    if ($errorMessage -match "circular reference") {
        Write-Host "成功: 检测到自引用循环!" -ForegroundColor Green
    } else {
        Write-Host "其他错误: $errorMessage" -ForegroundColor Yellow
    }
}

Write-Host "`n=== 测试接口参数循环引用检测 ===" -ForegroundColor Cyan

# 7. 创建接口使用有循环引用的类型 (应该失败,因为B引用A)
Write-Host "`n7. 创建接口使用类型B (类型B引用了类型A,但不形成环,应该成功)..." -ForegroundColor Yellow
try {
    $interfaceResponse = Invoke-RestMethod -Uri "$baseUrl/api/interfaces" -Method Post -ContentType "application/json" -Body (@{
        app_id = $appId
        name = "TestInterface"
        description = "测试接口"
        protocol = "http"
        url = "http://example.com/api/test"
        method = "POST"
        auth_type = "none"
        enabled = $true
        parameters = @(
            @{
                name = "data"
                type = "custom"
                ref = $typeBId
                location = "body"
                required = $true
                description = "使用类型B"
            }
        )
    } | ConvertTo-Json -Depth 10)
    Write-Host "成功: 接口创建成功" -ForegroundColor Green
} catch {
    Write-Host "错误: 接口创建失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Cyan
Write-Host "清理测试数据..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId" -Method Delete
    Write-Host "测试应用已删除" -ForegroundColor Green
} catch {
    Write-Host "删除应用失败: $($_.Exception.Message)" -ForegroundColor Yellow
}
