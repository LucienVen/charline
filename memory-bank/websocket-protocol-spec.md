# WebSocket 消息协议规范

> **文档版本**: v1.0  
> **创建时间**: 2026-01-29  
> **适用阶段**: Phase 3.1 - WebSocket 基础连接  
> **状态**: Draft

---

## 一、协议概述

### 1.1 设计原则

- **简洁性**: 使用 JSON 格式，易于调试和扩展
- **类型安全**: 明确的消息类型定义
- **可扩展性**: 支持未来新增消息类型
- **向后兼容**: 保留字段不删除，新增字段可选

### 1.2 消息结构

所有 WebSocket 消息使用统一的 JSON 结构：

```json
{
  "type": "message_type",
  "payload": {
    "field1": "value1",
    "field2": "value2"
  },
  "timestamp": "2026-01-29T10:30:00Z",
  "request_id": "req-uuid-1234"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | ✅ | 消息类型标识符 |
| `payload` | object | ✅ | 消息负载数据 |
| `timestamp` | string | ❌ | ISO 8601 格式时间戳 |
| `request_id` | string | ❌ | 请求追踪 ID（用于关联请求-响应） |

---

## 二、认证流程消息

### 2.1 Challenge Request（挑战请求）

**方向**: Client → Server  
**时机**: WebSocket 连接建立后，客户端首次请求认证

```json
{
  "type": "challenge_request",
  "payload": {},
  "timestamp": "2026-01-29T10:30:00Z",
  "request_id": "req-001"
}
```

**Payload 字段**: 无

---

### 2.2 Challenge Response（挑战响应）

**方向**: Server → Client  
**时机**: 服务端收到 challenge_request 后返回

```json
{
  "type": "challenge_response",
  "payload": {
    "nonce": "base64_encoded_32_bytes_random_value",
    "expires_at": "2026-01-29T10:30:30Z"
  },
  "timestamp": "2026-01-29T10:30:00Z",
  "request_id": "req-001"
}
```

**Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `nonce` | string | ✅ | Base64 编码的 32 字节随机值 |
| `expires_at` | string | ✅ | Nonce 过期时间（ISO 8601 格式） |

**安全要求**:
- Nonce 使用 `crypto/rand` 生成
- Nonce 有效期 30 秒
- Nonce 只能使用一次（原子消费）

---

### 2.3 Auth Request（认证请求）

**方向**: Client → Server  
**时机**: 客户端收到 challenge_response 后，使用 Ed25519 私钥签名 nonce

```json
{
  "type": "auth_request",
  "payload": {
    "nonce": "base64_encoded_nonce",
    "signature": "base64_encoded_ed25519_signature",
    "public_key": "base64_encoded_ed25519_public_key"
  },
  "timestamp": "2026-01-29T10:30:05Z",
  "request_id": "req-002"
}
```

**Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `nonce` | string | ✅ | 服务端下发的 nonce（原样返回） |
| `signature` | string | ✅ | Ed25519 签名（Base64 编码） |
| `public_key` | string | ✅ | Ed25519 公钥（Base64 编码，32 字节） |

**签名算法**:
```go
// 客户端签名逻辑
signature := ed25519.Sign(privateKey, []byte(nonce))
signatureBase64 := base64.StdEncoding.EncodeToString(signature)
```

---

### 2.4 Auth Response（认证响应）

**方向**: Server → Client  
**时机**: 服务端验证签名后返回

#### 成功响应

```json
{
  "type": "auth_response",
  "payload": {
    "success": true,
    "session_id": "sess-uuid-5678",
    "user_id": 123,
    "username": "alice",
    "session_token": "jwt_session_token_here",
    "expires_at": "2026-01-29T18:30:00Z"
  },
  "timestamp": "2026-01-29T10:30:06Z",
  "request_id": "req-002"
}
```

**成功 Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `success` | boolean | ✅ | 认证是否成功 |
| `session_id` | string | ✅ | 会话 ID（UUID 格式） |
| `user_id` | integer | ✅ | 用户 ID |
| `username` | string | ✅ | 用户名 |
| `session_token` | string | ✅ | Session Token（JWT 格式） |
| `expires_at` | string | ✅ | Session 过期时间 |

#### 失败响应

```json
{
  "type": "auth_response",
  "payload": {
    "success": false,
    "error_code": 3005,
    "error_message": "签名验证失败",
    "details": {
      "reason": "invalid_signature"
    }
  },
  "timestamp": "2026-01-29T10:30:06Z",
  "request_id": "req-002"
}
```

**失败 Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `success` | boolean | ✅ | 固定为 false |
| `error_code` | integer | ✅ | 错误码（参考 errors/codes.go） |
| `error_message` | string | ✅ | 用户友好的错误消息 |
| `details` | object | ❌ | 详细错误信息 |

**常见错误码**:

| 错误码 | 说明 |
|--------|------|
| 3004 | Nonce 无效（不存在、已使用或已过期） |
| 3005 | 签名验证失败 |
| 3006 | 公钥格式无效 |

---

## 三、心跳保活消息

### 3.1 Ping（心跳请求）

**方向**: Client → Server 或 Server → Client  
**时机**: 定期发送（建议间隔 30 秒）

```json
{
  "type": "ping",
  "payload": {},
  "timestamp": "2026-01-29T10:31:00Z"
}
```

**Payload 字段**: 无

---

### 3.2 Pong（心跳响应）

**方向**: Server → Client 或 Client → Server  
**时机**: 收到 ping 后立即响应

```json
{
  "type": "pong",
  "payload": {},
  "timestamp": "2026-01-29T10:31:00Z"
}
```

**Payload 字段**: 无

**心跳机制说明**:
- 客户端和服务端都可以主动发送 ping
- 收到 ping 必须立即回复 pong
- 如果 3 次心跳无响应，视为连接断开
- 心跳间隔建议 30 秒

---

## 四、错误消息

### 4.1 Error（通用错误）

**方向**: Server → Client  
**时机**: 服务端处理消息时发生错误

```json
{
  "type": "error",
  "payload": {
    "error_code": 5000,
    "error_message": "系统内部错误",
    "details": {
      "reason": "database_connection_failed",
      "retry_after": 5
    }
  },
  "timestamp": "2026-01-29T10:32:00Z",
  "request_id": "req-003"
}
```

**Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `error_code` | integer | ✅ | 错误码 |
| `error_message` | string | ✅ | 用户友好的错误消息 |
| `details` | object | ❌ | 详细错误信息 |

**错误码分段**:

| 范围 | 说明 |
|------|------|
| 1xxx | 参数/请求错误 |
| 2xxx | 资源相关错误 |
| 3xxx | 认证/授权错误 |
| 5xxx | 系统内部错误 |

---

## 五、连接管理消息

### 5.1 Close（关闭连接）

**方向**: Client → Server 或 Server → Client  
**时机**: 主动关闭连接前发送

```json
{
  "type": "close",
  "payload": {
    "reason": "user_logout",
    "message": "用户主动登出"
  },
  "timestamp": "2026-01-29T10:35:00Z"
}
```

**Payload 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `reason` | string | ✅ | 关闭原因代码 |
| `message` | string | ❌ | 关闭原因描述 |

**常见关闭原因**:

| reason | 说明 |
|--------|------|
| `user_logout` | 用户主动登出 |
| `session_expired` | Session 过期 |
| `server_shutdown` | 服务端关闭 |
| `duplicate_connection` | 重复连接（同一用户在其他地方登录） |

---

## 六、消息验证规则

### 6.1 必填字段验证

所有消息必须包含：
- `type` 字段（非空字符串）
- `payload` 字段（对象类型，可以为空对象 `{}`）

### 6.2 类型验证

- `type` 必须是已定义的消息类型
- 未知类型返回错误：
  ```json
  {
    "type": "error",
    "payload": {
      "error_code": 1001,
      "error_message": "未知消息类型",
      "details": {
        "received_type": "unknown_type"
      }
    }
  }
  ```

### 6.3 Payload 验证

- 每种消息类型的 payload 字段必须符合规范
- 缺少必填字段返回错误：
  ```json
  {
    "type": "error",
    "payload": {
      "error_code": 1001,
      "error_message": "缺少必填字段",
      "details": {
        "missing_fields": ["nonce", "signature"]
      }
    }
  }
  ```

### 6.4 长度限制

| 字段 | 最大长度 |
|------|----------|
| `type` | 64 字符 |
| `nonce` | 64 字符（Base64 编码后） |
| `signature` | 128 字符（Base64 编码后） |
| `public_key` | 64 字符（Base64 编码后） |
| `error_message` | 256 字符 |

---

## 七、实现示例

### 7.1 Go 消息结构定义

```go
// server/internal/protocol/message.go

package protocol

import (
    "encoding/json"
    "time"
)

// Message 统一消息结构
type Message struct {
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp string                 `json:"timestamp,omitempty"`
    RequestID string                 `json:"request_id,omitempty"`
}

// NewMessage 创建新消息
func NewMessage(msgType string, payload map[string]interface{}) *Message {
    return &Message{
        Type:      msgType,
        Payload:   payload,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}

// Marshal 序列化消息
func (m *Message) Marshal() ([]byte, error) {
    return json.Marshal(m)
}

// Unmarshal 反序列化消息
func Unmarshal(data []byte) (*Message, error) {
    var msg Message
    if err := json.Unmarshal(data, &msg); err != nil {
        return nil, err
    }
    return &msg, nil
}

// Validate 验证消息格式
func (m *Message) Validate() error {
    if m.Type == "" {
        return errors.New("message type is required")
    }
    if m.Payload == nil {
        return errors.New("message payload is required")
    }
    return nil
}
```

### 7.2 认证流程示例

```go
// 客户端认证流程
func (c *Client) Authenticate() error {
    // 1. 请求 challenge
    challengeReq := protocol.NewMessage("challenge_request", map[string]interface{}{})
    if err := c.Send(challengeReq); err != nil {
        return fmt.Errorf("send challenge request failed: %w", err)
    }

    // 2. 接收 nonce
    challengeResp, err := c.Receive()
    if err != nil {
        return fmt.Errorf("receive challenge response failed: %w", err)
    }

    nonce := challengeResp.Payload["nonce"].(string)

    // 3. 使用 Ed25519 私钥签名
    signature, err := c.signer.Sign(nonce)
    if err != nil {
        return fmt.Errorf("sign nonce failed: %w", err)
    }

    // 4. 发送认证请求
    authReq := protocol.NewMessage("auth_request", map[string]interface{}{
        "nonce":      nonce,
        "signature":  signature,
        "public_key": c.signer.PublicKeyBase64(),
    })
    if err := c.Send(authReq); err != nil {
        return fmt.Errorf("send auth request failed: %w", err)
    }

    // 5. 接收认证结果
    authResp, err := c.Receive()
    if err != nil {
        return fmt.Errorf("receive auth response failed: %w", err)
    }

    if !authResp.Payload["success"].(bool) {
        return fmt.Errorf("authentication failed: %s", authResp.Payload["error_message"])
    }

    c.sessionID = authResp.Payload["session_id"].(string)
    c.logger.Info("Authentication successful", zap.String("session_id", c.sessionID))

    return nil
}
```

---

## 八、版本演进

### 8.1 当前版本（v1.0）

- 支持基础认证流程（Challenge-Response）
- 支持心跳保活（Ping-Pong）
- 支持错误处理
- 支持连接管理

### 8.2 未来扩展（v2.0+）

计划支持的消息类型：

| 消息类型 | 说明 | 计划阶段 |
|----------|------|----------|
| `resume_request` | 断线恢复请求 | Phase 3.4 |
| `resume_response` | 断线恢复响应 | Phase 3.4 |
| `message` | 聊天消息 | Phase 4 |
| `message_ack` | 消息确认 | Phase 4 |
| `typing` | 输入状态 | Phase 5 |
| `presence` | 在线状态 | Phase 5 |

---

## 九、参考文档

- [Phase 3.1 实施计划](../nimbalyst-local/plans/phase3-1-websocket-basic-connection.md)
- [IM 系统认证策略挑战回复](./thinking-challenge/response-to-auth-challenge.md)
- [架构设计文档](./architecture.md)
- [错误码定义](../server/internal/errors/codes.go)

---

**文档维护者**: lxt  
**最后更新**: 2026-01-29
