-- 示例：带参数的 MCP 服务和工具
-- 插入示例服务
INSERT INTO "public"."mcp_service" 
("server_id", "display_name", "implementation_name", "protocol_version", "enabled", "metadata", "adapter", "start_mode") 
VALUES 
('demo-server', 'Demo Server', 'demo-server', '2025-03-26', true, '{"description": "Demonstration server with parameterized tools"}', 'builtin', 'auto');

-- 插入 Echo 工具（带参数）
INSERT INTO "public"."mcp_tool" 
("server_id", "tool_name", "description", "args_schema", "handler_type", "handler_config", "enabled") 
VALUES 
('demo-server', 'echo', 'Echo back the provided text with optional prefix', 
'{
  "type": "object",
  "title": "Echo Tool Parameters",
  "description": "Parameters for the echo tool",
  "properties": {
    "text": {
      "type": "string",
      "description": "The text to echo back"
    },
    "prefix": {
      "type": "string", 
      "description": "Optional prefix to add before the text"
    }
  },
  "required": ["text"]
}', 'builtin_echo', '{}', true);

-- 插入 Say Hi 工具（带参数）
INSERT INTO "public"."mcp_tool" 
("server_id", "tool_name", "description", "args_schema", "handler_type", "handler_config", "enabled") 
VALUES 
('demo-server', 'say_hi', 'Say hi to someone with their name', 
'{
  "type": "object",
  "title": "Say Hi Tool Parameters", 
  "description": "Parameters for the say hi tool",
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the person to greet"
    }
  },
  "required": []
}', 'builtin_say_hi', '{}', true);

-- 插入 Greet 工具（带参数）
INSERT INTO "public"."mcp_tool" 
("server_id", "tool_name", "description", "args_schema", "handler_type", "handler_config", "enabled") 
VALUES 
('demo-server', 'greet', 'Greet someone with custom greeting and name', 
'{
  "type": "object",
  "title": "Greet Tool Parameters",
  "description": "Parameters for the greet tool", 
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the person to greet"
    },
    "greeting": {
      "type": "string",
      "description": "The greeting to use (e.g., Hello, Hi, Good morning)"
    }
  },
  "required": []
}', 'builtin_greet', '{}', true);

-- 插入 Status 工具（无参数）
INSERT INTO "public"."mcp_tool" 
("server_id", "tool_name", "description", "args_schema", "handler_type", "handler_config", "enabled") 
VALUES 
('demo-server', 'status', 'Get system status', 
'{
  "type": "object",
  "title": "Status Tool Parameters",
  "description": "No parameters required for status check",
  "properties": {},
  "required": []
}', 'builtin_status', '{}', true);
