# Agent 通用工作约定

> 本文件只保留五条铁律，所有 agent 开工前必读。
> 完整规则、格式模板、详细说明见 [`docs/agent-guide/`](docs/agent-guide/)。

---

## 五条铁律

1. **开工前必读**：[`docs/CONTEXT.md`](docs/CONTEXT.md) + [`docs/TASKS.md`](docs/TASKS.md)，然后 `git status` + `git log --oneline -10`。
2. **不越权**：不动其他 agent 标记 `[- by xxx]` 的任务；不修改、不覆盖、不回滚他人未提交的改动；TASKS.md "已完成"区只能本人写入，发现归属有误记入 CONTEXT.md "待协调事项"，不改原条目。
3. **提交前更新文档**：完成工作 → 更新 TASKS.md → 追加 CHANGELOG.md →（如涉决策）追加 DECISIONS.md → 更新 CONTEXT.md → 再提交。提交信息必须遵循 [Conventional Commits](https://www.conventionalcommits.org/)。
4. **Git 禁区**：禁止 `git add .` / `git add -A` / `git reset --hard` / `git clean -fd` / force push / 未经用户允许 commit。允许 `git reset --soft` 撤销本地未推送提交。
5. **安全**：密钥、Token、密码不写入代码、配置、日志、文档或提交；凭据用引用机制（环境变量 / Windows Credential Manager）。

---

## 按需查阅

| 场景 | 文件 |
|---|---|
| 任务认领 / CHANGELOG / ADR 格式与模板 | [`docs/agent-guide/workflow.md`](docs/agent-guide/workflow.md) |
| Git 规则详细说明 | [`docs/agent-guide/git-rules.md`](docs/agent-guide/git-rules.md) |
| 代码风格 / 测试 / 架构约束 | [`docs/agent-guide/code-rules.md`](docs/agent-guide/code-rules.md) |
| V1 范围与产品边界 | [`docs/agent-guide/scope.md`](docs/agent-guide/scope.md) |
| 何时停下请求用户介入 | [`docs/agent-guide/escalation.md`](docs/agent-guide/escalation.md) |
| **模糊需求澄清流程** | [`docs/agent-guide/requirement-clarification.md`](docs/agent-guide/requirement-clarification.md) |
| 项目里程碑与验收状态 | [`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md) |
| 文档索引与字符规范 | [`docs/README.md`](docs/README.md) |

各 agent 的专属约定按需放在根目录 `<AGENT名>.md`（当前仅 `CODEX.md`），只补充差异不重复本文件。

---

**违反以上五条铁律的产出视为无效工作。** 规则不明确时，先记录到 `docs/CONTEXT.md` 的"待协调事项"，再用最保守、对工作区影响最小的方式推进。
