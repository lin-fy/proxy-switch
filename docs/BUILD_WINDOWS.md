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
C:\Users\yfeeling\AppData\Local\codex-go\runtime\go\bin
C:\Program Files\nodejs
```

## 安装包

NSIS 安装包：

```powershell
wails3 task windows:create:nsis:installer INSTALL_SCOPE=user
```

此命令需要 `makensis.exe`。用户级安装会写入 `%LOCALAPPDATA%\Programs\Codex Provider Hub`，不会要求管理员权限。

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
2. 创建对应平台适配器并执行校验和配置准备。
3. 只有路由勾选“激活时重启”时才启动 Codex。
4. 配置准备成功后才把路由保存为默认路由。

这组测试位于 `internal/application/route/activate_test.go`，用于防止界面层调整时破坏路由激活边界。
