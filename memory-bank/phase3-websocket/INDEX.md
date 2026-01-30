# Phase 3 WebSocket 文档索引

快速导航到相关文档。

## 📖 按用途分类

### 🎯 快速开始
- **想快速了解验证结果？** → [SUMMARY.md](./SUMMARY.md)
- **想了解如何测试？** → [websocket-verification.md](./websocket-verification.md)

### 🔧 问题排查
- **遇到 WebSocket 连接问题？** → [troubleshooting-guide.md](./troubleshooting-guide.md)
- **想了解常见问题？** → [README.md](./README.md#核心问题与解决方案)

### 📚 技术学习
- **想深入理解 WebSocket？** → [websocket-knowledge.md](./websocket-knowledge.md)
- **想了解实现细节？** → [verification-result.md](./verification-result.md)

### 📋 项目管理
- **想查看实施状态？** → [README.md](./README.md#phase-31-实施状态)
- **想了解技术决策？** → [README.md](./README.md#关键技术决策)

## 📄 按文档类型分类

### 总览文档
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [README.md](./README.md) | 目录总览，包含状态、决策、问题总结 | 所有人 |
| [SUMMARY.md](./SUMMARY.md) | 简洁的验证总结 | 快速了解 |
| [INDEX.md](./INDEX.md) | 本文档，快速导航 | 查找文档 |

### 技术文档
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [websocket-knowledge.md](./websocket-knowledge.md) | WebSocket 技术知识库 | 深入学习 |
| [troubleshooting-guide.md](./troubleshooting-guide.md) | 完整的问题排查过程 | 问题排查 |

### 验证文档
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [websocket-verification.md](./websocket-verification.md) | 验证指南和方法 | 测试人员 |
| [verification-result.md](./verification-result.md) | 详细验证结果 | 了解细节 |

## 🔍 按问题类型查找

### WebSocket 连接失败
1. 查看 [troubleshooting-guide.md](./troubleshooting-guide.md#问题-1-hijacker-接口缺失)
2. 检查是否实现 Hijacker 接口
3. 查看中间件配置

### 消息发送失败
1. 查看 [troubleshooting-guide.md](./troubleshooting-guide.md#问题-2-goroutine-启动顺序)
2. 检查 goroutine 启动顺序
3. 确认 WriteLoop 已启动

### 认证失败
1. 查看 [websocket-knowledge.md](./websocket-knowledge.md#认证流程)
2. 检查签名验证逻辑
3. 确认 nonce 有效性

### 性能问题
1. 查看 [websocket-knowledge.md](./websocket-knowledge.md#性能考虑)
2. 检查连接池配置
3. 优化 goroutine 数量

## 🎓 学习路径

### 初学者路径
1. [README.md](./README.md) - 了解整体架构
2. [SUMMARY.md](./SUMMARY.md) - 快速了解验证结果
3. [websocket-verification.md](./websocket-verification.md) - 学习如何测试

### 开发者路径
1. [websocket-knowledge.md](./websocket-knowledge.md) - 深入理解技术
2. [verification-result.md](./verification-result.md) - 了解实现细节
3. [troubleshooting-guide.md](./troubleshooting-guide.md) - 学习排查方法

### 问题排查路径
1. [troubleshooting-guide.md](./troubleshooting-guide.md) - 完整排查流程
2. [README.md](./README.md#核心问题与解决方案) - 常见问题
3. [websocket-knowledge.md](./websocket-knowledge.md) - 技术原理

## 📊 文档统计

| 类型 | 数量 | 总字数 |
|------|------|--------|
| 总览文档 | 3 | ~3,000 |
| 技术文档 | 2 | ~15,000 |
| 验证文档 | 2 | ~8,000 |
| **总计** | **7** | **~26,000** |

## 🔗 外部链接

### 相关项目文档
- [../implementation-plan.md](../implementation-plan.md) - 总体实施计划
- [../architecture.md](../architecture.md) - 架构设计文档
- [../progress.md](../progress.md) - 项目进度追踪

### 技术参考
- [RFC 6455 - WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [gorilla/websocket 文档](https://pkg.go.dev/github.com/gorilla/websocket)
- [Go http.Hijacker 接口](https://pkg.go.dev/net/http#Hijacker)

## 💡 使用建议

### 第一次阅读
建议按以下顺序阅读：
1. README.md (5分钟) - 了解整体
2. SUMMARY.md (3分钟) - 快速了解结果
3. troubleshooting-guide.md (15分钟) - 学习排查方法

### 遇到问题时
1. 先查看 README.md 的"核心问题与解决方案"
2. 如果没找到，查看 troubleshooting-guide.md
3. 如果需要深入理解，查看 websocket-knowledge.md

### 学习技术时
1. 从 websocket-knowledge.md 开始
2. 结合 verification-result.md 了解实现
3. 通过 troubleshooting-guide.md 学习实践

---

**最后更新**: 2026-01-30
**文档版本**: v1.0
**维护者**: Phase 3 开发团队
