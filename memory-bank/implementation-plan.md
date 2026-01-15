# CharLine IM 项目实施规划

## 项目概述

**项目名称**: CharLine - 命令行终端聊天软件
**架构**: Client-Server IM 通讯模型
**技术栈**: Golang 1.25.5 + WebSocket + SQLite
**部署目标**: 2核2G 云服务器

---

## 当前状态

### 已完成
- ✅ Git 仓库初始化
- ✅ 目录结构创建（server/client 分离）
- ✅ 设计文档撰写（memory-bank/）
- ✅ go.mod 模块声明

### 待解决
- ✅ Go 版本正确（1.25.5）
- ❌ 空的 main.go 文件
- ❌ 客户端缺少 cmd/main.go
- ❌ 无构建脚本
- ❌ 无依赖库配置

---

## 实施路线图（10阶段）

### Phase 0: 项目初始化（当前阶段）

**目标**: 能启动空程序，基础构建可用

#### 步骤 0.1: 验证项目配置
```
修改文件:
- server/go.mod: 保持 Go 1.25.5
- client/go.mod: 保持 Go 1.25.5
- 创建 client/cmd/main.go
```

#### 步骤 0.2: 创建构建脚本
```
创建文件:
- Makefile (根目录)
- server/.env.example
- client/.env.example
```

**Makefile 目标**:
```makefile
.PHONY: build server client clean deps

deps:         # 拉取依赖
server:       # 构建服务端
client:       # 构建客户端
run-server:   # 运行服务端
run-client:   # 运行客户端
clean:        # 清理构建产物
```

#### 步骤 0.3: 可运行的最小程序
```
server/cmd/main.go:
- 简单的 HTTP 服务（使用 go-chi）
- 健康检查端点 /health
- 优雅关闭

client/cmd/main.go:
- 命令行参数解析
- 简单的 "Hello" 输出
```

**检验标准**:
```bash
make run-server  # 能启动，监听端口
make run-client  # 能输出 Hello
curl localhost:8080/health  # 返回 200
```

---

### Phase 1: 邀请激活系统

**目标**: 服务端提供邀请码激活 API，返回 JWT Token

#### 服务端实现

**1.1 数据库层** (`server/internal/store/`)
```
invite.go:      # 邀请码存储（内存）
- 生成邀请码
- 验证邀请码
- 消耗邀请码
```

**1.2 配置管理** (`server/internal/config/`)
```
config.go:
- JWT 密钥配置
- 服务端口配置
- 环境变量加载
```

**1.3 认证模块** (`server/internal/auth/`)
```
jwt.go:
- Token 生成
- Token 验证
- 版本号机制
```

**1.4 API 层** (`server/internal/api/`)
```
handler.go:
- POST /api/invite/generate    # 生成邀请码
- POST /api/invite/activate    # 激活邀请码
```

#### API 规范

```http
POST /api/invite/generate
Response: {"code": "INV-123456"}

POST /api/invite/activate
Request:  {"code": "INV-123456", "username": "alice"}
Response: {"token": "eyJhbG...", "version": 1}
```

**检验标准**:
```bash
# 生成邀请码
curl -X POST http://localhost:8080/api/invite/generate
# {"code":"INV-ABC123"}

# 激活邀请码
curl -X POST http://localhost:8080/api/invite/activate \
  -H "Content-Type: application/json" \
  -d '{"code":"INV-ABC123","username":"alice"}'
# {"token":"eyJhbG...","version":1}

# 验证 token（添加 /api/validate 端点）
curl http://localhost:8080/api/validate \
  -H "Authorization: Bearer eyJhbG..."
# {"valid":true,"username":"alice"}
```

---

### Phase 2: 客户端注册与存储

**目标**: 客户端完成注册流程，本地 SQLite 存储凭证

#### 客户端实现

**2.1 数据存储** (`client/internal/store/`)
```
sqlite.go:
- 初始化数据库
- 存储用户凭证
- 存储服务器地址
- 数据库迁移
```

**Schema**:
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE,
    server_url TEXT,
    token TEXT,
    token_version INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**2.2 命令系统** (`client/internal/commands/`)
```
register.go:
- /join <url> <username> <invite_code>
- 调用激活 API
- 存储返回的 token
```

**2.3 CLI 入口** (`client/cmd/main.go`)
```
- 子命令解析
- /join 实现
```

**检验标准**:
```bash
# 客户端注册
./client join http://localhost:8080 alice INV-ABC123
# 注册成功！欢迎 alice

# 验证 SQLite 存储
sqlite3 charline.db "SELECT username, token FROM users;"
# alice|eyJhbG...
```

---

### Phase 3: WebSocket 基础通信

**目标**: 建立 WebSocket 长连接，实现双向聊天

#### 服务端实现

**3.1 WebSocket 处理** (`server/internal/websocket/`)
```
pool.go:        # 连接池管理
  - 管理所有活跃连接
  - 广播消息
  - 注册/注销连接

conn.go:        # 单个连接封装
  - 读写消息
  - 心跳检测
  - 优雅关闭

handler.go:     # WebSocket 升级
  - GET /ws (需 JWT 认证)
  - 升级 HTTP 为 WebSocket
```

**3.2 消息协议** (`server/internal/protocol/`)
```
message.go:
type Message struct {
    Type      string  `json:"type"`      // chat, heartbeat, system
    From      string  `json:"from"`      // 发送者
    To        string  `json:"to"`        // 接收者（单聊）或群组ID
    Content   string  `json:"content"`   // 消息内容
    Timestamp int64   `json:"timestamp"`
}
```

#### 客户端实现

**3.3 WebSocket 客户端** (`client/internal/chat/`)
```
client.go:
- 建立 WebSocket 连接
- 发送消息
- 接收消息
- 自动重连
```

**3.4 聊天界面** (`client/cmd/main.go`)
```
/chat 子命令:
- 进入聊天模式
- 实时显示接收消息
- 发送消息
```

**消息流程**:
```
Client A                    Server                    Client B
   |                         |                          |
   |---(1) WebSocket Connect-------------------------->|
   |                         |                          |
   |---(2) Message: "Hello"-------------------------->|
   |                         |                          |
   |                         |--(3) Broadcast---------->|
   |                         |                          |
   |<--(4) Message: "Hello"----------------------------|
```

**检验标准**:
```bash
# 终端 A（用户 alice）
./client chat
[连接成功] 已连接到服务器
> Hello World!
[系统] 消息已发送

# 终端 B（用户 bob）
./client chat
[连接成功] 已连接到服务器
[alice] Hello World!
```

---

### Phase 4: Kafka 消息队列（可选）

**目标**: 使用 Kafka 解耦消息接收与转发

#### 服务端改造

**4.1 Kafka 集成** (`server/internal/broker/`)
```
kafka.go:
- 生产者：接收 WebSocket 消息，写入 Kafka
- 消费者：从 Kafka 读取，广播到 WebSocket
```

**4.2 消息流转**
```
WebSocket → Kafka Producer → Kafka Topic → Kafka Consumer → WebSocket
   入站                            存储                            出站
```

**Topic 设计**:
```
chat.messages:     聊天消息
chat.presence:     在线状态
chat.system:       系统消息
```

**检验标准**:
- 消息通过 Kafka 转发
- 消费者组实现负载均衡
- Kafka 持久化验证

---

### Phase 5: 本地记录与命令系统

**目标**: 客户端存储聊天历史，完善命令体系

#### 客户端实现

**5.1 消息存储** (`client/internal/store/`)
```
messages.go:
- CREATE TABLE messages (id, from_user, content, timestamp)
- 存储接收的消息
- 查询历史记录
```

**5.2 命令完善** (`client/internal/commands/`)
```
/history:     # 显示聊天历史
/exit:        # 退出程序
/help:        # 帮助信息
/clear:       # 清屏
```

**检验标准**:
```bash
./client chat
> /history
[2025-01-15 10:30:00] alice: Hello
[2025-01-15 10:30:15] bob: Hi there
```

---

### Phase 6: 多群组支持

**目标**: 支持用户加入多个群组，消息按群组路由

#### 服务端实现

**6.1 群组管理** (`server/internal/groups/`)
```
manager.go:
- 群组创建
- 用户加入/离开
- 群组成员管理
```

**6.2 消息路由** (`server/internal/broker/`)
```
router.go:
- 根据 To 字段路由消息
- 单聊：直接转发给目标用户
- 群聊：转发给群组所有成员
```

#### 客户端实现

**6.3 群组切换** (`client/internal/commands/`)
```
/select:       # 选择当前群组
/groups:       # 列出所有群组
/join <group>: # 加入群组
```

**消息协议扩展**:
```json
{
    "type": "chat",
    "from": "alice",
    "to": "group:general",  // "user:bob" 或 "group:general"
    "content": "Hi everyone",
    "timestamp": 1736926800
}
```

---

### Phase 7: 安全加固

**目标**: HTTPS + Token 刷新机制

#### 安全措施

**7.1 HTTPS 配置**
```
server/.env:
- TLS_CERT=path/to/cert.pem
- TLS_KEY=path/to/key.pem
```

**7.2 Token 刷新**
```
POST /api/auth/refresh
Request:  {"token": "old_token"}
Response: {"token": "new_token", "version": 2}
```

**7.3 Token 作废机制**
```
- 版本号升级
- 旧版本 token 自动失效
- 客户端强制重新登录
```

---

### Phase 8: Docker 部署

**目标**: docker-compose 一键部署

#### 部署文件

**8.1 服务端 Dockerfile**
```dockerfile
FROM golang:1.25.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

**8.2 docker-compose.yml**
```yaml
version: '3.8'
services:
  charline-server:
    build: ./server
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=${JWT_SECRET}
    restart: unless-stopped

  kafka:
    image: bitnami/kafka:latest
    # ... Kafka 配置
```

**检验标准**:
```bash
docker-compose up -d
# 所有服务正常运行
```

---

### Phase 9: 测试与质量保障

**目标**: 完善的测试覆盖和代码质量工具

#### 测试实现

**9.1 单元测试**
```
server/internal/auth/jwt_test.go
server/internal/store/invite_test.go
client/internal/store/sqlite_test.go
```

**9.2 集成测试**
```
tests/integration/chat_test.go:
- 启动测试服务器
- 模拟多个客户端
- 验证消息收发
```

**9.3 代码质量工具**
```
.golangci.yml:  # lint 配置
Makefile:
  - make test       # 运行测试
  - make lint       # 代码检查
  - make coverage   # 覆盖率报告
```

---

## 关键技术决策

### 依赖库选择

| 功能 | 库 | 理由 |
|------|-----|------|
| Web 框架 | go-chi | 轻量，自由度高 |
| WebSocket | gorilla/websocket | 成熟稳定 |
| JWT | golang-jwt/jwt | 官方推荐 |
| SQLite | modernc.org/sqlite | 纯 Go，无 CGO |
| Kafka | kafka-go | 纯 Go 实现 |

### 架构原则

1. **模块化**: 按职责拆分，单一文件单一职责
2. **无全局变量**: 使用依赖注入
3. **读写分离**: 网络层不处理业务逻辑
4. **状态管理**: 显式结构体，可追踪

---

## 文件清单

### 需要创建的关键文件

#### 服务端
```
server/
├── cmd/
│   └── main.go                    # 入口
├── internal/
│   ├── api/handler.go             # HTTP 处理
│   ├── api/ws_handler.go          # WebSocket 处理
│   ├── auth/jwt.go                # JWT 实现
│   ├── broker/
│   │   ├── kafka.go               # Kafka 集成
│   │   └── router.go              # 消息路由
│   ├── config/config.go           # 配置管理
│   ├── groups/manager.go          # 群组管理
│   ├── protocol/message.go        # 消息协议
│   ├── store/
│   │   ├── invite.go              # 邀请码存储
│   │   └── offline.go             # 离线消息
│   └── websocket/
│       ├── pool.go                # 连接池
│       └── conn.go                # 连接封装
├── .env.example                   # 环境变量模板
└── Dockerfile                     # 镜像构建
```

#### 客户端
```
client/
├── cmd/
│   └── main.go                    # 入口
├── internal/
│   ├── auth/client.go             # 认证客户端
│   ├── chat/
│   │   ├── client.go              # WebSocket 客户端
│   │   └── ui.go                  # 聊天界面
│   ├── commands/
│   │   ├── register.go            # 注册命令
│   │   ├── chat.go                # 聊天命令
│   │   └── history.go             # 历史命令
│   ├── config/config.go           # 配置管理
│   └── store/
│       ├── sqlite.go              # SQLite 操作
│       └── messages.go            # 消息存储
├── .env.example
└── Dockerfile
```

#### 根目录
```
├── Makefile                       # 构建脚本
├── docker-compose.yml             # 部署配置
├── .golangci.yml                  # 代码质量配置
└── go.work                        # Go workspace
```

---

## 检验清单

### Phase 0 完成检验
- [ ] `make server` 构建成功
- [ ] `make client` 构建成功
- [ ] `make run-server` 服务启动
- [ ] `curl localhost:8080/health` 返回 200
- [ ] 客户端运行输出 Hello

### Phase 1 完成检验
- [ ] POST /api/invite/generate 生成邀请码
- [ ] POST /api/invite/activate 返回 JWT
- [ ] Token 验证端点工作正常
- [ ] 邀请码只能使用一次

### Phase 2 完成检验
- [ ] 客户端 /join 命令可用
- [ ] SQLite 存储用户凭证
- [ ] 注册后可查询到用户记录

### Phase 3 完成检验
- [ ] WebSocket 连接建立成功
- [ ] 两个客户端可以互相收发消息
- [ ] 心跳机制工作正常

---

## 开发注意事项

1. **文档同步**: 每完成一个阶段，更新 `memory-bank/@architecture.md` 和 `memory-bank/@progress.md`

2. **代码质量**:
   - 使用 `golangci-lint` 检查
   - 保持函数简洁（< 50 行）
   - 避免多层嵌套

3. **性能考虑**:
   - 控制 goroutine 数量
   - 使用对象池减少分配
   - 避免阻塞操作

4. **错误处理**:
   - 显式处理所有错误
   - 提供有意义的错误信息
   - 记录关键错误日志
