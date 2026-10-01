# 项目当前状态快照

> agent 协作的"单一事实源"。开工读，收工更新。只保留当前状态。
> 详细规则见 [`agent-guide/workflow.md`](agent-guide/workflow.md)。

最后更新：2026-10-02 · codex（仓库已公开，main 保护与 production 审批已生效）

## 一句话现状

V1 后端核心 + 桌面壳 + Vue 前端已完成并通过桌面原生验收（M4 全部完成，含窗口/DPI/浮层/键盘/IME 与 Mock 全链路走查）；NSIS 安装包全流程验收与 `BUILD_WINDOWS.md` 最终文档定稿完成，M5 全部完成。M6-A/B 的目标、证据矩阵和本地确定性验证已完成；当前先推进 M6-D，M6-C 真实 Provider 验收延后，M6-E 正式发布仍需两者完成。

## 下一步三项

1. 确认签名凭据可用性并准备 Tag 演练；GitHub 配置已完成，当前证书仍为自签名测试证书
2. 提供 CPA、OpenAI、自定义 Provider 凭据和 Codex Desktop 环境，完成真实端到端验收（**按用户安排稍后进行**）
3. M6-C/D 通过后统一切换 `1.0.0`、合并 `main` 并发布正式安装包

统一验证入口已补齐：`task check:frontend`、`task check:backend`、`task ci`；GitHub Actions 与本地入口均复用 `scripts/ci/verify.ps1`，兼容 Windows PowerShell 5.1，并在 Windows 构建时自动搜索 `GOPATH\bin` 的 Wails CLI。2026-10-02 本地完整验证通过并生成 portable ZIP。

## 当前卡点

- 真实 CPA / OpenAI / 自定义 Provider 凭据（**需用户提供**）
- 签名与 Tag 演练尚未完成：两个 Secret 名称已确认存在，证书为自签名测试证书；现有 Release 工作流要求签名 `Valid` 且带时间戳，凭据可用性和信任尚未验证

## 待协调事项

- 正式 `1.0.0` 签名需要受 Windows 信任链认可、包含私钥且可导出为 `.pfx` 的 Authenticode 证书；自签名证书不能作为公开发布替代品。

### 已裁决（归档）

- 2026-10-02：用户选择将仓库设为公开；GitHub API 已确认公开、`main` 保护和 `production` 审批生效（ADR-0012），原方案限制已解除
- 2026-10-01：TASKS.md codex 代记条目归属问题 → 删除，由 zcode 自己条目承担
- 2026-10-01：CI 与开发 agent、模型解耦，AI Review/失败诊断仅作为可选分析层（ADR-0011）

## 长期注意事项

- 当前已切换 GitHub PR 流程：任务 agent 在自己的分支同步 `dev`、解决冲突、验证并创建 PR；多个 agent 并行时由用户指定 reviewer/merger，只有一个活跃 agent 时允许自审并在分支保护允许时合并，但必须等待 CI 通过。不得覆盖他人未提交改动。
- `sources/` 下文件只读
- 不得提交密钥、临时文件、构建产物、CodeGraph 缓存
- 瞬时工作区状态直接 `git status`，不在本文档维护

---

**详细里程碑与验收状态**：[`planning/ROADMAP.md`](planning/ROADMAP.md)
**任务认领看板**：[`TASKS.md`](TASKS.md)
**变更历史**：`git log` + [`CHANGELOG.md`](CHANGELOG.md)
