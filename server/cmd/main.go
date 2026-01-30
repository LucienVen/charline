package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LucienVen/charline/pkg/sqlite"
	"github.com/LucienVen/charline/server/internal/config"
	"github.com/LucienVen/charline/server/internal/container"
	serverlogger "github.com/LucienVen/charline/server/internal/logger"
	"github.com/LucienVen/charline/server/internal/middleware"
	"github.com/LucienVen/charline/server/internal/router"
	"github.com/LucienVen/charline/server/internal/validator"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log, err := serverlogger.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// 3. 初始化验证器
	validator.Init()
	log.Info("验证器初始化成功")

	// 4. 初始化数据库
	dbCfg := cfg.GetDBConfig()
	db, err := sqlite.New(sqlite.Config{
		DataDir: dbCfg.DataDir,
		Name:    dbCfg.Name,
	})
	if err != nil {
		log.Error("数据库初始化失败", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("数据库连接成功", zap.String("path", db.Path()))

	// 5. 初始化容器（包含所有依赖）
	container, err := container.NewContainer(cfg, db, log)
	if err != nil {
		log.Error("容器初始化失败", zap.Error(err))
		os.Exit(1)
	}
	log.Info("服务层、控制器层初始化成功")
	log.Info("WebSocket 服务器初始化成功")

	// 6. 初始化中间件
	recovery := middleware.NewRecovery(log)
	log.Info("中间件初始化成功")

	// 7. 记录启动信息
	log.Info("Server starting",
		zap.String("address", cfg.GetAddress()),
		zap.String("env", cfg.Env),
		zap.String("log_level", cfg.LogLevel),
	)

	// 8. 创建路由
	r := router.NewRouter(&router.Routes{
		Controllers: &router.Controllers{
			Invite: container.InviteCtrl,
			Auth:   container.AuthCtrl,
		},
		WSServer: container.WSServer, // WebSocket 服务器
		Logger:   log,
		Middlewares: []func(http.Handler) http.Handler{
			recovery.Middleware,             // Panic 恢复（最外层）
			serverlogger.RequestLogger(log), // 请求日志
		},
	})

	// 9. 创建 HTTP 服务器
	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: r,
	}

	// 10. 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", zap.Error(err))
		}
	}()

	log.Info("Server started", zap.String("address", cfg.GetAddress()))

	// 11. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 12. 优雅关闭
	log.Info("Server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped")
}
