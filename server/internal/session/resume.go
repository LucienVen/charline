package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

const (
	// ResumeTokenLength Resume Token 长度（字节）
	ResumeTokenLength = 32

	// ResumeTokenTTL Resume Token 有效期（30秒）
	ResumeTokenTTL = 30 * time.Second
)

// ResumeTokenStore Resume Token 存储
// 用于断线恢复场景，Token 有效期 30 秒
type ResumeTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*ResumeTokenEntry // token -> entry
}

// ResumeTokenEntry Resume Token 条目
type ResumeTokenEntry struct {
	Token     string    // Resume Token
	SessionID string    // 关联的 Session ID
	ExpiresAt time.Time // 过期时间
}

// NewResumeTokenStore 创建 ResumeTokenStore
func NewResumeTokenStore() *ResumeTokenStore {
	return &ResumeTokenStore{
		tokens: make(map[string]*ResumeTokenEntry),
	}
}

// Generate 生成 Resume Token
// 返回 token 和过期时间
func (r *ResumeTokenStore) Generate(sessionID string) (string, time.Time, error) {
	// 生成 32 字节随机 Token
	tokenBytes := make([]byte, ResumeTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, fmt.Errorf("generate random token failed: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(ResumeTokenTTL)

	r.mu.Lock()
	defer r.mu.Unlock()

	// 存储 Token
	r.tokens[token] = &ResumeTokenEntry{
		Token:     token,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}

	return token, expiresAt, nil
}

// Consume 消费 Resume Token（原子操作）
// 成功返回 sessionID，失败返回错误
// Token 只能使用一次，使用后立即删除
func (r *ResumeTokenStore) Consume(token string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 查找 Token
	entry, exists := r.tokens[token]
	if !exists {
		return "", fmt.Errorf("resume token not found")
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		// 删除过期 Token
		delete(r.tokens, token)
		return "", fmt.Errorf("resume token expired")
	}

	// 获取 sessionID
	sessionID := entry.SessionID

	// 原子消费：立即删除 Token
	delete(r.tokens, token)

	return sessionID, nil
}

// Revoke 撤销 Resume Token
// 用于主动关闭 Session 时清理 Token
func (r *ResumeTokenStore) Revoke(token string) error {
	if token == "" {
		return nil // 空 Token 不需要撤销
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tokens[token]; !exists {
		return fmt.Errorf("resume token not found")
	}

	delete(r.tokens, token)
	return nil
}

// RevokeBySession 撤销指定 Session 的所有 Resume Token
func (r *ResumeTokenStore) RevokeBySession(sessionID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	revokedCount := 0
	tokensToDelete := make([]string, 0)

	// 查找该 Session 的所有 Token
	for token, entry := range r.tokens {
		if entry.SessionID == sessionID {
			tokensToDelete = append(tokensToDelete, token)
		}
	}

	// 删除 Token
	for _, token := range tokensToDelete {
		delete(r.tokens, token)
		revokedCount++
	}

	return revokedCount
}

// CleanExpired 清理过期的 Resume Token
// 返回清理的 Token 数量
func (r *ResumeTokenStore) CleanExpired() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiredTokens := make([]string, 0)

	// 查找过期的 Token
	for token, entry := range r.tokens {
		if now.After(entry.ExpiresAt) {
			expiredTokens = append(expiredTokens, token)
		}
	}

	// 删除过期的 Token
	for _, token := range expiredTokens {
		delete(r.tokens, token)
	}

	return len(expiredTokens)
}

// Count 返回当前 Token 总数
func (r *ResumeTokenStore) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tokens)
}

// IsValid 检查 Token 是否有效（不消费）
func (r *ResumeTokenStore) IsValid(token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.tokens[token]
	if !exists {
		return false
	}

	// 检查是否过期
	return time.Now().Before(entry.ExpiresAt)
}
