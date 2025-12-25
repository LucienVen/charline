# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

CharLine 是一个命令行终端聊天软件,采用邀请制,支持群组聊天。核心特点:
- **服务端轻量化**: 仅负责消息转发和基础认证,不存储聊天记录
- **客户端本地存储**: 所有聊天历史保存在客户端 SQLite 数据库
- **无状态认证**: 基于 JWT 的无过期时间令牌系统,通过版本号控制失效
- **实时通信**: 基于 WebSocket 的消息推送
- **可选 Kafka**: 支持通过 Kafka 进行消息队列扩展

## 系统架构

```
客户端 A (CLI)                    客户端 B (CLI)
    ↓                                  ↑
    ←------→ CharLine Server ←--------→
              (消息转发)
                 ↓
            Kafka (可选)
```

### 架构设计原则
1. **服务端不做持久化**: 仅在内存中维护邀请码和连接状态
2. **客户端持有数据**: 所有聊天记录存储在 `~/.charline/charline.db`
3. **无会话状态**: 服务端不维护 session,完全依赖 JWT
4. **隐私优先**: 服务端无法访问历史聊天记录

## 项目结构

```
charline/
├── server/                 # 服务端代码
│   ├── cmd/
│   │   └── charline-server/main.go
│   ├── internal/
│   │   ├── invite/        # 邀请码生成与验证
│   │   ├── auth/          # JWT 签发与验证
│   │   ├── transport/     # WebSocket 连接管理
│   │   └── dispatcher/    # 消息转发与 Kafka 投递
│   ├── go.mod
│   └── Dockerfile
├── client/                 # 客户端代码
│   ├── cmd/
│   │   └── charline-cli/main.go
│   ├── internal/
│   │   ├── config/        # 本地配置管理
│   │   ├── storage/       # SQLite 封装
│   │   ├── auth/          # 登录/注册/Token 刷新
│   │   ├── ui/            # 命令行界面
│   │   └── client/        # 服务端通信 (WebSocket)
│   ├── go.mod
│   └── Dockerfile
├── docs/                   # 详细设计文档
│   ├── charline_design.md
│   ├── charline_full_plan.md
│   └── charline_implementation_steps.md
└── Makefile                # 构建脚本
```

## 开发命令

### 构建命令
```bash
# 构建服务端
make build-server

# 构建客户端
make build-client

# 构建所有平台
make build-all
```

### 测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 代码质量
```bash
# 格式化代码
make fmt

# 运行 linter
make lint
```

### 运行
```bash
# 启动服务端
make run-server

# 启动客户端
make run-client
```

## 核心设计决策

### 认证机制
- **JWT Token 结构**: `{"uid": "username", "group": "default", "v": 1}`
  - 无过期时间 (exp)
  - 通过版本号 (v) 控制全局失效
  - 服务端验证签名 + 版本号匹配
- **本地存储**: `~/.charline/config.json`
  ```json
  {
    "username": "alice",
    "clientSecret": "RANDOM-SECRET",
    "token": "JWT-TOKEN",
    "ver": 1
  }
  ```
- **注册流程**: 使用一次性邀请码加入系统

### 存储策略
- **服务端**: 不持久化聊天记录,仅内存维护邀请码
- **客户端**: SQLite 数据库 `~/.charline/charline.db`
  ```sql
  CREATE TABLE messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      sender TEXT,
      content TEXT,
      group_name TEXT,
      ts DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```

### CLI 命令体系
- `charline join <url> <username> <group>` - 使用邀请码加入
- `charline login <username> <clientSecret>` - 已有用户登录
- `charline select <group>` - 选择群组
- `charline chat` - 进入聊天模式
- `exit` - 退出聊天界面
- `history` - 查看本地历史记录
- `logout` - 删除本地 token
- `version` - 显示客户端版本

## 技术栈

### Go 版本
- Go 1.25.5 (见 go.mod)

### 服务端依赖 (计划)
- **Web 框架**: Gin 或 Fiber (待确定)
- **WebSocket**: gorilla/websocket
- **JWT**: golang-jwt/jwt
- **Kafka**: kafka-go 或 segmentio/kafka-go (可选)

### 客户端依赖 (计划)
- **SQLite**: modernc.org/sqlite (纯 Go 实现,无 CGO)
- **WebSocket**: gorilla/websocket
- **CLI 框架**: cobra 或 spf13/cobra

## 开发阶段

项目分 10 个阶段实施,详见 [docs/charline_full_plan.md](docs/charline_full_plan.md):

1. **Phase 0**: 项目初始化 - 空可运行结构
2. **Phase 1**: 服务端 MVP - 邀请激活 API + JWT 发放
3. **Phase 2**: 客户端注册 + 本地 SQLite 存储
4. **Phase 3**: WebSocket 通信 - 实时聊天 MVP
5. **Phase 4**: Kafka 集成 - 消息队列
6. **Phase 5**: CLI 命令 - history/exit/help/select
7. **Phase 6**: 多群组支持
8. **Phase 7**: 安全优化 - HTTPS/Token 刷新
9. **Phase 8**: 部署上线 - docker-compose
10. **Phase 9**: 测试保障 - 单元测试/集成测试

## 重要文件位置

### 客户端数据目录
- 配置文件: `~/.charline/config.json`
- 数据库: `~/.charline/charline.db`
- Token: 存储在 config.json 中

### 服务端配置
- 环境变量: `.env` (见 .gitignore)
- JWT 密钥: 通过环境变量配置
- 邀请码: 内存存储 (可扩展为 Redis)

## 代码风格

遵循 Go 标准实践:
- 使用 `go fmt` 格式化代码
- 遵循 Effective Go 指南
- 错误处理必须显式处理,不能忽略
- 包注释遵循 godoc 规范

## 模块职责

### 服务端模块
- **invite**: 邀请码生成、验证、一次性使用检查
- **auth**: JWT 签发、验证、版本号检查
- **transport**: WebSocket 连接管理、消息广播
- **dispatcher**: 消息转发逻辑、Kafka 生产者/消费者

### 客户端模块
- **config**: 读写本地配置文件
- **storage**: SQLite 数据库操作封装
- **auth**: 注册/登录逻辑、Token 刷新
- **ui**: 命令行输入输出、聊天界面
- **client**: WebSocket 客户端、重连逻辑

## 安全注意事项

1. **Token 管理**: 服务端通过增加版本号可使所有旧 Token 失效
2. **Client Secret**: 用户换设备需要 username + clientSecret 重新登录
3. **HTTPS**: 生产环境必须使用 HTTPS
4. **输入验证**: 所有用户输入必须验证和清理
5. **敏感信息**: 永远不要在代码中硬编码密钥或密码

## 扩展方向 (未来)

- 群组权限管理
- 点对点私聊
- 用户在线状态
- 消息端到端加密 (E2EE)
- TUI 客户端 (bubbletea)
- Web 管理界面

## 参考文档

- [设计方案](docs/charline_design.md) - 详细架构设计
- [实施步骤](docs/charline_implementation_steps.md) - 开发步骤总结
- [完整计划](docs/charline_full_plan.md) - 10 阶段实施计划
