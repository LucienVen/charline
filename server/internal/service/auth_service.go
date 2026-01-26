package service

import (
	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"go.uber.org/zap"
)

// AuthService 认证服务
type AuthService struct {
	jwtManager *auth.Manager
	logger     *logger.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	jwtManager *auth.Manager,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		jwtManager: jwtManager,
		logger:     log,
	}
}

// ValidateToken 验证 Token
func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, *apperrors.BizError) {
	claims, bizErr := s.jwtManager.ValidateToken(tokenString)
	if bizErr != nil {
		s.logger.Warn("Token 验证失败",
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	s.logger.Info("Token 验证成功",
		zap.String("username", claims.Username))
	return claims, nil
}
