package errors

import (
	"fmt"
	"strings"
)

// ============================================
// 错误码定义（按语义分段）
// ============================================
//
// 0     - 成功
// 1xxx  - 参数/请求错误
// 2xxx  - 资源相关错误
// 3xxx  - 认证/授权错误
// 4xxx  - 业务逻辑错误
// 5xxx  - 系统内部错误
//

const (
	// 成功
	ErrCodeSuccess = 0

	// 1xxx - 参数/请求错误
	ErrCodeInvalidParam    = 1001 // 无效参数
	ErrCodeInvalidJSON     = 1002 // JSON 格式错误
	ErrCodeInvalidUsername = 1003 // 用户名格式无效
	ErrCodeMissingField    = 1004 // 缺少必填字段

	// 2xxx - 资源相关错误
	ErrCodeInviteNotFound = 2001 // 邀请码不存在
	ErrCodeInviteUsed     = 2002 // 邀请码已使用
	ErrCodeInviteInvalid  = 2003 // 邀请码格式无效
	ErrCodeUserNotFound   = 2004 // 用户不存在
	ErrCodeUserExists     = 2005 // 用户已存在

	// 3xxx - 认证/授权错误
	ErrCodeUnauthorized = 3001 // 未授权
	ErrCodeTokenExpired = 3002 // Token 已过期
	ErrCodeTokenInvalid = 3003 // Token 无效

	// 5xxx - 系统内部错误
	ErrCodeSystemError = 5000 // 系统错误
)

// ============================================
// 错误码映射表
// ============================================

var codeMessages = map[int]string{
	ErrCodeSuccess:           "成功",
	ErrCodeInvalidParam:      "参数错误",
	ErrCodeInvalidJSON:       "JSON 格式错误",
	ErrCodeInvalidUsername:   "用户名格式无效",
	ErrCodeMissingField:      "缺少必填字段",
	ErrCodeInviteNotFound:    "邀请码不存在",
	ErrCodeInviteUsed:        "邀请码已使用",
	ErrCodeInviteInvalid:     "邀请码格式无效",
	ErrCodeUserNotFound:      "用户不存在",
	ErrCodeUserExists:        "用户已存在",
	ErrCodeUnauthorized:      "未授权",
	ErrCodeTokenExpired:      "Token 已过期",
	ErrCodeTokenInvalid:      "Token 无效",
	ErrCodeSystemError:       "系统错误",
}

// ============================================
// 业务错误结构
// ============================================

// BizError 业务错误
type BizError struct {
	Code    int
	Message string
	Details map[string]interface{} // 详细信息（通过 RespondError 放入 data）
	cause   error                 // 原始错误，用于日志，不暴露给客户端
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Message
}

// GetCode 获取错误码
func (e *BizError) GetCode() int {
	return e.Code
}

// GetMsg 获取错误消息
func (e *BizError) GetMsg() string {
	return e.Message
}

// WithDetails 返回带详细信息的错误副本
func (e *BizError) WithDetails(details map[string]interface{}) *BizError {
	return &BizError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		cause:   e.cause,
	}
}

// WithDetail 返回带单个详细字段的错误副本（链式调用）
func (e *BizError) WithDetail(key string, value interface{}) *BizError {
	if e.Details == nil {
		return e.WithDetails(map[string]interface{}{key: value})
	}
	// 复制原有 details
	newDetails := make(map[string]interface{}, len(e.Details)+1)
	for k, v := range e.Details {
		newDetails[k] = v
	}
	newDetails[key] = value
	return &BizError{
		Code:    e.Code,
		Message: e.Message,
		Details: newDetails,
		cause:   e.cause,
	}
}

// WrapError 包装底层错误（用于日志）
func (e *BizError) WrapError(cause error) *BizError {
	return &BizError{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
		cause:   cause,
	}
}

// GetCause 获取底层错误
func (e *BizError) GetCause() error {
	return e.cause
}

// GetDetailedMsg 获取包含详细信息的错误消息（用于日志）
func (e *BizError) GetDetailedMsg() string {
	if e.Details == nil || len(e.Details) == 0 {
		return e.Message
	}
	var detailStrs []string
	for k, v := range e.Details {
		detailStrs = append(detailStrs, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s (%s)", e.Message, strings.Join(detailStrs, ", "))
}

// NewBizError 创建业务错误
func NewBizError(code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// ============================================
// 预定义错误实例
// ============================================

// 参数错误
var (
	ErrInvalidParam    = &BizError{Code: ErrCodeInvalidParam, Message: "参数错误"}
	ErrInvalidJSON     = &BizError{Code: ErrCodeInvalidJSON, Message: "JSON 格式错误"}
	ErrInvalidUsername = &BizError{Code: ErrCodeInvalidUsername, Message: "用户名格式无效"}
	ErrMissingField    = &BizError{Code: ErrCodeMissingField, Message: "缺少必填字段"}
)

// 资源错误
var (
	ErrInviteNotFound = &BizError{Code: ErrCodeInviteNotFound, Message: "邀请码不存在"}
	ErrInviteUsed     = &BizError{Code: ErrCodeInviteUsed, Message: "邀请码已使用"}
	ErrInviteInvalid  = &BizError{Code: ErrCodeInviteInvalid, Message: "邀请码格式无效"}
	ErrUserNotFound   = &BizError{Code: ErrCodeUserNotFound, Message: "用户不存在"}
	ErrUserExists     = &BizError{Code: ErrCodeUserExists, Message: "用户已存在"}
)

// 认证错误
var (
	ErrUnauthorized = &BizError{Code: ErrCodeUnauthorized, Message: "未授权"}
	ErrTokenExpired = &BizError{Code: ErrCodeTokenExpired, Message: "Token 已过期"}
	ErrTokenInvalid = &BizError{Code: ErrCodeTokenInvalid, Message: "Token 无效"}
)

// 系统错误
var (
	ErrSystemError = &BizError{Code: ErrCodeSystemError, Message: "系统错误"}
)

// ============================================
// 工具函数
// ============================================

// GetMsg 获取错误码对应的消息
func GetMsg(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
