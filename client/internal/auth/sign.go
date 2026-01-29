package auth

import (
	"crypto/ed25519"
	"encoding/base64"
)

// Sign 使用私钥对消息进行签名
// 返回 Base64 编码的签名
func (kp *KeyPair) Sign(message []byte) string {
	signature := ed25519.Sign(kp.PrivateKey, message)
	return base64.StdEncoding.EncodeToString(signature)
}
