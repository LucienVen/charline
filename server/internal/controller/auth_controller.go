package controller

import (
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/httputil"
	"github.com/LucienVen/charline/server/internal/service"
	"github.com/LucienVen/charline/server/internal/validator"
	"go.uber.org/zap"
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

// GetChallenge 获取登录挑战
// GET /api/v1/auth/challenge
func (c *AuthController) GetChallenge(w http.ResponseWriter, r *http.Request) {
	// 从 Authorization 头提取 Token
	authHeader := r.Header.Get("Authorization")
	token, err := auth.ParseTokenFromRequest(authHeader)
	if err != nil {
		c.logger.Warn("Token 解析失败",
			zap.String("auth_header", authHeader),
			zap.Int("error_code", err.Code))
		httputil.RespondError(w, err)
		return
	}

	// 调用 service 获取 challenge
	result, bizErr := c.authService.GetChallenge(token)
	if bizErr != nil {
		c.logger.Error("获取 challenge 失败",
			zap.Int("error_code", bizErr.Code),
			zap.Error(bizErr))
		httputil.RespondError(w, bizErr)
		return
	}

	c.logger.Info("Challenge 生成成功",
		zap.String("nonce", result.Nonce),
		zap.Int("expires_in", result.ExpiresIn))

	httputil.RespondSuccess(w, &httputil.ChallengeResponse{
		Nonce:     result.Nonce,
		ExpiresIn: result.ExpiresIn,
	})
}

// Login 登录验证
// POST /api/v1/auth/login
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req httputil.LoginRequest
	if !httputil.DecodeJSON(w, r, &req) {
		return
	}

	// 统一验证
	if err := validator.Validate(req); err != nil {
		validationErrors := validator.ParseError(err)
		c.logger.Warn("请求参数验证失败",
			zap.Any("validation_errors", validationErrors))
		httputil.RespondError(w,
			errors.ErrInvalidParam.WithDetails(map[string]interface{}{
				"validation_errors": validationErrors,
			}))
		return
	}

	// 调用 service 进行登录验证
	result, bizErr := c.authService.Login(req.Nonce, req.Signature)
	if bizErr != nil {
		c.logger.Error("登录验证失败",
			zap.String("nonce", req.Nonce),
			zap.Int("error_code", bizErr.Code),
			zap.Error(bizErr))
		httputil.RespondError(w, bizErr)
		return
	}

	c.logger.Info("登录成功",
		zap.Int("version", result.Version))

	httputil.RespondSuccess(w, &httputil.LoginResponse{
		Token:   result.Token,
		Version: result.Version,
	})
}
