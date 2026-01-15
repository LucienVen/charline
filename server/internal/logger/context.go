package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// contextKey 上下文键类型
type contextKey string

// RequestIDKey 请求 ID 上下文键
const RequestIDKey contextKey = "requestID"

// SetRequestID 设置请求 ID 到上下文
func SetRequestID(r *http.Request, requestID string) {
	ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
	*r = *r.WithContext(ctx)
}

// GetRequestID 从上下文获取请求 ID
func GetRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GenerateRequestID 生成新的请求 ID
func GenerateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// 如果随机生成失败，使用简单的时间戳
		return fmt.Sprintf("req-%x", time.Now().UnixNano())
	}
	return "req-" + hex.EncodeToString(b)
}
