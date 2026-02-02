// Package session 提供 WebSocket 会话管理功能
package session

import (
	"time"

	"github.com/google/uuid"
)

// SessionState 表示会话状态
type SessionState int

const (
	// SessionStateActive 活跃状态（连接正常）
	SessionStateActive SessionState = iota
	// SessionStateSuspended 挂起状态（连接断开，等待恢复）
	SessionStateSuspended
	// SessionStateClosed 关闭状态（会话已结束）
	SessionStateClosed
)

// String 返回状态的字符串表示
func (s SessionState) String() string {
	switch s {
	case SessionStateActive:
		return "active"
	case SessionStateSuspended:
		return "suspended"
	case SessionStateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Session 表示一个 WebSocket 会话
type Session struct {
	ID           string       // 会话唯一标识（UUID）
	UserID       int64        // 用户 ID
	DeviceID     string       // 设备标识（可选）
	ConnID       string       // WebSocket 连接 ID
	State        SessionState // 会话状态
	CreatedAt    time.Time    // 创建时间
	LastActiveAt time.Time    // 最后活跃时间
	ExpiresAt    time.Time    // 会话过期时间
	ResumeToken  string       // 断线恢复 Token
	ResumeExpiry time.Time    // Resume Token 过期时间
}

// 默认配置
const (
	// DefaultSessionTTL 默认会话有效期（8小时）
	DefaultSessionTTL = 8 * time.Hour
	// DefaultResumeTTL 默认 Resume Token 有效期（30秒）
	DefaultResumeTTL = 30 * time.Second
)

// NewSession 创建新会话
func NewSession(userID int64, deviceID, connID string) *Session {
	now := time.Now()
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		DeviceID:     deviceID,
		ConnID:       connID,
		State:        SessionStateActive,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(DefaultSessionTTL),
	}
}

// IsActive 检查会话是否处于活跃状态
func (s *Session) IsActive() bool {
	return s.State == SessionStateActive
}

// IsSuspended 检查会话是否处于挂起状态
func (s *Session) IsSuspended() bool {
	return s.State == SessionStateSuspended
}

// IsClosed 检查会话是否已关闭
func (s *Session) IsClosed() bool {
	return s.State == SessionStateClosed
}

// IsExpired 检查会话是否已过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsResumeTokenValid 检查 Resume Token 是否有效
func (s *Session) IsResumeTokenValid() bool {
	if s.ResumeToken == "" {
		return false
	}
	return time.Now().Before(s.ResumeExpiry)
}

// Touch 更新最后活跃时间
func (s *Session) Touch() {
	s.LastActiveAt = time.Now()
}

// Suspend 挂起会话，生成 Resume Token
func (s *Session) Suspend(resumeToken string) {
	s.State = SessionStateSuspended
	s.ResumeToken = resumeToken
	s.ResumeExpiry = time.Now().Add(DefaultResumeTTL)
	s.ConnID = "" // 清除连接 ID
}

// Resume 恢复会话
func (s *Session) Resume(newConnID string) {
	s.State = SessionStateActive
	s.ConnID = newConnID
	s.ResumeToken = ""
	s.ResumeExpiry = time.Time{}
	s.Touch()
}

// Close 关闭会话
func (s *Session) Close() {
	s.State = SessionStateClosed
	s.ConnID = ""
	s.ResumeToken = ""
}
