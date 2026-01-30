package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/LucienVen/charline/client/internal/auth"

	"github.com/gorilla/websocket"
)

// Client WebSocket 客户端
type Client struct {
	conn      *websocket.Conn // WebSocket 连接
	signer    *auth.Signer    // Ed25519 签名器
	serverURL string          // 服务器 URL
	userID    int64           // 用户 ID（认证后设置）
	send      chan []byte     // 发送消息通道
	closeChan chan struct{}   // 关闭信号通道
	closeOnce sync.Once       // 确保只关闭一次
	isClosed  bool            // 连接是否已关闭
	mu        sync.RWMutex    // 读写锁
}

// NewClient 创建新客户端
func NewClient(serverURL string, signer *auth.Signer) *Client {
	return &Client{
		serverURL: serverURL,
		signer:    signer,
		send:      make(chan []byte, 256),
		closeChan: make(chan struct{}),
		isClosed:  false,
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	// 建立 WebSocket 连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn

	// 启动读写循环
	go c.readLoop()
	go c.writeLoop()

	return nil
}

// Authenticate 处理认证流程
func (c *Client) Authenticate() error {
	// 等待服务器发送 challenge
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}

	// 解析 challenge 消息
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to parse challenge: %w", err)
	}

	if msg.Type != "challenge_response" {
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	// 解析 nonce
	var challenge ChallengePayload
	if err := json.Unmarshal(msg.Payload, &challenge); err != nil {
		return fmt.Errorf("failed to parse challenge payload: %w", err)
	}

	// 使用 Ed25519 签名 nonce
	signature, err := c.signer.Sign(challenge.Nonce)
	if err != nil {
		return fmt.Errorf("failed to sign nonce: %w", err)
	}

	// 构造认证请求
	authReq := AuthRequestPayload{
		PublicKey: c.signer.PublicKeyHex(),
		Signature: signature,
		Nonce:     challenge.Nonce,
	}

	authMsg, err := NewMessage("auth_request", authReq)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	authData, err := authMsg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	// 发送认证请求
	if err := c.Send(authData); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// 等待认证响应
	_, respData, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	var respMsg Message
	if err := json.Unmarshal(respData, &respMsg); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if respMsg.Type == "error" {
		var errPayload ErrorPayload
		if err := json.Unmarshal(respMsg.Payload, &errPayload); err != nil {
			return fmt.Errorf("authentication failed: unknown error")
		}
		return fmt.Errorf("authentication failed: %s - %s", errPayload.Code, errPayload.Message)
	}

	if respMsg.Type != "auth_response" {
		return fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}

	var authResp AuthResponsePayload
	if err := json.Unmarshal(respMsg.Payload, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !authResp.Success {
		return fmt.Errorf("authentication failed: %s", authResp.Message)
	}

	c.userID = authResp.UserID
	return nil
}

// Send 发送消息（异步）
func (c *Client) Send(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return fmt.Errorf("connection is closed")
	}

	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer is full")
	}
}

// readLoop 读取循环（goroutine）
func (c *Client) readLoop() {
	defer func() {
		c.Close()
	}()

	// 设置读取超时
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
			// 读取消息
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					// 异常关闭
					fmt.Printf("WebSocket error: %v\n", err)
				}
				return
			}

			// 处理接收到的消息
			c.handleMessage(data)
		}
	}
}

// writeLoop 写入循环（goroutine）
func (c *Client) writeLoop() {
	ticker := time.NewTicker(54 * time.Second) // 心跳间隔
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
			// 发送 Ping
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("Failed to parse message: %v\n", err)
		return
	}

	switch msg.Type {
	case "pong":
		// 心跳响应
		fmt.Println("Received pong")
	case "error":
		var errPayload ErrorPayload
		if err := json.Unmarshal(msg.Payload, &errPayload); err != nil {
			fmt.Printf("Failed to parse error: %v\n", err)
			return
		}
		fmt.Printf("Server error: %s - %s\n", errPayload.Code, errPayload.Message)
	default:
		// 其他消息类型
		fmt.Printf("Received message: %s\n", msg.Type)
	}
}

// Close 关闭连接
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.isClosed = true
		c.mu.Unlock()

		close(c.closeChan)
		close(c.send)
		if c.conn != nil {
			c.conn.Close()
		}
	})
	return nil
}

// IsClosed 连接是否已关闭
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

// GetUserID 获取用户 ID
func (c *Client) GetUserID() int64 {
	return c.userID
}
