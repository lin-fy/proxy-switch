# 项目当前状态快照

> agent 协作的"单一事实源"。开工读，收工更新。只保留当前状态。
> 详细规则见 [`agent-guide/workflow.md`](agent-guide/workflow.md)。

最后更新：2026-10-01 · codex（M5 收尾：BUILD_WINDOWS.md 定稿）

## 一句话现状

V1 后端核心 + 桌面壳 + Vue 前端已完成并通过桌面原生验收（M4 全部完成，含窗口/DPI/浮层/键盘/IME 与 Mock 全链路走查）；NSIS 安装包全流程验收与 `BUILD_WINDOWS.md` 最终文档定稿完成，M5 全部完成。剩余 M6 验收与发布，依赖用户凭据与 GitHub 配置。

## 下一步两项

1. GitHub 配置 `production` Environment + 签名 Secrets + `main` 分支保护，完成 Tag 发布演练（**需用户操作**）
2. 完成 CPA、OpenAI、自定义 Provider 的真实端到端验收（**需用户提供凭据**）

统一验证入口已补齐：`task check:frontend`、`task check:backend`、`task ci`；GitHub Actions 与本地入口均复用 `scripts/ci/verify.ps1`，不依赖任何 LLM。

## 当前卡点

- 真实 CPA / OpenAI / 自定义 Provider 凭据（**需用户提供**）
- GitHub 仓库 Secrets 配置（**需用户操作**）

## 待协调事项

- <暂无>

### 已裁决（归档）

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
