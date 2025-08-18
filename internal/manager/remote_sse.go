package manager

import (
	"McpServer/internal/logger"
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"McpServer/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// headerRoundTripper 实现 http.RoundTripper 接口，用于添加自定义头部
type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (hrt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range hrt.headers {
		req.Header.Set(key, value)
	}
	return hrt.base.RoundTrip(req)
}

// SSESessionInfo SSE会话信息
type SSESessionInfo struct {
	session     *mcp.ClientSession
	client      *mcp.Client
	lastUsed    time.Time
	config      *models.MCPServiceSSE
	activeConns int32 // 活跃连接数
}

// RemoteSSEManager 管理远程 SSE MCP 服务
type RemoteSSEManager struct {
	db       DatabaseServiceInterface
	sessions map[string]*SSESessionInfo
	mutex    sync.RWMutex
}

// NewRemoteSSEManager 创建新的远程 SSE 管理器
func NewRemoteSSEManager(db DatabaseServiceInterface) *RemoteSSEManager {
	return &RemoteSSEManager{
		db:       db,
		sessions: make(map[string]*SSESessionInfo),
	}
}

// GetOrCreateRemoteServer 获取或创建远程 SSE 服务器连接
func (rsm *RemoteSSEManager) GetOrCreateRemoteServer(serverID string) (*mcp.Server, error) {
	rsm.mutex.RLock()
	if sessionInfo, exists := rsm.sessions[serverID]; exists {
		// 更新最后使用时间和连接数
		sessionInfo.lastUsed = time.Now()
		atomic.AddInt32(&sessionInfo.activeConns, 1)
		rsm.mutex.RUnlock()

		// 返回一个代理服务器，将请求转发到远程客户端
		return rsm.createProxyServer(serverID, sessionInfo), nil
	}
	rsm.mutex.RUnlock()

	// 创建新连接
	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	// 双重检查
	if sessionInfo, exists := rsm.sessions[serverID]; exists {
		sessionInfo.lastUsed = time.Now()
		atomic.AddInt32(&sessionInfo.activeConns, 1)
		return rsm.createProxyServer(serverID, sessionInfo), nil
	}

	// 获取配置
	config, err := rsm.db.GetSSEServiceConfig(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSE service config: %w", err)
	}

	// 连接到远程服务
	session, client, err := rsm.connectToRemoteSSEService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote SSE service: %w", err)
	}

	logger.Info("Successfully connected to remote SSE service: %s", serverID)

	// 存储会话信息
	sessionInfo := &SSESessionInfo{
		session:     session,
		client:      client,
		lastUsed:    time.Now(),
		config:      config,
		activeConns: 1,
	}
	rsm.sessions[serverID] = sessionInfo

	return rsm.createProxyServer(serverID, sessionInfo), nil
}

// connectToRemoteSSEService 连接到远程 SSE 服务
func (rsm *RemoteSSEManager) connectToRemoteSSEService(config *models.MCPServiceSSE) (*mcp.ClientSession, *mcp.Client, error) {
	// 构建完整的 URL
	fullURL := config.BaseURL + config.SSEPath
	logger.Info("Connecting to remote SSE service: %s", fullURL)

	// 创建HTTP客户端，添加自定义头部
	var headers map[string]string
	if config.Headers != nil {
		headerMap := map[string]interface{}(config.Headers)
		headers = make(map[string]string)
		for k, v := range headerMap {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	// 创建传输选项 - 注意：URL 作为第一个参数传递给 NewSSEClientTransport
	options := &mcp.SSEClientTransportOptions{}

	// 如果有自定义头部，创建自定义的HTTP客户端
	if len(headers) > 0 {
		httpClient := &http.Client{
			Transport: &headerRoundTripper{
				base:    http.DefaultTransport,
				headers: headers,
			},
		}
		options.HTTPClient = httpClient
	}

	// 创建SSE传输
	transport := mcp.NewSSEClientTransport(fullURL, options)

	// 创建客户端
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-proxy-client",
		Version: "1.0.0",
	}, nil)

	// 启动客户端
	ctx := context.Background()
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to remote service: %w", err)
	}

	return session, client, nil
}

// createProxyServer 创建代理服务器
func (rsm *RemoteSSEManager) createProxyServer(serverID string, sessionInfo *SSESessionInfo) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-sse-proxy-server",
		Version: "1.0.0",
	}, nil)

	// 尝试获取远程服务器的工具列表
	toolsResult, err := sessionInfo.session.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		logger.Error("Failed to list tools from remote SSE service %s: %v", serverID, err)
		logger.Error("Returning basic proxy server for %s without pre-loaded tools", serverID)
		return server
	}

	// 为每个工具添加代理
	for _, toolPtr := range toolsResult.Tools {
		rsm.addProxyTool(server, sessionInfo, *toolPtr)
	}

	return server
}

// addProxyTool 添加代理工具
func (rsm *RemoteSSEManager) addProxyTool(server *mcp.Server, sessionInfo *SSESessionInfo, tool mcp.Tool) {
	// 创建符合 ToolHandlerFor 类型的处理器
	toolHandler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
		// 转换参数类型
		callParams := &mcp.CallToolParams{
			Name:      params.Name,
			Arguments: params.Arguments,
		}

		// 调用远程服务
		result, err := sessionInfo.session.CallTool(ctx, callParams)
		if err != nil {
			return nil, fmt.Errorf("failed to call remote tool %s: %w", tool.Name, err)
		}

		// 转换返回类型
		return &mcp.CallToolResultFor[any]{
			Content: result.Content,
			IsError: result.IsError,
		}, nil
	}

	mcp.AddTool(server, &tool, toolHandler)
}

// CleanupIdleSessions 清理空闲会话
func (rsm *RemoteSSEManager) CleanupIdleSessions(idleTimeout time.Duration) {
	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	now := time.Now()
	for serverID, sessionInfo := range rsm.sessions {
		if atomic.LoadInt32(&sessionInfo.activeConns) == 0 &&
			now.Sub(sessionInfo.lastUsed) > idleTimeout {
			logger.Info("Cleaning up idle SSE session for server: %s", serverID)

			// 关闭会话
			if sessionInfo.session != nil {
				sessionInfo.session.Close()
			}

			delete(rsm.sessions, serverID)
		}
	}
}
