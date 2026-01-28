package validator

import (
	"crypto/ed25519"
	"encoding/base64"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// 全局验证器实例
var validate *validator.Validate

// Init 初始化验证器并注册自定义验证规则
func Init() {
	validate = validator.New()

	// 注册自定义验证规则
	if err := registerCustomValidators(); err != nil {
		panic(err)
	}
}

// registerCustomValidators 注册自定义验证规则
func registerCustomValidators() error {
	// 用户名验证
	if err := validate.RegisterValidation("username", validateUsername); err != nil {
		return err
	}

	// Ed25519 公钥验证
	if err := validate.RegisterValidation("ed25519_public_key", validateEd25519PublicKey); err != nil {
		return err
	}

	return nil
}

// validateUsername 验证用户名格式
// 规则：3-20位，字母开头，仅包含字母数字下划线
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// 长度检查
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// 必须以字母开头
	if !unicode.IsLetter(rune(username[0])) {
		return false
	}

	// 只能包含字母、数字、下划线
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
	return matched
}

// validateEd25519PublicKey 验证 Ed25519 公钥
// 规则：Base64 编码，解码后长度为 32 字节
func validateEd25519PublicKey(fl validator.FieldLevel) bool {
	keyStr := fl.Field().String()

	// Base64 解码
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return false
	}

	// Ed25519 公钥长度必须为 32 字节
	if len(keyBytes) != ed25519.PublicKeySize {
		return false
	}

	return true
}

// Validate 验证结构体
func Validate(s interface{}) error {
	if validate == nil {
		Init()
	}
	return validate.Struct(s)
}

// GetValidator 获取全局验证器实例
func GetValidator() *validator.Validate {
	if validate == nil {
		Init()
	}
	return validate
}
