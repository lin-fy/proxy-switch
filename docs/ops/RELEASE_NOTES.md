# Codex Provider Hub 发布说明（草稿）

> 面向用户的 V1 发布说明草稿。正式发布时粘贴到 GitHub Release（替代 `--generate-notes` 自动内容），或用 `gh release edit v1.0.0 --notes-file docs/ops/RELEASE_NOTES.md`。

## Codex Provider Hub <版本号以实际发布 tag 为准>

Windows 桌面工具：把 Codex Desktop 的 Responses API 配置（Provider、Model、Route、Profile）管理起来，安全完成切换、测试、备份和恢复。

### 亮点

- **Provider / Model / Route 管理**：统一 Responses 协议，支持 CPA-Manager-Plus、OpenAI 与任意兼容服务；`/models` 一键同步，连接与模型测试内置。
- **路由激活**：选择 Provider + Model 组合成 Route，一键激活写入 Codex 配置；激活前后自动备份，失败自动回滚，可从「配置档案」页恢复最近备份。
- **多配置档案**：为不同工作环境保留独立的 Codex 配置文件，切换档案即时生效。
- **凭据安全**：本地状态文件不保存 API Key，凭据以环境变量名或 Windows Credential Manager 引用（`credential:<target>`）形式使用。
- **系统托盘**：关闭窗口隐藏到托盘，开机自启动可开关；路由激活可选自动重启 Codex（默认关闭）。
- **安装**：用户级 NSIS 安装包（无需管理员权限）与便携 ZIP 两种形态，均带 SHA256 校验与 Authenticode 签名。

### 下载

| 文件 | 说明 |
| --- | --- |
| `CodexProviderHub-<版本>-Setup.exe` | NSIS 安装包，安装到 `%LOCALAPPDATA%\Programs\Codex Provider Hub` |
| `CodexProviderHub-<版本>-portable.zip` | 便携版，解压即用 |
| `SHA256SUMS.txt` | 上述文件 SHA256 校验值 |

系统要求：Windows 10/11（x64），WebView2 Runtime（Windows 11 内置）。

> **签名说明**：0.1.0 为未签名构建，首次运行可能出现 SmartScreen"未知发布者"提示
> （选择"仍要运行"即可）。请用 `SHA256SUMS.txt` 校验下载文件完整性。受信签名将在
> 后续版本通过免费的开源签名服务（如 SignPath Foundation）补齐。

### 已知边界（V1）

- 仅适配 Codex Desktop（单平台适配器）；其他平台留待后续目标。
- 自动更新未包含，升级通过重新安装。
- 首次使用需要已有 Provider 服务地址与凭据。

### 升级与卸载

- 升级：直接运行新版安装包覆盖安装（用户级，配置与档案保留）。
- 卸载：系统「设置 → 应用」或安装目录卸载入口；Codex 配置备份保留在原位。
