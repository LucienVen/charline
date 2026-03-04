# Agent Guidelines (Codex / Charline)

> 本文件是本仓库代理规则的唯一来源（Single Source of Truth）。
> 所有规则变更仅在本文件维护与更新。

## 1. 项目目标
- 项目为 Golang 的 client-server 架构 IM 通讯应用。
- 以高性能、模块化、可维护、可演进为最高优先级。

## 2. 沟通规范
- 默认使用中文（除非用户明确要求英文）。
- 输出直接、结构化、可执行，不输出无关寒暄。
- 非用户要求时，不主动输出“计划/实施步骤”。
- 不使用 emoji。

## 3. 编码前强制阅读
开始写代码前必须先读 `memory-bank/document-navigation.md`，并按顺序阅读：
1. `memory-bank/design-document.md`
2. `memory-bank/tech-stack.md`
3. `memory-bank/architecture.md`
4. `memory-bank/progress.md`

涉及专题功能时，必须读取对应专题文档。  
`docs-backup/` 仅作历史参考，不能作为当前实现依据。

## 4. 架构与实现约束
- 禁止单一大文件堆逻辑，必须按职责拆分模块与文件。
- 保持分层边界清晰，避免跨层直接依赖。
- 遵循 Go 最佳实践：显式错误处理、显式依赖注入、最小化全局状态。
- 网络层仅处理连接、读写、心跳、重连，不混入业务逻辑。
- 并发实现应控制 goroutine 数量，避免阻塞链路和不必要拷贝。
- 代码变更后保持 gofmt 风格并进行必要验证。

## 5. 规则文件优先级
- 涉及 SQL、协议、消息格式、业务规则时，先读 `prompts/` 对应文件再实现。
- 若实现与 prompts 冲突，以 prompts 约束为准。

## 6. memory-bank 跟踪机制（必须执行）
以下情况需要同步更新文档：
- 新增主要功能 / 里程碑完成：`memory-bank/architecture.md`、`memory-bank/progress.md`
- 架构调整 / 模块职责变化：`memory-bank/architecture.md`
- 新增专题文档：`memory-bank/document-navigation.md`
- 数据库结构变更：`memory-bank/database-migration-system.md`、`memory-bank/architecture.md`
- 发现质量问题：`memory-bank/code-optimization-todo.md`

## 7. 默认工作流程
1. 读文档并确认上下文  
2. 设计模块边界与文件拆分  
3. 实现最小必要改动  
4. 执行构建/测试/检查  
5. 同步更新 memory-bank 文档并记录原因

## 8. 文档维护规则
- 代理规则发生变更时，直接修改 `AGENTS.md`。
- 不再维护 `AGENT.md` 的镜像或同步流程，避免规则漂移。
- 提交 PR 时需在描述中标注：已更新代理规则文档（`AGENTS.md`）。
