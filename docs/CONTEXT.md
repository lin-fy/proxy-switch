# 项目当前状态快照

> agent 协作的"单一事实源"。开工读，收工更新。只保留当前状态。
> 详细规则见 [`agent-guide/workflow.md`](agent-guide/workflow.md)。

最后更新：2026-10-01 · codex（NSIS 用户级安装包验收）

## 一句话现状

V1 后端核心 + 桌面壳 + Vue 前端基线已完成；桌面原生验收、安装包、真实端到端验收未完成。工作区已提交干净。

## 下一步三项

1. M4 收尾：Wails 桌面窗口、DPI、浮层、键盘/IME、真实后端流程验收
2. GitHub 配置 `production` Environment + 签名 Secrets + `main` 分支保护，完成 Tag 发布演练（**需用户操作**）
3. 补齐 NSIS / MSIX 安装包并验证

## 当前卡点

- 真实 CPA / OpenAI / 自定义 Provider 凭据（**需用户提供**）
- GitHub 仓库 Secrets 配置（**需用户操作**）

## 待协调事项

- <暂无>

### 已裁决（归档）

- 2026-10-01：TASKS.md codex 代记条目归属问题 → 删除，由 zcode 自己条目承担

## 长期注意事项

- `sources/` 下文件只读
- 不得提交密钥、临时文件、构建产物、CodeGraph 缓存
- 瞬时工作区状态直接 `git status`，不在本文档维护

---

**详细里程碑与验收状态**：[`planning/ROADMAP.md`](planning/ROADMAP.md)
**任务认领看板**：[`TASKS.md`](TASKS.md)
**变更历史**：`git log` + [`CHANGELOG.md`](CHANGELOG.md)
