# 变更日志

> 跨 agent 协作的交接记录。规则见 [`../AGENTS.md`](../AGENTS.md) 第 3.2 节。
>
> **格式对齐 [Conventional Commits](https://www.conventionalcommits.org/)**，未来可通过 `git-cliff` 等工具从 git log 自动重建。
>
> 琐碎改动（typo、格式化、重命名变量）**不记**，直接看 `git log` 即可。

---

## 2026-10-01 · codex

### fix(build): support user-local NSIS compiler

- **影响**：Windows NSIS 任务可通过 `MAKENSIS` 使用用户目录中的便携编译器，不再依赖 Chocolatey 或系统 `PATH`；已实际生成 amd64 用户级安装包，并验证安装、启动、升级、回滚和卸载。
- **关键文件**：`Taskfile.yml`、`build/windows/Taskfile.yml`、`docs/ops/BUILD_WINDOWS.md`
- **后续**：继续验证安装包升级/回滚，并补充 MSIX 路径。
- **commit**：24be915（代码分支 `codex/m5-installer-toolchain`）

## 2026-10-01 · codex

### build(windows): add NSIS toolchain setup and preflight

- **影响**：Windows 打包任务提供 NSIS 安装入口，并在生成安装器前给出明确的 `makensis.exe` 缺失提示。
- **关键文件**：`Taskfile.yml`、`build/windows/Taskfile.yml`
- **后续**：在具备 NSIS 的 Windows 环境完成真实安装、启动、升级/回滚和卸载验收；当前分支未改变系统级 Wails/NSIS 安装。
- **commit**：108cdc2（代码分支 `codex/m5-installer-toolchain`）

## 2026-10-01 · codex

### docs(agents): 建立多 agent 协作机制并明确提交前文档更新硬规则

- **影响**：所有 agent 协作流程
- **关键文件**：`AGENTS.md`、`CODEX.md`、`README.md`、`docs/README.md`、`docs/CONTEXT.md`、`docs/TASKS.md`、`docs/CHANGELOG.md`、`docs/DECISIONS.md`、`docs/product.md`（原 `PRODUCT_PLAN.md`）、`docs/planning/`、`docs/architecture/`、`docs/design/`、`docs/ops/`、`.agents/skills/review-pr/SKILL.md`
- **后续**：所有 agent 后续工作必须按 `AGENTS.md` 第 3.2 节顺序执行（TASKS → CHANGELOG → DECISIONS → CONTEXT → commit）
- **commit**：4d554f5

## 2026-10-01 · zcode

### feat(ui): align desktop UI with prototype and harden form flows

- **影响**：frontend/src 全部页面与壳层重做；Naive UI 主题改为从 tokens.css 计算值生成（单一令牌来源）；错误安全映射、激活 recovered=unknown、表单 path/错误摘要/脏关闭守卫、Profiles/Settings 懒加载、muted 对比度 ≥4.5:1；生产构建已确认剔除浏览器预览演示数据
- **关键文件**：`frontend/src/app/theme.ts`、`frontend/src/app/form.ts`（新）、`frontend/src/services/wails-api.ts`、`frontend/src/services/demo-data.ts`（新，仅 dev 生效）、`frontend/src/stores/workspace.ts`、`frontend/src/pages/`、`frontend/src/components/DesktopShell.vue`、`frontend/src/styles/`、`frontend/src/router.ts`、`frontend/index.html`
- **后续**：桌面原生窗口 DPI / 浮层 / 键盘 / IME 验收仍待（见 TASKS M4）
- **commit**：7da2b3a

### build: 复验 Wails 桌面构建并安装用户级 Go 工具链

- **影响**：Go 工具链改为用户级安装（go1.25.14，位于本机 AppData\Local\go-sdk\go，旧 codex-go runtime 已失效）；桌面 exe（约 11.4 MB）重新构建并启动冒烟通过；`docs/ops/BUILD_WINDOWS.md` 已更新 PATH 说明
- **关键文件**：`docs/ops/BUILD_WINDOWS.md`
- **commit**：随 docs(agents) 组提交（说明性文档变更）

### docs(collab): 按用户确认的分组提交工作区历史改动

- **影响**：工作区 88 个路径分 6 组提交完毕。其中 `feat(backend)`（2d2f821）与 `build(ci)`（417d4d6）两组是**整理 codex 之前未提交的工作**（后端 Provider 模型同步/引用完整性/凭据引用助手 + GitHub Actions/打包脚本），zcode 仅做分组、命名与提交，未改动代码内容；`chore(frontend)`（2046906）为前端 lint/format 工具链配置
- **关键文件**：见 `git log` 对应提交
- **commit**：f8a718e

## 2026-10-01 · codex (recorded by zcode)

### docs(agents): split rules into agent-guide and archive completed milestones

- **影响**：AGENTS.md 瘦身为通用规则入口，细则拆分到 docs/agent-guide/（code-rules / escalation / git-rules / scope / workflow）；ROADMAP 已完成的 M0–M2 里程碑归档到 docs/planning/archive/；相关文档交叉引用与行尾同步整理；CI 增加 paths-ignore 对纯文档变更跳过构建
- **关键文件**：`AGENTS.md`、`docs/agent-guide/`、`docs/planning/archive/M0-M2.md`、`docs/planning/ROADMAP.md`、`docs/CONTEXT.md`、`docs/TASKS.md`、`docs/README.md`、`.github/workflows/ci.yml`
- **后续**：无
- **commit**：ef01776（ci 跳过：8bd05b8；交叉引用与行尾整理：11df521，由 zcode 收尾提交）

## 2026-10-01 · codex

### docs(agents): let agents merge verified task branches

- **影响**：按用户指示，已验收的并行任务由所属 agent 在主工作树自行合并到 `dev`，无需用户执行或再次确认；同步流程和示例，保留他人改动保护及远端推送授权要求。
- **关键文件**：`AGENTS.md`、`docs/agent-guide/git-rules.md`、`docs/TASKS.md`、`docs/CONTEXT.md`
- **commit**：待下次文档提交回填
