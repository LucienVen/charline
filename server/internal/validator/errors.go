package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Error 验证错误
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// Error 实现 error 接口
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ParseError 解析验证错误，返回结构化的错误列表
func ParseError(err error) []Error {
	var errors []Error

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			// 生成友好的错误消息
			message := getErrorMessage(field, tag)

			errors = append(errors, Error{
				Field:   field,
				Message: message,
				Tag:     tag,
			})
		}
	}

	return errors
}

// getErrorMessage 根据字段和标签生成错误消息
func getErrorMessage(field, tag string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s 不能为空", field)
	case "username":
		return "用户名格式无效（3-20位，字母开头，含字母数字下划线）"
	case "ed25519_public_key":
		return "公钥格式无效（必须是有效的 Ed25519 Base64 公钥）"
	case "email":
		return "邮箱格式不正确"
	case "min":
		return fmt.Sprintf("%s 长度不能少于要求", field)
	case "max":
		return fmt.Sprintf("%s 长度不能超过要求", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}

// FormatAsMap 将验证错误转换为 map 格式
func FormatAsMap(errors []Error) map[string]string {
	result := make(map[string]string)
	for _, err := range errors {
		result[err.Field] = err.Message
	}
	return result
}

// FormatAsSlice 将验证错误转换为切片格式
func FormatAsSlice(errors []Error) []string {
	var result []string
	for _, err := range errors {
		result = append(result, err.Error())
	}
	return result
}

// ValidationErrorString 将错误列表格式化为字符串
func ValidationErrorString(errors []Error) string {
	if len(errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}
