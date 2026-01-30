# Phase 3: WebSocket 长连接架构

本目录包含 Phase 3 WebSocket 实现的所有文档。

## 文档列表

### 📋 总览文档
- [SUMMARY.md](./SUMMARY.md) - Phase 3.1 验证总结
  - 验证状态
  - 核心功能验证
  - 关键问题修复
  - 下一步计划

### 🔧 问题排查
- [troubleshooting-guide.md](./troubleshooting-guide.md) - **问题排查全过程** ⭐
  - 从发现问题到解决问题的完整记录
  - 日志分析方法
  - 调试技巧总结
  - 经验教训提炼

### 📚 技术文档
- [websocket-knowledge.md](./websocket-knowledge.md) - WebSocket 技术知识库
  - 连接池设计
  - HTTP 升级机制
  - goroutine 管理
  - 路由行为
  - 性能考虑

### ✅ 验证文档
- [websocket-verification.md](./websocket-verification.md) - WebSocket 连接验证指南
  - 验证方法
  - 测试步骤
  - 故障排查

- [verification-result.md](./verification-result.md) - 验证结果记录
  - 问题诊断与解决
  - 测试结果
  - 修改清单

## Phase 3.1 实施状态

### ✅ 已完成
1. WebSocket 服务器基础架构
2. 连接池管理
3. 消息协议定义
4. 认证流程设计
5. 客户端 WebSocket 实现
6. 基础连接验证
7. **关键问题修复**:
   - Hijacker 接口实现
   - goroutine 启动顺序优化

### 🔄 进行中
- 完整认证流程测试
- 心跳机制验证

### 📋 待完成
- 错误处理完善
- 并发连接测试
- 性能优化

## 关键技术决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| WebSocket 库 | gorilla/websocket v1.5.1 | 成熟稳定，Go 社区标准 |
| 端口配置 | 同端口 (8080) | HTTP 和 WebSocket 共用，部署简单 |
| Session 管理 | 简化版 | 仅标记 conn.userID，不实现完整 Session |
| 认证方式 | Ed25519 签名 + Challenge-Response | 安全性高，无需传输密码 |
| 消息格式 | JSON 序列化 | 易于调试，跨语言兼容 |

## 核心问题与解决方案

### 问题 1: Hijacker 接口缺失
**错误**: `websocket: response does not implement http.Hijacker`

**原因**: 日志中间件的 responseRecorder 未实现 Hijacker 接口

**解决**: 
```go
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
        return hijacker.Hijack()
    }
    return nil, nil, http.ErrNotSupported
}
```

### 问题 2: goroutine 启动顺序
**问题**: sendChallenge 在 WriteLoop 启动前调用

**解决**: 先启动 WriteLoop，再发送消息
```go
go conn.WriteLoop()
go conn.ReadLoop(s.handler)
if err := s.sendChallenge(conn); err != nil {
    conn.Close()
    return
}
```

## 验证结果

```
✓ WebSocket 连接成功
✓ 收到服务器消息: {"type":"challenge_response","payload":{"nonce":"..."},"timestamp":...}
```

## 相关文档

- [../implementation-plan.md](../implementation-plan.md) - 总体实施计划
- [../architecture.md](../architecture.md) - 架构设计文档
- [../progress.md](../progress.md) - 项目进度追踪

## 快速开始

### 启动服务器
```bash
cd server
JWT_SECRET=test-secret-key-with-at-least-32-bytes-length go run cmd/main.go
```

### 测试连接
```bash
cd client
go run cmd/main.go connect
```

### 查看日志
```bash
# 服务器日志会输出到控制台
# 包含 WebSocket 连接、认证等信息
```
