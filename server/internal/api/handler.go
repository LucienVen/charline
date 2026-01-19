package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/auth"
	"github.com/LucienVen/charline/server/internal/store"
	"github.com/LucienVen/charline/pkg/logger"
	"go.uber.org/zap"
)

// Handler API 处理器
type Handler struct {
	inviteStore *store.InviteStore
	jwtManager  *auth.Manager
	logger      *logger.Logger
}

// NewHandler 创建 API 处理器实例
func NewHandler(
	inviteStore *store.InviteStore,
	jwtManager *auth.Manager,
	log *logger.Logger,
) *Handler {
	return &Handler{
		inviteStore: inviteStore,
		jwtManager:  jwtManager,
		logger:      log,
	}
}

// ============================================
// 请求/响应结构
// ============================================

// generateInviteResponse 生成邀请码响应
type generateInviteResponse struct {
	Code string `json:"code"`
}

// activateInviteRequest 激活邀请码请求
type activateInviteRequest struct {
	Code     string `json:"code"`
	Username string `json:"username"`
}

// activateInviteResponse 激活邀请码响应
type activateInviteResponse struct {
	Token   string `json:"token"`
	Version int    `json:"version"`
}

// validateTokenResponse 验证 Token 响应
type validateTokenResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
}

// errorResponse 错误响应
type errorResponse struct {
	Error string `json:"error"`
}

// ============================================
// API 端点
// ============================================

// GenerateInviteCode 生成邀请码
// POST /api/invite/generate
func (h *Handler) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	code, err := h.inviteStore.Generate()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "生成邀请码失败")
		h.logger.Error("生成邀请码失败",
			zap.String("error", err.Error()))
		return
	}

	h.respondJSON(w, http.StatusOK, generateInviteResponse{Code: code})
	h.logger.Info("生成邀请码成功",
		zap.String("code", code))
}

// ActivateInviteCode 激活邀请码
// POST /api/invite/activate
func (h *Handler) ActivateInviteCode(w http.ResponseWriter, r *http.Request) {
	var req activateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	// 验证用户名格式
	if !isValidUsername(req.Username) {
		h.respondError(w, http.StatusBadRequest, apperrors.ErrUsernameInvalid.Error())
		return
	}

	// 激活邀请码
	if err := h.inviteStore.Activate(req.Code, req.Username); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInviteNotFound):
			h.respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrInviteAlreadyUsed):
			h.respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrInviteInvalid):
			h.respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, "激活邀请码失败")
		}
		h.logger.Error("激活邀请码失败",
			zap.String("code", req.Code),
			zap.String("username", req.Username),
			zap.String("error", err.Error()))
		return
	}

	// 生成 JWT Token
	token, err := h.jwtManager.GenerateToken(req.Username, 1)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "生成 Token 失败")
		h.logger.Error("生成 Token 失败",
			zap.String("username", req.Username),
			zap.String("error", err.Error()))
		return
	}

	h.respondJSON(w, http.StatusOK, activateInviteResponse{
		Token:   token,
		Version: 1,
	})
	h.logger.Info("激活邀请码成功",
		zap.String("code", req.Code),
		zap.String("username", req.Username))
}

// ValidateToken 验证 Token
// GET /api/validate
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// 从 Authorization 头提取 Token
	authHeader := r.Header.Get("Authorization")
	token, err := auth.ParseTokenFromRequest(authHeader)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, apperrors.ErrTokenMalformed.Error())
		return
	}

	// 验证 Token
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrTokenExpired):
			h.respondError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, apperrors.ErrTokenInvalid):
			h.respondError(w, http.StatusUnauthorized, err.Error())
		default:
			h.respondError(w, http.StatusUnauthorized, "Token 验证失败")
		}
		h.logger.Warn("Token 验证失败",
			zap.String("error", err.Error()))
		return
	}

	h.respondJSON(w, http.StatusOK, validateTokenResponse{
		Valid:    true,
		Username: claims.Username,
	})
	h.logger.Info("Token 验证成功",
		zap.String("username", claims.Username))
}

// ============================================
// 辅助方法
// ============================================

// respondJSON JSON 响应
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError 错误响应
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, errorResponse{Error: message})
}

// isValidUsername 验证用户名格式
// 规则: 3-20字符，仅允许字母、数字、下划线，必须以字母开头
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
	return matched
}
