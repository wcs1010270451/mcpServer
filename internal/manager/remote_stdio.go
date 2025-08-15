package manager

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"McpServer/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SessionInfo 会话信息
type SessionInfo struct {
	session      *mcp.ClientSession
	client       *mcp.Client
	lastUsed     time.Time
	config       *models.MCPServiceStdio
	activeConns  int32            // 活跃连接数
	userSessions map[string]int32 // 用户会话计数 (userID -> count)
	sessionKeys  map[string]bool  // 会话键集合 (for per_session strategy)
}

// RemoteStdioManager 管理远程 stdio MCP 服务
type RemoteStdioManager struct {
	db       DatabaseServiceInterface
	sessions map[string]*SessionInfo
	mutex    sync.RWMutex
	stopChan chan struct{}
}

// NewRemoteStdioManager 创建新的远程 stdio 管理器
func NewRemoteStdioManager(db DatabaseServiceInterface) *RemoteStdioManager {
	manager := &RemoteStdioManager{
		db:       db,
		sessions: make(map[string]*SessionInfo),
		stopChan: make(chan struct{}),
	}

	// 启动清理协程
	go manager.cleanupRoutine()

	return manager
}

// GetOrCreateRemoteServer 获取或创建远程服务器连接
func (rsm *RemoteStdioManager) GetOrCreateRemoteServer(serverID string) (*mcp.Server, error) {
	return rsm.GetOrCreateRemoteServerWithContext(serverID, "", "")
}

// GetOrCreateRemoteServerWithContext 获取或创建远程服务器连接（带用户和会话上下文）
func (rsm *RemoteStdioManager) GetOrCreateRemoteServerWithContext(serverID, userID, sessionKey string) (*mcp.Server, error) {
	// 获取配置以确定复用策略
	config, err := rsm.db.GetStdioServiceConfig(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stdio config: %w", err)
	}

	// 根据复用策略生成实际的 session 键
	actualSessionKey := rsm.generateSessionKey(serverID, userID, sessionKey, config.ReuseStrategy)

	rsm.mutex.RLock()
	if sessionInfo, exists := rsm.sessions[actualSessionKey]; exists {
		// 检查是否超过最大并发数
		if config.MaxConcurrent > 0 && atomic.LoadInt32(&sessionInfo.activeConns) >= int32(config.MaxConcurrent) {
			rsm.mutex.RUnlock()
			return nil, fmt.Errorf("service %s reached maximum concurrent connections (%d)", serverID, config.MaxConcurrent)
		}

		// 更新最后使用时间和连接数
		sessionInfo.lastUsed = time.Now()
		atomic.AddInt32(&sessionInfo.activeConns, 1)

		// 更新用户会话计数
		if userID != "" {
			if sessionInfo.userSessions == nil {
				sessionInfo.userSessions = make(map[string]int32)
			}
			sessionInfo.userSessions[userID]++
		}

		rsm.mutex.RUnlock()

		log.Printf("Reusing existing session for %s (strategy: %s, active: %d/%d)",
			serverID, config.ReuseStrategy, atomic.LoadInt32(&sessionInfo.activeConns), config.MaxConcurrent)

		// 返回一个代理服务器，将请求转发到远程客户端
		return rsm.createProxyServer(actualSessionKey, sessionInfo), nil
	}
	rsm.mutex.RUnlock()

	// 创建新连接
	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	// 双重检查
	if sessionInfo, exists := rsm.sessions[actualSessionKey]; exists {
		if config.MaxConcurrent > 0 && atomic.LoadInt32(&sessionInfo.activeConns) >= int32(config.MaxConcurrent) {
			return nil, fmt.Errorf("service %s reached maximum concurrent connections (%d)", serverID, config.MaxConcurrent)
		}

		sessionInfo.lastUsed = time.Now()
		atomic.AddInt32(&sessionInfo.activeConns, 1)

		if userID != "" {
			if sessionInfo.userSessions == nil {
				sessionInfo.userSessions = make(map[string]int32)
			}
			sessionInfo.userSessions[userID]++
		}

		return rsm.createProxyServer(actualSessionKey, sessionInfo), nil
	}

	log.Printf("Creating new session for %s (strategy: %s, max_concurrent: %d)",
		serverID, config.ReuseStrategy, config.MaxConcurrent)

	// 创建客户端连接
	session, client, err := rsm.connectToRemoteService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote service: %w", err)
	}

	sessionInfo := &SessionInfo{
		session:      session,
		client:       client,
		lastUsed:     time.Now(),
		config:       config,
		activeConns:  1,
		userSessions: make(map[string]int32),
		sessionKeys:  make(map[string]bool),
	}

	// 初始化用户会话计数
	if userID != "" {
		sessionInfo.userSessions[userID] = 1
	}

	rsm.sessions[actualSessionKey] = sessionInfo

	log.Printf("Successfully connected to remote stdio service: %s (session_key: %s)", serverID, actualSessionKey)
	return rsm.createProxyServer(actualSessionKey, sessionInfo), nil
}

// generateSessionKey 根据复用策略生成会话键
func (rsm *RemoteStdioManager) generateSessionKey(serverID, userID, sessionKey, reuseStrategy string) string {
	switch reuseStrategy {
	case "per_user":
		if userID != "" {
			return fmt.Sprintf("%s:user:%s", serverID, userID)
		}
		return fmt.Sprintf("%s:user:anonymous", serverID)
	case "per_session":
		if sessionKey != "" {
			return fmt.Sprintf("%s:session:%s", serverID, sessionKey)
		}
		return fmt.Sprintf("%s:session:%d", serverID, time.Now().UnixNano())
	case "shared":
		fallthrough
	default:
		return serverID // 共享模式，所有用户使用相同的服务器ID作为键
	}
}

// connectToRemoteService 连接到远程服务
func (rsm *RemoteStdioManager) connectToRemoteService(config *models.MCPServiceStdio) (*mcp.ClientSession, *mcp.Client, error) {
	ctx := context.Background()

	// 创建客户端
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-proxy-client",
		Version: "1.0.0",
	}, nil)

	// 创建命令
	cmd := exec.Command(config.Command, config.Args...)

	// 设置工作目录
	if config.Workdir != nil && *config.Workdir != "" {
		cmd.Dir = *config.Workdir
	}

	// 设置环境变量
	if config.Env != nil {
		cmd.Env = os.Environ()
		for key, value := range config.Env {
			if valueStr, ok := value.(string); ok {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, valueStr))
			}
		}
	}

	// 设置错误输出
	cmd.Stderr = log.Writer()

	// 创建传输
	transport := mcp.NewCommandTransport(cmd)

	log.Printf("Connecting to remote stdio service with command: %s %v", config.Command, config.Args)

	// 创建超时上下文
	timeout := time.Duration(config.StartupTimeoutMs) * time.Millisecond
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 连接
	session, err := client.Connect(connectCtx, transport)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}

	return session, client, nil
}

// createProxyServer 创建代理服务器，将 SSE 请求转发到远程 stdio 客户端
func (rsm *RemoteStdioManager) createProxyServer(serverID string, sessionInfo *SessionInfo) *mcp.Server {
	// 创建一个代理服务器
	server := mcp.NewServer(&mcp.Implementation{
		Name:    fmt.Sprintf("proxy-%s", serverID),
		Version: "1.0.0",
	}, nil)

	// 获取远程服务的工具列表
	ctx := context.Background()
	toolsResult, err := sessionInfo.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Printf("Failed to list tools from remote service %s: %v", serverID, err)
		return server
	}

	// 为每个远程工具创建代理工具
	for _, tool := range toolsResult.Tools {
		rsm.addProxyTool(server, sessionInfo, *tool)
	}

	return server
}

// addProxyTool 添加代理工具
func (rsm *RemoteStdioManager) addProxyTool(server *mcp.Server, sessionInfo *SessionInfo, tool mcp.Tool) {
	// 创建符合 ToolHandlerFor 类型的处理器
	toolHandler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
		// 增加活跃连接数
		atomic.AddInt32(&sessionInfo.activeConns, 1)
		sessionInfo.lastUsed = time.Now()

		// 转换参数类型
		callParams := &mcp.CallToolParams{
			Name:      params.Name,
			Arguments: params.Arguments,
		}

		result, err := sessionInfo.session.CallTool(ctx, callParams)
		if err != nil {
			// 减少活跃连接数
			atomic.AddInt32(&sessionInfo.activeConns, -1)
			return nil, fmt.Errorf("remote call failed: %w", err)
		}

		// 减少活跃连接数
		atomic.AddInt32(&sessionInfo.activeConns, -1)

		// 转换返回类型
		return &mcp.CallToolResultFor[any]{
			Content: result.Content,
			IsError: result.IsError,
		}, nil
	}

	mcp.AddTool(server, &tool, toolHandler)
	log.Printf("Added proxy tool: %s", tool.Name)
}

// cleanupRoutine 清理协程，根据配置策略管理会话生命周期
func (rsm *RemoteStdioManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rsm.cleanupIdleSessions()
		case <-rsm.stopChan:
			return
		}
	}
}

// cleanupIdleSessions 清理空闲会话
func (rsm *RemoteStdioManager) cleanupIdleSessions() {
	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for serverID, sessionInfo := range rsm.sessions {
		// 检查策略和空闲时间
		switch sessionInfo.config.ReuseStrategy {
		case "per_session":
			// 每个会话独立，当没有活跃连接且超过TTL时关闭
			if atomic.LoadInt32(&sessionInfo.activeConns) == 0 {
				idleDuration := now.Sub(sessionInfo.lastUsed)
				ttl := time.Duration(sessionInfo.config.IdleTtlMs) * time.Millisecond
				if idleDuration > ttl {
					log.Printf("Closing idle session %s (idle for %v)", serverID, idleDuration)
					toDelete = append(toDelete, serverID)
				}
			}
		case "singleton":
			// 单例模式，只在程序退出时关闭
			// 不做任何操作
		case "per_request":
			// 每个请求独立，立即关闭（在工具调用完成后）
			if atomic.LoadInt32(&sessionInfo.activeConns) == 0 {
				log.Printf("Closing per-request session %s", serverID)
				toDelete = append(toDelete, serverID)
			}
		}
	}

	// 清理标记的会话
	for _, serverID := range toDelete {
		if sessionInfo, exists := rsm.sessions[serverID]; exists {
			sessionInfo.session.Close()
			delete(rsm.sessions, serverID)
		}
	}
}

// CloseRemoteSession 关闭指定的远程会话
func (rsm *RemoteStdioManager) CloseRemoteSession(serverID string) {
	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	if sessionInfo, exists := rsm.sessions[serverID]; exists {
		sessionInfo.session.Close()
		delete(rsm.sessions, serverID)
		log.Printf("Manually closed remote session: %s", serverID)
	}
}

// CloseAll 关闭所有远程会话
func (rsm *RemoteStdioManager) CloseAll() {
	// 停止清理协程
	close(rsm.stopChan)

	rsm.mutex.Lock()
	defer rsm.mutex.Unlock()

	for serverID, sessionInfo := range rsm.sessions {
		sessionInfo.session.Close()
		log.Printf("Closed remote session: %s", serverID)
	}

	rsm.sessions = make(map[string]*SessionInfo)
}

// GetSessionStats 获取会话统计信息
func (rsm *RemoteStdioManager) GetSessionStats() map[string]interface{} {
	rsm.mutex.RLock()
	defer rsm.mutex.RUnlock()

	stats := make(map[string]interface{})
	totalSessions := 0
	totalConnections := int32(0)
	strategyStats := make(map[string]map[string]interface{})

	for sessionKey, sessionInfo := range rsm.sessions {
		totalSessions++
		activeConns := atomic.LoadInt32(&sessionInfo.activeConns)
		totalConnections += activeConns

		strategy := sessionInfo.config.ReuseStrategy
		if strategyStats[strategy] == nil {
			strategyStats[strategy] = map[string]interface{}{
				"sessions":    0,
				"connections": int32(0),
				"max_allowed": 0,
				"services":    []string{},
			}
		}

		strategyStats[strategy]["sessions"] = strategyStats[strategy]["sessions"].(int) + 1
		strategyStats[strategy]["connections"] = strategyStats[strategy]["connections"].(int32) + activeConns
		strategyStats[strategy]["max_allowed"] = sessionInfo.config.MaxConcurrent

		services := strategyStats[strategy]["services"].([]string)
		found := false
		for _, service := range services {
			if service == sessionInfo.config.ServerID {
				found = true
				break
			}
		}
		if !found {
			services = append(services, sessionInfo.config.ServerID)
			strategyStats[strategy]["services"] = services
		}

		// 详细会话信息
		sessionDetail := map[string]interface{}{
			"server_id":      sessionInfo.config.ServerID,
			"strategy":       strategy,
			"active_conns":   activeConns,
			"max_concurrent": sessionInfo.config.MaxConcurrent,
			"last_used":      sessionInfo.lastUsed,
			"user_sessions":  len(sessionInfo.userSessions),
		}
		stats[sessionKey] = sessionDetail
	}

	summary := map[string]interface{}{
		"total_sessions":    totalSessions,
		"total_connections": totalConnections,
		"by_strategy":       strategyStats,
		"session_details":   stats,
	}

	return summary
}
