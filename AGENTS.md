# Agent 通用工作约定

> 本文件只保留五条铁律，所有 agent 开工前必读。
> 完整规则、格式模板、详细说明见 [`docs/agent-guide/`](docs/agent-guide/)。

---

## 五条铁律

1. **开工前必读**：[`docs/CONTEXT.md`](docs/CONTEXT.md) + [`docs/TASKS.md`](docs/TASKS.md)，然后 `git status` + `git log --oneline -10`。
2. **不越权**：不动其他 agent 标记 `[- by xxx]` 的任务；不修改、不覆盖、不回滚他人未提交的改动；TASKS.md "已完成"区只能本人写入，发现归属有误记入 CONTEXT.md "待协调事项"，不改原条目。
3. **提交前更新文档**：任务 agent 在自己的分支完成代码和验证后，更新 TASKS.md → 追加 CHANGELOG.md →（如涉决策）追加 DECISIONS.md → 更新 CONTEXT.md，再创建 GitHub PR。多个 agent 并行时由用户指定 reviewer/merger；只有一个活跃 agent 时，该 agent 可自审并在分支保护允许时合并 PR，但仍必须通过 CI。提交信息必须遵循 [Conventional Commits](https://www.conventionalcommits.org/)。
4. **Git 禁区**：禁止 `git add .` / `git add -A` / `git reset --hard` / `git clean -fd` / force push / 未经用户允许 commit。允许 `git reset --soft` 撤销本地未推送提交。
5. **安全**：密钥、Token、密码不写入代码、配置、日志、文档或提交；凭据用引用机制（环境变量 / Windows Credential Manager）。

## 模型无关验证协议

项目验证不依赖 Codex、Pi、Claude Code 或任何特定模型。完成改动后按范围运行统一入口：

```powershell
task check:frontend   # 前端格式、lint、类型检查和生产构建
task check:backend    # Go 格式、模块、vet 和测试
task ci               # 全部检查 + Windows Wails 构建与 portable ZIP
```

GitHub Actions 复用 `scripts/ci/verify.ps1`；CI 本身不调用 LLM。失败诊断或 AI Review 属于独立的可选工作流，不能替代这些硬检查。

如果本机没有安装 Go Task，可直接调用同一实现：`pwsh -NoProfile -File ./scripts/ci/verify.ps1 -Scope frontend`、`-Scope backend` 或 `-BuildWindows`。

## Wails v3 beta 前端开发规则

Wails v3 仍处于 beta，前端开发必须先核对上游最新 beta 版本，再开始修改 Vue/Vite、Wails 绑定或桌面壳代码。Go 模块、`wails3` CLI、`@wailsio/runtime`、GitHub Actions 和构建文档必须使用同一个具体版本；禁止混用旧版 CLI/runtime，也不要在提交的依赖中使用浮动的 `latest`。

发现上游有更新时，先在同一分支完成版本对齐、重新生成 `frontend/bindings`，再实施前端功能改动，并运行 `task check:frontend` 和 `task ci`。当前规则只要求先完成版本核对；若升级涉及 API 或生成文件变化，必须单独记录并通过 PR 验证。

---

## 并行开发流程（多 agent 同时干活）

当多个 agent 需要同时开发时，使用 git worktree 隔离工作区。

### 核心原则

- **代码改动**：在 agent 自己的工作树，提交到 agent 自己的分支
- **文档更新**（TASKS/CONTEXT/CHANGELOG/DECISIONS）：任务相关更新在 agent 分支完成并随 PR 审查
- **合并**：多个 agent 并行时由用户指定 reviewer/merger；只有一个活跃 agent 时，该 agent 可自审并在分支保护允许时合并自己的 PR。任何情况下都必须通过 PR，禁止直接在主工作树合并分支

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

#### 1. 认领任务（由 root/Codex 在主工作树）

```bash
cd H:/code/proxy-switch
# root/Codex 编辑 docs/TASKS.md，把任务从 [ ] 改为 [- by <agent名> <日期>]
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

开发期间允许 agent 分支暂时落后于 `dev`。其他 agent 合并代码不会自动改动当前工作树。

#### 3.5 合并前同步版本（在自己的工作树）

准备合并时，先在 agent 自己的工作树同步主分支：

```bash
cd ../proxy-switch-<agent名>
git status --short
# 确认当前修改已保存并且工作区干净
git merge dev
# 如有冲突，只在当前 agent 分支解决
# 解决后重新运行本任务的测试和构建
```

同步后的分支必须重新验证通过，才能创建 PR。不得用重置、删除文件或覆盖他人修改的方式解决版本落后问题。

#### 4. 完成代码和文档（在自己的工作树）

```bash
cd ../proxy-switch-<agent名>
git status --short
git log --oneline -n 3
# 更新 TASKS.md、CHANGELOG.md、DECISIONS.md（如有需要）、CONTEXT.md
# 将自己的任务改为 [x by <agent名> <完成日期>]
git add <当前任务涉及的具体文件>
git commit -m "docs(collab): record <agent名> <任务简称> completion"
```

任务 agent 在 PR 描述中报告分支、提交、验证命令和结果；公共文档随 PR 一起审查。多个 agent 并行时用户人工指定 reviewer 和 merger；只有一个活跃 agent 时，任务 agent 可自审并在分支保护允许时合并，但必须等待 CI 通过。

#### 5. 创建 PR

任务 agent 确认自己的分支已经合并最新 `dev` 并完成验证，然后创建目标为 `dev` 的 GitHub PR。PR 描述必须包含改动范围、验证结果和未完成事项。

#### 6. 审查并合并 PR

多个 agent 并行时，用户指定的 reviewer 检查 PR 差异、验证记录和目标分支最新状态，指定的 merger 在审查通过后于 GitHub 合并 PR。只有一个活跃 agent 时，该 agent 可以执行同样的自审流程，并在分支保护允许时合并自己的 PR；如果分支保护要求独立批准，仍需由用户指定 reviewer。未通过或有冲突时退回 agent 分支处理，不能直接覆盖或绕过审查。

合并后确认 `dev` 状态和 CI 结果，再向用户报告合并提交。远端推送和发布仍按用户授权执行。

#### 7. 清理工作树（可选，由所属 agent 执行）

仅清理自己的、工作区干净且分支已合并的工作树；不得强制删除。

```bash
git worktree remove ../proxy-switch-<agent名>
git branch -d <agent名>/<任务简称>
```

### 规则

- 不要改其他 agent 的工作树
- 任务相关的 docs/ 协作文档在自己的分支更新并随 PR 审查
- 不要直接在主工作树合并分支或绕过 PR 修改任务相关协作文档
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

# zcode 合并前同步主分支并重新验证
git merge dev
# 解决冲突后重新运行验收

# zcode 更新 TASKS/CHANGELOG/CONTEXT 并创建 GitHub PR（目标 dev）
# 多 agent 时由用户指定 reviewer/merger；单 agent 时 zcode 自审，等待 CI，并在分支保护允许时合并

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
