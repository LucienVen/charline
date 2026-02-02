# Phase 3 WebSocket 问题排查与修复

本文档记录 Phase 3 WebSocket 实施过程中遇到的问题、排查过程和修复方案。

---

## 问题 1：客户端连接后卡住无响应

### 问题描述

**现象**：
```bash
$ ./bin/client connect
正在连接到服务器...
✓ 连接成功
Received message: challenge_response
正在进行身份认证...
# 卡住，没有后续输出
```

**影响**：客户端无法完成认证流程，WebSocket 连接建立后无法正常通信。

---

### 问题排查

#### 1. 检查服务端日志

```
[DEBUG] HandleConnection called
[DEBUG] WebSocket upgrade successful
[DEBUG] Connection created
[DEBUG] Read/Write loops started
[DEBUG] sendChallenge called
[DEBUG] Nonce generated: SWHhheET+huWVUVQRT2P/SmLP8FzGDBvn56f4WD7wZA=
[DEBUG] Challenge message created
[DEBUG] Message marshaled: {"type":"challenge_response","payload":{"nonce":"..."},"timestamp":1770020338598}
[DEBUG] Data sent to channel
[DEBUG] Challenge sent successfully
+0800 2026-02-02 16:18:58    INFO    HTTP request    {"method": "GET", "path": "/ws", "status": 200}
```

**结论**：服务端正常发送 challenge 消息，客户端也收到了（打印了 "Received message: challenge_response"）。

#### 2. 检查客户端代码

**问题代码**（`client/internal/websocket/client.go`）：

```go
// Connect() 方法
func (c *Client) Connect() error {
    // ...
    c.conn = conn
    
    // 启动读写循环
    go c.readLoop()   // ← 后台 goroutine 持续读取消息
    go c.writeLoop()
    
    return nil
}

// Authenticate() 方法
func (c *Client) Authenticate() error {
    // 等待服务器发送 challenge
    _, data, err := c.conn.ReadMessage()  // ← 尝试直接读取
    if err != nil {
        return fmt.Errorf("failed to read challenge: %w", err)
    }
    // ...
}
```

**问题分析**：

```
时间线：
T1: client.Connect() 调用
T2: readLoop() goroutine 启动（后台运行）
T3: client.Authenticate() 调用
T4: 服务端发送 challenge 消息
T5: readLoop() 读取到消息（消费掉）
T6: Authenticate() 调用 ReadMessage()，但消息已被 readLoop 消费
T7: Authenticate() 永久阻塞等待...
```

**根本原因**：**消息读取竞争条件**
- `readLoop` 和 `Authenticate()` 都尝试从同一个 WebSocket 连接读取消息
- `readLoop` 在后台持续运行，会消费所有消息
- `Authenticate()` 永远读不到消息，导致阻塞

---

### 修复方案

#### 方案：使用 Channel 进行消息分发

**设计思路**：
1. `readLoop` 统一读取所有消息
2. 根据消息类型，将消息分发到不同的 channel
3. `Authenticate()` 从专用的 `authChan` 接收消息

**修改代码**：

```go
// 1. Client 结构体添加 authChan
type Client struct {
    conn      *websocket.Conn
    signer    *auth.Signer
    serverURL string
    userID    int64
    send      chan []byte
    authChan  chan Message    // ← 新增：认证消息通道
    closeChan chan struct{}
    closeOnce sync.Once
    isClosed  bool
    mu        sync.RWMutex
}

// 2. NewClient 初始化 authChan
func NewClient(serverURL string, signer *auth.Signer) *Client {
    return &Client{
        serverURL: serverURL,
        signer:    signer,
        send:      make(chan []byte, 256),
        authChan:  make(chan Message, 1), // ← 缓冲 1 个消息
        closeChan: make(chan struct{}),
        isClosed:  false,
    }
}

// 3. Authenticate() 从 channel 接收
func (c *Client) Authenticate() error {
    // 从 authChan 接收 challenge（由 readLoop 发送）
    var msg Message
    select {
    case msg = <-c.authChan:
        // 收到消息
    case <-time.After(10 * time.Second):
        return fmt.Errorf("等待 challenge 超时")
    }

    if msg.Type != "challenge_response" {
        return fmt.Errorf("unexpected message type: %s", msg.Type)
    }
    
    // ... 后续认证逻辑
}

// 4. handleMessage() 将认证消息发送到 authChan
func (c *Client) handleMessage(data []byte) {
    var msg Message
    if err := json.Unmarshal(data, &msg); err != nil {
        fmt.Printf("Failed to parse message: %v\n", err)
        return
    }

    switch msg.Type {
    case "challenge_response", "auth_response", "error":
        // 认证相关消息，发送到 authChan
        select {
        case c.authChan <- msg:
            // 发送成功
        default:
            fmt.Printf("Warning: authChan is full, dropping message: %s\n", msg.Type)
        }
    case "pong":
        // 心跳响应
        fmt.Println("Received pong")
    default:
        // 其他消息类型
        fmt.Printf("Received message: %s\n", msg.Type)
    }
}
```

---

### 修复效果

**修复前**：
```bash
正在连接到服务器...
✓ 连接成功
Received message: challenge_response
正在进行身份认证...
# 卡住
```

**修复后**：
```bash
正在连接到服务器...
✓ 连接成功
正在进行身份认证...
❌ 认证失败: authentication failed: INVALID_SIGNATURE - Invalid signature format
```

✅ **消息读取竞争问题已解决**，客户端能正常收到 challenge 并发送认证请求。

---

## 问题 2：签名格式错误

### 问题描述

**现象**：
```
ERROR: authentication failed: INVALID_SIGNATURE - Invalid signature format
```

**影响**：客户端签名验证失败，无法完成认证。

---

### 问题排查

#### 1. 检查服务端验证逻辑

**代码位置**：`server/internal/websocket/handler.go:78`

```go
// 解码签名
signatureBytes, err := hex.DecodeString(authReq.Signature)
if err != nil || len(signatureBytes) != ed25519.SignatureSize {
    h.sendError(conn, "INVALID_SIGNATURE", "Invalid signature format")
    return
}
```

**结论**：服务端期望签名是 **hex 编码**的字符串。

#### 2. 检查客户端签名生成

**代码位置**：`client/internal/auth/signer.go:33`

```go
// Sign 对 nonce 进行签名
// 返回 base64 编码的签名
func (s *Signer) Sign(nonce string) (string, error) {
    if s.keyPair == nil || s.keyPair.PrivateKey == nil {
        return "", fmt.Errorf("签名器未初始化")
    }

    // 对 nonce 进行签名
    signature := ed25519.Sign(s.keyPair.PrivateKey, []byte(nonce))

    // 返回 base64 编码的签名
    return base64.StdEncoding.EncodeToString(signature), nil  // ← 使用 base64
}
```

**问题分析**：

| 组件 | 编码方式 | 代码位置 |
|------|---------|---------|
| 客户端签名 | `base64` | `client/internal/auth/signer.go:33` |
| 服务端验证 | `hex` | `server/internal/websocket/handler.go:78` |

**根本原因**：**编码格式不匹配**

---

### 修复方案

#### 方案：统一使用 hex 编码

**修改代码**（`client/internal/auth/signer.go`）：

```go
package auth

import (
    "crypto/ed25519"
    "encoding/hex"  // ← 改用 hex
    "fmt"
)

// Sign 对 nonce 进行签名
// 返回 hex 编码的签名
func (s *Signer) Sign(nonce string) (string, error) {
    if s.keyPair == nil || s.keyPair.PrivateKey == nil {
        return "", fmt.Errorf("签名器未初始化")
    }

    // 对 nonce 进行签名
    signature := ed25519.Sign(s.keyPair.PrivateKey, []byte(nonce))

    // 返回 hex 编码的签名（与服务端保持一致）
    return hex.EncodeToString(signature), nil  // ← 改为 hex
}
```

---

### 修复效果

**修复前**：
```
ERROR: authentication failed: INVALID_SIGNATURE - Invalid signature format
```

**修复后**：
```bash
正在连接到服务器...
✓ 连接成功
正在进行身份认证...
❌ 认证失败: USER_NOT_FOUND - User not found
```

✅ **签名格式问题已解决**，签名验证通过，进入用户查询阶段。

---

## 问题 3：数据库表结构缺失（待解决）

### 问题描述

**现象**：
```
ERROR: 查询用户失败 {"error": "SQL logic error: no such column: public_key (1)"}
```

**影响**：无法根据公钥查询用户，认证流程无法完成。

---

### 问题排查

**服务端日志**：
```
+0800 2026-02-02 16:41:10    ERROR    查询用户失败    {"public_key": "4285b49763222b338e6e5fac034f560287e22a3b348e433e58d8eb4850302fcf", "error": "SQL logic error: no such column: public_key (1)"}
```

**代码位置**：`server/internal/store/user.go:109`

```go
func (s *UserStore) GetByPublicKey(publicKey string) (*User, error) {
    var user User
    err := s.db.QueryRow(`
        SELECT id, username, public_key, token_version, created_at, updated_at
        FROM users
        WHERE public_key = ?
    `, publicKey).Scan(&user.ID, &user.Username, &user.PublicKey, &user.TokenVersion, &user.CreatedAt, &user.UpdatedAt)
    // ...
}
```

**根本原因**：数据库 `users` 表缺少 `public_key` 列。

---

### 修复方案

#### 方案：执行数据库迁移

**SQL 语句**：
```sql
-- 添加 public_key 列
ALTER TABLE users ADD COLUMN public_key TEXT NOT NULL DEFAULT '';

-- 创建唯一索引
CREATE UNIQUE INDEX idx_users_public_key ON users(public_key);
```

**执行方式**：
```bash
sqlite3 server/data/charline.db "ALTER TABLE users ADD COLUMN public_key TEXT NOT NULL DEFAULT ''; CREATE UNIQUE INDEX idx_users_public_key ON users(public_key);"
```

---

## 验证结果总结

### 流程验证

| 步骤 | 状态 | 说明 |
|------|------|------|
| 1. WebSocket 连接 | ✅ | 成功建立连接 |
| 2. 接收 Challenge | ✅ | 正常接收 nonce |
| 3. 签名生成 | ✅ | hex 编码正确 |
| 4. 发送认证请求 | ✅ | 消息发送成功 |
| 5. 签名验证 | ✅ | 服务端验证通过 |
| 6. Nonce 验证 | ✅ | 防重放检查通过 |
| 7. 用户查询 | ❌ | 数据库缺少 public_key 列 |

### 修改文件清单

| 文件 | 修改内容 | 状态 |
|------|---------|------|
| `client/internal/websocket/client.go` | 添加 authChan，修改消息分发逻辑 | ✅ 已完成 |
| `client/internal/auth/signer.go` | 签名编码从 base64 改为 hex | ✅ 已完成 |
| 数据库表结构 | 添加 public_key 列和索引 | ⏳ 待处理 |

---

## 技术要点总结

### 1. 并发消息处理

**问题**：多个 goroutine 竞争读取同一个 WebSocket 连接

**解决方案**：
- 使用 channel 进行消息分发
- `readLoop` 统一读取，通过 channel 分发到不同处理器
- 使用 `select` + `time.After` 实现超时机制

**架构图**：
```
┌─────────────────────────────────────────┐
│           Client Application            │
└─────────────────────────────────────────┘
                    │
                    ↓
┌─────────────────────────────────────────┐
│         Authenticate() Method           │
│  (从 authChan 接收认证消息)              │
└─────────────────────────────────────────┘
                    ↑
                    │ authChan
                    │
┌─────────────────────────────────────────┐
│         handleMessage() Method          │
│  (消息路由：认证消息 → authChan)         │
└─────────────────────────────────────────┘
                    ↑
                    │
┌─────────────────────────────────────────┐
│         readLoop() Goroutine            │
│  (统一读取所有 WebSocket 消息)           │
└─────────────────────────────────────────┘
                    ↑
                    │
┌─────────────────────────────────────────┐
│         WebSocket Connection            │
└─────────────────────────────────────────┘
```

### 2. 编码一致性

**问题**：客户端和服务端使用不同的编码格式

**解决方案**：
- 统一使用 hex 编码
- Ed25519 签名：64 字节 → hex 编码 → 128 字符
- 公钥：32 字节 → hex 编码 → 64 字符

**编码对比**：

| 数据 | 原始大小 | base64 编码 | hex 编码 |
|------|---------|------------|---------|
| Ed25519 签名 | 64 字节 | 88 字符 | 128 字符 |
| Ed25519 公钥 | 32 字节 | 44 字符 | 64 字符 |

### 3. 错误传播

**设计原则**：
- 认证错误通过 error channel 正确传递到调用方
- 使用结构化错误消息（错误码 + 详细信息）
- 客户端友好的错误提示

**错误消息示例**：
```json
{
  "type": "error",
  "payload": {
    "code": "INVALID_SIGNATURE",
    "message": "Invalid signature format"
  },
  "timestamp": 1770021670679
}
```

---

## 经验教训

### 1. 并发编程注意事项

- ❌ **错误做法**：多个 goroutine 直接读取同一个资源
- ✅ **正确做法**：使用 channel 进行消息传递和同步

### 2. 协议一致性

- ❌ **错误做法**：客户端和服务端使用不同的编码格式
- ✅ **正确做法**：在协议设计阶段明确编码格式，双方严格遵守

### 3. 调试技巧

- 使用详细的 DEBUG 日志追踪消息流向
- 在关键节点打印消息内容和状态
- 使用 `select` + `timeout` 避免永久阻塞

### 4. 测试策略

- 先测试基础连接，再测试认证流程
- 逐步验证每个环节（连接 → challenge → 签名 → 验证）
- 使用服务端日志和客户端日志对比分析

---

## 参考文档

- `memory-bank/websocket-protocol-spec.md` - WebSocket 消息协议规范
- `memory-bank/phase3-websocket/websocket-knowledge.md` - WebSocket 技术知识库
- `memory-bank/phase3-websocket/verification-result.md` - Phase 3.1 验证结果

## 问题 4：数据库路径混乱与表结构问题

### 问题描述

**现象**：
```
ERROR: 创建用户失败 {"error": "SQL logic error: table users has no column named public_key (1)"}
```

**影响**：用户注册失败，无法通过邀请码创建新用户。

---

### 问题排查

#### 1. 发现多个数据库文件

```bash
$ ls -lh server/data/*.db data/*.db
-rw-r--r--  56K  data/charline.db           # 项目根目录
-rw-r--r--  60K  server/data/charline.db    # server 子目录
-rw-r--r--  60K  server/data/server.db      # 另一个数据库
```

**问题**：存在 3 个数据库文件，不清楚 server 实际使用哪一个。

#### 2. 检查 server 实际使用的数据库

**方法 1：查看配置文件**
```bash
# server/.env
DB_PATH=./data/charline.db
```

**方法 2：使用 lsof 查看进程打开的文件**
```bash
$ lsof -p $(pgrep -f "./bin/server") | grep "\.db"
server  25595  3u  REG  /Users/liangliangtoo/code/charline/data/charline.db
```

**结论**：server 实际使用 `/Users/liangliangtoo/code/charline/data/charline.db`（项目根目录）

#### 3. 分析路径解析逻辑

**配置加载代码**（`server/internal/config/config.go:142-165`）：
```go
func findProjectRoot() string {
    dir, err := os.Getwd()
    // 向上查找 go.work 文件
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
            return dir
        }
        // ...
    }
}
```

**路径解析过程**：
1. 启动命令：`./bin/server`（从项目根目录）
2. 当前工作目录：`/Users/liangliangtoo/code/charline`
3. 配置中的相对路径：`./data/charline.db`
4. 实际解析路径：`/Users/liangliangtoo/code/charline/data/charline.db`

**根本原因**：
- 配置使用相对路径 `./data/charline.db`
- Server 从项目根目录启动
- 相对路径基于当前工作目录解析
- 导致使用项目根目录的 `data/charline.db`，而不是 `server/data/charline.db`

#### 4. 检查数据库表结构

```bash
$ sqlite3 data/charline.db "PRAGMA table_info(users);"
0|id|INTEGER|0||1
1|username|TEXT|1||0
2|token_version|INTEGER|0|1|0
3|server_url|TEXT|0||0
4|created_at|DATETIME|0|CURRENT_TIMESTAMP|0
5|last_login|DATETIME|0||0
6|public_key|TEXT|1|''|0
7|updated_at|DATETIME|0|CURRENT_TIMESTAMP|0
```

**问题**：
- `public_key` 和 `updated_at` 列存在
- 但表结构混乱（通过 ALTER TABLE 添加的列位置不正确）

#### 5. 查看表创建语句

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    token_version INTEGER DEFAULT 1,
    server_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME, public_key TEXT NOT NULL DEFAULT '', updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (username != '')
)
```

**问题**：`public_key` 和 `updated_at` 被添加到 `last_login` 行的中间，导致表结构混乱。

---

### 修复方案

#### 方案 1：清理冗余数据库文件

```bash
# 删除未使用的数据库
rm server/data/server.db

# 只保留实际使用的数据库
# data/charline.db（server 使用）
# server/data/charline.db（备份或废弃）
```

#### 方案 2：重建 users 表结构

**SQL 脚本**：
```sql
BEGIN TRANSACTION;

-- 创建新表
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    token_version INTEGER DEFAULT 1,
    server_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    public_key TEXT NOT NULL DEFAULT '',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (username != '')
);

-- 复制数据
INSERT INTO users_new (id, username, token_version, server_url, created_at, last_login, public_key, updated_at)
SELECT id, username, token_version, server_url, created_at, last_login, 
       COALESCE(public_key, ''), 
       COALESCE(updated_at, CURRENT_TIMESTAMP)
FROM users;

-- 删除旧表
DROP TABLE users;

-- 重命名新表
ALTER TABLE users_new RENAME TO users;

-- 创建索引
CREATE UNIQUE INDEX idx_users_public_key ON users(public_key) WHERE public_key != '';

COMMIT;
```

**执行方式**：
```bash
sqlite3 data/charline.db < rebuild_users.sql
```

#### 方案 3：在 server 启动时打印数据库路径

**修改代码**（`server/cmd/main.go:29-36`）：
```go
// 打印数据库配置路径
fmt.Printf("=== 配置信息 ===\n")
fmt.Printf("数据库配置路径 (cfg.DBPath): %s\n", cfg.DBPath)
dbCfg := cfg.GetDBConfig()
fmt.Printf("数据库目录 (DataDir): %s\n", dbCfg.DataDir)
fmt.Printf("数据库文件名 (Name): %s\n", dbCfg.Name)
fmt.Printf("===============\n\n")
```

**输出示例**：
```
=== 配置信息 ===
数据库配置路径 (cfg.DBPath): ./data/charline.db
数据库目录 (DataDir): data
数据库文件名 (Name): charline.db
===============

数据库连接成功 {"path": "data/charline.db"}
```

---

### 修复效果

**修复前**：
```
ERROR: 创建用户失败 {"error": "SQL logic error: table users has no column named public_key (1)"}
```

**修复后**：
- ✅ 数据库路径明确（启动时打印）
- ✅ 冗余数据库文件已清理
- ✅ users 表结构正确
- ✅ 用户注册功能正常

---

### 验证结果

```bash
# 1. 启动 server
$ ./bin/server
=== 配置信息 ===
数据库配置路径 (cfg.DBPath): ./data/charline.db
数据库目录 (DataDir): data
数据库文件名 (Name): charline.db
===============

# 2. 生成邀请码
$ curl -X POST http://localhost:8080/api/v1/invite/generate
{"code":0,"message":"成功","data":{"code":"INV-XXXXXXXX"}}

# 3. 注册用户
$ ./bin/client join INV-XXXXXXXX lucien
✓ 注册成功

# 4. 验证数据库
$ sqlite3 data/charline.db "SELECT id, username, public_key FROM users;"
1|lucien|4285b49763222b338e6e5fac034f560287e22a3b348e433e58d8eb4850302fcf
```

---

## 技术要点总结

### 4. 数据库路径管理

**问题**：相对路径基于当前工作目录解析，容易混淆

**最佳实践**：
1. **开发环境**：使用相对路径，从项目根目录启动
2. **生产环境**：使用绝对路径，避免路径混淆
3. **调试技巧**：启动时打印实际使用的数据库路径
4. **验证方法**：使用 `lsof` 查看进程实际打开的文件

**配置建议**：
```bash
# 开发环境（server/.env）
DB_PATH=./data/charline.db

# 生产环境（/etc/charline/.env）
DB_PATH=/var/lib/charline/charline.db
```

### 5. SQLite 表结构修改

**问题**：ALTER TABLE 添加的列位置不可控

**解决方案**：
1. 创建新表（正确的列顺序）
2. 复制数据到新表
3. 删除旧表
4. 重命名新表

**注意事项**：
- 使用事务保证原子性
- 使用 COALESCE 处理 NULL 值
- 重建索引和约束
- 备份原始数据

---

## 修改文件清单（完整）

| 文件 | 修改内容 | 状态 |
|------|---------|------|
| `client/internal/websocket/client.go` | 添加 authChan，修改消息分发逻辑 | ✅ 已完成 |
| `client/internal/auth/signer.go` | 签名编码从 base64 改为 hex | ✅ 已完成 |
| `server/cmd/main.go` | 添加数据库路径打印 | ✅ 已完成 |
| `data/charline.db` | 重建 users 表结构 | ✅ 已完成 |
| `server/data/server.db` | 删除冗余数据库文件 | ✅ 已完成 |

---

## 部署建议

### Docker 部署时的目录结构

**容器内部**：
```
/app/
├── server              # 可执行文件
├── data/               # 数据挂载点
├── logs/               # 日志挂载点
└── config/             # 配置挂载点
```

**宿主机**：
```
/data/charline/
├── data/               # 持久化数据
│   └── charline.db
├── logs/               # 持久化日志
└── config/             # 配置文件
    └── .env
```

**配置**：
```bash
# 容器内使用绝对路径
DB_PATH=/app/data/charline.db
```

---

## 经验教训（补充）

### 5. 数据库管理

- ❌ **错误做法**：使用相对路径 + 多个数据库文件
- ✅ **正确做法**：明确数据库路径，启动时打印验证

### 6. 表结构迁移

- ❌ **错误做法**：直接 ALTER TABLE 添加列
- ✅ **正确做法**：重建表保证结构正确

### 7. 调试技巧（补充）

- 使用 `lsof` 查看进程实际打开的文件
- 启动时打印关键配置路径
- 定期清理冗余文件
