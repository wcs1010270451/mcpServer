package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"McpServer/internal/models"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// DatabaseService 数据库服务
type DatabaseService struct {
	db *sql.DB
}

// DatabaseConfig 数据库配置接口
type DatabaseConfig interface {
	GetDSN() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() time.Duration
}

// NewDatabaseService 创建新的数据库服务
func NewDatabaseService(config DatabaseConfig) (*DatabaseService, error) {
	// 使用配置中的连接字符串
	connStr := config.GetDSN()

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(config.GetMaxOpenConns())
	db.SetMaxIdleConns(config.GetMaxIdleConns())
	db.SetConnMaxLifetime(config.GetConnMaxLifetime())

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database")

	return &DatabaseService{db: db}, nil
}

// Close 关闭数据库连接
func (ds *DatabaseService) Close() error {
	return ds.db.Close()
}

// GetEnabledServices 获取所有启用的 MCP 服务 (排除远程服务)
func (ds *DatabaseService) GetEnabledServices() ([]models.MCPService, error) {
	query := `
		SELECT server_id, display_name, implementation_name, protocol_version, 
		       enabled, metadata, created_at, updated_at, adapter, start_mode
		FROM mcp_service 
		WHERE enabled = true AND adapter = 'builtin'
		ORDER BY server_id
	`

	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []models.MCPService
	for rows.Next() {
		var service models.MCPService
		err := rows.Scan(
			&service.ServerID,
			&service.DisplayName,
			&service.ImplementationName,
			&service.ProtocolVersion,
			&service.Enabled,
			&service.Metadata,
			&service.CreatedAt,
			&service.UpdatedAt,
			&service.Adapter,
			&service.StartMode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, service)
	}

	return services, nil
}

// GetServiceByID 根据 server_id 获取服务
func (ds *DatabaseService) GetServiceByID(serverID string) (*models.MCPService, error) {
	query := `
		SELECT server_id, display_name, implementation_name, protocol_version, 
		       enabled, metadata, created_at, updated_at, adapter, start_mode
		FROM mcp_service 
		WHERE server_id = $1 AND enabled = true
	`

	var service models.MCPService
	err := ds.db.QueryRow(query, serverID).Scan(
		&service.ServerID,
		&service.DisplayName,
		&service.ImplementationName,
		&service.ProtocolVersion,
		&service.Enabled,
		&service.Metadata,
		&service.CreatedAt,
		&service.UpdatedAt,
		&service.Adapter,
		&service.StartMode,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service with id %s not found", serverID)
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// GetToolsByServerID 根据 server_id 获取该服务的所有工具
func (ds *DatabaseService) GetToolsByServerID(serverID string) ([]models.MCPTool, error) {
	query := `
		SELECT id, server_id, name, description, args_schema, enabled, 
		       created_at, updated_at, handler_type, handler_config
		FROM mcp_tool 
		WHERE server_id = $1 AND enabled = true
		ORDER BY name
	`

	rows, err := ds.db.Query(query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tools: %w", err)
	}
	defer rows.Close()

	var tools []models.MCPTool
	for rows.Next() {
		var tool models.MCPTool
		err := rows.Scan(
			&tool.ID,
			&tool.ServerID,
			&tool.Name,
			&tool.Description,
			&tool.ArgsSchema,
			&tool.Enabled,
			&tool.CreatedAt,
			&tool.UpdatedAt,
			&tool.HandlerType,
			&tool.HandlerConfig,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// GetServiceWithTools 获取服务及其所有工具
func (ds *DatabaseService) GetServiceWithTools(serverID string) (*models.ServiceWithTools, error) {
	service, err := ds.GetServiceByID(serverID)
	if err != nil {
		return nil, err
	}

	tools, err := ds.GetToolsByServerID(serverID)
	if err != nil {
		return nil, err
	}

	return &models.ServiceWithTools{
		Service: *service,
		Tools:   tools,
	}, nil
}

// GetStdioServiceConfig 获取远程 stdio 服务配置
func (ds *DatabaseService) GetStdioServiceConfig(serverID string) (*models.MCPServiceStdio, error) {
	query := `
		SELECT server_id, command, args, workdir, env, startup_timeout_ms, 
		       shutdown_timeout_ms, reuse_strategy, idle_ttl_ms, max_restarts, init_params
		FROM mcp_service_stdio 
		WHERE server_id = $1
	`

	var config models.MCPServiceStdio
	var args []string
	err := ds.db.QueryRow(query, serverID).Scan(
		&config.ServerID,
		&config.Command,
		pq.Array(&args),
		&config.Workdir,
		&config.Env,
		&config.StartupTimeoutMs,
		&config.ShutdownTimeoutMs,
		&config.ReuseStrategy,
		&config.IdleTtlMs,
		&config.MaxRestarts,
		&config.InitParams,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stdio config for server %s not found", serverID)
		}
		return nil, fmt.Errorf("failed to get stdio config: %w", err)
	}

	config.Args = args
	return &config, nil
}

// GetSSEServiceConfig 获取远程 SSE 服务配置
func (ds *DatabaseService) GetSSEServiceConfig(serverID string) (*models.MCPServiceSSE, error) {
	query := `
		SELECT server_id, base_url, sse_path, auth_type, auth_config, 
		       timeout_ms, connect_timeout_ms, retry_attempts, retry_delay_ms,
		       health_check_enabled, health_check_path, health_check_interval_ms,
		       headers, query_params, connection_pool_size, keep_alive,
		       follow_redirects, max_redirects, user_agent, created_at, updated_at
		FROM mcp_service_sse 
		WHERE server_id = $1
	`

	var config models.MCPServiceSSE
	err := ds.db.QueryRow(query, serverID).Scan(
		&config.ServerID,
		&config.BaseURL,
		&config.SSEPath,
		&config.AuthType,
		&config.AuthConfig,
		&config.TimeoutMs,
		&config.ConnectTimeoutMs,
		&config.RetryAttempts,
		&config.RetryDelayMs,
		&config.HealthCheckEnabled,
		&config.HealthCheckPath,
		&config.HealthCheckIntervalMs,
		&config.Headers,
		&config.QueryParams,
		&config.ConnectionPoolSize,
		&config.KeepAlive,
		&config.FollowRedirects,
		&config.MaxRedirects,
		&config.UserAgent,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sse config for server %s not found", serverID)
		}
		return nil, fmt.Errorf("failed to get sse config: %w", err)
	}

	return &config, nil
}

// IsRemoteStdioService 检查服务是否为远程 stdio 服务
func (ds *DatabaseService) IsRemoteStdioService(serverID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM mcp_service 
			WHERE server_id = $1 AND enabled = true AND adapter = 'remote_stdio'
		)
	`

	var exists bool
	err := ds.db.QueryRow(query, serverID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check remote stdio service: %w", err)
	}

	return exists, nil
}

// IsRemoteSSEService 检查服务是否为远程 SSE 服务
func (ds *DatabaseService) IsRemoteSSEService(serverID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM mcp_service 
			WHERE server_id = $1 AND enabled = true AND adapter = 'remote_sse'
		)
	`

	var exists bool
	err := ds.db.QueryRow(query, serverID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check remote sse service: %w", err)
	}

	return exists, nil
}

// GetEmployeeByName 根据姓名查询员工信息
func (ds *DatabaseService) GetEmployeeByName(name string) (*models.Employee, error) {
	query := `
		SELECT id, name, address, phone, enabled
		FROM employees 
		WHERE name = $1 AND enabled = true
	`

	var employee models.Employee
	err := ds.db.QueryRow(query, name).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Address,
		&employee.Phone,
		&employee.Enabled,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 未找到员工
		}
		return nil, fmt.Errorf("failed to query employee by name: %v", err)
	}

	return &employee, nil
}

// GetAllEmployees 获取所有启用的员工列表
func (ds *DatabaseService) GetAllEmployees() ([]models.Employee, error) {
	query := `
		SELECT id, name, address, phone, enabled
		FROM employees 
		WHERE enabled = true
		ORDER BY name
	`

	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %v", err)
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var employee models.Employee
		err := rows.Scan(
			&employee.ID,
			&employee.Name,
			&employee.Address,
			&employee.Phone,
			&employee.Enabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee row: %v", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating employee rows: %v", err)
	}

	return employees, nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
