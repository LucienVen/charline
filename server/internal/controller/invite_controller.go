package controller

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/service"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
)

// InviteController 邀请码控制器
type InviteController struct {
	inviteService *service.InviteService
	logger        *logger.Logger
}

// NewInviteController 创建邀请码控制器实例
func NewInviteController(
	inviteService *service.InviteService,
	log *logger.Logger,
) *InviteController {
	return &InviteController{
		inviteService: inviteService,
		logger:        log,
	}
}

// ============================================
// 请求/响应结构
// ============================================

// GenerateInviteCodeResponse 生成邀请码响应
type GenerateInviteCodeResponse struct {
	Code string `json:"code"`
}

// ActivateInviteCodeRequest 激活邀请码请求
type ActivateInviteCodeRequest struct {
	Code     string `json:"code"`
	Username string `json:"username"`
}

// ActivateInviteCodeResponse 激活邀请码响应
type ActivateInviteCodeResponse struct {
	Token   string `json:"token"`
	Version int    `json:"version"`
}

// ============================================
// HTTP 处理器
// ============================================

// GenerateInviteCode 生成邀请码
// POST /api/invite/generate
func (c *InviteController) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	code, bizErr := c.inviteService.Generate()
	if bizErr != nil {
		RespondError(w, bizErr)
		return
	}

	RespondSuccess(w, GenerateInviteCodeResponse{Code: code})
}

// ActivateInviteCode 激活邀请码
// POST /api/invite/activate
func (c *InviteController) ActivateInviteCode(w http.ResponseWriter, r *http.Request) {
	var req ActivateInviteCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, apperrors.ErrInvalidJSON)
		return
	}

	// 验证用户名格式（带详细错误信息）
	if bizErr := ValidateUsername(req.Username); bizErr != nil {
		RespondError(w, bizErr)
		return
	}

	token, version, bizErr := c.inviteService.Activate(req.Code, req.Username)
	if bizErr != nil {
		RespondError(w, bizErr)
		return
	}

	RespondSuccess(w, ActivateInviteCodeResponse{
		Token:   token,
		Version: version,
	})
}

// ============================================
// 验证逻辑
// ============================================

// ValidateUsername 验证用户名格式，返回详细的业务错误
// 规则: 3-20字符，仅允许字母、数字、下划线，必须以字母开头
func ValidateUsername(username string) *apperrors.BizError {
	// 检查长度
	if len(username) < 3 {
		return apperrors.ErrInvalidUsername.
			WithDetail("reason", "用户名长度必须在 3-20 字符之间").
			WithDetail("field", "username").
			WithDetail("provided", username).
			WithDetail("length", len(username)).
			WithDetail("min_length", 3).
			WithDetail("max_length", 20)
	}
	if len(username) > 20 {
		return apperrors.ErrInvalidUsername.
			WithDetail("reason", "用户名长度必须在 3-20 字符之间").
			WithDetail("field", "username").
			WithDetail("provided", username).
			WithDetail("length", len(username)).
			WithDetail("min_length", 3).
			WithDetail("max_length", 20)
	}

	// 检查格式：必须以字母开头
	matched, _ := regexp.MatchString(`^[a-zA-Z]`, username)
	if !matched {
		return apperrors.ErrInvalidUsername.
			WithDetail("reason", "用户名必须以字母开头").
			WithDetail("field", "username").
			WithDetail("provided", username)
	}

	// 检查格式：只允许字母、数字、下划线
	matched, _ = regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
	if !matched {
		return apperrors.ErrInvalidUsername.
			WithDetail("reason", "用户名只能包含字母、数字、下划线").
			WithDetail("field", "username").
			WithDetail("provided", username)
	}

	return nil // 验证通过
}

// IsValidUsername 验证用户名格式（简化版本，仅返回 bool）
// 规则: 3-20字符，仅允许字母、数字、下划线，必须以字母开头
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
	return matched
}
