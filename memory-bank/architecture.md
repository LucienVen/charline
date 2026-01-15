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
```

---

## 已完成模块

### Phase 0: 基础设施（已完成）

#### 服务端模块

**1. 配置管理** (`server/internal/config/config.go`)
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

**2. 日志系统** (`server/internal/logger/`)
- 职责：结构化日志与请求追踪
- 文件结构：
  - `logger.go`: zap 日志封装，自定义时间格式
  - `middleware.go`: HTTP 请求日志中间件
  - `context.go`: 请求 ID 上下文管理
- 日志格式：`+0800 2025-01-15 17:51:54 INFO cmd/main.go:35 | 消息内容 {字段}`
- 特性：
  - 时区支持（+0800）
  - 彩色输出（开发环境）
  - JSON 输出（生产环境）
  - 请求 ID 追踪
  - 调用堆栈（Error 级别）

**3. HTTP 服务** (`server/cmd/main.go`)
- 职责：服务入口与路由管理
- 框架：go-chi v5
- 端点：
  - GET /health: 健康检查
- 特性：
  - 优雅关闭（5秒超时）
  - 请求日志中间件
  - 信号处理（SIGINT/SIGTERM）

#### 客户端模块

**1. 基础框架** (`client/cmd/main.go`)
- 职责：命令行入口
- 支持命令：hello

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

---

## 技术栈

| 功能 | 库 | 版本 | 状态 |
|------|-----|------|------|
| Web 框架 | go-chi/chi/v5 | v5.2.4 | ✅ 已集成 |
| 日志 | uber-go/zap | v1.27.1 | ✅ 已集成 |
| 配置 | - | - | ✅ 已实现 |
| WebSocket | gorilla/websocket | - | ⏳ 待集成 |
| JWT | golang-jwt/jwt | - | ⏳ 待集成 |
| SQLite | modernc.org/sqlite | - | ⏳ 待集成 |

---

## 设计决策

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
- [ ] **进行中**: Phase 1 - 邀请激活系统
- [ ] **待定**: Phase 2 - 客户端注册
- [ ] **待定**: Phase 3 - WebSocket 通信

