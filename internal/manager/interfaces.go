package manager

import (
	"McpServer/internal/handlers"
	"McpServer/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DatabaseServiceInterface 数据库服务接口
type DatabaseServiceInterface interface {
	GetEnabledServices() ([]models.MCPService, error)
	GetServiceByID(serverID string) (*models.MCPService, error)
	GetToolsByServerID(serverID string) ([]models.MCPTool, error)
	GetServiceWithTools(serverID string) (*models.ServiceWithTools, error)
	GetStdioServiceConfig(serverID string) (*models.MCPServiceStdio, error)
	GetSSEServiceConfig(serverID string) (*models.MCPServiceSSE, error)
	IsRemoteStdioService(serverID string) (bool, error)
	IsRemoteSSEService(serverID string) (bool, error)
	GetEmployeeByName(name string) (*models.Employee, error)
	GetAllEmployees() ([]models.Employee, error)
}

// HandlerRegistryInterface 处理器注册表接口
type HandlerRegistryInterface interface {
	GetHandler(handlerType string) (handlers.ToolHandler, bool)
}

// MCPServerManagerInterface MCP服务器管理器接口
type MCPServerManagerInterface interface {
	GetServer(serverID string) (*mcp.Server, error)
	GetDB() DatabaseServiceInterface
}
