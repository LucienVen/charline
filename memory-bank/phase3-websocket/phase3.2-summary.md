# Phase 3.2 实施总结

## 提交信息
- **Commit**: c7433f1
- **日期**: 2026-02-02
- **标题**: fix: Phase 3.2 WebSocket 消息读取竞争与数据库问题修复

## 核心问题与解决方案

### 问题 1：WebSocket 消息读取竞争
**现象**: 客户端连接后打印 "Received message: challenge_response" 然后挂起

**根本原因**: 
- `readLoop` goroutine 和 `Authenticate()` 方法同时从 WebSocket 连接读取消息
- `readLoop` 先消费了 challenge 消息
- `Authenticate()` 永久阻塞等待

**解决方案**:
- 添加 `authChan chan Message` 用于认证消息分发
- `readLoop` 统一读取所有消息，按类型路由到不同 channel
- `Authenticate()` 从 `authChan` 接收消息，使用 `select` + `timeout` 防止阻塞

**修改文件**:
- `client/internal/websocket/client.go`: 添加 authChan 和消息分发逻辑

---

### 问题 2：签名编码不一致
**现象**: `authentication failed: INVALID_SIGNATURE - Invalid signature format`

**根本原因**:
- 客户端: `base64.StdEncoding.EncodeToString(signature)`
- 服务端: `hex.DecodeString(authReq.Signature)`
- 编码格式不匹配

**解决方案**:
- 统一使用 hex 编码
- 修改客户端签名生成: `hex.EncodeToString(signature)`

**修改文件**:
- `client/internal/auth/signer.go`: 改用 hex 编码

---

### 问题 3：数据库路径混乱
**现象**: 修改了 `server/data/charline.db`，但 server 仍报错 "no such column: public_key"

**根本原因**:
- 存在多个数据库文件: `data/charline.db`, `server/data/charline.db`, `server/data/server.db`
- 配置使用相对路径 `./data/charline.db`
- Server 从项目根目录启动，实际使用 `/Users/liangliangtoo/code/charline/data/charline.db`
- 一直在修改错误的数据库文件

**解决方案**:
1. 使用 `lsof` 确认 server 实际打开的数据库文件
2. 在 server 启动时打印数据库路径
3. 删除冗余数据库文件 `server/data/server.db`
4. 修改正确的数据库文件

**修改文件**:
- `server/cmd/main.go`: 添加数据库路径打印（第29-36行）

---

### 问题 4：数据库表结构混乱
**现象**: 通过 `ALTER TABLE` 添加的列位置不正确

**根本原因**:
- SQLite 的 `ALTER TABLE ADD COLUMN` 只能在表末尾添加列
- 多次 ALTER TABLE 导致列顺序混乱
- 表结构: `last_login DATETIME, public_key TEXT NOT NULL DEFAULT '', updated_at DATETIME`

**解决方案**:
- 使用事务重建表结构
- 创建新表 → 复制数据 → 删除旧表 → 重命名新表
- 确保列顺序正确

**执行命令**:
```sql
BEGIN TRANSACTION;
CREATE TABLE users_new (...);
INSERT INTO users_new SELECT ...;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
CREATE UNIQUE INDEX idx_users_public_key ON users(public_key);
COMMIT;
```

---

## Phase 3.2 Session 管理实现

### 核心功能
1. **Session 生命周期管理**
   - 创建: 认证成功后创建 Session
   - 查询: 支持按 ID、UserID、ConnID 查询
   - 挂起: 断线时生成 Resume Token
   - 恢复: 使用 Resume Token 恢复 Session
   - 关闭: 正常关闭时清理资源

2. **Resume Token 机制**
   - 32 字节随机 Token（base64 编码）
   - 30 秒过期时间
   - 原子消费（一次性使用）
   - 自动清理过期 Token

3. **多索引存储**
   - 按 Session ID 索引
   - 按 User ID 索引（支持多设备）
   - 按 Connection ID 索引

4. **后台清理**
   - 定期清理过期 Session（每分钟）
   - 清理过期 Resume Token

### 新增文件
- `server/internal/session/session.go` (128 行): Session 结构体和状态管理
- `server/internal/session/store.go` (243 行): 内存存储实现
- `server/internal/session/resume.go` (180 行): Resume Token 管理
- `server/internal/session/manager.go` (274 行): SessionManager 实现
- `client/internal/session/state.go` (87 行): 客户端 Session 状态

### 集成修改
- `server/internal/websocket/handler.go`: 认证成功后创建 Session
- `server/internal/websocket/conn.go`: 添加 sessionID 字段
- `server/internal/websocket/protocol.go`: AuthResponse 添加 session_id
- `server/internal/container/container.go`: 添加 SessionManager 依赖注入

---

## 文档更新

### 新增文档
- `memory-bank/phase3-websocket/troubleshooting.md` (833 行)
  - 问题 1: WebSocket 消息读取竞争
  - 问题 2: 签名编码不一致
  - 问题 3: 数据库路径混乱
  - 问题 4: 数据库表结构问题
  - 技术要点总结
  - 经验教训

### 更新文档
- `memory-bank/progress.md` (407 行新增)
  - Phase 3.2 问题修复记录
  - 数据库问题修复详情

---

## 统计数据

### 代码变更
- **修改文件**: 15 个
- **新增代码**: 2264 行
- **删除代码**: 72 行
- **净增加**: 2192 行

### 文件分类
- **核心功能**: 5 个新文件（Session 管理）
- **问题修复**: 3 个文件（消息分发、签名编码、启动日志）
- **集成修改**: 5 个文件（WebSocket、Container、Config）
- **文档更新**: 2 个文件

---

## 验证结果

### 功能验证
✅ WebSocket 连接建立成功
✅ Challenge-Response 认证流程正常
✅ 签名生成和验证通过
✅ Session 创建成功
✅ 数据库路径明确可见
✅ 用户注册功能正常

### 流程验证
| 步骤 | 状态 | 说明 |
|------|------|------|
| 1. WebSocket 连接 | ✅ | 成功建立连接 |
| 2. 接收 Challenge | ✅ | 正常接收 nonce |
| 3. 签名生成 | ✅ | hex 编码正确 |
| 4. 发送认证请求 | ✅ | 消息发送成功 |
| 5. 签名验证 | ✅ | 服务端验证通过 |
| 6. Nonce 验证 | ✅ | 防重放检查通过 |
| 7. Session 创建 | ✅ | Session ID 返回 |
| 8. 用户注册 | ✅ | 邀请码激活成功 |

---

## 技术要点

### 1. 并发消息处理
**架构**:
```
WebSocket Connection
        ↓
   readLoop (统一读取)
        ↓
  handleMessage (消息路由)
        ↓
   ┌────┴────┐
   ↓         ↓
authChan   其他 channel
   ↓
Authenticate()
```

**关键点**:
- 单一读取点，避免竞争
- Channel 分发，类型路由
- Select + Timeout，防止阻塞

### 2. 编码一致性
| 数据类型 | 原始大小 | hex 编码 | base64 编码 |
|---------|---------|---------|------------|
| Ed25519 签名 | 64 字节 | 128 字符 | 88 字符 |
| Ed25519 公钥 | 32 字节 | 64 字符 | 44 字符 |

**选择 hex 的原因**:
- 可读性好（纯字母数字）
- 长度固定可预测
- 服务端已使用 hex

### 3. 数据库路径管理
**开发环境**:
- 配置: `DB_PATH=./data/charline.db`
- 启动: `./bin/server`（从项目根目录）
- 实际路径: `/Users/liangliangtoo/code/charline/data/charline.db`

**生产环境建议**:
- 使用绝对路径: `DB_PATH=/var/lib/charline/charline.db`
- 或明确相对路径: `DB_PATH=./server/data/charline.db`

### 4. SQLite 表结构修改
**错误做法**:
```sql
ALTER TABLE users ADD COLUMN public_key TEXT;
ALTER TABLE users ADD COLUMN updated_at DATETIME;
-- 列位置不可控
```

**正确做法**:
```sql
BEGIN TRANSACTION;
CREATE TABLE users_new (...);  -- 正确的列顺序
INSERT INTO users_new SELECT ...;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
COMMIT;
```

---

## 经验教训

### 1. 并发编程
- ❌ 多个 goroutine 直接读取同一资源
- ✅ 使用 channel 进行消息传递和同步

### 2. 协议一致性
- ❌ 客户端和服务端使用不同编码
- ✅ 协议设计阶段明确编码格式

### 3. 调试技巧
- 使用详细的 DEBUG 日志
- 在关键节点打印状态
- 使用 `lsof` 查看进程实际打开的文件
- 启动时打印关键配置路径

### 4. 数据库管理
- ❌ 使用相对路径 + 多个数据库文件
- ✅ 明确数据库路径，启动时验证

### 5. 表结构迁移
- ❌ 直接 ALTER TABLE 添加列
- ✅ 重建表保证结构正确

---

## 下一步计划

### Phase 3.3: 消息队列与离线消息
- 实现消息队列机制
- 支持离线消息存储
- 实现消息确认机制

### Phase 3.4: 心跳与断线重连
- 实现心跳机制
- 实现断线重连逻辑
- 使用 Resume Token 恢复 Session

---

## 参考文档
- `memory-bank/phase3-websocket/troubleshooting.md` - 详细问题排查
- `memory-bank/progress.md` - 项目进度跟踪
- `memory-bank/websocket-protocol-spec.md` - WebSocket 协议规范
