# PowerShell版本的并发Session测试脚本
# 用于Windows环境测试多用户并发使用同一个handler

param(
    [string]$ServerUrl = "http://localhost:8080",
    [string]$ApiKey = "your-api-key-here",
    [switch]$Help
)

if ($Help) {
    Write-Host "用法: .\test_concurrent_sessions.ps1 [参数]"
    Write-Host "参数:"
    Write-Host "  -ServerUrl URL    设置服务器URL (默认: http://localhost:8080)"
    Write-Host "  -ApiKey KEY       设置API密钥 (默认: your-api-key-here)"
    Write-Host "  -Help             显示帮助"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\test_concurrent_sessions.ps1"
    Write-Host "  .\test_concurrent_sessions.ps1 -ServerUrl http://localhost:8080 -ApiKey mykey123"
    exit 0
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    $oldColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $Color
    Write-Host $Message
    $Host.UI.RawUI.ForegroundColor = $oldColor
}

Write-ColorOutput "=== 多用户并发Session测试 ===" "Cyan"
Write-Host "测试场景：多个用户同时连接 server_employee_info"
Write-Host ""

# 用户1测试函数
function Test-User1 {
    $userName = "用户1"
    
    Write-ColorOutput "[$userName] 开始测试" "Green"
    
    # 生成随机sessionID
    $sessionId = "USER1_$(Get-Date -Format 'yyyyMMddHHmmss')_$(Get-Random -Minimum 1000 -Maximum 9999)"
    Write-ColorOutput "[$userName] 使用SessionID: $sessionId" "Yellow"
    
    # 创建请求头
    $headers = @{
        "X-API-Key" = $ApiKey.Replace("X-API-Key: ", "")
        "Content-Type" = "application/json"
    }
    
    try {
        # 1. 建立SSE连接（后台任务）
        Write-ColorOutput "[$userName] 步骤1: 建立SSE连接" "Yellow"
        $sseHeaders = @{
            "X-API-Key" = $ApiKey.Replace("X-API-Key: ", "")
            "Accept" = "text/event-stream"
        }
        
        # 启动后台SSE连接
        $sseJob = Start-Job -ScriptBlock {
            param($url, $headers)
            try {
                Invoke-WebRequest -Uri "$url/mcp-server/sse?server_id=server_employee_info" -Headers $headers -TimeoutSec 30
            } catch {
                Write-Host "SSE连接错误: $_"
            }
        } -ArgumentList $ServerUrl, $sseHeaders
        
        Start-Sleep -Seconds 2  # 等待连接建立
        
        # 2. 获取工具列表
        Write-ColorOutput "[$userName] 步骤2: 获取工具列表" "Yellow"
        $toolsBody = @{
            jsonrpc = "2.0"
            id = "user1_tools"
            method = "tools/list"
        } | ConvertTo-Json
        
        $toolsResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $toolsBody
        
        Write-ColorOutput "[$userName] 工具列表响应:" "Green"
        $toolsResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
        # 3. 查询员工信息
        Write-ColorOutput "[$userName] 步骤3: 查询张三的信息" "Yellow"
        $queryBody = @{
            jsonrpc = "2.0"
            id = "user1_query"
            method = "tools/call"
            params = @{
                name = "employee_query"
                arguments = @{
                    name = "张三"
                }
            }
        } | ConvertTo-Json -Depth 3
        
        $queryResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $queryBody
        
        Write-ColorOutput "[$userName] 查询张三结果:" "Green"
        $queryResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
        # 4. 第二次查询
        Write-ColorOutput "[$userName] 步骤4: 查询李四的地址" "Yellow"
        $addressBody = @{
            jsonrpc = "2.0"
            id = "user1_address"
            method = "tools/call"
            params = @{
                name = "employee_address"
                arguments = @{
                    name = "李四"
                }
            }
        } | ConvertTo-Json -Depth 3
        
        $addressResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $addressBody
        
        Write-ColorOutput "[$userName] 查询李四地址结果:" "Green"
        $addressResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
    } catch {
        Write-ColorOutput "[$userName] 错误: $_" "Red"
    } finally {
        # 清理后台任务
        if ($sseJob) {
            Stop-Job $sseJob -ErrorAction SilentlyContinue
            Remove-Job $sseJob -ErrorAction SilentlyContinue
        }
    }
    
    Write-ColorOutput "[$userName] 测试完成" "Green"
    Write-Host "----------------------------------------"
}

# 用户2测试函数
function Test-User2 {
    $userName = "用户2"
    
    Write-ColorOutput "[$userName] 开始测试" "Green"
    
    # 生成随机sessionID
    $sessionId = "USER2_$(Get-Date -Format 'yyyyMMddHHmmss')_$(Get-Random -Minimum 1000 -Maximum 9999)"
    Write-ColorOutput "[$userName] 使用SessionID: $sessionId" "Yellow"
    
    # 创建请求头
    $headers = @{
        "X-API-Key" = $ApiKey.Replace("X-API-Key: ", "")
        "Content-Type" = "application/json"
    }
    
    try {
        # 1. 建立SSE连接（后台任务）
        Write-ColorOutput "[$userName] 步骤1: 建立SSE连接" "Yellow"
        $sseHeaders = @{
            "X-API-Key" = $ApiKey.Replace("X-API-Key: ", "")
            "Accept" = "text/event-stream"
        }
        
        # 启动后台SSE连接
        $sseJob = Start-Job -ScriptBlock {
            param($url, $headers)
            try {
                Invoke-WebRequest -Uri "$url/mcp-server/sse?server_id=server_employee_info" -Headers $headers -TimeoutSec 30
            } catch {
                Write-Host "SSE连接错误: $_"
            }
        } -ArgumentList $ServerUrl, $sseHeaders
        
        Start-Sleep -Seconds 2  # 等待连接建立
        
        # 2. 获取工具列表
        Write-ColorOutput "[$userName] 步骤2: 获取工具列表" "Yellow"
        $toolsBody = @{
            jsonrpc = "2.0"
            id = "user2_tools"
            method = "tools/list"
        } | ConvertTo-Json
        
        $toolsResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $toolsBody
        
        Write-ColorOutput "[$userName] 工具列表响应:" "Green"
        $toolsResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
        # 3. 查询员工信息
        Write-ColorOutput "[$userName] 步骤3: 查询王五的电话" "Yellow"
        $queryBody = @{
            jsonrpc = "2.0"
            id = "user2_query"
            method = "tools/call"
            params = @{
                name = "employee_phone"
                arguments = @{
                    name = "王五"
                }
            }
        } | ConvertTo-Json -Depth 3
        
        $queryResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $queryBody
        
        Write-ColorOutput "[$userName] 查询王五电话结果:" "Green"
        $queryResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
        # 4. 第二次查询
        Write-ColorOutput "[$userName] 步骤4: 查询赵六的完整信息" "Yellow"
        $fullBody = @{
            jsonrpc = "2.0"
            id = "user2_full"
            method = "tools/call"
            params = @{
                name = "employee_query"
                arguments = @{
                    name = "赵六"
                }
            }
        } | ConvertTo-Json -Depth 3
        
        $fullResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$sessionId" -Method POST -Headers $headers -Body $fullBody
        
        Write-ColorOutput "[$userName] 查询赵六完整信息结果:" "Green"
        $fullResponse | ConvertTo-Json -Depth 3
        Write-Host ""
        
    } catch {
        Write-ColorOutput "[$userName] 错误: $_" "Red"
    } finally {
        # 清理后台任务
        if ($sseJob) {
            Stop-Job $sseJob -ErrorAction SilentlyContinue
            Remove-Job $sseJob -ErrorAction SilentlyContinue
        }
    }
    
    Write-ColorOutput "[$userName] 测试完成" "Green"
    Write-Host "----------------------------------------"
}

# 用户3测试函数（无效session）
function Test-User3Invalid {
    $userName = "用户3(无效Session)"
    
    Write-ColorOutput "[$userName] 开始测试" "Green"
    
    # 使用无效sessionID
    $invalidSessionId = "INVALID_SESSION_12345"
    Write-ColorOutput "[$userName] 使用无效SessionID: $invalidSessionId" "Yellow"
    
    # 创建请求头
    $headers = @{
        "X-API-Key" = $ApiKey.Replace("X-API-Key: ", "")
        "Content-Type" = "application/json"
    }
    
    try {
        # 尝试使用无效session
        Write-ColorOutput "[$userName] 步骤1: 使用无效session查询工具" "Yellow"
        $toolsBody = @{
            jsonrpc = "2.0"
            id = "user3_invalid"
            method = "tools/list"
        } | ConvertTo-Json
        
        $errorResponse = Invoke-RestMethod -Uri "$ServerUrl/mcp-server/sse?sessionid=$invalidSessionId" -Method POST -Headers $headers -Body $toolsBody -ErrorAction Stop
        
        Write-ColorOutput "[$userName] 意外成功响应 (应该失败):" "Red"
        $errorResponse | ConvertTo-Json -Depth 3
        
    } catch {
        Write-ColorOutput "[$userName] 预期的错误响应 (正常):" "Red"
        Write-Host $_.Exception.Message
    }
    
    Write-Host ""
    Write-ColorOutput "[$userName] 测试完成" "Green"
    Write-Host "----------------------------------------"
}

# 主测试函数
function Main {
    Write-Host "检查服务器是否运行..."
    
    try {
        $healthResponse = Invoke-RestMethod -Uri "$ServerUrl/health" -TimeoutSec 5
        Write-ColorOutput "服务器运行正常" "Green"
    } catch {
        Write-ColorOutput "错误: 服务器未运行或地址不正确" "Red"
        Write-Host "请确保服务器在 $ServerUrl 运行"
        exit 1
    }
    
    Write-Host ""
    Write-Host "开始并发测试..."
    Write-Host ""
    
    # 并发执行用户1和用户2的测试
    Write-ColorOutput "=== 并发执行用户1和用户2 ===" "Cyan"
    
    $job1 = Start-Job -ScriptBlock ${function:Test-User1}
    $job2 = Start-Job -ScriptBlock ${function:Test-User2}
    
    # 等待并发测试完成
    Wait-Job $job1, $job2 | Out-Null
    
    # 显示结果
    Receive-Job $job1
    Receive-Job $job2
    
    # 清理任务
    Remove-Job $job1, $job2
    
    Write-ColorOutput "=== 测试无效Session ===" "Cyan"
    Test-User3Invalid
    
    Write-ColorOutput "=== 测试完成 ===" "Cyan"
    Write-Host "请检查服务器日志中的session管理信息："
    Write-Host "1. 两个用户应该有不同的sessionID"
    Write-Host "2. 每个用户的请求应该正确路由到自己的session"
    Write-Host "3. 无效session应该被拒绝"
    Write-Host "4. 服务器日志应该显示正确的session创建和路由信息"
}

# 运行主函数
Main
