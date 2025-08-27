-- 创建用户表
CREATE TABLE "public"."users" (
    "user_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "username" varchar(50) NOT NULL UNIQUE,                    -- 登录账号
    "password_hash" varchar(255) NOT NULL,                     -- 密码哈希值
    "name" varchar(100) NOT NULL,                              -- 用户姓名
    "phone" varchar(20) UNIQUE,                                -- 手机号
    "email" varchar(255),                                      -- 邮箱（可选）
    "avatar_url" text,                                          -- 头像URL
    "status" varchar(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')), -- 账号状态
    "last_login_at" timestamptz,                               -- 最后登录时间
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP,        -- 创建时间
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP         -- 更新时间
);

-- 设置表所有者
ALTER TABLE "public"."users" OWNER TO "wcs";

-- 创建必要索引
CREATE INDEX "idx_users_username" ON "public"."users" ("username");
CREATE INDEX "idx_users_phone" ON "public"."users" ("phone");
CREATE INDEX "idx_users_status" ON "public"."users" ("status");

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON "public"."users" 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 字段注释
COMMENT ON TABLE "public"."users" IS 'C端用户表';
COMMENT ON COLUMN "public"."users"."user_id" IS '用户唯一标识符';
COMMENT ON COLUMN "public"."users"."username" IS '登录用户名';
COMMENT ON COLUMN "public"."users"."password_hash" IS '密码哈希值';
COMMENT ON COLUMN "public"."users"."name" IS '用户姓名';
COMMENT ON COLUMN "public"."users"."phone" IS '手机号码';
COMMENT ON COLUMN "public"."users"."email" IS '邮箱地址';
COMMENT ON COLUMN "public"."users"."avatar_url" IS '用户头像URL';
COMMENT ON COLUMN "public"."users"."status" IS '账号状态：active-正常, inactive-未激活';
COMMENT ON COLUMN "public"."users"."last_login_at" IS '最后登录时间';
COMMENT ON COLUMN "public"."users"."created_at" IS '创建时间';
COMMENT ON COLUMN "public"."users"."updated_at" IS '更新时间';

-- 创建用户密钥表
CREATE TABLE "public"."user_keys" (
    "key_id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,                                   -- 关联用户ID
    "key_name" varchar(100) NOT NULL,                          -- 密钥名称
    "key_value" varchar(255) NOT NULL UNIQUE,                  -- 密钥值
    "key_type" varchar(20) DEFAULT 'api' CHECK (key_type IN ('api', 'access_token', 'refresh_token', 'secret')),
    "status" varchar(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'revoked')),
    "permissions" jsonb DEFAULT '[]'::jsonb,                   -- 密钥权限范围
    "expires_at" timestamptz,                                   -- 过期时间（可选）
    "last_used_at" timestamptz,                                -- 最后使用时间
    "usage_count" int4 DEFAULT 0,                              -- 使用次数
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP,        -- 创建时间
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP,        -- 更新时间
    
    -- 外键约束
    CONSTRAINT "fk_user_keys_user_id" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("user_id") ON DELETE CASCADE
);

-- 设置表所有者
ALTER TABLE "public"."user_keys" OWNER TO "wcs";

-- 创建索引
CREATE INDEX "idx_user_keys_user_id" ON "public"."user_keys" ("user_id");
CREATE INDEX "idx_user_keys_key_value" ON "public"."user_keys" ("key_value");
CREATE INDEX "idx_user_keys_status" ON "public"."user_keys" ("status");
CREATE INDEX "idx_user_keys_key_type" ON "public"."user_keys" ("key_type");
CREATE INDEX "idx_user_keys_expires_at" ON "public"."user_keys" ("expires_at");

-- 创建更新时间触发器
CREATE TRIGGER update_user_keys_updated_at 
    BEFORE UPDATE ON "public"."user_keys" 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 字段注释
COMMENT ON TABLE "public"."user_keys" IS '用户密钥表';
COMMENT ON COLUMN "public"."user_keys"."key_id" IS '密钥唯一标识符';
COMMENT ON COLUMN "public"."user_keys"."user_id" IS '关联的用户ID';
COMMENT ON COLUMN "public"."user_keys"."key_name" IS '密钥名称，用户自定义';
COMMENT ON COLUMN "public"."user_keys"."key_value" IS '密钥值，全局唯一';
COMMENT ON COLUMN "public"."user_keys"."key_type" IS '密钥类型：api-API密钥, access_token-访问令牌, refresh_token-刷新令牌, secret-密钥';
COMMENT ON COLUMN "public"."user_keys"."status" IS '密钥状态：active-有效, inactive-禁用, revoked-已撤销';
COMMENT ON COLUMN "public"."user_keys"."permissions" IS '密钥权限范围JSON数组';
COMMENT ON COLUMN "public"."user_keys"."expires_at" IS '密钥过期时间';
COMMENT ON COLUMN "public"."user_keys"."last_used_at" IS '最后使用时间';
COMMENT ON COLUMN "public"."user_keys"."usage_count" IS '密钥使用次数统计';
COMMENT ON COLUMN "public"."user_keys"."created_at" IS '创建时间';
COMMENT ON COLUMN "public"."user_keys"."updated_at" IS '更新时间';

-- 创建查询视图（可选，方便联合查询）
CREATE VIEW "public"."user_keys_view" AS
SELECT 
    uk.key_id,
    uk.user_id,
    u.username,
    u.name as user_name,
    uk.key_name,
    uk.key_value,
    uk.key_type,
    uk.status,
    uk.permissions,
    uk.expires_at,
    uk.last_used_at,
    uk.usage_count,
    uk.created_at,
    uk.updated_at,
    CASE 
        WHEN uk.expires_at IS NULL THEN false
        WHEN uk.expires_at > CURRENT_TIMESTAMP THEN false
        ELSE true
    END as is_expired
FROM user_keys uk
JOIN users u ON uk.user_id = u.user_id;

COMMENT ON VIEW "public"."user_keys_view" IS '用户密钥视图，包含用户信息和过期状态';
