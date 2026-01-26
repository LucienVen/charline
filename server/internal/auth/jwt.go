package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apperrors "github.com/LucienVen/charline/server/internal/errors"
)

const (
	// defaultIssuer 默认签发者
	defaultIssuer = "charline-server"
	// defaultExpiration 默认过期时间（24小时）
	defaultExpiration = 24 * time.Hour
)

// Claims JWT 声明
type Claims struct {
	Username string `json:"username"`
	Version  int    `json:"version"` // Token 版本号，用于作废旧 Token
	jwt.RegisteredClaims
}

// Manager JWT 管理器
type Manager struct {
	secretKey []byte
	issuer    string
	expire    time.Duration
}

// NewManager 创建 JWT 管理器
// secretKey: 签名密钥（建议32字节以上）
func NewManager(secretKey string) *Manager {
	return &Manager{
		secretKey: []byte(secretKey),
		issuer:    defaultIssuer,
		expire:    defaultExpiration,
	}
}

// SetExpiration 设置过期时间
func (m *Manager) SetExpiration(d time.Duration) {
	m.expire = d
}

// GenerateToken 生成 JWT Token
// username: 用户名
// version: Token 版本号（默认1）
func (m *Manager) GenerateToken(username string, version int) (string, *apperrors.BizError) {
	if username == "" {
		return "", apperrors.ErrInvalidUsername
	}

	now := time.Now()
	claims := Claims{
		Username: username,
		Version:  version,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   username,
			Audience:  []string{"charline-client"},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", apperrors.ErrSystemError
	}

	return tokenString, nil
}

// ValidateToken 验证 JWT Token
// 返回: (*Claims, 业务错误)
func (m *Manager) ValidateToken(tokenString string) (*Claims, *apperrors.BizError) {
	if tokenString == "" {
		return nil, apperrors.ErrTokenInvalid
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名算法: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, apperrors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrTokenInvalid
	}

	return claims, nil
}

// ParseTokenFromRequest 从 HTTP 请求中解析 Token
// 从 Authorization: Bearer <token> 头部提取
// 返回: (token string, 业务错误)
func ParseTokenFromRequest(authHeader string) (string, *apperrors.BizError) {
	if authHeader == "" {
		return "", apperrors.ErrTokenInvalid
	}

	// 检查前缀
	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", apperrors.ErrTokenInvalid
	}

	token := authHeader[len(prefix):]
	if token == "" {
		return "", apperrors.ErrTokenInvalid
	}

	return token, nil
}

// RefreshToken 刷新 Token（升级版本号）
// username: 用户名
// oldVersion: 旧 Token 版本号
// 返回: (newtoken string, newVersion int, 业务错误)
func (m *Manager) RefreshToken(username string, oldVersion int) (string, int, *apperrors.BizError) {
	newVersion := oldVersion + 1
	token, bizErr := m.GenerateToken(username, newVersion)
	if bizErr != nil {
		return "", 0, bizErr
	}
	return token, newVersion, nil
}
