package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"McpServer/internal/auth"
	"McpServer/internal/config"
	"McpServer/internal/database"
	"McpServer/internal/handlers"
	"McpServer/internal/logger"
	"McpServer/internal/manager"
)

var (
	configPath = flag.String("config", "config/config.dev.yaml", "path to config file")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// 从环境变量覆盖配置
	config.LoadConfigFromEnv(cfg)

	// 配置日志级别
	logger.SetLevelFromString(cfg.Logging.Level)

	logger.Info("Starting MCP Server with config: %s:%d", cfg.Server.Host, cfg.Server.Port)

	// 创建数据库服务
	db, err := database.NewDatabaseService(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to create database service: %v", err)
		return
	}
	defer db.Close()

	// 创建处理器注册表
	dbAdapter := handlers.NewDatabaseAdapter(db)
	handlerRegistry := handlers.NewToolHandlerRegistry(dbAdapter)

	// 创建 MCP 服务器管理器
	mcpManager := manager.NewMCPServerManager(db, handlerRegistry)

	// 从数据库加载内置服务器配置
	if err = mcpManager.LoadServersFromDatabase(); err != nil {
		logger.Fatal("Failed to load builtin servers from database", err)
	}

	// 创建会话管理器
	sessionManager := manager.NewSessionManager(mcpManager, db)

	// 创建认证中间件
	authMiddleware := auth.NewAuthMiddleware(&cfg.Auth)

	// 创建 HTTP 处理器
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		// 从路径中提取 server_id，格式为 /mcp-server/{server_id}/sse
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		var serverID string

		// 检查路径格式：/mcp-server/{server_id}/sse
		if len(pathParts) >= 3 && pathParts[0] == "mcp-server" && pathParts[2] == "sse" {
			serverID = pathParts[1]
		}

		if r.Method == "GET" && serverID != "" {
			// 初始连接请求，带有 server_id
			logger.Info("Handling initial connection with server_id: %s", serverID)
			sessionManager.HandleInitialConnection(w, r, serverID)
		} else {
			// 后续会话请求，基于 sessionId 路由
			logger.Info("Handling session request, method: %s, URL: %s", r.Method, r.URL.Path)
			sessionManager.HandleSessionRequest(w, r)
		}
	}

	// 设置路由
	mux := http.NewServeMux()

	// 为所有MCP相关端点添加认证
	// 修改路由格式：/mcp-server/{server_id}/sse
	mux.Handle("/mcp-server/", authMiddleware.Middleware(httpHandler))
	mux.Handle("/messages/", authMiddleware.Middleware(httpHandler))
	mux.Handle("/message", authMiddleware.Middleware(httpHandler))

	// 添加健康检查端点（不需要认证）
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 添加认证信息端点（不需要认证，用于调试）
	mux.HandleFunc("/auth/info", func(w http.ResponseWriter, r *http.Request) {
		if authMiddleware.IsEnabled() {
			w.Write([]byte("Authentication: Enabled\nHeader: " + authMiddleware.GetHeaderName()))
		} else {
			w.Write([]byte("Authentication: Disabled"))
		}
	})

	// 添加会话监控端点（需要认证）
	mux.Handle("/admin/sessions", authMiddleware.Middleware(func(w http.ResponseWriter, r *http.Request) {
		activeCount := sessionManager.GetActiveSessionCount()
		sessionInfo := sessionManager.GetSessionInfo()

		response := map[string]interface{}{
			"active_sessions": activeCount,
			"total_sessions":  len(sessionInfo),
			"sessions":        sessionInfo,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	addr := cfg.Server.GetServerAddr()
	logger.Info("Server starting on %s", addr)

	// 设置优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		sessionManager.Shutdown()
		os.Exit(0)
	}()

	if err = http.ListenAndServe(addr, mux); err != nil {
		logger.Error("Server failed: %v", err)
	}
}
