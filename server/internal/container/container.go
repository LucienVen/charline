package container

import (
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/config"
	"github.com/LucienVen/charline/server/internal/controller"
	"github.com/LucienVen/charline/server/internal/service"
	"github.com/LucienVen/charline/server/internal/session"
	"github.com/LucienVen/charline/server/internal/store"
	"github.com/LucienVen/charline/server/internal/websocket"
	"github.com/LucienVen/charline/pkg/sqlite"
)

// Container 依赖注入容器
type Container struct {
	InviteCtrl     *controller.InviteController
	AuthCtrl       *controller.AuthController
	WSServer       *websocket.Server        // WebSocket 服务器
	WSPool         *websocket.ConnectionPool // WebSocket 连接池
	SessionManager session.SessionManager   // Session 管理器
}

// NewContainer 创建并初始化所有依赖
func NewContainer(cfg *config.Config, db *sqlite.DB, log *logger.Logger) (*Container, error) {
	// 初始化 JWT 管理器
	jwtManager := auth.NewManager(cfg.GetJWTSecret())

	// 初始化存储层
	inviteStore := store.NewInviteStore(db, log)
	userStore := store.NewUserStore(db, log)
	nonceStore := store.NewNonceStore()

	// 初始化服务层
	inviteService := service.NewInviteService(inviteStore, jwtManager, log)
	authService := service.NewAuthService(jwtManager, userStore, nonceStore, log)

	// 初始化控制器层
	inviteCtrl := controller.NewInviteController(inviteService, log)
	authCtrl := controller.NewAuthController(authService, log)

	// 初始化 Session 管理器（Phase 3.2）
	sessionManager := session.NewManager(log)

	// 初始化 WebSocket 层（Phase 3.1 + 3.2）
	wsPool := websocket.NewConnectionPool()
	wsHandler := websocket.NewWSMessageHandler(nonceStore, userStore, wsPool, sessionManager)
	wsServer := websocket.NewServer(wsHandler)

	return &Container{
		InviteCtrl:     inviteCtrl,
		AuthCtrl:       authCtrl,
		WSServer:       wsServer,
		WSPool:         wsPool,
		SessionManager: sessionManager,
	}, nil
}
