package manager

import (
	"McpServer/internal/logger"
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"McpServer/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPSessionInfo 存储 HTTP 会话信息
type HTTPSessionInfo struct {
	ServerID     string
	SessionID    string
	Config       *models.MCPServiceSSE
	LastUsed     time.Time
	CreatedAt    time.Time
	IsActive     bool
	ConnectionID string // 用于跟踪连接
}

// SessionManager 管理 MCP 会话
type SessionManager struct {
	manager      MCPServerManagerInterface
	db           DatabaseServiceInterface
	mcpHandlers  map[string]http.Handler     // serverID -> MCP Handler
	sessions     map[string]*HTTPSessionInfo // sessionID -> HTTPSessionInfo
	handlerMutex sync.RWMutex

	// 会话清理配置
	sessionTimeout time.Duration // 会话超时时间
	cleanupTicker  *time.Ticker  // 清理定时器
	shutdownChan   chan bool     // 关闭信号
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(manager MCPServerManagerInterface, db DatabaseServiceInterface) *SessionManager {
	sm := &SessionManager{
		manager:        manager,
		db:             db,
		mcpHandlers:    make(map[string]http.Handler),
		sessions:       make(map[string]*HTTPSessionInfo),
		sessionTimeout: 30 * time.Minute, // 30分钟超时
		shutdownChan:   make(chan bool),
	}

	// 启动会话清理协程
	sm.startSessionCleanup()

	return sm
}

// startSessionCleanup 启动会话清理协程
func (sm *SessionManager) startSessionCleanup() {
	sm.cleanupTicker = time.NewTicker(5 * time.Minute) // 每5分钟清理一次

	go func() {
		for {
			select {
			case <-sm.cleanupTicker.C:
				sm.cleanupExpiredSessions()
			case <-sm.shutdownChan:
				sm.cleanupTicker.Stop()
				return
			}
		}
	}()

	logger.Info("Session cleanup started (timeout: %v, interval: 5m)", sm.sessionTimeout)
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	now := time.Now()
	expiredSessions := make([]string, 0)
	expiredServers := make(map[string]bool)

	sm.handlerMutex.Lock()

	// 找出过期的会话
	for sessionID, sessionInfo := range sm.sessions {
		if now.Sub(sessionInfo.LastUsed) > sm.sessionTimeout {
			expiredSessions = append(expiredSessions, sessionID)
			expiredServers[sessionInfo.ServerID] = true
		}
	}

	// 删除过期会话
	for _, sessionID := range expiredSessions {
		delete(sm.sessions, sessionID)
	}

	sm.handlerMutex.Unlock()

	if len(expiredSessions) > 0 {
		logger.Info("Cleaned up %d expired sessions: %v", len(expiredSessions), expiredSessions)

		// 检查是否有服务器没有活跃会话，清理其处理器
		sm.cleanupUnusedHandlers(expiredServers)
	}
}

// cleanupUnusedHandlers 清理没有活跃会话的处理器
func (sm *SessionManager) cleanupUnusedHandlers(potentialServers map[string]bool) {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	// 统计每个服务器的活跃会话数
	activeServers := make(map[string]int)
	for _, sessionInfo := range sm.sessions {
		activeServers[sessionInfo.ServerID]++
	}

	// 清理没有活跃会话的处理器
	for serverID := range potentialServers {
		if activeServers[serverID] == 0 {
			if _, exists := sm.mcpHandlers[serverID]; exists {
				delete(sm.mcpHandlers, serverID)
				logger.Info("Cleaned up handler for inactive server: %s", serverID)
			}
		}
	}
}

// Shutdown 关闭会话管理器
func (sm *SessionManager) Shutdown() {
	close(sm.shutdownChan)
	if sm.cleanupTicker != nil {
		sm.cleanupTicker.Stop()
	}
	logger.Info("Session manager shutdown")
}

// GetSessionInfo 获取会话信息（用于调试和监控）
func (sm *SessionManager) GetSessionInfo() map[string]*HTTPSessionInfo {
	sm.handlerMutex.RLock()
	defer sm.handlerMutex.RUnlock()

	// 复制一份以避免并发问题
	result := make(map[string]*HTTPSessionInfo)
	for k, v := range sm.sessions {
		sessionCopy := *v // 复制值
		result[k] = &sessionCopy
	}
	return result
}

// GetActiveSessionCount 获取活跃会话数量
func (sm *SessionManager) GetActiveSessionCount() int {
	sm.handlerMutex.RLock()
	defer sm.handlerMutex.RUnlock()

	count := 0
	for _, session := range sm.sessions {
		if session.IsActive {
			count++
		}
	}
	return count
}

// HandleInitialConnection 处理初始连接请求（GET 请求 + server_id）
func (sm *SessionManager) HandleInitialConnection(w http.ResponseWriter, r *http.Request, serverID string) {
	logger.Info("Handling initial connection for server: %s", serverID)

	// 检查是否已有缓存的处理器
	sm.handlerMutex.RLock()
	if handler, exists := sm.mcpHandlers[serverID]; exists {
		sm.handlerMutex.RUnlock()
		logger.Info("Using cached MCP handler for server: %s", serverID)
		handler.ServeHTTP(w, r)
		return
	}
	sm.handlerMutex.RUnlock()

	// 首先检查是否为远程 SSE 服务
	isSSE, err := sm.manager.GetDB().IsRemoteSSEService(serverID)
	if err != nil {
		logger.Info("Error checking SSE service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if isSSE {
		// 对于远程 SSE 服务，直接透传
		logger.Info("Handling remote SSE service: %s", serverID)
		sm.handleRemoteSSEProxy(w, r, serverID)
		return
	}

	// 获取本地服务器实例
	server, err := sm.manager.GetServer(serverID)
	if err != nil {
		log.Printf("Error: Server with ID '%s' not found: %v", serverID, err)
		http.Error(w, fmt.Sprintf("Server '%s' not found", serverID), http.StatusNotFound)
		return
	}

	// 创建新的 MCP SSE 处理器
	logger.Info("Creating new MCP SSE handler for server: %s", serverID)
	mcpHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		logger.Info("MCP SSE Handler called for server %s, method: %s, URL: %s", serverID, request.Method, request.URL.String())
		// 这里不在 GET 请求时创建 STDIO 会话，而是在后续的 POST 请求中按需创建
		return server
	})

	// 缓存处理器
	sm.handlerMutex.Lock()
	sm.mcpHandlers[serverID] = mcpHandler
	sm.handlerMutex.Unlock()

	logger.Info("Successfully created and cached MCP handler for server: %s", serverID)
	mcpHandler.ServeHTTP(w, r)
}

// generateSessionID 生成一个新的会话 ID
func (sm *SessionManager) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// generateConnectionID 生成连接 ID
func generateConnectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// isSessionExpired 检查会话是否已过期
func (sm *SessionManager) isSessionExpired(sessionInfo *HTTPSessionInfo) bool {
	return time.Since(sessionInfo.LastUsed) > sm.sessionTimeout
}

// isValidSessionID 验证sessionID格式（简单的长度和字符检查）
func (sm *SessionManager) isValidSessionID(sessionID string) bool {
	if len(sessionID) < 8 || len(sessionID) > 64 {
		return false
	}
	// 检查是否只包含字母数字字符和连字符
	for _, char := range sessionID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}
	return true
}

// createStdioSession 为 STDIO 服务创建虚拟会话
func (sm *SessionManager) createStdioSession(serverID string, r *http.Request) {
	// 从 URL 中提取可能的 sessionID，或生成一个新的
	sessionID := r.URL.Query().Get("sessionid")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}

	// 如果没有 sessionID，生成一个新的
	if sessionID == "" {
		sessionID = sm.generateSessionID()
		logger.Info("Generated new sessionID for STDIO service %s: %s", serverID, sessionID)
	}

	logger.Info("Creating STDIO virtual session: %s -> %s", sessionID, serverID)

	// 创建虚拟会话信息
	now := time.Now()
	sm.handlerMutex.Lock()
	sm.sessions[sessionID] = &HTTPSessionInfo{
		ServerID:     serverID,
		SessionID:    sessionID,
		Config:       nil, // STDIO 服务不需要 SSE 配置
		LastUsed:     now,
		CreatedAt:    now,
		IsActive:     true,
		ConnectionID: generateConnectionID(),
	}
	sm.handlerMutex.Unlock()

	logger.Info("Created virtual session for STDIO service: %s (sessionID: %s)", serverID, sessionID)
}

// HandleSessionRequest 处理会话请求（POST 请求或无 server_id 的请求）
func (sm *SessionManager) HandleSessionRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling session request: %s %s", r.Method, r.URL.String())

	// 检查是否是 POST 到 /message 或 /messages 端点，或者其他带 sessionId/session_id/sessionid 的请求
	sessionID := r.URL.Query().Get("sessionId")
	logger.Info("Checking sessionId (camelCase): '%s'", sessionID)
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
		logger.Info("Checking session_id (underscore): '%s'", sessionID)
	}
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionid") // 支持小写的 sessionid
		logger.Info("Checking sessionid (lowercase): '%s'", sessionID)
	}

	// 调试：显示所有查询参数
	log.Printf("All query parameters: %v", r.URL.Query())
	logger.Debug("All query parameters: %v", r.URL.Query())

	if sessionID != "" {
		logger.Debug("Detected request to %s with sessionId: %s", r.URL.Path, sessionID)
		logger.Debug("Routing to handleSessionMessage for sessionId: %s", sessionID)
		sm.handleSessionMessage(w, r, sessionID)
		return
	}

	// 检查是否是 POST 到 /message 或 /messages 端点（无 sessionId 参数的情况）
	if r.Method == "POST" && (strings.HasPrefix(r.URL.Path, "/message") || strings.HasPrefix(r.URL.Path, "/messages")) {
		logger.Debug("POST request to %s without sessionId, checking request body", r.URL.Path)
		// 可能需要从请求体中提取 sessionId，但这里先使用现有逻辑
	}

	// 遍历所有已缓存的处理器，让它们尝试处理这个请求
	sm.handlerMutex.RLock()
	handlers := make([]http.Handler, 0, len(sm.mcpHandlers))
	serverIDs := make([]string, 0, len(sm.mcpHandlers))
	for serverID, handler := range sm.mcpHandlers {
		handlers = append(handlers, handler)
		serverIDs = append(serverIDs, serverID)
	}
	sm.handlerMutex.RUnlock()

	if len(handlers) == 0 {
		logger.Error("No cached handlers available for session request")
		http.Error(w, "No active sessions", http.StatusNotFound)
		return
	}

	// 如果只有一个活跃的处理器，仍然需要验证session（如果提供了sessionID）
	if len(handlers) == 1 {
		// 检查是否有明确的sessionID参数，如果有则需要验证
		sessionID = r.URL.Query().Get("sessionId")
		if sessionID == "" {
			sessionID = r.URL.Query().Get("session_id")
		}
		if sessionID == "" {
			sessionID = r.URL.Query().Get("sessionid")
		}

		if sessionID != "" {
			// 有sessionID的请求必须通过正常的session验证流程
			logger.Info("Request with sessionID %s, routing to session validation", sessionID)
			sm.handleSessionMessage(w, r, sessionID)
			return
		}

		logger.Info("Using single cached handler for session request without sessionID: %s", serverIDs[0])
		handlers[0].ServeHTTP(w, r)
		return
	}

	// 如果有多个处理器，尝试从 URL 路径或其他信息推断目标服务器
	// 查看是否有服务器 ID 的线索
	var targetHandler http.Handler
	var targetServerID string

	// 1. 尝试从 URL 路径中推断目标服务器
	for i, serverID := range serverIDs {
		if strings.Contains(r.URL.String(), serverID) {
			targetHandler = handlers[i]
			targetServerID = serverID
			logger.Info("Inferred target server from URL: %s", serverID)
			break
		}
	}

	// 2. 尝试从 Referer 头推断目标服务器（如果请求来自特定的服务器页面）
	if targetHandler == nil {
		referer := r.Header.Get("Referer")
		if referer != "" {
			for i, serverID := range serverIDs {
				if strings.Contains(referer, serverID) {
					targetHandler = handlers[i]
					targetServerID = serverID
					logger.Info("Inferred target server from Referer: %s", serverID)
					break
				}
			}
		}
	}

	// 3. 对于缺少明确路由信息的请求，检查是否只有一个非 employee 服务器
	if targetHandler == nil {
		var nonEmployeeHandlers []http.Handler
		var nonEmployeeServerIDs []string

		for i, serverID := range serverIDs {
			if serverID != "server_employee_info" {
				nonEmployeeHandlers = append(nonEmployeeHandlers, handlers[i])
				nonEmployeeServerIDs = append(nonEmployeeServerIDs, serverID)
			}
		}

		if len(nonEmployeeHandlers) == 1 {
			targetHandler = nonEmployeeHandlers[0]
			targetServerID = nonEmployeeServerIDs[0]
			log.Printf("Using the only non-employee server: %s", targetServerID)
		}
	}

	// 如果无法推断，记录警告并拒绝请求
	if targetHandler == nil {
		logger.Error("Multiple handlers available (%d): %v, but cannot determine target server", len(handlers), serverIDs)
		logger.Error("Request URL: %s", r.URL.String())
		logger.Error("Please specify sessionId parameter to route to the correct server")
		http.Error(w, "Cannot determine target server. Multiple active sessions found. Please use sessionId parameter.",
			http.StatusBadRequest)
		return
	}

	log.Printf("Using inferred handler for server: %s", targetServerID)
	targetHandler.ServeHTTP(w, r)
}

// handleSessionMessage 处理基于 sessionId 的消息请求
func (sm *SessionManager) handleSessionMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	// 验证sessionID格式
	if !sm.isValidSessionID(sessionID) {
		logger.Error("Invalid sessionID format: %s", sessionID)
		http.Error(w, "Invalid session ID format", http.StatusBadRequest)
		return
	}

	// 查找会话信息
	sm.handlerMutex.RLock()
	sessionInfo, exists := sm.sessions[sessionID]

	// 如果会话存在，检查是否已过期
	if exists && sm.isSessionExpired(sessionInfo) {
		logger.Info("Session %s has expired, removing it", sessionID)
		sm.handlerMutex.RUnlock()

		// 删除过期会话
		sm.handlerMutex.Lock()
		delete(sm.sessions, sessionID)
		sm.handlerMutex.Unlock()

		exists = false
	} else {
		sm.handlerMutex.RUnlock()
	}

	// 添加调试信息
	logger.Debug("Looking up sessionId: %s", sessionID)
	logger.Debug("Available sessions: %d", len(sm.sessions))

	if !exists {
		// POST请求不应该创建新session，session应该在GET连接时已经建立
		logger.Error("Session not found: %s. Sessions must be established through initial GET connection, not POST requests.", sessionID)
		http.Error(w, "Session not found. Please establish connection first.", http.StatusNotFound)
		return
	}

	logger.Info("Found session for sessionId %s, forwarding to server: %s", sessionID, sessionInfo.ServerID)

	// 更新最后使用时间
	sessionInfo.LastUsed = time.Now()

	// 如果是 STDIO 服务（Config 为 nil），直接路由到对应的缓存处理器
	if sessionInfo.Config == nil {
		logger.Info("Routing STDIO session %s to server: %s", sessionID, sessionInfo.ServerID)

		sm.handlerMutex.RLock()
		handler, exists1 := sm.mcpHandlers[sessionInfo.ServerID]
		sm.handlerMutex.RUnlock()

		if !exists1 {
			logger.Error("Handler not found for STDIO server: %s", sessionInfo.ServerID)
			http.Error(w, "Server handler not available", http.StatusInternalServerError)
			return
		}

		// 直接使用缓存的处理器
		handler.ServeHTTP(w, r)
		return
	}

	// 对于 SSE 服务，构建远程 URL - 使用正确的端点格式
	remoteURL := sessionInfo.Config.BaseURL + "/messages/?session_id=" + sessionID

	logger.Info("Forwarding message to: %s", remoteURL)

	// 创建到远程服务的请求
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, r.Method, remoteURL, r.Body)
	if err != nil {
		logger.Error("Failed to create remote request: %v", err)
		http.Error(w, "Failed to create remote request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		if name == "Host" || name == "Connection" {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// 添加认证头部（如果需要）
	config := sessionInfo.Config
	if config.AuthType != "none" && config.AuthConfig != nil {
		switch config.AuthType {
		case "bearer_token":
			if token, ok := config.AuthConfig["token"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		case "api_key":
			if key, ok := config.AuthConfig["key"].(string); ok {
				headerName := "X-API-Key"
				if h, ok1 := config.AuthConfig["header"].(string); ok1 {
					headerName = h
				}
				req.Header.Set(headerName, key)
			}
		case "custom_header":
			for key, value := range config.AuthConfig {
				if valueStr, ok := value.(string); ok {
					req.Header.Set(key, valueStr)
				}
			}
		}
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send message to remote service: %v", err)
		http.Error(w, "Failed to send message", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	logger.Info("Remote service responded with status: %d", resp.StatusCode)
	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("Error copying response body: %v", err)
	}

	// 更新最后使用时间
	sm.handlerMutex.Lock()
	sessionInfo.LastUsed = time.Now()
	sm.handlerMutex.Unlock()
}

// CleanupHandler 清理特定服务器的处理器
func (sm *SessionManager) CleanupHandler(serverID string) {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	if _, exists := sm.mcpHandlers[serverID]; exists {
		delete(sm.mcpHandlers, serverID)
		logger.Info("Cleaned up MCP handler for server: %s", serverID)
	}
}

// handleRemoteSSEProxy 直接透传远程 SSE 服务
func (sm *SessionManager) handleRemoteSSEProxy(w http.ResponseWriter, r *http.Request, serverID string) {
	// 获取远程 SSE 服务配置
	config, err := sm.manager.GetDB().GetSSEServiceConfig(serverID)
	if err != nil {
		logger.Error("Failed to get SSE config for %s: %v", serverID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 构建远程 URL
	remoteURL := config.BaseURL + config.SSEPath

	// 移除 server_id 参数，保留其他查询参数
	query := r.URL.Query()
	query.Del("server_id")
	if len(query) > 0 {
		remoteURL += "?" + query.Encode()
	}

	logger.Debug("Proxying SSE request to: %s", remoteURL)
	logger.Debug("Original request method: %s, URL: %s", r.Method, r.URL.String())
	logger.Debug("Request headers: %v", r.Header)

	// 创建到远程服务的请求
	ctx := context.Background()

	// 对于 SSE 连接，通常初始请求应该是 GET
	method := r.Method
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/event-stream") || r.Method == "GET" {
		method = "GET"
		logger.Debug("Using GET method for SSE connection (Accept: %s)", acceptHeader)
	}

	// 对于 GET 请求，不应该发送 body
	var body io.Reader
	if method != "GET" {
		body = r.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, remoteURL, body)
	if err != nil {
		logger.Error("Failed to create remote request: %v", err)
		http.Error(w, "Failed to create remote request", http.StatusInternalServerError)
		return
	}

	// 复制请求头，但排除一些不需要的头
	for name, values := range r.Header {
		if name == "Host" || name == "Connection" {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// 添加认证头部
	if config.AuthType != "none" && config.AuthConfig != nil {
		switch config.AuthType {
		case "bearer_token":
			if token, ok := config.AuthConfig["token"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		case "api_key":
			if key, ok := config.AuthConfig["key"].(string); ok {
				headerName := "X-API-Key"
				if h, ok1 := config.AuthConfig["header"].(string); ok1 {
					headerName = h
				}
				req.Header.Set(headerName, key)
			}
		case "custom_header":
			for key, value := range config.AuthConfig {
				if valueStr, ok := value.(string); ok {
					req.Header.Set(key, valueStr)
				}
			}
		}
	}

	// 添加配置中的默认头部
	if config.Headers != nil {
		for key, value := range config.Headers {
			if valueStr, ok := value.(string); ok {
				req.Header.Set(key, valueStr)
			}
		}
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	logger.Debug("Sending request to remote: Method=%s, URL=%s", req.Method, req.URL.String())
	logger.Debug("Remote request headers: %v", req.Header)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to connect to remote SSE service %s: %v", serverID, err)
		http.Error(w, "Failed to connect to remote service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	logger.Debug("Successfully connected to remote SSE service %s, status: %d", serverID, resp.StatusCode)

	// 如果不是成功状态，返回错误
	if resp.StatusCode != http.StatusOK {
		logger.Error("Remote SSE service returned status: %d", resp.StatusCode)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 检查客户端是否支持 SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error("Client does not support SSE")
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// 开始流式传输
	reader := bufio.NewReader(resp.Body)
	for {
		line, err1 := reader.ReadBytes('\n')
		if err1 != nil {
			if err1 == io.EOF {
				logger.Error("Remote SSE stream ended for: %s", serverID)
			} else {
				logger.Error("Error reading from remote service %s: %v", serverID, err1)
			}
			return
		}

		// 检查是否包含 sessionId 信息
		lineStr := string(line)
		logger.Info("SSE line: %s", strings.TrimSpace(lineStr))

		if strings.HasPrefix(lineStr, "event: endpoint") {
			logger.Info("Detected endpoint event for %s", serverID)
		}

		// 检查多种可能的 sessionId 格式
		if strings.Contains(lineStr, "session_id=") || strings.Contains(lineStr, "sessionId=") || strings.Contains(lineStr, "sessionid=") {
			var sessionID string
			var endpointPath string

			// 处理 /messages/?session_id= 格式
			if strings.Contains(lineStr, "/messages/?session_id=") {
				parts := strings.Split(lineStr, "session_id=")
				if len(parts) > 1 {
					sessionID = strings.TrimSpace(parts[1])
					endpointPath = "/messages/"
				}
			}
			// 处理 /message?sessionId= 格式
			if strings.Contains(lineStr, "/message?sessionId=") {
				parts := strings.Split(lineStr, "sessionId=")
				if len(parts) > 1 {
					sessionID = strings.TrimSpace(parts[1])
					endpointPath = "/message"
				}
			}
			// 处理 sessionid= 格式（小写）
			if sessionID == "" && strings.Contains(lineStr, "sessionid=") {
				parts := strings.Split(lineStr, "sessionid=")
				if len(parts) > 1 {
					sessionID = strings.TrimSpace(parts[1])
					// 尝试推断端点路径
					if strings.Contains(lineStr, "/messages/") {
						endpointPath = "/messages/"
					} else if strings.Contains(lineStr, "/message") {
						endpointPath = "/message"
					} else {
						endpointPath = "/messages/" // 默认
					}
				}
			}

			if sessionID != "" {
				logger.Info("Extracted sessionId: %s for server: %s (endpoint: %s)", sessionID, serverID, endpointPath)

				// 存储会话信息
				now := time.Now()
				sm.handlerMutex.Lock()
				sm.sessions[sessionID] = &HTTPSessionInfo{
					ServerID:     serverID,
					SessionID:    sessionID,
					Config:       config,
					LastUsed:     now,
					CreatedAt:    now,
					IsActive:     true,
					ConnectionID: generateConnectionID(),
				}
				sm.handlerMutex.Unlock()
			}
		}

		// 写入客户端
		_, writeErr := w.Write(line)
		if writeErr != nil {
			logger.Error("Error writing to client for %s: %v", serverID, writeErr)
			return
		}

		// 立即刷新
		flusher.Flush()
	}
}

// CleanupAll 清理所有处理器
func (sm *SessionManager) CleanupAll() {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	count := len(sm.mcpHandlers)
	sm.mcpHandlers = make(map[string]http.Handler)
	logger.Info("Cleaned up all %d MCP handlers", count)
}
