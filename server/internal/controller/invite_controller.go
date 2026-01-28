package controller

import (
	"net/http"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/httputil"
	"github.com/LucienVen/charline/server/internal/service"
	"github.com/LucienVen/charline/server/internal/validator"
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

	// 统一验证
	if err := validator.Validate(req); err != nil {
		validationErrors := validator.ParseError(err)
		c.logger.Warn("请求参数验证失败",
			zap.Any("validation_errors", validationErrors))
		httputil.RespondWithError(w, http.StatusBadRequest,
			errors.ErrInvalidParam.WithDetails(map[string]interface{}{
				"validation_errors": validationErrors,
			}))
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
