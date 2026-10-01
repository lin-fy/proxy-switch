# Windows 构建与恢复

## 已验证的开发构建

在项目根目录执行：

```powershell
wails3 build
```

当前 V1 已验证 Windows amd64 可执行文件生成到 `bin/codex-provider-hub.exe`，并完成启动冒烟检查。

如果终端找不到工具，先把以下目录加入当前进程的 `PATH`（路径按本机安装位置调整）：

```text
C:\Users\yfeeling\go\bin
C:\Users\yfeeling\AppData\Local\go-sdk\go\bin
C:\Program Files\nodejs
```

2026-10-01 起本机 Go 工具链为用户级安装：`C:\Users\yfeeling\AppData\Local\go-sdk\go`（go1.25.14.windows-amd64.zip 解压，无需管理员权限）；旧 `codex-go` 运行时目录已不存在。

## 安装包

NSIS 安装包：

```powershell
wails3 task windows:create:nsis:installer INSTALL_SCOPE=user
```

此命令需要 `makensis.exe`。如果没有管理员权限或不希望安装系统级 NSIS，可以把官方安装程序安装到用户目录：

```powershell
$nsis = "$env:LOCALAPPDATA\proxy-switch-tools\nsis"
Start-Process .\nsis-3.12-setup.exe -Wait -ArgumentList "/S", "/D=$nsis"
$env:MAKENSIS = "$nsis\makensis.exe"
wails3 task windows:create:nsis:installer INSTALL_SCOPE=user MAKENSIS=$env:MAKENSIS
```

`MAKENSIS` 也可以直接指向现有的 `makensis.exe`，不需要修改系统 `PATH`。用户级安装会写入 `%LOCALAPPDATA%\Programs\Codex Provider Hub`，不会要求管理员权限。

MSIX 安装包：

```powershell
wails3 tool msix-install-tools
wails3 task windows:create:msix:package USE_MSIX_TOOL=true
```

MSIX 路径需要 Windows SDK 和 Microsoft MSIX Packaging Tool。两者都不是 Go/Wails 依赖，安装后重新打开终端再执行打包命令。

如果暂时没有 NSIS 或 MSIX 工具，可以生成无管理员权限的用户级 ZIP 安装包：

```powershell
wails3 build
powershell -ExecutionPolicy Bypass -File build/windows/portable/package.ps1
```

解压后运行 `install.ps1` 会安装到 `%LOCALAPPDATA%\Programs\Codex Provider Hub` 并创建开始菜单快捷方式；安装目录中的 `uninstall.ps1` 可安全卸载该目录。

最新 Windows amd64 ZIP 已实际验证：安装脚本能复制可执行文件并创建开始菜单快捷方式，卸载脚本能移除安装目录。

NSIS 用户级安装包已实际验证安装、启动、升级、回滚和静默卸载；卸载后 `%LOCALAPPDATA%\Programs\Codex Provider Hub` 目录会被移除。

## 运行时配置

- `CODEX_HOME`：覆盖 Codex 配置目录；未设置时使用用户目录下的 `.codex`。
- `CODEX_DESKTOP_EXECUTABLE`：指定 Codex Desktop 可执行文件路径；未设置时使用 `codex`。
- Provider 的 `auth_ref` 可以填写环境变量名，或填写 `credential:<target>` 从 Windows Credential Manager 读取 Generic Credential。

## 配置恢复

每次路由激活前，活动配置和选中的 Profile 配置都会写入同名 `.codex-provider-hub.bak` 备份。界面中的“恢复备份”会恢复最近一次备份；多文件写入中途失败时应用会自动回滚已写入文件。

不要把 API Key 直接填入 Provider 名称、地址或普通 JSON 状态文件。状态文件只保存凭据引用。

## 路由激活验收

应用层测试覆盖以下关键顺序：

1. 读取路由关联的 Provider、Model、全部 Provider 模型和 Codex Profile。
2. 创建对应平台适配器，先用同一套 Provider 凭据执行选中模型的 Responses 可达性检查，再执行校验和配置准备。
3. 只有路由勾选“激活时重启”时才重启 Codex；已有进程会先按配置的可执行文件名停止，再启动新进程，并把 Provider 传入启动环境。
4. 配置准备和按需重启都成功后才把路由保存为默认路由；可达性失败不会写入配置。

这组测试位于 `internal/application/route/activate_test.go`，用于防止界面层调整时破坏路由激活边界。

Codex 适配器测试还覆盖：无备份时返回明确错误、活动配置与选中 Profile 配置同时恢复，以及多文件写入失败后的活动配置回滚。

Provider 模型同步通过 `/models` 增量 Upsert；本地已有 Model 的启用状态保留，远端消失的 Model 不自动删除，新 Model 默认启用。Codex 目录写入独立的 `<profile-id>.models.json`，`model_catalog_json` 只保存该文件路径；目录、活动配置和 Profile 配置使用同一轮备份和事务式恢复。

V1 验收加固还覆盖激活前可达性检查、Provider 凭据的子进程注入、引用删除保护和默认 Route 删除恢复。当前机器能运行 Codex CLI，但 Provider Hub 状态中没有 CPA 配置，因此 CPA 双模型和真实 Codex 重启链路仍需在目标环境手工复验。
