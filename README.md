# proxy-switch

Windows 桌面工具：管理 Codex Desktop 的 Responses API Provider、Model、Route 和 Profile。

> 项目目标：Provider 只配置一次，通过 Route 选择 Platform、Provider、Model，安全完成切换、测试、备份和恢复。

## 当前状态

V1 功能开发与验收基本完成：后端核心、桌面壳、Vue 前端已通过桌面原生验收（窗口/DPI/浮层/键盘/IME），Windows 用户级 NSIS 安装包全流程验证通过。剩余验收事项依赖真实凭据与 GitHub 仓库配置，见 [`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md)。

## 快速开始

```bash
# 开发环境要求：Go 1.25+, Node.js 22+, Wails 3 (v3.0.0-beta.26), Task (go-task)

# 后端测试
go test ./...

# 前端开发
cd frontend
npm install
npm run dev

# Windows 构建
task build
```

完成代码修改后可用同一套模型无关验证入口：

```powershell
task check:frontend
task check:backend
task ci
```

完整构建与安装说明见 [`docs/ops/BUILD_WINDOWS.md`](docs/ops/BUILD_WINDOWS.md)。

## 文档

- **产品**：[`docs/product.md`](docs/product.md) — 产品定位、核心模型、架构决策
- **规划**：[`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md) — V1 目标、里程碑、验收状态
- **架构**：[`docs/architecture/TECH_STACK.md`](docs/architecture/TECH_STACK.md) · [`docs/architecture/UI_ARCHITECTURE_SPEC.md`](docs/architecture/UI_ARCHITECTURE_SPEC.md)
- **构建**：[`docs/ops/BUILD_WINDOWS.md`](docs/ops/BUILD_WINDOWS.md) · [`docs/ops/CI_CD.md`](docs/ops/CI_CD.md)
- **发布**：[`docs/ops/RELEASE_CHECKLIST.md`](docs/ops/RELEASE_CHECKLIST.md) — Tag 发布演练与正式发布步骤
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
