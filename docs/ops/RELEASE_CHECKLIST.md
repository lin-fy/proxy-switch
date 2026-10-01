# Tag 发布演练与正式发布清单

> 配套 [`../../.github/workflows/release.yml`](../../.github/workflows/release.yml)。
> 版本一致性由 `scripts/ci/validate-release.ps1` 在流水线内强制校验，本地不要绕过。

## 1. 一次性配置（GitHub 仓库，需仓库管理员操作）

| 事项 | 位置 | 说明 |
| --- | --- | --- |
| `production` Environment | Settings → Environments | `release.yml` 声明 `environment: production`；建议勾选 required reviewers，发布时人工确认 |
| `WINDOWS_CERTIFICATE_BASE64` | Environments → production → Secrets | 代码签名证书 `.pfx` 的 base64：`[Convert]::ToBase64String([IO.File]::ReadAllBytes("cert.pfx"))` |
| `WAILS_WINDOWS_CERT_PASSWORD` | Environments → production → Secrets | 证书私钥密码 |
| `main` 分支保护 | Settings → Branches | 发布 tag 必须位于 `main`（`validate-release.ps1` 校验 `--is-ancestor $tagCommit origin/main`） |

缺少任一 Secret 时，流水线会在签名阶段明确失败，不会产出未签名产物。

## 2. 版本一致性（发布前自检）

以下四处必须完全一致，tag 为 `vX.Y.Z`（不含 prerelease 后缀，各段 ≤ 65535）：

- [ ] `build/config.yml` → `version: "X.Y.Z"`
- [ ] `build/windows/info.json` → `fixed.file_version` 与 `info."0000".ProductVersion`
- [ ] `build/windows/nsis/wails_tools.nsh` → `INFO_PRODUCTVERSION`
- [ ] `build/windows/wails.exe.manifest` → `assemblyIdentity@version`（流水线不校验，仅保持一致）
- [ ] `build/windows/msix/app_manifest.xml`、`template.xml` → `Version="X.Y.Z.0"`（MSIX 路径使用）
- [ ] 界面尾栏版本（`frontend/src/components/DesktopShell.vue`）仅为展示，不参与校验

版本号在发布演练前统一即可（当前基线 `0.1.0`；发布时确定目标版本并同步以上全部位置，`wails.exe.manifest` 为三段、MSIX 为四段格式）。

### 已知问题

- exe「属性 → 详细信息」中的文件版本/产品版本当前显示为空：版本资源字节已嵌入（info.json 经 winres 写入 syso），但 Windows VersionInfo API 读不到，属 winres/wails3 上游问题；NSIS 安装包注册表 `DisplayVersion` 不受影响（来自 `INFO_PRODUCTVERSION`）。可向上游反馈。

## 3. 发布流程

1. 完成验收后，通过 PR 将 `dev` 合并到 `main`。
2. 在 `main` 打 tag 并推送：

   ```bash
   git tag -a v1.0.0 -m "Codex Provider Hub 1.0.0" && git push origin v1.0.0
   ```

   也可以在 GitHub Actions 手动运行 Release 工作流并填入已有 tag（`workflow_dispatch`）。
3. 流水线自动执行：模型无关验证与 Windows 构建 → NSIS 安装包 → exe 与安装包签名 → Authenticode 验证（必须带时间戳）→ 用签名后可执行文件重建 portable ZIP → 生成 `SHA256SUMS.txt` → 创建 GitHub Release 并上传三个产物。
4. 人工核验：Release 页下载安装包，校验 SHA256，干净环境安装、启动、卸载。

## 4. 演练说明

- 演练即按上述流程走一遍 `v1.0.0`：失败可删除 Release 与 tag 后修正重跑，不影响仓库状态。
- `--generate-notes` 使用 GitHub 自动生成的变更说明；正式发布时可替换为
  [`docs/ops/RELEASE_NOTES.md`](RELEASE_NOTES.md) 的定稿内容（`gh release edit --notes-file`）。
- 构建机要求由 runner 提供（Windows + NSIS + Windows SDK signtool），本地无需配置签名环境。
