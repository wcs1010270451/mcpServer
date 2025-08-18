package models

import "time"

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
