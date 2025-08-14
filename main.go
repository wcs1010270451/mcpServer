package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"McpServer/internal/database"
	"McpServer/internal/handlers"
	"McpServer/internal/manager"
)

var (
	host = flag.String("host", "0.0.0.0", "host to listen on")
	port = flag.String("port", "9001", "port to listen on")
)

func main() {
	flag.Parse()

	// 创建数据库服务
	db, err := database.NewDatabaseService()
	if err != nil {
		log.Fatalf("Failed to create database service: %v", err)
	}
	defer db.Close()

	// 创建处理器注册表
	handlerRegistry := handlers.NewToolHandlerRegistry()

		// 创建 MCP 服务器管理器
	mcpManager := manager.NewMCPServerManager(db, handlerRegistry)

	// 从数据库加载服务器配置
	if err := mcpManager.LoadServersFromDatabase(); err != nil {
		log.Fatalf("Failed to load servers from database: %v", err)
	}

	// 创建会话管理器
	sessionManager := manager.NewSessionManager(mcpManager, db)

	// 创建 HTTP 处理器
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		serverID := r.URL.Query().Get("server_id")

		if r.Method == "GET" && serverID != "" {
			// 初始连接请求，带有 server_id
			log.Printf("Handling initial connection with server_id: %s", serverID)
			sessionManager.HandleInitialConnection(w, r, serverID)
		} else {
			// 后续会话请求，基于 sessionId 路由
			log.Printf("Handling session request, method: %s, URL: %s", r.Method, r.URL.Path)
			sessionManager.HandleSessionRequest(w, r)
		}
	}

	// 设置路由
	mux := http.NewServeMux()
	mux.Handle("/mcp-server/sse", http.HandlerFunc(httpHandler))
	mux.Handle("/messages/", http.HandlerFunc(httpHandler))
	mux.Handle("/message", http.HandlerFunc(httpHandler))

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
