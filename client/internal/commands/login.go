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

// LoginConfig Login 命令配置
type LoginConfig struct {
	ServerURL string // 服务端 URL
}

// LoginResult Login 结果
type LoginResult struct {
	Token   string // 新的 JWT token
	Version int    // 新的凭证版本
}

// ChallengeResponse Challenge 响应
type ChallengeResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
	Data    struct {
		Nonce     string `json:"nonce"`      // 随机挑战值
		ExpiresIn int    `json:"expires_in"` // 过期时间（秒）
	} `json:"data"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Nonce     string `json:"nonce"`     // 挑战值
	Signature string `json:"signature"` // 签名
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
	Data    struct {
		Token   string `json:"token"`   // 新的 JWT token
		Version int    `json:"version"` // 新的凭证版本
	} `json:"data"`
}

// Login 执行登录流程
func Login(cfg *LoginConfig) (*LoginResult, error) {
	// 1. 检查是否有凭证
	if !store.HasCredential() {
		return nil, fmt.Errorf("未找到凭证，请先执行 join 命令")
	}

	// 2. 加载当前 token
	token, err := store.GetCurrentToken()
	if err != nil {
		return nil, fmt.Errorf("加载 token 失败: %w", err)
	}

	// 3. 获取 challenge
	challenge, err := doGetChallenge(cfg.ServerURL, token)
	if err != nil {
		return nil, fmt.Errorf("获取 challenge 失败: %w", err)
	}

	// 4. 加载密钥对
	kp, err := auth.Load()
	if err != nil {
		return nil, fmt.Errorf("加载密钥对失败: %w", err)
	}

	// 5. 对 nonce 进行签名
	signature := kp.Sign([]byte(challenge.Data.Nonce))

	// 6. 发送登录请求
	loginResp, err := doLoginRequest(cfg.ServerURL, challenge.Data.Nonce, signature)
	if err != nil {
		return nil, fmt.Errorf("登录请求失败: %w", err)
	}

	// 7. 更新本地 token 和版本
	if err := store.UpdateToken(loginResp.Data.Token, loginResp.Data.Version); err != nil {
		return nil, fmt.Errorf("更新 token 失败: %w", err)
	}

	return &LoginResult{
		Token:   loginResp.Data.Token,
		Version: loginResp.Data.Version,
	}, nil
}

// doGetChallenge 获取登录挑战
func doGetChallenge(serverURL, token string) (*ChallengeResponse, error) {
	// 构造 HTTP 请求
	url := fmt.Sprintf("%s/api/v1/auth/challenge", serverURL)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置 Authorization 头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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
	var resp ChallengeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应码
	if resp.Code != 0 {
		return nil, fmt.Errorf("获取 challenge 失败: %s (code: %d)", resp.Message, resp.Code)
	}

	return &resp, nil
}

// doLoginRequest 发送登录请求
func doLoginRequest(serverURL, nonce, signature string) (*LoginResponse, error) {
	// 构造请求体
	req := LoginRequest{
		Nonce:     nonce,
		Signature: signature,
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构造 HTTP 请求
	url := fmt.Sprintf("%s/api/v1/auth/login", serverURL)
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
	var resp LoginResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应码
	if resp.Code != 0 {
		return nil, fmt.Errorf("登录失败: %s (code: %d)", resp.Message, resp.Code)
	}

	return &resp, nil
}
