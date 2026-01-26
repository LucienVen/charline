package service

import (
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/store"
	"go.uber.org/zap"
)

// InviteService 邀请码服务
type InviteService struct {
	inviteStore *store.InviteStore
	jwtManager  *auth.Manager
	logger      *logger.Logger
}

// NewInviteService 创建邀请码服务实例
func NewInviteService(
	inviteStore *store.InviteStore,
	jwtManager *auth.Manager,
	log *logger.Logger,
) *InviteService {
	return &InviteService{
		inviteStore: inviteStore,
		jwtManager:  jwtManager,
		logger:      log,
	}
}

// Generate 生成邀请码
func (s *InviteService) Generate() (string, *apperrors.BizError) {
	code, err := s.inviteStore.Generate()
	if err != nil {
		s.logger.Error("生成邀请码失败",
			zap.String("error", err.Error()))
		return "", apperrors.ErrSystemError
	}

	s.logger.Info("生成邀请码成功",
		zap.String("code", code))
	return code, nil
}

// Activate 激活邀请码，返回 Token
func (s *InviteService) Activate(code, username string) (token string, version int, bizErr *apperrors.BizError) {
	// 激活邀请码
	if bizErr = s.inviteStore.Activate(code, username); bizErr != nil {
		s.logger.Error("激活邀请码失败",
			zap.String("code", code),
			zap.String("username", username),
			zap.Int("code", bizErr.Code))
		return "", 0, bizErr
	}

	// 生成 JWT Token
	tokenStr, err := s.jwtManager.GenerateToken(username, 1)
	if err != nil {
		s.logger.Error("生成 Token 失败",
			zap.String("username", username),
			zap.String("error", err.Error()))
		return "", 0, apperrors.ErrSystemError
	}

	s.logger.Info("激活邀请码成功",
		zap.String("code", code),
		zap.String("username", username))
	return tokenStr, 1, nil
}
