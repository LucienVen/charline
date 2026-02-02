package session

import (
	"fmt"
	"time"

	"github.com/LucienVen/charline/pkg/logger"
	"go.uber.org/zap"
)

// SessionManager Session 管理器接口
type SessionManager interface {
	// Create 创建新 Session
	Create(userID int64, deviceID string, connID string) (*Session, error)

	// Get 根据 sessionID 获取 Session
	Get(sessionID string) (*Session, bool)

	// GetByUser 获取用户的所有 Session（所有设备）
	GetByUser(userID int64) []*Session

	// GetByConn 根据 connID 获取 Session
	GetByConn(connID string) (*Session, bool)

	// Touch 更新 Session 最后活跃时间
	Touch(sessionID string) error

	// Suspend 挂起 Session（断线时调用）
	// 返回 Resume Token 和过期时间
	Suspend(sessionID string) (resumeToken string, expiresAt time.Time, err error)

	// Resume 恢复 Session（重连时调用）
	// 使用 Resume Token 恢复，并绑定新的 connID
	Resume(resumeToken string, newConnID string) (*Session, error)

	// Close 关闭 Session（正常关闭）
	Close(sessionID string) error

	// Cleanup 清理过期的 Session 和 Resume Token
	Cleanup()

	// Count 返回当前 Session 总数
	Count() int

	// CountByUser 返回指定用户的 Session 数量
	CountByUser(userID int64) int
}

// Manager SessionManager 实现
type Manager struct {
	sessionStore     *SessionStore
	resumeTokenStore *ResumeTokenStore
	logger           *logger.Logger
	stopCleanup      chan struct{}
}

// NewManager 创建 SessionManager
func NewManager(log *logger.Logger) *Manager {
	m := &Manager{
		sessionStore:     NewSessionStore(),
		resumeTokenStore: NewResumeTokenStore(),
		logger:           log,
		stopCleanup:      make(chan struct{}),
	}

	// 启动后台清理 goroutine
	go m.startCleanupRoutine()

	return m
}

// Create 创建新 Session
func (m *Manager) Create(userID int64, deviceID string, connID string) (*Session, error) {
	// 创建 Session
	sess := NewSession(userID, deviceID, connID)

	// 添加到存储
	if err := m.sessionStore.Add(sess); err != nil {
		return nil, fmt.Errorf("add session failed: %w", err)
	}

	m.logger.Info("Session created",
		zap.String("sessionID", sess.ID),
		zap.Int64("userID", userID),
		zap.String("deviceID", deviceID),
		zap.String("connID", connID))

	return sess, nil
}

// Get 根据 sessionID 获取 Session
func (m *Manager) Get(sessionID string) (*Session, bool) {
	return m.sessionStore.Get(sessionID)
}

// GetByUser 获取用户的所有 Session（所有设备）
func (m *Manager) GetByUser(userID int64) []*Session {
	return m.sessionStore.GetByUser(userID)
}

// GetByConn 根据 connID 获取 Session
func (m *Manager) GetByConn(connID string) (*Session, bool) {
	return m.sessionStore.GetByConn(connID)
}

// Touch 更新 Session 最后活跃时间
func (m *Manager) Touch(sessionID string) error {
	sess, exists := m.sessionStore.Get(sessionID)
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	sess.Touch()

	if err := m.sessionStore.Update(sess); err != nil {
		return fmt.Errorf("update session failed: %w", err)
	}

	return nil
}

// Suspend 挂起 Session（断线时调用）
func (m *Manager) Suspend(sessionID string) (string, time.Time, error) {
	// 1. 获取 Session
	sess, exists := m.sessionStore.Get(sessionID)
	if !exists {
		return "", time.Time{}, fmt.Errorf("session %s not found", sessionID)
	}

	// 2. 检查 Session 状态
	if !sess.IsActive() {
		return "", time.Time{}, fmt.Errorf("session %s is not active", sessionID)
	}

	// 3. 生成 Resume Token
	resumeToken, expiresAt, err := m.resumeTokenStore.Generate(sessionID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate resume token failed: %w", err)
	}

	// 4. 挂起 Session
	sess.Suspend(resumeToken)

	// 5. 更新存储
	if err := m.sessionStore.Update(sess); err != nil {
		// 如果更新失败，撤销 Resume Token
		_ = m.resumeTokenStore.Revoke(resumeToken)
		return "", time.Time{}, fmt.Errorf("update session failed: %w", err)
	}

	m.logger.Info("Session suspended",
		zap.String("sessionID", sessionID),
		zap.String("resumeToken", resumeToken),
		zap.Time("expiresAt", expiresAt))

	return resumeToken, expiresAt, nil
}

// Resume 恢复 Session（重连时调用）
func (m *Manager) Resume(resumeToken string, newConnID string) (*Session, error) {
	// 1. 消费 Resume Token（原子操作）
	sessionID, err := m.resumeTokenStore.Consume(resumeToken)
	if err != nil {
		return nil, fmt.Errorf("consume resume token failed: %w", err)
	}

	// 2. 获取 Session
	sess, exists := m.sessionStore.Get(sessionID)
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// 3. 检查 Session 状态
	if !sess.IsSuspended() {
		return nil, fmt.Errorf("session %s is not suspended", sessionID)
	}

	// 4. 检查 Session 是否过期
	if sess.IsExpired() {
		// 删除过期 Session
		_ = m.sessionStore.Remove(sessionID)
		return nil, fmt.Errorf("session %s expired", sessionID)
	}

	// 5. 恢复 Session
	sess.Resume(newConnID)

	// 6. 更新存储
	if err := m.sessionStore.Update(sess); err != nil {
		return nil, fmt.Errorf("update session failed: %w", err)
	}

	m.logger.Info("Session resumed",
		zap.String("sessionID", sessionID),
		zap.String("newConnID", newConnID))

	return sess, nil
}

// Close 关闭 Session（正常关闭）
func (m *Manager) Close(sessionID string) error {
	// 1. 获取 Session
	sess, exists := m.sessionStore.Get(sessionID)
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 2. 撤销 Resume Token（如果有）
	if sess.ResumeToken != "" {
		_ = m.resumeTokenStore.Revoke(sess.ResumeToken)
	}

	// 3. 关闭 Session
	sess.Close()

	// 4. 从存储删除
	if err := m.sessionStore.Remove(sessionID); err != nil {
		return fmt.Errorf("remove session failed: %w", err)
	}

	m.logger.Info("Session closed", zap.String("sessionID", sessionID))

	return nil
}

// Cleanup 清理过期的 Session 和 Resume Token
func (m *Manager) Cleanup() {
	// 清理过期的 Resume Token
	expiredTokens := m.resumeTokenStore.CleanExpired()
	if expiredTokens > 0 {
		m.logger.Info("Cleaned expired resume tokens", zap.Int("count", expiredTokens))
	}

	// 清理过期的 Session
	expiredSessions := m.sessionStore.RemoveExpired()
	if expiredSessions > 0 {
		m.logger.Info("Cleaned expired sessions", zap.Int("count", expiredSessions))
	}
}

// Count 返回当前 Session 总数
func (m *Manager) Count() int {
	return m.sessionStore.Count()
}

// CountByUser 返回指定用户的 Session 数量
func (m *Manager) CountByUser(userID int64) int {
	return m.sessionStore.CountByUser(userID)
}

// startCleanupRoutine 启动后台清理 goroutine
func (m *Manager) startCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	m.logger.Info("Session cleanup routine started")

	for {
		select {
		case <-ticker.C:
			m.Cleanup()

		case <-m.stopCleanup:
			m.logger.Info("Session cleanup routine stopped")
			return
		}
	}
}

// Stop 停止 SessionManager
func (m *Manager) Stop() {
	close(m.stopCleanup)
	m.logger.Info("SessionManager stopped")
}
