package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/LucienVen/charline/server/internal/session"
)

// Connection WebSocket 连接封装
type Connection struct {
	id             string                 // 连接唯一标识
	conn           *websocket.Conn        // WebSocket 连接
	userID         int64                  // 用户ID（认证后设置）
	sessionID      string                 // Session ID（认证后设置）
	sessionManager session.SessionManager // Session 管理器
	send           chan []byte            // 发送消息通道
	closeChan      chan struct{}          // 关闭信号通道
	closeOnce      sync.Once              // 确保只关闭一次
	isClosed       bool                   // 连接是否已关闭
	mu             sync.RWMutex           // 读写锁
}

// NewConnection 创建新连接
func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		id:        uuid.New().String(),
		conn:      conn,
		userID:    0, // 未认证
		send:      make(chan []byte, 256),
		closeChan: make(chan struct{}),
		isClosed:  false,
	}
}

// ID 获取连接ID
func (c *Connection) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

// SetUserID 设置用户ID（认证成功后调用）
func (c *Connection) SetUserID(userID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userID = userID
}

// GetUserID 获取用户ID
func (c *Connection) GetUserID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// IsAuthenticated 是否已认证
func (c *Connection) IsAuthenticated() bool {
	return c.GetUserID() > 0
}

// Send 发送消息（异步）
func (c *Connection) Send(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return ErrConnectionClosed
	}

	select {
	case c.send <- data:
		return nil
	default:
		// 发送通道已满，连接可能阻塞
		return ErrSendBufferFull
	}
}

// ReadLoop 读取循环（goroutine）
func (c *Connection) ReadLoop(handler MessageHandler) {
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
				}
				return
			}

			// 解析消息
			msg, err := Unmarshal(data)
			if err != nil {
				// 发送错误响应
				errMsg, _ := NewMessage(MessageTypeError, ErrorPayload{
					Code:    "INVALID_MESSAGE",
					Message: "Failed to parse message",
				})
				errData, _ := errMsg.Marshal()
				c.Send(errData)
				continue
			}

			// 处理消息
			handler.HandleMessage(c, msg)
		}
	}
}

// WriteLoop 写入循环（goroutine）
func (c *Connection) WriteLoop() {
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

// Close 关闭连接
func (c *Connection) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		sessionID := c.sessionID
		sessionManager := c.sessionManager
		c.isClosed = true
		c.mu.Unlock()
		// 如果有关联的 Session，挂起它
		if sessionID != "" && sessionManager != nil {
			sessionManager.Suspend(sessionID)
		}
		close(c.closeChan)
		close(c.send)
		c.conn.Close()
	})
	return nil
}

func (c *Connection) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

// 错误定义
var (
	ErrConnectionClosed = &ConnectionError{Code: "CONNECTION_CLOSED", Message: "Connection is closed"}
	ErrSendBufferFull   = &ConnectionError{Code: "SEND_BUFFER_FULL", Message: "Send buffer is full"}
)

// ConnectionError 连接错误
type ConnectionError struct {
	Code    string
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleMessage(conn *Connection, msg *Message)
}
