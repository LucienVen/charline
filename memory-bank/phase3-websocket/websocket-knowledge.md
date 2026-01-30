# WebSocket 技术知识库

本文档记录 Charline 项目中 WebSocket 实现的核心技术细节和常见问题解答。

---

## 一、WebSocket 连接池设计

### 1.1 ConnectionPool 结构

```go
type ConnectionPool struct {
    connections map[int64]*Connection  // userID -> Connection
    mu          sync.RWMutex
}
```

**设计特点：**
- **动态管理**：不是预分配的连接池，而是连接管理器
- **按需创建**：每个 WebSocket 连接建立时才创建 Connection 对象
- **动态扩展**：Go 的 map 会自动扩容，无需预先指定容量
- **初始状态**：`Count()` 在初始化时返回 0 是正常的

**为什么不需要预分配容量？**
1. Go 的 map 是动态数据结构，会根据元素数量自动扩容
2. WebSocket 连接是长连接，连接数相对稳定
3. 预分配反而可能浪费内存（如果实际连接数远小于预分配数）
4. 动态扩容的性能开销可以忽略不计

---

## 二、HTTP 升级为 WebSocket 的原理

### 2.1 协议升级机制（RFC 7230）

HTTP/1.1 支持 **协议升级（Protocol Upgrade）** 机制，允许客户端和服务器协商切换到其他协议。

**升级流程：**

```
客户端                                    服务器
  |                                         |
  |--- HTTP GET /ws ----------------------->|
  |    Upgrade: websocket                   |
  |    Connection: Upgrade                  |
  |    Sec-WebSocket-Key: xxx               |
  |                                         |
  |<-- HTTP 101 Switching Protocols --------|
  |    Upgrade: websocket                   |
  |    Connection: Upgrade                  |
  |    Sec-WebSocket-Accept: yyy            |
  |                                         |
  |<======= WebSocket 双向通信 ============>|
```

### 2.2 关键 HTTP 头

**客户端请求头：**
```http
GET /ws HTTP/1.1
Host: localhost:8080
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
```

**服务器响应头：**
```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

### 2.3 代码实现（gorilla/websocket）

```go
// server/internal/router/router.go
r.Get("/ws", routes.WSServer.HandleConnection)

// server/internal/websocket/server.go
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // 升级 HTTP 连接为 WebSocket
    wsConn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return  // 升级失败
    }
    
    // 创建连接封装
    conn := NewConnection(wsConn)
    
    // 启动读写循环
    go conn.ReadLoop(s.handler)
    go conn.WriteLoop()
}
```

**`upgrader.Upgrade()` 做了什么？**
1. 验证 HTTP 请求头（Upgrade、Connection、Sec-WebSocket-Key）
2. 计算 `Sec-WebSocket-Accept` 响应值
3. 发送 `101 Switching Protocols` 响应
4. **劫持（Hijack）底层 TCP 连接**
5. 返回 `*websocket.Conn` 对象，用于后续 WebSocket 通信

### 2.4 TCP 连接劫持

```go
// gorilla/websocket 内部实现（简化）
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Conn, error) {
    // 1. 验证请求头
    if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
        return nil, errors.New("missing Connection: Upgrade")
    }
    
    // 2. 劫持 TCP 连接
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        return nil, errors.New("response does not implement http.Hijacker")
    }
    
    netConn, brw, err := hijacker.Hijack()
    if err != nil {
        return nil, err
    }
    
    // 3. 发送 101 响应
    brw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
    brw.WriteString("Upgrade: websocket\r\n")
    brw.WriteString("Connection: Upgrade\r\n")
    brw.WriteString("Sec-WebSocket-Accept: " + computeAcceptKey(challengeKey) + "\r\n")
    brw.WriteString("\r\n")
    brw.Flush()
    
    // 4. 返回 WebSocket 连接对象
    return &Conn{conn: netConn}, nil
}
```

**关键点：**
- `http.Hijacker` 接口允许接管底层 TCP 连接
- 一旦劫持成功，HTTP 层不再处理这个连接
- 后续所有数据都通过 WebSocket 协议传输

---

## 三、Goroutine 管理

### 3.1 每个连接的 Goroutine 数量

**答案：每个 WebSocket 连接创建 2 个 goroutine**

```go
// server/internal/websocket/server.go:54-55
go conn.ReadLoop(s.handler)  // 协程 1：读取循环
go conn.WriteLoop()           // 协程 2：写入循环
```

**为什么需要两个 goroutine？**
1. **读写分离**：避免读写操作相互阻塞
2. **并发处理**：可以同时接收和发送消息
3. **避免死锁**：如果用单个 goroutine，可能在等待读取时无法发送
4. **性能优化**：充分利用多核 CPU

**资源影响：**
- 1000 个并发连接 = 2000 个 goroutine
- 每个 goroutine 约 2KB 栈空间
- 2000 个 goroutine ≈ 4MB 内存（栈空间）
- Go 的 goroutine 非常轻量，这个开销是可接受的

### 3.2 ReadLoop 实现

```go
// server/internal/websocket/conn.go
func (c *Connection) ReadLoop(handler MessageHandler) {
    defer func() {
        c.Close()
    }()
    
    // 设置读取超时和心跳
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        select {
        case <-c.closeChan:
            return
        default:
            // 读取 WebSocket 消息
            _, data, err := c.conn.ReadMessage()
            if err != nil {
                return
            }
            
            // 解析并处理消息
            msg, err := Unmarshal(data)
            if err != nil {
                // 发送错误响应
                continue
            }
            
            // 处理消息
            handler.HandleMessage(c, msg)
        }
    }
}
```

**关键点：**
- `c.conn.ReadMessage()` 是阻塞调用，等待 WebSocket 帧到达
- 不经过 HTTP 层，直接从 TCP 连接读取
- 支持心跳机制（Ping/Pong）
- 连接关闭时自动退出循环

### 3.3 WriteLoop 实现

```go
// server/internal/websocket/conn.go
func (c *Connection) WriteLoop() {
    ticker := time.NewTicker(54 * time.Second)  // 心跳间隔
    defer func() {
        ticker.Stop()
        c.Close()
    }()
    
    for {
        select {
        case <-c.closeChan:
            // 发送关闭消息
            c.conn.WriteMessage(websocket.CloseMessage, []byte{})
            return
            
        case data := <-c.send:
            // 设置写入超时
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            
            // 发送消息
            if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
                return
            }
            
        case <-ticker.C:
            // 发送心跳 Ping
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

**关键点：**
- 通过 channel (`c.send`) 接收要发送的消息
- 定时发送心跳 Ping（54 秒间隔）
- 支持优雅关闭（发送 CloseMessage）

---

## 四、路由与消息处理

### 4.1 连接生命周期

```
【第一次连接】
客户端 → HTTP GET /ws → chi 路由 → HandleConnection → 协议升级 → sendChallenge → ReadLoop/WriteLoop 启动

【后续消息】
客户端 → WebSocket 帧 → TCP 连接 → ReadLoop 直接读取 → 处理消息
                                    ↑
                            不经过 HTTP 层
                            不经过 chi 路由
                            不触发 HandleConnection
```

### 4.2 是否每次都走路由？

**答案：不会。**

| 阶段 | 是否经过 chi 路由 | 是否触发 sendChallenge | 处理方式 |
|------|------------------|----------------------|---------|
| 第一次连接 | ✅ 是 | ✅ 是（仅一次） | HTTP 升级 → WebSocket |
| 后续消息 | ❌ 否 | ❌ 否 | WebSocket 帧 → ReadLoop |

**原因：**
1. **协议升级后，连接不再是 HTTP**
2. 后续消息是 **WebSocket 帧**，不是 HTTP 请求
3. WebSocket 帧直接通过 TCP 连接传输
4. `ReadLoop` 直接调用 `c.conn.ReadMessage()` 读取帧
5. **完全绕过 HTTP 层和路由层**

### 4.3 sendChallenge 只执行一次

```go
// server/internal/websocket/server.go:47-51
// 发送初始挑战（nonce）
if err := s.sendChallenge(conn); err != nil {
    conn.Close()
    return
}
```

**执行时机：**
- 仅在 `HandleConnection` 中执行
- 仅在协议升级成功后执行
- 每个连接只执行一次
- 后续消息不会再触发

**认证流程：**
```
1. 客户端连接 /ws
2. 服务器发送 challenge_response（包含 nonce）
3. 客户端签名 nonce，发送 auth_request
4. 服务器验证签名，发送 auth_response
5. 认证成功，连接保持打开
6. 后续消息直接通过 WebSocket 帧传输
```

---

## 五、WebSocket 长连接特性

### 5.1 为什么是"长连接"？

1. **一次握手，持久连接**
   - HTTP 升级只发生一次
   - 连接建立后保持打开状态
   - 不需要重复建立连接

2. **双向通信**
   - 服务器可以主动推送消息
   - 客户端可以随时发送消息
   - 不需要轮询（polling）

3. **低延迟**
   - 消息直接通过 TCP 传输
   - 无 HTTP 头开销
   - 无连接建立开销

4. **状态保持**
   - 连接关联用户 ID（`conn.userID`）
   - 认证状态持久化
   - 不需要每次请求都认证

### 5.2 连接关闭场景

连接会在以下情况关闭：
1. 客户端主动断开
2. 服务器主动关闭
3. 网络故障
4. 读写超时
5. 心跳超时（60 秒无 Pong 响应）

---

## 六、性能考虑

### 6.1 Goroutine 数量控制

**当前设计：**
- 每个连接 2 个 goroutine
- 1000 个连接 = 2000 个 goroutine

**是否需要优化？**
- Go 的 goroutine 非常轻量（2KB 栈）
- 2000 个 goroutine ≈ 4MB 内存
- 对于大多数应用，这个开销可以接受

**如果需要支持 10 万+ 连接：**
- 考虑使用 goroutine 池
- 考虑使用 epoll/kqueue（如 gnet 库）
- 考虑使用消息队列解耦

### 6.2 内存管理

**当前设计：**
```go
type Connection struct {
    conn      *websocket.Conn
    userID    int64
    send      chan []byte  // 缓冲区 256
    closeChan chan struct{}
    closeOnce sync.Once
    isClosed  bool
    mu        sync.RWMutex
}
```

**内存占用估算（每个连接）：**
- `*websocket.Conn`: ~1KB
- `send` channel (256 buffer): ~256 * 平均消息大小
- 其他字段: ~100 bytes
- **总计**: ~1-2KB + 消息缓冲区

**优化建议：**
- 如果消息量大，考虑调整 `send` channel 缓冲区大小
- 如果连接数多，考虑使用对象池（sync.Pool）

### 6.3 心跳机制

**当前设计：**
- 服务器每 54 秒发送 Ping
- 客户端收到 Ping 后自动回复 Pong
- 服务器 60 秒未收到 Pong 则关闭连接

**为什么需要心跳？**
1. 检测连接是否存活
2. 防止 NAT 超时
3. 防止防火墙关闭空闲连接
4. 及时清理僵尸连接

---

## 七、常见问题

### Q1: 为什么初始化时 `Count()` 返回 0？

**A:** ConnectionPool 是连接管理器，不是预分配池。连接在建立时才添加到 pool 中。

### Q2: 是否需要预分配连接池容量？

**A:** 不需要。Go 的 map 会自动扩容，预分配反而可能浪费内存。

### Q3: 每个 /ws 请求都会创建 2 个 goroutine 吗？

**A:** 是的。每个 WebSocket 连接创建 ReadLoop 和 WriteLoop 两个 goroutine。

### Q4: 后续消息还会走 chi 路由吗？

**A:** 不会。协议升级后，消息通过 WebSocket 帧传输，直接由 ReadLoop 处理，不经过 HTTP 层。

### Q5: sendChallenge 会每次都执行吗？

**A:** 不会。只在连接建立时执行一次，后续消息不会触发。

### Q6: 如何支持更多并发连接？

**A:** 
- 当前设计可支持数千到数万连接
- 如需支持 10 万+ 连接，考虑：
  - 使用 goroutine 池
  - 使用 epoll/kqueue（如 gnet）
  - 使用消息队列解耦
  - 水平扩展（多实例 + 负载均衡）

---

## 八、参考资料

- [RFC 6455 - The WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [RFC 7230 - HTTP/1.1 Protocol Upgrade](https://tools.ietf.org/html/rfc7230#section-6.7)
- [gorilla/websocket 文档](https://pkg.go.dev/github.com/gorilla/websocket)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)

---

**最后更新：** 2026-01-30  
**维护者：** Charline 开发团队
