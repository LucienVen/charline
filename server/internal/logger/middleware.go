package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RequestLogger HTTP 请求日志中间件
func RequestLogger(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 生成请求 ID
			requestID := GetRequestID(r)
			if requestID == "" {
				requestID = GenerateRequestID()
				SetRequestID(r, requestID)
			}

			// 创建响应记录器
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 调用下一个处理器
			next.ServeHTTP(recorder, r)

			// 计算耗时
			duration := time.Since(start)

			// 记录请求日志
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", recorder.statusCode),
				zap.Float64("duration_ms", float64(duration.Milliseconds())),
				zap.String("request_id", requestID),
				zap.String("ip", getClientIP(r)),
			)
		})
	}
}

// responseRecorder 响应记录器，用于捕获状态码
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 拦截状态码
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// getClientIP 获取客户端 IP
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// 尝试从 X-Real-IP 获取
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// 使用 RemoteAddr
	return r.RemoteAddr
}
