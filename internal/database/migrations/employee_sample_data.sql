-- 员工示例数据和工具配置

-- 插入示例员工数据
INSERT INTO "public"."employees" ("name", "address", "phone", "enabled") 
VALUES 
('张三', '北京市朝阳区建国路88号', '13800138001', true),
('李四', '上海市浦东新区陆家嘴环路1000号', '13800138002', true),
('王五', '深圳市南山区科技园南区R2-B栋', '13800138003', true),
('赵六', '广州市天河区珠江新城花城大道85号', '13800138004', true),
('钱七', '杭州市西湖区文三路508号', '13800138005', true)
ON CONFLICT (name) DO NOTHING;

-- 如果 demo-server 不存在，先创建
INSERT INTO "public"."mcp_service" 
("server_id", "display_name", "implementation_name", "protocol_version", "enabled", "metadata", "adapter", "start_mode") 
VALUES 
('server_employee_info', '查询员工信息', 'server_employee_infor', '2025-03-26', true, '{"description": "Demonstration server with parameterized tools"}', 'builtin', 'auto')
ON CONFLICT (server_id) DO NOTHING;

-- 插入员工查询工具
INSERT INTO "public"."mcp_tool" 
("server_id", "name", "description", "args_schema", "handler_type", "handler_config", "enabled") 
VALUES 
('server_employee_info', 'employee_query', '根据员工姓名查询员工地址和电话', 
'{
  "type": "object",
  "title": "Employee Query Tool Parameters",
  "description": "Parameters for querying employee information",
  "properties": {
    "name": {
      "type": "string",
      "description": "员工姓名"
    }
  },
  "required": ["name"]
}', 'builtin_employee_query', '{}', true),

('server_employee_info', 'employee_address', '根据员工姓名查询员工地址', 
'{
  "type": "object",
  "title": "Employee Address Query Parameters",
  "description": "Parameters for querying employee address",
  "properties": {
    "name": {
      "type": "string",
      "description": "员工姓名"
    }
  },
  "required": ["name"]
}', 'builtin_employee_address', '{}', true),

('server_employee_info', 'employee_phone', '根据员工姓名查询员工电话', 
'{
  "type": "object",
  "title": "Employee Phone Query Parameters",
  "description": "Parameters for querying employee phone",
  "properties": {
    "name": {
      "type": "string",
      "description": "员工姓名"
    }
  },
  "required": ["name"]
}', 'builtin_employee_phone', '{}', true)

ON CONFLICT (server_id, name) DO NOTHING;
