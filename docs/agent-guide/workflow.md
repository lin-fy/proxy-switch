# 协作流程详细规则

> 本文档是 [`../../AGENTS.md`](../../AGENTS.md) 铁律 1~3 的展开说明。涉及任务认领、CHANGELOG、ADR、CONTEXT 的格式与模板。

## 一、开工前完整流程

铁律 1 是最低要求。完整流程按顺序：

1. `AGENTS.md`（五条铁律）
2. 你的专属入口（如果存在，如 `CODEX.md`）
3. `docs/CONTEXT.md`
4. `docs/TASKS.md`
5. `git status` + `git log --oneline -10`
6. 需要了解里程碑时读 `docs/planning/ROADMAP.md`
7. 需要写 CHANGELOG / ADR 时回到本文件查格式

发现上下文与仓库实际状态不一致时，以仓库代码 + `docs/planning/ROADMAP.md` 最新记录为准，并主动更新文档。

并行开发时，agent 分支在开发期间可以暂时落后于 `dev`。准备合并前，必须在自己的 worktree 执行 `git merge dev`，在该分支解决冲突并重新验证；验证通过后才能合并到主工作树。

## 二、任务认领

- 开始非平凡任务前，在 `docs/TASKS.md` 把对应条目从 `[ ]` 改为 `[- by <agent名> <日期>]`。
- 格式：`[- by codex 2026-10-01] 补齐 NSIS 安装包`。
- 看到其他 agent 的 `[- by xxx]`：**不接管、不修改、不并行重做**。冲突记入 `docs/CONTEXT.md` "待协调事项"。
- 完成改为 `[x by <agent名> <完成日期>]`。
- **"已完成"区只能本人写入**。发现归属有误：不改原条目，记入 `docs/CONTEXT.md` "待协调事项"，由用户裁决。

## 三、CHANGELOG 规则

仓库同时依赖 `git log` 和 `docs/CHANGELOG.md`：

- `git log` 是唯一事实源（Conventional Commits 强制）。
- `docs/CHANGELOG.md` 是跨 agent 交接的展开版，只记"另一个 agent 接手时需要上下文"的改动。

**必须写 CHANGELOG**：
- 开发新需求 / 完成任务（同步更新 TASKS）
- 引入新技术 / 技术选型（同步写 ADR）
- 修改公共 API、数据模型、配置格式
- 新增/调整 agent 协作规则
- 影响其他 agent 正在用的文件或行为

**不写**：typo、格式化、重命名、代码风格、只动自己任务范围的实现细节。

**格式**（对齐 Conventional Commits，未来可用 `git-cliff` 重建）：

```markdown
## YYYY-MM-DD · <agent名>

### <type>(<scope>): <subject>

- **影响**：<一句话说明范围>
- **关键文件**：<文件路径列表>
- **后续**：<需要跟进的事项，可选>
- **commit**：<提交后补充 hash>
```

`<type>` 取值：`feat` / `fix` / `docs` / `refactor` / `test` / `chore` / `perf` / `build` / `ci`。

**操作顺序**：完成工作 → TASKS → CHANGELOG（hash 留空）→ DECISIONS（如涉决策）→ CONTEXT → 提交 → 接力回填 hash。

**接力回填**：本次文档提交 hash 留空，下次任何 agent 提交文档时顺手补回上次的 hash。

## 四、CONTEXT.md 规则

`docs/CONTEXT.md` 是项目当前状态的"单一事实源"。

- 开始工作时读它，结束工作时更新它
- 只保留**当前**状态，历史由 `git log` + `CHANGELOG.md` 承担
- 保持精简（目标 < 800 token），只回答"现在在做什么、卡在哪、下一步是什么"

## 五、ADR 格式

任何影响架构、技术选型、模块边界的决策，在 `docs/DECISIONS.md` 追加：

```markdown
## ADR-XXXX：<决策标题>

- 日期：YYYY-MM-DD
- 决策人：<agent名或人名>
- 状态：proposed / accepted / superseded by ADR-YYYY
- 背景：<为什么需要决策>
- 决策：<选了什么>
- 理由：<为什么选它>
- 后果：<带来的影响和约束>
```

编号从 ADR-0001 递增，不重用。
