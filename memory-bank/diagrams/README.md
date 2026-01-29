# 项目图表目录

本目录存放项目相关的流程图、时序图、架构图等。

## 文件列表

| 文件名 | 描述 | 格式 | 关联文档 |
| --- | --- | --- | --- |
| `phase2.1-auth-flow.mmd` | Phase 2.1 Nonce 签名登录认证流程 | Mermaid | `../progress.md` |

## 使用说明

### Mermaid 图表

Mermaid 是一个基于文本的图表工具，支持在 Markdown 中直接渲染。

**在线预览**：
- [Mermaid Live Editor](https://mermaid.live/)
- GitHub/GitLab 自动渲染 Markdown 中的 Mermaid 代码块

**本地预览**：
```bash
# 安装 Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# 生成 PNG
mmdc -i phase2.1-auth-flow.mmd -o phase2.1-auth-flow.png

# 生成 SVG
mmdc -i phase2.1-auth-flow.mmd -o phase2.1-auth-flow.svg
```

**VS Code 插件**：
- [Markdown Preview Mermaid Support](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid)

## 图表规范

### 时序图（Sequence Diagram）

- 使用 `sequenceDiagram` 关键字
- 参与者命名清晰（Client、Server、Database）
- 使用 `Note` 添加关键说明
- 使用 `-->>` 表示返回消息
- 使用 `->>` 表示请求消息

### 命名规范

- 文件名：`phaseX.Y-功能描述.mmd`
- 标题：简洁明了，包含阶段和功能
- 注释：中文，清晰说明关键步骤

---

**最后更新**: 2025-01-29
