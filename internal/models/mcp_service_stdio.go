package models

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
	MaxConcurrent     int      `json:"max_concurrent" db:"max_concurrent"`
	IdleTtlMs         int      `json:"idle_ttl_ms" db:"idle_ttl_ms"`
	MaxRestarts       int      `json:"max_restarts" db:"max_restarts"`
	InitParams        JSONB    `json:"init_params" db:"init_params"`
}
