# 文档导航（Document Navigation）

本文件用于定义本项目所有设计文档的阅读顺序与依赖关系。  
**任何代码编写之前，必须严格按顺序阅读以下文档。**

> 项目会持续使用 claude code 进行研究学习与规划，故在过程中生成的文档，放在本文档中进行汇总导航
>
> - **name**: 文档名
> - **desc**: 文档描述
> - **author**: 作者（本人使用 lxt，AI 生成使用 claude code）

---

## 一、核心文档（必读，按顺序）

| 序号 | 文档名 | 描述 | 作者 | 用途 |
| --- | --- | --- | --- | --- |
| 1 | `design-document.md` | 基础设计想法 | lxt | 项目背景、核心目标、初始设计 |
| 2 | `tech-stack.md` | 技术栈说明 | lxt | 技术选型、依赖说明、约束条件 |
| 3 | `architecture.md` | 架构与里程碑记录 | claude code | 当前架构设计、模块划分、关键流程 |
| 4 | `progress.md` | 已完成步骤追踪 | claude code | 当前进度、已完成事项、待办事项 |

---

## 二、实施与规划文档（按需阅读）

| 文档名 | 描述 | 作者 | 用途 |
| --- | --- | --- | --- |
| `implementation-plan.md` | 实施规划 | claude code | 详细的 10 阶段开发路线图 |
| `usages.md` | 工具使用说明 | claude code | Makefile 等工具使用方法 |

---

## 三、专题文档（涉及相关功能时必读）


| 文档名 | 描述 | 作者 | 用途 |
| --- | --- | --- | --- |
| `charline_jwt_ed25519_auth.md` | 认证方案总结 | lxt | JWT + Ed25519 私钥签名认证方案 |
| `phase2-1-login-auth.md` | Phase 2.1 实现记录 | claude code | Nonce 签名登录实现与审查 |
| `websocket-protocol-spec.md` | WebSocket 协议规范 | claude code | WebSocket 消息格式、认证流程、心跳机制 |
| `session-management-faq.md` | Session 管理 FAQ | claude code | 多设备、Resume Token、断线检测常见问题 |
| `database-migration-system.md` | 数据库迁移系统 | claude code | SQLite 迁移系统工作原理、使用方法、最佳实践 |
| `code-optimization-todo.md` | 代码优化待办 | claude code | golangci-lint 检查发现的代码质量问题 |
| `thinking-challenge/response-to-auth-challenge.md` | IM 架构重整方案 | claude code | 回应 ChatGPT 5.2 挑战，Phase 3 设计方向 |

---

## 四、辅助资源（按需参考）

- **`prompts/` 目录**：当涉及协议设计、SQL 生成、规则约束时必须阅读
- **`README.md`**：项目启动方式、运行说明
- **`docs-backup/` 目录**：之前的规划记录，仅供参考，切勿直接使用

---

## 五、强制规则

- ❌ 不允许在未阅读核心文档的情况下开始写代码
- ❌ 不允许凭经验假设架构
- ❌ 不允许跳过 design-document 或 architecture 文档
- ✅ 必须按顺序阅读核心文档（1→2→3→4）
- ✅ 涉及专题功能时必须阅读对应专题文档

---

## 六、阅读完成标记（可选）

在开始实现前，请在此处标记：

- [ ] 已阅读 design-document.md
- [ ] 已阅读 tech-stack.md
- [ ] 已阅读 architecture.md
- [ ] 已阅读 progress.md
- [ ] 已阅读相关专题文档

---

## 七、文档更新规则

当发生以下情况时，必须更新对应文档：

| 情况 | 需要更新的文档 |
| --- | --- |
| 新增主要功能 | `architecture.md`, `progress.md` |
| 完成一个里程碑 | `progress.md`, `architecture.md` |
| 架构发生调整 | `architecture.md` |
| 模块职责发生变化 | `architecture.md` |
| 新增专题文档 | `document-navigation.md`（本文件） |
| 发现代码质量问题 | `code-optimization-todo.md` |
| 数据库结构变更 | `database-migration-system.md`, `architecture.md` |

---

**最后更新时间**: 2026-02-03
