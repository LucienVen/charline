# 项目进度（Progress）

> 本文件用于记录：已完成事项、当前进行中事项、待办事项。  
> 每完成一个任务或阶段，必须更新本文件。

---

## 已完成 ✅

### Phase 0: 项目基础设施（2025-01-15 完成）

#### 服务端
- [x] 目录结构（server/client 分离）
- [x] Go 模块配置（go.mod 1.25.5）
- [x] HTTP 服务器（go-chi v5）
- [x] 健康检查端点（/health）
- [x] 配置管理系统（环境变量）
- [x] 结构化日志系统（zap）
  - [x] 自定义时间格式（+0800 2025-01-15 17:51:54）
  - [x] 请求日志中间件
  - [x] 请求 ID 追踪
  - [x] 开发环境彩色输出
  - [x] 按状态码分级日志（2xx/INFO, 4xx/WARN, 5xx/ERROR）
- [x] 优雅关闭机制

#### 客户端
- [x] 目录结构
- [x] Go 模块配置
- [x] 命令行入口（main.go）
- [x] hello 命令实现

#### 构建系统
- [x] Makefile（build/server/client/run-*等目标）
- [x] 环境变量模板（.env.example）

#### 文档
- [x] design-document.md - 基础设计想法
- [x] document-navigation.md - 文档导航
- [x] implementation-plan.md - 10阶段实施规划
- [x] usages.md - Makefile 使用说明
- [x] architecture.md - 架构设计记录

---

### 日志抽象为公共库（2025-01-16 完成）

#### 公共库
- [x] `pkg/logger/config.go` - 通用日志配置接口
- [x] `pkg/logger/context.go` - 请求 ID 管理
- [x] `pkg/logger/logger.go` - 核心 Logger
- [x] `pkg/go.mod` - pkg 模块定义
- [x] `go.work` - Go Workspace 配置

#### Server 重构
- [x] `server/internal/logger/logger.go` - 使用 pkg/logger
- [x] `server/internal/logger/middleware.go` - 更新 context API
- [x] `server/internal/logger/context.go` - 删除（移至 pkg）
- [x] `server/go.mod` - 添加 pkg 依赖（local replace）

#### Client 集成
- [x] `client/internal/config/config.go` - 客户端配置
- [x] `client/internal/logger/logger.go` - 日志适配器
- [x] `client/go.mod` - 添加 pkg 依赖（local replace）
- [x] `client/cmd/main.go` - 集成日志系统

#### 验证
- [x] Server 构建成功，日志正常输出
- [x] Client 构建成功，日志正常输出
- [x] 请求 ID 格式一致

---

### HTTP 请求日志级别优化（2025-01-19 完成）

**变更**:
- `server/internal/logger/middleware.go` - 根据状态码动态选择日志级别
  - 2xx/3xx → INFO
  - 4xx → WARN
  - 5xx → ERROR

**效果**:
- 200 OK → INFO（蓝色）
- 404 Not Found → WARN（黄色）
- 500 Internal Server Error → ERROR（红色）

---

### Phase 1: 邀请激活系统（2025-01-19 完成）

**目标**: 服务端提供邀请码激活 API，返回 JWT Token

#### pkg/sqlite 公共库
- [x] `pkg/sqlite/sqlite.go` - 数据库连接管理
- [x] `pkg/sqlite/migrations.go` - 迁移系统（go:embed）
- [x] `pkg/sqlite/migrations/001_init.sql` - 建表 SQL（含注释）
- [x] `pkg/go.mod` - 添加 modernc.org/sqlite v1.34.4

#### server/internal/errors 统一错误码
- [x] `server/internal/errors/codes.go` - 错误码定义
  - ERR_INVITE_xxx: 邀请码相关错误
  - ERR_AUTH_xxx: JWT 认证相关错误
  - ERR_USER_xxx: 用户相关错误

#### server/internal/store 邀请码存储
- [x] `server/internal/store/invite.go` - 邀请码业务逻辑
  - Generate() - 生成 INV-XXXXXXXX 格式邀请码
  - Validate() - 验证邀请码有效性
  - Activate() - 激活邀请码
  - IsUsed() - 检查是否已使用

#### server/internal/auth JWT 认证
- [x] `server/internal/auth/jwt.go` - JWT 生成/验证
  - NewManager() - 创建 JWT 管理器
  - GenerateToken() - 生成 JWT Token
  - ValidateToken() - 验证 Token
  - ParseTokenFromRequest() - 从请求头解析 Token

#### server/internal/api HTTP 处理
- [x] `server/internal/api/handler.go` - API 端点实现
  - GenerateInviteCode() - POST /api/invite/generate
  - ActivateInviteCode() - POST /api/invite/activate
  - ValidateToken() - GET /api/validate

#### 配置更新
- [x] `server/internal/config/config.go` - 添加 JWT_SECRET、DB_PATH
- [x] `server/.env.example` - 新增环境变量模板
- [x] `server/cmd/main.go` - 集成所有模块

#### API 端点
- [x] POST /api/invite/generate - 生成邀请码
- [x] POST /api/invite/activate - 激活邀请码，返回 JWT Token
- [x] GET /api/validate - 验证 Token

#### 依赖
- [x] modernc.org/sqlite v1.34.4 - 纯 Go SQLite 实现
- [x] github.com/golang-jwt/jwt v5.3.0 - JWT 认证

#### 测试验证
```bash
# 生成邀请码
curl -X POST http://localhost:8080/api/invite/generate
# {"code":"INV-XCJGURX8"}

# 激活邀请码
curl -X POST http://localhost:8080/api/invite/activate \
  -H "Content-Type: application/json" \
  -d '{"code":"INV-XCJGURX8","username":"alice"}'
# {"token":"eyJhbG...","version":1}

# 验证 Token
curl http://localhost:8080/api/validate \
  -H "Authorization: Bearer eyJhbG..."
# {"valid":true,"username":"alice"}
```

---

## 当前进行中 🚧

暂无

---

## 待办事项 📋

### Phase 2: 客户端注册与存储
- [ ] client/internal/store/sqlite.go - SQLite 存储（复用 pkg/sqlite）
- [ ] client/internal/commands/register.go - /join 命令
- [ ] 客户端配置管理

### Phase 3: WebSocket 基础通信
- [ ] server/internal/websocket/pool.go - 连接池
- [ ] server/internal/websocket/conn.go - 连接封装
- [ ] server/internal/protocol/message.go - 消息协议
- [ ] client/internal/chat/client.go - WebSocket 客户端

### Phase 4-9: 后续阶段
详见 `implementation-plan.md`

---

## 问题与解决 🔧

### 2025-01-15 日志系统编译错误修复
**问题**: 
- config.go 中未使用的 zap 导入
- context.go 中指针赋值错误
- logger.go 中 sugar.With 类型不匹配

**解决**:
- 移除未使用的导入
- 修改 SetRequestID: `*r = *r.WithContext(ctx)`
- 简化 Logger 结构，移除 sugar 字段

### 2025-01-16 Logger 抽象为公共库
**任务**: 将 server/internal/logger 抽象为 pkg/logger

**变更**:
- 创建 `pkg/logger` 公共库
- 定义 `logger.Config` 接口
- Server 和 Client 分别实现配置适配器
- 使用 `go.work` 管理多模块
- 使用 `replace` 指令引用本地 pkg

### 2025-01-19 HTTP 请求日志级别优化
**问题**: 访问不存在的路径返回 404，但日志仍使用 INFO 级别

**解决**: 根据 HTTP 状态码动态选择日志级别
- 2xx/3xx → INFO
- 4xx → WARN
- 5xx → ERROR

### 2025-01-19 Phase 1 邀请激活系统实现
**技术选型**:
- SQLite: 轻量级，适合 2核2G 服务器
- pkg/sqlite: Server 和 Client 共用
- 统一错误码: ERR_模块_具体错误 格式

**架构分层**:
```
Handler (HTTP 处理)
    ↓
Store (业务逻辑)
    ↓
SQLite (数据访问)
```

**遇到的问题**:
- logger.String/Int 方法不存在 → 改用 zap.String/zap.Int
- main.go 导入路径错误 → 使用 serverlogger 别名
- 配置验证失败 → 确保设置环境变量

---

## 下一步行动

1. Phase 2: 客户端注册与存储
2. 实现 /join 命令
3. 客户端使用 pkg/sqlite 存储用户凭证

---

### 目录结构优化（2025-01-21 完成）

**问题**:
- 根目录 `data/` 有歧义（无法区分是 server 还是 client 的数据）
- SQLite 数据库文件不应上传到远程仓库

**变更**:
- `data/charline.db` → `server/data/server.db`
- 新建 `client/data/` 目录（空，为 Phase 2 准备）
- 更新 `server/internal/config/config.go` 默认路径

**.gitignore 更新**:
```gitignore
# Database files
*.db
*.db-shm
*.db-wal

# Data directories
server/data/
client/data/
```

**配置路径更新**:
- `DBPath` 默认值: `server/data/server.db`
- `GetDBConfig()` 返回: `server/data/server.db`

---

### 错误码和响应结构重构（2025-01-22 完成）

**问题**: 
- 原有 `api/handler.go` 文件过大，职责不清晰
- 错误响应缺少详细信息，用户无法知道具体错误原因
- 缺少 Controller-Service 分层架构

**变更**:

#### 1. 错误码数字化（按语义分段）
```
0     - 成功
1xxx  - 参数/请求错误
2xxx  - 资源相关错误
3xxx  - 认证/授权错误
5xxx  - 系统内部错误
```

#### 2. 统一响应结构
```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

#### 3. BizError 详细错误信息增强
```go
type BizError struct {
    Code    int
    Message string
    Details map[string]interface{} // 详细信息
    cause   error                 // 原始错误（用于日志）
}

// 链式调用方法
func (e *BizError) WithDetail(key string, value interface{}) *BizError
func (e *BizError) WithDetails(details map[string]interface{}) *BizError
func (e *BizError) WrapError(cause error) *BizError
func (e *BizError) GetCause() error
func (e *BizError) GetDetailedMsg() string
```

#### 4. 目录结构重构
```
server/internal/
├── controller/           # 新增：Controller 层（原 api）
│   ├── invite_controller.go
│   ├── auth_controller.go
│   └── response.go
├── service/              # 新增：Service 层（业务逻辑）
│   ├── invite_service.go
│   └── auth_service.go
├── store/                # 保留：数据访问层
│   └── invite.go         # 修改：返回 *errors.BizError
├── auth/                 # 保留：JWT 工具层
│   └── jwt.go            # 修改：返回 *errors.BizError
└── errors/               # 扩展：错误码系统
    └── codes.go          # 修改：新增 Details 字段和链式方法
```

#### 5. 分层架构
```
Controller (controller/)  →  请求/响应解析，调用 Service
         ↓
Service (service/)        →  业务逻辑编排，调用 Store/Auth
         ↓
Store/Auth (store/auth/)  →  数据访问、JWT 处理
```

#### 6. 响应格式示例

**无详细信息时（现有）**：
```json
{
    "code": 1003,
    "message": "用户名格式无效"
}
```

**有详细信息时（新）**：
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

#### 7. 关键文件变更
| 文件 | 操作 |
| --- | --- |
| server/internal/errors/codes.go | 扩展 BizError 结构，新增链式方法 |
| server/internal/controller/response.go | 统一响应结构，处理 Details |
| server/internal/controller/invite_controller.go | 新建，从 handler.go 迁移 |
| server/internal/controller/auth_controller.go | 新建，从 handler.go 迁移 |
| server/internal/service/invite_service.go | 新建，业务逻辑层 |
| server/internal/service/auth_service.go | 新建，认证业务逻辑 |
| server/internal/store/invite.go | 修改，返回 *errors.BizError |
| server/internal/auth/jwt.go | 修改，返回 *errors.BizError |
| server/cmd/main.go | 修改，集成新架构 |
| server/internal/api/ | 删除，重命名为 controller |

#### 8. 使用示例
```go
// Controller 层手动添加详细信息
if len(req.Username) < 3 {
    bizErr := apperrors.ErrInvalidUsername.
        WithDetail("reason", "用户名长度必须在 3-20 字符之间").
        WithDetail("field", "username").
        WithDetail("provided", req.Username).
        WithDetail("length", len(req.Username))
    RespondError(w, bizErr)
    return
}

// Service 层包装底层错误（用于日志）
if err := db.Exec(...); err != nil {
    return "", 0, errors.ErrSystemError.
        WrapError(err).
        WithDetail("operation", "invite_activate")
}
```

#### 9. 向后兼容性
- 现有预定义错误无需修改，继续正常工作
- 无 Details 时响应格式保持不变
- 所有现有 API 端点路径不变

---

### 依赖注入容器 + 路由模块化（2025-01-26 完成）

**变更内容**:

#### 1. 新增 container 依赖注入容器
- [x] `server/internal/container/container.go`
  - `NewContainer()` - 组装所有依赖
  - 返回包含所有 Controllers 的容器结构

#### 2. 新增 router 路由模块
- [x] `server/internal/router/router.go`
  - `NewRouter()` - 创建并配置路由
  - 支持路由分组：`/api/v1/*`
  - 支持全局中间件配置
  - 开发环境自动打印路由列表

#### 3. 配置管理改进
- [x] `server/internal/config/config.go` - 添加 godotenv 加载
- [x] `client/internal/config/config.go` - 添加 godotenv 加载
- [x] 多路径查找：当前目录 `.env` → 项目根目录 `.env`

#### 4. main.go 简化
- [x] 从 118 行简化至 91 行
- [x] 使用 container.NewContainer() 组装依赖
- [x] 使用 router.NewRouter() 构建路由

#### 5. 依赖更新
- [x] `github.com/joho/godotenv v1.5.1` - .env 文件加载

#### 6. API 路由变更
| 旧路由 | 新路由 |
| --- | --- |
| `/api/invite/generate` | `/api/v1/invite/generate` |
| `/api/invite/activate` | `/api/v1/invite/activate` |
| `/api/validate` | `/api/v1/validate` |
| `/health` | `/health`（不变） |

#### 7. 目录结构变更
```
server/internal/
├── container/           # 新增
│   └── container.go
├── router/              # 新增
│   └── router.go
├── controller/          # 保留
├── service/             # 保留
├── store/               # 保留
├── auth/                # 保留
├── errors/              # 保留
├── config/              # 修改（支持 .env）
└── logger/              # 保留
```

#### 8. 架构改进
**职责分离**:
- `container`: 依赖组装
- `router`: 路由配置
- `main`: 启动流程编排

**依赖关系**:
```
main.go (91行)
    ├─ config.Load()
    ├─ logger.New()
    ├─ sqlite.New()
    ├─ container.NewContainer()  ← 新增
    │   ├─ jwtManager
    │   ├─ store
    │   ├─ services
    │   └─ controllers
    └─ router.NewRouter()         ← 更新
        └─ Controllers
            ├─ Invite
            └─ Auth
```

#### 9. 测试验证
```bash
# 服务启动成功，路由打印正常
=== Registered Routes ===
  POST   /api/v1/invite/activate
  POST   /api/v1/invite/generate
  GET    /api/v1/validate
  GET    /health
```

---
