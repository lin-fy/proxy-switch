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

## 并行开发流程（多 agent 同时干活）

当多个 agent 需要同时开发时，使用 git worktree 隔离工作区。

### 核心原则

- **代码改动**：在 agent 自己的工作树，提交到 agent 自己的分支
- **文档更新**（TASKS/CONTEXT/CHANGELOG/DECISIONS）：**只在主工作树**，提交到 dev 分支
- **合并**：任务验收通过后，由负责该任务的 agent 在主工作树自行合并到 `dev`，无需用户执行或再次确认

### 工作树结构

```
H:/code/proxy-switch              # 主工作树（dev 分支）
├── 用途：文档更新、最终合并
└── 提交：docs(tasks): ... / docs(collab): ...

H:/code/proxy-switch-codex        # codex 的工作树（codex/<任务> 分支）
├── 用途：代码开发
└── 提交：feat(...): ... / fix(...): ...

H:/code/proxy-switch-zcode        # zcode 的工作树（zcode/<任务> 分支）
├── 用途：代码开发
└── 提交：feat(...): ... / fix(...): ...
```

### 命名规范

- **工作树路径**：`../proxy-switch-<agent名>`（如 `../proxy-switch-codex`）
- **分支名**：`<agent名>/<任务简称>`（如 `codex/m5-nsis`、`zcode/m4-dpi-fix`）

### 完整流程

#### 1. 认领任务（在主工作树）

```bash
cd H:/code/proxy-switch
# 编辑 docs/TASKS.md，把任务从 [ ] 改为 [- by <agent名> <日期>]
git add docs/TASKS.md
git commit -m "docs(tasks): <agent名> claims <任务名>"
```

#### 2. 创建工作树（在主工作树）

```bash
git worktree add ../proxy-switch-<agent名> -b <agent名>/<任务简称>
```

#### 3. 开发代码（在自己的工作树）

```bash
cd ../proxy-switch-<agent名>
# 改代码...
git add <具体文件>
git commit -m "feat(<scope>): <改动说明>"
# 可以多次提交
```

#### 4. 完成后更新文档（回到主工作树）

```bash
cd H:/code/proxy-switch
# 编辑 docs/TASKS.md，把任务从 [- by <agent名>] 改为 [x by <agent名> <日期>]
# 追加 docs/CHANGELOG.md（格式见 agent-guide/workflow.md）
# 更新 docs/CONTEXT.md（如有需要）
git add docs/TASKS.md docs/CHANGELOG.md docs/CONTEXT.md
git commit -m "docs(collab): record <agent名> <任务简称> completion"
```

#### 5. 合并前检查

agent 确认主工作树处于 `dev`、工作区干净，并检查待合并差异和任务验收结果。不得覆盖他人未提交的改动。

#### 6. agent 自行合并（在主工作树）

```bash
cd H:/code/proxy-switch
git merge --no-ff <agent名>/<任务简称> -m "<type>(<scope>): merge <任务简称>"
git status --short
```

合并后检查结果，并向用户报告合并提交和验证情况。推送远端仍按用户授权执行。出现涉及他人工作或需求取舍的冲突时，按冲突处理规则协调，不覆盖或回滚他人改动。

#### 7. 清理工作树（可选，由所属 agent 执行）

仅清理自己的、工作区干净且分支已合并的工作树；不得强制删除。

```bash
git worktree remove ../proxy-switch-<agent名>
git branch -d <agent名>/<任务简称>
```

### 规则

- 不要改其他 agent 的工作树
- 不要在自己的工作树更新 docs/ 下的协作文档（TASKS/CONTEXT/CHANGELOG/DECISIONS）
- 协作文档只在主工作树更新
- 代码提交用 Conventional Commits（feat/fix/refactor 等）
- 文档提交用 `docs(tasks):` / `docs(collab):`

### 示例

```bash
# zcode 认领 M4 DPI 验收任务
cd H:/code/proxy-switch
# 编辑 TASKS.md: [- by zcode 2026-10-02] M4 DPI 验收
git add docs/TASKS.md
git commit -m "docs(tasks): zcode claims M4 DPI acceptance"

# zcode 建工作树
git worktree add ../proxy-switch-zcode -b zcode/m4-dpi-fix

# zcode 在自己的工作树开发
cd ../proxy-switch-zcode
# ... 改代码 ...
git add frontend/src/pages/ConfigPage.vue
git commit -m "fix(ui): correct DPI scaling on 960x640 window"

# zcode 完成后回到主工作树更新文档
cd H:/code/proxy-switch
# 编辑 TASKS.md: [x by zcode 2026-10-02]
# 追加 CHANGELOG.md
git add docs/TASKS.md docs/CHANGELOG.md
git commit -m "docs(collab): record zcode m4-dpi-fix completion"

# zcode 检查主工作树状态和任务验收结果后自行合并
git status --short
git merge --no-ff zcode/m4-dpi-fix -m "fix(ui): merge M4 DPI acceptance"
git status --short
# 向用户报告合并结果；推送远端按用户授权执行

# zcode 清理自己的已合并、干净工作树（可选）
git worktree remove ../proxy-switch-zcode
git branch -d zcode/m4-dpi-fix
```

---

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
