package logger

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/LucienVen/charline/pkg/logger"
	"go.uber.org/zap"
)

// RequestLogger HTTP 请求日志中间件
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 生成请求 ID
			requestID := logger.GetRequestID(r.Context())
			if requestID == "" {
				requestID = logger.GenerateRequestID()
				ctx := logger.SetRequestID(r.Context(), requestID)
				r = r.WithContext(ctx)
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

			// 根据状态码选择日志级别
			// 2xx/3xx -> INFO (正常)
			// 4xx -> WARN (客户端错误)
			// 5xx -> ERROR (服务器错误)
			switch {
			case recorder.statusCode >= 500:
				log.Error("HTTP request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", recorder.statusCode),
					zap.Float64("duration_ms", float64(duration.Milliseconds())),
					zap.String("request_id", requestID),
					zap.String("ip", getClientIP(r)),
				)
			case recorder.statusCode >= 400:
				log.Warn("HTTP request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", recorder.statusCode),
					zap.Float64("duration_ms", float64(duration.Milliseconds())),
					zap.String("request_id", requestID),
					zap.String("ip", getClientIP(r)),
				)
			default:
				log.Info("HTTP request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", recorder.statusCode),
					zap.Float64("duration_ms", float64(duration.Milliseconds())),
					zap.String("request_id", requestID),
					zap.String("ip", getClientIP(r)),
				)
			}
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

// Hijack 实现 http.Hijacker 接口（WebSocket 需要）
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
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
