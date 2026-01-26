package controller

import (
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/service"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
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
// 响应结构
// ============================================

// ValidateTokenResponse 验证 Token 响应
type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
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
		RespondError(w, apperrors.ErrTokenInvalid)
		return
	}

	claims, bizErr := c.authService.ValidateToken(token)
	if bizErr != nil {
		RespondError(w, bizErr)
		return
	}

	RespondSuccess(w, ValidateTokenResponse{
		Valid:    true,
		Username: claims.Username,
	})
}
