package logger

import (
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/client/internal/config"
	"go.uber.org/zap/zapcore"
)

// New 创建客户端日志实例
func New(cfg *config.Config) (*logger.Logger, error) {
	adapter := &configAdapter{cfg}
	return logger.New(adapter)
}

// configAdapter 将 client.Config 适配为 logger.Config
type configAdapter struct {
	*config.Config
}

// IsDevelopment 判断是否为开发环境
func (a *configAdapter) IsDevelopment() bool {
	return a.Config.IsDevelopment()
}

// GetLogLevel 获取日志级别
func (a *configAdapter) GetLogLevel() zapcore.Level {
	return a.Config.GetZapLevel()
}
