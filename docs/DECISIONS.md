# 技术决策记录（ADR）

> 架构和技术选型决策记录。格式见 [`../AGENTS.md`](../AGENTS.md) 第 3.4 节。编号从 ADR-0001 开始递增，**不重用编号**。

---

## ADR-0001：使用 Wails 3 而非 Tauri / Electron / 远程网页

- 日期：2026-09-30（V1 启动时确定）
- 决策人：用户 + codex
- 状态：accepted
- 背景：需要 Windows 桌面壳承载 Vue 前端，调用 Go 后端能力。
- 决策：Wails 3，锁定版本 `v3.0.0-beta.26`。
- 理由：Go 后端一体化、体积小、Windows 集成度好；不需要引入 Node.js 后端。
- 后果：跨平台支持受限于 Wails 3 能力；版本锁定在 beta，存在 API 变更风险，需要持续关注上游。

## ADR-0002：前端使用 Vue 3 + Pinia + Naive UI + Tailwind v4

- 日期：2026-09-30
- 决策人：用户 + codex
- 状态：accepted
- 背景：需要选择 Vue 生态下的 UI 组件库与样式方案。
- 决策：Vue 3 + TypeScript + Pinia 状态管理 + Naive UI 控件 + Tailwind v4 管布局与令牌。
- 理由：Naive UI 提供完整桌面控件且支持主题定制；Tailwind v4 与 Vite 集成度高；Pinia 是 Vue 3 官方推荐。比较依据见 [`design/UI_LIBRARY_COMPARISON.md`](design/UI_LIBRARY_COMPARISON.md)。
- 后果：不引入 Reka UI / shadcn-vue；混合样式边界明确——Tailwind 管布局，Naive 管控件。

## ADR-0003：Provider 作为数据而非工厂

- 日期：2026-09-30
- 决策人：用户 + codex
- 状态：accepted
- 背景：需要支持 CPA、OpenAI、自定义 Responses Provider，是否要为每家建独立工厂类。
- 决策：Provider 作为数据模型，不为每个供应商创建独立工厂。
- 理由：V1 只接 Responses API，协议统一；用工厂会导致不必要的抽象。
- 后果：新增 Provider 类型只需数据扩展，不需要新增代码分支。

## ADR-0004：V1 仅支持 Codex Desktop 平台适配器

- 日期：2026-09-30
- 决策人：用户 + codex
- 状态：accepted
- 背景：产品规划上未来要支持 Claude Code / Gemini CLI / OpenCode，V1 范围怎么定。
- 决策：V1 只实现 `CodexAdapter`，但保留 `PlatformAdapter` 接口和注册表机制。
- 理由：集中精力打通主链路，避免多平台同时推进分散风险。
- 后果：V1.1+ 新增平台时需新增 Adapter 实现，不需要重构现有架构。

## ADR-0005：凭据使用引用机制，不直接存储

- 日期：2026-09-30
- 决策人：用户 + codex
- 状态：accepted
- 背景：API Key 等敏感信息的存储方式。
- 决策：状态文件只保存 `credential:<target>` 或 `${ENV_VAR}` 引用，运行时通过 Windows Credential Manager 或环境变量解析。
- 理由：避免密钥进入 Git、备份、日志；符合最小权限原则。
- 后果：用户需要额外配置凭据来源；提供 UI 引导。

## ADR-0006：建立多 agent 协作机制（AGENTS.md + 各家入口 + docs 协作文件）

- 日期：2026-10-01
- 决策人：用户 + codex
- 状态：accepted
- 背景：用户使用多个 AI agent（Codex / Claude / Gemini / pi / zcode）协同工作，原有 `AGENTS.md` 只覆盖 Codex，缺少统一协作规范，文档结构混乱。
- 决策：
  1. `AGENTS.md` 作为所有 agent 的通用规则底座；
  2. `CODEX.md` / `CLAUDE.md` / `GEMINI.md` / `PI.md` / `ZCODE.md` 等专属入口**按需存在**——只有当某个 agent 确实需要差异化约定时才创建，没有专属文件的 agent 直接遵守 `AGENTS.md`；
  3. `docs/CONTEXT.md` 作为当前状态快照（single source of truth）；
  4. `docs/TASKS.md` 作为任务认领看板；
  5. `docs/CHANGELOG.md` 作为追加式变更日志；
  6. `docs/DECISIONS.md`（本文件）作为 ADR 记录；
  7. `docs/` 按 `planning/` `architecture/` `design/` `ops/` 分类归档；
  8. `PRODUCT_PLAN.md` 改名 `docs/product.md`，`PROJECT_INSTRUCTIONS.md` 改名 `CODEX.md`。
- 理由：统一入口降低 agent 接入成本；状态、任务、日志、决策四类高频更新文档单独成文避免相互污染；按主题归档让 docs/ 结构清晰。
- 后果：所有 agent 必须遵守新规则；历史文档引用路径已全部更新；后续新增文档需遵循同一分类约定。

## ADR-0007：浏览器预览演示数据模式（仅开发构建生效）

- 日期：2026-10-01
- 决策人：zcode 实现，用户确认（交付物是桌面工具而非网页）
- 状态：accepted
- 背景：前端界面走查需要可运行的数据环境，但应用能力依赖 Wails 宿主；纯浏览器打开生产构建只能看到后端错误态，无法走查界面状态。
- 决策：`frontend/src/services/wails-api.ts` 仅在「开发构建（`import.meta.env.DEV`）且检测不到 Wails runtime（`window._wails.environment`）」时启用内存演示数据（`services/demo-data.ts`），界面在侧栏与底部状态栏显示「预览数据」标记。生产构建经产物检查确认演示数据被完整剔除。
- 理由：不修改 Go API 与绑定，让 `npm run dev` 成为界面开发的快速预览环境；最终交付形态保持为 Wails 桌面工具。
- 后果：浏览器 `npm run dev` 中看到的是演示数据而非真实配置；任何真实行为验证必须在 `wails3 build` 产出的桌面 exe 中进行。

## ADR-0008：并行分支在合并前同步最新 dev

- 日期：2026-10-01
- 决策人：用户 + codex
- 状态：accepted
- 背景：多个 agent 并行开发时，各自 worktree 会自然落后于 `dev`；强制实时同步会打断开发，直接合并又可能把旧代码带入主分支。
- 决策：开发期间允许 agent 分支暂时落后；准备交付时，agent 必须在自己的 worktree 执行 `git merge dev`，在该分支解决冲突并重新运行任务验证，验证通过后由 root/Codex 合并到 `dev`。
- 理由：把版本冲突处理放在隔离工作树中，避免主工作树出现半解决状态；保留已有提交，不要求 force push 或改写分支历史。
- 后果：合并前需要额外一次同步和验证；被忽略的依赖、缓存和构建产物仍需通过独立初始化流程生成，Git 合并不会自动补齐。

## ADR-0009：当前阶段由 root/Codex 统一集成

- 日期：2026-10-01
- 决策人：用户 + codex
- 状态：superseded by ADR-0010
- 背景：多个 agent 同时完成任务时，多个 agent 直接操作主工作树会产生合并竞态；本地锁只能降低风险，不能阻止绕过流程的操作。
- 决策：在 zcode 任务和本轮 Codex 任务全部结束前，任务 agent 只负责自己的分支、同步、冲突处理和验证；公共文档更新与 `dev` 合并由当前对话的 root/Codex 统一执行。两项任务结束后再切换 GitHub PR 和 Merge Queue。
- 理由：主工作树只有一个写入者，流程简单且可见；后续用 GitHub 分支保护和合并队列承接多人集成。
- 后果：root/Codex 需要按顺序处理 agent 的交付；任务 agent 不能自行修改主工作树或合并分支。该过渡方案已由 ADR-0010 替代。

## ADR-0010：切换 GitHub PR 并由用户指定审查与合并者

- 日期：2026-10-01
- 决策人：用户 + codex
- 状态：accepted
- 背景：zcode 任务和当前规则梳理已完成，需要从本地单一集成者流程切换到可见、可审查的 GitHub 协作流程。
- 决策：任务 agent 在自己的分支完成代码、文档和验证后创建目标为 `dev` 的 GitHub PR。多个 agent 并行时由用户人工指定 reviewer 和 merger；只有一个活跃 agent 时，该 agent 可以自审并在分支保护允许时合并自己的 PR，但必须等待 CI 通过；若分支保护要求独立批准，仍需指定 reviewer。合并后以 PR 和 CI 结果作为代码完成证据。
- 理由：PR 提供差异、审查意见、检查结果和合并记录；多人并行时人工指定可以根据任务风险选择合适的审查者和合并者，单 agent 时保留完整 PR/CI 轨迹又不增加等待，同时不绕过 GitHub 分支保护。
- 后果：任务完成需要远端分支和 PR；本地 worktree 仍用于开发和版本同步，主工作树不直接合并 agent 分支。

## ADR-0011：CI 与开发 agent、模型解耦

- 日期：2026-10-01
- 决策人：用户 + codex
- 状态：accepted
- 背景：项目会在 Codex、Pi、Claude Code 等 agent 之间切换；把 CI 绑定某个模型会让开发入口和发布可靠性随 agent 改变。
- 决策：`scripts/ci/verify.ps1` 作为唯一验证实现，`Taskfile.yml` 提供前端、后端和完整 CI 入口；GitHub Actions 只执行这些确定性检查。AI Review 或失败诊断只能作为独立的可选分析层。
- 理由：任何 agent 都能调用相同命令，CI 结果可复现，模型或中转服务不可用时仍能完成构建、测试和发布。
- 后果：需要 AI 分析时必须另行配置凭据、权限和失败触发策略；AI 不直接拥有生产发布权限。
