# 代码优化待办事项

> 本文件记录 golangci-lint 检查发现的代码质量问题，用于后续优化。
> 已记录的问题会在 .golangci.yml 中配置忽略，避免重复提示。

---

## 服务端优化事项（10 个）

### errcheck - 未检查错误返回值（6 个）

| 文件 | 行号 | 问题 | 优先级 | 状态 |
| --- | --- | --- | --- | --- |
| `server/cmd/main.go` | 36 | `defer log.Sync()` 未检查错误 | 低 | 待优化 |
| `server/cmd/main.go` | 52 | `defer db.Close()` 未检查错误 | 低 | 待优化 |
| `server/internal/httputil/response.go` | 29 | `json.NewEncoder(w).Encode()` 未检查错误 | 中 | 待优化 |
| `server/internal/httputil/response.go` | 58 | `json.NewEncoder(w).Encode()` 未检查错误 | 中 | 待优化 |
| `server/internal/router/router.go` | 65 | `w.Write([]byte("OK"))` 未检查错误 | 低 | 待优化 |
| `server/internal/store/invite.go` | 111 | `tx.Rollback()` 未检查错误 | 中 | 待优化 |

**优化建议**：
- defer 中的 Close() 和 Sync() 可以使用 `_ = xxx()` 显式忽略
- HTTP 响应的 Encode() 和 Write() 应该检查错误并记录日志
- 事务回滚应该检查错误并记录

### staticcheck - 代码优化建议（4 个）

| 文件 | 行号 | 问题 | 建议 | 状态 |
| --- | --- | --- | --- | --- |
| `server/internal/errors/codes.go` | 146 | `if e.Details == nil \|\| len(e.Details) == 0` | 改为 `if len(e.Details) == 0` | 待优化 |
| `server/internal/httputil/response.go` | 54 | `if bizErr.Details != nil && len(bizErr.Details) > 0` | 改为 `if len(bizErr.Details) > 0` | 待优化 |
| `server/internal/logger/logger.go` | 27 | `return a.Config.GetZapLevel()` | 改为 `return a.GetZapLevel()` | 待优化 |
| `server/internal/store/invite.go` | 284 | `if !((c >= '0' && c <= '9') \|\| (c >= 'A' && c <= 'Z'))` | 应用德摩根定律简化 | 待优化 |

**优化建议**：
- Go 中 `len(nil)` 返回 0，可以省略 nil 检查
- 嵌入字段可以直接调用方法
- 复杂条件可以简化提高可读性

---

## 客户端优化事项（6 个）

### errcheck - 未检查错误返回值（4 个）

| 文件 | 行号 | 问题 | 优先级 | 状态 |
| --- | --- | --- | --- | --- |
| `client/cmd/main.go` | 28 | `defer log.Sync()` 未检查错误 | 低 | 待优化 |
| `client/internal/commands/join.go` | 128 | `defer httpResp.Body.Close()` 未检查错误 | 低 | 待优化 |
| `client/internal/commands/login.go` | 117 | `defer httpResp.Body.Close()` 未检查错误 | 低 | 待优化 |
| `client/internal/commands/login.go` | 170 | `defer httpResp.Body.Close()` 未检查错误 | 低 | 待优化 |

**优化建议**：
- defer 中的 Close() 和 Sync() 可以使用 `_ = xxx()` 显式忽略

### staticcheck - 代码优化建议（2 个）

| 文件 | 行号 | 问题 | 建议 | 状态 |
| --- | --- | --- | --- | --- |
| `client/cmd/main.go` | 81 | `return a.Config.GetZapLevel()` | 改为 `return a.GetZapLevel()` | 待优化 |
| `client/internal/logger/logger.go` | 27 | `return a.Config.GetZapLevel()` | 改为 `return a.GetZapLevel()` | 待优化 |

**优化建议**：
- 嵌入字段可以直接调用方法

---

## 优化优先级说明

- **高**：影响功能或安全性，需要尽快修复
- **中**：影响代码质量，建议在下次迭代修复
- **低**：代码风格问题，可以延后处理

---

## 更新记录

| 日期 | 更新内容 |
| --- | --- |
| 2025-01-29 | 初始创建，记录 golangci-lint 首次检查结果 |

