package session

import (
	"fmt"
	"sync"
	"time"
)

// SessionStore Session 内存存储
// 使用三层索引结构：
// 1. sessions: sessionID -> Session (主索引，唯一数据源)
// 2. userSessions: userID -> []sessionID (支持多设备)
// 3. connSessions: connID -> sessionID (快速查询)
type SessionStore struct {
	mu           sync.RWMutex
	sessions     map[string]*Session  // sessionID -> Session
	userSessions map[int64][]string   // userID -> []sessionID (支持多设备)
	connSessions map[string]string    // connID -> sessionID
}

// NewSessionStore 创建 SessionStore
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions:     make(map[string]*Session),
		userSessions: make(map[int64][]string),
		connSessions: make(map[string]string),
	}
}

// Add 添加 Session
func (s *SessionStore) Add(sess *Session) error {
	if sess == nil {
		return fmt.Errorf("session is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查 sessionID 是否已存在
	if _, exists := s.sessions[sess.ID]; exists {
		return fmt.Errorf("session %s already exists", sess.ID)
	}

	// 检查 connID 是否已被占用
	if existingSessionID, exists := s.connSessions[sess.ConnID]; exists {
		return fmt.Errorf("connID %s already bound to session %s", sess.ConnID, existingSessionID)
	}

	// 添加到主索引
	s.sessions[sess.ID] = sess

	// 添加到用户索引（支持多设备）
	s.userSessions[sess.UserID] = append(s.userSessions[sess.UserID], sess.ID)

	// 添加到连接索引
	s.connSessions[sess.ConnID] = sess.ID

	return nil
}

// Get 根据 sessionID 获取 Session
func (s *SessionStore) Get(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, exists := s.sessions[sessionID]
	return sess, exists
}

// GetByUser 获取用户的所有 Session（所有设备）
func (s *SessionStore) GetByUser(userID int64) []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionIDs := s.userSessions[userID]
	if len(sessionIDs) == 0 {
		return nil
	}

	sessions := make([]*Session, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		if sess, ok := s.sessions[id]; ok {
			sessions = append(sessions, sess)
		}
	}

	return sessions
}

// GetByConn 根据 connID 获取 Session
func (s *SessionStore) GetByConn(connID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionID, exists := s.connSessions[connID]
	if !exists {
		return nil, false
	}

	sess, exists := s.sessions[sessionID]
	return sess, exists
}

// Update 更新 Session
func (s *SessionStore) Update(sess *Session) error {
	if sess == nil {
		return fmt.Errorf("session is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查 Session 是否存在
	existing, exists := s.sessions[sess.ID]
	if !exists {
		return fmt.Errorf("session %s not found", sess.ID)
	}

	// 如果 ConnID 发生变化（断线恢复场景）
	if existing.ConnID != sess.ConnID {
		// 删除旧的 connID 映射
		delete(s.connSessions, existing.ConnID)

		// 添加新的 connID 映射
		s.connSessions[sess.ConnID] = sess.ID
	}

	// 更新主索引
	s.sessions[sess.ID] = sess

	return nil
}

// Remove 删除 Session
func (s *SessionStore) Remove(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取 Session
	sess, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 从主索引删除
	delete(s.sessions, sessionID)

	// 从用户索引删除
	userSessionIDs := s.userSessions[sess.UserID]
	for i, id := range userSessionIDs {
		if id == sessionID {
			// 删除该元素
			s.userSessions[sess.UserID] = append(userSessionIDs[:i], userSessionIDs[i+1:]...)
			break
		}
	}

	// 如果用户没有其他 Session，删除用户索引
	if len(s.userSessions[sess.UserID]) == 0 {
		delete(s.userSessions, sess.UserID)
	}

	// 从连接索引删除
	delete(s.connSessions, sess.ConnID)

	return nil
}

// RemoveExpired 删除过期的 Session
// 返回删除的 Session 数量
func (s *SessionStore) RemoveExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredIDs := make([]string, 0)

	// 查找过期的 Session
	for id, sess := range s.sessions {
		if sess.IsExpired() {
			expiredIDs = append(expiredIDs, id)
		}
	}

	// 删除过期的 Session
	for _, id := range expiredIDs {
		sess := s.sessions[id]

		// 从主索引删除
		delete(s.sessions, id)

		// 从用户索引删除
		userSessionIDs := s.userSessions[sess.UserID]
		for i, sid := range userSessionIDs {
			if sid == id {
				s.userSessions[sess.UserID] = append(userSessionIDs[:i], userSessionIDs[i+1:]...)
				break
			}
		}

		// 如果用户没有其他 Session，删除用户索引
		if len(s.userSessions[sess.UserID]) == 0 {
			delete(s.userSessions, sess.UserID)
		}

		// 从连接索引删除
		delete(s.connSessions, sess.ConnID)
	}

	_ = now // 避免未使用变量警告
	return len(expiredIDs)
}

// Count 返回当前 Session 总数
func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sessions)
}

// CountByUser 返回指定用户的 Session 数量
func (s *SessionStore) CountByUser(userID int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.userSessions[userID])
}

// GetAllActive 获取所有活跃的 Session
func (s *SessionStore) GetAllActive() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		if sess.IsActive() {
			sessions = append(sessions, sess)
		}
	}

	return sessions
}
