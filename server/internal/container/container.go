package container

import (
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/config"
	"github.com/LucienVen/charline/server/internal/controller"
	"github.com/LucienVen/charline/server/internal/service"
	"github.com/LucienVen/charline/server/internal/store"
	"github.com/LucienVen/charline/pkg/sqlite"
)

// Container 依赖注入容器
type Container struct {
	InviteCtrl *controller.InviteController
	AuthCtrl   *controller.AuthController
}

// NewContainer 创建并初始化所有依赖
func NewContainer(cfg *config.Config, db *sqlite.DB, log *logger.Logger) (*Container, error) {
	// 初始化 JWT 管理器
	jwtManager := auth.NewManager(cfg.GetJWTSecret())

	// 初始化存储层
	inviteStore := store.NewInviteStore(db, log)

	// 初始化服务层
	inviteService := service.NewInviteService(inviteStore, jwtManager, log)
	authService := service.NewAuthService(jwtManager, log)

	// 初始化控制器层
	inviteCtrl := controller.NewInviteController(inviteService, log)
	authCtrl := controller.NewAuthController(authService, log)

	return &Container{
		InviteCtrl: inviteCtrl,
		AuthCtrl:   authCtrl,
	}, nil
}
