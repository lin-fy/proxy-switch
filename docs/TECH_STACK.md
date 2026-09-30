# Codex Provider Hub：桌面工具技术栈

> 状态：技术选型基线（UI 迁移进行中）  
> 目标：让前端看起来像桌面配置工具，同时保持现有 Wails + Go 业务链路，不把应用迁移成远程网页。

## 1. 最终推荐栈

| 层 | 选择 | 说明 |
| --- | --- | --- |
| 桌面壳 | Wails 3 `v3.0.0-beta.26` | 当前项目已经使用，负责窗口、托盘、开机启动、Go 服务和本地资源嵌入 |
| 后端 | Go | 保留现有 DDD/application/infrastructure 分层与 CodexAdapter |
| UI 框架 | Vue 3.5.x + `<script setup>` | 用组件和响应式状态替代当前单文件 DOM 拼接 |
| 语言 | TypeScript 5.6.x | 沿用现有严格类型设置和 Wails 生成 DTO |
| 构建 | Vite 7.x | 当前项目已经使用，保留开发服务器和生产构建命令 |
| 页面导航 | Vue Router 5.x + `createMemoryHistory` | 桌面窗口内的一级页面，不在界面暴露 URL |
| 全局状态 | Pinia 4.x | 管理 Provider / Model / Route / Profile 汇总和激活状态 |
| UI 组件库 | Naive UI 2.45.x | 提供 Dialog、Form、Select、Message、Notification、Skeleton 等完整控件；用主题覆盖适配桌面令牌 |
| 图标 | `lucide-vue-next` 1.x | 统一 SVG 图标风格、尺寸和可访问名称 |
| 样式 | 原生 CSS + Naive UI `themeOverrides` | 自有桌面工具布局和语义令牌，不引入 Tailwind 的大量工具类 |
| 字体 | 本地打包 IBM Plex Sans + JetBrains Mono | 避免运行时请求 Google Fonts；UI 和模型/URL 数据分开用字形 |
| 测试 | 现有 Go 测试；后续给 store 增加 Vitest | 只测试异步状态和数据转换，不为静态模板堆快照测试 |

## 2. 为什么不换 Tauri

CC Switch 使用 Tauri，但它的视觉形态来自 React 组件架构、页面信息层级和桌面交互，而不是来自 Tauri 本身。Wails 和 Tauri 都使用 WebView 渲染前端。

当前项目已经具备：

- Wails 3 Go 服务绑定；
- Windows 系统托盘；
- 关闭窗口隐藏和托盘恢复；
- Windows 开机启动；
- `frontend/dist` 嵌入 Go 二进制；
- Codex 配置备份、写入和恢复能力。

因此只替换前端渲染层，能以最小风险获得“桌面工具形态”。迁移到 Tauri 会同时引入 Rust 命令层、窗口生命周期迁移、绑定迁移和打包迁移，不能解决当前的网页式布局问题。

## 3. 依赖变化规划

### 保留

```json
{
  "@wailsio/runtime": "3.0.0-beta.26",
  "typescript": "^5.6.3",
  "vite": "^7.0.0"
}
```

### 增加

```json
{
  "vue": "^3.5.43",
  "vue-router": "^5.3.1",
  "pinia": "^4.0.3",
  "naive-ui": "^2.45.3",
  "lucide-vue-next": "^1.0.0"
}
```

开发依赖增加：

```json
{
  "@vitejs/plugin-vue": "^6.0.0",
  "vue-tsc": "^3.3.11"
}
```

版本号在正式实施时按当前 lockfile 可用版本锁定；此处的版本范围表达兼容主版本，不要求现在安装依赖。

### 暂不增加

- Tailwind CSS：当前界面需要稳定的桌面令牌和少量布局规则，工具类会把壳层结构拆散。
- Reka UI / shadcn-vue：它们适合自建组件系统；当前项目已有完整控件需求，暂不承担额外的基础组件维护成本。比较依据见 [`UI_LIBRARY_COMPARISON.md`](UI_LIBRARY_COMPARISON.md)。
- Nuxt：应用没有 SSR、SEO 或远程页面需求。
- VueUse：首版的窗口和异步能力用 Vue / Wails 原生 API 即可。
- TanStack Query：资源数量少，Pinia store 足以管理刷新、缓存和错误状态。
- 动画库：只需要 CSS transition，激活流程重点是状态反馈，不是复杂动画。
- 图表和编辑器：当前 V1 没有对应产品需求。

## 4. 构建与运行约束

### 开发

```text
Wails dev
 ├── Vite dev server（frontend）
 ├── Wails Go service
 └── WebView 加载本地开发资源
```

开发端口继续由现有 `VITE_PORT` 管理。Vue 入口必须能在普通 Vite 预览中启动，但生产功能只能通过 Wails 绑定访问。

### 生产

```text
vue-tsc --noEmit
  → vite build
  → frontend/dist
  → Go //go:embed all:frontend/dist
  → wails3 build
```

生产构建不得加载远程 CSS、远程字体、CDN 图标或远程脚本。所有字体、图标和 UI 原语代码必须随应用打包。

前端 `package.json` 的脚本基线调整为：

```json
{
  "dev": "vite",
  "build": "vue-tsc --noEmit && vite build",
  "preview": "vite preview"
}
```

`tsc` 仍可用于只检查 `.ts` 文件，但不能替代 `vue-tsc` 对 `.vue` SFC 的模板和 props 类型检查。

## 5. 前端运行时边界

### 唯一后端入口

Vue 页面只能调用 `src/services/wails-api.ts`，不得直接导入自动生成的 Wails 绑定。自动生成目录保持原样，重新生成后不需要手动合并业务 UI 代码。

### DTO 策略

前端沿用现有 DTO：

- `ProviderDTO`：Provider 基本资料和凭据引用；
- `ModelDTO`：Provider 归属、模型 ID、显示名、启用状态；
- `RouteDTO`：Provider / Model 组合、默认标记、是否重启；
- `ProfileDTO`：Codex 配置档案和配置路径。

不在 Vue 层复制一套后端实体。页面所需的派生值（当前 Provider、模型数量、可激活状态）由 store 计算。

### 凭据安全

- UI 只编辑环境变量名或 `credential:target` 引用，不把 API Key 写入普通状态持久化。
- 凭据字段默认使用密码输入和可选显示按钮；不进入日志和错误 toast。
- 不在模板中渲染完整请求头、查询参数或 credential 值。
- 所有外部 URL 只作为文本展示和后端测试参数，不通过 iframe 或远程页面加载。

## 6. 目录与页面架构基线

```text
frontend/src/
├── app/              # App、路由和依赖装配
├── layouts/          # DesktopShell
├── pages/            # Config、Profiles、Settings
├── features/         # providers、routes、profiles、settings
├── stores/           # Pinia 状态与异步动作
├── services/         # Wails API 适配和错误转换
├── components/ui/    # Naive UI 的业务级封装，不重复造基础控件
├── styles/           # tokens、base、desktop
└── main.ts           # Vue 入口
```

页面、store、service 三层职责固定：

```text
Page / Feature Component
          ↓ 调用语义动作
Pinia Store
          ↓ 调用服务函数
Wails API Adapter
          ↓ 调用生成绑定
Go Application Service
```

组件不负责决定后端方法名；service 不负责显示 toast；store 不直接操作 DOM。

## 7. UI 组件基线

首批只需要以下业务封装（底层控件来自 Naive UI）：

- `Button`：primary / secondary / ghost / danger；
- `IconButton`：必须有可访问名称；
- `Input`、`Textarea`、`Select`、`Checkbox`；
- `Dialog`、`DialogHeader`、`DialogFooter`；
- `Badge`：active / neutral / warning / danger；
- `Toast` 和页面级 `StatusNotice`；
- `EmptyState`、`ListSkeleton`、`ErrorState`；
- `ProviderIcon`：首字母或后续接入的可信品牌资源。

这些组件使用项目语义令牌，不直接暴露一堆颜色和间距 props。需要新的视觉变化时先扩展令牌或变体，而不是在页面里临时写样式。

## 8. 技术验收标准

### 桌面形态

- 运行时只加载本地资源，不打开浏览器新标签或远程页面。
- 窗口缩放到 960 × 640 时导航、工具栏和主操作仍可用。
- 主内容滚动时导航和顶部工具栏保持固定。
- 关闭窗口进入托盘，托盘恢复窗口后页面状态可继续使用。

### 工程质量

- `vue-tsc --noEmit` 通过；
- `npm run build` 通过；
- Wails Windows 构建通过；
- 生成绑定重新生成后，Vue 页面无需直接修改绑定文件；
- Provider / Model / Route / Profile 的现有功能逐项完成对照验证。

### 体验质量

- 首屏能看到当前路由和当前 Profile；
- Provider 切换无需打开编辑表单；
- 编辑和新增使用可聚焦、可取消、可恢复焦点的弹窗；
- 加载、测试、激活、保存、错误和恢复状态都有明确反馈；
- 深色主题下普通文本对比度不低于 4.5:1；
- 键盘操作和 `prefers-reduced-motion` 有明确行为。
