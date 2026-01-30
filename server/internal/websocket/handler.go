package websocket

import (
	"crypto/ed25519"
	"encoding/hex"
	"time"

	"github.com/LucienVen/charline/server/internal/store"
)

// WSMessageHandler WebSocket 消息处理器
type WSMessageHandler struct {
	nonceStore *store.NonceStore
	userStore  *store.UserStore
	pool       *ConnectionPool
}

// NewWSMessageHandler 创建消息处理器
func NewWSMessageHandler(nonceStore *store.NonceStore, userStore *store.UserStore, pool *ConnectionPool) *WSMessageHandler {
	return &WSMessageHandler{
		nonceStore: nonceStore,
		userStore:  userStore,
		pool:       pool,
	}
}

// HandleMessage 处理 WebSocket 消息
func (h *WSMessageHandler) HandleMessage(conn *Connection, msg *Message) {
	switch msg.Type {
	case MessageTypeAuthRequest:
		h.handleAuthRequest(conn, msg)
	case MessageTypePing:
		h.handlePing(conn, msg)
	default:
		// 未认证连接只能发送认证请求
		if !conn.IsAuthenticated() {
			h.sendError(conn, "UNAUTHORIZED", "Authentication required")
			return
		}

		// 已认证连接的其他消息处理
		h.handleAuthenticatedMessage(conn, msg)
	}
}

// handleAuthRequest 处理认证请求
func (h *WSMessageHandler) handleAuthRequest(conn *Connection, msg *Message) {
	// 解析认证请求
	var authReq AuthRequestPayload
	if err := msg.UnmarshalPayload(&authReq); err != nil {
		h.sendError(conn, "INVALID_PAYLOAD", "Failed to parse auth request")
		return
	}

	// 验证必填字段
	if authReq.PublicKey == "" || authReq.Signature == "" || authReq.Nonce == "" {
		h.sendError(conn, "INVALID_REQUEST", "Missing required fields")
		return
	}

	// 解码公钥
	publicKeyBytes, err := hex.DecodeString(authReq.PublicKey)
	if err != nil || len(publicKeyBytes) != ed25519.PublicKeySize {
		h.sendError(conn, "INVALID_PUBLIC_KEY", "Invalid public key format")
		return
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)

	// 解码签名
	signatureBytes, err := hex.DecodeString(authReq.Signature)
	if err != nil || len(signatureBytes) != ed25519.SignatureSize {
		h.sendError(conn, "INVALID_SIGNATURE", "Invalid signature format")
		return
	}

	// 验证并消费 nonce（防止重放攻击）
	_, valid := h.nonceStore.Consume(authReq.Nonce)
	if !valid {
		h.sendError(conn, "INVALID_NONCE", "Nonce is invalid or expired")
		return
	}

	// 验证签名
	if !ed25519.Verify(publicKey, []byte(authReq.Nonce), signatureBytes) {
		h.sendError(conn, "SIGNATURE_VERIFICATION_FAILED", "Signature verification failed")
		return
	}

	// 根据公钥查找用户
	user, err := h.userStore.GetByPublicKey(authReq.PublicKey)
	if err != nil {
		h.sendError(conn, "USER_NOT_FOUND", "User not found")
		return
	}

	// 设置连接的用户 ID
	conn.SetUserID(user.ID)

	// 将连接添加到连接池
	h.pool.Add(conn)

	// 发送认证成功响应
	authResp := AuthResponsePayload{
		Success: true,
		UserID:  user.ID,
	}

	respMsg, err := NewMessage(MessageTypeAuthResponse, authResp)
	if err != nil {
		h.sendError(conn, "INTERNAL_ERROR", "Failed to create response")
		return
	}

	respData, err := respMsg.Marshal()
	if err != nil {
		h.sendError(conn, "INTERNAL_ERROR", "Failed to marshal response")
		return
	}

	conn.Send(respData)
}

// handlePing 处理 Ping 消息
func (h *WSMessageHandler) handlePing(conn *Connection, msg *Message) {
	// 发送 Pong 响应
	pongMsg, err := NewMessage(MessageTypePong, map[string]interface{}{
		"timestamp": time.Now().UnixMilli(),
	})
	if err != nil {
		return
	}

	pongData, err := pongMsg.Marshal()
	if err != nil {
		return
	}

	conn.Send(pongData)
}

// handleAuthenticatedMessage 处理已认证连接的消息
func (h *WSMessageHandler) handleAuthenticatedMessage(conn *Connection, msg *Message) {
	// TODO: 实现业务消息处理逻辑
	// 例如：聊天消息、状态更新等
}

// sendError 发送错误消息
func (h *WSMessageHandler) sendError(conn *Connection, code string, message string) {
	errPayload := ErrorPayload{
		Code:    code,
		Message: message,
	}

	errMsg, err := NewMessage(MessageTypeError, errPayload)
	if err != nil {
		return
	}

	errData, err := errMsg.Marshal()
	if err != nil {
		return
	}

	conn.Send(errData)
}
