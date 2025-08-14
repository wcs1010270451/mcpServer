-- 插入示例 MCP 服务配置
INSERT INTO mcp_service (
    server_id, 
    display_name, 
    implementation_name, 
    protocol_version, 
    enabled, 
    metadata, 
    adapter, 
    start_mode
) VALUES 
(
    'greeter1', 
    'Greeter Service 1', 
    'greeter1', 
    '2024-11-05', 
    true, 
    '{"description": "First greeting service", "version": "1.0"}', 
    'builtin', 
    'on_demand'
),
(
    'greeter2', 
    'Greeter Service 2', 
    'greeter2', 
    '2024-11-05', 
    true, 
    '{"description": "Second greeting service", "version": "1.0"}', 
    'builtin', 
    'on_demand'
),
(
    'greeter3', 
    'Greeter Service 3', 
    'greeter3', 
    '2024-11-05', 
    true, 
    '{"description": "Third greeting service", "version": "1.0"}', 
    'builtin', 
    'on_demand'
),
(
    'echo_service', 
    'Echo Service', 
    'echo_service', 
    '2024-11-05', 
    true, 
    '{"description": "Echo service for testing", "version": "1.0"}', 
    'builtin', 
    'on_demand'
),
(
    'status_service', 
    'Status Service', 
    'status_service', 
    '2024-11-05', 
    true, 
    '{"description": "System status service", "version": "1.0"}', 
    'builtin', 
    'on_demand'
);

-- 插入示例工具配置
INSERT INTO mcp_tool (
    server_id, 
    name, 
    description, 
    args_schema, 
    enabled, 
    handler_type, 
    handler_config
) VALUES 
-- greeter1 服务的工具
(
    'greeter1', 
    'greet1', 
    'Say Hi to someone', 
    '{"type": "object", "properties": {"name": {"type": "string", "description": "Name to greet"}}, "required": ["name"]}', 
    true, 
    'say_hi', 
    '{}'
),

-- greeter2 服务的工具
(
    'greeter2', 
    'greet2', 
    'Say Hello to someone', 
    '{"type": "object", "properties": {"name": {"type": "string", "description": "Name to greet"}}, "required": ["name"]}', 
    true, 
    'say_hello', 
    '{}'
),

-- greeter3 服务的工具
(
    'greeter3', 
    'greet3', 
    'Say NotFond to someone', 
    '{"type": "object", "properties": {"name": {"type": "string", "description": "Name to greet"}}, "required": ["name"]}', 
    true, 
    'say_notfond', 
    '{}'
),

-- echo_service 的工具
(
    'echo_service', 
    'echo', 
    'Echo back the input message', 
    '{"type": "object", "properties": {"message": {"type": "string", "description": "Message to echo"}, "repeat": {"type": "integer", "description": "Number of times to repeat", "default": 1}}}', 
    true, 
    'builtin_echo', 
    '{}'
),

-- status_service 的工具
(
    'status_service', 
    'get_status', 
    'Get system status', 
    '{"type": "object", "properties": {}}', 
    true, 
    'builtin_status', 
    '{}'
),

-- 通用问候工具
(
    'greeter1', 
    'custom_greet', 
    'Custom greeting with configurable message', 
    '{"type": "object", "properties": {"name": {"type": "string", "description": "Name to greet"}, "greeting": {"type": "string", "description": "Custom greeting word", "default": "Hello"}}, "required": ["name"]}', 
    true, 
    'builtin_greet', 
    '{"default_greeting": "Hi there"}'
),

(
    'greeter2', 
    'custom_greet', 
    'Custom greeting with configurable message', 
    '{"type": "object", "properties": {"name": {"type": "string", "description": "Name to greet"}, "greeting": {"type": "string", "description": "Custom greeting word", "default": "Hello"}}, "required": ["name"]}', 
    true, 
    'builtin_greet', 
    '{"default_greeting": "Welcome"}'
);

-- 查询验证数据
SELECT 
    s.server_id,
    s.display_name,
    s.implementation_name,
    s.enabled as service_enabled,
    COUNT(t.id) as tool_count
FROM mcp_service s
LEFT JOIN mcp_tool t ON s.server_id = t.server_id AND t.enabled = true
WHERE s.enabled = true
GROUP BY s.server_id, s.display_name, s.implementation_name, s.enabled
ORDER BY s.server_id;

-- 查询工具详情
SELECT 
    t.server_id,
    t.name,
    t.description,
    t.handler_type,
    t.enabled
FROM mcp_tool t
JOIN mcp_service s ON t.server_id = s.server_id
WHERE s.enabled = true AND t.enabled = true
ORDER BY t.server_id, t.name;
