package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
)

// Config 客户端配置
type Config struct {
	Env       string // 环境: development | production
	LogLevel  string // 日志级别: debug | info | warn | error
	LogFormat string // 日志格式: console | json
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	// 尝试加载 .env 文件（如果存在）
	// 按优先级尝试：项目根目录 .env > 当前目录 .env
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		Env:       getEnv("ENV", "development"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "console"),
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

// GetProjectRoot 获取项目根目录
func (c *Config) GetProjectRoot() string {
	return findProjectRoot()
}

// findProjectRoot 查找项目根目录（包含 go.work 文件的目录）
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// 向上查找 go.work 文件
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 到达根目录，返回当前目录
			break
		}
		dir = parent
	}

	// 找不到 go.work，返回当前目录
	return "."
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
