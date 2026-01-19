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
	"github.com/LucienVen/charline/server/internal/api"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/config"
	serverlogger "github.com/LucienVen/charline/server/internal/logger"
	"github.com/LucienVen/charline/server/internal/store"
	"github.com/go-chi/chi/v5"
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

	// 3. 初始化数据库
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

	// 4. 初始化 JWT 管理器
	jwtManager := auth.NewManager(cfg.GetJWTSecret())
	log.Info("JWT 管理器初始化成功")

	// 5. 初始化邀请码存储
	inviteStore := store.NewInviteStore(db, log)
	log.Info("邀请码存储初始化成功")

	// 6. 初始化 API 处理器
	apiHandler := api.NewHandler(inviteStore, jwtManager, log)
	log.Info("API 处理器初始化成功")

	// 7. 记录启动信息
	log.Info("Server starting",
		zap.String("address", cfg.GetAddress()),
		zap.String("env", cfg.Env),
		zap.String("log_level", cfg.LogLevel),
	)

	// 8. 创建路由
	r := chi.NewRouter()

	// 9. 注册中间件
	r.Use(serverlogger.RequestLogger(log))

	// 10. 注册路由
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API 路由
	r.Post("/api/invite/generate", apiHandler.GenerateInviteCode)
	r.Post("/api/invite/activate", apiHandler.ActivateInviteCode)
	r.Get("/api/validate", apiHandler.ValidateToken)

	// 11. 创建 HTTP 服务器
	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: r,
	}

	// 12. 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", zap.Error(err))
		}
	}()

	log.Info("Server started", zap.String("address", cfg.GetAddress()))

	// 13. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 14. 优雅关闭
	log.Info("Server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped")
}
