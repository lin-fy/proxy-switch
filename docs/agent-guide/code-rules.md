# 代码与架构规则

> [`../../AGENTS.md`](../../AGENTS.md) 铁律之外的技术约束。

## 技术栈约定

- 遵循 [`../architecture/TECH_STACK.md`](../architecture/TECH_STACK.md) 和 [`../architecture/UI_ARCHITECTURE_SPEC.md`](../architecture/UI_ARCHITECTURE_SPEC.md)
- Wails 3（锁定 `v3.0.0-beta.26`）+ Go + Vue 3 + TypeScript + Pinia + Naive UI + Tailwind v4
- 不引入新的 UI 组件库、状态管理库或 CSS 框架

## DDD 分层

严格的依赖方向：

```
interfaces → application → domain
infrastructure → application/domain
```

- `domain` 不依赖任何基础设施、UI 或 Wails 类型
- `application` 只编排用例，通过端口/接口访问外部能力
- `infrastructure` 实现配置文件、凭据、HTTP、进程和 Codex 适配等外部能力
- `interfaces` 只负责 Wails 绑定、DTO 转换和错误呈现，不承载业务规则
- `cmd` 是组合根，只组装依赖和启动程序

## 测试

- 非平凡逻辑必须配套测试，参考 `internal/` 下已有的 `*_test.go` 模式
- 完成 Go 改动后默认运行 `go test ./...` 和 `go vet ./...`
- 完成前端改动后默认运行 `npm run lint`、`npm run format:check`、`npm run build`（在 `frontend/`）
- 验证失败时不要提交

## 文档同步

修改以下内容时**必须**同步更新对应文档：

| 改动 | 同步更新 |
|---|---|
| 公共 API、数据模型、配置格式 | `docs/product.md` + 相关架构文档 |
| 技术选型、架构决策 | `docs/DECISIONS.md`（新增 ADR） |
| 里程碑进度 | `docs/planning/ROADMAP.md` |

## 字符与格式

- 所有文档遵守 `docs/README.md` 的字符与格式约定
- 编辑器配置由 `.editorconfig` 声明（UTF-8 / LF / 末尾换行）
