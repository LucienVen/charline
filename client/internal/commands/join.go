// Package commands 提供客户端命令实现
package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LucienVen/charline/client/internal/auth"
	"github.com/LucienVen/charline/client/internal/store"
)

// JoinConfig Join 命令配置
type JoinConfig struct {
	ServerURL string // 服务端 URL
	Code      string // 邀请码
	Username  string // 用户名
}

// JoinResult Join 结果
type JoinResult struct {
	Token   string // JWT token
	Version int    // 凭证版本
}

// ActivateRequest 激活请求
type ActivateRequest struct {
	Code      string `json:"code"`       // 邀请码
	Username  string `json:"username"`   // 用户名
	PublicKey string `json:"public_key"` // Base64 编码的公钥
}

// ActivateResponse 激活响应
type ActivateResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
	Data    struct {
		Token   string `json:"token"`   // JWT token
		Version int    `json:"version"` // 凭证版本
	} `json:"data"`
}

// Join 执行 join 流程
func Join(cfg *JoinConfig) (*JoinResult, error) {
	// 1. 检查是否已有凭证
	if store.HasCredential() {
		return nil, fmt.Errorf("已有凭证，无需重复 join")
	}

	// 2. 确保 ~/.charline/ 目录存在
	if _, err := store.EnsurePrivateDir(); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 3. 生成 Ed25519 密钥对
	kp, err := auth.Generate()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 4. 保存密钥对
	if err := kp.Save(); err != nil {
		return nil, fmt.Errorf("保存密钥对失败: %w", err)
	}

	// 5. 构造激活请求
	req := ActivateRequest{
		Code:      cfg.Code,
		Username:  cfg.Username,
		PublicKey: kp.PubKeyBase64(),
	}

	// 6. 调用 server API
	resp, err := doActivateRequest(cfg.ServerURL, req)
	if err != nil {
		return nil, fmt.Errorf("激活请求失败: %w", err)
	}

	// 7. 保存 token
	if err := store.SaveToken(resp.Data.Token); err != nil {
		return nil, fmt.Errorf("保存 token 失败: %w", err)
	}

	// 8. 保存完整凭证
	cred := &store.Credential{
		Token:    resp.Data.Token,
		Username: cfg.Username,
		Version:  resp.Data.Version,
	}
	if err := store.SaveCredential(cred); err != nil {
		return nil, fmt.Errorf("保存凭证失败: %w", err)
	}

	return &JoinResult{
		Token:   resp.Data.Token,
		Version: resp.Data.Version,
	}, nil
}

// doActivateRequest 发送激活请求
func doActivateRequest(serverURL string, req ActivateRequest) (*ActivateResponse, error) {
	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构造 HTTP 请求
	url := fmt.Sprintf("%s/api/v1/invite/activate", serverURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var resp ActivateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应码
	if resp.Code != 0 {
		return nil, fmt.Errorf("激活失败: %s (code: %d)", resp.Message, resp.Code)
	}

	return &resp, nil
}
