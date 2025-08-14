-- 插入必应中文MCP服务配置

-- 1. 首先在 mcp_service 表中插入基础服务信息
INSERT INTO "public"."mcp_service" (
    "server_id", 
    "display_name", 
    "implementation_name", 
    "protocol_version", 
    "enabled", 
    "metadata", 
    "adapter", 
    "start_mode"
) VALUES (
    'bing-cn-mcp-server', 
    'Bing CN MCP Server', 
    'bing-cn-mcp-server', 
    '2024-11-05', 
    true, 
    '{"description": "ModelScope托管的必应中文MCP服务", "provider": "ModelScope", "url": "https://mcp.api-inference.modelscope.net"}', 
    'remote_sse', 
    'on_demand'
);

-- 2. 然后在 mcp_service_sse 表中插入SSE特定配置
INSERT INTO "public"."mcp_service_sse" (
    "server_id",
    "base_url",
    "sse_path",
    "auth_type",
    "auth_config",
    "timeout_ms",
    "connect_timeout_ms",
    "retry_attempts",
    "retry_delay_ms",
    "health_check_enabled",
    "health_check_path",
    "headers",
    "query_params",
    "user_agent"
) VALUES (
    'bing-cn-mcp-server',
    'https://mcp.api-inference.modelscope.net',
    '/6675c445a78944/sse',
    'none',
    '{}',
    300000, -- 5分钟超时（适合长时间流式传输）
    30000,  -- 30秒连接超时
    2,      -- 重试2次
    3000,   -- 重试间隔3秒
    false,  -- 不启用健康检查（避免对公共服务造成压力）
    null,
    '{"Accept": "text/event-stream", "Cache-Control": "no-cache", "Connection": "keep-alive"}',
    '{}',
    'MCP-Proxy/1.0'
);

-- 验证插入结果
SELECT 
    s.server_id,
    s.display_name,
    s.adapter,
    s.enabled,
    sse.base_url,
    sse.sse_path,
    sse.auth_type,
    sse.timeout_ms
FROM mcp_service s
JOIN mcp_service_sse sse ON s.server_id = sse.server_id
WHERE s.server_id = 'bing-cn-mcp-server';

-- 如果需要删除（测试用）
-- DELETE FROM mcp_service WHERE server_id = 'bing-cn-mcp-server';
