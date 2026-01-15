package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Config 应用配置
type Config struct {
	Env       string // 环境: development | production
	LogLevel  string // 日志级别: debug | info | warn | error
	LogFormat string // 日志格式: console | json
	Port      int    // 服务器端口
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	cfg := &Config{
		Env:       getEnv("ENV", "development"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "console"),
		Port:      getEnvAsInt("SERVER_PORT", 8080),
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证环境
	if c.Env != "development" && c.Env != "production" {
		return fmt.Errorf("无效的环境: %s (允许: development, production)", c.Env)
	}

	// 验证日志级别
	if !isValidLogLevel(c.LogLevel) {
		return fmt.Errorf("无效的日志级别: %s (允许: debug, info, warn, error)", c.LogLevel)
	}

	// 验证日志格式
	if c.LogFormat != "console" && c.LogFormat != "json" {
		return fmt.Errorf("无效的日志格式: %s (允许: console, json)", c.LogFormat)
	}

	// 验证端口
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", c.Port)
	}

	return nil
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// GetZapLevel 获取 zap 日志级别
func (c *Config) GetZapLevel() zapcore.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetAddress 获取服务器监听地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取整数环境变量
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// isValidLogLevel 验证日志级别是否有效
func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
