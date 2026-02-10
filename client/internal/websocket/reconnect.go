package websocket

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/LucienVen/charline/client/internal/session"
)

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	MaxRetries int           // 最大重试次数（默认 5）
	BaseDelay  time.Duration // 基础延迟（默认 1s）
	MaxDelay   time.Duration // 最大延迟（默认 30s）
}

// DefaultReconnectConfig 默认重连配置
var DefaultReconnectConfig = ReconnectConfig{
	MaxRetries: 5,
	BaseDelay:  1 * time.Second,
	MaxDelay:   30 * time.Second,
}

// ReconnectState 重连状态
type ReconnectState int

const (
	ReconnectStateIdle       ReconnectState = iota // 空闲
	ReconnectStateConnecting                       // 正在连接
	ReconnectStateResuming                         // 正在恢复
	ReconnectStateFailed                           // 失败
)

// String 返回状态字符串
func (s ReconnectState) String() string {
	switch s {
	case ReconnectStateIdle:
		return "idle"
	case ReconnectStateConnecting:
		return "connecting"
	case ReconnectStateResuming:
		return "resuming"
	case ReconnectStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ReconnectCallback 重连回调
type ReconnectCallback func(success bool, attempt int, err error)

// ReconnectManager 重连管理器
type ReconnectManager struct {
	client       *Client              // WebSocket 客户端
	session      *session.State       // Session 状态
	config       ReconnectConfig      // 重连配置
	state        ReconnectState       // 当前状态
	attempt      int                  // 当前重试次数
	onReconnect  ReconnectCallback    // 重连回调
	stopChan     chan struct{}        // 停止信号
	mu           sync.Mutex           // 互斥锁
}

// NewReconnectManager 创建重连管理器
func NewReconnectManager(client *Client, sess *session.State, config *ReconnectConfig) *ReconnectManager {
	cfg := DefaultReconnectConfig
	if config != nil {
		if config.MaxRetries > 0 {
			cfg.MaxRetries = config.MaxRetries
		}
		if config.BaseDelay > 0 {
			cfg.BaseDelay = config.BaseDelay
		}
		if config.MaxDelay > 0 {
			cfg.MaxDelay = config.MaxDelay
		}
	}

	return &ReconnectManager{
		client:   client,
		session:  sess,
		config:   cfg,
		state:    ReconnectStateIdle,
		attempt:  0,
		stopChan: make(chan struct{}),
	}
}

// SetCallback 设置重连回调
func (r *ReconnectManager) SetCallback(callback ReconnectCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onReconnect = callback
}

// Start 开始重连
// 返回是否成功恢复连接
func (r *ReconnectManager) Start() bool {
	r.mu.Lock()
	if r.state == ReconnectStateConnecting || r.state == ReconnectStateResuming {
		r.mu.Unlock()
		return false // 已经在重连中
	}
	r.state = ReconnectStateConnecting
	r.attempt = 0
	r.mu.Unlock()

	for r.attempt < r.config.MaxRetries {
		select {
		case <-r.stopChan:
			r.setState(ReconnectStateIdle)
			return false
		default:
		}

		r.attempt++

		// 计算延迟（指数退避）
		delay := r.calculateDelay(r.attempt)
		fmt.Printf("[Reconnect] 尝试 %d/%d，等待 %v...\n", r.attempt, r.config.MaxRetries, delay)

		// 等待延迟
		select {
		case <-time.After(delay):
		case <-r.stopChan:
			r.setState(ReconnectStateIdle)
			return false
		}

		// 尝试恢复（如果有有效的 Resume Token）
		if r.session.IsResumeTokenValid() {
			r.setState(ReconnectStateResuming)
			if r.tryResume() {
				r.notifyCallback(true, r.attempt, nil)
				r.setState(ReconnectStateIdle)
				return true
			}
		}

		// Resume 失败或无效，尝试重新连接
		r.setState(ReconnectStateConnecting)
		if r.tryReconnect() {
			r.notifyCallback(true, r.attempt, nil)
			r.setState(ReconnectStateIdle)
			return true
		}
	}

	// 所有重试都失败
	r.setState(ReconnectStateFailed)
	r.notifyCallback(false, r.attempt, fmt.Errorf("max retries exceeded"))
	return false
}

// Stop 停止重连
func (r *ReconnectManager) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.stopChan:
		// 已经关闭
	default:
		close(r.stopChan)
	}
}

// GetState 获取当前状态
func (r *ReconnectManager) GetState() ReconnectState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

// GetAttempt 获取当前重试次数
func (r *ReconnectManager) GetAttempt() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.attempt
}

// tryResume 尝试使用 Resume Token 恢复
func (r *ReconnectManager) tryResume() bool {
	token, valid := r.session.GetResumeToken()
	if !valid {
		return false
	}

	fmt.Printf("[Reconnect] 尝试使用 Resume Token 恢复...\n")

	// 1. 建立新连接
	if err := r.client.Connect(); err != nil {
		fmt.Printf("[Reconnect] 连接失败: %v\n", err)
		return false
	}

	// 2. 发送 Resume 请求
	if err := r.client.Resume(token); err != nil {
		fmt.Printf("[Reconnect] Resume 失败: %v\n", err)
		r.client.Close()
		// 清除无效的 Resume Token
		r.session.ClearResumeToken()
		return false
	}

	fmt.Printf("[Reconnect] Resume 成功\n")
	return true
}

// tryReconnect 尝试重新连接并认证
func (r *ReconnectManager) tryReconnect() bool {
	fmt.Printf("[Reconnect] 尝试重新连接...\n")

	// 1. 建立新连接
	if err := r.client.Connect(); err != nil {
		fmt.Printf("[Reconnect] 连接失败: %v\n", err)
		return false
	}

	// 2. 重新认证
	if err := r.client.Authenticate(); err != nil {
		fmt.Printf("[Reconnect] 认证失败: %v\n", err)
		r.client.Close()
		return false
	}

	fmt.Printf("[Reconnect] 重新连接成功\n")
	return true
}

// calculateDelay 计算延迟（指数退避 + 抖动）
func (r *ReconnectManager) calculateDelay(attempt int) time.Duration {
	// 指数退避: baseDelay * 2^(attempt-1)
	delay := float64(r.config.BaseDelay) * math.Pow(2, float64(attempt-1))

	// 限制最大延迟
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// 添加 10% 抖动
	jitter := delay * 0.1
	delay = delay + (jitter * (0.5 - float64(time.Now().UnixNano()%100)/100))

	return time.Duration(delay)
}

// setState 设置状态
func (r *ReconnectManager) setState(state ReconnectState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = state
}

// notifyCallback 通知回调
func (r *ReconnectManager) notifyCallback(success bool, attempt int, err error) {
	r.mu.Lock()
	callback := r.onReconnect
	r.mu.Unlock()

	if callback != nil {
		callback(success, attempt, err)
	}
}

// Reset 重置重连管理器（用于新的重连周期）
func (r *ReconnectManager) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = ReconnectStateIdle
	r.attempt = 0

	// 重新创建 stopChan
	select {
	case <-r.stopChan:
		r.stopChan = make(chan struct{})
	default:
	}
}
