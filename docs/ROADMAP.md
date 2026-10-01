# Codex Provider Hub：V1 目标、里程碑与执行状态

> 项目目录：`H:\code\proxy-switch`  
> 当前版本目标：Windows V1  
> 最后更新：2026-10-01

本文档是目标模式的执行依据和项目状态真源。它回答四个问题：最终要交付什么、分几段完成、每段如何验收、现在马上做什么。

产品范围和架构原则见 [`PRODUCT_PLAN.md`](../PRODUCT_PLAN.md)，执行约束见 [`PROJECT_INSTRUCTIONS.md`](../PROJECT_INSTRUCTIONS.md)。如果实现过程发现方案需要变化，先更新本文档和相关方案文档，再继续编码。

## 1. 目标 Goal

交付一个 Windows 优先的 Wails 桌面工具，让用户通过 Provider、Model 和 Route 管理 Codex Desktop 的 Responses API 配置，并安全完成切换、测试、备份和恢复。

V1 首先服务 CPA-Manager-Plus / CLIProxyAPI，也支持 OpenAI 和其他兼容 Responses API 的 Provider。CPA 在产品中作为普通 Provider 管理，应用不实现代理转发。

### V1 Definition of Done

- [ ] Windows 程序可以安装、启动、隐藏到托盘并退出。
- [x] Provider、Model、Route、Profile 的后端模型和持久化已实现。
- [ ] Provider 可以通过界面新增、编辑、删除和测试连接。
- [ ] Route 可以选择 Provider/Model、激活并设置默认路由。
- [x] Codex `config.toml` 读取、局部写入、备份和失败恢复已实现。
- [x] 多个 Codex Profile 和模型目录同步已实现。
- [x] 环境变量和 Windows Credential Manager 凭据引用已实现，普通状态文件不保存 API Key。
- [x] Responses API `/models` 和最小 `/responses` 测试已实现。
- [ ] 自动重启默认关闭，用户勾选后激活路由可以重启 Codex。
- [x] 系统托盘常驻和开机启动开关已实现。
- [ ] CPA、OpenAI、一个自定义 Responses Provider 各完成一次真实请求验证。
- [ ] Windows 安装包、恢复说明和发布文档完成。

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
| 当前工作区构建 | [!] | Go 测试和 vet 仍被 `PlatformAdapter.Launch/Restart` 签名与调用方不一致阻塞；`frontend` 的 `vue-tsc --noEmit` 与 `npm run build` 已通过（最近构建约 636 kB raw / 192 kB gzip）；Wails Windows 构建尚未复验。 |
| Vue UI 迁移 | [-] | Vue 3、Pinia、Naive UI、Tailwind v4、配置/Profiles/Settings 页面和 Wails API 适配层已有可构建基线；浏览器窗口和导航冒烟检查已完成，Wails、DPI、焦点、IME、对比度和完整流程验收仍待完成。 |
| Windows 安装包 | [-] | 用户级 ZIP 安装/卸载已验证；NSIS 和 MSIX 仍待工具链补齐。 |
| 真实端到端验收 | [ ] | 仍需可用的 CPA、OpenAI 和自定义 Provider 以及 Codex Desktop 实际请求记录。 |

## 3. 里程碑 Milestones

### M0 — 方案与基线

状态：`[x]`

- [x] 明确产品边界、V1 范围和不做事项。
- [x] 锁定 Wails 3 `v3.0.0-beta.26`、Go、Vue 3、TypeScript、Naive UI、Pinia。
- [x] 确定 DDD 分层、PlatformAdapter 和 Provider/Model/Route/Profile 模型。
- [x] 建立 Git `dev` 分支和项目执行文档。

出口标准：方案可以直接拆成实现任务，且不依赖聊天上下文才能恢复进度。

### M1 — 核心后端与桌面壳

状态：`[!]`（实现已在工作区，当前集成验证被构建回归阻塞）

- [x] 本地 JSON 状态存储和四类 Repository。
- [x] Wails 服务绑定、DTO 和前端生成绑定。
- [x] Windows 窗口、托盘、关闭窗口隐藏和开机启动。
- [x] Provider、Model、Route、Profile 的基本 CRUD。

出口标准：程序可以启动，Wails API 可以读写本地状态。

### M2 — Codex 配置与安全切换

状态：`[x]`

- [x] 定位用户级 Codex 配置目录和活动 `config.toml`。
- [x] 通过 Codex 适配器局部写入 Provider、Model、Profile 和模型目录。
- [x] 写入前备份，多文件写入失败回滚，提供手动恢复入口。
- [x] 激活成功后才更新默认 Route。

出口标准：切换失败不会留下半写入配置，恢复后可以继续使用原配置。

### M3 — Responses Provider 与 CPA 适配

状态：`[x]`（真实服务验收待完成）

- [x] Provider 统一使用 Responses 协议。
- [x] 支持环境变量和 `credential:<target>` 凭据引用。
- [x] 连接测试调用 `/models`，模型测试调用最小 `/responses`。
- [ ] 用 CPA-Manager-Plus 的真实地址、凭据和两个模型完成端到端验证。

出口标准：CPA、OpenAI 和一个自定义 Responses Provider 都能添加、测试、激活并完成实际请求。

### M4 — Vue 桌面界面

状态：`[-]`（前端基线已实现，桌面验收待完成）

- [x] Vue 入口、Pinia workspace store、Wails API 适配层和 Naive UI 主题。
- [x] 配置页的 Provider / Model / Route 列表、编辑和激活流程已接入。
- [x] Profiles 页面：档案切换、创建、删除和恢复备份的前端基线已接入。
- [x] Settings 页面：Codex 状态、启动、自动启动和恢复入口的前端基线已接入。
- [x] 加载、保存、测试、激活、错误和恢复状态的统一反馈已接入。
- [ ] 完成 Wails Windows 构建、窗口尺寸、DPI、浮层、焦点、键盘/IME、对比度和真实后端流程验收。

出口标准：在 960 × 640 窗口中，Provider、Model、Route、Profile 和 Settings 全流程可操作，前端构建通过。

### M5 — Windows 构建与安装

状态：`[-]`

- [x] Windows amd64 可执行文件构建和启动冒烟验证。
- [x] 无管理员权限的用户级 ZIP 安装和卸载验证。
- [ ] 补齐 NSIS 或 MSIX 安装包，并验证安装、启动、升级/回滚和卸载。
- [ ] 发布 `BUILD_WINDOWS.md` 的最终步骤和故障恢复说明。

出口标准：干净 Windows 环境可以按文档安装并启动 V1。

### M6 — V1 验收与发布

状态：`[ ]`

- [ ] 运行 Go 测试、`go vet`、前端类型检查和生产构建。
- [ ] 完成 CPA → Proxy Switch → Codex Desktop → 模型请求的真实链路。
- [ ] 完成 OpenAI、CPA、自定义 Provider 和两个 Codex Profile 的验收记录。
- [ ] 清理文档中的临时状态，更新版本号、README 和发布说明。
- [ ] 将验收后的 `dev` 合并到 `main`，关闭 V1 goal。

## 4. 计划 Plan

### 当前执行顺序

```text
修复 Vue 构建
  ↓
完成 Profiles / Settings 页面
  ↓
跑 Go + 前端 + Wails 验证
  ↓
补齐安装包
  ↓
真实 Provider / Codex 端到端验收
  ↓
更新文档并发布 V1
```

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

只保留进入实现阶段后马上要做的三项：

1. 同步 `PlatformAdapter.Launch/Restart` 的调用方和实现，让 `go test ./...` 与 `go vet ./...` 通过。
2. 重新运行 Wails Windows 构建，并按 960 × 640、1120 × 760、125% / 150% DPI 验证壳层、浮层和主区滚动。
3. 完成表单焦点、键盘/IME、对比度、reduced-motion 和 Provider → Codex 的真实流程验收，再处理 NSIS/MSIX 安装包。

## 6. 技术决策记录

- 使用 Wails 3，不迁移到 Tauri、Electron 或远程网页。
- 使用 Vue 3 + TypeScript + Naive UI + Pinia；生成的 Wails 绑定不得手改。
- Provider 是数据，不为 CPA、OpenAI 等供应商分别创建工厂。
- V1 只有 Codex 平台适配器，其他平台留到新的 goal。
- Codex Desktop 若不支持 `--profile`，继续使用备份后切换活动配置文件的兼容方案。
- 自动更新属于 V1.1，不阻塞 V1 安装和验收。
