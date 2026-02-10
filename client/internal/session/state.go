package session

import (
	"sync"
	"time"
)

// State 客户端 Session 状态
type State struct {
	SessionID    string    // Session ID（服务端返回）
	UserID       int64     // 用户 ID
	DeviceID     string    // 设备标识
	ResumeToken  string    // Resume Token（用于断线恢复）
	ResumeExpiry time.Time // Resume Token 过期时间
	CreatedAt    time.Time // 创建时间
	LastActive   time.Time // 最后活跃时间
	mu           sync.RWMutex
}

// NewState 创建新的 Session 状态
func NewState(sessionID string, userID int64, deviceID string) *State {
	now := time.Now()
	return &State{
		SessionID:  sessionID,
		UserID:     userID,
		DeviceID:   deviceID,
		CreatedAt:  now,
		LastActive: now,
	}
}

// GetSessionID 获取 Session ID
func (s *State) GetSessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionID
}

// GetUserID 获取用户 ID
func (s *State) GetUserID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UserID
}

// GetDeviceID 获取设备 ID
func (s *State) GetDeviceID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.DeviceID
}

// Touch 更新最后活跃时间
func (s *State) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActive = time.Now()
}

// IsValid 检查 Session 是否有效
// Session ID 不为空即认为有效
func (s *State) IsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SessionID != ""
}

// Clear 清空 Session 状态
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SessionID = ""
	s.UserID = 0
	s.DeviceID = ""
	s.ResumeToken = ""
	s.ResumeExpiry = time.Time{}
}

// SetResumeToken 设置 Resume Token
func (s *State) SetResumeToken(token string, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResumeToken = token
	s.ResumeExpiry = expiry
}

// GetResumeToken 获取 Resume Token
// 返回 token 和是否有效
func (s *State) GetResumeToken() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ResumeToken == "" {
		return "", false
	}
	if time.Now().After(s.ResumeExpiry) {
		return "", false
	}
	return s.ResumeToken, true
}

// ClearResumeToken 清除 Resume Token
func (s *State) ClearResumeToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResumeToken = ""
	s.ResumeExpiry = time.Time{}
}

// IsResumeTokenValid 检查 Resume Token 是否有效
func (s *State) IsResumeTokenValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ResumeToken == "" {
		return false
	}
	return time.Now().Before(s.ResumeExpiry)
}

// GetLastActive 获取最后活跃时间
func (s *State) GetLastActive() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastActive
}

// GetCreatedAt 获取创建时间
func (s *State) GetCreatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CreatedAt
}
