import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import {
  describeError,
  wailsApi,
  type ModelDTO,
  type ProfileDTO,
  type ProviderDTO,
  type RouteDTO,
} from '../services/wails-api';

export type WorkspacePhase = 'idle' | 'loading' | 'ready' | 'saving' | 'testing' | 'activating' | 'error';
export type ActivationStage = 'idle' | 'updating' | 'ready' | 'error';

export const useWorkspaceStore = defineStore('workspace', () => {
  const providers = ref<ProviderDTO[]>([]);
  const models = ref<ModelDTO[]>([]);
  const routes = ref<RouteDTO[]>([]);
  const profiles = ref<ProfileDTO[]>([]);
  const selectedProfileID = ref('');
  const autostart = ref(false);
  const codexRunning = ref(false);
  const phase = ref<WorkspacePhase>('idle');
  const message = ref('准备连接本地 Codex 配置');
  const error = ref(false);
  const errorDetail = ref('');
  const activationStage = ref<ActivationStage>('idle');
  // 后端契约没有恢复结果事件:激活失败时保持 unknown(null),由界面提供恢复入口。
  const activationRecovered = ref<boolean | null>(null);
  const activatingRouteID = ref<string | null>(null);
  // 档案创建成功但配置路径保存失败时记录已创建档案,重试时只补存路径,不再重复创建。
  const pendingProfile = ref<ProfileDTO | null>(null);

  const currentRoute = computed(() => routes.value.find((item) => item.default));
  const currentProvider = computed(() =>
    currentRoute.value ? providers.value.find((item) => item.id === currentRoute.value?.provider_id) : undefined,
  );
  const currentModel = computed(() =>
    currentRoute.value
      ? models.value.find(
          (item) => item.provider_id === currentRoute.value?.provider_id && item.id === currentRoute.value?.model_id,
        )
      : undefined,
  );
  const selectedProfile = computed(() => profiles.value.find((item) => item.id === selectedProfileID.value));
  const modelsByProvider = computed(() => {
    const result = new Map<string, ModelDTO[]>();
    for (const item of models.value) result.set(item.provider_id, [...(result.get(item.provider_id) ?? []), item]);
    return result;
  });
  const busy = computed(() => ['loading', 'saving', 'testing', 'activating'].includes(phase.value));

  function setNotice(nextMessage: string, isError = false): void {
    message.value = nextMessage;
    error.value = isError;
    errorDetail.value = isError ? nextMessage : '';
  }

  /** 只拉取数据,不发通知、不吞异常;供需要自定义消息的调用方组合。 */
  async function reload(): Promise<void> {
    const [nextProviders, nextModels, nextRoutes, nextProfiles, nextAutostart, nextCodexRunning] = await Promise.all([
      wailsApi.listProviders(),
      wailsApi.listModels(),
      wailsApi.listRoutes(),
      wailsApi.listProfiles(),
      wailsApi.autostartEnabled(),
      wailsApi.codexRunning(),
    ]);
    providers.value = nextProviders ?? [];
    models.value = nextModels ?? [];
    routes.value = nextRoutes ?? [];
    profiles.value = nextProfiles ?? [];
    autostart.value = nextAutostart;
    codexRunning.value = nextCodexRunning;
    if (!profiles.value.some((item) => item.id === selectedProfileID.value))
      selectedProfileID.value = profiles.value[0]?.id ?? '';
  }

  async function refresh(preserveMessage = false): Promise<void> {
    if (busy.value) return;
    phase.value = 'loading';
    try {
      await reload();
      phase.value = 'ready';
      if (!preserveMessage) setNotice('本地配置已同步');
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause, '读取本地配置失败'), true);
    }
  }

  async function runAction(
    action: () => Promise<unknown>,
    successMessage: string,
    actionPhase: WorkspacePhase = 'saving',
  ): Promise<boolean> {
    if (busy.value) {
      setNotice('上一步操作尚未完成,请稍候', true);
      return false;
    }
    phase.value = actionPhase;
    try {
      await action();
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause), true);
      return false;
    }
    phase.value = 'ready';
    setNotice(successMessage);
    try {
      await reload();
      return true;
    } catch {
      // 写入已成功:报告刷新失败,不能声称操作失败。
      setNotice(`${successMessage},但状态刷新失败,请点击刷新重试`, true);
      return true;
    }
  }

  async function activate(routeID: string, profileID = selectedProfileID.value): Promise<boolean> {
    if (phase.value === 'activating') return false;
    if (!profileID) {
      activationStage.value = 'error';
      activationRecovered.value = null;
      setNotice('请先创建并选择一个配置档案', true);
      return false;
    }
    phase.value = 'activating';
    activationStage.value = 'updating';
    activationRecovered.value = null;
    activatingRouteID.value = routeID;
    try {
      await wailsApi.activateRoute(routeID, profileID);
    } catch (cause) {
      phase.value = 'error';
      activationStage.value = 'error';
      activationRecovered.value = null;
      activatingRouteID.value = null;
      setNotice(describeError(cause, '路由激活失败'), true);
      return false;
    }
    activationStage.value = 'ready';
    activatingRouteID.value = null;
    setNotice('路由已激活,Codex 配置已更新');
    try {
      await reload();
    } catch {
      setNotice('路由已激活,但状态刷新失败,请点击刷新重试', true);
    }
    phase.value = 'ready';
    return true;
  }

  async function testProvider(id: string): Promise<boolean> {
    return runAction(() => wailsApi.testProvider(id), 'Provider 连接测试成功', 'testing');
  }

  async function syncProviderModels(providerID: string): Promise<boolean> {
    return runAction(() => wailsApi.syncProviderModels(providerID), 'Provider 模型已同步', 'testing');
  }

  async function testModel(providerID: string, id: string): Promise<boolean> {
    return runAction(() => wailsApi.testModel(providerID, id), '模型 Responses 调用成功', 'testing');
  }

  async function startCodex(): Promise<boolean> {
    return runAction(wailsApi.startCodex, '已请求启动 Codex');
  }

  async function setAutostart(enabled: boolean): Promise<void> {
    const previous = autostart.value;
    autostart.value = enabled;
    const result = await runAction(
      () => wailsApi.setAutostart(enabled),
      enabled ? '已开启开机自动启动' : '已关闭开机自动启动',
    );
    // 只有写入确实失败才回滚;写成功但刷新失败按已保存处理。
    if (!result) autostart.value = previous;
  }

  async function createProvider(id: string, name: string, baseURL: string, authRef: string): Promise<boolean> {
    return runAction(() => wailsApi.createProvider(id, name, baseURL, authRef), 'Provider 已添加');
  }

  async function saveProvider(item: ProviderDTO): Promise<boolean> {
    return runAction(() => wailsApi.saveProvider(item), 'Provider 已保存');
  }

  async function deleteProvider(id: string): Promise<boolean> {
    return runAction(() => wailsApi.deleteProvider(id), 'Provider 已删除');
  }

  async function createModel(providerID: string, id: string, name: string): Promise<boolean> {
    return runAction(() => wailsApi.createModel(providerID, id, name), 'Model 已添加');
  }

  async function saveModel(item: ModelDTO, successMessage = 'Model 已保存'): Promise<boolean> {
    return runAction(() => wailsApi.saveModel(item), successMessage);
  }

  async function deleteModel(providerID: string, id: string): Promise<boolean> {
    return runAction(() => wailsApi.deleteModel(providerID, id), 'Model 已删除');
  }

  async function createRoute(input: {
    id: string;
    name: string;
    providerID: string;
    modelID: string;
    restart: boolean;
  }): Promise<boolean> {
    return runAction(async () => {
      const created = await wailsApi.createRoute(input.id, input.name, input.providerID, input.modelID);
      // CreateRoute 不接收 restart 字段:创建后立即用现有保存方法补齐,保证用户勾选被持久化。
      if (input.restart) await wailsApi.saveRoute({ ...created, restart_on_activate: true });
    }, '路由已添加');
  }

  async function saveRoute(item: RouteDTO): Promise<boolean> {
    return runAction(() => wailsApi.saveRoute(item), '路由已保存');
  }

  async function deleteRoute(id: string): Promise<boolean> {
    return runAction(() => wailsApi.deleteRoute(id), '路由已删除');
  }

  async function createProfile(item: ProfileDTO): Promise<boolean> {
    if (busy.value) {
      setNotice('上一步操作尚未完成,请稍候', true);
      return false;
    }
    phase.value = 'saving';
    const retry = pendingProfile.value?.id === item.id ? pendingProfile.value : null;
    if (retry) {
      // 首次创建已成功,重试只补存配置路径。
      return runAction(() => wailsApi.saveProfile({ ...retry, config_path: item.config_path }), '配置档案已创建').then(
        (result) => {
          if (result) pendingProfile.value = null;
          return result;
        },
      );
    }
    let created: ProfileDTO;
    try {
      created = await wailsApi.createProfile(item.id, item.name);
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause, '创建配置档案失败'), true);
      return false;
    }
    if (!item.config_path) {
      phase.value = 'ready';
      setNotice('配置档案已创建');
      try {
        await reload();
        return true;
      } catch {
        setNotice('配置档案已创建,但状态刷新失败,请点击刷新重试', true);
        return true;
      }
    }
    try {
      await wailsApi.saveProfile({ ...created, config_path: item.config_path });
      phase.value = 'ready';
      setNotice('配置档案已创建');
      try {
        await reload();
        return true;
      } catch {
        setNotice('配置档案已创建,但状态刷新失败,请点击刷新重试', true);
        return true;
      }
    } catch (cause) {
      pendingProfile.value = { ...created, config_path: item.config_path };
      phase.value = 'error';
      setNotice(`档案已创建,但保存配置路径失败:${describeError(cause, '请重试以补存路径')}`, true);
      return false;
    }
  }

  async function saveProfile(item: ProfileDTO): Promise<boolean> {
    const result = await runAction(() => wailsApi.saveProfile(item), '配置档案已保存');
    if (result && pendingProfile.value?.id === item.id) pendingProfile.value = null;
    return result;
  }

  async function deleteProfile(id: string): Promise<boolean> {
    const result = await runAction(() => wailsApi.deleteProfile(id), '配置档案已删除');
    if (result && pendingProfile.value?.id === id) pendingProfile.value = null;
    return result;
  }

  async function restoreProfile(id: string): Promise<boolean> {
    return runAction(() => wailsApi.restoreProfile(id), 'Codex 配置已恢复');
  }

  return {
    providers,
    models,
    routes,
    profiles,
    selectedProfileID,
    selectedProfile,
    autostart,
    codexRunning,
    phase,
    message,
    error,
    errorDetail,
    activationStage,
    activationRecovered,
    activatingRouteID,
    pendingProfile,
    currentRoute,
    currentProvider,
    currentModel,
    modelsByProvider,
    busy,
    refresh,
    activate,
    testProvider,
    syncProviderModels,
    testModel,
    startCodex,
    setAutostart,
    createProvider,
    saveProvider,
    deleteProvider,
    createModel,
    saveModel,
    deleteModel,
    createRoute,
    saveRoute,
    deleteRoute,
    createProfile,
    saveProfile,
    deleteProfile,
    restoreProfile,
  };
});
