# 内存泄漏测试脚本
# 用于验证修复效果

param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$AppCount = 100,
    [int]$Iterations = 5
)

Write-Host "================================" -ForegroundColor Cyan
Write-Host "内存泄漏压力测试" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Apps per iteration: $AppCount" -ForegroundColor Yellow
Write-Host "Iterations: $Iterations" -ForegroundColor Yellow
Write-Host ""

function Test-ServerHealth {
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method GET -TimeoutSec 5
        return $true
    } catch {
        return $false
    }
}

function Create-Applications {
    param([int]$Count, [int]$Offset)
    
    Write-Host "Creating $Count applications..." -ForegroundColor Green
    $created = 0
    
    for ($i = 1; $i -le $Count; $i++) {
        $id = $Offset + $i
        $body = @{
            name = "test_app_$id"
            path = "test_path_$id"
            protocol = "sse"
            description = "Test application $id"
        } | ConvertTo-Json
        
        try {
            $response = Invoke-RestMethod -Uri "$BaseUrl/api/applications" `
                -Method POST `
                -ContentType "application/json" `
                -Body $body `
                -TimeoutSec 10
            $created++
            
            if ($i % 10 -eq 0) {
                Write-Host "  Created $i/$Count applications..." -ForegroundColor Gray
            }
        } catch {
            Write-Host "  Failed to create app $id: $_" -ForegroundColor Red
        }
    }
    
    Write-Host "  Successfully created $created applications" -ForegroundColor Green
    return $created
}

function Update-Applications {
    param([int]$Count, [int]$Offset)
    
    Write-Host "Updating $Count applications..." -ForegroundColor Green
    $updated = 0
    
    # 获取所有应用
    try {
        $apps = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method GET
        $appList = $apps.applications
        
        foreach ($app in $appList) {
            $body = @{
                description = "Updated at $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            } | ConvertTo-Json
            
            try {
                Invoke-RestMethod -Uri "$BaseUrl/api/applications/$($app.id)" `
                    -Method PUT `
                    -ContentType "application/json" `
                    -Body $body `
                    -TimeoutSec 10 | Out-Null
                $updated++
            } catch {
                Write-Host "  Failed to update app $($app.id): $_" -ForegroundColor Red
            }
        }
    } catch {
        Write-Host "  Failed to get applications: $_" -ForegroundColor Red
    }
    
    Write-Host "  Successfully updated $updated applications" -ForegroundColor Green
    return $updated
}

function Delete-Applications {
    Write-Host "Deleting all applications..." -ForegroundColor Green
    $deleted = 0
    
    try {
        $apps = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method GET
        $appList = $apps.applications
        
        foreach ($app in $appList) {
            try {
                Invoke-RestMethod -Uri "$BaseUrl/api/applications/$($app.id)" `
                    -Method DELETE `
                    -TimeoutSec 10 | Out-Null
                $deleted++
            } catch {
                Write-Host "  Failed to delete app $($app.id): $_" -ForegroundColor Red
            }
        }
    } catch {
        Write-Host "  Failed to get applications: $_" -ForegroundColor Red
    }
    
    Write-Host "  Successfully deleted $deleted applications" -ForegroundColor Green
    return $deleted
}

function Get-ApplicationCount {
    try {
        $apps = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method GET
        return $apps.applications.Count
    } catch {
        return 0
    }
}

# 主测试流程
Write-Host "Checking server health..." -ForegroundColor Cyan
if (-not (Test-ServerHealth)) {
    Write-Host "ERROR: Server is not responding at $BaseUrl" -ForegroundColor Red
    Write-Host "Please ensure the server is running before running this test." -ForegroundColor Yellow
    exit 1
}
Write-Host "Server is healthy!" -ForegroundColor Green
Write-Host ""

# 清理现有数据
Write-Host "Cleaning up existing data..." -ForegroundColor Cyan
Delete-Applications | Out-Null
Write-Host ""

# 开始测试迭代
for ($iteration = 1; $iteration -le $Iterations; $iteration++) {
    Write-Host "================================" -ForegroundColor Magenta
    Write-Host "Iteration $iteration of $Iterations" -ForegroundColor Magenta
    Write-Host "================================" -ForegroundColor Magenta
    
    $offset = ($iteration - 1) * $AppCount
    
    # 创建应用
    $created = Create-Applications -Count $AppCount -Offset $offset
    Start-Sleep -Seconds 2
    
    # 更新应用（触发事件）
    $updated = Update-Applications -Count $AppCount -Offset $offset
    Start-Sleep -Seconds 2
    
    # 删除应用（测试资源释放）
    $deleted = Delete-Applications
    Start-Sleep -Seconds 2
    
    # 验证清理
    $remaining = Get-ApplicationCount
    
    Write-Host ""
    Write-Host "Iteration $iteration Summary:" -ForegroundColor Cyan
    Write-Host "  Created: $created" -ForegroundColor White
    Write-Host "  Updated: $updated" -ForegroundColor White
    Write-Host "  Deleted: $deleted" -ForegroundColor White
    Write-Host "  Remaining: $remaining" -ForegroundColor White
    
    if ($remaining -gt 0) {
        Write-Host "  WARNING: $remaining applications not deleted!" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ All applications cleaned up successfully" -ForegroundColor Green
    }
    
    Write-Host ""
    
    # 等待下一次迭代
    if ($iteration -lt $Iterations) {
        Write-Host "Waiting 3 seconds before next iteration..." -ForegroundColor Gray
        Start-Sleep -Seconds 3
    }
}

Write-Host "================================" -ForegroundColor Cyan
Write-Host "测试完成!" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "请检查服务器日志和内存使用情况" -ForegroundColor Yellow
Write-Host "建议使用以下工具进行进一步分析:" -ForegroundColor Yellow
Write-Host "  1. pprof: http://localhost:8080/debug/pprof/" -ForegroundColor White
Write-Host "  2. 服务器日志中的内存统计信息" -ForegroundColor White
Write-Host ""
