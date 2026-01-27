package service

import (
	apperrors "github.com/LucienVen/charline/server/internal/errors"
)

// InviteServiceInterface 邀请码服务接口
// 用于 Controller 层依赖注入，避免循环定义
type InviteServiceInterface interface {
	// Generate 生成邀请码
	Generate() (string, *apperrors.BizError)

	// Activate 激活邀请码，创建用户并返回 Token
	// 参数: code - 邀请码, username - 用户名, publicKey - Ed25519 公钥 (Base64 编码)
	// 返回: token - JWT Token, version - 凭证版本, 业务错误
	Activate(code, username, publicKey string) (token string, version int, bizErr *apperrors.BizError)
}
