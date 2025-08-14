# MCP Server with Database Configuration

这是一个基于 Go 语言的 MCP (Model Context Protocol) 服务器，支持从 PostgreSQL 数据库动态加载服务配置。

## 功能特性

- 📊 **数据库驱动**: 从 PostgreSQL 数据库动态加载 MCP 服务和工具配置
- 🚀 **动态路由**: 支持通过 `server_id` 访问不同的 MCP 服务
- 🔧 **可扩展处理器**: 插件化的工具处理器系统
- 📡 **SSE 协议**: 使用 Server-Sent Events 协议提供 HTTP 接口
- ⚙️ **环境配置**: 支持环境变量配置数据库连接

## 项目结构

```
mcpServer/
├── main.go          # 原始版本 (已弃用)
├── main_new.go      # 新版本主程序
├── models.go        # 数据模型定义
├── database.go      # 数据库服务层
├── func.go          # 工具处理器注册表
├── stdioClient.go   # 客户端示例
├── sample_data.sql  # 示例数据
├── go.mod           # Go 模块定义
└── README.md        # 项目文档
```

## 数据库设计

### mcp_service 表
存储 MCP 服务的基本配置信息。

### mcp_tool 表
存储每个服务的工具配置，包括参数模式和处理器类型。

## 安装和运行

### 1. 准备数据库

首先创建 PostgreSQL 数据库并执行表结构创建 SQL：

```sql
-- 创建数据库表 (已在您的环境中完成)
-- 执行 sample_data.sql 插入示例数据
\i sample_data.sql
```

### 2. 配置环境变量

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=wcs
export DB_PASSWORD=your_password
export DB_NAME=postgres
export DB_SSLMODE=disable
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 运行服务器

```bash
# 使用新版本
go run main_new.go models.go database.go func.go

# 或者指定主机和端口
go run main_new.go models.go database.go func.go -host 0.0.0.0 -port 8080
```

## 使用方式

### 访问服务

服务器支持多种访问方式：

1. **通过 server_id 路径**:
   ```
   http://localhost:8080/greeter1
   http://localhost:8080/echo_service
   ```

2. **通过查询参数**:
   ```
   http://localhost:8080/?server_id=greeter1
   http://localhost:8080/any-path?server_id=echo_service
   ```

### 可用服务和工具

根据示例数据，包含以下服务：

- **greeter1**: 包含 `greet1`, `custom_greet` 工具
- **greeter2**: 包含 `greet2`, `custom_greet` 工具  
- **greeter3**: 包含 `greet3` 工具
- **echo_service**: 包含 `echo` 工具
- **status_service**: 包含 `get_status` 工具

### 工具处理器类型

内置的处理器类型：

- `builtin_echo`: 回显输入参数
- `builtin_greet`: 可配置的问候处理器
- `builtin_status`: 系统状态处理器
- `say_hi`, `say_hello`, `say_notfond`: 兼容旧版本的处理器

## 扩展开发

### 添加新的处理器

在 `func.go` 的 `RegisterBuiltinHandlers` 方法中添加：

```go
r.RegisterHandler("your_handler_type", func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
    // 您的处理逻辑
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: "Your response"},
        },
    }, nil
})
```

### 添加新服务

直接在数据库中插入新的服务和工具配置：

```sql
-- 添加新服务
INSERT INTO mcp_service (server_id, display_name, implementation_name, ...) VALUES (...);

-- 添加对应的工具
INSERT INTO mcp_tool (server_id, name, description, handler_type, ...) VALUES (...);
```

重启服务器后，新配置将自动生效。

## 开发说明

- `main.go` 是原始的硬编码版本，已被 `main_new.go` 替代
- 新版本完全基于数据库配置，无需修改代码即可添加新服务
- 支持热重载：重启服务器后会自动加载数据库中的最新配置
- 兼容原有的路径访问方式，便于平滑迁移

## 故障排除

1. **数据库连接失败**: 检查环境变量配置和数据库连接权限
2. **服务未找到**: 确认数据库中对应的 `server_id` 且 `enabled=true`
3. **工具调用失败**: 检查工具的 `handler_type` 是否已注册
