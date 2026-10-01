---
name: review-pr
description: Review Codex Provider Hub pull requests for correctness, Windows regressions, configuration safety, and V1 scope. Use when reviewing a PR, diff, or change before merge.
---

# Review a pull request

Read `AGENTS.md`, `CODEX.md`, and `docs/planning/ROADMAP.md` before reviewing the change. Inspect the complete base-to-head diff, then report only actionable findings with severity, file, line, impact, and a concrete fix.

Check the areas in `references/review-checklist.md`, prioritizing correctness and data-loss or credential risks over style. Run `scripts/review-checks.ps1` after dependencies are installed; pass `-BuildWindows` when the change affects Wails, generated bindings, packaging, or Windows behavior.

Do not treat a clean build as proof that configuration rollback, route activation, or credential handling is correct. If there are no actionable findings, state `No blocking findings` and list the checks that passed.
