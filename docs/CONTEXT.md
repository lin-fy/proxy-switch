# 项目当前状态快照

> agent 协作的"单一事实源"。开工读，收工更新。只保留当前状态。
> 详细规则见 [`agent-guide/workflow.md`](agent-guide/workflow.md)。

最后更新：2026-10-02 · codex（Codex 配置保真/回滚加固）

## 一句话现状

V1 后端核心 + 桌面壳 + Vue 前端已完成并通过桌面原生验收（M4 全部完成，含窗口/DPI/浮层/键盘/IME 与 Mock 全链路走查）；NSIS 安装包全流程验收与 `BUILD_WINDOWS.md` 最终文档定稿完成，M5 全部完成。M6-A/B 的目标、证据矩阵和本地确定性验证已完成；当前先推进 M6-D，M6-C 真实 Provider 验收延后，M6-E 正式发布仍需两者完成。

## 下一步

1. 确认签名凭据可用性；GitHub 配置已完成，当前证书仍为自签名测试证书；按用户决定，Tag 等代码进入 `main` 后再创建
2. 提供 CPA、OpenAI、自定义 Provider 凭据和 Codex Desktop 环境，完成真实端到端验收（**按用户安排稍后进行**）
3. M6-C/D 通过后统一切换 `1.0.0`、合并 `main` 并发布正式安装包
4. 前端开发前核对最新 Wails v3 beta；当前基线已统一到已核对的 `v3.0.0-beta.27`

统一验证入口已补齐：`task check:frontend`、`task check:backend`、`task ci`；GitHub Actions 与本地入口均复用 `scripts/ci/verify.ps1`，兼容 Windows PowerShell 5.1，并在 Windows 构建时自动搜索 `GOPATH\bin` 的 Wails CLI。2026-10-02 本地完整验证通过并生成 portable ZIP。PR #10 已合并到 `dev`（`38354c8`），包含文档 PR 的 `verify` 触发修复；PR #11 已合并到 `dev`（`0b33c73`），统一 UI 版本来源并改善 Windows 依赖安装；合并后 Windows CI 通过（运行 `36967821792`）。

PR #9 已在独立工作树同步 `dev` 并保留双方文档记录；补齐 Provider/Model 删除引用保护、并发写入和损坏 JSON 的回归测试。本地完整 Windows 验证通过；合并仍须以最新 PR head 的 GitHub `verify` 通过为准。

2026-10-02 Codex parity 加固已完成：TOML 局部补丁保留未知配置并对不安全表示 fail-closed，模型目录按显式所有权标记更新，备份与激活失败恢复此前工件状态，并拒绝目标路径碰撞。Go 1.25.14 下 `go test ./internal/...` 通过；根包 `go test ./...` 仍需先生成 `frontend/dist` embed 产物。

Provider `/models` 现要求响应包含 `data` 数组，异常响应会 fail-closed，避免同步时清空本地模型。第一阶段剩余主要为真实 Provider/Codex Desktop 验收，以及 CC Switch parity 的用户级功能：当前配置识别/导入、官方登录保护与预设入口、Auth mode UI；这些需要单独设计和真实 Codex auth 环境，不在本轮盲目扩展。

Phase 2 已落地只读 `InspectCodexConfig` 与显式 Profile 导入：配置档案页可将当前 `config.toml` 原子复制到隔离的 `profiles/<id>/config.toml`，保留未知配置、MCP 与 OAuth 语义；不会读取、复制或返回 auth secret，也不覆盖已有导入目标。Provider/Model 记录不自动推断，真实 Codex Desktop/auth 环境仍需端到端验收。

Provider 当前配置页已补充只读 preset/auth-mode 元数据；`auth_ref` 只允许环境变量名或 `credential:<target>` 引用，派生字段不落盘，也不接受原始 API Key。官方 OAuth 登录仍需真实 Codex 环境验收。

Windows 原生验收命令与逐项清单见 `docs/ops/CODEX_NATIVE_ACCEPTANCE.md`，覆盖 Provider/AuthRef、Profile 导入、备份恢复、DPI、托盘和真实 Responses 请求。云端 Linux 可完成 internal/前端确定性检查，但根包 Wails 测试需要 GTK4/WebKitGTK，不能替代 Windows 原生验收。

Provider 添加页已接入只读预设选择：公开 Responses 预设只填充连接信息与环境变量引用占位；官方 OAuth 预设不可直接创建，改为引导当前 Codex Profile 导入，避免伪造或复制认证内容。Provider health/status 与 failover 仍是后续高价值切片。

Provider 健康与 failover 顺序最小切片已完成：健康检查仅在用户菜单点击时调用，返回内存状态/超时/耗时，不自动联网；Route `priority` 只保存和展示顺序，当前不自动 failover。

Settings/Codex 已补 Auth Center 只读状态：本机 `auth.json` 仅被分类为 OAuth、API Key 引用、已配置、缺失或格式错误并显示是否存在凭据；token、key 名称和文件内容永不返回，登录/账号切换仍待真实环境与独立设计。

Settings 数据区已补显式安全 JSON 导出：Provider/Model/Route/Profile 与 schema/export 时间可下载；Provider headers/query 参数和不符合引用格式的历史 `auth_ref` 会被剔除，导出不读取 `auth.json`。导入/合并需要独立冲突与回滚设计，Session/Skills 仍缺后端数据源。

Provider 卡片现保留最近 5 次显式健康检查的内存历史并在 Tooltip 显示；不自动联网或持久化，自动 failover/监控仍待单独策略。

Settings 数据区已补安全导入预览与确认：严格接受导出 `schema_version: 1`，拒绝未知/secret 字段并返回冲突/缺失引用；无冲突时只新增资源并以单次 state.json 事务提交，冲突或转换错误全量拒绝，不覆盖已有状态。

Store 导入事务在最终写入前再次校验启用 Model、Provider/Model/Route 引用及 Profile 路径冲突，避免预览后状态变化或绕过 Wails 造成部分写入。

Provider/Model/Route 的 Create 已补 ID 冲突保护，重复新建不会覆盖已有记录；Save 仍用于显式编辑，导入预览冲突列表尚不自动合并。

Route Delete 现检查 Profile.RouteID 引用，引用中的 Route 不可删除，避免配置档案孤儿引用；RouteService 已由 main 注入 Profile repository。

Profile Save 现检查非空 RouteID 是否存在，拒绝通过编辑/Wails API 写入孤儿引用；ProfileService 已由 main 注入 Route repository。

Model Create/Save 现检查 Provider 是否存在，拒绝孤儿 Model；ModelService 已由 main 注入 Provider repository，Route 层仍做 Provider/Model 双重引用校验。

配置页编辑已有 Model 时锁定 Provider，避免 UI 改变 ProviderID 造成旧 Route 孤儿引用；新建 Model 仍可选 Provider。

Profile Save 现拒绝不同档案共享非空 ConfigPath/ModelCatalogPath，避免激活互相覆盖；空路径默认语义保留。

## 当前卡点

- 真实 CPA / OpenAI / 自定义 Provider 凭据（**需用户提供**）
- 签名与 Tag 演练尚未完成：两个 Secret 名称已确认存在，证书为自签名测试证书；现有 Release 工作流要求签名 `Valid` 且带时间戳，凭据可用性和信任尚未验证

## 待协调事项

- 正式 `1.0.0` 签名需要受 Windows 信任链认可、包含私钥且可导出为 `.pfx` 的 Authenticode 证书；自签名证书不能作为公开发布替代品。
- PR #11 已合并且合并后 Windows CI 通过；其后续收尾记录已由 PR #14 更新。
- 已新增 Wails v3 beta 前端版本规则，并将 Go 模块、CLI、runtime、CI 和构建文档统一到 `v3.0.0-beta.27`；升级后的完整 Windows CI 已通过（运行 `36971499051`）。

### 已裁决（归档）

- 2026-10-02：用户选择将仓库设为公开；GitHub API 已确认公开、`main` 保护和 `production` 审批生效（ADR-0012），原方案限制已解除
- 2026-10-01：TASKS.md codex 代记条目归属问题 → 删除，由 zcode 自己条目承担
- 2026-10-01：CI 与开发 agent、模型解耦，AI Review/失败诊断仅作为可选分析层（ADR-0011）

## 长期注意事项

- 当前已切换 GitHub PR 流程：任务 agent 在自己的分支同步 `dev`、解决冲突、验证并创建 PR；多个 agent 并行时由用户指定 reviewer/merger，只有一个活跃 agent 时允许自审并在分支保护允许时合并，但必须等待 CI 通过。不得覆盖他人未提交改动。
- `sources/` 下文件只读
- 不得提交密钥、临时文件、构建产物、CodeGraph 缓存
- 瞬时工作区状态直接 `git status`，不在本文档维护

---

**详细里程碑与验收状态**：[`planning/ROADMAP.md`](planning/ROADMAP.md)
**任务认领看板**：[`TASKS.md`](TASKS.md)
**变更历史**：`git log` + [`CHANGELOG.md`](CHANGELOG.md)
