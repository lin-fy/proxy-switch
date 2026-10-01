# proxy-switch

Windows 桌面工具：管理 Codex Desktop 的 Responses API Provider、Model、Route 和 Profile。

> 项目目标：Provider 只配置一次，通过 Route 选择 Platform、Provider、Model，安全完成切换、测试、备份和恢复。

## 当前状态

V1 开发中。后端核心与桌面壳已完成，Vue 前端迁移基线已建立，桌面原生验收、安装包与真实端到端验收未完成。

详见 [`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md)。

## 快速开始

```bash
# 开发环境要求：Go 1.24+, Node.js 20+, Wails 3 (v3.0.0-beta.26)

# 后端测试
go test ./...

# 前端开发
cd frontend
npm install
npm run dev

# Windows 构建
task build
```

完整构建与安装说明见 [`docs/ops/BUILD_WINDOWS.md`](docs/ops/BUILD_WINDOWS.md)。

## 文档

- **产品**：[`docs/product.md`](docs/product.md) — 产品定位、核心模型、架构决策
- **规划**：[`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md) — V1 目标、里程碑、验收状态
- **架构**：[`docs/architecture/TECH_STACK.md`](docs/architecture/TECH_STACK.md) · [`docs/architecture/UI_ARCHITECTURE_SPEC.md`](docs/architecture/UI_ARCHITECTURE_SPEC.md)
- **构建**：[`docs/ops/BUILD_WINDOWS.md`](docs/ops/BUILD_WINDOWS.md) · [`docs/ops/CI_CD.md`](docs/ops/CI_CD.md)
- **完整索引**：[`docs/README.md`](docs/README.md)

## AI agent 协作

本仓库使用 AI agent 协同开发。**任何 agent 进入仓库前必须先读 [`AGENTS.md`](AGENTS.md)**。

| 文件 | 用途 |
|---|---|
| [`AGENTS.md`](AGENTS.md) | 所有 agent 的通用规则 |
| [`CODEX.md`](CODEX.md) | Codex 专属约定 |

其他 agent（Claude / Gemini / pi / zcode 等）按需补充各自的 `<AGENT名>.md`；没有专属文件时直接遵守 `AGENTS.md`。

协作状态：

- 当前状态：[`docs/CONTEXT.md`](docs/CONTEXT.md)
- 任务看板：[`docs/TASKS.md`](docs/TASKS.md)
- 变更日志：[`docs/CHANGELOG.md`](docs/CHANGELOG.md)
- 技术决策：[`docs/DECISIONS.md`](docs/DECISIONS.md)

## 许可

见 [`LICENSE`](LICENSE)。
