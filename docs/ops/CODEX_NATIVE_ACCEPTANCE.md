# Codex Desktop 原生验收清单

本文档用于在真实 Windows 10/11 + Codex Desktop 环境完成第一版验收。云端 Linux 工作区可以运行契约测试和前端构建，但不能代替 Windows WebView、DPI、Credential Manager 或 Codex Desktop 真实请求。

## 1. 前置条件

- Windows 10/11 x64、WebView2 Runtime、已安装 Codex Desktop
- Go 1.25.14、Node.js/npm、Task；首次构建先执行 `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.27`
- 可用的 CPA、OpenAI 或其他 Responses Provider 地址；真实 API Key 只通过环境变量或 Windows Credential Manager 提供
- 复制现有 Codex 配置目录到临时备份，并为验收设置独立 `CODEX_HOME`

PowerShell 示例：

```powershell
$env:CODEX_HOME = Join-Path $env:TEMP "proxy-switch-codex-acceptance"
$env:CODEX_DESKTOP_EXECUTABLE = "C:\Path\To\Codex.exe"
New-Item -ItemType Directory -Force $env:CODEX_HOME | Out-Null
Copy-Item "$HOME\.codex\config.toml" "$env:CODEX_HOME\config.toml" -ErrorAction SilentlyContinue
Copy-Item "$HOME\.codex\auth.json" "$env:CODEX_HOME\auth.json" -ErrorAction SilentlyContinue
```

不要把 API Key 直接粘贴到界面、`state.json`、日志或验收记录。

## 2. 构建和确定性检查

在仓库根目录运行：

```powershell
task check:frontend
task check:backend
task ci
```

若只验证后端契约，可运行：

```powershell
go test ./internal/...
go vet ./internal/...
```

契约覆盖包括：AuthRef 只能是环境变量/`credential:<target>` 引用；原始 API Key、空引用、Malformed TOML、重复导入目标均被拒绝；Provider DTO 只返回 preset/auth 模式和“引用已配置”状态，不返回凭据值；Profile 导入只复制 `config.toml`，不读写 `auth.json`。

## 3. 原生桌面验收

- [ ] 启动 `bin\proxy-switch.exe`，窗口可见，托盘单击可隐藏/恢复，退出菜单能结束进程
- [ ] 在 960×640 和 1120×760 逻辑尺寸检查页面滚动、Provider 行、Route 行和弹窗边界
- [ ] 在 Windows 显示缩放 125% 与 150% 各检查一次；文字、焦点环、下拉菜单和通知不裁切
- [ ] 配置页 Provider 卡片显示 preset（已识别时）、认证模式和“凭据引用已配置”，不显示完整密钥
- [ ] Provider 编辑只填写环境变量名或 `credential:<target>`；填入类似 `sk-live-secret` 后保存应失败且状态文件不变
- [ ] Profile 页读取当前 `config.toml` / `auth.json` 状态；点击“导入当前 Codex”后确认新档案路径为 `profiles\<id>\config.toml`
- [ ] 导入后逐字比较源/目标 `config.toml`，确认注释、未知键、MCP 和 OAuth 相关配置保留；确认 `auth.json` 未被复制或修改
- [ ] 用两个 Provider 和至少两个已启用 Model 创建 Route；停用 Model 后不能激活引用它的 Route
- [ ] 激活 Route 前确认 `.proxy-switch.bak` 存在；故意制造不可写目标后重试，确认活动配置、目录和备份标记恢复到激活前状态
- [ ] 点击“恢复备份”，确认活动 `config.toml` 与模型目录回到激活前内容
- [ ] 用真实 Provider 运行连接测试、同步 `/models`、最小 `/responses` 模型测试；失败提示不得包含 API Key、请求头或响应体
- [ ] 可选勾选“激活时重启 Codex”，确认旧进程停止后新进程启动；未勾选时激活不重启
- [ ] 设置页开关能读写开机启动；重启应用后状态保持

## 4. 记录结果

记录 Windows 版本、WebView2 版本、Codex Desktop 版本、显示缩放、构建提交、Provider 类型和每项清单结果。真实 CPA/OpenAI/自定义 Provider 请求仍需用户提供凭据；没有凭据时将对应项目标记为“未测”，不要填写推测结果。

云端限制：Linux 工作区可通过 `go test ./internal/...`、`go vet ./internal/...` 和前端 `npm run format:check && npm run lint && npm run build` 验证；`go test ./...` / `go vet ./...` 的根包需要 GTK4/WebKitGTK 的 Linux 原生依赖，当前环境缺失，因此不能作为 Windows 原生验收替代品。
