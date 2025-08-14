package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolHandler 定义工具处理器的接口
type ToolHandler func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error)

// DatabaseService 数据库服务接口（避免循环依赖）
type DatabaseService interface {
	GetEmployeeByName(name string) (*Employee, error)
	GetAllEmployees() ([]Employee, error)
}

// Employee 员工模型（避免循环依赖）
type Employee struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Enabled bool   `json:"enabled"`
}

// ToolHandlerRegistry 工具处理器注册表
type ToolHandlerRegistry struct {
	handlers map[string]ToolHandler
	db       DatabaseService
}

// NewToolHandlerRegistry 创建新的工具处理器注册表
func NewToolHandlerRegistry(db DatabaseService) *ToolHandlerRegistry {
	registry := &ToolHandlerRegistry{
		handlers: make(map[string]ToolHandler),
		db:       db,
	}

	// 注册内置处理器
	registry.RegisterBuiltinHandlers()

	return registry
}

// NewToolHandlerRegistryWithoutDB 创建不依赖数据库的工具处理器注册表
func NewToolHandlerRegistryWithoutDB() *ToolHandlerRegistry {
	registry := &ToolHandlerRegistry{
		handlers: make(map[string]ToolHandler),
		db:       nil,
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
	// Echo 处理器 - 回显输入的文本
	r.RegisterHandler("builtin_echo", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: No arguments provided"},
				},
				IsError: true,
			}, nil
		}

		// 获取要回显的文本
		text, exists := args["text"].(string)
		if !exists || text == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: 'text' parameter is required"},
				},
				IsError: true,
			}, nil
		}

		// 可选的前缀参数
		prefix, _ := args["prefix"].(string)
		if prefix != "" {
			text = prefix + " " + text
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
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

	// Say Hi 处理器 - 专门的打招呼工具
	r.RegisterHandler("builtin_say_hi", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		name, _ := args["name"].(string)
		if name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hi there!"},
				},
			}, nil
		}

		message := fmt.Sprintf("Hi %s!", name)
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

	// 员工查询处理器
	r.RegisterHandler("builtin_employee_query", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: No arguments provided"},
				},
				IsError: true,
			}, nil
		}

		// 获取员工姓名参数
		name, exists := args["name"].(string)
		if !exists || name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: 'name' parameter is required"},
				},
				IsError: true,
			}, nil
		}

		// 查询员工信息
		if r.db == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: Database service not available"},
				},
				IsError: true,
			}, nil
		}

		employee, err := r.db.GetEmployeeByName(name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error querying employee: %v", err)},
				},
				IsError: true,
			}, nil
		}

		if employee == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Employee '%s' not found", name)},
				},
			}, nil
		}

		// 格式化员工信息
		message := fmt.Sprintf("员工信息:\n姓名: %s\n地址: %s\n电话: %s",
			employee.Name, employee.Address, employee.Phone)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	})

	// 员工地址查询处理器
	r.RegisterHandler("builtin_employee_address", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: No arguments provided"},
				},
				IsError: true,
			}, nil
		}

		// 获取员工姓名参数
		name, exists := args["name"].(string)
		if !exists || name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: 'name' parameter is required"},
				},
				IsError: true,
			}, nil
		}

		// 查询员工信息
		if r.db == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: Database service not available"},
				},
				IsError: true,
			}, nil
		}

		employee, err := r.db.GetEmployeeByName(name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error querying employee: %v", err)},
				},
				IsError: true,
			}, nil
		}

		if employee == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Employee '%s' not found", name)},
				},
			}, nil
		}

		// 只返回地址信息
		message := fmt.Sprintf("%s 的地址: %s", employee.Name, employee.Address)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil
	})

	// 员工电话查询处理器
	r.RegisterHandler("builtin_employee_phone", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// 类型断言获取参数
		args, ok := params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: No arguments provided"},
				},
				IsError: true,
			}, nil
		}

		// 获取员工姓名参数
		name, exists := args["name"].(string)
		if !exists || name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: 'name' parameter is required"},
				},
				IsError: true,
			}, nil
		}

		// 查询员工信息
		if r.db == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: Database service not available"},
				},
				IsError: true,
			}, nil
		}

		employee, err := r.db.GetEmployeeByName(name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error querying employee: %v", err)},
				},
				IsError: true,
			}, nil
		}

		if employee == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Employee '%s' not found", name)},
				},
			}, nil
		}

		// 只返回电话信息
		message := fmt.Sprintf("%s 的电话: %s", employee.Name, employee.Phone)

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
