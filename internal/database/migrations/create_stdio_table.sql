-- 创建 mcp_service_stdio 表（如果不存在）

-- 创建 mcp_service_stdio 表
CREATE TABLE IF NOT EXISTS "public"."mcp_service_stdio" (
    "server_id" varchar(255) NOT NULL,
    "command" varchar(500) NOT NULL,
    "args" jsonb NOT NULL DEFAULT '[]',
    "workdir" varchar(500),
    "env" jsonb NOT NULL DEFAULT '{}',
    "startup_timeout_ms" int NOT NULL DEFAULT 30000,
    "shutdown_timeout_ms" int NOT NULL DEFAULT 5000,
    "reuse_strategy" varchar(20) NOT NULL DEFAULT 'shared',
    "max_concurrent" int NOT NULL DEFAULT 10,
    "idle_ttl_ms" int NOT NULL DEFAULT 300000,
    "max_restarts" int NOT NULL DEFAULT 3,
    "init_params" jsonb NOT NULL DEFAULT '{}',
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    
    CONSTRAINT "mcp_service_stdio_pkey" PRIMARY KEY ("server_id"),
    CONSTRAINT "fk_stdio_service" FOREIGN KEY ("server_id") 
        REFERENCES "public"."mcp_service"("server_id") 
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "check_reuse_strategy" 
        CHECK ("reuse_strategy" IN ('shared', 'per_user', 'per_session')),
    CONSTRAINT "check_max_concurrent" 
        CHECK ("max_concurrent" >= 0 AND "max_concurrent" <= 1000),
    CONSTRAINT "check_startup_timeout" 
        CHECK ("startup_timeout_ms" > 0 AND "startup_timeout_ms" <= 300000),
    CONSTRAINT "check_shutdown_timeout" 
        CHECK ("shutdown_timeout_ms" > 0 AND "shutdown_timeout_ms" <= 60000),
    CONSTRAINT "check_idle_ttl" 
        CHECK ("idle_ttl_ms" >= 0),
    CONSTRAINT "check_max_restarts" 
        CHECK ("max_restarts" >= 0 AND "max_restarts" <= 10)
);

-- 添加字段注释
COMMENT ON TABLE "public"."mcp_service_stdio" IS 'STDIO MCP服务配置表';
COMMENT ON COLUMN "public"."mcp_service_stdio"."server_id" IS '服务器ID，关联mcp_service表';
COMMENT ON COLUMN "public"."mcp_service_stdio"."command" IS '启动命令';
COMMENT ON COLUMN "public"."mcp_service_stdio"."args" IS '命令参数数组';
COMMENT ON COLUMN "public"."mcp_service_stdio"."workdir" IS '工作目录';
COMMENT ON COLUMN "public"."mcp_service_stdio"."env" IS '环境变量';
COMMENT ON COLUMN "public"."mcp_service_stdio"."startup_timeout_ms" IS '启动超时时间（毫秒）';
COMMENT ON COLUMN "public"."mcp_service_stdio"."shutdown_timeout_ms" IS '关闭超时时间（毫秒）';
COMMENT ON COLUMN "public"."mcp_service_stdio"."reuse_strategy" IS '复用策略: shared(共享进程), per_user(每用户独立), per_session(每会话独立)';
COMMENT ON COLUMN "public"."mcp_service_stdio"."max_concurrent" IS '最大并发连接数，0表示无限制';
COMMENT ON COLUMN "public"."mcp_service_stdio"."idle_ttl_ms" IS '空闲超时时间（毫秒），0表示不超时';
COMMENT ON COLUMN "public"."mcp_service_stdio"."max_restarts" IS '最大重启次数';
COMMENT ON COLUMN "public"."mcp_service_stdio"."init_params" IS '初始化参数';

-- 创建索引
CREATE INDEX IF NOT EXISTS "idx_stdio_reuse_strategy" ON "public"."mcp_service_stdio"("reuse_strategy");
CREATE INDEX IF NOT EXISTS "idx_stdio_updated_at" ON "public"."mcp_service_stdio"("updated_at");

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_stdio_modtime ON "public"."mcp_service_stdio";
CREATE TRIGGER update_stdio_modtime 
    BEFORE UPDATE ON "public"."mcp_service_stdio" 
    FOR EACH ROW EXECUTE FUNCTION update_modified_column();
