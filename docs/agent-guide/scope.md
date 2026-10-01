# V1 范围约束

> [`../../AGENTS.md`](../../AGENTS.md) 之外的产品边界说明。任何超出 V1 范围的需求**不要静默实现**。

## V1 锁定

- Windows only
- Wails 3（`v3.0.0-beta.26`）+ Go
- 只实现 Codex Desktop 平台适配器
- Provider 只配置一次，通过 Route 选择 Platform、Provider、Model
- 首版只接入 Responses API Provider
- 不与 CC Switch 同时管理 Codex 配置

## V1 不做

- 协议转换、自建代理、账户池、智能路由
- 云同步、插件系统
- 自动更新（V1.1）
- 其他平台（Claude Code / Gemini CLI / OpenCode，V2）

## 范围变更流程

任何超出 V1 范围的需求：

1. **不要静默实现**
2. 在 `docs/CONTEXT.md` "待协调事项"中提出，说明需求、影响、建议归属（V1.1 / V2 / 独立 goal）
3. 等待用户裁决

## V1 完成标准

见 [`../planning/ROADMAP.md`](../planning/ROADMAP.md) 第 1 节 "Definition of Done"。

**核心判定**：文件写入成功 ≠ V1 完成。必须跑通完整链路：

```
Proxy Switch → 选择 Route → 备份并更新 Codex 配置 → 启动 Codex Desktop → 实际请求模型成功
```
