# API 测试脚本
$baseUrl = "http://localhost:8080/api"

Write-Host "Testing MCP Adapter API..." -ForegroundColor Green

# 测试创建应用
Write-Host "`n1. Testing Application Creation..." -ForegroundColor Yellow
$appData = @{
    name = "测试应用"
    description = "这是一个测试应用"
    path = "v1.0.0"
    protocol = "http"
    enabled = $true
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/applications" -Method POST -Body $appData -ContentType "application/json"
    Write-Host "✓ Application created successfully: $($response.name)" -ForegroundColor Green
    $appId = $response.id
} catch {
    Write-Host "✗ Failed to create application: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 测试获取所有应用
Write-Host "`n2. Testing Get All Applications..." -ForegroundColor Yellow
try {
    $apps = Invoke-RestMethod -Uri "$baseUrl/applications" -Method GET
    Write-Host "✓ Retrieved $($apps.Count) applications" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to get applications: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试创建接口
Write-Host "`n3. Testing Interface Creation..." -ForegroundColor Yellow
$interfaceData = @{
    app_id = $appId
    name = "test_api"
    description = "测试接口"
    protocol = "HTTP"
    url = "https://api.example.com/test"
    auth_type = "none"
    enabled = $true
    options = '{"method":"GET","parameters":[],"defaultParams":[],"defaultHeaders":[]}'
} | ConvertTo-Json

try {
    $interface = Invoke-RestMethod -Uri "$baseUrl/interfaces" -Method POST -Body $interfaceData -ContentType "application/json"
    Write-Host "✓ Interface created successfully: $($interface.name)" -ForegroundColor Green
    $interfaceId = $interface.id
} catch {
    Write-Host "✗ Failed to create interface: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试获取所有接口
Write-Host "`n4. Testing Get All Interfaces..." -ForegroundColor Yellow
try {
    $interfaces = Invoke-RestMethod -Uri "$baseUrl/interfaces" -Method GET
    Write-Host "✓ Retrieved $($interfaces.Count) interfaces" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to get interfaces: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试更新接口
Write-Host "`n5. Testing Interface Update..." -ForegroundColor Yellow
$updateData = @{
    app_id = $appId
    name = "updated_test_api"
    description = "更新后的测试接口"
    protocol = "HTTP"
    url = "https://api.example.com/updated"
    auth_type = "bearer"
    enabled = $true
    options = '{"method":"POST","parameters":[],"defaultParams":[],"defaultHeaders":[]}'
} | ConvertTo-Json

try {
    $updatedInterface = Invoke-RestMethod -Uri "$baseUrl/interfaces/$interfaceId" -Method PUT -Body $updateData -ContentType "application/json"
    Write-Host "✓ Interface updated successfully: $($updatedInterface.name)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to update interface: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试删除接口
Write-Host "`n6. Testing Interface Deletion..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/interfaces/$interfaceId" -Method DELETE
    Write-Host "✓ Interface deleted successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to delete interface: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试删除应用
Write-Host "`n7. Testing Application Deletion..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/applications/$appId" -Method DELETE
    Write-Host "✓ Application deleted successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to delete application: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nAPI Testing Complete!" -ForegroundColor Green