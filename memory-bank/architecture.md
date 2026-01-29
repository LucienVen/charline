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

### 2025-01-26: 配置文件项目根目录检测

**背景**:
- 从不同目录运行时，SQLite 数据库路径可能不正确
- client 和 server 需要统一的路径解析机制

**变更**:
- 添加 `findProjectRoot()` 函数，通过向上查找 `go.work` 确定项目根目录
- DB_PATH 配置支持完整路径，默认为相对路径
- GetDBConfig() 从完整路径解析 DataDir 和 Name

**配置优先级**:
```
1. DB_PATH 环境变量（完整路径）→ 最高优先级
2. <projectRoot>/server/data/server.db → 默认值
```

**代码示例**:
```go
// 从完整路径解析
func (c *Config) GetDBConfig() DBConfig {
    dir := filepath.Dir(c.DBPath)   // server/data
    name := filepath.Base(c.DBPath) // server.db
    return DBConfig{DataDir: dir, Name: name}
}
```

---

### 2025-01-28: 验证器统一化优化

**背景**:
- invite_controller.go 中存在分散的验证逻辑（`isValidUsername()` 和魔法数字）
- 每个控制器都需要重复实现相同的验证规则
- 验证逻辑与控制器耦合，难以测试和复用
- 缺少结构化的验证错误信息

**新架构**:
```
┌─────────────────────────────────────────────────────────────┐
│                      Server (最新架构)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Controller │  │   Validator  │  │    httputil  │      │
│  │  HTTP 处理   │  │   字段验证   │  │  DTO + tags  │      │
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
| Validator | `server/internal/validator/validator.go` | 验证器核心（自定义规则注册） |
| Validator | `server/internal/validator/errors.go` | 验证错误处理与友好消息 |

**目录结构**:
```
server/internal/
├── validator/                    # 新增：验证器包
│   ├── validator.go              # 验证器核心 + 自定义规则
│   └── errors.go                 # 错误处理 + 友好消息
├── controller/                   # 修改：使用验证器
│   ├── invite_controller.go      # 移除 isValidUsername()
│   └── auth_controller.go
├── httputil/                     # 修改：添加验证标签
│   ├── dto.go                    # 添加 validate tags
│   ├── response.go
│   └── request.go
├── service/
├── store/
├── auth/
├── errors/
├── config/
└── logger/
```

**自定义验证规则**:
```go
// username: 3-20位，字母开头，含字母数字下划线
validateUsername(fl validator.FieldLevel) bool

// ed25519_public_key: Base64 编码，解码后 32 字节
validateEd25519PublicKey(fl validator.FieldLevel) bool
```

**DTO 验证标签**:
```go
type ActivateInviteRequest struct {
    Code      string `json:"code" validate:"required"`
    Username  string `json:"username" validate:"required,username"`
    PublicKey string `json:"public_key" validate:"required,ed25519_public_key"`
}
```

**控制器验证流程**:
```
Before:
  ├─ 解析 JSON
  ├─ 手动验证用户名（isValidUsername）
  ├─ 手动验证公钥长度（魔法数字）
  └─ 调用 Service

After:
  ├─ 解析 JSON
  ├─ 统一验证（validator.Validate）
  │   ├─ username 规则
  │   ├─ ed25519_public_key 规则
  │   └─ required 检查
  ├─ 解析验证错误（友好消息）
  └─ 调用 Service
```

**依赖注入**:
```go
// main.go 初始化
func main() {
    // ... 配置、日志 ...
    
    // 3. 初始化验证器（新增）
    validator.Init()
    log.Info("验证器初始化成功")
    
    // ... 数据库、容器 ...
}
```

**错误处理设计**:
```go
// 验证错误结构
type Error struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Tag     string `json:"tag,omitempty"`
}

// 错误消息示例
{
  "code": 1001,
  "message": "参数错误",
  "data": {
    "validation_errors": [
      {
        "field": "Username",
        "message": "用户名格式无效（3-20位，字母开头，含字母数字下划线）",
        "tag": "username"
      }
    ]
  }
}
```

**优化效果**:

| 维度 | 优化前 | 优化后 | 改进 |
| --- | --- | --- | --- |
| **代码复用** | 每个控制器重复实现 | 统一验证器包 | ⬆️ 100% |
| **可维护性** | 分散在各个控制器 | 集中管理 | ⬆️ 显著提升 |
| **可测试性** | 与控制器耦合 | 独立测试 | ⬆️ 易测试 |
| **声明性** | 命令式验证代码 | struct tag 声明 | ⬆️ 更清晰 |
| **代码行数** | 5-8 行验证逻辑 | 1 行调用 | -80% |
| **错误信息** | 简单错误码 | 结构化详细错误 | ⬆️ 用户体验好 |

**设计原则**:
- **单一职责**: 验证器只负责字段验证
- **声明式**: 通过 struct tag 定义规则
- **不耦合**: 不依赖 Web 框架（纯 validator 库）
- **可扩展**: 支持自定义验证规则

**技术栈**:
- 库: `github.com/go-playground/validator/v10`
- 标准: Go 标准库 + validator 自定义规则

---


### 2025-01-28: Phase 1 代码优化 - httputil 合并 + Recovery 中间件

**背景**:
- httputil 包存在两个功能重复的响应函数（RespondError/RespondWithError），容易混淆
- 缺少 Panic 恢复机制，单个请求 panic 可能导致整体服务崩溃

**新架构**:
```
┌─────────────────────────────────────────────────────────────┐
│                      Server (最新架构)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Controller  │  │   Validator  │  │   httputil   │      │
│  │  HTTP 处理   │  │   字段验证   │  │  响应工具    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Middleware  │  │    Service   │  │    Store     │      │
│  │ Panic+Logger │  │  业务逻辑   │  │   数据访问   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**新增模块**:

| 模块 | 文件 | 职责 |
| --- | --- | --- |
| Middleware | `server/internal/middleware/recovery.go` | Panic 恢复中间件 |

**目录结构**:
```
server/internal/
├── middleware/                   # 新增：中间件包
│   └── recovery.go              # Panic 恢复
├── controller/
├── service/
├── store/
├── validator/
├── httputil/                     # 修改：合并响应函数
│   ├── response.go              # 删除 RespondWithError
│   ├── request.go
│   └── dto.go
├── errors/
├── config/
└── logger/
```

**中间件架构**:
```
HTTP Request
    ↓
┌────────────────────────────────┐
│ Recovery Middleware (最外层)   │ ← 捕获 panic
│ - defer recover()              │
│ - 记录 panic 详情               │
│ - 返回 500 错误                │
└────────────────────────────────┘
    ↓
┌────────────────────────────────┐
│ RequestLogger Middleware       │ ← 记录请求日志
│ - 记录请求信息                 │
│ - 记录响应状态码               │
└────────────────────────────────┘
    ↓
┌────────────────────────────────┐
│ Router & Controller             │ ← 业务处理
└────────────────────────────────┘
```

**httputil 优化**:

**Before:**
```go
// 两个函数功能重复
RespondError(w, bizErr)                   // 自动判断 HTTP 状态
RespondWithError(w, httpStatus, bizErr)    // 手动指定 HTTP 状态（已删除）
```

**After:**
```go
// 统一为一个函数
RespondError(w, bizErr)    // 自动根据 bizErr.Code 判断 HTTP 状态
```

**HTTP 状态码映射规则**:
```go
func getHTTPStatus(code int) int {
    switch {
    case code >= 5000: return http.StatusInternalServerError  // 系统错误
    case code >= 3000: return http.StatusUnauthorized      // 认证错误
    case code >= 2000: return http.StatusBadRequest      // 资源错误
    case code >= 1000: return http.StatusBadRequest      // 参数错误
    default:           return http.StatusOK               // 成功
    }
}
```

**中间件注册顺序**:
```go
r := router.NewRouter(&router.Routes{
    Middlewares: []func(http.Handler) http.Handler{
        recovery.Middleware,              // 1️⃣ 最外层：Panic 恢复
        serverlogger.RequestLogger(log),  // 2️⃣ 内层：请求日志
    },
})
```

**设计原则**:
- **Recovery 优先**: 必须是最外层中间件，确保捕获所有 panic
- **顺序重要**: Recovery → Logger → Router → Controller
- **职责单一**: Recovery 只负责 panic 恢复，不处理业务逻辑

**优化效果**:

| 维度 | 优化前 | 优化后 | 改进 |
| --- | --- | --- | --- |
| **API 一致性** | 2 个相似函数 | 1 个统一函数 | ⬆️ 消除混淆 |
| **代码行数** | 88 行 | 68 行 | -20 行 |
| **服务稳定性** | Panic 导致崩溃 | 自动恢复 | ⬆️ 显著提升 |
| **可调试性** | Panic 难以追踪 | 详细日志 | ⬆️ 易定位问题 |

**技术栈**:
- 标准: Go recover() + defer
- 日志: zap（记录 panic 详情）
- 错误处理: BizError 统一响应

---


### 2025-01-29: Phase 3 架构设计方向

**背景**:
- ChatGPT 5.2 指出当前架构是 "CLI 工具形态"，缺少 IM 系统核心要素
- Phase 0-2.1 完成了基础认证，但需要演进到长连接 IM 架构

**新架构**:
```
┌─────────────────────────────────────────────────────────────┐
│                      IM 长连接架构（Phase 3）                 │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  WebSocket   │  │   Session    │  │  Protocol    │      │
│  │  连接管理    │  │   会话管理   │  │  消息协议    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Heartbeat   │  │   Resume     │  │  Connection  │      │
│  │  心跳保活    │  │   断线恢复   │  │  连接池      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**Token 分层架构**:
```
Layer 1: Identity Token (Ed25519 私钥)
         └─ 永久存储，用于签名证明身份

Layer 2: Refresh Token (JWT, 7天有效)
         └─ 用于获取新的 Session Token

Layer 3: Session Token (内存，连接绑定)
         └─ WebSocket 连接建立时生成

Layer 4: Resume Token (30秒有效)
         └─ 断线时下发，用于快速重连
```

**Session 管理设计**:
```go
type Session struct {
    ID           string
    UserID       int64
    DeviceID     string
    ConnID       string        // WebSocket 连接 ID
    State        SessionState  // ACTIVE, SUSPENDED, CLOSED
    CreatedAt    time.Time
    LastActiveAt time.Time
    ResumeToken  string        // 断线恢复 Token
    ResumeExpiry time.Time
}

type SessionManager interface {
    Create(userID int64, deviceID string, conn *websocket.Conn) (*Session, error)
    Get(sessionID string) (*Session, bool)
    GetByUser(userID int64) []*Session
    Suspend(sessionID string) (resumeToken string, error)
    Resume(resumeToken string, conn *websocket.Conn) (*Session, error)
    Close(sessionID string) error
}
```

**连接生命周期**:

1. **首次连接**:
   ```
   Client → WebSocket Connect → Server
   Client ← Challenge (nonce) ← Server
   Client → Auth (signature) → Server
   Client ← Session Created ← Server
   Client ↔ Message Stream ↔ Server
   ```

2. **心跳保活**:
   ```
   Client → Ping → Server
   Client ← Pong ← Server
   ```

3. **断线恢复**:
   ```
   Client ✕ Connection Lost
   Client → WebSocket Reconnect → Server
   Client → Resume {resume_token} → Server
   Client ← Session Restored ← Server
   ```

**Phase 3 实施规划**:

- **Phase 3.1**: WebSocket 基础连接
  - server/internal/websocket/ (5 个文件)
  - client/internal/websocket/client.go

- **Phase 3.2**: Session 管理层
  - server/internal/session/ (4 个文件)
  - client/internal/session/state.go

- **Phase 3.3**: 心跳与断线检测
  - server/internal/websocket/heartbeat.go
  - client/internal/websocket/keepalive.go

- **Phase 3.4**: 自动重连与恢复
  - server/internal/session/resume.go
  - client/internal/websocket/reconnect.go

**与现有代码关系**:

| 现有模块        | 保留/修改 | 说明                                 |
|-----------------|-----------|--------------------------------------|
| Ed25519 密钥对  | ✅ 保留   | 作为 Identity Token                  |
| Nonce 签名登录  | ✅ 保留   | 用于 WebSocket 连接认证              |
| JWT Token       | ⚡ 调整   | 改为 Refresh Token，不再用于每次操作 |
| /auth/challenge | ⚡ 调整   | 改为 WebSocket 握手阶段调用          |
| /auth/login     | ⚡ 调整   | 改为 WebSocket 认证消息              |

**设计原则**:
- **渐进式演进**: Phase 0-2.1 的认证基础完全保留
- **架构转型**: 从 CLI 工具转向 IM 长连接架构
- **Session 优先**: 连接建立后靠 Session，不再频繁验证 Token
- **断线恢复**: 支持快速重连和消息续传

**参考文档**:
- `thinking-challenge/IM-system-auth-challenge.md` - ChatGPT 5.2 挑战
- `thinking-challenge/response-to-auth-challenge.md` - 架构重整方案

---
