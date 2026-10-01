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
| [`agent-guide/`](agent-guide/) | agent 详细规则（按需查阅） |
| [`planning/`](planning/) | 里程碑规划、实施计划；`archive/` 存已完成里程碑 |
| [`architecture/`](architecture/) | 技术栈、架构规范 |
| [`design/`](design/) | UI 原型、设计选型 |
| [`ops/`](ops/) | 构建、CI/CD、运维 |

## Agent 入口

仓库根目录的 `AGENTS.md` 是五条铁律（每次开工必读，约 600 token）。详细规则在 [`agent-guide/`](agent-guide/) 按需查阅。各家 agent 的专属差异约定按需放在根目录 `<AGENT名>.md`（当前仅 `CODEX.md`）。

## 写作约定

- 所有文档使用简体中文，技术术语可保留英文。
- 时间格式统一 `YYYY-MM-DD`。
- 路径引用使用相对路径，跨目录引用时检查链接有效性。
- 文档移动或改名后必须更新所有引用它的地方（用 `grep -rln "<旧文件名>" --include="*.md"` 检查）。

## 字符与格式约定

所有 Markdown 文档必须遵守：

- **编码**：UTF-8（不带 BOM）
- **行尾符**：LF（`\n`），不用 CRLF
- **文件结尾**：以单个换行符结尾
- **标点**：
  - 中文句子里用中文标点（，。；：？！）
  - 英文句子和代码里用英文标点（,.;:?!
  - 不混用（避免“中文句子用英文逗号”或反之）
- **引号**：统一用直角引号 `「」` 或英文双引号 `"`，不用弯引号 `“”‘’`
- **破折号**：用两个连字符 `--` 或中文破折号 `——`，不用 en-dash `–` 或 em-dash `—`
- **省略号**：中文用 `……`，英文用 `...`，不用 `…`
- **空格**：
  - 中英文之间加一个空格（如“使用 Vue 3”）
  - 中文与半角标点之间不加空格
  - 全角标点前后不加空格

## 工具

仓库根目录 `.editorconfig` 已声明 `end_of_line = lf`、`charset = utf-8`、`insert_final_newline = true`。编辑器应自动遵守；手工编辑后用 `file <path>` 检查。
