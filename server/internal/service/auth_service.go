package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"strconv"

	"github.com/LucienVen/charline/pkg/logger"
	"github.com/LucienVen/charline/server/internal/auth"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
	"github.com/LucienVen/charline/server/internal/store"
	"go.uber.org/zap"
)

// AuthService 认证服务
type AuthService struct {
	jwtManager *auth.Manager
	userStore  *store.UserStore
	nonceStore *store.NonceStore
	logger     *logger.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(
	jwtManager *auth.Manager,
	userStore *store.UserStore,
	nonceStore *store.NonceStore,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		jwtManager: jwtManager,
		userStore:  userStore,
		nonceStore: nonceStore,
		logger:     log,
	}
}

// ChallengeResult Challenge 响应
type ChallengeResult struct {
	Nonce     string
	ExpiresIn int
}

// GetChallenge 获取挑战
func (s *AuthService) GetChallenge(tokenString string) (*ChallengeResult, *apperrors.BizError) {
	// 验证 token
	claims, bizErr := s.jwtManager.ValidateToken(tokenString)
	if bizErr != nil {
		s.logger.Warn("Token 验证失败",
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	// 从 claims 获取 username，然后查询用户 ID
	user, bizErr := s.userStore.GetByUsername(claims.Username)
	if bizErr != nil {
		s.logger.Error("查询用户失败",
			zap.String("username", claims.Username),
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	// 生成 nonce（使用 userID 的字符串形式）
	nonce, err := s.nonceStore.Generate(strconv.FormatInt(user.ID, 10))
	if err != nil {
		s.logger.Error("生成 nonce 失败",
			zap.Int64("user_id", user.ID),
			zap.String("error", err.Error()))
		return nil, apperrors.ErrSystemError.WrapError(err)
	}

	s.logger.Info("生成 challenge 成功",
		zap.String("username", claims.Username),
		zap.Int64("user_id", user.ID))

	return &ChallengeResult{
		Nonce:     nonce,
		ExpiresIn: 30, // 30 秒
	}, nil
}

// LoginResult Login 响应
type LoginResult struct {
	Token   string
	Version int
}

// Login 登录验证
func (s *AuthService) Login(nonce, signature string) (*LoginResult, *apperrors.BizError) {
	// 消费 nonce
	entry, ok := s.nonceStore.Consume(nonce)
	if !ok {
		s.logger.Warn("Nonce 消费失败",
			zap.String("nonce", nonce))
		return nil, apperrors.ErrInvalidNonce.WithDetail("reason", "nonce 不存在、已使用或已过期")
	}

	// 从 userID 字符串转换回 int64
	userID, err := strconv.ParseInt(entry.UserID, 10, 64)
	if err != nil {
		s.logger.Error("解析 userID 失败",
			zap.String("user_id", entry.UserID),
			zap.String("error", err.Error()))
		return nil, apperrors.ErrSystemError.WrapError(err)
	}

	// 从数据库获取用户公钥
	user, bizErr := s.userStore.GetByID(userID)
	if bizErr != nil {
		s.logger.Error("查询用户失败",
			zap.String("user_id", entry.UserID),
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	// 验签
	pubKeyBytes, err := base64.StdEncoding.DecodeString(user.PublicKey)
	if err != nil {
		s.logger.Error("公钥解码失败",
			zap.String("username", user.Username),
			zap.String("error", err.Error()))
		return nil, apperrors.ErrInvalidPublicKey.WrapError(err)
	}

	if !verifySignature(pubKeyBytes, []byte(nonce), signature) {
		s.logger.Warn("签名验证失败",
			zap.String("username", user.Username))
		return nil, apperrors.ErrSignatureInvalid
	}

	// 生成新 token（版本号递增）
	newVersion := user.TokenVersion + 1
	newToken, bizErr := s.jwtManager.GenerateToken(user.Username, newVersion)
	if bizErr != nil {
		s.logger.Error("生成 token 失败",
			zap.String("username", user.Username),
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	// 更新数据库中的 token 版本号
	if bizErr := s.userStore.UpdateTokenVersion(user.ID, newVersion); bizErr != nil {
		s.logger.Error("更新 token 版本失败",
			zap.String("username", user.Username),
			zap.Int("new_version", newVersion),
			zap.String("error", bizErr.Message))
		return nil, bizErr
	}

	s.logger.Info("登录成功",
		zap.String("username", user.Username),
		zap.Int("version", newVersion))

	return &LoginResult{
		Token:   newToken,
		Version: newVersion,
	}, nil
}

// verifySignature 验证 Ed25519 签名
func verifySignature(pubKey, message []byte, signatureBase64 string) bool {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubKey, message, signature)
}

// ValidateToken 验证 Token
func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, *apperrors.BizError) {
	return s.jwtManager.ValidateToken(tokenString)
}
