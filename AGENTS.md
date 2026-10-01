# Agent 通用工作约定

本文件是本仓库所有 AI agent 协作的统一规则。任何 agent（Codex / Claude / Gemini / pi / zcode 等）进入本仓库开始工作前，**必须先完整阅读本文件**。

各 agent 的专属约定按需放在根目录对应的入口文件中（如 `CODEX.md`、`CLAUDE.md`、`GEMINI.md`、`PI.md`、`ZCODE.md` 等）。**不是所有 agent 都必须有专属文件**——只有当某个 agent 确实需要补充差异约定时才创建；没有专属文件的 agent 直接遵守本文件即可。专属文件与本文件冲突时以本文件为准。

---

## 一、开工前必读

每次会话开始，按顺序读取以下文件，建立对项目的最新认知：

1. **本文件**（`AGENTS.md`）— 通用规则
2. **你的专属入口**（如果存在，如 `CODEX.md` / `CLAUDE.md` / `GEMINI.md` / `PI.md` / `ZCODE.md` 等）— 差异化约定
3. **`docs/CONTEXT.md`** — 项目当前状态快照（最近在做什么、卡在哪、下一步）
4. **`docs/TASKS.md`** — 任务认领表，确认哪些任务已被其他 agent 认领
5. **`docs/planning/ROADMAP.md`** — 里程碑级规划
6. **`git status`** + **`git log --oneline -10`** — 仓库实际状态

发现上下文与仓库实际状态不一致时，以仓库代码 + `docs/planning/ROADMAP.md` 最新记录为准，并主动更新文档。

## 二、项目文档地图

```
proxy-switch/
├── AGENTS.md                       # 本文件：所有 agent 的通用规则
├── CODEX.md                        # Codex agent 专属约定
├── <其他 agent>.md                 # 按需存在（如 CLAUDE.md / GEMINI.md / PI.md / ZCODE.md 等）
├── README.md                       # 项目门面（人类视角）
└── docs/
    ├── README.md                   # 文档索引
    ├── product.md                  # 产品定位、核心模型、架构决策
    ├── CONTEXT.md                  # 当前状态快照（每次工作开始/结束都要更新）
    ├── TASKS.md                    # 任务认领与进度
    ├── CHANGELOG.md                # 变更日志（agent 工作记录）
    ├── DECISIONS.md                # 技术决策记录（ADR）
    ├── planning/                   # 规划类
    │   ├── ROADMAP.md              # 里程碑、目标、验收状态
    │   └── FRONTEND_IMPLEMENTATION_PLAN.md
    ├── architecture/               # 架构类
    │   ├── TECH_STACK.md
    │   └── UI_ARCHITECTURE_SPEC.md
    ├── design/                     # 设计类
    │   ├── UI_PROTOTYPE.md
    │   ├── UI_LIBRARY_COMPARISON.md
    │   └── ui-prototype.svg
    └── ops/                        # 运维类
        ├── BUILD_WINDOWS.md
        └── CI_CD.md
```

## 三、多 agent 协作机制

### 3.1 任务认领

- 开始任何非平凡任务前，在 `docs/TASKS.md` 把对应条目从 `[ ]` 改为 `[- by <agent名>]`，并注明日期。
- 任务格式：`[- by codex 2026-10-01] 补齐 NSIS 安装包`。
- 看到其他 agent 标记的 `[- by xxx]` 任务：**不要接管、不要修改、不要并行重做**。如有冲突，在 `docs/CONTEXT.md` 的"待协调事项"中提出。
- 完成任务后把标记改为 `[x by <agent名> <完成日期>]`。
- **"已完成"区只能由执行任务本人的 agent 写入**。不要代其他 agent 记录完成条目——你不掌握实际上下文，容易记错归属。发现历史条目归属有误时，**不要修改原条目**，在 `docs/CONTEXT.md` 的"待协调事项"中提出，由用户裁决。

### 3.2 变更日志

仓库**同时**依赖 `git log` 和 `docs/CHANGELOG.md`：

- **`git log` 是唯一事实源**，所有提交都必须遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范（见第四节）。
- **`docs/CHANGELOG.md` 是跨 agent 交接的展开版**，只记录"另一个 agent 接着干活时需要知道上下文"的改动。琐碎改动（typo、格式化、重命名、微调样式）**不写**，看 `git log` 就够了。

**什么时候必须写 CHANGELOG**：

- 开发了新需求 / 完成了任务（同步把 `TASKS.md` 对应条目标记为 `[x]`）
- 引入新技术 / 做了技术选型（同步按第 3.4 节追加 ADR 到 `DECISIONS.md`）
- 修改了公共 API、数据模型、配置格式
- 新增 / 调整了 agent 协作规则（修改 `AGENTS.md` 本身）
- 影响了其他 agent 可能正在用的文件或行为

**什么时候可以不写**：typo、格式化、重命名、代码风格调整、只动自己任务范围内的实现细节。

**格式**（与 Conventional Commits 对齐，未来可用 `git-cliff` 自动重建）：

```markdown
## YYYY-MM-DD · <agent名>

### <type>(<scope>): <subject>

- **影响**：<一句话说明这次改动的范围>
- **关键文件**：<文件路径列表>
- **后续**：<需要其他 agent / 人跟进的事项，可选>
- **commit**：<提交后补充 hash>
```

`<type>` 取值与 commit type 一致：`feat` / `fix` / `docs` / `refactor` / `test` / `chore` / `perf` / `build` / `ci`。

**操作顺序**：完成工作 → 更新 `TASKS.md` → 追加 `CHANGELOG.md`（先不填 commit hash） →（如涉决策）追加 `DECISIONS.md` → 更新 `CONTEXT.md` → 提交代码 → 回到 `CHANGELOG.md` 补上 commit hash（可作为单独 `docs:` 提交或下一次提交一并提交）。

### 3.3 状态快照

`docs/CONTEXT.md` 是项目当前状态的"单一事实源"（single source of truth）。

- **开始工作时**：先读它，了解最近在做什么。
- **结束工作时**：更新它，反映最新状态、卡点、下一步。
- 只保留**当前**状态，历史过程由 `git log` 和 `docs/CHANGELOG.md` 共同承担。

### 3.4 技术决策

任何影响架构、技术选型、模块边界的决策，在 `docs/DECISIONS.md` 追加一条 ADR（Architecture Decision Record）：

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

## 四、Git 规则

- 开始任务前执行 `git status`，确认工作区状态。
- 修改完成后检查 `git diff`，确认改动范围符合预期。
- 只暂存当前任务修改的文件；**禁止 `git add .` 和 `git add -A`**。
- **禁止 `git reset --hard`**、**`git clean -fd`**、**force push**。
- 未经用户明确要求不要执行 `git commit`。
- **不修改、不覆盖、不回滚其他 agent 或用户尚未提交的修改**。如有冲突，在 `docs/CONTEXT.md` 的"待协调事项"中提出，等待用户裁决。
- 提交信息**必须**遵循 [Conventional Commits](https://www.conventionalcommits.org/)：
  - 格式：`<type>(<scope>): <subject>`，scope 可选
  - `type` 取值：`feat`（新功能）/ `fix`（修复）/ `docs`（文档）/ `refactor`（重构）/ `test`（测试）/ `chore`（杂项）/ `perf`（性能）/ `build`（构建）/ `ci`（CI）
  - `subject` 用祈使句、不超过 72 字符
  - 示例：`feat(provider): add OpenAI Responses adapter`、`fix(codex): restore backup on partial write failure`、`docs(agents): clarify CHANGELOG rules`
- 提交粒度：一个提交只做一件事。文档更新和代码改动分开提交（除非文档是代码改动的直接说明）。

## 五、代码与测试

- 遵循 `docs/architecture/TECH_STACK.md` 和 `docs/architecture/UI_ARCHITECTURE_SPEC.md` 中的技术约定。
- 非平凡逻辑必须配套测试，参考 `internal/` 下已有的 `*_test.go` 模式。
- 修改公共 API、数据模型、配置格式时，必须同步更新 `docs/product.md` 和相关架构文档。
- 严格遵守 DDD 分层：`interfaces → application → domain`；`infrastructure → application/domain`。`domain` 不依赖任何基础设施、UI 或 Wails 类型。

## 六、安全与敏感信息

- API Key、Token、密码等敏感信息**永远不写入**代码、配置、日志、文档或提交记录。
- 凭据使用引用机制：环境变量或 Windows Credential Manager（见 `docs/product.md`）。
- 错误信息和日志不得包含完整密钥、请求头或响应体。

## 七、范围约束（V1）

- Windows only。
- Wails 3（当前锁定 `v3.0.0-beta.26`）+ Go。
- 只实现 Codex Desktop 平台适配器。
- Provider 只配置一次，通过 Route 选择 Platform、Provider、Model。
- 首版只接入 Responses API Provider。
- 不实现协议转换、自建代理、账户池、智能路由、云同步、插件系统。

任何超出 V1 范围的需求，**不要静默实现**，在 `docs/CONTEXT.md` 的"待协调事项"中提出，由用户决定放到 V1.1 / V2 还是独立 goal。

## 八、何时停下请求用户介入

以下情况必须暂停并向用户确认，**不要自行决策**：

- 需要用户提供真实凭据、API 地址或测试环境。
- 需要删除不可恢复的数据。
- 外部服务不可用，无法完成验证。
- 方案会改变 V1 产品边界或数据格式。
- 发现其他 agent 留下的修改与当前任务冲突，无法判断如何处理。

---

**违反以上规则的产出视为无效工作。** 规则不明确时，先记录疑问到 `docs/CONTEXT.md`，再用最保守、对工作区影响最小的方式推进。
