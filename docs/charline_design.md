# CharLine 设计方案

CharLine 是一个**字符界面（CLI）聊天软件**，采用邀请制，支持群组聊天，消息通过云服务器转发，无中心化存储日志，客户端本地可持久化消息历史。

---

## 1. 系统目标

- 提供跨平台 CLI 聊天体验
- 无中心化聊天记录存储
- 用户通过**一次性邀请码**加入系统
- 群组聊天为默认模式（预留多群组扩展）
- 客户端本地保存用户数据与消息历史
- 服务器负责**转发消息**与**基础认证**

---

## 2. 架构概览

```
          +--------------------+
          | 2核2G 云服务器      |
          | CharLine Server    |
          |---------------------|
          | 1. 邀请验证         |
          | 2. JWT 认证         |
          | 3. 消息转发         |
          | 4. Kafka 投递       |
          +---------+-----------+
                    |
         Kafka (可选，用于扩展)
                    |
   +----------------+----------------+
   |                                 |
   v                                 v
客户端 A                        客户端 B
charline CLI                    charline CLI
本地 sqlite                     本地 sqlite
本地 token 配置                 本地 token 配置
```

---

## 3. 关键技术路线

| 模块              | 方案                                                         |
|-------------------|--------------------------------------------------------------|
| 通信协议           | TCP 或 Websocket（优先 WS，易扩展）                         |
| 客户端本地数据     | SQLite（纯 Go 驱动 `modernc.org/sqlite`）                   |
| 服务端存储         | 无（不落盘）                                                 |
| 状态管理           | 无会话状态，依赖 JWT                                         |
| Token 机制         | JWT（无过期，但含版本号，版本号不匹配即失效）               |
| 邀请机制           | 服务器生成一次性邀请码，使用后即冻结                        |
| 消息投递扩展       | 可选接入 Kafka                                               |
| 聊天记录存储       | 客户端本地 SQLite，文件随客户端存在                         |

---

## 4. 邀请 / 注册流程

### 4.1 邀请码格式
```
https://server/join?code=xxx
```

### 4.2 客户端加入命令
```bash
charline join <inviteURL> <username> default
```

### 4.3 加入后服务端发回数据
- `username`
- `clientSecret`
- `token` (JWT)

### 4.4 客户端本地保存

默认保存路径示例：

```
~/.charline/config.json
~/.charline/charline.db
```

`config.json` 内容结构：

```json
{
  "username": "alice",
  "clientSecret": "RANDOM-CLIENT-SECRET",
  "token": "JWT-TOKEN-STRING",
  "ver": 1
}
```

---

## 5. Token 认证方案（JWT）

### 5.1 Payload 示例

```json
{
  "uid": "alice",
  "ver": 1
}
```

### 5.2 特点

- **无 exp（不过期）**
- 验签 + `ver` 版本匹配才有效
- 服务端无存储，无 session

### 5.3 Token 验证逻辑

```go
jwt.ParseWithClaims(...)
if claims.Ver != serverVer {
    return "token 版本过旧，请重新登录获取新 token"
}
```

### 5.4 用户登录行为

老用户登录（换设备）：

```bash
charline login <username> <clientSecret>
```

验证通过后颁发新 token，不需要邀请码。

---

## 6. 客户端 SQLite

### 6.1 设计

数据库文件位置：
```
~/.charline/charline.db
```

建表示例：

```sql
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender TEXT,
    content TEXT,
    group_name TEXT,
    ts DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 6.2 使用场景

- 保存聊天历史
- 离线可阅读
- 退出再进入时从本地加载

---

## 7. 消息收发流程

### 7.1 发送消息

```
客户端 → 服务器 → 转发到群组所有在线用户
                        ↳ 可选写入 Kafka（扩展）
```

### 7.2 接收消息

- 进入群组后监听 WS / TCP
- 不退出指令期间持续接收与打印消息
- 退出指令后仍接收消息（后台线程落库）

---

## 8. 推荐 CLI 指令体系

| 指令                                           | 说明 |
|------------------------------------------------|------|
| `charline join <url> <username> <group>`       | 使用邀请码加入系统 |
| `charline login <username> <clientSecret>`     | 老用户重新登录获取 token |
| `charline select <group>`                      | 进入群组 |
| `charline chat`                                 | 开始聊天模式 |
| `exit`                                          | 退出聊天（不退出程序） |
| `history`                                       | 查看本地聊天记录 |
| `logout`                                        | 删除本地 token |
| `version`                                       | 显示客户端版本号 |

---

## 9. 服务端模块划分

```
/server
  ├── cmd
  │     └── charline-server/main.go
  ├── internal
  │     ├── invite       # 邀请码生成与一次性校验
  │     ├── auth         # JWT 签发与验证
  │     ├── transport    # Websocket/TCP 连接管理
  │     └── dispatcher   # 消息转发、Kafka 投递
```

---

## 10. 客户端模块划分

```
/client
  ├── cmd
  │     └── charline-cli/main.go
  ├── internal
  │     ├── config       # 本地配置: username/token/clientSecret
  │     ├── storage      # SQLite 封装
  │     ├── auth         # login/join/token 刷新
  │     ├── ui           # 命令行字符界面输入输出
  │     └── client       # 与服务器通信 (ws/tcp)
```

---

## 11. 故障与安全机制

| 风险点                  | 处理方式 |
|-------------------------|-----------|
| clientSecret/Token 泄露 | 更新服务端版本号 ver，使旧 token 自动失效 |
| 账号丢失（无配置文件）  | 需要邀请码重新加入 |
| 私钥泄露                | 更换 JWT 签名密钥，所有 token 失效 |

---

## 12. 后续可扩展方向

- 群组权限管理
- 点对点私聊
- 用户状态（在线/离线）
- 头像与基本资料
- 消息加密（E2EE）

---

## 总结

CharLine 是一个**纯终端字符界面聊天系统**，通过**本地 SQLite + JWT 无状态认证 + 邀请制注册**构建，重点不在于中心服务，而是**客户端自持数据 + 服务端轻逻辑转发**的方式实现低成本、可拓展、安全可控的聊天系统。

---
