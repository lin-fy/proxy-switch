# 变更日志

> 跨 agent 协作的交接记录。规则见 [`../AGENTS.md`](../AGENTS.md) 第 3.2 节。
>
> **格式对齐 [Conventional Commits](https://www.conventionalcommits.org/)**，未来可通过 `git-cliff` 等工具从 git log 自动重建。
>
> 琐碎改动（typo、格式化、重命名变量）**不记**，直接看 `git log` 即可。

---

## 2026-10-01 · codex

### docs(agents): 建立多 agent 协作机制并明确提交前文档更新硬规则

- **影响**：所有 agent 协作流程
- **关键文件**：`AGENTS.md`、`CODEX.md`、`README.md`、`docs/README.md`、`docs/CONTEXT.md`、`docs/TASKS.md`、`docs/CHANGELOG.md`、`docs/DECISIONS.md`、`docs/product.md`（原 `PRODUCT_PLAN.md`）、`docs/planning/`、`docs/architecture/`、`docs/design/`、`docs/ops/`、`.agents/skills/review-pr/SKILL.md`
- **后续**：所有 agent 后续工作必须按 `AGENTS.md` 第 3.2 节顺序执行（TASKS → CHANGELOG → DECISIONS → CONTEXT → commit）
- **commit**：<提交后补充 hash>
