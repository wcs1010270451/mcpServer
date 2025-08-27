#!/bin/bash

# 测试多用户并发使用同一个handler的脚本
# 用于验证session隔离是否正常工作

SERVER_URL="http://localhost:8080"
AUTH_HEADER="X-API-Key: your-api-key-here"  # 根据你的认证配置修改

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== 多用户并发Session测试 ===${NC}"
echo "测试场景：多个用户同时连接 server_employee_info"
echo

# 用户1的测试函数
test_user1() {
    local user_name="用户1"
    local session_file="/tmp/user1_session.txt"
    
    echo -e "${GREEN}[${user_name}] 开始测试${NC}"
    
    # 1. 建立连接 (GET请求)
    echo -e "${YELLOW}[${user_name}] 步骤1: 建立SSE连接${NC}"
    curl -s -H "$AUTH_HEADER" \
         -H "Accept: text/event-stream" \
         "$SERVER_URL/mcp-server/sse?server_id=server_employee_info" \
         > "$session_file" &
    
    local sse_pid=$!
    sleep 2  # 等待连接建立
    
    # 2. 生成随机sessionID（模拟客户端行为）
    local session_id="USER1_$(date +%s)_$(shuf -i 1000-9999 -n 1)"
    echo -e "${YELLOW}[${user_name}] 使用SessionID: $session_id${NC}"
    
    # 3. 获取工具列表
    echo -e "${YELLOW}[${user_name}] 步骤2: 获取工具列表${NC}"
    local tools_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user1_tools",
            "method": "tools/list"
        }')
    
    echo -e "${GREEN}[${user_name}] 工具列表响应:${NC}"
    echo "$tools_response" | jq '.' 2>/dev/null || echo "$tools_response"
    echo
    
    # 4. 使用工具查询员工信息
    echo -e "${YELLOW}[${user_name}] 步骤3: 查询张三的信息${NC}"
    local query_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user1_query",
            "method": "tools/call",
            "params": {
                "name": "employee_query",
                "arguments": {
                    "name": "\u5f20\u4e09"
                }
            }
        }')
    
    echo -e "${GREEN}[${user_name}] 查询张三结果:${NC}"
    echo "$query_response" | jq '.' 2>/dev/null || echo "$query_response"
    echo
    
    # 5. 第二次查询（测试session状态保持）
    echo -e "${YELLOW}[${user_name}] 步骤4: 查询李四的地址${NC}"
    local address_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user1_address",
            "method": "tools/call",
            "params": {
                "name": "employee_address",
                "arguments": {
                    "name": "\u674e\u56db"
                }
            }
        }')
    
    echo -e "${GREEN}[${user_name}] 查询李四地址结果:${NC}"
    echo "$address_response" | jq '.' 2>/dev/null || echo "$address_response"
    echo
    
    # 清理
    kill $sse_pid 2>/dev/null
    rm -f "$session_file"
    
    echo -e "${GREEN}[${user_name}] 测试完成${NC}"
    echo "----------------------------------------"
}

# 用户2的测试函数
test_user2() {
    local user_name="用户2"
    local session_file="/tmp/user2_session.txt"
    
    echo -e "${GREEN}[${user_name}] 开始测试${NC}"
    
    # 1. 建立连接 (GET请求)
    echo -e "${YELLOW}[${user_name}] 步骤1: 建立SSE连接${NC}"
    curl -s -H "$AUTH_HEADER" \
         -H "Accept: text/event-stream" \
         "$SERVER_URL/mcp-server/sse?server_id=server_employee_info" \
         > "$session_file" &
    
    local sse_pid=$!
    sleep 2  # 等待连接建立
    
    # 2. 生成随机sessionID（模拟客户端行为）
    local session_id="USER2_$(date +%s)_$(shuf -i 1000-9999 -n 1)"
    echo -e "${YELLOW}[${user_name}] 使用SessionID: $session_id${NC}"
    
    # 3. 获取工具列表
    echo -e "${YELLOW}[${user_name}] 步骤2: 获取工具列表${NC}"
    local tools_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user2_tools",
            "method": "tools/list"
        }')
    
    echo -e "${GREEN}[${user_name}] 工具列表响应:${NC}"
    echo "$tools_response" | jq '.' 2>/dev/null || echo "$tools_response"
    echo
    
    # 4. 使用工具查询员工信息
    echo -e "${YELLOW}[${user_name}] 步骤3: 查询王五的电话${NC}"
    local query_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user2_query",
            "method": "tools/call",
            "params": {
                "name": "employee_phone",
                "arguments": {
                    "name": "\u738b\u4e94"
                }
            }
        }')
    
    echo -e "${GREEN}[${user_name}] 查询王五电话结果:${NC}"
    echo "$query_response" | jq '.' 2>/dev/null || echo "$query_response"
    echo
    
    # 5. 第二次查询（测试session状态保持）
    echo -e "${YELLOW}[${user_name}] 步骤4: 查询赵六的完整信息${NC}"
    local full_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user2_full",
            "method": "tools/call",
            "params": {
                "name": "employee_query",
                "arguments": {
                    "name": "\u8d75\u516d"
                }
            }
        }')
    
    echo -e "${GREEN}[${user_name}] 查询赵六完整信息结果:${NC}"
    echo "$full_response" | jq '.' 2>/dev/null || echo "$full_response"
    echo
    
    # 清理
    kill $sse_pid 2>/dev/null
    rm -f "$session_file"
    
    echo -e "${GREEN}[${user_name}] 测试完成${NC}"
    echo "----------------------------------------"
}

# 用户3的测试函数（测试错误的sessionID）
test_user3_invalid() {
    local user_name="用户3(无效Session)"
    
    echo -e "${GREEN}[${user_name}] 开始测试${NC}"
    
    # 使用一个无效的sessionID
    local invalid_session_id="INVALID_SESSION_12345"
    echo -e "${YELLOW}[${user_name}] 使用无效SessionID: $invalid_session_id${NC}"
    
    # 尝试直接使用工具（应该失败）
    echo -e "${YELLOW}[${user_name}] 步骤1: 使用无效session查询工具${NC}"
    local error_response=$(curl -s -X POST \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        "$SERVER_URL/mcp-server/sse?sessionid=$invalid_session_id" \
        -d '{
            "jsonrpc": "2.0",
            "id": "user3_invalid",
            "method": "tools/list"
        }')
    
    echo -e "${RED}[${user_name}] 无效session响应 (应该失败):${NC}"
    echo "$error_response"
    echo
    
    echo -e "${GREEN}[${user_name}] 测试完成${NC}"
    echo "----------------------------------------"
}

# 主测试流程
main() {
    echo "检查服务器是否运行..."
    if ! curl -s "$SERVER_URL/health" > /dev/null; then
        echo -e "${RED}错误: 服务器未运行或地址不正确${NC}"
        echo "请确保服务器在 $SERVER_URL 运行"
        exit 1
    fi
    
    echo -e "${GREEN}服务器运行正常${NC}"
    echo
    
    echo "开始并发测试..."
    echo
    
    # 并发执行用户1和用户2的测试
    echo -e "${BLUE}=== 并发执行用户1和用户2 ===${NC}"
    test_user1 &
    test_user2 &
    
    # 等待并发测试完成
    wait
    
    echo -e "${BLUE}=== 测试无效Session ===${NC}"
    test_user3_invalid
    
    echo -e "${BLUE}=== 测试完成 ===${NC}"
    echo "请检查日志中的session管理信息："
    echo "1. 两个用户应该有不同的sessionID"
    echo "2. 每个用户的请求应该正确路由到自己的session"
    echo "3. 无效session应该被拒绝"
    echo "4. 服务器日志应该显示正确的session创建和路由信息"
}

# 使用说明
usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -u URL     设置服务器URL (默认: http://localhost:8080)"
    echo "  -k KEY     设置API密钥 (默认: your-api-key-here)"
    echo "  -h         显示帮助"
    echo
    echo "示例:"
    echo "  $0"
    echo "  $0 -u http://localhost:8080 -k mykey123"
}

# 解析命令行参数
while getopts "u:k:h" opt; do
    case $opt in
        u)
            SERVER_URL="$OPTARG"
            ;;
        k)
            AUTH_HEADER="X-API-Key: $OPTARG"
            ;;
        h)
            usage
            exit 0
            ;;
        \?)
            echo "无效选项: -$OPTARG" >&2
            usage
            exit 1
            ;;
    esac
done

# 检查依赖
if ! command -v curl &> /dev/null; then
    echo -e "${RED}错误: 需要安装 curl${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}警告: 建议安装 jq 以美化JSON输出${NC}"
fi

# 执行主函数
main
