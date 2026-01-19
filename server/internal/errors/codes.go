package errors

// ============================================
// 统一错误码定义
// 格式: ERR_模块_具体错误
// ============================================

import "errors"

// 通用错误
var (
	// ErrInternal 服务器内部错误
	ErrInternal = errors.New("INTERNAL_ERROR")
	// ErrInvalidRequest 无效请求
	ErrInvalidRequest = errors.New("INVALID_REQUEST")
)

// 邀请码相关错误 (ERR_INVITE_xxx)
var (
	// ErrInviteNotFound 邀请码不存在
	ErrInviteNotFound = errors.New("ERR_INVITE_NOT_FOUND")
	// ErrInviteAlreadyUsed 邀请码已被使用
	ErrInviteAlreadyUsed = errors.New("ERR_INVITE_ALREADY_USED")
	// ErrInviteInvalid 邀请码格式无效
	ErrInviteInvalid = errors.New("ERR_INVITE_INVALID")
	// ErrInviteGenerateFailed 生成邀请码失败
	ErrInviteGenerateFailed = errors.New("ERR_INVITE_GENERATE_FAILED")
)

// JWT 认证相关错误 (ERR_AUTH_xxx)
var (
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = errors.New("ERR_AUTH_TOKEN_INVALID")
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = errors.New("ERR_AUTH_TOKEN_EXPIRED")
	// ErrTokenMalformed Token 格式错误
	ErrTokenMalformed = errors.New("ERR_AUTH_TOKEN_MALFORMED")
)

// 用户相关错误 (ERR_USER_xxx)
var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("ERR_USER_NOT_FOUND")
	// ErrUserExists 用户已存在
	ErrUserExists = errors.New("ERR_USER_EXISTS")
	// ErrUsernameInvalid 用户名格式无效
	ErrUsernameInvalid = errors.New("ERR_USER_USERNAME_INVALID")
)
