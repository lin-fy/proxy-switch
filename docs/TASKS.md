# 任务认领表

> 多 agent 协作的任务看板。规则见 [`../AGENTS.md`](../AGENTS.md) 第三节。
>
> 状态标记：`[ ]` 未认领 · `[- by <agent> <日期>]` 进行中 · `[x by <agent> <完成日期>]` 已完成 · `[!]` 阻塞

最后更新：2026-10-01

## 进行中

- <暂无>

## 待认领（按优先级排序）

### M4 — Vue 桌面界面收尾

- [ ] Wails 桌面窗口尺寸、DPI、浮层行为验收（960×640 和 1120×760）
- [ ] 键盘导航与 IME 输入在桌面端的验收
- [ ] 真实后端流程端到端走查（Provider/Model/Route/Profile 全流程）

### M5 — Windows 安装包

- [x by codex 2026-10-01] 补齐 NSIS 或 MSIX 安装包工具链
- [x by codex 2026-10-01] 验证安装、启动、升级/回滚、卸载全流程

### M6 — V1 验收

- [ ] 在 GitHub 配置 `production` Environment、签名 Secrets、`main` 分支保护（**需要用户操作**）
- [ ] 完成一次 Tag 发布演练
- [ ] 用真实 CPA、OpenAI、自定义 Provider 完成端到端验收（**需要用户提供凭据**）

### 持续性工作

- [ ] 每次代码改动后同步更新 `docs/CONTEXT.md` 和 `docs/CHANGELOG.md`
- [ ] 重要技术决策追加到 `docs/DECISIONS.md`

## 已完成

- [x by codex 2026-10-01] 文档结构整理：建立多 agent 协作机制（AGENTS.md + 各 agent 入口 + docs/CONTEXT/TASKS/CHANGELOG/DECISIONS）
- [x by zcode 2026-10-01] 前端基线缺口修复与视觉重做（令牌单一来源、错误安全映射、表单 path/错误摘要、懒加载、960/1120 走查）
- [x by zcode 2026-10-01] 桌面构建复验与用户级 Go 1.25 工具链安装（wails3 build + 启动冒烟 + 生产包演示数据剔除检查）
- [x by zcode 2026-10-01] 提交工作区全部未提交改动（6 组方案经用户确认；backend/ci 两组为整理 codex 之前未提交的工作）

## 阻塞

- <暂无>
