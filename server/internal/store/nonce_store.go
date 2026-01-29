package store

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	"time"
)

// NonceEntry Nonce 存储条目
type NonceEntry struct {
	UserID    string
	ExpiresAt time.Time
	Used      bool
}

// NonceStore Nonce 内存存储（防重放）
type NonceStore struct {
	mu   sync.RWMutex
	data map[string]NonceEntry
}

// NewNonceStore 创建 Nonce 存储
func NewNonceStore() *NonceStore {
	return &NonceStore{
		data: make(map[string]NonceEntry),
	}
}

// Generate 生成新的 nonce
// 长度: 32 字节 (Base64 后 ~44 字符)
// 过期: 30 秒
func (s *NonceStore) Generate(userID string) (string, error) {
	// 生成随机字节
	randBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, randBytes); err != nil {
		return "", err
	}
	nonce := base64.StdEncoding.EncodeToString(randBytes)

	// 存储
	s.mu.Lock()
	s.data[nonce] = NonceEntry{
		UserID:    userID,
		ExpiresAt: time.Now().Add(30 * time.Second),
		Used:      false,
	}
	s.mu.Unlock()

	return nonce, nil
}

// Consume 消费 nonce（原子操作，验证后删除）
func (s *NonceStore) Consume(nonce string) (*NonceEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[nonce]
	if !exists {
		return nil, false // 不存在或已使用
	}

	if entry.Used {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(s.data, nonce)
		return nil, false // 已过期
	}

	// 标记为已使用并删除
	entry.Used = true
	delete(s.data, nonce)

	return &entry, true
}

// Cleanup 清理过期 nonce（定期调用）
func (s *NonceStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for nonce, entry := range s.data {
		if now.After(entry.ExpiresAt) {
			delete(s.data, nonce)
		}
	}
}
