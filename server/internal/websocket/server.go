package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Server WebSocket 服务器
type Server struct {
	upgrader *websocket.Upgrader
	handler  *WSMessageHandler
}

// NewServer 创建 WebSocket 服务器
func NewServer(handler *WSMessageHandler) *Server {
	return &Server{
		upgrader: &websocket.Upgrader{
			// 读缓冲区大小
			ReadBufferSize: 1024,
			// 写缓冲区大小
			WriteBufferSize: 1024,
			// 允许跨域（生产环境需要配置具体域名）
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			// 握手超时
			HandshakeTimeout: 10 * time.Second,
		},
		handler: handler,
	}
}

// HandleConnection 处理 WebSocket 连接（HTTP Handler）
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[DEBUG] HandleConnection called")
	
	// 升级 HTTP 连接为 WebSocket
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 升级失败，upgrader 已经发送了错误响应
		fmt.Printf("[ERROR] WebSocket upgrade failed: %v\n", err)
		return
	}
	fmt.Println("[DEBUG] WebSocket upgrade successful")

	// 创建连接封装
	conn := NewConnection(wsConn)
	fmt.Println("[DEBUG] Connection created")

	// 先启动读写循环
	go conn.WriteLoop()
	go conn.ReadLoop(s.handler)
	fmt.Println("[DEBUG] Read/Write loops started")

	// 发送初始挑战（nonce）
	if err := s.sendChallenge(conn); err != nil {
		fmt.Printf("[ERROR] Send challenge failed: %v\n", err)
		conn.Close()
		return
	}
	fmt.Println("[DEBUG] Challenge sent successfully")
}

// sendChallenge 发送挑战消息
func (s *Server) sendChallenge(conn *Connection) error {
	fmt.Println("[DEBUG] sendChallenge called")
	
	// 生成 nonce
	nonce, err := s.handler.nonceStore.Generate("")
	if err != nil {
		fmt.Printf("[ERROR] Generate nonce failed: %v\n", err)
		return err
	}
	fmt.Printf("[DEBUG] Nonce generated: %s\n", nonce)

	// 创建挑战消息
	challengeMsg, err := NewMessage(MessageTypeChallengeResponse, ChallengePayload{
		Nonce: nonce,
	})
	if err != nil {
		fmt.Printf("[ERROR] Create message failed: %v\n", err)
		return err
	}
	fmt.Println("[DEBUG] Challenge message created")

	// 序列化消息
	data, err := challengeMsg.Marshal()
	if err != nil {
		fmt.Printf("[ERROR] Marshal message failed: %v\n", err)
		return err
	}
	fmt.Printf("[DEBUG] Message marshaled: %s\n", string(data))

	// 发送挑战
	err = conn.Send(data)
	if err != nil {
		fmt.Printf("[ERROR] Send data failed: %v\n", err)
	} else {
		fmt.Println("[DEBUG] Data sent to channel")
	}
	return err
}
