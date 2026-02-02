package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
)

// Config 应用配置
type Config struct {
	Env       string // 环境: development | production
	LogLevel  string // 日志级别: debug | info | warn | error
	LogFormat string // 日志格式: console | json
	Port      int    // 服务器端口
	JWTSecret string // JWT 签名密钥
	DBPath    string // 数据库文件路径（完整路径）
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	// 获取项目根目录
	projectRoot := findProjectRoot()
	
	// 尝试加载 .env 文件（按优先级）
	// 1. 项目根目录/.env
	// 2. 当前目录/.env
	// 3. server/.env
	_ = godotenv.Load(filepath.Join(projectRoot, ".env"))
	_ = godotenv.Load()
	_ = godotenv.Load(filepath.Join(projectRoot, "server/.env"))

	cfg := &Config{
		Env:       getEnv("ENV", "development"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "console"),
		Port:      getEnvAsInt("SERVER_PORT", 8080),
		JWTSecret: getEnv("JWT_SECRET", ""),
		// DB_PATH 优先使用环境变量，否则使用基于项目根目录的路径
		DBPath: getEnv("DB_PATH", filepath.Join(projectRoot, "server/data/server.db")),
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

	// 验证 JWT 密钥
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET 不能为空")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET 长度不足（至少32字节）")
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

// GetJWTSecret 获取 JWT 密钥
func (c *Config) GetJWTSecret() string {
	return c.JWTSecret
}

// GetDBConfig 获取数据库配置
func (c *Config) GetDBConfig() DBConfig {
	// 从完整路径中解析目录和文件名
	dir := filepath.Dir(c.DBPath)
	name := filepath.Base(c.DBPath)

	return DBConfig{
		DataDir: dir,
		Name:    name,
	}
}

// DBConfig 数据库配置
type DBConfig struct {
	DataDir string // 数据目录
	Name    string // 数据库文件名
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
