# 简单的curl命令测试

如果你不想运行脚本，可以直接使用这些curl命令进行测试：

## 1. 检查服务器健康状态
```bash
curl http://localhost:9001/health
```

## 2. 用户1测试流程

### 步骤1: 建立SSE连接
```bash
# 在终端1中运行（保持连接）
curl -H "X-API-Key: abcdefg" \
     -H "Accept: text/event-stream" \
     "http://localhost:9001/mcp-server/sse?server_id=server_employee_info"
```

### 步骤2: 初始化Session（新终端）
```bash
# 2a. 发送initialize请求
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "init_1",
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-01-07",
      "capabilities": {
        "tools": {}
      },
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'

# 2b. 发送initialized通知
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized"
  }'
```

### 步骤3: 获取工具列表
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user1_tools",
    "method": "tools/list"
  }'
```

### 步骤4: 查询员工信息
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user1_query",
    "method": "tools/call",
    "params": {
      "name": "employee_query",
      "arguments": {
        "name": "张三"
      }
    }
  }'
```

### 步骤5: 查询员工地址
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user1_address",
    "method": "tools/call",
    "params": {
      "name": "employee_address",
      "arguments": {
        "name": "李四"
      }
    }
  }'
```

## 3. 用户2测试流程（同时进行）

### 步骤1: 建立SSE连接
```bash
# 在终端3中运行（保持连接）
curl -H "X-API-Key: abcdefg" \
     -H "Accept: text/event-stream" \
     "http://localhost:9001/mcp-server/sse?server_id=server_employee_info"
```

### 步骤2: 初始化Session（新终端）
```bash
# 2a. 发送initialize请求
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER2_TEST_002" \
  -d '{
    "jsonrpc": "2.0",
    "id": "init_2",
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-01-07",
      "capabilities": {
        "tools": {}
      },
      "clientInfo": {
        "name": "test-client-2",
        "version": "1.0.0"
      }
    }
  }'

# 2b. 发送initialized通知
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER2_TEST_002" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized"
  }'
```

### 步骤3: 获取工具列表
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER2_TEST_002" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user2_tools",
    "method": "tools/list"
  }'
```

### 步骤4: 查询员工电话
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER2_TEST_002" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user2_query",
    "method": "tools/call",
    "params": {
      "name": "employee_phone",
      "arguments": {
        "name": "王五"
      }
    }
  }'
```

### 步骤5: 查询完整员工信息
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER2_TEST_002" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user2_full",
    "method": "tools/call",
    "params": {
      "name": "employee_query",
      "arguments": {
        "name": "赵六"
      }
    }
  }'
```

## 4. 测试无效Session

```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=INVALID_SESSION_123" \
  -d '{
    "jsonrpc": "2.0",
    "id": "invalid_test",
    "method": "tools/list"
  }'
```

## 5. 测试Session混乱

尝试用用户1的sessionID调用用户2的查询：
```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "cross_session_test",
    "method": "tools/call",
    "params": {
      "name": "employee_query",
      "arguments": {
        "name": "钱七"
      }
    }
  }'
```

## 测试重点关注

1. **SessionID隔离**: USER1_TEST_001 和 USER2_TEST_002 的请求应该独立处理
2. **无效Session处理**: INVALID_SESSION_123 应该被拒绝
3. **服务器日志**: 检查session创建、路由和验证的日志
4. **并发处理**: 两个用户同时使用应该不会互相干扰

## 预期结果

- ✅ 有效sessionID的请求应该成功返回员工信息
- ❌ 无效sessionID的请求应该返回错误
- ✅ 每个sessionID应该独立维护状态
- ✅ 服务器日志应该显示正确的session管理信息

## 中文编码问题解决方案

如果遇到中文字符显示为乱码（如`����`），使用以下方法：

### 方法1: 使用Unicode转义序列
将中文字符转换为Unicode编码：
- 张三 → `\u5f20\u4e09`
- 李四 → `\u674e\u56db`  
- 王五 → `\u738b\u4e94`
- 赵六 → `\u8d75\u516d`
- 钱七 → `\u94b1\u4e03`

```bash
curl -X POST \
  -H "X-API-Key: abcdefg" \
  -H "Content-Type: application/json" \
  "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" \
  -d '{
    "jsonrpc": "2.0",
    "id": "user1_query",
    "method": "tools/call",
    "params": {
      "name": "employee_query",
      "arguments": {
        "name": "\u674e\u56db"
      }
    }
  }'
```

### 方法2: 使用PowerShell (Windows)
```powershell
$body = @{
    jsonrpc = "2.0"
    id = "user1_query"
    method = "tools/call"
    params = @{
        name = "employee_query"
        arguments = @{
            name = "李四"
        }
    }
} | ConvertTo-Json -Depth 3

$headers = @{
    "X-API-Key" = "abcdefg"
    "Content-Type" = "application/json"
}

Invoke-RestMethod -Uri "http://localhost:9001/mcp-server/sse?sessionid=USER1_TEST_001" -Method POST -Headers $headers -Body ([System.Text.Encoding]::UTF8.GetBytes($body))
```

## 替换配置

请根据你的实际配置替换：
- `your-api-key-here` -> 你的实际API密钥
- `localhost:8080` -> 你的服务器地址
- 如果认证方式不同，请调整`X-API-Key`头部
