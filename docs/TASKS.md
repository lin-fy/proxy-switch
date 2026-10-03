# 任务认领表

> 多 agent 协作的任务看板。规则见 [`../AGENTS.md`](../AGENTS.md) 第三节。
>
> 状态标记：`[ ]` 未认领 · `[- by <agent> <日期>]` 进行中 · `[x by <agent> <完成日期>]` 已完成 · `[!]` 阻塞

最后更新：2026-10-02

## 进行中

- [x by zcode 2026-10-02] 产品命名统一为 Proxy Switch（PR #23）：安装/exe/界面/发布产物全部改名，重发 v0.1.0（产物 `ProxySwitch-0.1.0-*`，SHA256 与属性页核验通过）；Go module 路径与 CHANGELOG 历史保持不变

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
- [x by zcode 2026-10-02] M6 发布准备：README 刷新、发布说明草稿与发布演练清单（PR #6/#14 合并）
- [x by codex 2026-10-02] 按用户授权将仓库公开；配置 `production` Environment、两个测试签名 Secrets、发布审批人和 `main` 分支保护（凭据可用性与签名信任尚待演练验证）
- [x by zcode 2026-10-02] Tag 发布演练：main 同步 + v0.1.0 tag + 无签名发布（PR #15~#21）；Release 已发布（3 产物 SHA256 校验一致、属性页版本 0.1.0 正确）；顺带验证测试签名证书可用但信任链不通、新增 unsigned 手动开关
- [ ] 用真实 CPA、OpenAI、自定义 Provider 完成端到端验收（**需要用户提供凭据**）

### 持续性工作

- [ ] 每次代码改动后同步更新 `docs/CONTEXT.md` 和 `docs/CHANGELOG.md`
- [ ] 重要技术决策追加到 `docs/DECISIONS.md`

## 已完成

- [x by codex 2026-10-02] Codex 配置保真与原子恢复加固：局部 TOML 补丁保留未知配置并对不安全多行/引号键表示安全失败；模型目录仅按显式 `proxy_switch_managed` 所有权更新，保护用户预设；备份标记、目标碰撞和异常回滚补齐回归测试；Go 1.25.14 下 `go test ./internal/...` 通过，根包完整测试仍需前端 `frontend/dist` embed 产物。
- [x by codex 2026-10-02] Provider `/models` 响应契约加固：缺失 `data` 数组时连接/同步失败，不再将异常响应误判为空模型列表；Go 1.25.14 下 provider 与 internal 测试通过。
- [x by codex 2026-10-02] Phase 2 Codex 配置只读探测与 Profile 导入：Wails `InspectCodexConfig` 返回 config.toml/auth.json 状态、当前 Provider/Model 与可导入原因；配置档案页显式导入当前 `config.toml` 到隔离 Profile，保留未知配置/MCP/OAuth 语义，绝不复制或修改 `auth.json`，并覆盖 malformed/重复目标/危险 ID 回归测试。
- [x by codex 2026-10-02] Provider preset/auth-mode 安全显示：Provider DTO 与当前配置页展示只读 preset 元数据、认证模式和凭据引用是否配置；AuthRef 仅接受环境变量或 `credential:<target>` 引用，状态文件不保存原始密钥，补领域/Wails 回归测试。
- [x by codex 2026-10-02] Windows 原生验收收口：新增可执行的 Codex Desktop/Provider/AuthRef/Profile/备份恢复/DPI/托盘验收清单与命令，记录 Linux 云端缺少 GTK4/WebKitGTK 导致根包 Wails 测试受限；补 Wails 原始凭据拒绝契约测试。
- [x by codex 2026-10-02] Provider 预设快速添加：新增只读 `ListProviderPresets` 契约与配置页预设选择器；预设仅填充公开连接信息和凭据引用占位，官方 OAuth 引导 Profile 导入，不复制或写入 secret；补 Wails/前端构建验证。
- [x by codex 2026-10-02] Provider 健康与 failover 顺序最小切片：健康检查仅由用户显式触发并返回安全状态/超时/耗时 DTO；Route 新增非负 priority，配置页按 priority 展示/编辑顺序，暂不自动切换或发送后台请求。
- [x by codex 2026-10-02] Auth Center 只读状态：Settings/Codex 展示 auth.json 的派生认证类型与凭据存在性，覆盖 OAuth/API key/缺失/格式错误，不返回或写入认证内容。
- [x by codex 2026-10-02] Provider/Profile 安全导出：Settings 显式导出 schema-versioned JSON，包含四类资源与安全 auth_ref 引用，移除 headers/query 参数及非法历史凭据，补 Wails secret-leak 回归测试；导入/合并与 Session/Skills 页面待后续数据契约。
- [x by codex 2026-10-02] Provider 健康历史：内存保留最近 5 次用户显式健康检查并在 Provider 卡片 Tooltip 展示时间/状态/耗时；不自动联网、不持久化、不执行自动 failover。
- [x by codex 2026-10-02] 安全导入预览与确认：严格解析导出 `schema_version: 1` JSON，拒绝 headers/query/auth/token 等未知或 secret 字段，报告冲突/缺失引用；无冲突时用户确认后以单次原子事务只新增资源，冲突全量拒绝。
- [x by codex 2026-10-02] 导入原子引用加固：Store 写入前再次拒绝停用 Model 路由、共享 Profile 路径等绕过预览/竞态情况，失败不写入任何资源并补回归测试。
- [x by codex 2026-10-02] 资源创建冲突保护：Provider/Model/Route Create 在 Save 前拒绝已存在 ID，避免新建静默覆盖；编辑继续使用 Save，补 application focused tests。
- [x by codex 2026-10-02] Route 引用完整性：Route Delete 检查 Profile.RouteID 引用并阻止孤儿档案，main wiring 注入 Profile repository，补回归测试。
- [x by codex 2026-10-02] Profile 引用校验：Profile Save 对非空 RouteID 检查 Route 存在性，拒绝孤儿引用并补 application focused tests；main wiring 传入 Route repository。
- [x by codex 2026-10-02] Model 引用校验：Model Create/Save 注入 Provider repository 并拒绝不存在 Provider，防止编辑/Wails API 写入孤儿 Model；补 application tests。
- [x by codex 2026-10-02] Model 编辑引用保护：编辑已有 Model 时锁定 Provider 选择，避免被 Route 引用的 Model 被 UI 移到另一 Provider；新建流程仍支持选择 Provider。
- [x by codex 2026-10-02] Profile 路径冲突保护：Profile Save 拒绝不同档案共享非空 ConfigPath/ModelCatalogPath，避免激活互相覆盖；补 application focused tests。

- [x by zcode 2026-10-02] 版本展示同源与构建卫生（PR #11 尾栏注入 0.1.0、install-or-skip）+ 收尾（PR #14 exe 版本元数据 info.json 0409 + validator、NDialog aria 兜底）；上游反馈 wails#6210 / naive-ui#8231
- [x by codex 2026-10-02] 补齐 Provider/Model 删除引用保护、`store.go` 并发写入与损坏 JSON 安全失败的针对性测试；PR #9 同步 dev 后完整 Windows 验证通过；去锁变异测试检验并发回归用例有效，本机无 C 编译器，未运行 race detector

- [x by codex 2026-10-02] 新增 Wails v3 beta 前端开发规则并统一到 `v3.0.0-beta.27`：前端改动前核对最新具体版本，统一 Go 模块、CLI、runtime、CI 和文档，重新生成绑定并通过 `task ci`

- [x by codex 2026-10-02] 按用户授权合并 PR #11 到 `dev`：统一 UI 版本来源并修复 Windows 依赖安装路径；合并提交 `0b33c73`，合并后 Windows CI 通过（运行 `36967821792`）

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
