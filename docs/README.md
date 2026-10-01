# 文档索引

本目录是 proxy-switch 项目的文档中心。结构如下：

## 顶层

| 文件 | 作用 | 更新频率 |
|---|---|---|
| [`product.md`](product.md) | 产品定位、核心模型、架构原则 | 随产品演进 |
| [`CONTEXT.md`](CONTEXT.md) | **当前状态快照**，agent 开工/收工必读 | 每次工作会话 |
| [`TASKS.md`](TASKS.md) | 任务认领与进度 | 每次任务变动 |
| [`CHANGELOG.md`](CHANGELOG.md) | 跨 agent 交接的变更记录（与 git log 对齐，琐碎改动不记） | 有重要产出时 |
| [`DECISIONS.md`](DECISIONS.md) | 技术决策记录（ADR） | 有重要决策时 |

## 子目录

| 目录 | 内容 |
|---|---|
| [`planning/`](planning/) | 里程碑规划、实施计划 |
| [`architecture/`](architecture/) | 技术栈、架构规范 |
| [`design/`](design/) | UI 原型、设计选型 |
| [`ops/`](ops/) | 构建、CI/CD、运维 |

## Agent 入口

仓库根目录的 `AGENTS.md` 是所有 AI agent 的通用规则。各家 agent 的专属差异约定按需放在根目录的 `<AGENT名>.md` 中（当前仅有 `CODEX.md`），没有专属文件的 agent 直接遵守 `AGENTS.md`。**任何 agent 进入仓库都应先读 `AGENTS.md`。**

## 写作约定

- 所有文档使用简体中文，技术术语可保留英文。
- 时间格式统一 `YYYY-MM-DD`。
- 路径引用使用相对路径，跨目录引用时检查链接有效性。
- 文档移动或改名后必须更新所有引用它的地方（用 `grep -rln "<旧文件名>" --include="*.md"` 检查）。
