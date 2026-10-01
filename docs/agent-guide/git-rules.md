# Git 规则详细说明

> [`../../AGENTS.md`](../../AGENTS.md) 铁律 4 的展开说明。

## 禁区（违反视为无效工作）

- `git add .` / `git add -A`
- `git reset --hard`
- `git clean -fd`
- force push（`git push --force` / `--force-with-lease`）
- 未经用户明确要求的 `git commit`

## 允许但谨慎

- `git reset --soft`：仅用于撤销**本地未推送**的提交（例如提交分组错误时重新组织）。使用前必须确认该提交未 push 到远程。

## 必须遵守

- 开始任务前 `git status`，确认工作区状态
- 修改完成后 `git diff`，确认改动范围符合预期
- 只暂存当前任务修改的文件
- 不修改、不覆盖、不回滚其他 agent 或用户尚未提交的改动
- 开发期间允许 agent 分支暂时落后于 `dev`；准备合并时必须在自己的 worktree 执行 `git merge dev`，解决冲突并重新验证
- 并行任务验收通过后，由任务 agent 创建目标为 `dev` 的 GitHub PR；用户指定 reviewer/merger 审查并合并，未指定前不得合并。远端推送仍按用户授权执行。

## 提交信息规范

严格遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <subject>
```

- `type`：`feat`（新功能）/ `fix` / `docs` / `refactor` / `test` / `chore` / `perf` / `build` / `ci`
- `scope`：可选，模块名（如 `provider` / `codex` / `ui` / `agents` / `ci`）
- `subject`：祈使句，≤ 72 字符，首字母小写，结尾不加句号

示例：
- `feat(provider): add OpenAI Responses adapter`
- `fix(codex): restore backup on partial write failure`
- `docs(agents): clarify CHANGELOG rules`

## 提交粒度

- 一个提交只做一件事
- 文档更新和代码改动**分开提交**（除非文档是代码改动的直接说明）
- 多个逻辑无关的改动必须拆成多个提交

## 冲突处理

发现其他 agent 留下的修改与当前任务冲突时：
1. 不要擅自处理
2. 在 `docs/CONTEXT.md` "待协调事项"中描述冲突
3. 等待用户裁决

分支版本落后不属于冲突。先在 agent 自己的 worktree 合并最新 `dev`；只有合并后无法判断或无法安全解决的代码冲突，才按上述流程暂停协调。
