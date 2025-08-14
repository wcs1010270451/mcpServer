# MCP Server

一个基于 Go 语言开发的 Model Context Protocol (MCP) 服务器，支持本地服务、远程 stdio 服务和远程 SSE 服务的统一管理和代理。

## 🚀 特性

- **多协议支持**: 支持本地 MCP 服务、远程 stdio 服务和远程 SSE 服务
- **数据库驱动**: 通过 PostgreSQL 数据库动态配置和管理 MCP 服务
- **智能路由**: 基于 `server_id` 和 `sessionId` 的智能请求路由
- **会话管理**: 支持会话缓存、超时管理和自动清理
- **工具代理**: 无缝代理本地和远程工具调用
- **配置化**: 支持 YAML 配置文件和环境变量
- **可扩展**: 模块化架构，易于扩展新功能

## 📁 项目结构

```
mcpServer/
├── main.go                           # 主程序入口
├── go.mod                           # Go 模块文件
├── go.sum                           # 依赖校验文件
├── .gitignore                       # Git 忽略文件
├── README.md                        # 项目文档
├── config/                          # 配置文件目录
│   └── config.dev.yaml             # 开发环境配置
├── internal/                        # 内部包
│   ├── config/                      # 配置管理
│   │   └── config.go               # 配置结构和加载器
│   ├── database/                    # 数据库层
│   │   ├── database.go             # 数据库服务
│   │   └── migrations/             # 数据库迁移文件
│   │       ├── mcp_service_sse_table.sql
│   │       └── insert_bing_cn_mcp.sql
│   ├── models/                      # 数据模型
│   │   └── models.go               # 数据结构定义
│   ├── handlers/                    # 工具处理器
│   │   └── func.go                 # 内置工具处理器
│   └── manager/                     # 管理器层
│       ├── interfaces.go           # 接口定义
│       ├── mcp_manager.go          # MCP服务管理器
│       ├── remote_stdio.go         # 远程stdio服务管理
│       ├── remote_sse.go           # 远程SSE服务管理
│       └── session_manager.go      # 会话管理器
└── stdioClient.go                   # stdio客户端工具
```

## 🛠️ 安装和构建

### 前置要求

- Go 1.21 或更高版本
- PostgreSQL 数据库
- Git

### 克隆项目

```bash
git clone <your-repo-url>
cd mcpServer
```

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build .
```

## ⚙️ 配置

### 配置文件

创建 `config/config.dev.yaml` 文件：

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 9001

# 数据库配置
database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "your_password"
  database: "mcp_server"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: "5m"

# 日志配置
logging:
  level: "info"
  format: "text"
  file: ""

# 远程服务配置
remote:
  default_timeout: "30s"
  default_connect_timeout: "10s"
  default_retry_attempts: 3
  default_retry_delay: "3s"
  session_cleanup_interval: "30s"
  default_idle_ttl: "5m"

# 工具配置
tools:
  builtin_echo: true
  builtin_greet: true
  builtin_status: true
```

### 环境变量

支持以下环境变量覆盖配置：

```bash
# 数据库配置
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=your_password
export DB_DATABASE=mcp_server

# 服务器配置
export SERVER_HOST=0.0.0.0
export SERVER_PORT=9001

# 日志配置
export LOG_LEVEL=debug
```

## 🗄️ 数据库设置

### 创建数据库表

执行以下 SQL 创建必要的表：

```sql
-- 主服务表
CREATE TABLE "public"."mcp_service" (
    "server_id" varchar(255) PRIMARY KEY,
    "display_name" varchar(255) NOT NULL,
    "implementation_name" varchar(255) NOT NULL,
    "protocol_version" varchar(50) DEFAULT '2025-03-26',
    "enabled" boolean DEFAULT true,
    "metadata" jsonb,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "adapter" varchar(50) DEFAULT 'local',
    "start_mode" varchar(50) DEFAULT 'auto'
);

-- 工具表
CREATE TABLE "public"."mcp_tool" (
    "id" serial PRIMARY KEY,
    "server_id" varchar(255) REFERENCES mcp_service(server_id),
    "tool_name" varchar(255) NOT NULL,
    "description" text,
    "args_schema" jsonb,
    "handler_type" varchar(100) NOT NULL,
    "handler_config" jsonb,
    "enabled" boolean DEFAULT true,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

-- 远程SSE服务配置表
CREATE TABLE "public"."mcp_service_sse" (
    "server_id" varchar(255) PRIMARY KEY REFERENCES mcp_service(server_id),
    "base_url" varchar(500) NOT NULL,
    "sse_path" varchar(255) NOT NULL,
    "auth_type" varchar(50) DEFAULT 'none',
    "auth_config" jsonb DEFAULT '{}',
    "timeout_ms" integer DEFAULT 300000,
    "connect_timeout_ms" integer DEFAULT 30000,
    "retry_attempts" integer DEFAULT 2,
    "retry_delay_ms" integer DEFAULT 3000,
    -- ... 其他字段
);
```

## 🚀 运行

### 基本运行

```bash
./McpServer
```

### 指定配置文件

```bash
./McpServer -config /path/to/config.yaml
```

### 开发模式

```bash
go run . -config config/config.dev.yaml
```

## 📡 API 接口

### SSE 连接

```bash
# 初始连接（指定服务ID）
curl -N -H "Accept: text/event-stream" \
  "http://localhost:9001/mcp-server/sse?server_id=your-server-id"
```

### 工具调用

```bash
# 会话消息（使用从SSE获得的sessionId）
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' \
  "http://localhost:9001/messages/?session_id=your-session-id"
```

## 🔧 服务类型

### 1. 本地服务 (Local)

- 直接在服务器进程中运行的 MCP 服务
- 支持内置工具：echo、greet、status
- 通过数据库配置工具和处理器

### 2. 远程 Stdio 服务 (Remote Stdio)

- 通过命令行启动的外部 MCP 服务
- 支持进程生命周期管理
- 支持会话复用策略

### 3. 远程 SSE 服务 (Remote SSE)

- 通过 HTTP SSE 连接的远程 MCP 服务
- 支持透明代理
- 支持会话路由和缓存

## 🔄 会话管理

系统支持智能会话管理：

1. **初始连接**: 使用 `server_id` 参数建立连接
2. **会话路由**: 自动提取 `sessionId` 并缓存会话信息
3. **后续请求**: 基于 `sessionId` 路由到正确的服务实例
4. **自动清理**: 空闲会话自动清理机制

## 🔍 监控和日志

- 支持结构化日志输出
- 详细的连接和会话管理日志
- 工具调用和错误跟踪
- 性能指标监控

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

如有问题或建议，请创建 Issue 或联系项目维护者。

---

**注意**: 请确保在生产环境中使用前，正确配置数据库连接信息和安全设置。