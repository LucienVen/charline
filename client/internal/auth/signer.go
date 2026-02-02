package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

// Signer 签名器
type Signer struct {
	keyPair *KeyPair
}

// NewSigner 创建签名器
func NewSigner(kp *KeyPair) *Signer {
	return &Signer{
		keyPair: kp,
	}
}

// Sign 对 nonce 进行签名
// 返回 hex 编码的签名
func (s *Signer) Sign(nonce string) (string, error) {
	if s.keyPair == nil || s.keyPair.PrivateKey == nil {
		return "", fmt.Errorf("签名器未初始化")
	}

	// 对 nonce 进行签名
	signature := ed25519.Sign(s.keyPair.PrivateKey, []byte(nonce))

	// 返回 hex 编码的签名（与服务端保持一致）
	return hex.EncodeToString(signature), nil
}

// PublicKeyHex 返回公钥的十六进制编码
func (s *Signer) PublicKeyHex() string {
	if s.keyPair == nil || s.keyPair.PublicKey == nil {
		return ""
	}
	return hex.EncodeToString(s.keyPair.PublicKey)
}
