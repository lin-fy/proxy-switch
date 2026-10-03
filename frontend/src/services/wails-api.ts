import * as generated from '../../bindings/proxy-switch/internal/interfaces/wails/app';
import type {
  ModelDTO,
  ProfileDTO,
  ProviderDTO,
  ProviderHealthDTO,
  ProviderPresetDTO,
  WorkspaceImportPreviewDTO,
  RouteDTO,
  CodexConfigStatusDTO,
} from '../../bindings/proxy-switch/internal/interfaces/wails/models';

declare global {
  interface Window {
    /** Wails 宿主注入的运行时对象;浏览器中不存在 environment 字段。 */
    _wails?: { environment?: { OS?: string } };
  }
}

export type {
  CodexConfigStatusDTO,
  ModelDTO,
  ProfileDTO,
  ProviderDTO,
  ProviderHealthDTO,
  ProviderPresetDTO,
  RouteDTO,
  WorkspaceImportPreviewDTO,
};

/** 适配层统一接口:Wails 绑定与浏览器预览演示数据都实现它。 */
export interface WailsApi {
  listProviders(): Promise<ProviderDTO[] | null>;
  listProviderPresets(): Promise<ProviderPresetDTO[]>;
  listModels(): Promise<ModelDTO[] | null>;
  listModelsByProvider(providerID: string): Promise<ModelDTO[] | null>;
  listRoutes(): Promise<RouteDTO[] | null>;
  listProfiles(): Promise<ProfileDTO[] | null>;
  createProvider(id: string, name: string, baseURL: string, authRef: string): Promise<ProviderDTO>;
  saveProvider(item: ProviderDTO): Promise<void>;
  deleteProvider(id: string): Promise<void>;
  testProvider(id: string): Promise<void>;
  checkProviderHealth(id: string): Promise<ProviderHealthDTO>;
  syncProviderModels(providerID: string): Promise<ModelDTO[] | null>;
  createModel(providerID: string, id: string, name: string): Promise<ModelDTO>;
  saveModel(item: ModelDTO): Promise<void>;
  deleteModel(providerID: string, id: string): Promise<void>;
  testModel(providerID: string, id: string): Promise<void>;
  createRoute(id: string, name: string, providerID: string, modelID: string): Promise<RouteDTO>;
  saveRoute(item: RouteDTO): Promise<void>;
  deleteRoute(id: string): Promise<void>;
  exportWorkspace(): Promise<string>;
  previewWorkspaceImport(payload: string): Promise<WorkspaceImportPreviewDTO>;
  importWorkspace(payload: string): Promise<WorkspaceImportPreviewDTO>;
  activateRoute(routeID: string, profileID: string): Promise<void>;
  createProfile(id: string, name: string): Promise<ProfileDTO>;
  inspectCodexConfig(): Promise<CodexConfigStatusDTO>;
  importCodexConfig(id: string, name: string): Promise<ProfileDTO>;
  saveProfile(item: ProfileDTO): Promise<void>;
  deleteProfile(id: string): Promise<void>;
  restoreProfile(id: string): Promise<void>;
  codexRunning(): Promise<boolean>;
  startCodex(): Promise<void>;
  autostartEnabled(): Promise<boolean>;
  setAutostart(enabled: boolean): Promise<void>;
}

const generatedApi: WailsApi = {
  listProviders: () => generated.ListProviders(),
  listProviderPresets: () => generated.ListProviderPresets(),
  listModels: () => generated.ListModels(),
  listModelsByProvider: (providerID) => generated.ListModelsByProvider(providerID),
  listRoutes: () => generated.ListRoutes(),
  listProfiles: () => generated.ListProfiles(),
  createProvider: (id, name, baseURL, authRef) => generated.CreateProvider(id, name, baseURL, authRef),
  saveProvider: (item) => generated.SaveProvider(item),
  deleteProvider: (id) => generated.DeleteProvider(id),
  testProvider: (id) => generated.TestProvider(id),
  checkProviderHealth: (id) => generated.CheckProviderHealth(id),
  syncProviderModels: (providerID) => generated.SyncProviderModels(providerID),
  createModel: (providerID, id, name) => generated.CreateModel(providerID, id, name),
  saveModel: (item) => generated.SaveModel(item),
  deleteModel: (providerID, id) => generated.DeleteModel(providerID, id),
  testModel: (providerID, id) => generated.TestProviderModel(providerID, id),
  createRoute: (id, name, providerID, modelID) => generated.CreateRoute(id, name, providerID, modelID),
  saveRoute: (item) => generated.SaveRoute(item),
  deleteRoute: (id) => generated.DeleteRoute(id),
  exportWorkspace: () => generated.ExportWorkspace(),
  previewWorkspaceImport: (payload) => generated.PreviewWorkspaceImport(payload),
  importWorkspace: (payload) => generated.ImportWorkspace(payload),
  activateRoute: (routeID, profileID) => generated.ActivateRoute(routeID, profileID),
  createProfile: (id, name) => generated.CreateProfile(id, name),
  inspectCodexConfig: () => generated.InspectCodexConfig(),
  importCodexConfig: (id, name) => generated.ImportCodexConfig(id, name),
  saveProfile: (item) => generated.SaveProfile(item),
  deleteProfile: (id) => generated.DeleteProfile(id),
  restoreProfile: (id) => generated.RestoreCodexConfig(id),
  codexRunning: () => generated.CodexRunning(),
  startCodex: () => generated.StartCodex(),
  autostartEnabled: () => generated.AutostartEnabled(),
  setAutostart: (enabled) => generated.SetAutostart(enabled),
};

/**
 * 浏览器预览模式:仅在开发构建且 Wails runtime(宿主注入的 environment)不存在时启用,
 * 使用内存演示数据让界面可以在浏览器中走查;生产构建与桌面运行不受影响。
 */
const isPreview = import.meta.env.DEV && typeof window !== 'undefined' && !window._wails?.environment;

export const isPreviewMode: boolean = isPreview;

export const wailsApi: WailsApi = isPreview
  ? await import('./demo-data').then((module) => module.createDemoApi())
  : generatedApi;

/*
 * 后端错误的安全映射:把未知异常转换为可操作的中文提示。
 * 原始异常文本、凭据与请求参数不允许出现在提示、toast 或 DOM 中。
 */
const SAFE_ERROR_PATTERNS: Array<[RegExp, string]> = [
  [
    /network|fetch|failed to fetch|timeout|timed out|econnrefused|connection refused|connect|dns|proxy|tls|ssl|certificate|socket|refused/i,
    '无法连接到 Provider。请检查 API 地址、本机网络和代理设置后重试。',
  ],
  [
    /401|403|unauthorized|forbidden|invalid[ _-]?api[ _-]?key|api[ _-]?key|credential|auth|token|凭据/i,
    '凭据校验未通过。请检查环境变量名或 credential: 引用是否指向有效凭据。',
  ],
  [/404|not found|no such|path|endpoint/i, '服务端接口不存在。请确认 Responses API 地址包含版本路径(如 /v1)。'],
  [/429|rate limit|too many|quota|额度/i, '请求过于频繁或额度不足。请稍后重试,或检查 Provider 额度。'],
  [/profile|档案/i, '配置档案不可用。请选择有效的档案后重试。'],
  [/disabled|停用|未启用/i, '所选模型已停用。请先在 Provider 中启用该模型，再创建或激活路由。'],
  [
    /toml|config|backup|restore|rollback|备份|恢复/i,
    'Codex 配置写入或恢复出现问题。可在「配置档案」页恢复最近备份后重试。',
  ],
  [/json|decode|parse|格式|invalid/i, '返回数据无法解析。请确认 Provider 兼容 Responses API 后重试。'],
];

export function describeError(error: unknown, fallback = '操作失败'): string {
  const raw = error instanceof Error ? error.message : typeof error === 'string' ? error : '';
  const text = raw.trim();
  if (!text) return `${fallback},请检查配置后重试。`;
  for (const [pattern, safeMessage] of SAFE_ERROR_PATTERNS) {
    if (pattern.test(text)) return safeMessage;
  }
  return `${fallback},请检查地址与凭据引用后重试;如持续失败,可在「配置档案」页恢复最近备份。`;
}
