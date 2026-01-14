

# 文档导航（Document Navigation）

本文件用于定义本项目所有设计文档的阅读顺序与依赖关系。  
**任何代码编写之前，必须严格按顺序阅读以下文档。**



>  项目会持续使用 claude code 进行研究学习与规划，故在过程中生成的文档，放在本文档中进行汇总导航
>
>  name 文档名
>  desc 文档描述
>  author 作者（如本人使用lxt,  Ai 生成使用 claude code）



| name               | desc                                 | author |
| ------------------ | ------------------------------------ | ------ |
| design-document.md | 基础设计想法                         | lxt    |
| tech-stack.md      | 技术栈                               | lxt    |
| architecture.md    | 主要功能或完成里程碑后，文件文档记录 |        |
| progress.md        | 跟踪已完成的步骤                     |        |
|                    |                                      |        |
|                    |                                      |        |





# 文档导航（Document Navigation）

本文件用于定义本项目所有设计文档的阅读顺序与依赖关系。  
**任何代码编写之前，必须严格按顺序阅读以下文档。**

---

## 一、必读顺序（强制）

请按以下顺序逐一阅读，不得跳过：

1. `memory-bank/@design-document.md`
   - 项目背景
   - 核心目标
   - 初始设计想法

2. `memory-bank/@tech-stack.md`
   - 技术选型
   - 依赖说明
   - 约束条件

3. `memory-bank/@architecture.md`
   - 当前架构设计
   - 模块划分
   - 关键流程

4. `memory-bank/@progress.md`
   - 当前进度
   - 已完成事项
   - 待办事项

---

## 二、辅助文档（按需）

以下文档在涉及对应内容时必须阅读：

- `prompts/` 目录下所有 prompt 文件  
  - 当涉及协议设计、SQL 生成、规则约束时必须阅读

- `README.md`
  - 项目启动方式
  - 运行说明

---

## 三、强制规则

- 不允许在未阅读上述文档的情况下开始写代码
- 不允许凭经验假设架构
- 不允许跳过 design 或 architecture 文档

---

## 四、阅读完成标记（可选）

在开始实现前，请在此处标记：

- [ ] 已阅读 design-document.md
- [ ] 已阅读 tech-stack.md
- [ ] 已阅读 architecture.md
- [ ] 已阅读 progress.md





***

注：docs-backup/ 目录下文档为之前的规划记录，仅供记录与参考，切勿直接使用
