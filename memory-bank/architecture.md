# 架构设计（Architecture）

> 本文件用于记录：系统架构设计、模块划分、关键流程、重要决策。  
> 每次完成主要功能或里程碑后，必须更新本文件。

---

## 系统架构

### 整体架构
```
┌─────────────────────────────────────────────────────────────┐
│                         Client                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Commands   │  │     Chat     │  │    Store     │      │
│  │   /join等    │  │   WebSocket  │  │   SQLite     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                        WebSocket / HTTP
                            │
┌─────────────────────────────────────────────────────────────┐
│                         Server                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │     API      │  │   WebSocket  │  │    Auth      │      │
│  │  HTTP Handler│  │     Pool     │  │     JWT      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Config    │  │    Logger    │  │    Store     │      │
│  │  环境变量配置 │  │  结构化日志  │  │ 邀请码/离线  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                    ┌───────┴───────┐
                    │   pkg/        │
                    │   logger/     │
                    │  公共日志库    │
                    └───────────────┘
```

---

## 已完成模块

### Phase 0: 基础设施（已完成）

#### 公共库

**1. 日志系统** (`pkg/logger/`)
- 职责：通用结构化日志库，server 和 client 共用
- 文件结构：
  - `logger.go`: zap 日志封装，自定义时间格式
  - `context.go`: 请求 ID 上下文管理（通用）
  - `config.go`: 通用日志配置接口
- 日志格式：`+0800 2025-01-15 17:51:54 INFO cmd/main.go:35 | 消息内容 {字段}`
- 特性：
  - 时区支持（+0800）
  - 彩色输出（开发环境）
  - JSON 输出（生产环境）
  - 请求 ID 追踪
  - 调用堆栈（Error 级别）

#### 服务端模块

**2. 配置管理** (`server/internal/config/config.go`)
- 职责：环境变量加载与验证
- 支持配置：
  - ENV: development | production
  - LOG_LEVEL: debug | info | warn | error
  - LOG_FORMAT: console | json
  - SERVER_PORT: 1-65535
- 关键方法：
  - `Load()`: 加载配置
  - `Validate()`: 验证配置有效性
  - `GetZapLevel()`: 获取日志级别
  - `GetAddress()`: 获取监听地址

**3. 日志适配器** (`server/internal/logger/`)
- 职责：将 server.Config 适配为 logger.Config
- 文件结构：
  - `logger.go`: 配置适配器，调用 pkg/logger
  - `middleware.go`: HTTP 请求日志中间件（server 特有）

**4. HTTP 服务** (`server/cmd/main.go`)
- 职责：服务入口与路由管理
- 框架：go-chi v5
- 端点：
  - GET /health: 健康检查
- 特性：
  - 优雅关闭（5秒超时）
  - 请求日志中间件
  - 信号处理（SIGINT/SIGTERM）

#### 客户端模块

**5. 配置管理** (`client/internal/config/config.go`)
- 职责：环境变量加载与验证
- 支持配置：
  - ENV: development | production
  - LOG_LEVEL: debug | info | warn | error
  - LOG_FORMAT: console | json
- 关键方法：
  - `Load()`: 加载配置
  - `Validate()`: 验证配置有效性
  - `GetZapLevel()`: 获取日志级别

**6. 日志适配器** (`client/internal/logger/`)
- 职责：将 client.Config 适配为 logger.Config
- 文件：
  - `logger.go`: 配置适配器，调用 pkg/logger

**7. 基础框架** (`client/cmd/main.go`)
- 职责：命令行入口
- 支持命令：hello
- 已集成日志系统

#### 构建系统

**Makefile**
- `make deps`: 拉取依赖
- `make build`: 构建所有
- `make server`: 构建服务端
- `make client`: 构建客户端
- `make run-server`: 运行服务端
- `make run-client`: 运行客户端
- `make clean`: 清理构建产物
- `make test`: 运行测试
- `make lint`: 代码检查

**Go Workspace** (`go.work`)
- 多模块工作区支持
- 模块列表：pkg, server, client

---

## 技术栈

| 功能 | 库 | 版本 | 状态 |
| --- | --- | --- | --- |
| 日志 | uber-go/zap | v1.27.1 | ✅ 已集成 |
| Web 框架 | go-chi/chi/v5 | v5.2.4 | ✅ 已集成 |
| 配置 | - | - | ✅ 已实现 |
| WebSocket | gorilla/websocket | - | ⏳ 待集成 |
| JWT | golang-jwt/jwt | v5.3.0 | ✅ 已集成 |
| SQLite | modernc.org/sqlite | v1.34.4 | ✅ 已集成 |

---

## 设计决策

### 日志抽象决策
- **选择**: 将 logger 抽象为 `pkg/logger` 公共库
- **理由**:
  - server 和 client 共享核心日志功能
  - 减少代码重复
  - 统一日志格式和请求 ID 生成
  - 便于未来扩展到其他组件

### 日志格式决策
- **选择**: 自定义格式 `+0800 2025-01-15 17:51:54 INFO file.go(line) | msg {fields}`
- **理由**:
  - 时区前置便于调试
  - 文件位置便于定位问题
  - 结构化字段便于解析
  - 开发环境彩色输出提高可读性

### 框架选择
- **选择**: go-chi 而非 gin / echo
- **理由**:
  - 轻量级，无过度封装
  - 兼容 net/http，易于扩展
  - 中间件机制灵活

### 配置管理
- **选择**: 环境变量 + 默认值
- **理由**:
  - 12-Factor App 最佳实践
  - 容器友好
  - 敏感信息不进代码库

### Go Workspace
- **选择**: 使用 go.work 管理多模块
- **理由**:
  - 本地开发无需推送远程仓库
  - 方便 pkg 模块被 server/client 引用
  - 统一依赖版本管理

---

## 待实现模块

### Phase 1: 邀请激活系统（下一步）
- `server/internal/store/invite.go`: 邀请码存储
- `server/internal/auth/jwt.go`: JWT 认证
- `server/internal/api/handler.go`: API 处理器

### Phase 2: 客户端注册
- `client/internal/store/sqlite.go`: SQLite 存储
- `client/internal/commands/register.go`: 注册命令

### Phase 3: WebSocket 通信
- `server/internal/websocket/pool.go`: 连接池
- `server/internal/websocket/conn.go`: 连接封装
- `server/internal/protocol/message.go`: 消息协议
- `client/internal/chat/client.go`: WebSocket 客户端

---

## 里程碑

- [x] **2025-01-15**: Phase 0 完成 - 基础设施搭建
  - 服务端 HTTP 服务
  - 日志系统
  - 配置管理
  - 构建脚本
- [x] **2025-01-16**: 日志抽象为公共库
  - 创建 `pkg/logger` 公共库
  - Server 和 Client 共用日志功能
  - Go Workspace 多模块管理
- [ ] **进行中**: Phase 1 - 邀请激活系统
- [ ] **待定**: Phase 2 - 客户端注册
- [ ] **待定**: Phase 3 - WebSocket 通信

### 数据存储决策
- **选择**: Server 和 Client 各自独立数据目录
- **结构**:
  - `server/data/server.db` - 服务端邀请码、用户数据
  - `client/data/client.db` - 客户端凭证、聊天历史
- **理由**:
  - 架构清晰，server/client 是独立进程
  - 数据隔离，互不干扰
  - 便于部署和备份
  - 数据库文件不入版本控制

---

## 架构演进记录

### 2025-01-22: Controller-Service-Store 三层架构重构

**背景**: 
- 原 `api/handler.go` 职责过重，包含 HTTP 处理、业务逻辑、数据访问
- 错误响应缺少详细信息，用户体验不佳
- 缺少清晰的分层架构

**新架构**:
```
┌─────────────────────────────────────────────────────────────┐
│                      Server (重构后)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Controller  │  │   Service    │  │    Store     │      │
│  │  HTTP 处理   │  │   业务逻辑   │  │   数据访问   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    errors/   │  │    auth/     │  │    Config    │      │
│  │  错误码系统  │  │    JWT       │  │   日志配置   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**模块职责**:

| 层级 | 模块 | 职责 |
| --- | --- | --- |
| Controller | `controller/` | 请求/响应解析，参数验证，调用 Service |
| Service | `service/` | 业务逻辑编排，事务管理，调用 Store/Auth |
| Store | `store/` | 数据访问（SQLite），CRUD 操作 |
| Auth | `auth/` | JWT 生成/验证，Token 管理 |
| Errors | `errors/` | 错误码定义，BizError 结构，链式调用 |

**错误处理设计**:
```go
// BizError 结构
type BizError struct {
    Code    int                    // 数字化错误码
    Message string                 // 用户友好消息
    Details map[string]interface{} // 详细错误信息（可选）
    cause   error                  // 底层错误（用于日志）
}

// 链式调用示例
ErrInvalidUsername.
    WithDetail("reason", "用户名长度必须在 3-20 字符之间").
    WithDetail("field", "username").
    WithDetail("provided", "ab")
```

**响应格式**:
```json
{
    "code": 1003,
    "message": "用户名格式无效",
    "data": {
        "reason": "用户名长度必须在 3-20 字符之间",
        "field": "username",
        "provided": "ab",
        "length": 2,
        "min_length": 3,
        "max_length": 20
    }
}
```

**目录结构变更**:
```
server/internal/
├── controller/           # 新增：HTTP 处理层
│   ├── invite_controller.go
│   ├── auth_controller.go
│   └── response.go       # 统一响应结构
├── service/              # 新增：业务逻辑层
│   ├── invite_service.go
│   └── auth_service.go
├── store/                # 修改：数据访问层
│   └── invite.go         # 返回 *errors.BizError
├── auth/                 # 修改：JWT 工具层
│   └── jwt.go            # 返回 *errors.BizError
├── errors/               # 扩展：错误码系统
│   └── codes.go          # 新增 Details、链式方法
├── config/               # 保留：配置管理
├── logger/               # 保留：日志适配器
└── api/                  # 删除：已重命名为 controller
```

**错误码分段**:
```
0     - 成功
1xxx  - 参数/请求错误（1001: 无效参数, 1002: JSON错误, 1003: 用户名无效）
2xxx  - 资源相关错误（2001: 邀请码不存在, 2002: 邀请码已使用）
3xxx  - 认证/授权错误（3001: 未授权, 3002: Token过期）
5xxx  - 系统内部错误（5000: 系统错误）
```

**向后兼容性**:
- 预定义错误无需修改，继续正常工作
- 无 Details 时响应格式保持不变
- API 端点路径不变

---

### 2025-01-26: 依赖注入容器 + 路由模块化

**背景**:
- main.go 中直接初始化 service/controller，代码耦合度高
- 路由注册逻辑混在 main.go 中，不易维护
- 缺少依赖注入容器，难以扩展

**新架构**:
```
┌─────────────────────────────────────────────────────────────┐
│                      Server (最新架构)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Container  │  │    Router    │  │  Controller  │      │
│  │  依赖注入容器 │  │  路由分组    │  │  HTTP 处理   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Service    │  │    Store     │  │    Auth      │      │
│  │  业务逻辑层  │  │   数据访问   │  │    JWT       │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**新增模块**:

| 模块 | 文件 | 职责 |
| --- | --- | --- |
| Container | `container/container.go` | 依赖注入容器，组装所有依赖 |
| Router | `router/router.go` | 路由注册、分组、中间件管理 |

**目录结构**:
```
server/internal/
├── container/                    # 新增：依赖注入容器
│   └── container.go              # NewContainer() 返回所有 controllers
├── router/                       # 新增：路由模块
│   └── router.go                 # NewRouter() 支持路由分组、中间件
├── controller/                   # 保留：HTTP 处理层
├── service/                      # 保留：业务逻辑层
├── store/                        # 保留：数据访问层
├── auth/                         # 保留：JWT 工具层
├── errors/                       # 保留：错误码系统
├── config/                       # 保留：配置管理（支持 .env）
└── logger/                       # 保留：日志适配器
```

**路由分组设计**:
```
/health                          → 健康检查（无认证）
/api/v1/invite/generate         → POST 生成邀请码
/api/v1/invite/activate         → POST 激活邀请码
/api/v1/validate                → GET  验证 Token
```

**中间件架构**:
```go
type Routes struct {
    Controllers *Controllers
    Logger      *logger.Logger
    Middlewares []func(http.Handler) http.Handler  // 全局中间件切片
}

// 使用示例
r := router.NewRouter(&router.Routes{
    Controllers: &router.Controllers{
        Invite: container.InviteCtrl,
        Auth:   container.AuthCtrl,
    },
    Logger: log,
    Middlewares: []func(http.Handler) http.Handler{
        serverlogger.RequestLogger(log),  // 可添加多个
        // middleware.Recoverer,
        // middleware.Timeout(60s),
    },
})
```

**配置管理改进**:
- 添加 `godotenv` 支持，自动加载 `.env` 文件
- 支持多路径查找：当前目录 `.env` → 项目根目录 `.env`
- Server 和 Client 的 config.go 都已更新

**main.go 简化**:
```go
// 重构前：118 行，直接初始化所有依赖
jwtManager := auth.NewManager(...)
inviteStore := store.NewInviteStore(...)
inviteService := service.NewInviteService(...)
authService := service.NewAuthService(...)
inviteCtrl := controller.NewInviteController(...)
authCtrl := controller.NewAuthController(...)

// 重构后：91 行，通过容器组装
container, err := container.NewContainer(cfg, db, log)
r := router.NewRouter(&router.Routes{
    Controllers: &router.Controllers{...},
    Logger: log,
    Middlewares: [...],
})
```

**开发体验改进**:
- 开发环境启动时自动打印注册的路由列表
```
=== Registered Routes ===
  POST   /api/v1/invite/activate
  POST   /api/v1/invite/generate
  GET    /api/v1/validate
  GET    /health
```

**依赖关系**:
```
main.go
    ↓
container.NewContainer()
    ├─ auth.NewManager()
    ├─ store.NewInviteStore()
    ├─ service.NewInviteService()
    ├─ service.NewAuthService()
    ├─ controller.NewInviteController()
    └─ controller.NewAuthController()
    ↓
router.NewRouter()
    └─ 使用 Controllers 构建路由
```

---
