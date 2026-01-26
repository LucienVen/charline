package router

import (
	"fmt"
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/controller"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Controllers 控制器集合
type Controllers struct {
	Invite *controller.InviteController
	Auth   *controller.AuthController
}

// Routes 路由依赖注入结构
type Routes struct {
	Controllers *Controllers                       // 控制器集合
	Logger      *logger.Logger                     // 日志
	Middlewares []func(http.Handler) http.Handler // 全局中间件
}

// NewRouter 创建并配置路由
func NewRouter(routes *Routes) chi.Router {
	r := chi.NewRouter()

	// 注册全局中间件
	for _, mw := range routes.Middlewares {
		r.Use(mw)
	}

	// 健康检查路由（无需认证）
	r.Get("/health", healthHandler)

	// API 路由分组
	r.Route("/api/v1", func(r chi.Router) {
		// 邀请相关路由
		r.Route("/invite", func(r chi.Router) {
			r.Post("/generate", routes.Controllers.Invite.GenerateInviteCode)
			r.Post("/activate", routes.Controllers.Invite.ActivateInviteCode)
		})

		// 认证相关路由
		r.Get("/validate", routes.Controllers.Auth.ValidateToken)
	})

	// 开发环境打印所有路由
	printRoutes(r, routes.Logger)

	return r
}

// healthHandler 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// printRoutes 打印所有路由（开发环境）
func printRoutes(r chi.Router, log *logger.Logger) {
	if log == nil {
		return
	}

	log.Info("=== Registered Routes ===")
	if err := chi.Walk(r, printRoute); err != nil {
		log.Error("Failed to walk routes", zap.Error(err))
	}
}

// printRoute 打印单条路由
func printRoute(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	fmt.Printf("  %-6s %s\n", method, route)
	return nil
}
