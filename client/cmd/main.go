package main

import (
	"fmt"
	"os"

	"github.com/LucienVen/charline/client/internal/config"
	"github.com/LucienVen/charline/client/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.New(cfg)
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

	switch command {
	case "hello":
		log.Info("Hello command executed")
		fmt.Println("Hello from CharLine Client!")
	default:
		log.Warn("Unknown command", zap.String("command", command))
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	log.Info("Client stopped")
}

func printHelp() {
	fmt.Println("CharLine Client")
	fmt.Println("Usage: client <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  hello    - Print hello message")
}
