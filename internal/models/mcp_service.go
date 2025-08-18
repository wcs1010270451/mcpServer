package models

import "time"

// MCPService 表示 mcp_service 表的数据模型
type MCPService struct {
	ServerID           string    `json:"server_id" db:"server_id"`
	DisplayName        string    `json:"display_name" db:"display_name"`
	ImplementationName string    `json:"implementation_name" db:"implementation_name"`
	ProtocolVersion    string    `json:"protocol_version" db:"protocol_version"`
	Enabled            bool      `json:"enabled" db:"enabled"`
	Metadata           JSONB     `json:"metadata" db:"metadata"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	Adapter            string    `json:"adapter" db:"adapter"`
	StartMode          string    `json:"start_mode" db:"start_mode"`
}

// ServiceWithTools 包含服务及其工具的完整信息
type ServiceWithTools struct {
	Service MCPService `json:"service"`
	Tools   []MCPTool  `json:"tools"`
}
