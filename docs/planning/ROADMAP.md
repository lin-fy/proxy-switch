# Codex Provider Hub：V1 目标、里程碑与执行状态

> 项目目录：`H:\code\proxy-switch`  
> 当前版本目标：Windows V1.0.0
> 最后更新：2026-10-02

本文档是目标模式的执行依据和项目状态真源。它回答四个问题：最终要交付什么、分几段完成、每段如何验收、现在马上做什么。

M4 桌面界面和 M5 Windows 安装包已完成；当前仅剩 M6 的真实 Provider/Codex 验收、GitHub 发布配置和正式 Tag 发布。具体边界与验收门槛见本文件第 1、3、4 节。

产品范围和架构原则见 [`../product.md`](../product.md)，执行约束见 [`../../CODEX.md`](../../CODEX.md)。如果实现过程发现方案需要变化，先更新本文档和相关方案文档，再继续编码。

## 1. 目标 Goal

交付一个 Windows 优先的 Wails 桌面工具，让用户通过 Provider、Model 和 Route 管理 Codex Desktop 的 Responses API 配置，并安全完成切换、测试、备份和恢复。

V1 首先服务 CPA-Manager-Plus / CLIProxyAPI，也支持 OpenAI 和其他兼容 Responses API 的 Provider。CPA 在产品中作为普通 Provider 管理，应用不实现代理转发。

### V1 Definition of Done

- [x] Windows 程序可以安装、启动、隐藏到托盘并退出。
- [x] Provider、Model、Route、Profile 的后端模型和持久化已实现。
- [x] Provider 可以通过界面新增、编辑、删除和测试连接。
- [x] Route 可以选择 Provider/Model、激活并设置默认路由。
- [x] Codex `config.toml` 读取、局部写入、备份和失败恢复已实现。
- [x] 多个 Codex Profile 和模型目录同步已实现。
- [x] 环境变量和 Windows Credential Manager 凭据引用已实现，普通状态文件不保存 API Key。
- [x] Responses API `/models` 和最小 `/responses` 测试已实现。
- [x] 自动重启默认关闭，用户勾选后激活路由可以重启 Codex。
- [x] 系统托盘常驻和开机启动开关已实现。
- [ ] CPA、OpenAI、一个自定义 Responses Provider 各完成一次真实请求验证。
- [x] Windows 安装包、恢复说明和发布文档完成。

“文件写入成功”不等于 V1 完成。最终必须通过这条真实链路：

```text
Proxy Switch
    ↓
选择 Route（CPA Provider + Model）
    ↓
备份并更新 Codex 配置
    ↓
启动 / 重启 Codex Desktop
    ↓
实际请求模型成功
```

## 2. 当前状态 Current

状态标记统一使用：`[ ]` 未开始，`[-]` 进行中，`[x]` 已完成，`[!]` 阻塞。

| 能力 | 状态 | 当前事实与验收证据 |
| --- | --- | --- |
| 项目方案和边界 | [x] | V1 锁定 Windows、Wails 3、Go、Codex 单平台适配器和 Responses API。 |
| 领域与应用层 | [x] | Provider / Model / Route / Profile、Repository、Service 和 Activator 已存在。 |
| Codex 配置适配 | [x] | 活动配置、Profile 配置、模型目录、备份、恢复和回滚逻辑已实现。 |
| Provider 连通性 | [x] | 支持 `/models` 和最小 `/responses` 请求，凭据只使用引用。 |
| Wails 桌面壳 | [x] | Wails 3、窗口关闭隐藏、托盘显示/隐藏/退出已接入。 |
| 开机启动 | [x] | Windows `HKCU\\...\\Run` 开关已接入。 |
| 当前工作区构建 | [x] | `gofmt`、`go test ./...`、`go vet ./...`、前端格式检查、lint、生产构建和 Wails Windows amd64 构建均已复验；统一验证脚本还可生成 portable ZIP。2026-10-01 前端重构后桌面 exe 重新构建（约 11.4 MB）并完成启动冒烟：WebView2 初始化、窗口渲染真实后端空态、生产包确认不含浏览器预览演示数据。 |
| Vue UI 迁移 | [x] | Vue 3、Pinia、Naive UI、Tailwind v4、配置/Profiles/Settings 页面和 Wails API 适配层已接入；错误安全映射、错误摘要聚焦、激活消息保留已验证；2026-10-01 桌面原生验收完成：窗口 1120×760/最小 960×640 实测、150% DPI 渲染、浮层/Tab/Escape/IME 验证、Mock Responses 全链路走查通过（详见 CHANGELOG）。 |
| Windows 安装包 | [x] | 用户级 NSIS 安装包已生成，并完成安装、启动、升级/回滚和卸载验证；MSIX 作为可选路径保留。 |
| 真实端到端验收 | [ ] | 仍需可用的 CPA、OpenAI 和自定义 Provider 以及 Codex Desktop 实际请求记录。 |

## 3. 里程碑 Milestones

> M0~M2 已完成，归档于 [`archive/M0-M2.md`](archive/M0-M2.md)。以下仅保留活跃里程碑。

### M3 — Responses Provider 与 CPA 适配

状态：`[x]`（真实服务验收待完成）

- [x] Provider 统一使用 Responses 协议。
- [x] 支持环境变量和 `credential:<target>` 凭据引用。
- [x] 连接测试调用 `/models`，模型测试调用最小 `/responses`。
- [ ] 用 CPA-Manager-Plus 的真实地址、凭据和两个模型完成端到端验证。

出口标准：CPA、OpenAI 和一个自定义 Responses Provider 都能添加、测试、激活并完成实际请求。

### M4 — Vue 桌面界面

状态：`[x]`（前端基线、浏览器走查与桌面原生验收全部完成）

- [x] Vue 入口、Pinia workspace store、Wails API 适配层和 Naive UI 主题。
- [x] 配置页的 Provider / Model / Route 列表、编辑和激活流程已接入。
- [x] Profiles 页面：档案切换、创建、删除和恢复备份的前端基线已接入。
- [x] Settings 页面：Codex 状态、启动、自动启动和恢复入口的前端基线已接入。
- [x] 加载、保存、测试、激活、错误和恢复状态的统一反馈已接入。
- [x] 收敛令牌/字体单一来源（`theme.ts` 从 `tokens.css` 计算值生成主题，系统字体栈），补齐 NForm 字段 path 和错误摘要焦点。
- [x] 创建 Route 时保留 `restart_on_activate`，并让激活 `recovered` 只来自明确后端证据（失败时保持 unknown 并提供恢复入口）。
- [x] 将原始后端错误转换为安全提示（分类映射，浏览器 401 用例验证无原始异常泄漏）。
- [x] 修正文字对比度（muted 调整为 `#8ca3a1`，所有表面 ≥ 4.5:1，计算证据）并按 960×640 / 1120×760 截图走查修复行列表溢出。
- [x] 完成窗口尺寸、DPI、浮层、键盘/IME 和真实后端流程验收（2026-10-01 桌面原生验收：150% DPI 下默认 1680×1140 物理 = 1120×760 逻辑、最小跟踪尺寸 1440×960 物理 = 960×640 逻辑实测通过；浮层/Tab/Escape/IME 组合输入验证；本地 Mock Responses API 全链路走查通过，发现并修复最小窗口裁切、模型同步无 UI 入口、错误态清页三处缺陷，见 CHANGELOG）。

出口标准：在 960 × 640 窗口中，Provider、Model、Route、Profile 和 Settings 全流程可操作，前端构建通过。

### M5 — Windows 构建与安装

状态：`[x]`

- [x] Windows amd64 可执行文件构建和启动冒烟验证。
- [x] 无管理员权限的用户级 ZIP 安装和卸载验证。
- [x] 补齐 NSIS 用户级安装包，并验证安装、启动和卸载。
- [x] 验证安装包升级/回滚。
- [x] 发布 [`../ops/BUILD_WINDOWS.md`](../ops/BUILD_WINDOWS.md) 的最终步骤和故障恢复说明。

出口标准：干净 Windows 环境可以按文档安装并启动 V1。

### M6 — V1 验收与发布

状态：`[ ]`

- [x] 运行 Go 测试、`go vet`、前端类型检查和生产构建（2026-10-02：`task ci`/`scripts/ci/verify.ps1 -BuildWindows` 通过；验证器会自动搜索 `GOPATH\bin` 中的 Wails CLI，并生成 portable ZIP）。
- [ ] 完成 CPA → Proxy Switch → Codex Desktop → 模型请求的真实链路。
- [ ] 完成 OpenAI、CPA、自定义 Provider 和两个 Codex Profile 的验收记录。
- [x] 清理文档中的临时状态，刷新 README 和发布说明（2026-10-01：README 状态与环境要求刷新；新增 docs/ops/RELEASE_CHECKLIST.md 发布演练清单——含发布时需统一的 6 处版本位置——与 docs/ops/RELEASE_NOTES.md 发布说明草稿；版本号统一与 Tag 演练按用户决定滞后）。
- [ ] 将验收后的 `dev` 合并到 `main`，关闭 V1 goal。

### CI/CD 实施状态

- [x] `.github/workflows/ci.yml`：Windows runner、Go/前端检查、Wails 构建和 7 天构建产物。
- [x] `.github/workflows/release.yml`：Tag 校验、版本元数据校验、NSIS/portable 构建、签名、SHA256 和 GitHub Release。
- [x] `.agents/skills/review-pr`：项目级 Codex Review Skill 和检查清单。
- [ ] 在 GitHub 配置 `production` Environment、签名证书 Secrets 和 `main` 分支保护。
- [ ] 在 GitHub Actions 上完成一次真实 Tag 发布演练。

## 4. 计划 Plan

### 当前执行顺序

```text
M6-A：对齐状态、目标和验收证据
  ↓
M6-B：本地确定性验证与 Windows 产物检查（已完成）
  ↓
M6-D：GitHub production Environment、签名 Secrets、main 保护与 Tag 演练（当前先行）
  ↓
M6-C：CPA / OpenAI / 自定义 Provider 真实端到端验收（按用户安排延后）
  ↓
M6-E：确认 1.0.0、合并 main、发布签名安装包并关闭 V1 goal
```

### M6 验收证据与门槛

| 阶段 | 必须证明 | 当前状态 | 证据或依赖 |
| --- | --- | --- | --- |
| M6-A | 文档状态与代码、已合并验收结果一致 | [x] | 本文件、`docs/CONTEXT.md`、`docs/TASKS.md` |
| M6-B | Go/前端/Wails/portable ZIP 验证通过，且本地找不到 Wails 时给出明确预检 | [x] | `scripts/ci/verify.ps1`、`docs/ops/CI_CD.md`、`docs/ops/BUILD_WINDOWS.md` |
| M6-C | 三类 Provider 各完成真实请求；CPA 验证两个模型；两个 Profile 可切换、激活、恢复 | [ ] | 需要用户提供 Provider 凭据与 Codex Desktop 环境 |
| M6-D | GitHub Environment、签名 Secrets、`main` 分支保护和一次 Tag 发布演练完成 | [ ] | 需要仓库管理员操作，流程见 `docs/ops/RELEASE_CHECKLIST.md` |
| M6-E | 版本资源、Tag、Release 产物和安装验证全部一致 | [ ] | 依赖 M6-C、M6-D；目标 Tag 为 `v1.0.0` |

### 执行原则

1. 先完成 Provider → Codex 配置 → Route 激活 → 实际请求的主链路，再做视觉细节。
2. 任何 Codex 配置写入都遵循“备份 → 写入 → 验证 → 失败恢复”。
3. Provider 只保存凭据引用；日志、错误和界面不显示完整密钥、请求头或响应体。
4. 先修复当前构建阻塞，再扩展页面；不在 V1 引入第二个平台、协议转换、云同步或账户池。
5. 每个里程碑完成后先更新本文档，再提交代码；未通过验证的能力不能标记为完成。

### 可暂停并请求用户介入的情况

- 需要用户提供真实 Provider 地址、凭据或 Codex Desktop 环境。
- 需要删除不可恢复的数据。
- 外部服务不可用，无法完成真实请求验证。
- 方案会改变 V1 产品边界或数据格式。

除此之外，目标模式应继续拆分任务、实现、验证、修复并更新本文档。

## 5. 当前下一步 Next

只保留进入实现阶段后马上要做的两项：

1. 在 GitHub 配置 `production` Environment、签名 Secrets 和 `main` 分支保护，完成一次 Tag 发布演练（**需要用户操作**）。
2. 完成 CPA、OpenAI、自定义 Provider 的真实端到端验收（**需要用户提供凭据**）。

## 6. 技术决策记录

- 使用 Wails 3，不迁移到 Tauri、Electron 或远程网页。
- 使用 Vue 3 + TypeScript + Pinia；Tailwind CSS v4 管布局与令牌，Naive UI 管控件；生成的 Wails 绑定不得手改。
- Provider 是数据，不为 CPA、OpenAI 等供应商分别创建工厂。
- V1 只有 Codex 平台适配器，其他平台留到新的 goal。
- Codex Desktop 若不支持 `--profile`，继续使用备份后切换活动配置文件的兼容方案。
- 自动更新属于 V1.1，不阻塞 V1 安装和验收。
