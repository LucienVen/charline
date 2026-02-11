package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LucienVen/charline/client/internal/auth"
	"github.com/LucienVen/charline/client/internal/websocket"
)

// ConnectConfig Connect 命令配置
type ConnectConfig struct {
	ServerURL string // WebSocket 服务器 URL (ws://host:port/ws)
}

// ConnectResult Connect 命令结果
type ConnectResult struct {
	UserID int64 // 用户 ID
}

// Connect 连接到服务器并保持长连接
func Connect(cfg *ConnectConfig) (*ConnectResult, error) {
	// 1. 加载密钥对
	keyPair, err := auth.Load()
	if err != nil {
		return nil, fmt.Errorf("加载密钥对失败: %w", err)
	}

	// 2. 创建签名器
	signer := auth.NewSigner(keyPair)

	// 3. 创建 WebSocket 客户端
	client := websocket.NewClient(cfg.ServerURL, signer)

	// 4. 启用自动重连（在连接前配置）
	client.EnableAutoReconnect(&websocket.ReconnectConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	})

	// 5. 设置断线回调
	client.SetDisconnectCallback(func() {
		fmt.Println("\n⚠️  连接已断开，正在尝试重连...")
	})

	// 6. 设置重连回调
	client.SetReconnectCallback(func(success bool) {
		if success {
			fmt.Println("✓ 重连成功")
		} else {
			fmt.Println("✗ 重连失败，请手动重新连接")
		}
	})

	// 7. 连接到服务器
	fmt.Println("正在连接到服务器...")
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	fmt.Println("✓ 连接成功")

	// 8. 等待一小段时间让 readLoop 启动
	time.Sleep(100 * time.Millisecond)

	// 9. 执行认证
	fmt.Println("正在进行身份认证...")
	if err := client.Authenticate(); err != nil {
		client.Close()
		return nil, fmt.Errorf("认证失败: %w", err)
	}
	fmt.Printf("✓ 认证成功 (用户 ID: %d)\n", client.GetUserID())

	// 10. 保持连接，等待中断信号
	fmt.Println("\n已连接到服务器，按 Ctrl+C 断开连接")
	fmt.Println("自动重连已启用 (最多重试 5 次)")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 11. 关闭连接
	fmt.Println("\n正在断开连接...")
	client.Close()
	fmt.Println("✓ 已断开连接")

	return &ConnectResult{
		UserID: client.GetUserID(),
	}, nil
}
