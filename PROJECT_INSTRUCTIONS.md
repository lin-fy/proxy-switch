# Codex Provider Hub：V1 项目指令

## 当前目标

当前阶段只有一个长期 goal：完成 Codex Provider Hub Windows V1。

每轮对话、开发、验证、修复和文档更新都属于同一个 V1 goal，不拆成独立用户任务。

V1 验收完成后关闭当前 goal。V1.1、V2 或其他平台支持必须创建新的独立 goal。

## 首轮初始化

1. 初始化 Git，创建合理的 `.gitignore` 和初始基线提交。
2. 初始化或启用 CodeGraph；理解代码时优先使用 CodeGraph。
3. CodeGraph 不可用时再使用普通文件搜索。
4. `sources/` 中的同步文件只读，不得修改、移动或删除。

## 模型分工

- `gpt-5.6-sol`：项目统筹、需求拆解、架构决策、审查、集成和验收。
- `gpt-5.6-luna`（max）：执行独立实现、测试、调查和文档子任务。
- 子任务都属于当前 V1 goal，不创建新的用户级项目任务。
- 指定模型不可用时必须说明，不静默替换。

## V1 范围

- Windows only。
- Wails + Go。
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

- 使用 `PlatformAdapter` 策略封装平台配置、启动和验证。
- 使用适配器注册表/工厂按平台 ID 创建适配器。
- V1 只有 `CodexAdapter`。
- Provider 作为数据，不为每个供应商创建工厂。
- 暂不实现协议工厂、自建代理、协议转换、账户池、智能路由、云同步和插件系统。

## 执行流程

1. 读取当前 goal、项目指令、项目文档和 Git 状态。
2. 先完成 Codex Desktop 配置读取、配置档案和模型目录的技术验证。
3. 验证 `Codex → CPA → 两个模型`。
4. 验证一个直连 Responses Provider。
5. 再实现核心数据、CodexAdapter、路由和备份恢复。
6. 最后实现 Wails UI、托盘、开机启动和安装包。
7. 每个非平凡逻辑保留至少一个可运行测试或验证。
8. 不私自扩大 V1 范围；发现约束时优先寻找兼容实现。

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

