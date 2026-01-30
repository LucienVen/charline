# WebSocket 连接问题排查全过程

## 问题背景

**目标**: 验证 Phase 3.1 WebSocket 基础连接功能是否正常工作

**初始状态**: 
- 服务端和客户端代码已实现
- 需要验证 WebSocket 连接能否成功建立
- 需要验证服务器能否发送挑战消息



[toc]





---

## 排查过程

### 第一步: 初始测试

#### 1.1 编译服务器
```bash
cd /Users/liangliangtoo/code/charline/server
go build -o /tmp/charline-server cmd/main.go
```

**结果**: ✅ 编译成功

#### 1.2 启动服务器
```bash
JWT_SECRET=test-secret-key-with-at-least-32-bytes-length /tmp/charline-server
```

**观察日志**:
```
+0800 2026-01-30 15:41:46  INFO  验证器初始化成功
+0800 2026-01-30 15:41:46  INFO  数据库连接成功
+0800 2026-01-30 15:41:46  INFO  WebSocket 服务器初始化成功
+0800 2026-01-30 15:41:46  INFO  Server started  {"address": ":8080"}
+0800 2026-01-30 15:41:46  INFO  === Registered Routes ===
  GET    /ws
```

**结果**: ✅ 服务器启动成功，`/ws` 路由已注册

#### 1.3 测试 WebSocket 连接
```go
conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
```

**结果**: ❌ 连接失败
```
websocket: bad handshake
```

**服务器日志**:
```
+0800 2026-01-30 15:42:11  ERROR  HTTP request  {"method": "GET", "path": "/ws", "status": 500}
```

**初步判断**: 
- 路由注册正常
- 请求到达服务器
- 服务器返回 500 错误
- 没有 panic 日志

---

### 第二步: 分析问题

#### 2.1 检查代码结构

**查看 HandleConnection 方法**:

```go
// server/internal/websocket/server.go
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // 升级 HTTP 连接为 WebSocket
    wsConn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        // 升级失败，upgrader 已经发送了错误响应
        return  // ⚠️ 没有日志输出
    }
    
    // 创建连接封装
    conn := NewConnection(wsConn)
    
    // 发送初始挑战（nonce）
    if err := s.sendChallenge(conn); err != nil {
        conn.Close()
        return  // ⚠️ 没有日志输出
    }
    
    // 启动读写循环
    go conn.ReadLoop(s.handler)
    go conn.WriteLoop()
}
```

**发现问题 1**: 
- `HandleConnection` 中没有任何日志输出
- 错误被静默吞掉
- 无法定位具体失败点

#### 2.2 添加调试日志

**修改代码添加日志**:
```go
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
    fmt.Println("[DEBUG] HandleConnection called")
    
    wsConn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Printf("[ERROR] WebSocket upgrade failed: %v\n", err)
        return
    }
    fmt.Println("[DEBUG] WebSocket upgrade successful")
    
    conn := NewConnection(wsConn)
    fmt.Println("[DEBUG] Connection created")
    
    // ... 后续步骤也添加日志
}
```

#### 2.3 重新测试

**重新编译并启动服务器**:
```bash
go build -o /tmp/charline-server cmd/main.go
JWT_SECRET=test-secret-key-with-at-least-32-bytes-length /tmp/charline-server
```

**再次测试连接**:
```bash
go run test_ws.go
```

**新的日志输出**:
```
[DEBUG] HandleConnection called
[ERROR] WebSocket upgrade failed: websocket: response does not implement http.Hijacker
```

**关键发现**: 
- 问题定位到 WebSocket 升级阶段
- 错误信息: `response does not implement http.Hijacker`
- 这是一个接口实现问题

---

### 第三步: 深入分析

#### 3.1 理解 Hijacker 接口

**WebSocket 升级原理**:
1. 客户端发送 HTTP Upgrade 请求
2. 服务器需要"劫持"底层 TCP 连接
3. 将 HTTP 连接转换为 WebSocket 连接
4. 这需要 `http.ResponseWriter` 实现 `http.Hijacker` 接口

**Hijacker 接口定义**:
```go
type Hijacker interface {
    Hijack() (net.Conn, *bufio.ReadWriter, error)
}
```

#### 3.2 检查中间件

**查看请求处理链**:
```go
// server/cmd/main.go
r := router.NewRouter(&router.Routes{
    Middlewares: []func(http.Handler) http.Handler{
        recovery.Middleware,             // Panic 恢复
        serverlogger.RequestLogger(log), // 请求日志 ⚠️
    },
})
```

**查看日志中间件**:
```go
// server/internal/logger/middleware.go
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 创建响应记录器
            recorder := &responseRecorder{
                ResponseWriter: w,
                statusCode:     http.StatusOK,
            }
            
            // 调用下一个处理器
            next.ServeHTTP(recorder, r)  // ⚠️ 传递的是 recorder，不是原始 w
        })
    }
}

type responseRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
    r.statusCode = statusCode
    r.ResponseWriter.WriteHeader(statusCode)
}
// ⚠️ 没有实现 Hijack() 方法
```

**问题根源找到**:
- 日志中间件包装了 `http.ResponseWriter`
- `responseRecorder` 只实现了 `WriteHeader()` 方法
- 没有实现 `Hijack()` 方法
- WebSocket 升级时调用 `Hijack()` 失败

---

### 第四步: 解决方案设计

#### 4.1 解决思路

**方案**: 在 `responseRecorder` 中实现 `Hijack()` 方法

**实现原理**:
1. 检查底层 `ResponseWriter` 是否支持 `Hijacker`
2. 如果支持，转发调用
3. 如果不支持，返回错误

#### 4.2 代码实现

```go
// server/internal/logger/middleware.go

// Hijack 实现 http.Hijacker 接口（WebSocket 需要）
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
        return hijacker.Hijack()
    }
    return nil, nil, http.ErrNotSupported
}
```

**关键点**:
- 使用类型断言检查底层 ResponseWriter
- 转发 Hijack 调用到底层实现
- 保持中间件的透明性

---

### 第五步: 验证修复

#### 5.1 重新编译测试

```bash
go build -o /tmp/charline-server cmd/main.go
JWT_SECRET=test-secret-key-with-at-least-32-bytes-length /tmp/charline-server
```

#### 5.2 测试连接

```bash
go run test_ws.go
```

**结果**: ✅ 连接成功！

```
✓ WebSocket 连接成功
✓ 收到服务器消息: {"type":"challenge_response","payload":{"nonce":"..."},"timestamp":...}
```

**服务器日志**:
```
[DEBUG] HandleConnection called
[DEBUG] WebSocket upgrade successful
[DEBUG] Connection created
[DEBUG] Read/Write loops started
[DEBUG] sendChallenge called
[DEBUG] Nonce generated: ...
[DEBUG] Challenge message created
[DEBUG] Message marshaled: ...
[DEBUG] Data sent to channel
```

---

### 第六步: 发现第二个问题

#### 6.1 代码审查

**查看 HandleConnection 的执行顺序**:
```go
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
    wsConn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    conn := NewConnection(wsConn)
    
    // 发送初始挑战（nonce）
    if err := s.sendChallenge(conn); err != nil {  // ⚠️ 先发送
        conn.Close()
        return
    }
    
    // 启动读写循环
    go conn.ReadLoop(s.handler)  // ⚠️ 后启动
    go conn.WriteLoop()
}
```

#### 6.2 分析 Send 机制

**查看 Send 方法**:
```go
// server/internal/websocket/conn.go
func (c *Connection) Send(data []byte) error {
    select {
    case c.send <- data:  // ⚠️ 发送到 channel
        return nil
    default:
        return ErrSendBufferFull
    }
}
```

**查看 WriteLoop**:
```go
func (c *Connection) WriteLoop() {
    for {
        select {
        case data := <-c.send:  // ⚠️ 从 channel 读取
            // 发送消息
            c.conn.WriteMessage(websocket.TextMessage, data)
        }
    }
}
```

**问题分析**:
1. `sendChallenge` 调用 `conn.Send(data)`
2. `Send` 将数据放入 `c.send` channel
3. 但此时 `WriteLoop` 还没启动
4. 虽然 channel 有缓冲(256)，但数据无法真正发送
5. 这是一个时序问题

#### 6.3 修复方案

**调整执行顺序**:
```go
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
    wsConn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    conn := NewConnection(wsConn)
    
    // 先启动读写循环
    go conn.WriteLoop()  // ✅ 先启动
    go conn.ReadLoop(s.handler)
    
    // 发送初始挑战（nonce）
    if err := s.sendChallenge(conn); err != nil {  // ✅ 后发送
        conn.Close()
        return
    }
}
```

**原理**:
- 先启动 `WriteLoop`，确保有 goroutine 在监听 channel
- 再调用 `sendChallenge`，数据能够立即被处理
- 避免时序竞争问题

---

## 问题总结

### 问题 1: Hijacker 接口缺失

**现象**:
```
websocket: response does not implement http.Hijacker
```

**根本原因**:
- 日志中间件的 `responseRecorder` 包装了 `http.ResponseWriter`
- 未实现 `http.Hijacker` 接口
- WebSocket 升级需要 Hijacker 接口来劫持 TCP 连接

**解决方案**:
```go
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
        return hijacker.Hijack()
    }
    return nil, nil, http.ErrNotSupported
}
```

**影响范围**:
- 文件: `server/internal/logger/middleware.go`
- 影响: 所有 WebSocket 连接

### 问题 2: goroutine 启动顺序

**现象**:
- 虽然连接成功，但可能存在时序问题
- 数据发送到 channel 但没有接收者

**根本原因**:

- `sendChallenge` 在 `WriteLoop` 启动前调用
- 异步发送机制依赖 goroutine 已启动

**解决方案**:
```go
// 先启动 WriteLoop
go conn.WriteLoop()
go conn.ReadLoop(s.handler)

// 再发送消息
if err := s.sendChallenge(conn); err != nil {
    conn.Close()
    return
}
```

**影响范围**:
- 文件: `server/internal/websocket/server.go`
- 影响: 初始挑战消息发送

---

## 排查技巧总结

### 1. 日志驱动调试

**原则**: 在关键路径添加日志
```go
fmt.Println("[DEBUG] 步骤描述")
fmt.Printf("[ERROR] 错误信息: %v\n", err)
```

**位置**:
- 函数入口/出口
- 错误处理分支
- 状态转换点
- 异步操作启动点

### 2. 错误信息分析

**步骤**:
1. 记录完整错误信息
2. 理解错误含义
3. 定位错误来源
4. 追踪调用链

**示例**:
```
websocket: response does not implement http.Hijacker
         ↓
ResponseWriter 缺少 Hijacker 接口
         ↓
检查中间件是否包装了 ResponseWriter
         ↓
找到 responseRecorder 未实现接口
```

### 3. 代码审查

**检查点**:
- 接口实现完整性
- 异步操作时序
- 资源初始化顺序
- 错误处理覆盖

### 4. 分层验证

**策略**:
1. 验证服务器启动
2. 验证路由注册
3. 验证连接建立
4. 验证消息发送
5. 验证消息接收

### 5. 对比参考实现

**方法**:
- 查看 gorilla/websocket 官方示例
- 对比标准实现
- 理解最佳实践
- 识别偏差点

---

## 经验教训

### 1. 中间件设计原则

**透明性**: 中间件应该对底层接口透明
```go
// ❌ 错误: 只包装部分方法
type responseRecorder struct {
    http.ResponseWriter
}

// ✅ 正确: 转发所有接口
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
        return hijacker.Hijack()
    }
    return nil, nil, http.ErrNotSupported
}
```

### 2. 异步操作时序

**原则**: 先启动接收者，再发送数据
```go
// ❌ 错误顺序
sendData()
go receiveLoop()

// ✅ 正确顺序
go receiveLoop()
sendData()
```

### 3. 错误处理

**原则**: 不要静默吞掉错误
```go
// ❌ 错误: 静默返回
if err != nil {
    return
}

// ✅ 正确: 记录错误
if err != nil {
    log.Error("操作失败", zap.Error(err))
    return
}
```

### 4. 调试日志

**原则**: 临时调试日志要明确标记
```go
// ✅ 使用 [DEBUG] 前缀
fmt.Println("[DEBUG] HandleConnection called")

// 生产环境移除或使用日志级别控制
if log.Level() == DebugLevel {
    log.Debug("HandleConnection called")
}
```

---

## 工具和方法

### 1. 日志分析工具

```bash
# 实时查看日志
tail -f /tmp/server.log

# 过滤错误日志
grep ERROR /tmp/server.log

# 查看最近的日志
tail -20 /tmp/server.log
```

### 2. 网络调试工具

```bash
# 检查端口占用
lsof -ti:8080

# 测试 HTTP 连接
curl -v http://localhost:8080/health

# 测试 WebSocket (需要 websocat)
websocat ws://localhost:8080/ws
```

### 3. Go 调试技巧

```go
// 打印变量
fmt.Printf("变量: %+v\n", variable)

// 打印堆栈
debug.PrintStack()

// 类型断言检查
if hijacker, ok := w.(http.Hijacker); ok {
    fmt.Println("支持 Hijacker")
}
```

---

## 参考资料

### WebSocket 相关
- [RFC 6455 - WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [gorilla/websocket 文档](https://pkg.go.dev/github.com/gorilla/websocket)
- [Go http.Hijacker 接口](https://pkg.go.dev/net/http#Hijacker)

### 中间件设计
- [Go HTTP 中间件模式](https://www.alexedwards.net/blog/making-and-using-middleware)
- [ResponseWriter 包装最佳实践](https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body)

### 调试技巧
- [Go 调试指南](https://golang.org/doc/diagnostics.html)
- [日志最佳实践](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)

---

## 附录: 完整测试代码

### 测试程序
```go
// test_ws.go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/gorilla/websocket"
)

func main() {
    // 连接到 WebSocket 服务器
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
    if err != nil {
        log.Fatal("连接失败:", err)
    }
    defer conn.Close()

    fmt.Println("✓ WebSocket 连接成功")

    // 设置读取超时
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))

    // 读取服务器发送的挑战消息
    _, message, err := conn.ReadMessage()
    if err != nil {
        log.Fatal("读取消息失败:", err)
    }

    fmt.Printf("✓ 收到服务器消息: %s\n", string(message))
}
```

### 编译运行
```bash
# 编译测试程序
go mod init test
go get github.com/gorilla/websocket@v1.5.1
go run test_ws.go
```
