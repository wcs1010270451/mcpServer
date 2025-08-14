package manager

import (
	"bufio"
	"context"
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
	ServerID  string
	SessionID string
	Config    *models.MCPServiceSSE
	LastUsed  time.Time
}

// SessionManager 管理 MCP 会话
type SessionManager struct {
	manager      MCPServerManagerInterface
	db           DatabaseServiceInterface
	mcpHandlers  map[string]http.Handler     // serverID -> MCP Handler
	sessions     map[string]*HTTPSessionInfo // sessionID -> HTTPSessionInfo
	handlerMutex sync.RWMutex
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(manager MCPServerManagerInterface, db DatabaseServiceInterface) *SessionManager {
	return &SessionManager{
		manager:     manager,
		db:          db,
		mcpHandlers: make(map[string]http.Handler),
		sessions:    make(map[string]*HTTPSessionInfo),
	}
}

// HandleInitialConnection 处理初始连接请求（GET 请求 + server_id）
func (sm *SessionManager) HandleInitialConnection(w http.ResponseWriter, r *http.Request, serverID string) {
	log.Printf("Handling initial connection for server: %s", serverID)

	// 检查是否已有缓存的处理器
	sm.handlerMutex.RLock()
	if handler, exists := sm.mcpHandlers[serverID]; exists {
		sm.handlerMutex.RUnlock()
		log.Printf("Using cached MCP handler for server: %s", serverID)
		handler.ServeHTTP(w, r)
		return
	}
	sm.handlerMutex.RUnlock()

	// 首先检查是否为远程 SSE 服务
	isSSE, err := sm.manager.GetDB().IsRemoteSSEService(serverID)
	if err != nil {
		log.Printf("Error checking SSE service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if isSSE {
		// 对于远程 SSE 服务，直接透传
		log.Printf("Handling remote SSE service: %s", serverID)
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
	log.Printf("Creating new MCP SSE handler for server: %s", serverID)
	mcpHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		log.Printf("MCP SSE Handler called for server %s, method: %s, URL: %s", serverID, request.Method, request.URL.String())
		return server
	})

	// 缓存处理器
	sm.handlerMutex.Lock()
	sm.mcpHandlers[serverID] = mcpHandler
	sm.handlerMutex.Unlock()

	log.Printf("Successfully created and cached MCP handler for server: %s", serverID)
	mcpHandler.ServeHTTP(w, r)
}

// HandleSessionRequest 处理会话请求（POST 请求或无 server_id 的请求）
func (sm *SessionManager) HandleSessionRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling session request: %s %s", r.Method, r.URL.String())

	// 检查是否是 POST 到 /message 或 /messages 端点
	if r.Method == "POST" && (strings.HasPrefix(r.URL.Path, "/message") || strings.HasPrefix(r.URL.Path, "/messages")) {
		sessionID := r.URL.Query().Get("sessionId")
		if sessionID == "" {
			sessionID = r.URL.Query().Get("session_id")
		}
		if sessionID != "" {
			log.Printf("Detected POST request to %s with sessionId: %s", r.URL.Path, sessionID)
			sm.handleSessionMessage(w, r, sessionID)
			return
		}
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
		log.Printf("No cached handlers available for session request")
		http.Error(w, "No active sessions", http.StatusNotFound)
		return
	}

	// 如果只有一个活跃的处理器，直接使用它
	if len(handlers) == 1 {
		log.Printf("Using single cached handler for session request: %s", serverIDs[0])
		handlers[0].ServeHTTP(w, r)
		return
	}

	// 如果有多个处理器，我们需要更智能的路由策略
	// 这里简化处理：使用第一个可用的处理器
	log.Printf("Multiple handlers available (%d), using first one: %s", len(handlers), serverIDs[0])
	handlers[0].ServeHTTP(w, r)
}

// handleSessionMessage 处理基于 sessionId 的消息请求
func (sm *SessionManager) handleSessionMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	// 查找会话信息
	sm.handlerMutex.RLock()
	sessionInfo, exists := sm.sessions[sessionID]
	sm.handlerMutex.RUnlock()

	if !exists {
		log.Printf("Session not found: %s", sessionID)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	log.Printf("Found session for sessionId %s, forwarding to server: %s", sessionID, sessionInfo.ServerID)

	// 构建远程 URL - 使用正确的端点格式
	remoteURL := sessionInfo.Config.BaseURL + "/messages/?session_id=" + sessionID

	log.Printf("Forwarding message to: %s", remoteURL)

	// 创建到远程服务的请求
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, r.Method, remoteURL, r.Body)
	if err != nil {
		log.Printf("Failed to create remote request: %v", err)
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
				if h, ok := config.AuthConfig["header"].(string); ok {
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
		log.Printf("Failed to send message to remote service: %v", err)
		http.Error(w, "Failed to send message", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Printf("Remote service responded with status: %d", resp.StatusCode)

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
		log.Printf("Error copying response body: %v", err)
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
		log.Printf("Cleaned up MCP handler for server: %s", serverID)
	}
}

// handleRemoteSSEProxy 直接透传远程 SSE 服务
func (sm *SessionManager) handleRemoteSSEProxy(w http.ResponseWriter, r *http.Request, serverID string) {
	// 获取远程 SSE 服务配置
	config, err := sm.manager.GetDB().GetSSEServiceConfig(serverID)
	if err != nil {
		log.Printf("Failed to get SSE config for %s: %v", serverID, err)
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

	log.Printf("Proxying SSE request to: %s", remoteURL)
	log.Printf("Original request method: %s, URL: %s", r.Method, r.URL.String())
	log.Printf("Request headers: %v", r.Header)

	// 创建到远程服务的请求
	ctx := context.Background()

	// 对于 SSE 连接，通常初始请求应该是 GET
	method := r.Method
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/event-stream") || r.Method == "GET" {
		method = "GET"
		log.Printf("Using GET method for SSE connection (Accept: %s)", acceptHeader)
	}

	// 对于 GET 请求，不应该发送 body
	var body io.Reader
	if method != "GET" {
		body = r.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, remoteURL, body)
	if err != nil {
		log.Printf("Failed to create remote request: %v", err)
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
				if h, ok := config.AuthConfig["header"].(string); ok {
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

	log.Printf("Sending request to remote: Method=%s, URL=%s", req.Method, req.URL.String())
	log.Printf("Remote request headers: %v", req.Header)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to connect to remote SSE service %s: %v", serverID, err)
		http.Error(w, "Failed to connect to remote service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Printf("Successfully connected to remote SSE service %s, status: %d", serverID, resp.StatusCode)

	// 如果不是成功状态，返回错误
	if resp.StatusCode != http.StatusOK {
		log.Printf("Remote SSE service returned status: %d", resp.StatusCode)
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
		log.Printf("Client does not support SSE")
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// 开始流式传输
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Remote SSE stream ended for: %s", serverID)
			} else {
				log.Printf("Error reading from remote service %s: %v", serverID, err)
			}
			return
		}

		// 检查是否包含 sessionId 信息
		lineStr := string(line)
		log.Printf("SSE line: %s", strings.TrimSpace(lineStr))

		if strings.HasPrefix(lineStr, "event: endpoint") {
			log.Printf("Detected endpoint event for %s", serverID)
		}

		// 检查多种可能的 sessionId 格式
		if strings.Contains(lineStr, "session_id=") || strings.Contains(lineStr, "sessionId=") {
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

			if sessionID != "" {
				log.Printf("Extracted sessionId: %s for server: %s (endpoint: %s)", sessionID, serverID, endpointPath)

				// 存储会话信息
				sm.handlerMutex.Lock()
				sm.sessions[sessionID] = &HTTPSessionInfo{
					ServerID:  serverID,
					SessionID: sessionID,
					Config:    config,
					LastUsed:  time.Now(),
				}
				sm.handlerMutex.Unlock()
			}
		}

		// 写入客户端
		_, writeErr := w.Write(line)
		if writeErr != nil {
			log.Printf("Client disconnected during SSE proxy for: %s", serverID)
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
	log.Printf("Cleaned up all %d MCP handlers", count)
}
