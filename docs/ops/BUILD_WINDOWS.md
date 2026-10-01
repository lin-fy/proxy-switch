# Windows 构建、安装与恢复

本文档是 V1 在 Windows 上构建、打包、安装、升级、卸载和故障恢复的最终操作说明。适用于干净 Windows 10/11 环境，以及本机开发复验。

## 1. 工具链前提

| 工具 | 版本 / 位置 | 用途 |
| --- | --- | --- |
| Go | `go.mod` 要求 1.25.x | 后端构建与测试 |
| Node.js | 20+（CI 使用 22） | 前端构建 |
| Wails CLI | `v3.0.0-beta.26` | 桌面构建与打包 |
| NSIS | 3.12（系统级或用户级） | 生成 NSIS 安装包 |
| WebView2 Runtime | Windows 10/11 通常已内置 | 运行桌面界面 |

安装 Wails CLI：

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.26
```

如果终端找不到工具，先把以下目录加入当前进程的 `PATH`（路径按本机安装位置调整）：

```text
%USERPROFILE%\go\bin
%LOCALAPPDATA%\go-sdk\go\bin
C:\Program Files\nodejs
```

2026-10-01 起本机 Go 工具链为用户级安装：`%LOCALAPPDATA%\go-sdk\go`（go1.25.14.windows-amd64.zip 解压，无需管理员权限）；旧 `codex-go` 运行时目录已不存在。

## 2. 开发构建

在仓库根目录执行：

```powershell
wails3 build
```

产物为 `bin/codex-provider-hub.exe`（Windows amd64，约 11.4 MB），并已完成启动冒烟检查。

也可以使用统一任务入口：

```powershell
task build          # 等价于 wails3 build
task ci             # 完整验证 + Windows 构建 + portable ZIP
```

## 3. 验证

提交或发布前运行模型无关的确定性检查：

```powershell
task check:frontend   # 前端格式、lint、类型检查与生产构建
task check:backend    # gofmt、go mod tidy 校验、go vet、go test
task ci               # 上述全部 + wails3 build + portable ZIP
```

`task ci` 复用 `scripts/ci/verify.ps1 -BuildWindows`，与 GitHub Actions 执行同一套检查；任何 agent 或本地环境都应通过该入口验证，不依赖特定 LLM。

## 4. 安装包

### 4.1 NSIS 用户级安装包（推荐）

```powershell
wails3 task windows:create:nsis:installer INSTALL_SCOPE=user
```

- 产物：`bin/codex-provider-hub-amd64-installer.exe`。
- 安装位置：`%LOCALAPPDATA%\Programs\Codex Provider Hub`，不需要管理员权限。
- 安装过程会创建开始菜单和桌面快捷方式，并在缺少 WebView2 时静默安装运行时。
- 卸载信息写入 `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Codex Provider HubCodex Provider Hub`。

此命令需要 `makensis.exe`。如果没有管理员权限或不希望安装系统级 NSIS，可以把官方安装程序安装到用户目录：

```powershell
$nsis = "$env:LOCALAPPDATA\proxy-switch-tools\nsis"
Start-Process .\nsis-3.12-setup.exe -Wait -ArgumentList "/S", "/D=$nsis"
$env:MAKENSIS = "$nsis\makensis.exe"
wails3 task windows:create:nsis:installer INSTALL_SCOPE=user MAKENSIS=$env:MAKENSIS
```

`MAKENSIS` 也可以直接指向现有的 `makensis.exe`，不需要修改系统 `PATH`。

静默安装与静默卸载（供自动化验收）：

```powershell
# 静默安装到当前用户目录
Start-Process .\bin\codex-provider-hub-amd64-installer.exe -Wait -ArgumentList '/S'

# 静默卸载
& "$env:LOCALAPPDATA\Programs\Codex Provider Hub\uninstall.exe" /S
```

### 4.2 Portable ZIP（无需 NSIS）

```powershell
wails3 build
powershell -ExecutionPolicy Bypass -File build/windows/portable/package.ps1
```

- 产物：`bin/codex-provider-hub-windows-amd64.zip`。
- 解压后运行 `install.ps1` 会安装到 `%LOCALAPPDATA%\Programs\Codex Provider Hub` 并创建开始菜单快捷方式。
- 安装目录中的 `uninstall.ps1` 只在存在 `.codex-provider-hub-install` 标记时删除目录，可安全卸载。

### 4.3 MSIX（可选路径）

```powershell
wails3 tool msix-install-tools
wails3 task windows:create:msix:package USE_MSIX_TOOL=true
```

MSIX 路径需要 Windows SDK 和 Microsoft MSIX Packaging Tool，两者都不是 Go/Wails 依赖，安装后重新打开终端再执行打包命令。MSIX 不属于 V1 发布阻塞项。

## 5. 签名与发布

正式发布由 `.github/workflows/release.yml` 完成：校验 tag 与版本元数据、构建、签名 exe 与 NSIS 安装包、生成 `SHA256SUMS.txt` 并创建 GitHub Release。

发布前必须确保三处版本一致（`scripts/ci/validate-release.ps1` 会校验）：

- `build/config.yml` 的 `info.version`
- `build/windows/info.json` 的 `fixed.file_version` 与 `ProductVersion`
- `build/windows/nsis/wails_tools.nsh` 的 `INFO_PRODUCTVERSION`

`production` Environment 需要以下 Secrets：

```text
WINDOWS_CERTIFICATE_BASE64
WAILS_WINDOWS_CERT_PASSWORD
```

本地签名与打包：

```powershell
wails3 task windows:sign:installer ARCH=amd64 INSTALL_SCOPE=user SIGN_CERTIFICATE="C:\path\to\cert.pfx"
```

签名证书只写入临时目录，不进入产物；发布流程会校验 Authenticode 签名和时间戳，两者缺失都会失败。

## 6. 运行时配置

| 环境变量 | 说明 |
| --- | --- |
| `CODEX_HOME` | 覆盖 Codex 配置目录；未设置时使用用户目录下的 `.codex` |
| `CODEX_DESKTOP_EXECUTABLE` | 指定 Codex Desktop 可执行文件路径；未设置时使用 `codex` |
| `CPH_REMOTE_DEBUG_PORT` | 仅验收用；设置后开启 CDP 远程调试端口，默认关闭 |

Provider 的 `auth_ref` 可以填写环境变量名（也接受 `env:` 前缀），或填写 `credential:<target>` 从 Windows Credential Manager 读取 Generic Credential。

普通状态文件位于 `%APPDATA%\CodexProviderHub\state.json`，只保存凭据引用，不保存 API Key。不要把密钥填入 Provider 名称、地址或状态文件。

## 7. 卸载与残留

NSIS 卸载会移除：

- 安装目录 `%LOCALAPPDATA%\Programs\Codex Provider Hub`
- 开始菜单与桌面快捷方式
- `HKCU` 卸载注册项
- `%APPDATA%\codex-provider-hub.exe` 下的 WebView2 数据目录

卸载**不会**自动删除：

- `%APPDATA%\CodexProviderHub\state.json`（应用状态）
- Codex 自身的配置、Profile 和 `.codex-provider-hub.bak` 备份

如需彻底清理，确认不再需要配置后再手动删除上述路径。删除前建议先备份 `%APPDATA%\CodexProviderHub` 和 Codex 配置目录。

## 8. 故障恢复

### 8.1 构建阶段

| 现象 | 处理 |
| --- | --- |
| `wails3: command not found` | 执行第 1 节的安装命令，并把 `%USERPROFILE%\go\bin` 加入当前进程 `PATH` |
| `frontend/dist` 缺失或前端构建失败 | 进入 `frontend` 执行 `npm ci`，再回到根目录重试 |
| Wails 报图标或资源缺失 | 执行 `wails3 task common:generate:icons` 后重试 |
| `task ci` 报 `go mod tidy changed go.mod or go.sum` | 在根目录执行 `go mod tidy` 并提交 `go.mod`/`go.sum`，不要忽略该检查 |
| `makensis` 找不到 | 设置 `MAKENSIS` 指向 `makensis.exe`，见 4.1 |
| MSIX 打包失败 | 安装 Windows SDK 和 MSIX Packaging Tool 后重开终端；或改用 NSIS / portable 路径 |

### 8.2 安装与升级阶段

| 现象 | 处理 |
| --- | --- |
| 安装包提示架构不支持 | 安装包为 amd64，确认目标机为 64 位；ARM64 需单独构建 |
| 安装时要求管理员权限 | 使用 `INSTALL_SCOPE=user` 生成的用户级安装包，机器级安装才需要 UAC |
| 安装被“应用正在运行”阻塞 | 从托盘退出应用，或在任务管理器结束 `codex-provider-hub.exe` 后重试 |
| 升级后仍是旧版本 | 确认安装包版本高于已安装版本，且安装目录未被其他进程占用；必要时先卸载旧版再安装 |
| 升级异常需回滚 | 重新运行上一个版本的安装包覆盖安装；应用状态文件向后兼容，不会被安装包删除 |
| 卸载后残留注册项 | 删除 `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Codex Provider HubCodex Provider Hub` |

### 8.3 启动与运行时

| 现象 | 处理 |
| --- | --- |
| 窗口空白或无法渲染 | 确认已安装 WebView2 Runtime；安装包会在缺失时自动安装 |
| 托盘图标存在但窗口不显示 | 点击托盘图标或使用托盘菜单“显示主窗口” |
| 关闭窗口后进程仍在 | 这是预期行为：关闭窗口仅隐藏到托盘，需通过托盘菜单“退出”结束进程 |
| 找不到 Codex Desktop | 设置 `CODEX_DESKTOP_EXECUTABLE` 指向实际可执行文件 |
| Codex 使用了错误的配置目录 | 检查是否设置了 `CODEX_HOME`；未设置时使用 `%USERPROFILE%\.codex` |

### 8.4 Codex 配置损坏或激活失败

每次路由激活前，活动配置、选中的 Profile 配置和模型目录都会写入同名 `.codex-provider-hub.bak` 备份。多文件写入中途失败时应用会自动回滚已写入文件。

界面中的“恢复备份”会恢复最近一次备份。若需要手工恢复：

1. 退出 Codex Provider Hub 和 Codex Desktop。
2. 在 `%CODEX_HOME%`（默认 `%USERPROFILE%\.codex`）找到 `config.toml.codex-provider-hub.bak`、Profile 配置备份和 `*.models.json.codex-provider-hub.bak`。
3. 将备份文件复制回对应的原始路径（去掉 `.codex-provider-hub.bak` 后缀）。
4. 如果存在 `.codex-provider-hub.bak.missing` 标记，说明激活前该文件不存在，应删除对应原始文件而不是恢复。
5. 重新启动 Codex Desktop 验证配置生效。

若备份缺失，界面恢复会返回明确错误；此时只能从 Codex 自身的配置或版本控制中恢复，应用不会伪造默认配置。

### 8.5 发布阶段

| 现象 | 处理 |
| --- | --- |
| Tag 校验失败 | 确认 tag 形如 `v0.1.0`、位于 `main`、且与三处版本元数据一致 |
| 缺少 `WINDOWS_CERTIFICATE_BASE64` | 在 GitHub `production` Environment 配置签名 Secrets |
| 签名或时间戳校验失败 | 检查证书是否有效、密码是否正确、时间戳服务是否可达；发布流程要求签名带时间戳 |
| Release 已存在 | 删除或重命名冲突的 tag/Release 后重新运行，不要复用已发布 tag |

## 9. V1 安装验收清单

- [ ] `task ci` 在干净检出上通过。
- [ ] 生成 `bin/codex-provider-hub.exe` 并完成启动冒烟。
- [ ] 生成用户级 NSIS 安装包或 portable ZIP。
- [ ] 无管理员权限完成安装、启动、隐藏到托盘、退出。
- [ ] 覆盖安装完成升级，回滚到旧版可正常启动。
- [ ] 卸载后安装目录、快捷方式和注册项被移除。
- [ ] Codex 配置备份存在，恢复备份可回到激活前状态。

上述安装、启动、升级/回滚和卸载步骤已在本机实际验证；真实 CPA / OpenAI / 自定义 Provider 的请求验收属于 M6，需要用户提供凭据，见 [`../planning/ROADMAP.md`](../planning/ROADMAP.md)。
