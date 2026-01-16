package logger

import "go.uber.org/zap/zapcore"

// Config 通用日志配置接口
type Config interface {
	IsDevelopment() bool
	GetLogLevel() zapcore.Level
}
