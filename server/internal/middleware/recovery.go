package middleware

import (
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/httputil"
	"go.uber.org/zap"
)

// Recovery Panic 恢复中间件
// 捕获 panic 并转换为 500 错误响应，防止服务崩溃
type Recovery struct {
	logger *logger.Logger
}

// NewRecovery 创建 Recovery 中间件实例
func NewRecovery(log *logger.Logger) *Recovery {
	return &Recovery{
		logger: log,
	}
}

// Middleware 返回中间件函数
func (m *Recovery) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 详情
				m.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("query", r.URL.RawQuery),
				)

				// 返回 500 错误响应
				httputil.RespondError(w, errors.ErrSystemError.WithDetail("reason", "内部服务器错误"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
