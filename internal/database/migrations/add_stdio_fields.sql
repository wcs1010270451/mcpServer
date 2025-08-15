-- 如果 mcp_service_stdio 表已存在，添加缺失的字段

-- 添加 reuse_strategy 字段（如果不存在）
DO $$ 
BEGIN 
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'mcp_service_stdio' AND column_name = 'reuse_strategy'
    ) THEN
        ALTER TABLE "public"."mcp_service_stdio" 
        ADD COLUMN "reuse_strategy" varchar(20) NOT NULL DEFAULT 'shared';
        
        ALTER TABLE "public"."mcp_service_stdio" 
        ADD CONSTRAINT "check_reuse_strategy" 
        CHECK ("reuse_strategy" IN ('shared', 'per_user', 'per_session'));
    END IF;
END $$;

-- 添加 max_concurrent 字段（如果不存在）
DO $$ 
BEGIN 
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'mcp_service_stdio' AND column_name = 'max_concurrent'
    ) THEN
        ALTER TABLE "public"."mcp_service_stdio" 
        ADD COLUMN "max_concurrent" int NOT NULL DEFAULT 10;
        
        ALTER TABLE "public"."mcp_service_stdio" 
        ADD CONSTRAINT "check_max_concurrent" 
        CHECK ("max_concurrent" >= 0 AND "max_concurrent" <= 1000);
    END IF;
END $$;

-- 添加字段注释
COMMENT ON COLUMN "public"."mcp_service_stdio"."reuse_strategy" IS '复用策略: shared(共享进程), per_user(每用户独立), per_session(每会话独立)';
COMMENT ON COLUMN "public"."mcp_service_stdio"."max_concurrent" IS '最大并发连接数，0表示无限制';

-- 为现有记录设置合理的默认值
UPDATE "public"."mcp_service_stdio" 
SET 
    "reuse_strategy" = 'shared',
    "max_concurrent" = 10
WHERE "reuse_strategy" IS NULL OR "reuse_strategy" = '';

-- 验证表结构
SELECT 
    column_name, 
    data_type, 
    is_nullable, 
    column_default
FROM information_schema.columns 
WHERE table_name = 'mcp_service_stdio' 
    AND column_name IN ('reuse_strategy', 'max_concurrent')
ORDER BY column_name;
