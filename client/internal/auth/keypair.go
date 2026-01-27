// Package auth 提供客户端认证相关功能
package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/LucienVen/charline/client/internal/store"
)

const (
	// PrivateKeyPEMType PEM 类型标识
	PrivateKeyPEMType = "PRIVATE KEY"
	// PublicKeyPEMType PEM 类型标识
	PublicKeyPEMType = "PUBLIC KEY"
)

// KeyPair Ed25519 密钥对
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// Generate 生成新的 Ed25519 密钥对
func Generate() (*KeyPair, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	return &KeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

// Load 从 ~/.charline/id_ed25519 加载密钥对
func Load() (*KeyPair, error) {
	// 获取私钥路径
	privKeyPath, err := store.KeyPath()
	if err != nil {
		return nil, fmt.Errorf("获取私钥路径失败: %w", err)
	}

	// 读取私钥文件
	privKeyData, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 解析 PEM 格式
	block, _ := pem.Decode(privKeyData)
	if block == nil || block.Type != PrivateKeyPEMType {
		return nil, fmt.Errorf("无效的私钥格式")
	}

	// Ed25519 私钥是 64 字节
	if len(block.Bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("私钥长度错误: 期望 %d 字节，实际 %d 字节",
			ed25519.PrivateKeySize, len(block.Bytes))
	}

	privKey := ed25519.PrivateKey(block.Bytes)
	pubKey := privKey.Public().(ed25519.PublicKey)

	return &KeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

// Save 保存密钥对到 ~/.charline/
// 私钥权限: 600
// 公钥权限: 644
func (kp *KeyPair) Save() error {
	// 确保目录存在
	if _, err := store.EnsurePrivateDir(); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 保存私钥
	if err := kp.savePrivateKey(); err != nil {
		return err
	}

	// 保存公钥
	if err := kp.savePublicKey(); err != nil {
		return err
	}

	return nil
}

// savePrivateKey 保存私钥到文件
func (kp *KeyPair) savePrivateKey() error {
	privKeyPath, err := store.KeyPath()
	if err != nil {
		return fmt.Errorf("获取私钥路径失败: %w", err)
	}

	// 编码为 PEM 格式
	block := &pem.Block{
		Type:  PrivateKeyPEMType,
		Bytes: kp.PrivateKey,
	}
	privKeyPEM := pem.EncodeToMemory(block)

	// 写入文件，权限 600
	if err := os.WriteFile(privKeyPath, privKeyPEM, store.PrivateFilePerm); err != nil {
		return fmt.Errorf("保存私钥失败: %w", err)
	}

	return nil
}

// savePublicKey 保存公钥到文件
func (kp *KeyPair) savePublicKey() error {
	pubKeyPath, err := store.KeyPubPath()
	if err != nil {
		return fmt.Errorf("获取公钥路径失败: %w", err)
	}

	// 编码为 PEM 格式
	block := &pem.Block{
		Type:  PublicKeyPEMType,
		Bytes: kp.PublicKey,
	}
	pubKeyPEM := pem.EncodeToMemory(block)

	// 写入文件，权限 644
	if err := os.WriteFile(pubKeyPath, pubKeyPEM, store.PublicFilePerm); err != nil {
		return fmt.Errorf("保存公钥失败: %w", err)
	}

	return nil
}

// PubKeyBytes 返回公钥字节数组
func (kp *KeyPair) PubKeyBytes() []byte {
	return kp.PublicKey
}

// PubKeyBase64 返回 Base64 编码的公钥（用于发送给 server）
func (kp *KeyPair) PubKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKey)
}
