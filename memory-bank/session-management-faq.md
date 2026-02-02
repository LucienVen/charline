# Session 管理常见问题解答（FAQ）

> **文档版本**: v1.0  
> **创建时间**: 2026-02-02  
> **作者**: claude code  
> **适用阶段**: Phase 3.2 - Session 管理层

---

## 一、多设备与 Session 关系

### Q1: 同个用户，多设备，是否使用同一个 Session？

**答案：不是，每个设备使用独立的 Session**

#### 设计理由

1. **每个设备有独立的 WebSocket 连接**
   - 手机 APP → WebSocket Conn A → Session A
   - 电脑客户端 → WebSocket Conn B → Session B
   - 网页版 → WebSocket Conn C → Session C

2. **Session 与 Connection 绑定**
   ```go
   type Session struct {
       ID       string
       UserID   int64
       DeviceID string  // 设备标识
       ConnID   string  // WebSocket 连接 ID（一对一绑定）
       // ...
   }
   ```

3. **断线恢复需要设备级别的 Resume Token**
   - 手机断线 → 使用 Session A 的 Resume Token 恢复
   - 电脑断线 → 使用 Session B 的 Resume Token 恢复
   - 互不干扰

#### 数据结构设计

```go
// 正确设计：支持多设备
userSessions map[int64][]string  // userID -> []sessionID

// 示例数据
userSessions[123] = [
    "session-mobile-abc",    // 手机
    "session-desktop-def",   // 电脑
    "session-web-ghi"        // 网页
]
```

#### 实际场景

```
用户 ID: 123
├─ Session A (手机)
│   ├─ DeviceID: "mobile-device-001"
│   ├─ ConnID: "ws-conn-abc"
│   └─ ResumeToken: "resume-token-mobile"
├─ Session B (电脑)
│   ├─ DeviceID: "desktop-device-002"
│   ├─ ConnID: "ws-conn-def"
│   └─ ResumeToken: "resume-token-desktop"
└─ Session C (网页)
    ├─ DeviceID: "web-device-003"
    ├─ ConnID: "ws-conn-ghi"
    └─ ResumeToken: "resume-token-web"
```

---

### Q2: 假设场景，用户多个设备登录，是否能都收到信息？

**答案：是的，所有设备都能收到消息**

#### 实现机制

```go
// 给用户 123 发送消息
func SendMessageToUser(userID int64, message string) {
    // 1. 获取该用户的所有 Session（所有在线设备）
    sessions := sessionStore.GetByUser(userID)
    
    // 2. 遍历所有 Session，给每个设备发送消息
    for _, sess := range sessions {
        if sess.IsActive() {
            // 3. 通过 ConnID 获取 WebSocket 连接
            conn := connectionPool.Get(sess.ConnID)
            
            // 4. 发送消息
            conn.Send(message)
        }
    }
}
```

#### 消息广播流程

```
服务端收到：给用户 123 发送消息 "Hello"
        ↓
SessionStore.GetByUser(123)
返回: [Session A, Session B, Session C]
        ↓
    ┌───────┼───────┐
    ↓       ↓       ↓
Session A  Session B  Session C
(手机)     (电脑)     (网页)
    ↓       ↓       ↓
Conn ABC  Conn DEF  Conn GHI
    ↓       ↓       ↓
手机收到   电脑收到   网页收到
"Hello"   "Hello"   "Hello"
```

---

## 二、Resume Token 机制

### Q3: Resume Token 管理的作用是什么？

**答案：Resume Token 是用于断线恢复的临时凭证**，让客户端在短时间内快速重连，无需重新走完整认证流程。

#### 核心优势

| 维度 | 传统方案（无 Resume Token） | 优化方案（有 Resume Token） |
|------|---------------------------|---------------------------|
| **重连流程** | 完整认证（Challenge + 签名） | 直接使用 Token 恢复 |
| **耗时** | 1-3 秒 | 100-300 毫秒 |
| **用户体验** | 明显卡顿 | 几乎无感知 |
| **服务端压力** | 需要验证签名 | 只需验证 Token |

#### 生命周期

```
1. 生成时机：WebSocket 连接意外断开
   ↓
2. 有效期：30 秒
   ↓
3. 使用时机：客户端在 30 秒内重连
   ↓
4. 消费方式：原子操作，只能用一次
   ↓
5. 过期/撤销：30 秒后自动过期，或主动撤销
```

#### 关键特性

1. **原子消费（只能用一次）**
   ```go
   func (r *ResumeTokenStore) Consume(token string) (string, error) {
       // 检查 Token 是否存在和有效
       // ...
       
       // ⚠️ 关键：立即删除，防止重放攻击
       delete(r.tokens, token)
       
       return sessionID, nil
   }
   ```

2. **短有效期（30 秒）**
   - ✅ 足够长：覆盖大部分网络抖动场景
   - ✅ 足够短：降低安全风险

3. **与 Session 绑定**
   - 一个 Session 同时只能有一个有效的 Resume Token

---

### Q4: 如何检测断线？

**答案：采用三层检测机制**

#### Layer 1: 应用层心跳（主动检测）

```go
const (
    PingInterval   = 30 * time.Second  // 心跳间隔
    PongTimeout    = 10 * time.Second  // Pong 响应超时
    MaxMissedPings = 3                 // 最大允许丢失次数
)

// 心跳检测流程
0s    → 发送 Ping #1
      ← 收到 Pong #1 (正常)
      
30s   → 发送 Ping #2
      ← 收到 Pong #2 (正常)
      
60s   → 发送 Ping #3
      ✕ 未收到 Pong (超时 10s)
      ⚠️ missedPings = 1
      
90s   → 发送 Ping #4
      ✕ 未收到 Pong (超时 10s)
      ⚠️ missedPings = 2
      
120s  → 发送 Ping #5
      ✕ 未收到 Pong (超时 10s)
      ⚠️ missedPings = 3
      
      ❌ 达到最大丢失次数 → 判定断线
```

#### Layer 2: 读写错误检测（被动检测）

```go
func (c *Connection) ReadLoop() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            // ⚠️ 被动检测：读取错误 = 连接断开
            c.handleDisconnect()
            return
        }
        c.handleMessage(message)
    }
}
```

**触发场景：**
- 客户端主动关闭连接
- 网络物理断开
- 客户端进程崩溃
- 防火墙/代理关闭连接

#### Layer 3: TCP KeepAlive（操作系统层）

```go
// 启用 TCP KeepAlive
tcpConn.SetKeepAlive(true)
tcpConn.SetKeepAlivePeriod(30 * time.Second)

// TCP KeepAlive 参数（Linux）
tcp_keepalive_time = 30s   // 空闲 30 秒后开始发送探测包
tcp_keepalive_intvl = 10s  // 探测包间隔 10 秒
tcp_keepalive_probes = 3   // 最多发送 3 个探测包

总超时时间 = 30s + (10s × 3) = 60s
```

#### 不同场景的检测时间

| 场景 | 检测方式 | 检测时间 |
|------|---------|---------|
| **客户端主动关闭** | 读写错误 | 立即（< 1 秒） |
| **网络物理断开** | 心跳超时 | 90-120 秒 |
| **客户端进程崩溃** | 心跳超时 | 90-120 秒 |
| **WiFi 切换** | 读写错误 + 心跳 | 5-30 秒 |
| **移动网络切换** | 读写错误 + 心跳 | 10-60 秒 |
| **防火墙关闭连接** | TCP KeepAlive | 60 秒 |

---

### Q5: 如何区分意外断线和用户主动断线？

**答案：通过 WebSocket 关闭码区分**

#### 方案：WebSocket 关闭码（标准方法）

```go
func (c *Connection) handleCloseError(closeErr *websocket.CloseError) {
    switch closeErr.Code {
    case websocket.CloseNormalClosure:
        // 1000: 正常关闭
        c.handleNormalClose()  // ❌ 不生成 Resume Token
        
    case websocket.CloseGoingAway:
        // 1001: 客户端离开（关闭页面/APP）
        c.handleNormalClose()  // ❌ 不生成 Resume Token
        
    case websocket.CloseAbnormalClosure:
        // 1006: 异常关闭（TCP 连接断开）
        c.handleUnexpectedDisconnect()  // ✅ 生成 Resume Token
        
    case websocket.CloseNoStatusReceived:
        // 1005: 未收到状态码
        c.handleUnexpectedDisconnect()  // ✅ 生成 Resume Token
        
    default:
        c.handleUnexpectedDisconnect()
    }
}
```

#### WebSocket 关闭码对照表

| 关闭码 | 名称 | 含义 | 处理方式 |
|--------|------|------|---------|
| **1000** | Normal Closure | 正常关闭 | ❌ 不生成 Resume Token |
| **1001** | Going Away | 客户端离开（关闭页面/APP） | ❌ 不生成 Resume Token |
| **1002** | Protocol Error | 协议错误 | ❌ 不生成 Resume Token |
| **1005** | No Status Received | 未收到状态码 | ✅ 生成 Resume Token |
| **1006** | Abnormal Closure | 异常关闭（TCP 连接断开） | ✅ 生成 Resume Token |
| **1011** | Internal Error | 服务端内部错误 | ✅ 生成 Resume Token |

#### 客户端实现

**主动关闭（不需要重连）：**

```go
// 用户主动登出
func (c *Client) Logout() error {
    // 1. 发送 close 消息
    closeMsg := protocol.NewMessage("close", map[string]interface{}{
        "reason": "user_logout",
    })
    c.Send(closeMsg)
    
    // 2. 使用正常关闭码关闭连接
    c.conn.WriteMessage(
        websocket.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, "logout"),
    )
    
    // 3. 清空 Resume Token（不需要重连）
    c.resumeToken = ""
    c.conn.Close()
    
    return nil
}
```

**意外断线（需要重连）：**

```go
func (c *Client) ReadLoop() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            // 检查是否是正常关闭
            if closeErr, ok := err.(*websocket.CloseError); ok {
                if closeErr.Code == websocket.CloseNormalClosure {
                    return  // 正常关闭，不重连
                }
            }
            
            // 意外断线，启动自动重连
            go c.autoReconnect()
            return
        }
        c.handleMessage(message)
    }
}
```

#### 完整流程对比

**场景 A: 用户主动登出**

```
客户端                          服务端
  │                              │
  │─────close 消息────────────────►│
  │  (reason: user_logout)       │
  │                              │
  │─────CloseNormalClosure───────►│
  │  (关闭码: 1000)              │
  │                              │
  │                              ├─检测到正常关闭
  │                              ├─❌ 不生成 Resume Token
  │                              ├─关闭 Session (CLOSED)
  │                              └─清理资源
  │                              
  ❌ 不尝试重连
```

**场景 B: 意外断线（网络抖动）**

```
客户端                          服务端
  │                              │
  ✕ 网络断开（没有发送任何消息）  │
  │                              │
  │                              ├─检测到异常关闭 (1006)
  │                              ├─✅ 生成 Resume Token
  │                              ├─Suspend Session
  │                              └─等待重连（30秒）
  │                              
  │─────重新连接──────────────────►│
  │─────Resume Request───────────►│
  │  (resume_token: xxx)         │
  │                              ├─验证 Token
  │                              ├─恢复 Session (ACTIVE)
  │◄────Resume Response──────────┤
  │                              
  ✅ 恢复正常通信
```

---

## 三、设计决策总结

### 判断逻辑

```go
if 收到 close 消息 || 关闭码 == 1000/1001 {
    // 正常关闭
    ❌ 不生成 Resume Token
    直接关闭 Session
} else {
    // 意外断线
    ✅ 生成 Resume Token
    Suspend Session
    等待重连（30秒）
}
```

### 推荐方案

**两者结合：**
1. **优先使用 WebSocket 关闭码**（标准方法）
2. **辅助使用 close 消息**（提供更多上下文）

---

## 四、参考文档

- [架构设计文档](./architecture.md)
- [WebSocket 协议规范](./websocket-protocol-spec.md)
- [Phase 3.1 实施计划](../nimbalyst-local/plans/phase3-1-websocket-basic-connection.md)
- [IM 系统认证策略挑战回复](./thinking-challenge/response-to-auth-challenge.md)

---

**文档维护者**: claude code  
**最后更新**: 2026-02-02
