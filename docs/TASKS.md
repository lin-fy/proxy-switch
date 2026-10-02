# 任务认领表

> 多 agent 协作的任务看板。规则见 [`../AGENTS.md`](../AGENTS.md) 第三节。
>
> 状态标记：`[ ]` 未认领 · `[- by <agent> <日期>]` 进行中 · `[x by <agent> <完成日期>]` 已完成 · `[!]` 阻塞

最后更新：2026-10-02

## 进行中

- <暂无>

## 待认领（按优先级排序）

### M4 — Vue 桌面界面收尾

- [x by zcode 2026-10-01] Wails 桌面窗口尺寸、DPI、浮层行为验收（960×640 和 1120×760）
- [x by zcode 2026-10-01] 键盘导航与 IME 输入在桌面端的验收
- [x by zcode 2026-10-01] 真实后端流程端到端走查（Provider/Model/Route/Profile 全流程）

### M5 — Windows 安装包

- [x by codex 2026-10-01] 补齐 NSIS 或 MSIX 安装包工具链
- [x by codex 2026-10-01] 验证安装、启动、升级/回滚、卸载全流程
- [x by codex 2026-10-01] 发布 `docs/ops/BUILD_WINDOWS.md` 最终步骤和故障恢复说明（M5 收尾）

### M6 — V1 验收

- [x by codex 2026-10-02] M6 目标树、验收证据矩阵与 1.0 发布前验证预检
- [- by zcode 2026-10-01] M6 发布准备：README 刷新、发布说明草稿与发布演练清单（版本统一与 Tag 演练按用户决定滞后，待新任务/目标后启动）
- [x by codex 2026-10-02] 按用户授权将仓库公开；配置 `production` Environment、两个测试签名 Secrets、发布审批人和 `main` 分支保护（凭据可用性与签名信任尚待演练验证）
- [ ] 完成一次 Tag 发布演练
- [ ] 用真实 CPA、OpenAI、自定义 Provider 完成端到端验收（**需要用户提供凭据**）

### 持续性工作

- [ ] 每次代码改动后同步更新 `docs/CONTEXT.md` 和 `docs/CHANGELOG.md`
- [ ] 重要技术决策追加到 `docs/DECISIONS.md`

## 已完成

- [x by codex 2026-10-02] 按用户授权合并 PR #10 到 `dev`：V1 目标/发布门槛文档与文档 PR 的 `verify` 触发修复；合并提交 `38354c8`，合并后 Windows CI 通过（运行 `36964118419`）

- [x by codex 2026-10-01] 后端强制「Route 只能引用已启用 Model」：`route.Service` 创建/保存与 `Activator` 激活时校验 `Enabled`，阻止停用模型仍被激活；补 `ErrModelDisabled` 与前端安全错误映射

- [x by zcode 2026-10-01] 前端产物分包与可访问性标签：manualChunks 拆分（业务 index 638→37 kB，单 chunk 回到警告线内）、配置档案选择器空态无名 tab stop 修复；CI 通过后 PR #4 合并
- [x by codex 2026-10-01] 完善模型无关 CI 入口与验证文档；`task check:frontend`、`task check:backend`、`task ci` 全部通过，Windows exe 与 portable ZIP 生成成功；分支 `codex/agent-independent-ci`
- [x by codex 2026-10-02] 将 CI 与发布 PowerShell 脚本兼容到 Windows PowerShell 5.1；显式检查原生命令退出码，保留 PowerShell 7 CI 行为
- [x by codex 2026-10-01] 审查并集成 GitHub PR #1；修复 Windows CI 换行与 Wails 图标路径问题，PR 已合并到 `dev`（`4537f4c`）
- [x by zcode 2026-10-01] M4 桌面验收：窗口 1120×760/最小 960×640 落地与 DPI 实测、最小窗口裁切修复、Tab/Escape/IME 验证、Mock Responses 全链路走查（同步模型入口缺失与错误态清页两个 P1 修复）；分支 `zcode/m4-acceptance`
- [x by codex 2026-10-01] 明确并行分支版本同步规则：开发期间允许落后，合并前在 agent worktree 合并最新 `dev` 并重新验证
- [x by codex 2026-10-01] 按用户确认切换 GitHub PR 流程：多 agent 时人工指定 reviewer/merger，单 agent 时允许自审并在分支保护允许时合并且必须通过 CI
- [x by codex 2026-10-01] 文档结构整理：建立多 agent 协作机制（AGENTS.md + 各 agent 入口 + docs/CONTEXT/TASKS/CHANGELOG/DECISIONS）
- [x by zcode 2026-10-01] 前端基线缺口修复与视觉重做（令牌单一来源、错误安全映射、表单 path/错误摘要、懒加载、960/1120 走查）
- [x by zcode 2026-10-01] 桌面构建复验与用户级 Go 1.25 工具链安装（wails3 build + 启动冒烟 + 生产包演示数据剔除检查）
- [x by zcode 2026-10-01] 提交工作区全部未提交改动（6 组方案经用户确认；backend/ci 两组为整理 codex 之前未提交的工作）

## 阻塞

- <暂无>
