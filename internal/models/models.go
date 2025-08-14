package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

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

// JSONB 类型用于处理 PostgreSQL 的 JSONB 字段
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// ServiceWithTools 包含服务及其工具的完整信息
type ServiceWithTools struct {
	Service MCPService `json:"service"`
	Tools   []MCPTool  `json:"tools"`
}

// MCPServiceStdio 表示 mcp_service_stdio 表的数据模型
type MCPServiceStdio struct {
	ServerID          string   `json:"server_id" db:"server_id"`
	Command           string   `json:"command" db:"command"`
	Args              []string `json:"args" db:"args"`
	Workdir           *string  `json:"workdir" db:"workdir"`
	Env               JSONB    `json:"env" db:"env"`
	StartupTimeoutMs  int      `json:"startup_timeout_ms" db:"startup_timeout_ms"`
	ShutdownTimeoutMs int      `json:"shutdown_timeout_ms" db:"shutdown_timeout_ms"`
	ReuseStrategy     string   `json:"reuse_strategy" db:"reuse_strategy"`
	IdleTtlMs         int      `json:"idle_ttl_ms" db:"idle_ttl_ms"`
	MaxRestarts       int      `json:"max_restarts" db:"max_restarts"`
	InitParams        JSONB    `json:"init_params" db:"init_params"`
}

// MCPServiceSSE 表示 mcp_service_sse 表的数据模型
type MCPServiceSSE struct {
	ServerID              string    `json:"server_id" db:"server_id"`
	BaseURL               string    `json:"base_url" db:"base_url"`
	SSEPath               string    `json:"sse_path" db:"sse_path"`
	AuthType              string    `json:"auth_type" db:"auth_type"`
	AuthConfig            JSONB     `json:"auth_config" db:"auth_config"`
	TimeoutMs             int       `json:"timeout_ms" db:"timeout_ms"`
	ConnectTimeoutMs      int       `json:"connect_timeout_ms" db:"connect_timeout_ms"`
	RetryAttempts         int       `json:"retry_attempts" db:"retry_attempts"`
	RetryDelayMs          int       `json:"retry_delay_ms" db:"retry_delay_ms"`
	HealthCheckEnabled    bool      `json:"health_check_enabled" db:"health_check_enabled"`
	HealthCheckPath       *string   `json:"health_check_path" db:"health_check_path"`
	HealthCheckIntervalMs int       `json:"health_check_interval_ms" db:"health_check_interval_ms"`
	Headers               JSONB     `json:"headers" db:"headers"`
	QueryParams           JSONB     `json:"query_params" db:"query_params"`
	ConnectionPoolSize    int       `json:"connection_pool_size" db:"connection_pool_size"`
	KeepAlive             bool      `json:"keep_alive" db:"keep_alive"`
	FollowRedirects       bool      `json:"follow_redirects" db:"follow_redirects"`
	MaxRedirects          int       `json:"max_redirects" db:"max_redirects"`
	UserAgent             string    `json:"user_agent" db:"user_agent"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// Employee 表示 employees 表的数据模型
type Employee struct {
	ID      int64  `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Address string `json:"address" db:"address"`
	Phone   string `json:"phone" db:"phone"`
	Enabled bool   `json:"enabled" db:"enabled"`
}

// JSONB 方法实现

// Value 实现 driver.Valuer 接口
//func (j JSONB) Value() (driver.Value, error) {
//	if j == nil {
//		return nil, nil
//	}
//	return json.Marshal(j)
//}

// Scan 实现 sql.Scanner 接口
//func (j *JSONB) Scan(value interface{}) error {
//	if value == nil {
//		*j = nil
//		return nil
//	}
//
//	switch v := value.(type) {
//	case []byte:
//		return json.Unmarshal(v, j)
//	case string:
//		return json.Unmarshal([]byte(v), j)
//	default:
//		return errors.New("cannot scan non-string value into JSONB")
//	}
//}

// ToJSONSchema 将 JSONB 转换为 JSON Schema
func (j *JSONB) ToJSONSchema() (*jsonschema.Schema, error) {
	if j == nil {
		return nil, nil
	}

	// 将 JSONB 转换为 map
	data := map[string]interface{}(*j)
	if data == nil {
		return nil, nil
	}

	// 创建 JSON Schema
	schema := &jsonschema.Schema{}

	// 设置基本属性
	if schemaType, ok := data["type"].(string); ok {
		schema.Type = schemaType
	}

	if title, ok := data["title"].(string); ok {
		schema.Title = title
	}

	if description, ok := data["description"].(string); ok {
		schema.Description = description
	}

	// 处理 properties
	if properties, ok := data["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*jsonschema.Schema)
		for propName, propValue := range properties {
			if propMap, ok := propValue.(map[string]interface{}); ok {
				propSchema := &jsonschema.Schema{}

				if propType, ok := propMap["type"].(string); ok {
					propSchema.Type = propType
				}

				if propDesc, ok := propMap["description"].(string); ok {
					propSchema.Description = propDesc
				}

				schema.Properties[propName] = propSchema
			}
		}
	}

	// 处理 required 字段
	if required, ok := data["required"].([]interface{}); ok {
		schema.Required = make([]string, len(required))
		for i, req := range required {
			if reqStr, ok := req.(string); ok {
				schema.Required[i] = reqStr
			}
		}
	}

	return schema, nil
}
