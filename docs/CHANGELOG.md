# 变更日志

> 跨 agent 协作的交接记录。规则见 [`../AGENTS.md`](../AGENTS.md) 第 3.2 节。
>
> **格式对齐 [Conventional Commits](https://www.conventionalcommits.org/)**，未来可通过 `git-cliff` 等工具从 git log 自动重建。
>
> 琐碎改动（typo、格式化、重命名变量）**不记**，直接看 `git log` 即可。

---

## 2026-10-02 · codex

### ops(github): make repository public and enable release protections

- **影响**：按用户明确授权将 `lin-fy/proxy-switch` 从私有改为公开，解决原分支保护和 Environment 审批的方案限制。
- **验证**：GitHub API 确认 `visibility=public`；`main` 要求 PR、`verify` 通过、最新分支和对话解决，规则对管理员生效，禁止强推/删除；单人仓库强制批准人数为 0。`production` 已添加 `lin-fy` 为审批人，允许本人批准；两个 Secret 名称仍存在，未读取其值。
- **关键文件**：`docs/TASKS.md`、`docs/planning/ROADMAP.md`、`docs/DECISIONS.md`、`docs/CONTEXT.md`
- **后续**：当前 Secrets 为自签名测试凭据，签名和 Tag 演练仍待验证；未创建 Tag 或 Release。
- **commit**：待下次文档提交回填

## 2026-10-02 · codex

### docs(planning): align V1.0 goal tree and M6 acceptance gates

- **影响**：修正 ROADMAP 中已完成的 Provider/Route/自动重启状态，新增 M6-A 到 M6-E 的执行顺序与验收证据矩阵，并同步产品状态与 CI/CD 的 `v1.0.0` 发布目标。
- **关键文件**：`docs/planning/ROADMAP.md`、`docs/product.md`、`docs/ops/CI_CD.md`、`docs/CONTEXT.md`、`docs/TASKS.md`
- **验证**：`scripts/ci/verify.ps1 -BuildWindows` 通过；生成 Windows amd64 portable ZIP；工作区无生成文件改动。
- **后续**：M6-C 需要真实 Provider 凭据，M6-D 需要 GitHub Environment、签名 Secrets 和分支保护。
- **commit**：待提交后补充

### docs(planning): prioritize GitHub release configuration

- **影响**：按用户选择先推进 M6-D GitHub 发布配置；M6-C 真实 Provider 验收延后，M6-E 仍需两者都通过后执行。
- **关键文件**：`docs/CONTEXT.md`、`docs/planning/ROADMAP.md`
- **后续**：按 `docs/ops/RELEASE_CHECKLIST.md` 完成仓库设置；当前不创建 Tag、不切换 `1.0.0`。

### docs(ops): record GitHub branch protection plan limitation

- **影响**：只读审计确认仓库为私有仓库，当前 GitHub 方案对分支保护和 Rulesets 返回 403；`production` Environment 尚未创建，远端没有 Tag。
- **关键文件**：`docs/CONTEXT.md`
- **后续**：用户需升级 GitHub 方案或将仓库设为公开后，才能完成 `main` 分支保护；Environment 和 Secrets 仍需用户配置。

### ops(github): create production Environment

- **影响**：已在 `lin-fy/proxy-switch` 创建 `production` Environment，并配置 `WINDOWS_CERTIFICATE_BASE64` 与 `WAILS_WINDOWS_CERT_PASSWORD` 自签名测试凭据；当前没有保护规则。
- **后续**：正式发布仍需替换为受信任的 Windows Authenticode PFX 证书，并解决私有仓库的分支保护方案限制。

### fix(ci): support Windows PowerShell 5.1 in verification scripts

- **影响**：CI 验证、发布版本守卫及其测试不再要求 PowerShell 7.3；通过显式检查原生命令退出码，保留 5.1 下的失败检测行为。
- **关键文件**：`scripts/ci/verify.ps1`、`scripts/ci/validate-release.ps1`、`scripts/ci/validate-release.test.ps1`、`.agents/skills/review-pr/scripts/review-checks.ps1`
- **后续**：在目标分支的完整 GitHub Windows CI 中复验。
- **commit**：待提交后补充

## 2026-10-01 · codex

### fix(route): reject disabled model when creating or activating routes

- **影响**：修复 V1 主链路不变量漏洞。前端只允许选择已启用 Model，但后端 `route.Service` 创建/保存与 `Activator` 激活都不校验 `Enabled`，已停用模型的旧 Route 仍可被激活并写入 Codex 配置。现新增 `ErrModelDisabled`，`Create`/`Save`/`Activate` 均在可达性检查、配置写入和重启之前拦截。
- **关键文件**：`internal/application/route/service.go`、`internal/application/route/activate.go`、`internal/application/route/service_test.go`、`internal/application/route/activate_test.go`、`frontend/src/services/wails-api.ts`
- **后续**：无。激活失败时前端按安全映射提示“所选模型已停用”。
- **commit**：`078bf92`

## 2026-10-01 · codex

### docs(ops): finalize Windows build, install and recovery guide

- **影响**：M5 收尾。`docs/ops/BUILD_WINDOWS.md` 由开发笔记升级为最终操作说明，补齐工具链前提、统一验证入口、NSIS/portable/MSIX 三条打包路径、签名与发布、运行时配置、卸载残留清理，以及构建/安装升级/启动/Codex 配置/发布分阶段故障恢复和 V1 安装验收清单。
- **关键文件**：`docs/ops/BUILD_WINDOWS.md`、`docs/TASKS.md`、`docs/planning/ROADMAP.md`、`docs/CONTEXT.md`
- **后续**：M5 全部完成；剩余 M6 依赖用户配置 GitHub `production` Environment、签名 Secrets、`main` 分支保护和真实 Provider 凭据。
- **commit**：`336a7dd`（PR #3）

## 2026-10-01 · codex

### ci: standardize agent-independent verification entrypoints

- **影响**：Codex、Pi、Claude Code 等 agent 共用 `task check:frontend`、`task check:backend` 和 `task ci`；GitHub Actions 继续只执行硬检查，不依赖 LLM。
- **关键文件**：`Taskfile.yml`、`scripts/ci/verify.ps1`、`.github/workflows/ci.yml`、`docs/ops/CI_CD.md`、`AGENTS.md`
- **验证**：前端/后端范围入口、完整 `task ci` 与 Windows amd64 exe/portable ZIP 生成均通过；Wails 从 `GOPATH\bin` 自动查找，仅修改验证进程的 PATH。
- **后续**：如需 AI Review 或失败诊断，应作为独立可选工作流接入，不能替代 CI 结果。
- **commit**：`ae7271b`

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
- **commit**：381c19e

## 2026-10-01 · codex

### docs(agents): define pre-merge branch synchronization

- **影响**：并行 worktree 的版本同步和合并前验收流程
- **关键文件**：`AGENTS.md`、`docs/agent-guide/workflow.md`、`docs/agent-guide/git-rules.md`、`docs/CONTEXT.md`、`docs/TASKS.md`、`docs/DECISIONS.md`
- **后续**：补充新 worktree 的依赖和构建输入初始化入口
- **commit**：918de3f

## 2026-10-01 · zcode

### fix(desktop): M4 桌面原生验收与三处缺陷修复

- **影响**：M4 三项验收全部完成。① 窗口规格落地：默认 1024×768 → 1120×760，新增最小 960×640（WM_GETMINMAXINFO 实测 1440×960 物理 @150% DPI = 960×640 逻辑，permonitorv2 生效）；② 最小窗口裁切修复：移除 `html/body/#app` 的 `min-width:960px`（视口比窗口逻辑尺寸小约 14px 系统边框，原规则致右侧内容被静默裁切）；③ 新增「同步模型」Provider 菜单入口（后端 /models 增量 Upsert 此前无 UI 触达路径）；④ 操作失败不再清空整个配置页（错误横幅内联、数据保留，仅无数据时独占页面）；⑤ 新增 `CPH_REMOTE_DEBUG_PORT` 门控 CDP 端口供自动化验收，默认关闭
- **关键文件**：`main.go`、`frontend/src/styles/base.css`、`frontend/src/pages/ConfigPage.vue`
- **验收证据**：CDP + 本地 Mock Responses API（127.0.0.1:8471）全链路走查：Provider 创建/持久化、测试连接（Bearer 凭据 → /models 200）、模型同步 2 条、Route 创建（级联选择、重启默认关）、Profile 创建自动选中、路由激活写入沙箱 CODEX_HOME 的 config.toml + 模型目录、备份恢复回到激活前状态、引用删除保护（Provider/Model 被引用时拒删且错误安全化）、Tab 顺序（侧栏→顶栏→主区）、Escape 关闭浮层、IME 组合输入提交；go test ./... 全过，dev 同步后重新构建冒烟通过；测试数据已清理（state.json 全空）
- **后续**：M4 出口标准达成；真实外部 Provider（CPA/OpenAI）请求验证仍属 M6（需用户凭据）
- **commit**：zcode/m4-acceptance 分支 d27e070、309f17f、49d6290（合并提交见下次回填）

## 2026-10-01 · codex

### docs(agents): switch to user-assigned GitHub PR integration

- **影响**：并行分支交付、公共文档审查和 `dev` 集成职责
- **关键文件**：`AGENTS.md`、`docs/agent-guide/git-rules.md`、`docs/agent-guide/workflow.md`、`docs/CONTEXT.md`、`docs/TASKS.md`、`docs/DECISIONS.md`
- **后续**：多个 agent 并行时由用户人工指定 reviewer/merger；单 agent 可在分支保护允许时自审并合并，但必须等待 CI；后续可配置分支保护和 Merge Queue
- **commit**：待下次文档提交回填

## 2026-10-01 · codex

### fix(ci): keep checkout line endings consistent on Windows

- **影响**：GitHub PR #1 的 Windows CI 在格式检查阶段报 17 个文件失败；本地以 `core.autocrlf=true` 检出可复现。新增 `.gitattributes`，让文本文件统一使用 LF，与 EditorConfig、Prettier 保持一致，二进制仍由 Git 自动识别。
- **关键文件**：`.gitattributes`
- **后续**：等待 PR 的完整 Windows CI 通过再合并
- **commit**：待下次文档提交回填

### fix(build): generate platform-specific icons

- **影响**：Windows CI 不再尝试打开 macOS 专用的 `darwin/icons.icns` 路径；macOS 仍保留 ICNS 和 Assets.car 生成参数。
- **关键文件**：`build/Taskfile.yml`
- **后续**：等待 PR 的完整 Windows CI 通过再合并
- **commit**：待下次文档提交回填

## 2026-10-01 · zcode

### perf(ui): split vendor chunks and name the profile selects

- **影响**：vite `manualChunks` 把依赖拆为 naive-ui（437.74 kB）/ framework（129.50 kB）/ vendor（28.93 kB）/ icons（5.90 kB），业务代码 index 从 638.10 kB 降至 36.97 kB，单 chunk 全部低于 500 kB 警告线；桌面端经 go:embed 打包总字节不变，收益是业务改动不再翻新库代码 hash。顶栏与设置页的配置档案 NSelect 在无档案时禁用（naive-ui 非过滤选择器空值时内部焦点 div 无可访问名称、attrs 无法透传到内部元素属上游限制），消除无名 tab stop；运行时审计三个页面可聚焦控件 0 个无名称
- **关键文件**：`frontend/vite.config.ts`、`frontend/src/components/DesktopShell.vue`、`frontend/src/pages/SettingsPage.vue`
- **后续**：模态/对话框右上角 X 关闭按钮（NCard/NDialog 内部）无可访问名称为上游限制，可考虑向 naive-ui 反馈；`wails/custom.js` 404 为运行时探测可选文件的上游行为，与本次分包无关
- **commit**：`65762fb`（PR #4）

## 2026-10-01 · zcode

### docs(release): M6 发布准备：README 刷新、发布清单与说明草稿（版本统一滞后，按用户决定）

- **影响**：发布前置文档准备（版本统一按用户决定滞后，当前保持 0.1.0 基线）：梳理出发布流水线版本校验共 6 处需保持一致（`build/config.yml`、`build/windows/info.json` 两处、`build/windows/nsis/wails_tools.nsh`，另加 `wails.exe.manifest` 与 MSIX 两清单为建议同步点），发布时按 RELEASE_CHECKLIST 逐项统一；README 当前状态与环境要求刷新（Go 1.25+/Node 22+）；新增 `scripts/ci/validate-release.ps1` 对 tag `v1.0.0` 的四点校验；README「当前状态」与开发环境要求刷新（Go 1.25+/Node 22+）；新增 `docs/ops/RELEASE_CHECKLIST.md`（production Environment/签名 Secrets/main 分支保护等用户侧一次性配置 + 版本自检 + 发布流程 + 演练说明）与 `docs/ops/RELEASE_NOTES.md`（V1 用户发布说明草稿）
- **关键文件**：`README.md`、`docs/ops/RELEASE_CHECKLIST.md`（新）、`docs/ops/RELEASE_NOTES.md`（新）、`docs/planning/ROADMAP.md`、`build/windows/wails.exe.manifest`、`build/windows/msix/app_manifest.xml`、`build/windows/msix/template.xml`、`docs/TASKS.md`
- **后续**：Tag 发布演练与版本号统一按用户决定滞后，待新增任务/目标落地后再启动；演练仍需用户先配置 production Environment 与签名 Secrets（见 RELEASE_CHECKLIST 第 1 节）；exe「属性 → 详细信息」版本显示为空为 winres/wails3 上游问题（版本资源字节已嵌入但 API 读不到），可向上游反馈
- **commit**：待 PR 合并后回填
