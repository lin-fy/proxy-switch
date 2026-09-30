# Codex Provider Hub：V1 项目指令

## 当前目标

当前阶段只有一个长期 goal：完成 Codex Provider Hub Windows V1。

每轮对话、开发、验证、修复和文档更新都属于同一个 V1 goal，不拆成独立用户任务。

V1 验收完成后关闭当前 goal。V1.1、V2 或其他平台支持必须创建新的独立 goal。

## 文档关系

- `PRODUCT_PLAN.md`：产品定位、范围和架构原则。
- `docs/ROADMAP.md`：V1 的统一目标、里程碑、验收状态和当前下一步；目标模式每轮都必须读取并维护。
- `docs/TECH_STACK.md`、`docs/UI_ARCHITECTURE_SPEC.md`：前端技术和界面实现规范。
- `docs/BUILD_WINDOWS.md`：Windows 构建、安装和恢复操作说明。

当聊天上下文与仓库文档不一致时，以代码和 `docs/ROADMAP.md` 的最新验证记录为准；发现差异先更新文档。

## 首轮初始化

1. 初始化 Git，创建合理的 `.gitignore` 和初始基线提交。
2. 初始化或启用 CodeGraph；理解代码时优先使用 CodeGraph。
3. CodeGraph 不可用时再使用普通文件搜索。
4. `sources/` 中的同步文件只读，不得修改、移动或删除。

## 模型分工

| 任务 | 模型 |
| --- | --- |
| 需求拆解、架构设计、跨模块方案 | `gpt-6.1-sol` + `xhigh` |
| 涉及公共 API、数据库迁移、安全、并发、核心算法 | `gpt-6.1-sol` + `xhigh` |
| 范围明确的功能实现 | `gpt-5.6-luna` + `max` |
| 单元测试、集成测试、调查、文档 | `gpt-5.6-luna` + `max` |
| 代码审查、合并冲突、失败测试分析、最终验收 | `gpt-6.1-sol` + `xhigh` |

- 子任务都属于当前 V1 goal，不创建新的用户级项目任务。
- 指定模型不可用时必须说明，不静默替换。

## V1 范围

- Windows only。
- Wails 3（当前锁定 `v3.0.0-beta.26`）+ Go。
- 只实现 Codex Desktop 平台适配器。
- Provider 只配置一次，通过 Route 选择 Platform、Provider、Model。
- CPA 作为普通 Provider，不实现独立的 CPA 产品逻辑。
- 首版只接入 Responses API Provider。
- 不与 CC Switch 同时管理 Codex 配置。

必须完成：

- CPA、OpenAI、自定义 Responses Provider 的添加、编辑、删除和连接测试。
- 模型同步或手动添加。
- Route 创建、编辑、删除、默认路由和激活。
- 多个 Codex 配置档案。
- Codex 模型目录同步，使 Codex 自带模型下拉列表显示全部已配置模型。
- 配置写入前备份，失败自动恢复。
- 自动重启 Codex 默认关闭，用户可勾选。
- 系统托盘常驻。
- 开机启动开关。
- Windows 安装包。
- 测试、文档和故障恢复说明。

自动更新和其他平台放入 V1.1 或后续独立 goal，不阻塞 V1 完成。

## 架构约束

- 采用 DDD 分层架构，保持领域逻辑与 Wails、HTTP、文件系统和 Windows API 解耦。
- 依赖方向只能由外向内：`interfaces → application → domain`；`infrastructure → application/domain`。
- `domain` 不依赖任何基础设施、UI 或 Wails 类型。
- `application` 只编排用例，通过端口/接口访问外部能力。
- `infrastructure` 实现配置文件、凭据、HTTP、进程和 Codex 适配等外部能力。
- `interfaces` 只负责 Wails 绑定、DTO 转换和错误呈现，不承载业务规则。
- `cmd` 是组合根，只负责组装依赖和启动程序。
- 使用 `PlatformAdapter` 策略封装平台配置、启动和验证。
- 使用适配器注册表/工厂按平台 ID 创建适配器。
- V1 只有 `CodexAdapter`。
- Provider 作为数据，不为每个供应商创建工厂。
- 暂不实现协议工厂、自建代理、协议转换、账户池、智能路由、云同步和插件系统。

推荐目录：

```text
cmd/codex-provider-hub/
internal/
  domain/{provider,model,route,profile}/
  application/{provider,model,route,profile}/
  infrastructure/{config,credential,codex,launcher,provider}/
  interfaces/wails/
frontend/
docs/
```

领域层优先使用标准库和纯 Go 类型；不把 Wails 生成类型、TOML 结构或 Windows API 传入领域层。

## 执行流程

1. 读取当前 goal、项目指令、项目文档和 Git 状态。
2. 先完成 Codex Desktop 配置读取、配置档案和模型目录的技术验证。
3. 验证 `Codex → CPA → 两个模型`。
4. 验证一个直连 Responses Provider。
5. 再实现核心数据、CodexAdapter、路由和备份恢复。
6. 最后实现 Wails UI、托盘、开机启动和安装包。
7. 每个非平凡逻辑保留至少一个可运行测试或验证。
8. 不私自扩大 V1 范围；发现约束时优先寻找兼容实现。

## 进度保存规则

- 每个实施阶段完成后，必须先更新文档并同步 CodeGraph，再创建一次 Git 提交。
- 长时间任务、工具链安装或可能中断前，先保存当前可运行进度。
- 阶段提交应使用清晰的 Conventional Commit 消息，便于从最近阶段恢复。
- 阶段边界尽量保持工作区干净；不得提交密钥、临时文件、构建产物或 CodeGraph 缓存。
- 后续工作从最近一次阶段提交继续，不重做已经提交的内容。

如果 Codex Desktop 不支持 `--profile`，使用备份后切换活动 `config.toml` 的方式实现。模型下拉列表是 V1 硬性要求，不得静默删除。

## V1 完成标准

- Windows 构建可安装并运行。
- Provider、Model、Route 全流程可用。
- CPA、OpenAI、自定义 Responses Provider 可用。
- 多个 Codex 配置档案可切换。
- Codex 自带模型下拉列表显示全部配置模型。
- 配置备份、恢复和异常提示可用。
- 自动重启默认关闭且勾选后有效。
- 托盘常驻和开机启动可用。
- 测试、文档和 Git 版本基线完整。
## Git 分支策略

- 日常开发、提交和验证统一使用 `dev` 分支。
- `main` 分支保留为 1.0 发布分支，在 V1 完成并验收前不更新。
- 发布 1.0 时再将已验收的 `dev` 合并或快进到 `main`。
