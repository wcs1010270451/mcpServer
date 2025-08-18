package manager

import (
	"McpServer/internal/logger"
	"McpServer/internal/models"
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerManager 管理所有 MCP 服务器实例
type MCPServerManager struct {
	db              DatabaseServiceInterface
	handlerRegistry HandlerRegistryInterface
	servers         map[string]*mcp.Server
	remoteManager   *RemoteStdioManager
	sseManager      *RemoteSSEManager
}

// NewMCPServerManager 创建新的服务器管理器
func NewMCPServerManager(db DatabaseServiceInterface, handlerRegistry HandlerRegistryInterface) *MCPServerManager {
	return &MCPServerManager{
		db:              db,
		handlerRegistry: handlerRegistry,
		servers:         make(map[string]*mcp.Server),
		remoteManager:   NewRemoteStdioManager(db),
		sseManager:      NewRemoteSSEManager(db),
	}
}

// LoadServersFromDatabase 从数据库加载所有服务器配置
func (m *MCPServerManager) LoadServersFromDatabase() error {
	services, err := m.db.GetEnabledServices()
	if err != nil {
		return fmt.Errorf("failed to load builtin services: %w", err)
	}

	logger.Info("Loading %d builtin services from database", len(services))

	for _, service := range services {
		logger.Info("Loading builtin service: %s", service.ServerID)

		server, err1 := m.createServerFromConfig(service)
		if err1 != nil {
			logger.Info("Failed to create server for builtin service %s: %v", service.ServerID, err1)
			continue
		}

		m.servers[service.ServerID] = server
		logger.Info("Successfully loaded builtin service: %s", service.ServerID)
	}

	return nil
}

// createServerFromConfig 根据配置创建服务器
func (m *MCPServerManager) createServerFromConfig(service models.MCPService) (*mcp.Server, error) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-builtin-server",
		Version: "1.0.0",
	}, nil)

	// 获取工具配置
	tools, err := m.db.GetToolsByServerID(service.ServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools for service %s: %w", service.ServerID, err)
	}

	// 添加工具
	for _, tool := range tools {
		logger.Info("Adding tool: %s to server: %s", tool.Name, service.ServerID)

		// 获取处理器
		handler, exists := m.handlerRegistry.GetHandler(tool.HandlerType)
		if !exists {
			logger.Warn("No handler found for type: %s", tool.HandlerType)
			continue
		}

		// 创建工具定义
		toolDef := mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}

		// 转换 JSONB 到 JSON Schema
		if tool.ArgsSchema != nil {
			schema, err1 := tool.ArgsSchema.ToJSONSchema()
			if err1 != nil {
				logger.Warn("Failed to convert args schema for tool %s: %v", tool.Name, err1)
			} else if schema != nil {
				toolDef.InputSchema = schema
			}
		}

		// 创建符合 ToolHandlerFor 类型的处理器
		toolHandler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
			// 转换参数类型
			callParams := &mcp.CallToolParams{
				Name:      params.Name,
				Arguments: params.Arguments,
			}
			result, err1 := handler(ctx, session, callParams)
			if err1 != nil {
				return nil, err1
			}
			// 转换返回类型
			return &mcp.CallToolResultFor[any]{
				Content: result.Content,
				IsError: result.IsError,
			}, nil
		}

		mcp.AddTool(server, &toolDef, toolHandler)
	}

	return server, nil
}

// GetServer 根据 server_id 获取对应的 MCP 服务器
func (m *MCPServerManager) GetServer(serverID string) (*mcp.Server, error) {
	// 检查是否是远程 stdio 服务
	isRemoteStdio, err := m.db.IsRemoteStdioService(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if service is remote stdio: %w", err)
	}
	if isRemoteStdio {
		logger.Info("Getting remote stdio server for: %s", serverID)
		return m.remoteManager.GetOrCreateRemoteServer(serverID)
	}

	// 检查是否是远程 SSE 服务
	isRemoteSSE, err := m.db.IsRemoteSSEService(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if service is remote SSE: %w", err)
	}
	if isRemoteSSE {
		logger.Info("Getting remote SSE server for: %s", serverID)
		return m.sseManager.GetOrCreateRemoteServer(serverID)
	}

	// 本地服务
	if server, exists := m.servers[serverID]; exists {
		logger.Info("Using builtin MCP server for: %s", serverID)
		return server, nil
	}

	return nil, fmt.Errorf("server not found: %s", serverID)
}

// GetDB 获取数据库服务接口
func (m *MCPServerManager) GetDB() DatabaseServiceInterface {
	return m.db
}
