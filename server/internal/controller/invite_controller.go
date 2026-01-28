package controller

import (
	"net/http"
	"regexp"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/httputil"
	"github.com/LucienVen/charline/server/internal/service"
	"go.uber.org/zap"
)

// InviteController 邀请码控制器
type InviteController struct {
	inviteService service.InviteServiceInterface
	logger        *logger.Logger
}

// NewInviteController 创建邀请码控制器实例
func NewInviteController(
	inviteService service.InviteServiceInterface,
	log *logger.Logger,
) *InviteController {
	return &InviteController{
		inviteService: inviteService,
		logger:        log,
	}
}

// ActivateInviteCode 激活邀请码
// POST /api/v1/invite/activate
func (c *InviteController) ActivateInviteCode(w http.ResponseWriter, r *http.Request) {
	var req httputil.ActivateInviteRequest
	if !httputil.DecodeJSON(w, r, &req) {
		return
	}

	// 验证用户名格式
	if !isValidUsername(req.Username) {
		httputil.RespondWithError(w, http.StatusBadRequest, errors.ErrInvalidUsername)
		return
	}

	// 验证公钥格式 (Base64 编码的 Ed25519 公钥约 44 字符)
	if len(req.PublicKey) < 40 || len(req.PublicKey) > 50 {
		httputil.RespondWithError(w, http.StatusBadRequest, errors.ErrInvalidParam.WithDetail("reason", "公钥格式不正确"))
		return
	}

	// 调用 service 激活邀请码
	token, version, err := c.inviteService.Activate(req.Code, req.Username, req.PublicKey)
	if err != nil {
		c.logger.Error("激活邀请码失败",
			zap.String("code", req.Code),
			zap.String("username", req.Username),
			zap.Int("error_code", err.Code),
			zap.Error(err),
		)
		httputil.RespondError(w, err)
		return
	}

	c.logger.Info("邀请码激活成功",
		zap.String("code", req.Code),
		zap.String("username", req.Username))

	httputil.RespondSuccess(w, &httputil.ActivateInviteResponse{
		Token:   token,
		Version: version,
	})
}

// GenerateInviteCode 生成邀请码
// POST /api/v1/invite/generate
func (c *InviteController) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	code, err := c.inviteService.Generate()
	if err != nil {
		c.logger.Error("生成邀请码失败", zap.Int("error_code", err.Code), zap.Error(err))
		httputil.RespondError(w, err)
		return
	}

	c.logger.Info("生成邀请码成功", zap.String("code", code))
	httputil.RespondSuccess(w, &httputil.GenerateInviteCodeResponse{
		Code: code,
	})
}

// isValidUsername 验证用户名格式
// 规则: 3-20个字符，字母开头，仅包含字母、数字、下划线
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	pattern := `^[a-zA-Z][a-zA-Z0-9_]*$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}
