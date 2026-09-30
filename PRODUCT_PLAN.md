# Codex Provider Hub：需求与架构草案

> 状态：V1 核心链路和 Windows 桌面壳已完成，进入桌面能力与验收加固阶段
>
> 目标：供应商只配置一次，通过路由生成目标平台配置并启动对应平台。

## 1. 产品定位

这是一个以供应商为中心、以路由为核心的桌面管理工具。

它不是代理服务器，也不重复实现 CPA / CLIProxyAPI 的转发能力。它负责：

1. 管理 CPA、OpenAI 和其他供应商。
2. 管理供应商可用模型。
3. 创建“平台 → 供应商 → 模型”的路由。
4. 按平台规则写入配置。
5. 启动、重启并验证目标平台。

第一阶段只实现 Codex Desktop，架构保留后续接入 Claude Code、Gemini CLI、OpenCode 的边界。

## 已确认决策

- 首版只支持 Windows。
- 首版只实现 Codex 平台适配器。
- 路由激活时是否自动重启 Codex 由用户勾选，默认关闭。
- 首版只接入 Responses API 供应商。
- CPA 作为普通 Provider 管理，不在应用内做特殊的独立产品模块。
- 不与 CC Switch 同时管理 Codex 配置；Provider Hub 作为唯一配置管理入口。
- Codex 内置模型下拉列表需要显示已配置的全部可用模型。
- 支持多个 Codex 配置档案。
- 支持系统托盘常驻和开机启动；自动更新放到 V1.1。

## 2. 核心模型

```text
Provider（供应商）
    ↓
Model（模型）
    ↓
Route（路由）
    ↓
PlatformAdapter（平台适配器）
    ↓
Launch（启动目标平台）
```

### Provider

供应商只配置一次，包含：

- `id`
- 名称
- API 地址
- 协议类型
- 认证引用
- 自定义请求头和查询参数
- 模型列表

API Key 不存入普通 JSON；优先使用 Windows Credential Manager 或环境变量引用。

### Model

模型属于供应商，最小字段为：

- `provider_id`
- `model_id`
- 显示名称
- 是否启用

上下文长度、推理能力等元数据后续再补充。

模型列表需要同步到 Codex 的模型目录配置，使 Codex 自带模型选择界面显示已配置的全部可用模型。

### Route

路由表示目标平台如何使用某个供应商的某个模型：

```text
Codex → CPA → gpt-5
Codex → 供应商 A → model-x
```

如同一供应商对不同平台使用不同路径或协议，只在路由上增加覆盖字段，不重复保存密钥。

## 3. 路由激活流程

```text
选择路由
  ↓
校验供应商、模型和协议
  ↓
检查服务可达性
  ↓
备份目标平台当前配置
  ↓
通过平台适配器生成配置
  ↓
必要时重启目标平台
  ↓
执行连接验证
```

任何写入失败都必须恢复最近一次有效备份。

## 4. 技术架构

```text
Wails UI
    ↓
RouteService
    ├── ProviderRegistry
    ├── ModelRegistry
    ├── RouteRegistry
    ├── AdapterFactory / Registry
    ├── BackupService
    ├── CredentialService
    └── Launcher
```

### 设计模式使用边界

- 策略模式：`PlatformAdapter`，封装不同平台的配置、启动和验证逻辑。
- 工厂/注册表：根据平台 ID 创建适配器。
- Provider 保持为数据，不为每个供应商创建工厂。
- 协议适配器暂不抽象；只有出现第二种实际协议时再增加。

### DDD 分层与依赖方向

V1 采用 DDD 分层架构，业务规则与桌面框架、配置文件和 Windows API 解耦：

```text
interfaces (Wails / DTO)
          ↓
application (用例 / 编排)
          ↓
domain (Provider / Model / Route / Profile)

infrastructure (JSON / TOML / Credential / HTTP / Process / Codex)
          └────────实现 application/domain 所需端口
```

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

规则：

- `domain` 不依赖 Wails、HTTP、文件系统或 Windows API。
- `application` 只编排用例，通过端口访问外部能力。
- `infrastructure` 实现端口和平台适配。
- `interfaces` 只做 Wails 绑定、DTO 转换和错误呈现。
- `cmd` 负责依赖组装和启动。

最小接口方向：

```text
PlatformAdapter
├── Validate(route)
├── Prepare(route)
├── Launch(route)
└── IsRunning()
```

第一版只有 `CodexAdapter`。

## 5. Codex MVP

### 必须支持

- 读取用户级 Codex `config.toml`。
- 添加、更新自定义 Provider。
- 设置当前 `model_provider`。
- 设置当前 `model`。
- 写入前自动备份。
- 写入失败自动恢复。
- 检查 Codex 是否运行。
- 启动或重启 Codex。
- 测试 API 地址、模型和 Responses API 调用。
- 打开 CPAMP 管理页面。

### 供应商范围

- CPA / CLIProxyAPI。
- OpenAI。
- 一个自定义 Responses API 供应商。

CPA 与其他供应商在产品层级上统一走 Provider 管理和路由流程。

### 暂不支持

- 自建代理转发。
- 多平台完整实现。
- 自动协议转换。
- 账户池和智能路由。
- MCP、Skills、Prompts 管理。
- 云同步、插件和多用户权限。

以下能力属于首版范围：

- 多个 Codex 配置档案。
- 系统托盘常驻。
- 开机启动。
- 自动更新（V1.1）。

## 6. 兼容性边界

Codex 自定义 Provider 当前应按 Responses API 设计。

供应商分为三类：

1. 直接支持 Responses API：可以直接接入 Codex。
2. 由 CPA 转换为 Responses API：Codex 连接 CPA。
3. 仅支持 Chat Completions 或原生协议：不能保证直连，需要 CPA 或其他适配层。

CPA-Manager-Plus 负责管理和观测；实际请求转发由 CPA / CLIProxyAPI 完成。

## 7. 界面范围

### 当前路由

显示当前平台、供应商、模型和服务状态，提供启动、切换、测试和打开管理页。

### 供应商

添加、编辑、删除、测试连接和刷新模型。

### 路由

创建、复制、启用、停用和启动路由。

### 设置

包含 Codex 配置档案、备份目录、自动重启、启动检查、托盘常驻、开机启动和更新设置。

## 8. 技术取舍

- 桌面框架：Wails。
- 后端：Go。
- 首版平台：Windows。
- 前端：先使用项目成员熟悉的轻量方案，不引入额外状态管理库。
- 本地数据：JSON 文件即可，暂不引入数据库。
- 密钥：Windows Credential Manager 或环境变量引用。
- 日志：默认脱敏，不记录 API Key、管理密钥和完整响应体。

## 9. 首个验证任务

先不做完整 UI，完成最小链路验证：

1. 手动配置 Codex 自定义 Provider。
2. 连接 CPA。
3. 调用 CPA 中两个不同模型。
4. 测试流式输出和工具调用。
5. 再验证一个直连 Responses API 供应商。
6. 确认 Codex 桌面端重启后读取新配置。
7. 验证多个 Codex 配置档案切换。
8. 验证 Codex 内置模型下拉列表显示同步后的全部模型。
9. 验证托盘常驻和开机启动；自动更新在 V1.1 单独验证。

核心配置验证通过后已进入 Wails 界面、桌面能力和安装验收阶段。

## 10. 待确认事项

- Codex Desktop 当前版本是否在启动时读取同一份用户级 `config.toml`。
- Codex 模型目录配置的实际格式和刷新时机。
- CPA 对外提供的 Codex/Responses API 实际地址和认证方式。
- 多个 Codex 配置档案与模型目录之间的对应关系。
- 自动更新采用 Wails 生态方案还是 Windows 原生安装包更新方案。

## 11. 当前技术判断

官方 Codex 配置支持：

- 用户级 `config.toml`。
- 位于 `$CODEX_HOME` 下的命名配置档案。
- 通过 `--profile` 选择配置档案。
- 通过 `model_catalog_json` 指定启动时加载的模型目录。

因此，应用内的一个 Codex 配置档案至少应包含：

```text
profile-name.config.toml
profile-name.models.json
```

需要验证 Codex Desktop 是否暴露或继承 CLI 的 `--profile` 行为。若桌面端不支持，首版回退为：

```text
选中档案 → 备份当前 config.toml → 激活档案内容 → 按需重启 Codex
```

模型下拉列表同步依赖 `model_catalog_json` 的实际 JSON schema，先通过最小样例验证，不在未确认 schema 前实现复杂的模型目录编辑器。

### 本机只读观察

- 本机存在用户级 Codex 配置目录和 `config.toml`。
- 当前配置使用自定义 Provider，当前模型为 `gpt-5.6-sol`。
- 当前配置未发现显式的 `model_catalog_json` 配置。
- 当前配置目录未发现命名的 `*.config.toml` 档案文件。
- 以上仅为配置入口观察，不读取或记录认证凭据。

### V1 开发验证记录

- Go 1.27.1 便携工具链已就绪。
- Wails CLI `wails3` v3.0.0-beta.26 已安装。
- DDD 领域层和应用端口已通过 `go test ./...`。
- Wails v3 vanilla-ts 壳已成功构建 Windows amd64 可执行文件，前端绑定使用 `@wailsio/runtime`。
- 已实现 Codex 配置写入、写入前备份/失败恢复、模型目录同步和基础 Provider/Route 管理界面。
- 多个 Profile 已支持独立配置文件；激活时同步写入活动 `config.toml`，兼容不支持 `--profile` 的 Codex Desktop。
- 多文件写入失败会回滚已写入的活动配置，激活成功后路由会成为当前默认路由。
- 已接入 Wails v3 系统托盘：关闭窗口隐藏、托盘单击切换、菜单显示/隐藏/退出。
- 已接入 Wails v3 开机启动：Windows 使用 `HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run`，界面提供启用/禁用开关。
- Go 测试、前端构建、Wails Windows amd64 构建和启动冒烟验证已通过。

## 12. 实施顺序

### 阶段 A：配置与兼容性验证

1. 确认 Codex Desktop 的配置读取路径。
2. 验证一个自定义 Responses Provider。
3. 验证 CPA 和两个 CPA 模型。
4. 验证直连 Responses Provider。
5. 验证多个配置档案的切换方式。
6. 解析并验证模型目录 JSON。

### 阶段 B：核心应用

1. Provider、Model、Route 数据存储。
2. CodexAdapter 和适配器注册表。
3. 配置备份、恢复和原子写入。
4. 路由激活与连接测试。
5. 模型目录生成。

### 阶段 C：Windows 桌面能力

1. ~~系统托盘常驻。~~ 已完成：关闭窗口隐藏，托盘菜单支持显示/隐藏/退出。
2. ~~开机启动开关。~~ 已完成：使用 Wails v3 Autostart，界面可切换。
3. 默认关闭的自动重启开关。
4. 安装包和版本回滚验证。

自动更新属于 V1.1，优先使用成熟的 Wails/Windows 安装更新方案，不自行实现下载、替换和回滚器。
