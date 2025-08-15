-- STDIO 服务复用策略示例配置

-- 清理现有的示例数据
DELETE FROM "public"."mcp_service_stdio" WHERE "server_id" LIKE '%-demo';

-- 1. 共享模式 - 时间服务 (适合无状态、轻量级服务)
INSERT INTO "public"."mcp_service_stdio" (
    "server_id", 
    "command", 
    "args", 
    "workdir", 
    "env", 
    "startup_timeout_ms", 
    "shutdown_timeout_ms", 
    "reuse_strategy", 
    "max_concurrent",
    "idle_ttl_ms", 
    "max_restarts", 
    "init_params"
) VALUES (
    'time-shared-demo',
    'uvx',
    '["mcp-server-time"]',
    NULL,
    '{}',
    30000,   -- 30秒启动超时
    5000,    -- 5秒关闭超时
    'shared', -- 所有用户共享同一个进程
    100,     -- 最大100个并发连接
    300000,  -- 5分钟空闲超时
    3,       -- 最大重启3次
    '{}'
);

-- 2. 每用户独立模式 - 个人助手服务 (适合有用户状态的服务)
INSERT INTO "public"."mcp_service_stdio" (
    "server_id", 
    "command", 
    "args", 
    "workdir", 
    "env", 
    "startup_timeout_ms", 
    "shutdown_timeout_ms", 
    "reuse_strategy", 
    "max_concurrent",
    "idle_ttl_ms", 
    "max_restarts", 
    "init_params"
) VALUES (
    'assistant-per-user-demo',
    'python',
    '["-m", "personal_assistant"]',
    NULL,
    '{"USER_CONTEXT": "enabled"}',
    60000,     -- 60秒启动超时
    15000,     -- 15秒关闭超时
    'per_user', -- 每个用户独立进程
    5,         -- 每个用户最多5个连接
    1800000,   -- 30分钟空闲超时
    2,         -- 最大重启2次
    '{"enable_memory": true}'
);

-- 3. 每会话独立模式 - 数据分析服务 (适合需要会话隔离的服务)
INSERT INTO "public"."mcp_service_stdio" (
    "server_id", 
    "command", 
    "args", 
    "workdir", 
    "env", 
    "startup_timeout_ms", 
    "shutdown_timeout_ms", 
    "reuse_strategy", 
    "max_concurrent",
    "idle_ttl_ms", 
    "max_restarts", 
    "init_params"
) VALUES (
    'datalab-per-session-demo',
    'jupyter',
    '["kernel", "--kernel=python3"]',
    '/tmp/datalab',
    '{"JUPYTER_CONFIG_DIR": "/tmp/jupyter"}',
    90000,       -- 90秒启动超时
    30000,       -- 30秒关闭超时
    'per_session', -- 每个会话独立进程
    3,           -- 每个会话最多3个连接
    600000,      -- 10分钟空闲超时
    1,           -- 最大重启1次
    '{"isolation": true}'
);

-- 4. 计算密集型服务 - 限制并发数
INSERT INTO "public"."mcp_service_stdio" (
    "server_id", 
    "command", 
    "args", 
    "workdir", 
    "env", 
    "startup_timeout_ms", 
    "shutdown_timeout_ms", 
    "reuse_strategy", 
    "max_concurrent",
    "idle_ttl_ms", 
    "max_restarts", 
    "init_params"
) VALUES (
    'heavy-compute-demo',
    'python',
    '["-m", "heavy_computation_service"]',
    NULL,
    '{"OMP_NUM_THREADS": "4"}',
    120000,    -- 120秒启动超时
    60000,     -- 60秒关闭超时
    'shared',  -- 共享模式
    2,         -- 严格限制只有2个并发连接
    120000,    -- 2分钟空闲超时
    1,         -- 最大重启1次
    '{"max_memory": "2GB"}'
);

-- 显示所有配置
SELECT 
    server_id,
    reuse_strategy,
    max_concurrent,
    idle_ttl_ms / 1000 as idle_timeout_seconds,
    command || ' ' || args::text as full_command
FROM "public"."mcp_service_stdio" 
WHERE server_id LIKE '%-demo'
ORDER BY reuse_strategy, max_concurrent;
