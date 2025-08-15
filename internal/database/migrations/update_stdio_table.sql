-- 更新 mcp_service_stdio 表，添加最大并发数字段

-- 如果表已存在 max_concurrent 字段，先删除（为了重新创建）
DO $$ 
BEGIN 
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'mcp_service_stdio' 
        AND column_name = 'max_concurrent'
    ) THEN
        ALTER TABLE "public"."mcp_service_stdio" DROP COLUMN "max_concurrent";
    END IF;
END $$;

-- 添加 max_concurrent 字段
ALTER TABLE "public"."mcp_service_stdio" 
ADD COLUMN "max_concurrent" int NOT NULL DEFAULT 10 
COMMENT '最大并发连接数，0表示无限制';

-- 更新复用策略字段的注释，明确不同策略的含义
COMMENT ON COLUMN "public"."mcp_service_stdio"."reuse_strategy" IS '复用策略: shared(共享进程), per_user(每用户独立), per_session(每会话独立)';

-- 为现有记录设置合理的默认值
UPDATE "public"."mcp_service_stdio" 
SET 
    "reuse_strategy" = 'shared',
    "max_concurrent" = CASE 
        WHEN "server_id" LIKE '%time%' THEN 50      -- 时间服务，轻量级
        WHEN "server_id" LIKE '%search%' THEN 20    -- 搜索服务，中等负载
        WHEN "server_id" LIKE '%ai%' THEN 5         -- AI服务，重负载
        ELSE 10                                     -- 默认值
    END
WHERE "reuse_strategy" IS NULL OR "reuse_strategy" = '';

-- 添加约束确保合理的配置
ALTER TABLE "public"."mcp_service_stdio" 
ADD CONSTRAINT "check_max_concurrent" 
CHECK ("max_concurrent" >= 0 AND "max_concurrent" <= 1000);

ALTER TABLE "public"."mcp_service_stdio" 
ADD CONSTRAINT "check_reuse_strategy" 
CHECK ("reuse_strategy" IN ('shared', 'per_user', 'per_session'));

-- 创建示例配置
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
) VALUES 
-- 时间服务：高并发，共享进程
('time-service-demo', 'uvx', '["mcp-server-time"]', NULL, '{}', 30000, 5000, 'shared', 50, 300000, 3, '{}'),

-- 计算服务：中等并发，共享进程  
('calc-service-demo', 'python', '["-m", "calculator_mcp"]', NULL, '{}', 45000, 10000, 'shared', 20, 180000, 2, '{}'),

-- 个人助手：低并发，每用户独立
('personal-assistant-demo', 'node', '["personal-assistant.js"]', NULL, '{}', 60000, 15000, 'per_user', 5, 900000, 1, '{}')

ON CONFLICT (server_id) DO UPDATE SET
    "reuse_strategy" = EXCLUDED."reuse_strategy",
    "max_concurrent" = EXCLUDED."max_concurrent";
