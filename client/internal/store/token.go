package store

import (
	"fmt"
)

// UpdateToken 更新 Token 和版本号
// 用于登录成功后更新本地凭证
func UpdateToken(token string, version int) error {
	// 加载现有凭证
	cred, err := LoadCredential()
	if err != nil {
		return fmt.Errorf("加载凭证失败: %w", err)
	}

	// 更新 token 和版本
	cred.Token = token
	cred.Version = version

	// 保存更新后的凭证
	if err := SaveCredential(cred); err != nil {
		return fmt.Errorf("保存凭证失败: %w", err)
	}

	// 同时更新 token.jwt 文件（向后兼容）
	if err := SaveToken(token); err != nil {
		return fmt.Errorf("保存 token 文件失败: %w", err)
	}

	return nil
}

// GetCurrentToken 获取当前 Token
func GetCurrentToken() (string, error) {
	cred, err := LoadCredential()
	if err != nil {
		return "", fmt.Errorf("加载凭证失败: %w", err)
	}

	return cred.Token, nil
}

// GetCurrentVersion 获取当前 Token 版本号
func GetCurrentVersion() (int, error) {
	cred, err := LoadCredential()
	if err != nil {
		return 0, fmt.Errorf("加载凭证失败: %w", err)
	}

	return cred.Version, nil
}
