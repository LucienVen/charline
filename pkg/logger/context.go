package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// contextKey 上下文键类型
type contextKey string

// RequestIDKey 请求 ID 上下文键
const RequestIDKey contextKey = "requestID"

// SetRequestID 设置请求 ID 到上下文
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID 从上下文获取请求 ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
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
