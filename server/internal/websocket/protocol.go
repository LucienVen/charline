package websocket

import (
	"encoding/json"
	"time"
)

// 消息类型常量
const (
	// 认证相关
	MessageTypeChallengeResponse = "challenge_response" // 服务端发送 nonce 挑战
	MessageTypeAuthRequest       = "auth_request"       // 客户端发送签名认证
	MessageTypeAuthResponse      = "auth_response"      // 服务端返回认证结果

	// 断线恢复
	MessageTypeResumeRequest  = "resume_request"  // 客户端发送恢复请求
	MessageTypeResumeResponse = "resume_response" // 服务端返回恢复结果

	// 心跳
	MessageTypePing = "ping"
	MessageTypePong = "pong"

	// 错误
	MessageTypeError = "error"
)

// Message WebSocket 消息结构
type Message struct {
	Type      string          `json:"type"`                // 消息类型
	Payload   json.RawMessage `json:"payload"`             // 消息载荷（JSON）
	Timestamp int64           `json:"timestamp"`           // 时间戳（毫秒）
	RequestID string          `json:"request_id,omitempty"` // 请求ID（可选）
}

// ChallengePayload 挑战载荷
type ChallengePayload struct {
	Nonce string `json:"nonce"` // 随机数
}

// AuthRequestPayload 认证请求载荷
type AuthRequestPayload struct {
	PublicKey string `json:"public_key"`        // Ed25519 公钥（hex）
	Signature string `json:"signature"`         // 签名（hex）
	Nonce     string `json:"nonce"`             // 原始 nonce
	DeviceID  string `json:"device_id,omitempty"` // 设备标识（可选，用于多设备管理）
}

// AuthResponsePayload 认证响应载荷
type AuthResponsePayload struct {
	Success   bool   `json:"success"`            // 是否成功
	UserID    int64  `json:"user_id,omitempty"`  // 用户ID（成功时）
	SessionID string `json:"session_id,omitempty"` // Session ID（成功时）
	ResumeToken  string `json:"resume_token,omitempty"`  // Resume Token（成功时）
	ResumeExpiry int64  `json:"resume_expiry,omitempty"` // Resume Token 过期时间（Unix毫秒）
	Message   string `json:"message,omitempty"`  // 错误信息（失败时）
}

// ErrorPayload 错误载荷

// ResumeRequestPayload 恢复请求载荷
type ResumeRequestPayload struct {
	ResumeToken string `json:"resume_token"`        // Resume Token
	DeviceID    string `json:"device_id,omitempty"` // 设备标识（可选）
}

// ResumeResponsePayload 恢复响应载荷
type ResumeResponsePayload struct {
	Success      bool   `json:"success"`                 // 是否成功
	SessionID    string `json:"session_id,omitempty"`    // Session ID（成功时）
	UserID       int64  `json:"user_id,omitempty"`       // 用户ID（成功时）
	ResumeToken  string `json:"resume_token,omitempty"`  // 新的 Resume Token（成功时）
	ResumeExpiry int64  `json:"resume_expiry,omitempty"` // 新 Token 过期时间（Unix毫秒）
	Message      string `json:"message,omitempty"`       // 错误信息（失败时）
}
type ErrorPayload struct {
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
}

// NewMessage 创建新消息
func NewMessage(msgType string, payload interface{}) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now().UnixMilli(),
	}, nil
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

// UnmarshalPayload 反序列化载荷到指定类型
func (m *Message) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(m.Payload, v)
}
