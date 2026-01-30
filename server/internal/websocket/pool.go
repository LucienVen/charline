package websocket

import (
	"sync"
)

// ConnectionPool WebSocket 连接池
type ConnectionPool struct {
	connections map[int64]*Connection // userID -> Connection
	mu          sync.RWMutex          // 读写锁
}

// NewConnectionPool 创建连接池
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[int64]*Connection),
	}
}

// Add 添加连接
func (p *ConnectionPool) Add(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn.userID > 0 {
		p.connections[conn.userID] = conn
	}
}

// Remove 移除连接
func (p *ConnectionPool) Remove(userID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.connections, userID)
}

// GetByUser 根据用户ID获取连接
func (p *ConnectionPool) GetByUser(userID int64) (*Connection, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	conn, exists := p.connections[userID]
	return conn, exists
}

// Count 获取连接数
func (p *ConnectionPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.connections)
}

// GetAll 获取所有连接（用于广播等场景）
func (p *ConnectionPool) GetAll() []*Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	conns := make([]*Connection, 0, len(p.connections))
	for _, conn := range p.connections {
		conns = append(conns, conn)
	}
	return conns
}
