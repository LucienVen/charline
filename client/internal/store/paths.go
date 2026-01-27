// Package store 提供客户端本地存储功能
package store

import (
	"os"
	"path/filepath"
)

const (
	// CharlineDirName 数据目录名称
	CharlineDirName = ".charline"

	// PrivateKeyFile 私钥文件名
	PrivateKeyFile = "id_ed25519"

	// PublicKeyFile 公钥文件名
	PublicKeyFile = "id_ed25519.pub"

	// TokenFile JWT token 文件名
	TokenFile = "token.jwt"

	// CredentialFile 凭证信息文件名
	CredentialFile = "credential.json"

	// DirPerm 目录权限 (700)
	DirPerm = 0700

	// PrivateFilePerm 私钥文件权限 (600)
	PrivateFilePerm = 0600

	// PublicFilePerm 公钥文件权限 (644)
	PublicFilePerm = 0644
)

// GetCharlineDir 返回 ~/.charline/ 目录路径
// 不会创建目录，仅返回路径
func GetCharlineDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, CharlineDirName), nil
}

// EnsurePrivateDir 确保 ~/.charline/ 目录存在且权限正确 (700)
// 如果目录不存在则创建
func EnsurePrivateDir() (string, error) {
	dir, err := GetCharlineDir()
	if err != nil {
		return "", err
	}

	// 检查目录是否存在
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// 创建目录，权限 700
		if err := os.MkdirAll(dir, DirPerm); err != nil {
			return "", err
		}
		return dir, nil
	}
	if err != nil {
		return "", err
	}

	// 目录存在，检查是否为目录
	if !info.IsDir() {
		return "", os.ErrExist
	}

	return dir, nil
}

// KeyPath 返回私钥文件路径
func KeyPath() (string, error) {
	dir, err := GetCharlineDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PrivateKeyFile), nil
}

// KeyPubPath 返回公钥文件路径
func KeyPubPath() (string, error) {
	dir, err := GetCharlineDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, PublicKeyFile), nil
}

// TokenPath 返回 token 文件路径
func TokenPath() (string, error) {
	dir, err := GetCharlineDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, TokenFile), nil
}

// CredentialPath 返回凭证文件路径
func CredentialPath() (string, error) {
	dir, err := GetCharlineDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CredentialFile), nil
}
