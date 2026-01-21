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
