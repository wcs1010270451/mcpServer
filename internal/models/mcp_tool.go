package models

import "time"

// MCPTool 表示 mcp_tool 表的数据模型
type MCPTool struct {
	ID            int64     `json:"id" db:"id"`
	ServerID      string    `json:"server_id" db:"server_id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	ArgsSchema    JSONB     `json:"args_schema" db:"args_schema"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	HandlerType   string    `json:"handler_type" db:"handler_type"`
	HandlerConfig JSONB     `json:"handler_config" db:"handler_config"`
}
