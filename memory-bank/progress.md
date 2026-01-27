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

### Phase 2: 客户端注册与存储（2025-01-27 完成）

**目标**: 实现客户端 `/join` 命令，支持 Ed25519 密钥对生成、JWT Token 存储、凭证安全管理

#### 客户端新建文件

- [x] `client/internal/store/paths.go` - `~/.charline/` 路径管理
  - `GetCharlineDir()` - 返回用户数据目录
  - `EnsurePrivateDir()` - 确保目录存在且权限正确 (700)
  - `KeyPath()` / `KeyPubPath()` / `TokenPath()` - 密钥和 Token 路径

- [x] `client/internal/auth/keypair.go` - Ed25519 密钥对管理
  - `Generate()` - 生成 Ed25519 密钥对
  - `Load()` - 从 `~/.charline/id_ed25519` 加载
  - `Save()` - 保存密钥对（PEM 格式，私钥 600 权限）
  - `PubKeyBase64()` - 返回 Base64 编码公钥

- [x] `client/internal/store/credential.go` - 凭证存储
  - `SaveToken()` / `LoadToken()` - JWT Token 读写
  - `SaveCredential()` / `LoadCredential()` - 完整凭证管理
  - `HasCredential()` - 检查是否已有凭证

- [x] `client/internal/auth/signer.go` - Nonce 签名器
  - `NewSigner()` - 创建签名器
  - `Sign()` - 对 nonce 进行 Ed25519 签名（Base64 输出）

- [x] `client/internal/commands/join.go` - Join 命令实现
  - `Join()` - 执行完整 join 流程
  - 生成密钥对 → 调用 API → 保存凭证

- [x] `client/cmd/main.go` - 添加 join 命令处理
  - `handleJoin()` - 命令参数解析和执行

#### 服务端修改文件

- [x] `pkg/sqlite/migrations/002_add_public_key.sql` - 数据库迁移
  - `ALTER TABLE users ADD COLUMN public_key TEXT NOT NULL`
  - `CREATE UNIQUE INDEX idx_users_public_key ON users(public_key)`

- [x] `server/internal/controller/invite_controller.go` - 添加 public_key 支持
  - `ActivateInviteCodeRequest` 添加 `PublicKey` 字段
  - 公钥格式验证（40-50 字符 Base64）

- [x] `server/internal/service/invite_service.go` - 激活逻辑更新
  - `Activate()` 新增 `publicKey` 参数

- [x] `server/internal/store/invite.go` - 完整激活事务
  - 创建用户（绑定 public_key）
  - 创建用户-邀请码关联
  - 更新邀请码状态

#### 验证步骤

```bash
# 1. 启动服务端
make run-server

# 2. 生成邀请码
curl -X POST http://localhost:8080/api/v1/invite/generate
# {"code":"INV-XXXXXXXX"}

# 3. 客户端执行 join
./bin/charline join INV-XXXXXXXX alice
# ✓ Join 成功！
#   用户名: alice
#   凭证版本: 1

# 4. 验证文件
ls -la ~/.charline/
# -rw-------  id_ed25519      (私钥, 600)
# -rw-r--r--  id_ed25519.pub  (公钥)
# -rw-r--r--  token.jwt       (JWT Token)

# 5. 验证服务端用户
sqlite3 server/data/server.db "SELECT username, public_key FROM users;"
# alice    | Base64EncodedPublicKey...
```

#### 安全特性

- 私钥文件权限 `600`（仅用户可读写）
- 目录权限 `700`（仅用户可访问）
- 私钥永不网络传输，仅传输公钥

---

## 当前进行中 🚧

### Phase 2.1: Nonce 签名登录

**目标**: 实现 `/auth login` 命令，基于 Ed25519 PoP（Proof of Possession）认证

**参考文档**: `memory-bank/phase2-1-login-auth.md`

#### 核心流程

```
1. 客户端读取本地私钥和 token
2. GET /api/v1/auth/challenge (携带旧 token)
3. 服务端验证 token，生成随机 nonce (5-30秒有效)
4. 客户端用私钥签名 nonce
5. POST /api/v1/auth/login (nonce + signature)
6. 服务端验签，更新 token 版本
7. 返回新 token
```

#### 关键文件

| 客户端新增 | 描述 |
| --- | --- |
| `client/internal/commands/login.go` | `/auth login` 命令入口 |
| `client/internal/store/token.go` | Token 管理（加载/保存） |

| 服务端新增 | 描述 |
| --- | --- |
| `server/internal/controller/auth_controller.go` | /challenge 和 /login 端点 |
| `server/internal/service/auth_service.go` | 认证业务逻辑 |
| `server/internal/store/nonce_store.go` | Nonce 内存存储（防重放） |

#### 代码审查

详见 `memory-bank/phase2-1-login-auth.md` 第五节「人工代码审查思路」

---

## 待办事项 📋

### Phase 2.1: Nonce 签名登录
- [ ] `server/internal/store/nonce_store.go` - Nonce 内存存储
- [ ] `server/internal/service/auth_service.go` - Challenge/Login 业务逻辑
- [ ] `server/internal/controller/auth_controller.go` - /auth/* 端点
- [ ] `client/internal/commands/login.go` - /auth login 命令
- [ ] `client/internal/store/token.go` - Token 独立管理
- [ ] `server/internal/router/router.go` - 添加 /auth/* 路由

### Phase 3: WebSocket 基础通信
- [ ] server/internal/websocket/pool.go - 连接池
- [ ] server/internal/websocket/conn.go - 连接封装
- [ ] server/internal/protocol/message.go - 消息协议
- [ ] client/internal/chat/client.go - WebSocket 客户端

### Phase 4-9: 后续阶段
详见 `implementation-plan.md`
---

---

### Phase 2 完善与代码优化（2025-01-27 完成）

**本轮更新内容：**

#### 1. 客户端 Phase 2 模块完成

**新建文件：**
| 文件 | 描述 |
| --- | --- |
| `client/internal/auth/keypair.go` | Ed25519 密钥对生成/加载（PEM 格式） |
| `client/internal/auth/signer.go` | Nonce 签名器（Ed25519） |
| `client/internal/store/paths.go` | ~/.charline/ 路径管理 |
| `client/internal/store/credential.go` | 凭证存储（JWT + 元数据） |
| `client/internal/commands/join.go` | /join 命令实现 |

**修改文件：**
| 文件 | 修改内容 |
| --- | --- |
| `client/cmd/main.go` | 添加 join 命令处理 |
| `client/go.mod` | 添加 godotenv 依赖 |
| `client/internal/config/config.go` | 添加 .env 加载和项目根目录检测 |

#### 2. 服务端 public_key 支持

**新建文件：**
| 文件 | 描述 |
| --- | --- |
| `pkg/sqlite/migrations/002_add_public_key.sql` | 数据库迁移（users 表添加 public_key） |

**修改文件：**
| 文件 | 修改内容 |
| --- | --- |
| `server/internal/controller/invite_controller.go` | 添加 public_key 字段，支持公钥验证 |
| `server/internal/service/invite_service.go` | Activate() 新增 publicKey 参数 |
| `server/internal/store/invite.go` | 完整激活事务（用户创建+关联+邀请码更新） |

#### 3. inviteService 接口统一

**新建文件：**
| 文件 | 描述 |
| --- | --- |
| `server/internal/service/interfaces.go` | `InviteServiceInterface` 统一接口定义 |

**修改文件：**
| 文件 | 修改内容 |
| --- | --- |
| `server/internal/controller/invite_controller.go` | 使用 `service.InviteServiceInterface`，移除重复定义 |

#### 4. 响应结构统一

**修改文件：**
| 文件 | 修改内容 |
| --- | --- |
| `server/internal/controller/invite_controller.go` | 移除重复的 `writeSuccess`/`writeError`，改用 `response.go` 的 `RespondSuccess`/`RespondError` |

#### 5. 文档更新

**新建文档：**
| 文件 | 描述 |
| --- | --- |
| `memory-bank/phase2-1-login-auth.md` | Phase 2.1 Nonce 签名登录完整实施文档（含代码审查思路） |

**修改文档：**
| 文件 | 修改内容 |
| --- | --- |
| `memory-bank/document-navigation.md` | 添加 phase2-1-login-auth.md 引用 |
| `memory-bank/progress.md` | 标记 Phase 2 完成，更新待办事项 |

#### 代码统计

```
 11 files changed, 478 insertions(+), 511 deletions(-)
```

- **客户端新增**：auth/、store/、commands/ 三个包
- **服务端优化**：接口统一、响应结构统一
- **文档完善**：Phase 2.1 实施指南 + 代码审查清单

---

### 下一步计划

1. **Phase 2.1**: 实现 `/auth login` 命令（Nonce 签名登录）
2. **Phase 3**: WebSocket 通信基础

详见 `memory-bank/phase2-1-login-auth.md`
