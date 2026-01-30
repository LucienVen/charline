# WebSocket 连接验证指南

本文档说明如何验证 Charline 项目的 WebSocket 连接功能。

---

## 一、验证环境准备

### 1. 编译项目

```bash
# 编译服务器
cd /path/to/charline/server
go build -o ../bin/server cmd/main.go

# 编译客户端
cd /path/to/charline/client
go build -o ../bin/charline cmd/main.go
```

### 2. 启动服务器

```bash
cd /path/to/charline
./bin/server

# 或者后台启动
nohup ./bin/server > /tmp/server.log 2>&1 &
```

**验证服务器启动成功：**
```bash
# 检查端口监听
lsof -i:8080

# 检查健康端点
curl http://localhost:8080/health
# 应返回: OK

# 查看注册的路由
tail -20 /tmp/server.log | grep "/ws"
# 应看到: GET    /ws
```

---

## 二、验证方法

### 方法 1：使用客户端命令（推荐）

#### Step 1: 生成邀请码

```bash
curl -X POST http://localhost:8080/api/v1/invite/generate \
  -H "Content-Type: application/json" \
  -d '{"count": 1}'

# 响应示例:
# {"code":0,"message":"成功","data":{"code":"INV-XXXXXXXX"}}
```

#### Step 2: 加入服务器

```bash
./bin/charline join INV-XXXXXXXX testuser
```

**预期输出：**
```
✓ Join 成功！
  用户名: testuser
  凭证版本: 1
  凭证已保存到 ~/.charline/
```

#### Step 3: 建立 WebSocket 连接

```bash
./bin/charline connect
```

**预期输出：**
```
正在连接到服务器...
✓ 连接成功
正在进行身份认证...
✓ 认证成功 (用户 ID: 1)

已连接到服务器，按 Ctrl+C 断开连接
```

**验证成功标志：**
- ✅ 连接成功
- ✅ 收到 challenge_response
- ✅ 认证成功
- ✅ 连接保持打开

---

### 方法 2：使用 Go 测试程序

创建测试文件 `test_ws.go`：

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws"
	fmt.Printf("连接到 %s...\n", url)

	// 建立 WebSocket 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("✗ 连接失败: %v", err)
	}
	defer conn.Close()

	fmt.Println("✓ WebSocket 连接成功！")

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 读取服务器发送的第一条消息（challenge）
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("✗ 读取消息失败: %v", err)
	}

	fmt.Printf("收到消息: %s\n", string(message))

	// 解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Fatalf("✗ 解析消息失败: %v", err)
	}

	if msg["type"] == "challenge_response" {
		fmt.Println("✓ 收到 challenge_response，WebSocket 握手成功！")
	} else {
		fmt.Printf("✗ 意外的消息类型: %v\n", msg["type"])
	}
}
```

运行测试：

```bash
go mod init test_ws
go get github.com/gorilla/websocket@v1.5.1
go run test_ws.go
```

**预期输出：**
```
连接到 ws://localhost:8080/ws...
✓ WebSocket 连接成功！
收到消息: {"type":"challenge_response","payload":{"nonce":"..."}}
✓ 收到 challenge_response，WebSocket 握手成功！
```

---

### 方法 3：使用 websocat 工具

安装 websocat：

```bash
# macOS
brew install websocat

# Linux
cargo install websocat
```

测试连接：

```bash
websocat ws://localhost:8080/ws
```

**预期行为：**
- 连接成功后，立即收到一条 JSON 消息（challenge_response）
- 消息格式：`{"type":"challenge_response","payload":{"nonce":"..."}}`

---

## 三、验证检查清单

### 服务器端检查

- [ ] 服务器成功启动，监听 8080 端口
- [ ] `/ws` 路由已注册
- [ ] 健康检查端点 `/health` 正常响应
- [ ] 日志显示 "WebSocket 服务器初始化成功"

### 客户端连接检查

- [ ] WebSocket 连接成功建立（HTTP 101 Switching Protocols）
- [ ] 收到服务器发送的 `challenge_response` 消息
- [ ] 消息格式正确（包含 type 和 payload 字段）
- [ ] payload 中包含 nonce 字段

### 认证流程检查

- [ ] 客户端发送 `auth_request`（包含公钥、签名、nonce）
- [ ] 服务器验证签名成功
- [ ] 服务器返回 `auth_response`（success: true）
- [ ] 连接关联到 userID
- [ ] 连接添加到连接池

---

## 四、常见问题排查

### 问题 1：连接失败 "websocket: bad handshake"

**可能原因：**
- 服务器未启动或端口被占用
- WebSocket 路由未正确注册
- 服务器在处理请求时发生 panic

**排查步骤：**
```bash
# 1. 检查服务器是否运行
lsof -i:8080

# 2. 检查服务器日志
tail -50 /tmp/server.log

# 3. 测试 HTTP 端点
curl http://localhost:8080/health

# 4. 查看路由注册
grep "/ws" /tmp/server.log
```

### 问题 2：连接成功但未收到消息

**可能原因：**
- `sendChallenge` 函数执行失败
- NonceStore.Generate() 出错
- 消息序列化失败

**排查步骤：**
```bash
# 查看服务器日志中的错误
grep "ERROR" /tmp/server.log

# 启用 debug 日志
LOG_LEVEL=debug ./bin/server
```

### 问题 3：认证失败

**可能原因：**
- 客户端密钥对不存在或损坏
- 签名验证失败
- Nonce 已过期或被消费

**排查步骤：**
```bash
# 检查客户端凭证
ls -la ~/.charline/

# 重新生成凭证
rm -rf ~/.charline/
./bin/charline join <invite-code> <username>

# 查看服务器认证日志
grep "auth" /tmp/server.log
```

---

## 五、成功验证的标志

当看到以下输出时，说明 WebSocket 连接验证成功：

### 客户端输出

```
正在连接到服务器...
✓ 连接成功
正在进行身份认证...
✓ 认证成功 (用户 ID: 1)

已连接到服务器，按 Ctrl+C 断开连接
```

### 服务器日志

```
+0800 2026-01-30 15:00:00  INFO  Server started  {"address": ":8080"}
+0800 2026-01-30 15:00:05  INFO  WebSocket connection established  {"remote_addr": "127.0.0.1:xxxxx"}
+0800 2026-01-30 15:00:05  INFO  Challenge sent  {"nonce": "..."}
+0800 2026-01-30 15:00:05  INFO  Authentication successful  {"user_id": 1}
+0800 2026-01-30 15:00:05  INFO  Connection added to pool  {"user_id": 1, "pool_size": 1}
```

---

## 六、下一步

WebSocket 连接验证成功后，可以继续进行：

1. **消息收发测试**：测试客户端和服务器之间的消息传输
2. **心跳机制测试**：验证 Ping/Pong 心跳是否正常工作
3. **断线重连测试**：测试网络中断后的重连机制
4. **并发连接测试**：测试多个客户端同时连接
5. **性能测试**：测试连接数和消息吞吐量

---

**最后更新：** 2026-01-30  
**维护者：** Charline 开发团队
