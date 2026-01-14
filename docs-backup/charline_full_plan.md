# CharLine 全阶段实施步骤与项目结构草稿

## 全阶段概览（可执行表）

| 阶段 | 任务目标 | 关键产出 | 责任方 |
|------|-----------|-----------|--------|
| 1. 项目初始化 | 确定架构、建仓库 | Git 仓库、readme、大纲文档 | 全体 |
| 2. 服务端雏形 | 邀请码 / 注册 / JWT 验证接口 | 可运行的 REST 服务与注册流程 | Server |
| 3. 客户端基础 | CLI 注册 / 登录、保存 Token | 可登录的 CLI 应用、Token 本地化 | Client |
| 4. 消息转发通道 | Kafka 接入、消息发送 API | 消息进入 Kafka Topic | Server |
| 5. 消息接收机制 | 监听 Kafka / WebSocket 推送 | 可实时接收消息的客户端 | Server/Client |
| 6. 本地存档 | SQLite 持久化收发消息 | 本地聊天记录与查询工具 | Client |
| 7. 群聊与状态管理 | 群组选取、退出、在线状态 | 简易群组协议与命令设计 | Server/Client |
| 8. 部署上线 | 云服务器（2c2g）、日志、TLS | 可供用户使用的可部署方案 | Server |
| 9. 验收与文档 | 用户手册、debug指南、故障处理 | 初版 v1.0 发布 | 全体 |

---


建议实施步骤按“最小可运行 → 功能迭代 → 扩展优化”推进，分阶段如下：

------

## 阶段 0：基础准备

**目标：搭好环境与工程骨架**

1. 创建 Git 仓库与主目录
2. 拆分 `client/` 和 `server/` 两个独立工程结构
3. 制定 Go module 路径、版本要求（Go ≥ 1.20）
4. 初始化依赖包（gin/fiber、gorilla websocket、kafka-go、sqlite3等）
5. 编写 README：项目目标与运行方式总览

产出物：

- 项目目录结构
- 可执行的 `go run client` 和 `go run server`（空逻辑）

------

## 阶段 1：服务端基础功能 MVP

**目标：能启动、能验证邀请、能返回 Token**

1. 实现 `/invite/activate` API
2. 实现用户结构体、邀请码结构体、本地存储（json或轻量boltdb）
3. 注册成功后发放 Token（JWT）
4. Token 中包含：username、过期时间、群组名

产出物：

- 邀请加入流程可跑通
- curl/HTTP 测试可验证注册 & Token

------

## 阶段 2：客户端注册与本地存储

**目标：客户端能加入，并把用户信息本地落盘**

1. 实现加入命令 `charline <url> <username> <group>`
2. 调用 `/invite/activate`
3. 返回数据写入 sqlite 本地：
   - users
   - groups
4. 校验本地是否已有用户，避免重复注册

产出物：

- 客户端首次启动 → 注册成功 → 本地有数据

------

## 阶段 3：WebSocket 通信打通

**目标：连接、发送、接收消息**

1. 服务端实现 `/ws`，校验 token
2. 客户端实现 WS 连接、断线重连
3. 输入文本发送消息（简单打印处理）
4. 服务端广播消息给所有在线客户端

产出物：

- 两个客户端终端可以互相看到消息
- 聊天 MVP 完成（无持久化）

------

## 阶段 4：Kafka 接入

**目标：消息路径从客户端→服务端→Kafka→服务端→客户端**

1. docker-compose 启 Kafka
2. 服务端 Producer：收到客户端消息 → push Kafka Topic
3. 服务端 Consumer：消费消息 → 分发 WebSocket
4. 保障基础投递可靠性：`acks=all`

产出物：

- Kafka 在消息链路中工作
- 服务端内部不保存聊天记录

------

## 阶段 5：客户端 sqlite 历史记录与命令模式

**目标：聊天内容本地落盘、命令模式可切换**

1. 接收到消息后写入 sqlite
2. `history` 命令 显示历史记录
3. `exit` 命令 → 退出聊天界面
4. `/help` `/select` `/groups` 指令支持

产出物：

- 客户端支持指令交互
- 可查看聊天历史

------

## 阶段 6：多群组支持（预留扩展）

**目标：添加新群组/切换群组可用**

1. 服务端群组表
2. Kafka 多 Topic
3. `charline select <group>` 切换消费目标
4. 群组在线人员列表 `/users`

产出物：

- 单账号多群组支持
- 可动态切换群组

------

## 阶段 7：安全优化

**目标：提升安全基础**

1. 全部接口启用 HTTPS（证书可用 Let’s Encrypt）
2. Token 过期刷新机制
3. 禁止客户端自定义 sender 字段（由服务端注入）
4. 邀请码生成可限制过期时间

选项（可延期）

- 离线消息拉取（短轮询/SSE）
- 消息端到端加密（AES/RSA 混合）

------

## 阶段 8：基础运维部署

**目标：上线一个可长期运行的版本**

1. 云服务器（2C2G）
2. docker-compose：server + kafka + nginx（可选）
3. Systemd 或 supervisor 托管 `charline-server`
4. 日志分级输出（error/info）

产出物：

- 可稳定运行的部署方案
- 部署文档

------

## 阶段 9：质量与测试保障

**目标：项目可维护**

1. CLI 自动化测试（Go test）
2. API 单元测试
3. WebSocket 压测（tiny-http/ws-bench）
4. `make build` `make test` `make run` 命令

产出物：

- 基础测试覆盖率
- 统一构建流程

------

## 阶段 10：未来拓展（可选路线）

- 用户私聊（点对点 WS 通道）
- 离线消息缓存（Redis List）
- 文件/图片传输（Base64 或 MinIO）
- WebUI 客户端（tuido/termui → WebSocket 复用）

------

# 全阶段概览（可执行表）

| 阶段 | 目标           | 完成标志            |
| ---- | -------------- | ------------------- |
| 0    | 准备           | 能启动空程序        |
| 1    | 邀请激活       | Token 返回          |
| 2    | 客户端注册存储 | sqlite 存储成功     |
| 3    | 基础聊天       | WS 双向可聊天       |
| 4    | Kafka          | Kafka 全链路通      |
| 5    | 本地记录/命令  | /history 等可用     |
| 6    | 多群组         | select 可用         |
| 7    | 安全           | HTTPS + Token刷新   |
| 8    | 部署上线       | docker-compose 运行 |
| 9    | 测试保障       | make + go test      |
| 10   | 拓展           | 后续项目演进        |


## 技术流程（参考）

```
[Client] --注册邀请码--> [Server]
[Server] --发JWT Token--> [Client保存]
[Client] --发送消息--> [Server → Kafka]
[Kafka] --订阅消息--> [Server]
[Server] --推送消息(WebSocket)--> [Client]
[Client] --入库SQLite--> 本地历史记录
```

---

## Token 方案（简版）

- JWT 无过期
- 本地存放：`~/.charline/token`
- payload 示例：

```json
{
  "uid": "user_abc",
  "group": "default",
  "v": 1
}
```

- 若版本号冲突，可提示更新但旧 token 不作废

---

## 建议的初始目录结构草稿

```
charline/
├── server/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── api/           # HTTP handler
│   │   ├── auth/          # JWT、邀请码
│   │   ├── broker/        # Kafka producer/consumer
│   │   ├── groups/        # 群组逻辑
│   │   └── config/
│   ├── go.mod
│   └── Dockerfile
│
├── client/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── auth/          # Token 读写
│   │   ├── chat/          # 消息循环
│   │   ├── store/         # SQLite 数据访问
│   │   ├── commands/      # CLI 命令解析
│   │   └── config/
│   ├── go.mod
│   └── Dockerfile
│
├── docs/
│   
│   
│   
│
└── Makefile            # 编译与发布脚本入口
```

---

## 推荐开发顺序（落地版）

1. 新建仓库并写 `README.md`
2. 服务端注册与邀请码接口 → JWT 发放
3. 客户端注册/登录命令 → token 存储
4. Kafka 通道打通（最小可用）
5. WebSocket / 长轮询接收
6. SQLite 接入落地消息
7. CLI 聊天指令集 (`select`, `exit`)
8. 打包编译 + 发布脚本
9. 部署上线 & 测试

---

## 后续扩展路线（参考）

- 私聊、消息加密、TLS
- 多群组体系
- 服务端缓存优化（可选 Redis）
- 二次开发协议（gRPC / QUIC）
- TUI 客户端（bubbletea）

---

## 当前输出物列表

- `charline_design.md`（设计）
- `charline_implementation_steps.md`（实施）
- `charline_full_plan.md`（当前文档）

---



