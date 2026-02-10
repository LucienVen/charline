package websocket

import (
	"github.com/LucienVen/charline/server/internal/session"
)

// SetSessionInfo 设置 Session 信息（认证/Resume 成功后调用）
func (c *Connection) SetSessionInfo(sessionID string, sessionManager session.SessionManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = sessionID
	c.sessionManager = sessionManager
}

// GetSessionID 获取 Session ID
func (c *Connection) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}
