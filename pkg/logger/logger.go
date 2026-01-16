package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger zap 日志封装
type Logger struct {
	*zap.Logger
}

// New 创建新的日志实例
func New(cfg Config) (*Logger, error) {
	// 构建 encoder 配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据环境选择编码器
	var encoder zapcore.Encoder
	if cfg.IsDevelopment() {
		// 开发环境: console 格式，更易读
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// 生产环境: json 格式
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 输出到 stdout
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), cfg.GetLogLevel())

	// 创建 logger
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过包装层
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{
		Logger: zapLogger,
	}, nil
}

// customTimeEncoder 自定义时间编码器（本地时区）
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 格式: +0800 2025-01-15 17:35:12
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	enc.AppendString(fmt.Sprintf("%s%02d%02d %s",
		sign, hours, minutes,
		t.Format("2006-01-02 15:04:05"),
	))
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// With 创建带字段的子 logger
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}
