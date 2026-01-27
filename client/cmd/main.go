package main

import (
	"fmt"
	"os"

	"github.com/LucienVen/charline/client/internal/commands"
	"github.com/LucienVen/charline/client/internal/config"
	pkglogger "github.com/LucienVen/charline/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := newClientLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// 记录启动信息
	log.Info("Client starting",
		zap.String("env", cfg.Env),
		zap.String("log_level", cfg.LogLevel),
	)

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	cmdArgs := os.Args[2:] // 清晰分离：仅包含命令之后的参数

	switch command {
	case "hello":
		log.Info("Hello command executed")
		fmt.Println("Hello from CharLine Client!")

	case "join":
		handleJoin(log, cfg, cmdArgs)

	default:
		log.Warn("Unknown command", zap.String("command", command))
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}

	log.Info("Client stopped")
}

// newClientLogger 创建客户端日志实例
func newClientLogger(cfg *config.Config) (*pkglogger.Logger, error) {
	adapter := &configAdapter{cfg}
	return pkglogger.New(adapter)
}

// configAdapter 将 client.Config 适配为 logger.Config
type configAdapter struct {
	*config.Config
}

func (a *configAdapter) IsDevelopment() bool {
	return a.Config.IsDevelopment()
}

func (a *configAdapter) GetLogLevel() zapcore.Level {
	return a.Config.GetZapLevel()
}

func handleJoin(log *pkglogger.Logger, cfg *config.Config, args []string) {
	// 验证参数数量
	if len(args) != 2 {
		fmt.Println("Usage: charline join <invite-code> <username>")
		os.Exit(1)
	}

	code := args[0]
	username := args[1]

	log.Info("Executing join command",
		zap.String("username", username),
	)

	// 构造 Join 配置
	joinCfg := &commands.JoinConfig{
		ServerURL: cfg.ServerURL,
		Code:      code,
		Username:  username,
	}

	// 执行 join 流程
	result, err := commands.Join(joinCfg)
	if err != nil {
		log.Error("Join failed",
			zap.String("username", username),
			zap.Error(err),
		)
		fmt.Fprintf(os.Stderr, "Join 失败: %v\n", err)
		os.Exit(1)
	}

	// 成功
	log.Info("Join successful",
		zap.String("username", username),
		zap.Int("version", result.Version),
	)

	fmt.Printf("✓ Join 成功！\n")
	fmt.Printf("  用户名: %s\n", username)
	fmt.Printf("  凭证版本: %d\n", result.Version)
	fmt.Printf("  凭证已保存到 ~/.charline/\n")
}

func printHelp() {
	fmt.Println("CharLine Client")
	fmt.Println("Usage: charline <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  hello                          - Print hello message")
	fmt.Println("  join <invite-code> <username>  - Join with invitation code")
}
