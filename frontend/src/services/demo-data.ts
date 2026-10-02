import type { WailsApi } from './wails-api';
import type {
  ModelDTO,
  ProfileDTO,
  ProviderDTO,
  RouteDTO,
} from '../../bindings/proxy-switch/internal/interfaces/wails/models';

/*
 * 浏览器预览模式的内存演示数据,只能在 wails-api.ts 的预览开关下被导入
 * (开发构建 + 无 Wails runtime)。数据与 DTO 结构一致,支持完整增删改流程,
 * 用于在浏览器中走查界面状态;不代表真实后端行为。
 */

const providers: ProviderDTO[] = [
  {
    id: 'cpa',
    name: 'CPA',
    base_url: 'http://127.0.0.1:8317/v1',
    protocol: 'Responses API',
    auth_ref: 'CPA_API_KEY',
  },
  {
    id: 'openai',
    name: 'OpenAI',
    base_url: 'https://api.openai.com/v1',
    protocol: 'Responses API',
    auth_ref: 'OPENAI_API_KEY',
  },
];

const models: ModelDTO[] = [
  { provider_id: 'cpa', id: 'gpt-5.6-sol', name: 'GPT-5.6 Sol', enabled: true },
  { provider_id: 'cpa', id: 'gpt-5.6-luna', name: 'GPT-5.6 Luna', enabled: true },
  { provider_id: 'cpa', id: 'gpt-4.1-mini', name: '', enabled: false },
  { provider_id: 'openai', id: 'gpt-5.3-codex', name: 'GPT-5.3 Codex', enabled: true },
  { provider_id: 'openai', id: 'o4-mini', name: '', enabled: false },
];

const routes: RouteDTO[] = [
  {
    id: 'work-sol',
    name: '工作主力',
    platform_id: 'codex',
    provider_id: 'cpa',
    model_id: 'gpt-5.6-sol',
    restart_on_activate: true,
    default: true,
  },
  {
    id: 'openai-daily',
    name: 'OpenAI 日常',
    platform_id: 'codex',
    provider_id: 'openai',
    model_id: 'gpt-5.3-codex',
    restart_on_activate: false,
    default: false,
  },
];

const profiles: ProfileDTO[] = [
  { id: 'work', name: '工作', config_path: 'profiles/work.config.toml' },
  { id: 'personal', name: '个人', config_path: '' },
];

let autostart = false;
let codexRunning = true;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function requireProvider(id: string): ProviderDTO {
  const item = providers.find((provider) => provider.id === id);
  if (!item) throw new Error(`provider ${id} not found`);
  return item;
}

export function createDemoApi(): WailsApi {
  return {
    async listProviders() {
      await sleep(120);
      return clone(providers);
    },
    async listModels() {
      await sleep(120);
      return clone(models);
    },
    async listModelsByProvider(providerID) {
      await sleep(80);
      return clone(models.filter((model) => model.provider_id === providerID));
    },
    async listRoutes() {
      await sleep(120);
      return clone(routes);
    },
    async listProfiles() {
      await sleep(120);
      return clone(profiles);
    },
    async createProvider(id, name, baseURL, authRef) {
      await sleep(150);
      if (providers.some((provider) => provider.id === id)) throw new Error(`provider id "${id}" already exists`);
      const created: ProviderDTO = {
        id,
        name,
        base_url: baseURL,
        protocol: 'Responses API',
        auth_ref: authRef || undefined,
      };
      providers.push(created);
      return clone(created);
    },
    async saveProvider(item) {
      await sleep(150);
      requireProvider(item.id);
      providers.splice(
        providers.findIndex((provider) => provider.id === item.id),
        1,
        clone(item),
      );
    },
    async deleteProvider(id) {
      await sleep(150);
      requireProvider(id);
      providers.splice(
        providers.findIndex((provider) => provider.id === id),
        1,
      );
      for (let index = models.length - 1; index >= 0; index -= 1)
        if (models[index].provider_id === id) models.splice(index, 1);
      for (let index = routes.length - 1; index >= 0; index -= 1)
        if (routes[index].provider_id === id) routes.splice(index, 1);
    },
    async testProvider(id) {
      await sleep(600);
      requireProvider(id);
    },
    async syncProviderModels(providerID) {
      await sleep(500);
      requireProvider(providerID);
      return clone(models.filter((model) => model.provider_id === providerID));
    },
    async createModel(providerID, id, name) {
      await sleep(150);
      requireProvider(providerID);
      if (models.some((model) => model.provider_id === providerID && model.id === id))
        throw new Error(`model "${id}" already exists`);
      const created: ModelDTO = { provider_id: providerID, id, name, enabled: false };
      models.push(created);
      return clone(created);
    },
    async saveModel(item) {
      await sleep(150);
      requireProvider(item.provider_id);
      const index = models.findIndex((model) => model.provider_id === item.provider_id && model.id === item.id);
      if (index < 0) throw new Error(`model "${item.id}" not found`);
      models.splice(index, 1, clone(item));
    },
    async deleteModel(providerID, id) {
      await sleep(150);
      const index = models.findIndex((model) => model.provider_id === providerID && model.id === id);
      if (index < 0) throw new Error(`model "${id}" not found`);
      models.splice(index, 1);
      for (let routeIndex = routes.length - 1; routeIndex >= 0; routeIndex -= 1) {
        const route = routes[routeIndex];
        if (route.provider_id === providerID && route.model_id === id) routes.splice(routeIndex, 1);
      }
    },
    async testModel(providerID, id) {
      await sleep(700);
      const provider = requireProvider(providerID);
      if (!models.some((model) => model.provider_id === providerID && model.id === id))
        throw new Error(`model "${id}" not found`);
      if (provider.auth_ref === 'OPENAI_API_KEY')
        throw new Error('401 unauthorized: invalid api key (demo failure sample)');
    },
    async createRoute(id, name, providerID, modelID) {
      await sleep(150);
      if (routes.some((route) => route.id === id)) throw new Error(`route id "${id}" already exists`);
      requireProvider(providerID);
      const created: RouteDTO = {
        id,
        name,
        platform_id: 'codex',
        provider_id: providerID,
        model_id: modelID,
        restart_on_activate: false,
        default: false,
      };
      routes.push(created);
      return clone(created);
    },
    async saveRoute(item) {
      await sleep(150);
      const index = routes.findIndex((route) => route.id === item.id);
      if (index < 0) throw new Error(`route "${item.id}" not found`);
      routes.splice(index, 1, clone(item));
    },
    async deleteRoute(id) {
      await sleep(150);
      const index = routes.findIndex((route) => route.id === id);
      if (index < 0) throw new Error(`route "${id}" not found`);
      routes.splice(index, 1);
    },
    async activateRoute(routeID, profileID) {
      await sleep(800);
      const route = routes.find((item) => item.id === routeID);
      if (!route) throw new Error(`route "${routeID}" not found`);
      if (!profiles.some((profile) => profile.id === profileID)) throw new Error(`profile "${profileID}" not found`);
      for (const item of routes) item.default = item.id === routeID;
      codexRunning = true;
    },
    async createProfile(id, name) {
      await sleep(150);
      if (profiles.some((profile) => profile.id === id)) throw new Error(`profile id "${id}" already exists`);
      const created: ProfileDTO = { id, name };
      profiles.push(created);
      return clone(created);
    },
    async saveProfile(item) {
      await sleep(150);
      const index = profiles.findIndex((profile) => profile.id === item.id);
      if (index < 0) throw new Error(`profile "${item.id}" not found`);
      profiles.splice(index, 1, clone(item));
    },
    async deleteProfile(id) {
      await sleep(150);
      const index = profiles.findIndex((profile) => profile.id === id);
      if (index < 0) throw new Error(`profile "${id}" not found`);
      profiles.splice(index, 1);
    },
    async restoreProfile() {
      await sleep(600);
    },
    async codexRunning() {
      await sleep(60);
      return codexRunning;
    },
    async startCodex() {
      await sleep(400);
      codexRunning = true;
    },
    async autostartEnabled() {
      await sleep(60);
      return autostart;
    },
    async setAutostart(enabled) {
      await sleep(200);
      autostart = enabled;
    },
  };
}
