package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolHandler 定义工具处理器的接口
type ToolHandler func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error)

// ToolHandlerRegistry 工具处理器注册表
type ToolHandlerRegistry struct {
	handlers map[string]ToolHandler
}

// NewToolHandlerRegistry 创建新的工具处理器注册表
func NewToolHandlerRegistry() *ToolHandlerRegistry {
	registry := &ToolHandlerRegistry{
		handlers: make(map[string]ToolHandler),
	}

	// 注册内置处理器
	registry.RegisterBuiltinHandlers()

	return registry
}

// RegisterHandler 注册处理器
func (r *ToolHandlerRegistry) RegisterHandler(handlerType string, handler ToolHandler) {
	r.handlers[handlerType] = handler
}

// GetHandler 获取处理器
func (r *ToolHandlerRegistry) GetHandler(handlerType string) (ToolHandler, bool) {
	handler, exists := r.handlers[handlerType]
	return handler, exists
}

// RegisterBuiltinHandlers 注册内置处理器
func (r *ToolHandlerRegistry) RegisterBuiltinHandlers() {
	// Echo 处理器 - 简单返回输入参数
	r.RegisterHandler("builtin_echo", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		message := fmt.Sprintf("Echo from tool '%s' with args: %v", params.Name, params.Arguments)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	})

	// Greet 处理器 - 问候处理器
	r.RegisterHandler("builtin_greet", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		name, _ := args["name"].(string)
		if name == "" {
			name = "World"
		}

		greeting, _ := args["greeting"].(string)
		if greeting == "" {
			greeting = "Hello"
		}

		message := fmt.Sprintf("%s, %s!", greeting, name)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	})

	// Status 处理器 - 返回系统状态
	r.RegisterHandler("builtin_status", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		message := "System is running normally"
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	})

	// 保留原有的处理器以兼容现有数据
	r.RegisterHandler("say_hi", r.createLegacyHandler("Hi"))
	r.RegisterHandler("say_hello", r.createLegacyHandler("Hello"))
	r.RegisterHandler("say_notfond", r.createLegacyHandler("NotFond"))
}

// createLegacyHandler 创建兼容旧版本的处理器
func (r *ToolHandlerRegistry) createLegacyHandler(prefix string) ToolHandler {
	return func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		name, _ := args["name"].(string)
		if name == "" {
			name = "Anonymous"
		}

		message := fmt.Sprintf("%s %s", prefix, name)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	}
}
