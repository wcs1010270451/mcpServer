-- MCP SSE 远程服务配置表
CREATE TABLE "public"."mcp_service_sse" (
  "server_id" text COLLATE "pg_catalog"."default" NOT NULL,
  "base_url" text COLLATE "pg_catalog"."default" NOT NULL,
  "sse_path" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT '/sse'::text,
  "auth_type" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'none'::text,
  "auth_config" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "timeout_ms" int4 NOT NULL DEFAULT 30000,
  "connect_timeout_ms" int4 NOT NULL DEFAULT 10000,
  "retry_attempts" int4 NOT NULL DEFAULT 3,
  "retry_delay_ms" int4 NOT NULL DEFAULT 1000,
  "health_check_enabled" bool NOT NULL DEFAULT true,
  "health_check_path" text COLLATE "pg_catalog"."default" DEFAULT '/health'::text,
  "health_check_interval_ms" int4 NOT NULL DEFAULT 60000,
  "headers" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "query_params" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "connection_pool_size" int4 NOT NULL DEFAULT 5,
  "keep_alive" bool NOT NULL DEFAULT true,
  "follow_redirects" bool NOT NULL DEFAULT true,
  "max_redirects" int4 NOT NULL DEFAULT 5,
  "user_agent" text COLLATE "pg_catalog"."default" DEFAULT 'MCP-Proxy/1.0'::text,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  CONSTRAINT "mcp_service_sse_pkey" PRIMARY KEY ("server_id"),
  CONSTRAINT "mcp_service_sse_server_id_fkey" FOREIGN KEY ("server_id") REFERENCES "public"."mcp_service" ("server_id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "mcp_service_sse_auth_type_check" CHECK (auth_type IN ('none', 'bearer_token', 'api_key', 'basic_auth', 'custom_header'))
);

-- 设置表所有者
ALTER TABLE "public"."mcp_service_sse" 
  OWNER TO "wcs";

-- 创建索引
CREATE INDEX "idx_mcp_service_sse_auth_type" ON "public"."mcp_service_sse" USING btree (
  "auth_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

CREATE INDEX "idx_mcp_service_sse_health_check" ON "public"."mcp_service_sse" USING btree (
  "health_check_enabled" "pg_catalog"."bool_ops" ASC NULLS LAST
);

-- 字段注释
COMMENT ON TABLE "public"."mcp_service_sse" IS 'MCP SSE远程服务配置表';

COMMENT ON COLUMN "public"."mcp_service_sse"."server_id" IS '服务ID，关联mcp_service表';

COMMENT ON COLUMN "public"."mcp_service_sse"."base_url" IS '远程SSE服务的基础URL，如：http://remote-server:8080';

COMMENT ON COLUMN "public"."mcp_service_sse"."sse_path" IS 'SSE端点路径，通常为/sse或/mcp-server/sse';

COMMENT ON COLUMN "public"."mcp_service_sse"."auth_type" IS '认证类型：none-无认证, bearer_token-Bearer令牌, api_key-API密钥, basic_auth-基础认证, custom_header-自定义头部';

COMMENT ON COLUMN "public"."mcp_service_sse"."auth_config" IS '认证配置JSON，根据auth_type包含不同字段：
- bearer_token: {"token": "xxx"}
- api_key: {"key": "xxx", "header": "X-API-Key"}
- basic_auth: {"username": "xxx", "password": "xxx"}
- custom_header: {"header_name": "value"}';

COMMENT ON COLUMN "public"."mcp_service_sse"."timeout_ms" IS '请求超时时间（毫秒）';

COMMENT ON COLUMN "public"."mcp_service_sse"."connect_timeout_ms" IS '连接超时时间（毫秒）';

COMMENT ON COLUMN "public"."mcp_service_sse"."retry_attempts" IS '重试次数';

COMMENT ON COLUMN "public"."mcp_service_sse"."retry_delay_ms" IS '重试间隔时间（毫秒）';

COMMENT ON COLUMN "public"."mcp_service_sse"."health_check_enabled" IS '是否启用健康检查';

COMMENT ON COLUMN "public"."mcp_service_sse"."health_check_path" IS '健康检查路径';

COMMENT ON COLUMN "public"."mcp_service_sse"."health_check_interval_ms" IS '健康检查间隔时间（毫秒）';

COMMENT ON COLUMN "public"."mcp_service_sse"."headers" IS '自定义HTTP请求头，JSON格式：{"Content-Type": "application/json"}';

COMMENT ON COLUMN "public"."mcp_service_sse"."query_params" IS '默认查询参数，JSON格式：{"param1": "value1"}';

COMMENT ON COLUMN "public"."mcp_service_sse"."connection_pool_size" IS '连接池大小';

COMMENT ON COLUMN "public"."mcp_service_sse"."keep_alive" IS '是否保持长连接';

COMMENT ON COLUMN "public"."mcp_service_sse"."follow_redirects" IS '是否跟随重定向';

COMMENT ON COLUMN "public"."mcp_service_sse"."max_redirects" IS '最大重定向次数';

COMMENT ON COLUMN "public"."mcp_service_sse"."user_agent" IS 'HTTP User-Agent头';

COMMENT ON COLUMN "public"."mcp_service_sse"."created_at" IS '创建时间';

COMMENT ON COLUMN "public"."mcp_service_sse"."updated_at" IS '更新时间';

-- 示例数据
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
    'remote_sse_demo', 
    'Remote SSE Demo Service', 
    'remote_sse_demo', 
    '2024-11-05', 
    true, 
    '{"description": "Remote SSE service demonstration"}', 
    'remote_sse', 
    'on_demand'
);

INSERT INTO "public"."mcp_service_sse" (
    "server_id",
    "base_url",
    "sse_path",
    "auth_type",
    "auth_config",
    "timeout_ms",
    "headers",
    "query_params"
) VALUES (
    'remote_sse_demo',
    'http://remote-mcp-server:8080',
    '/mcp-server/sse',
    'bearer_token',
    '{"token": "your-bearer-token-here"}',
    30000,
    '{"Content-Type": "text/event-stream", "Cache-Control": "no-cache"}',
    '{"version": "1.0"}'
);

-- 查询验证
SELECT 
    s.server_id,
    s.display_name,
    s.adapter,
    sse.base_url,
    sse.sse_path,
    sse.auth_type
FROM mcp_service s
JOIN mcp_service_sse sse ON s.server_id = sse.server_id
WHERE s.enabled = true AND s.adapter = 'remote_sse';
