package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// Credential 用户凭证
type Credential struct {
	Token    string `json:"token"`     // JWT token
	Username string `json:"username"`  // 用户名
	Version  int    `json:"version"`   // 凭证版本
}

// SaveToken 保存 JWT token 到 ~/.charline/token.jwt
func SaveToken(token string) error {
	tokenPath, err := TokenPath()
	if err != nil {
		return fmt.Errorf("获取 token 路径失败: %w", err)
	}

	// 写入文件，权限 600
	if err := os.WriteFile(tokenPath, []byte(token), PrivateFilePerm); err != nil {
		return fmt.Errorf("保存 token 失败: %w", err)
	}

	return nil
}

// LoadToken 从 ~/.charline/token.jwt 加载 token
func LoadToken() (string, error) {
	tokenPath, err := TokenPath()
	if err != nil {
		return "", fmt.Errorf("获取 token 路径失败: %w", err)
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("读取 token 失败: %w", err)
	}

	return string(data), nil
}

// SaveCredential 保存完整凭证信息
func SaveCredential(cred *Credential) error {
	credPath, err := CredentialPath()
	if err != nil {
		return fmt.Errorf("获取凭证路径失败: %w", err)
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化凭证失败: %w", err)
	}

	// 写入文件，权限 600
	if err := os.WriteFile(credPath, data, PrivateFilePerm); err != nil {
		return fmt.Errorf("保存凭证失败: %w", err)
	}

	return nil
}

// LoadCredential 加载完整凭证信息
func LoadCredential() (*Credential, error) {
	credPath, err := CredentialPath()
	if err != nil {
		return nil, fmt.Errorf("获取凭证路径失败: %w", err)
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("读取凭证失败: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("解析凭证失败: %w", err)
	}

	return &cred, nil
}

// HasCredential 检查是否已有凭证（判断是否需要 join）
func HasCredential() bool {
	credPath, err := CredentialPath()
	if err != nil {
		return false
	}

	// 检查文件是否存在
	_, err = os.Stat(credPath)
	return err == nil
}
