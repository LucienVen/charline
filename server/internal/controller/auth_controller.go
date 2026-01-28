package controller

import (
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/httputil"
	"github.com/LucienVen/charline/server/internal/service"
)

// AuthController 认证控制器
type AuthController struct {
	authService *service.AuthService
	logger      *logger.Logger
}

// NewAuthController 创建认证控制器实例
func NewAuthController(
	authService *service.AuthService,
	log *logger.Logger,
) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      log,
	}
}

// ============================================
// HTTP 处理器
// ============================================

// ValidateToken 验证 Token
// GET /api/validate
func (c *AuthController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// 从 Authorization 头提取 Token
	authHeader := r.Header.Get("Authorization")
	token, err := auth.ParseTokenFromRequest(authHeader)
	if err != nil {
		httputil.RespondError(w, errors.ErrTokenInvalid)
		return
	}

	claims, bizErr := c.authService.ValidateToken(token)
	if bizErr != nil {
		httputil.RespondError(w, bizErr)
		return
	}

	httputil.RespondSuccess(w, &httputil.ValidateTokenResponse{
		Valid:    true,
		Username: claims.Username,
	})
}
