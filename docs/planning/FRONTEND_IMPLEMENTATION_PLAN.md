# Tailwind CSS v4 + Naive UI 桌面前端实施计划

日期：2026-10-01。状态：设计与计划已整理，实施验收尚未完成。

## 1. 本阶段交付与边界

本阶段交付设计规范、接口边界、实施顺序和验收清单。**不安装依赖、不修改前端源码、不修改 Go API、不直接修改 Wails 自动生成绑定。** 保留 Wails 3 + Go，不迁移 Tauri。`sources/` 始终只读。

工作树已有 Vue 前端和其他未提交改动。本计划保留这些事实，不回滚现有工作，也不把已有部分实现视为完成验收。后续开始实施时先核对工作树，沿用可复用部分；不要重复安装、搭建或提交其他工作的文件。

配套文档：[`../architecture/TECH_STACK.md`](../architecture/TECH_STACK.md)、[`../architecture/UI_ARCHITECTURE_SPEC.md`](../architecture/UI_ARCHITECTURE_SPEC.md)、[`../design/UI_PROTOTYPE.md`](../design/UI_PROTOTYPE.md)。本计划中的混合样式边界取代旧文档里“只用原生 CSS、不引入 Tailwind”的结论。

原型用于页面层级和布局参考。若旧原型文字与本计划冲突，以本计划为准：校验失败先聚焦错误摘要，再由摘要定位字段；没有现有 API 支撑的设置不虚构交互。

## 2. 依赖与构建基线

| 分类 | 目标版本范围 |
| --- | --- |
| Vue | `vue ^3.5.43` |
| 路由 | `vue-router ^5.3.1` |
| 状态 | `pinia ^4.0.3` |
| 控件 | `naive-ui ^2.45.3` |
| 图标 | `lucide-vue-next ^1.0.0` |
| Vue Vite 插件 | `@vitejs/plugin-vue ^6.0.9` |
| Tailwind | `tailwindcss ^4.3.3`、`@tailwindcss/vite ^4.3.3` |
| Vue 类型检查 | `vue-tsc ^3.3.11` |

保留现有 TypeScript、Vite 与 Wails runtime。本阶段不调整它们的版本。正式实施时检查 Wails runtime 与 Go 侧锁定版本的兼容性，不能仅依赖 `latest` 的名称判断。当前基线的 `package.json` 使用 `@wailsio/runtime: latest`，lockfile 实际解析为 `3.0.0-beta.26`；发布前必须确认并固定这组版本，避免生产构建随 registry 漂移。

`frontend/vite.config.ts` 使用 `vue()` 与 `tailwindcss()`；构建脚本为 `vue-tsc --noEmit && vite build`。不创建 `tailwind.config.js` 或 PostCSS 配置。实施时使用 lockfile 固定实际解析版本。

当前 lockfile 中 `lucide-vue-next 1.0.0` 带有弃用提示，推荐包名为 `@lucide/vue`。本计划保留用户指定依赖；如要换包名，应作为明确的选型变更记录，不静默替换。

## 3. 样式与主题边界

Tailwind 负责壳层、导航、Grid/Flex、尺寸、间距、响应式布局及语义令牌；Naive UI 负责控件外观和交互。`DesktopShell` 自行组织，不使用后台管理模板。

`frontend/src/styles/tokens.css` 是颜色的唯一来源：

| CSS 变量 | Tailwind v4 映射 | 示例 |
| --- | --- | --- |
| `--surface-0` / `--surface-1` / `--surface-2` | `--color-surface-0` / `--color-surface-1` / `--color-surface-2` | `bg-surface-0`、`bg-surface-1` |
| `--surface-selected` | `--color-surface-selected` | `bg-surface-selected` |
| `--border-subtle` | `--color-border-subtle` | `border-border-subtle` |
| `--text-primary` / `--text-secondary` | `--color-text-primary` / `--color-text-secondary` | `text-text-primary` |
| `--accent` | `--color-accent` | `text-accent` |
| `--danger` | `--color-danger` | `text-danger` |

映射采用 `@theme inline`。仅导入 Tailwind 的 theme 与 utilities，不导入 Preflight。`base.css` 负责项目 reset、字体、focus、滚动条和 reduced-motion；`desktop.css` 只保留桌面特有视觉规则，不与 Tailwind 重复声明壳层布局。

`App.vue` 使用 `NConfigProvider`、`darkTheme`、`zhCN`、`dateZhCN`、`themeOverrides`，并嵌套 `NMessageProvider`、`NDialogProvider`、`NNotificationProvider`。

Naive UI 主题只通过公开 API 设置，不覆盖 `.n-*` 实现类。现有运行时曾因颜色计算不接受 `var(--accent)` 而无法挂载：实施时先加载 tokens.css，再读取 CSS 变量计算值，把实际颜色传给公开 themeOverrides。禁止在 theme.ts 中维护第二套手写十六进制颜色；验收要确认令牌改变后两套外观同步。

默认深色。字体优先本机 Segoe UI / 中文系统字体，数据采用 Consolas；如果选用 IBM Plex Sans / JetBrains Mono，必须确实随包分发并记录文件体积，不能用系统字体别名冒充已打包字体。普通文字（含次要文字、占位符）对背景至少 4.5:1，当前较暗的 muted 令牌需实测调整。

## 4. 壳层与页面

推荐窗口 1120 × 760，最小窗口 960 × 640。固定左导航 216 px、顶部工具栏 52 px，底部状态区固定；壳层约束在窗口高度内，只有主内容区滚动。浮层通过控件的公开挂载选项进入顶层，不受主区 overflow 裁剪。

| 路由 | 主要任务与动作 |
| --- | --- |
| `/config` | 当前 Provider / Model / Route 摘要；Provider 切换、更多菜单、模型管理、路由编辑与激活 |
| `/profiles` | 档案选择、新建、编辑、删除和备份恢复 |
| `/settings` | 应用自动启动、Codex 启动/状态、当前档案、数据恢复、关于 |

使用 memory history，启动显式进入 `/config`。顶部工具栏提供当前档案选择、刷新和 Codex 状态；底部展示操作主消息。首页遵循原型的“当前使用 + Provider 行列表”，不采用营销式 Hero、常驻编辑表单或多层卡片。Provider 没有可用模型/路由时说明下一步；不能把第一条未激活路由标成“当前使用”。

| 场景 | Naive UI 控件 |
| --- | --- |
| Provider / Model / Route / Profile 编辑 | `NModal` + `NForm` + `NFormItem` |
| 配置档案选择 | `NSelect` |
| Provider 更多菜单 | `NDropdown` |
| 自动启动 | `NSwitch` |
| 状态标签 | `NTag` |
| 初次加载 | `NSkeleton` |
| 删除与恢复确认 | `NDialogProvider` |
| 测试与激活结果 | `useMessage` / `useNotification` |
| 空列表 | `NEmpty` |

没有现有 API 的设置只能显示准确的只读事实或从本阶段范围移除；不得新增虚假的路径选择/设置保存操作，也不得为了 UI 改 Go 接口。

## 5. 数据与异步职责

调用链固定为 Vue Component → Pinia Store → `services/wails-api.ts` → 自动生成 binding → Go Application Service。组件仅导入 DTO 类型和 store，不运行适配层函数，不展示后端原始错误。

沿用 `ProviderDTO`、`ModelDTO`、`RouteDTO`、`ProfileDTO`，不改字段或复制新实体。唯一适配层保留：

| 资源/能力 | 后端方法 |
| --- | --- |
| Provider | `ListProviders` / `CreateProvider` / `SaveProvider` / `DeleteProvider` |
| Model | `ListModels` / `CreateModel` / `SaveModel` / `DeleteModel` |
| Route | `ListRoutes` / `CreateRoute` / `SaveRoute` / `DeleteRoute` |
| Profile | `ListProfiles` / `CreateProfile` / `SaveProfile` / `DeleteProfile` |
| 激活 | `ActivateRoute` |
| 测试 | `TestProvider` / `TestProviderModel` |
| Codex | `CodexRunning` / `StartCodex` |
| 自动启动 | `AutostartEnabled` / `SetAutostart` |
| 恢复 | `RestoreCodexConfig` |

Store 并行刷新四种资源，计算明确的已激活路由，管理 loading / saving / testing / activating / error。成功保存后刷新受影响资源；激活或恢复后刷新 Route、Profile、Codex 状态及摘要。保留旧列表，避免刷新闪烁。

异步动作返回明确的成功/失败结果，不能用 `void` API 返回的 undefined 判断失败。自动启动只在写入确实失败时回滚；写成功但刷新失败须报告“已保存，状态刷新失败”，不能声称保存失败。禁止冲突动作重入，过期读取不得覆盖新状态，成功消息不能被普通刷新覆盖。

`ActivateRoute` 当前是一次调用，前端不能凭计时模拟“正在备份/正在重启”。维护 activationStage（idle / updating / ready / error）与 recovered（true / false / unknown）；调用中的文案为“正在更新 Codex 配置”。只在现有后端契约明确报告恢复成功时使用 true；否则保留 unknown 并提供恢复入口。禁止通过错误中包含“恢复/restore”猜测成功，不新增后端阶段事件。

未知错误转换为安全、可操作的中文信息，分别建议检查地址/凭据引用、选择有效档案、重试同步或恢复备份。不将原始异常、凭据、完整请求参数放入 toast 或 DOM。

新建路由必须把用户选定的 restart_on_activate 保存下来；模型与 Provider 必须匹配且启用。新建档案的创建和路径保存若为两次调用，第二步失败后须保留已创建 ID 供重试，不能重复 CreateProfile。

## 6. 表单与键盘规范

- NFormItem 的 path 对应 model/rules 字段，真实 label 关联输入 ID；仅传 rules 而没有 path 不算校验完成。
- 提交校验失败时保留输入，在弹窗顶部显示 `tabindex="-1"` 错误摘要并聚焦；摘要可跳转对应字段。后端失败也在当前表单内反馈。
- NModal 开启自动聚焦、焦点圈定、Escape 关闭、关闭后恢复到触发控件；Provider/Profile 编辑禁止遮罩误关。脏表单关闭须确认。
- Enter 提交必须连接实际提交按钮/表单处理器，不能仅声明 submit 监听器。中文输入期间检查 isComposing 和组合输入状态；确认候选词时不提交。
- 测试与保存分开。保存/测试/激活按钮提供进行中反馈和防重入；删除/恢复确认说明影响，失败不能自动关闭确认框并误报成功。
- 导航为 nav，内容为 main；按钮、开关和 Select 有名称。Tab 顺序与视觉顺序一致，focus 可见。reduced-motion 下禁用位移与连续动画。

## 7. 后续实施顺序与退出条件

| 步骤 | 工作 | 退出条件 |
| --- | --- | --- |
| 1 | 核对依赖、Vite、令牌和主题注入 | 指定插件与脚本齐全；无 Preflight；无双份颜色；普通 Vite 可挂载 |
| 2 | 统一适配层与 store | 上述所有方法保留；组件无直接调用；void 成功、失败回滚、刷新失败分别验证 |
| 3 | 收敛 DesktopShell 与配置页 | Tailwind 管壳层；当前摘要真实；Provider 行列表可切换；主区独立滚动 |
| 4 | 编辑、档案、设置和反馈 | CRUD/测试/激活/恢复逐项可用；restart 和档案路径确实保存；确认框无误报 |
| 5 | 无障碍与窗口验收 | 下表交互、尺寸、缩放项目通过并保存证据 |
| 6 | Windows 生产构建与体积记录 | 本地资源检查、Wails 构建、启动计时和大小记录完整 |

不额外拆一套 shell store、workspace service 或基础控件包装目录，除非复用和独立状态实际需要。各步先更新验收记录和 CodeGraph，再仅提交本步文件；不混入已有后端、绑定或格式化工作。

## 8. 验证方法与记录模板

以下均为后续实施验收；计划完成不等于这些项目已通过。

| 项目 | 证明方法 | 当前证据状态 |
| --- | --- | --- |
| 类型与生产构建 | 单独运行 `vue-tsc --noEmit`、`npm run build`，保存退出码和产物 | 2026-10-01 通过；4431 modules，主 chunk 637.88 kB raw / 193.66 kB gzip，Profiles/Settings 已拆为懒加载 chunk（6.37 / 5.33 kB），仍有单 chunk 超过 500 kB 的 Vite 提示 |
| Wails Windows 构建 | 沿用项目 Windows task、生成 exe 并启动 WebView | 2026-10-01 复验：go1.25.14（用户级安装于 AppData\Local\go-sdk）+ wails3 v3.0.0-beta.26 生成 `bin/codex-provider-hub.exe`（约 11.4 MB，windowsgui），启动冒烟通过：WebView2 初始化成功、窗口渲染真实后端空态；NSIS/MSIX 安装包仍待工具链补齐 |
| 本地资源 | 检查生产 HTML/CSS/JS 加载入口，离线运行并检查资源请求；SVG namespace 和文档链接不等于远程加载 | 2026-10-01 静态检查仅发现随包字体和用户输入的 URL 占位符；浏览器预览演示数据（demo-data.ts）确认已被生产构建完整剔除（仅存占位文案），离线运行与 WebView 请求记录仍待做 |
| 令牌与字体单一来源 | 检查 `themeOverrides` 是否从 `tokens.css` 的计算值生成，且字体确实为系统字体或随包资源；不得维护第二套硬编码颜色 | 2026-10-01 已收敛：`theme.ts` 在 App setup 阶段读取 `tokens.css` 计算值生成 themeOverrides，无第二套硬编码颜色；字体改为真实系统字体栈（`--font-ui` = Segoe UI / 中文系统字体，`--font-mono` = Consolas），伪造的 IBM Plex / JetBrains Mono @font-face 已删除 |
| 960 × 640 / 1120 × 760 | 两尺寸下空态、长列表、弹窗、菜单、通知截图；检查壳层高度和滚动区域 | 2026-10-01 浏览器预览模式下 960×640 与 1120×760 截图走查通过，发现并修复行列表 grid 轨道 max-content 溢出；原生窗口验收仍待做 |
| Windows 125% / 150% | 原生 Windows/WebView 在两缩放下检查焦点、文字、浮层边界 | 未验收；浏览器 viewport 不能冒充 DPI 测试 |
| 固定导航/主区滚动 | 长列表滚动，比较工具栏、导航、状态区位置和主区 scrollTop | 2026-10-01 浏览器走查：导航、工具栏、底部状态区固定，仅主内容区滚动（截图证据）；原生窗口核对待做 |
| 浮层 | 两尺寸打开最长表单、底部菜单和通知，检查边界与可滚动内容 | 未完整验收 |
| Modal 焦点 | 记录打开时 activeElement、Tab/Shift+Tab 圈定、Esc、关闭后焦点 | 未完整验收 |
| 校验与保留输入 | 空值、空白、无效 URL、后端失败；错误摘要聚焦且输入不丢失 | 待做，已有规则声明不足以证明 |
| NForm 字段语义 | 逐项核对每个 `NFormItem` 的 `path`、label/id、错误摘要和失败后焦点；保存字段输入 | 2026-10-01 全部 NFormItem 已补显式 `path`；提交失败时表单顶部显示错误摘要（真实字段文案）、自动聚焦且点击条目可跳转对应字段（浏览器验证）；完整逐项核对仍待真实后端 |
| Route 创建语义 | 新建 Route 后核对 `restart_on_activate` 与用户输入一致，且不新增后端 API | 2026-10-01 已补齐：CreateRoute 后按用户勾选通过现有 SaveRoute 保存 `restart_on_activate`，未新增后端 API；真实后端一致性核对待做 |
| 激活恢复状态 | 仅使用后端明确证据设置 `recovered`；错误文案不能作为状态协议 | 2026-10-01 已移除错误文本解析：激活失败时 `recovered` 恒为 unknown，界面提示到「配置档案」页恢复备份；真实后端证据待接入 |
| 错误信息边界 | 将后端异常转换为安全、可操作的中文提示，不把原始异常直接放入界面 | 2026-10-01 `describeError` 已建立分类安全映射（网络/凭据/404/限流/档案/配置写入）；浏览器 401 用例验证 toast 与页脚仅显示安全提示，原始异常未泄漏 |
| 键盘与 IME | 键盘导航、Enter 提交、Escape 取消、中文候选确认不提交 | 未完整验收 |
| 对比度 | 计算文字对所有实际背景比率，含 muted、selected、占位符、控件 | 2026-10-01 计算通过：muted 调整为 `#8ca3a1`，对 surface-0/1/2/selected 分别为 6.90/6.35/5.75/4.70:1，secondary ≥ 5.94:1，primary ≥ 10.8:1；界面截图复核 |
| reduced-motion | 启用系统减少动效，检查实际过渡/动画不位移 | 声明存在；运行行为待验证 |
| Go API / binding 不变 | 实施差异与开始基线逐文件核对，不把他人改动算入本阶段 | 本阶段不修改；工作树已有其他改动 |

体积记录字段：构建版本/日期、所有 JS chunk 的 raw 字节和 gzip 字节、懒加载 chunk 名称与触发页面、CSS 体积、实际字体文件及总字节、Wails exe 字节、安装包字节。不得用 npm 解压大小代替性能证据。

WebView 首显时间从进程启动记录到主界面可交互，说明机器、Windows/DPI、冷/热启动条件和重复次数。至少重复 5 次，报告中位数和范围。未测量字段写“未测量”，不填 0。当前没有已验证的懒加载 chunk，后续可对 Profiles/Settings 使用路由级按需加载，再记录首屏与总包变化。

## 9. 完成定义

本阶段完成：三份文档对混合栈、接口/主题/页面边界、实施次序和验收方法保持一致，全部要求有明确后续验证路径；不要求本阶段安装或运行未来代码。

实施完成：第 8 节全部门槛有相应范围的实际证据，CRUD/测试/激活/恢复全链路通过，未测指标和恢复状态没有被猜测值替代。Windows 构建、DPI 或 WebView 测量缺失时仍属于未验收，不宣称 Windows 前端已交付。

## 10. 本阶段目标覆盖审计

本阶段审计对象是规范与计划。下表将用户要求逐项对应到可执行的约束及后续证据；“已覆盖”表示要求已纳入方案，不表示现有代码通过实施验收。

| 目标要求 | 规范位置 | 覆盖结论与后续证据 |
| --- | --- | --- |
| 保留 Wails 3 + Go，不改 Go API、DTO 或生成绑定；不安装依赖、不改前端源码 | 第 1、5、9 节 | 已覆盖；本阶段提交仅含文档，实施阶段按开始基线核对差异 |
| 九个指定依赖、Vue/Tailwind Vite 插件、类型检查后构建、不用旧版配置 | 第 2、7、8 节 | 已覆盖；后续检查 manifest/lockfile、配置和两个构建命令 |
| 全部指定颜色令牌、`@theme inline` 映射、无 Preflight、自有 reset 和公共主题 API | 第 3、8 节 | 已覆盖；后续检查生产 CSS、Naive UI 主题计算值及单一令牌来源 |
| darkTheme、zhCN、dateZhCN 及四个 Provider | 第 3 节；页面规范第 6 节 | 已覆盖；后续运行验证中文控件和通知/弹窗上下文 |
| 三页路由、自有 DesktopShell、固定导航/顶部/底部、主区独立滚动及控件映射 | 第 4、7、8 节 | 已覆盖；后续检查两种窗口尺寸、长列表及浮层边界 |
| 唯一适配层、四类资源 CRUD、激活/测试/Codex/自动启动/恢复全部方法及 Pinia 职责 | 第 5、7、8 节 | 已覆盖；后续逐项验证成功、失败、刷新失败和恢复未知状态 |
| 类型检查、生产/Wails 构建、资源全部本地加载 | 第 8 节；技术栈第 4、8 节 | 已覆盖；已有前端构建记录仅作基线，后续获取原生构建与离线资源请求证据 |
| 960 × 640、1120 × 760、125%/150% DPI | 第 4、8 节 | 已覆盖；浏览器尺寸检查与原生 DPI 检查分别记录 |
| Modal 焦点、Esc、圈定/恢复、错误摘要、保留输入、Enter/IME、键盘、对比度及减少动效 | 第 6、8 节 | 已覆盖；每项均指定行为验证方法，未测项目仍标为未验收 |
| JS raw/gzip、懒加载 chunk、字体、首显、exe/安装包大小 | 第 8 节 | 已覆盖；后续记录真实产物和机器/启动条件，不以 npm 解压体积代替性能 |

本阶段交付已具备后续实施顺序、退出条件和完整验收路径。尚未完成的实现与原生实测在第 8 节保留，不作为本阶段文档交付已完成的证据。
