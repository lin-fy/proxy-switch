# Codex Provider Hub：UI 页面框架与实现规范

> 状态：实现规范（部分已落地）  
> 配套原型：[`UI_PROTOTYPE.md`](UI_PROTOTYPE.md)  
> 目标：在不改变 Go 业务接口的前提下，把当前单文件 DOM 页面重构为可维护的 Vue 3 桌面工具界面。

## 1. 约束与目标

### 必须保留

- Wails 3 桌面壳和 Go 业务层。
- `frontend/bindings/...` 下的 Wails 自动生成绑定；生成文件不可手改。
- Provider、Model、Route、Profile 的现有业务语义。
- Codex 启动、连接测试、自动启动、备份恢复等后端能力。
- Windows 托盘和关闭窗口隐藏行为。

### 本次 UI 重构解决的问题

- `frontend/src/main.ts` 不再承担所有 HTML 拼接、状态、表单和事件分支。
- 首页从“全量配置表单”改为“当前配置 + Provider 切换列表”。
- 编辑操作进入弹窗上下文，页面保持稳定。
- 一级页面有明确边界，主内容区独立滚动。
- 视觉状态与异步状态统一，避免操作后只显示模糊的“操作完成”。

### 不在本阶段引入

- 不迁移到 Tauri；Wails 已经满足桌面壳需求。
- 不引入 SSR、Nuxt、Electron 或远程网页加载。
- 不添加后端数据库或新的领域对象。
- 不为每个 Provider 建立独立组件或独立 API 层。
- 不把所有通用组件抽成独立 npm 包。

## 2. 页面与路由模型

使用 Vue Router 5 的 `createMemoryHistory()` 管理一级页面。Memory history 不在桌面窗口显示 URL，重启后由本地设置恢复默认页即可。

```text
/config       配置切换首页（默认）
/profiles     Codex 配置档案
/settings     应用、Codex、数据和关于
```

页面级导航只有这三项。Provider 详情、添加模型、删除确认和恢复备份都属于当前页的覆盖层，不增加顶级路由。

### 导航规则

- `配置`是默认入口，进入时刷新一次汇总数据。
- 从列表打开编辑弹窗时保留背景页面滚动位置。
- 保存或删除成功后关闭弹窗，回到原列表位置并刷新对应资源。
- 直接触发激活不会改变当前页面。
- 切换页面时取消未完成的列表刷新请求；激活流程不可被普通页面切换取消。

## 3. 推荐目录结构

```text
frontend/src/
├── app/
│   ├── App.vue                    # 应用壳，只负责布局和页面出口
│   ├── router.ts                  # memory history + 页面路由
│   └── providers.ts               # 全局依赖装配
├── layouts/
│   └── DesktopShell.vue           # 导航栏、工具栏、状态栏、主内容槽
├── pages/
│   ├── ConfigPage.vue             # 当前路由 + Provider 列表
│   ├── ProfilesPage.vue           # Profile 列表与备份恢复
│   └── SettingsPage.vue           # 应用、Codex、数据、关于
├── features/
│   ├── providers/
│   │   ├── ProviderList.vue
│   │   ├── ProviderListItem.vue
│   │   ├── ProviderEditorDialog.vue
│   │   ├── ProviderModelList.vue
│   │   └── provider.types.ts
│   ├── routes/
│   │   ├── ActiveRouteSummary.vue
│   │   ├── RouteActivationButton.vue
│   │   └── route.types.ts
│   ├── profiles/
│   │   ├── ProfileList.vue
│   │   ├── ProfileEditorDialog.vue
│   │   └── profile.types.ts
│   └── settings/
│       ├── SettingsSection.vue
│       └── SettingRow.vue
├── stores/
│   ├── workspace.store.ts          # providers/models/routes/profiles 汇总
│   ├── shell.store.ts              # 当前页、当前 profile、消息、窗口状态
│   └── settings.store.ts           # autostart 等设置
├── services/
│   ├── wails-api.ts                # 生成绑定的唯一适配层
│   ├── workspace-service.ts        # 组合刷新和激活流程
│   └── error-message.ts            # 错误转用户可读文本
├── components/
│   ├── ui/                         # Naive UI 的业务级封装，不重复造基础控件
│   └── icons/                      # Lucide Vue 图标封装（如需统一尺寸）
├── styles/
│   ├── tokens.css                 # 语义颜色、尺寸、动效令牌
│   ├── base.css                   # reset、字体、focus、滚动条
│   └── desktop.css                # 壳层布局和桌面专用规则
├── main.ts                        # createApp、Pinia、Router、全局样式
└── vite-env.d.ts
```

组件边界按“业务功能”划分，不按视觉块堆一层层 `Card`。一个组件只有在拥有自己的状态、提交动作或可复用的交互时才单独拆分。基础控件由 Naive UI 提供，项目只封装业务语义和主题配置。

## 4. 组件树

```text
App
└── DesktopShell
    ├── AppBrand
    ├── PrimaryNav
    ├── Topbar
    │   ├── ProfileSwitcher
    │   ├── CodexStatus
    │   └── RefreshButton
    ├── RouterView
    │   ├── ConfigPage
    │   │   ├── ActiveRouteSummary
    │   │   ├── ProviderToolbar
    │   │   ├── ProviderList
    │   │   │   └── ProviderListItem × N
    │   │   └── StatusNotice
    │   ├── ProfilesPage
    │   │   └── ProfileList
    │   └── SettingsPage
    │       └── SettingsSection × N
    ├── GlobalDialogHost
    │   ├── ProviderEditorDialog
    │   ├── ProfileEditorDialog
    │   ├── ConfirmDialog
    │   └── RestoreBackupDialog
    └── AppStatusBar
```

`DesktopShell` 不知道 Provider 的字段和 Wails 方法名；它只消费 store 暴露的 `currentProfile`、`codexRunning`、`busy` 和导航动作。

## 5. 数据层和 Wails 边界

### 绑定原则

`services/wails-api.ts` 是生成绑定和 Vue 的唯一边界。页面和组件不得直接导入自动生成的 `app.ts`，这样生成绑定路径变化时只需改一处。

```ts
import * as generated from '../bindings/codex-provider-hub/internal/interfaces/wails/app'

export const wailsApi = {
  listProviders: () => generated.ListProviders(),
  listModels: () => generated.ListModels(),
  listRoutes: () => generated.ListRoutes(),
  listProfiles: () => generated.ListProfiles(),
  activateRoute: (routeId: string, profileId: string) =>
    generated.ActivateRoute(routeId, profileId),
  // 其余方法按同样方式映射，不重新定义 DTO 字段。
}
```

这段是接口形状示例，不是要求现在直接实现的代码。

### 现有 API 映射

| UI 能力 | Wails 方法 |
| --- | --- |
| 读取 Provider | `ListProviders` |
| 新增 / 保存 / 删除 Provider | `CreateProvider` / `SaveProvider` / `DeleteProvider` |
| 读取 / 保存 Model | `ListModels` / `ListModelsByProvider` / `CreateModel` / `SaveModel` / `DeleteModel` |
| 读取 / 保存 Route | `ListRoutes` / `CreateRoute` / `SaveRoute` / `DeleteRoute` |
| 激活 Route | `ActivateRoute(routeID, profileID)` |
| 读取 / 保存 Profile | `ListProfiles` / `CreateProfile` / `SaveProfile` / `DeleteProfile` |
| 测试 Provider / Model | `TestProvider` / `TestProviderModel` |
| Codex 状态 | `CodexRunning` / `StartCodex` |
| 自动启动 | `AutostartEnabled` / `SetAutostart` |
| 恢复备份 | `RestoreCodexConfig` |

### Store 责任

`workspace.store.ts` 负责：

- 并行拉取四种资源；
- 建立 `providerId → models` 的索引；
- 计算当前默认 Route、当前 Provider、当前 Model；
- 保持 `loading / refreshing / activating / error` 等状态；
- 在保存、删除、激活后刷新受影响的数据；
- 把后端错误转换成页面可展示的阶段和恢复建议。

`shell.store.ts` 负责：

- 当前一级页面；
- 当前 Profile 选择；
- 顶部短消息和 `role=alert` 错误消息；
- 弹窗打开状态和正在编辑的对象 ID。

组件只能调用 store 的动作，例如 `workspace.activateRoute(routeId)`，不在模板中拼接 `Promise.all` 或捕获原始错误。

## 6. Naive UI 接入规范

- `App.vue` 最外层使用 `NConfigProvider`，统一 `darkTheme`、`themeOverrides`、`zhCN` 和 `dateZhCN`；同时放置 `NMessageProvider`、`NDialogProvider`、`NNotificationProvider`。
- 使用具名导入和 Vite tree shaking，不全量注册 Naive UI；不把 Naive 的内部 class 当成业务 API。
- `NModal` / `NDrawer` 用于编辑上下文，默认启用自动聚焦和焦点圈定；Provider 编辑弹窗设置 `mask-closable=false`，关闭前检查脏表单。
- Provider / Profile 表单使用 `NForm`、`NFormItem` 和规则校验；错误摘要仍由页面提供，不能只依赖字段下方的红字。
- `useMessage()`、`useDialog()`、`useNotification()` 只能在对应 Provider 后代组件的 `setup()` 中调用；Pinia 和 service 不直接显示 UI 消息。
- 不使用 Naive 的全屏后台布局组件来生成页面壳层；壳层、导航、Provider 列表和状态摘要按 [`UI_PROTOTYPE.md`](UI_PROTOTYPE.md) 自有布局实现。
- 组件库只负责控件行为和基础外观，桌面视觉令牌集中在 `styles/tokens.css` 和 `NConfigProvider.theme-overrides`。

完整的组件库比较、版本依据和未实测项见 [`UI_LIBRARY_COMPARISON.md`](UI_LIBRARY_COMPARISON.md)。

## 7. 表单与弹窗规范

- 所有字段使用真实 `<label>` 和关联 `id`。
- 表单提交前做同步字段校验；后端错误显示在表单顶部，同时保留字段输入。
- 提交失败后将焦点移动到错误摘要，摘要项目可跳到具体字段。
- 弹窗打开时把焦点放到标题或第一个字段；关闭时恢复到打开弹窗的触发按钮。
- `Escape` 关闭可取消弹窗；有未保存输入时先显示确认。
- 删除必须使用可恢复路径或确认对话框，并说明受影响对象。
- 所有异步按钮有明确进行中状态，进行中不改变按钮尺寸。
- 测试连接和保存是两个动作，不合并成“保存并测试”。

## 8. 视觉与 CSS 实现规范

### 令牌优先

组件只使用语义令牌，例如 `var(--surface-2)`、`var(--text-secondary)`、`var(--accent)`，禁止在单个组件内写新的十六进制颜色。

### 布局规则

- 壳层使用 `display: grid`：`216px minmax(0, 1fr)`。
- 主内容区使用 `min-width: 0; min-height: 0; overflow: auto`，防止列表把窗口撑出边界。
- 列表使用行布局和分隔线，不嵌套多层卡片。
- 主要按钮使用实心强调色；次要动作使用低对比度边界；危险动作使用语义红色。
- 不使用页面级 radial gradient、装饰性网格、玻璃模糊或大面积投影。
- 自定义滚动条宽度保持 8 px，轨道透明，滑块使用 `border-subtle` 以上对比度。

### 图标规则

- 使用 `lucide-vue-next`，保持 16 px / 18 px / 20 px 三档尺寸和统一 1.75 px 描边。
- 图标按钮必须有 `aria-label` 或可见文字。
- 不使用 emoji 或 Unicode 字符作为系统图标。
- Provider 品牌没有可靠 SVG 时使用统一的首字母标记，不现场绘制品牌 Logo。

## 9. 异步流程状态机

```text
idle
 ├─ refresh → loading → ready | loadError
 ├─ create/edit → saving → ready | saveError
 ├─ test → testing → testPassed | testFailed
 └─ activate → backingUp → writing → restarting? → ready | activateError
```

每个状态只能有一个用户可读的主消息。`activateError` 必须带 `stage` 和 `recovered` 字段，页面据此显示“已恢复备份”或“需要手动恢复”。

## 10. 无障碍与桌面操作

- 导航使用 `<nav>`，主内容使用 `<main>`，弹窗使用 `role="dialog"` 和 `aria-modal="true"`。
- Tab 顺序必须与视觉顺序一致；不通过正数 `tabindex` 调整顺序。
- 所有焦点环都可见，焦点不能被固定工具栏或弹窗遮挡。
- 颜色不是唯一状态提示；激活状态同时使用标签、图标和文本。
- 支持 `Enter` 提交、`Escape` 关闭、方向键操作列表选择（如组件库提供原生支持）。
- `prefers-reduced-motion: reduce` 时不做位移和连续动画。
- 普通文字和背景对比度不低于 4.5:1；边界和焦点至少保持可辨识对比。

## 11. 迁移顺序（只描述实施边界）

1. 新建 Vue + Vite 入口和 `DesktopShell`，暂时保留现有 Wails 绑定。
2. 建立 `wails-api` 和 workspace store，先做到“读数据与错误状态不变”。
3. 实现配置页壳层、当前摘要和 Provider 列表。
4. 实现 Provider / Profile 弹窗，接入原有创建、保存、删除和测试方法。
5. 实现 Route 激活阶段状态和 Profile 页面。
6. 实现设置页，迁移自动启动、Codex 启动和恢复备份入口。
7. 删除旧 `main.ts` / 旧页面样式前，完成功能对照和窗口尺寸验收。

每一步都保持 Go API 不变；若发现 UI 需要新的后端能力，应先记录为 API 变更，而不是在组件中绕过服务层。
