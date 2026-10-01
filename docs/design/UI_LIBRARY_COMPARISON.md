# Vue 3 桌面工具组件库选型：Naive UI / Element Plus / Reka UI

> 调研日期：2026-09-30。范围：现有 Wails + Go 桌面应用的 Vue 3 前端；本次只更新设计规范。版本来自 npm 查询，能力来自官方文档/源码，不代表已在本项目运行验证。

## 结论

**本项目推荐 Naive UI。** 用它提供表单、弹窗、选择器和消息等控件，用项目自己的 `DesktopShell`、Provider 列表和状态摘要控制桌面布局。替换此前 Reka UI 方案，避免维护一套基础控件外观、校验和通知系统。

这里的“好”指更适合本项目的功能和维护成本，不是通用排名。用户已经选择 Vue 3，希望获得 CC Switch 那样的桌面工具体验；业务集中在配置编辑、切换、测试和恢复，尚不需要高度特殊的交互控件。

Naive UI、Element Plus 和 Reka UI 都是 Web UI 技术，放在 Wails 的 WebView 中运行。它们不会变成 WinUI 原生控件，也不负责托盘、系统标题栏或文件选择器。这些由 Wails / Go 提供。桌面感由固定窗口结构、信息密度和交互流程决定。

## 对比

| 维度 | Naive UI | Element Plus | Reka UI + 自有 CSS / shadcn-vue |
| --- | --- | --- | --- |
| 定位 | Vue 3 成品组件库 | Vue 3 成品组件库，管理系统常用 | Reka 为无样式交互原语；shadcn-vue 分发可修改的组件源码 |
| 本次 npm 版本 | 2.45.3 | 2.14.7 | Reka 2.10.5；shadcn-vue CLI 2.8.2 |
| Provider 编辑表单 | `NForm` / `NFormItem`，有校验规则与字段反馈 | `ElForm` 等，也有完整表单能力 | Reka 本身不提供等价完整业务表单校验；需要组合自行校验或其他表单方案 |
| 编辑与确认 | `NModal` / `NDrawer` / `NDialogProvider` | `ElDialog` / `ElDrawer` / `ElMessageBox` | Dialog / AlertDialog / Sheet 可组成；外观由项目负责 |
| 消息与通知 | `NMessageProvider` / `NNotificationProvider` | Message / Notification | shadcn-vue 可组合通知方案；需统一外观与生命周期 |
| 深色主题 | `darkTheme` + `NConfigProvider` | 深色 CSS 变量文件 + `html.dark` | 自有语义令牌或 shadcn-vue 主题变量 |
| 主题定制 | 类型化 `GlobalThemeOverrides`，全局及组件级配置 | CSS 变量，必要时 SCSS 变量 | 自由度最高，样式源码由项目持有 |
| 紧凑密度 | 组件 size 和高度令牌；有小尺寸选项 | 有 size 与主题变量 | 自行定义每个控件尺寸及状态 |
| 图标 | 可接 Lucide；无需使用默认推荐 xicons | 可接自选图标库 | 可接 Lucide |
| 可访问性 | Modal 有自动聚焦、焦点圈定、Esc 配置；组合页面仍须实测 | 也需检查页面组合和控件使用 | Reka 明确以 WAI-ARIA、键盘和焦点管理为核心原则；页面组合仍须实测 |
| 维护成本（本项目判断） | 较低：统一控件主题及表单 API | 较低；项目尚无 Element 资产可复用 | 较高：直接用 Reka 要设计维护外观；shadcn-vue 要维护复制进来的组件源码 |
| 最适用条件 | 中小型 Vue 工具，需要完整控件且保留自定义布局 | 团队熟悉 Element，或已有大量表格/后台组件 | 产品需要充分控制控件结构、特殊交互和独立设计系统 |

Element Plus 同样能够做好桌面界面；不存在“Element 必然像后台，Naive 必然像桌面”的技术结论。本项目选择 Naive，依据是用户偏好、完整控件范围、类型化主题配置和较低的定制负担。

Reka 的优势也保留：若未来出现 Naive 难以提供的组合交互，应先检查是否能通过 Naive 插槽解决，再评估局部替代；现在不混用两套组件库。

## 原型对应的 Naive UI 组件

| 原型部分 | 建议实现 | 边界 |
| --- | --- | --- |
| 固定导航、顶部工具栏、状态栏 | 自有 Grid/Flex + `NMenu` / `NButton` | 版式遵循原型，避免直接套完整后台模板 |
| 当前档案选择 | `NSelect` | 不把选择动作等同于激活 Codex 配置 |
| Provider 列表 | 自有语义列表 + `NButton` / `NTag` / `NDropdown` | 保留易扫描的名称、地址和模型信息；无需 `NDataTable` |
| Provider / Profile 编辑 | 受控 `NModal` + `NForm` / `NFormItem` | 完整字段编辑、脏表单保护、保存失败保留输入 |
| 模型列表与编辑 | 自有列表 + `NInput` / `NSwitch` / `NSelect` | 不为两三行列表添加分页与复杂数据表格 |
| 测试结果与失败信息 | `NAlert` + `NMessageProvider` | 错误需要可读原因，不能只显示红色或短暂 toast |
| 删除 / 恢复确认 | `NDialogProvider` | 显示影响范围，区分不可逆删除和恢复配置 |
| 设置开关 | `NSwitch` | 操作失败要恢复显示值，不能仅修改本地 UI |
| 首次加载 / 空态 | `NSkeleton` / `NEmpty` | 首次加载保持壳层，刷新时保留列表 |

## 接入时的关键约束

1. 在应用最外层使用 `NConfigProvider`，统一 `darkTheme`、`themeOverrides`、`zhCN` 和 `dateZhCN`。消息、对话框和通知 Provider 必须在这个配置作用域内。
2. `useMessage()` / `useDialog()` 只能在相应 Provider 的后代组件 `setup()` 中调用；不要在 Pinia 模块顶层、API service 或提供 Provider 的同一组件里直接注入。
3. 业务 store 返回结果或维护操作状态，组件负责消息展示。默认不使用 `createDiscreteApi`：官方文档说明其上下文和 DOM 容器独立，需要手动同步配置，混用容易产生主题不一致。
4. 编辑表单使用受控 `NModal`，保持焦点圈定；设置 `mask-closable=false` 避免点击背景丢输入。Esc、关闭按钮和取消按钮走同一套脏表单检查。
5. 使用具名组件导入，由 Vite/Rollup tree shaking 清理未用代码；不全量 `app.use(naive)` 注册。Naive 的控件样式由 css-render 按使用注入，不需要额外全局组件 CSS 文件。
6. 优先通过公开 `themeOverrides`、props 和 slots 定制，避免用 `:deep()` 批量改内部 `.n-*` 结构。布局 CSS 只管自己的窗口和业务区域。

## 体积、性能与验证边界

Naive 官方声明组件可 tree shake，npm 包含 date-fns、highlight.js、lodash 等依赖；这些依赖出现在安装清单里不等于全部进入本应用产物。Element 和 Reka 也不能仅按 npm 解压大小比较。

作为研究记录，2026-09-30 npm 元数据显示包解压体积约为：Naive UI 51.9 MB、Element Plus 44.0 MB、Reka UI 8.5 MB、shadcn-vue 1.3 MB。它们分别代表完整组件库、完整组件库、无样式原语和源码分发工具，口径不同；不能把这些数字当作 WebView 首屏大小或安装包大小。

**本次未安装或构建三套方案，因此不提供“Naive 更小/更快”的数字结论。** 正式实施时，在相同 Vue/Vite 基线、同等表单/弹窗/列表场景下记录：入口 JS 的 raw / gzip 大小、懒加载 chunk、字体体积、WebView 首次显示及交互时间。打包体积同时报告字体、Wails 可执行文件和安装包，避免把控件 JS 与整个桌面包混为一谈。

正式接入验收至少覆盖：

- 960×640、1120×760，Windows 125% / 150% 缩放；滚动只发生在内容区。
- NModal 打开、Tab 圈定、NSelect 下拉层、Esc、取消后的焦点恢复。
- 字段可访问名称、校验错误朗读、错误摘要聚焦；不能假设 `NFormItem` 显示 label 就满足所有关联要求。
- 深色主题在列表、弹窗、下拉菜单、message、dialog 上一致；文字与边界对比度实测。
- 中文 IME 输入时 Enter 不误提交；刷新和保存不清空草稿。
- 没有后端进度/恢复信息时只显示“正在切换”或“恢复状态待核实”，不由组件库虚构“备份完成”。

## 调研依据

- [Naive UI 官方 README：组件范围、tree shaking、主题、CSS 注入](https://github.com/tusen-ai/naive-ui/blob/main/README.md)
- [Naive UI 官方主题文档](https://github.com/tusen-ai/naive-ui/blob/main/demo/pages/docs/customize-theme/enUS/index.md)
- [Naive UI Config Provider 文档：主题、语言、component-options](https://github.com/tusen-ai/naive-ui/blob/main/src/config-provider/demos/enUS/index.demo-entry.md)
- [Naive UI Modal 文档：auto-focus / close-on-esc / trap-focus](https://github.com/tusen-ai/naive-ui/blob/main/src/modal/demos/enUS/index.demo-entry.md)
- [Naive UI Form 文档：规则、label-props、validate](https://github.com/tusen-ai/naive-ui/blob/main/src/form/demos/enUS/index.demo-entry.md)
- [Naive UI 离散 API 文档：独立上下文约束](https://github.com/tusen-ai/naive-ui/blob/main/src/discrete/demos/zhCN/index.demo-entry.md)
- [Element Plus 主题](https://element-plus.org/en-US/guide/theming.html) / [深色模式](https://element-plus.org/en-US/guide/dark-mode.html)
- [Reka UI 定位、可访问性及无样式原语](https://reka-ui.com/docs/overview/introduction)
- [shadcn-vue 源码分发与组件所有权](https://www.shadcn-vue.com/docs/introduction.html)
- npm 元数据：[`naive-ui`](https://www.npmjs.com/package/naive-ui)、[`element-plus`](https://www.npmjs.com/package/element-plus)、[`reka-ui`](https://www.npmjs.com/package/reka-ui)、[`shadcn-vue`](https://www.npmjs.com/package/shadcn-vue)。官方 main 文档可能含尚未发布字段，实现时以锁定版本的类型声明为准。
