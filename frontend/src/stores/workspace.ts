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

  const currentRoute = computed(() => routes.value.find((item) => item.default) ?? routes.value[0]);
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

  async function refresh(): Promise<void> {
    phase.value = 'loading';
    try {
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
      phase.value = 'ready';
      setNotice('本地配置已同步');
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause, '读取本地配置失败'), true);
    }
  }

  async function runAction<T>(
    action: () => Promise<T>,
    successMessage: string,
    actionPhase: WorkspacePhase = 'saving',
  ): Promise<T | undefined> {
    phase.value = actionPhase;
    try {
      const result = await action();
      phase.value = 'ready';
      setNotice(successMessage);
      await refresh();
      return result;
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause), true);
      return undefined;
    }
  }

  async function activate(routeID: string, profileID = selectedProfileID.value): Promise<boolean> {
    if (!profileID) {
      setNotice('请先创建或选择一个配置档案', true);
      return false;
    }
    phase.value = 'activating';
    try {
      await wailsApi.activateRoute(routeID, profileID);
      setNotice('路由已激活，Codex 配置已更新');
      await refresh();
      return true;
    } catch (cause) {
      phase.value = 'error';
      setNotice(describeError(cause, '路由激活失败，配置可能已自动恢复'), true);
      return false;
    }
  }

  async function testProvider(id: string): Promise<void> {
    await runAction(() => wailsApi.testProvider(id), 'Provider 连接测试成功', 'testing');
  }

  async function testModel(providerID: string, id: string): Promise<void> {
    await runAction(() => wailsApi.testModel(providerID, id), '模型 Responses 调用成功', 'testing');
  }

  async function startCodex(): Promise<void> {
    await runAction(wailsApi.startCodex, '已请求启动 Codex', 'saving');
  }

  async function setAutostart(enabled: boolean): Promise<void> {
    const previous = autostart.value;
    autostart.value = enabled;
    const result = await runAction(
      () => wailsApi.setAutostart(enabled),
      enabled ? '已开启开机自动启动' : '已关闭开机自动启动',
    );
    if (result === undefined) autostart.value = previous;
  }

  async function createProfile(item: ProfileDTO): Promise<void> {
    await runAction(async () => {
      const created = await wailsApi.createProfile(item.id, item.name);
      if (item.config_path) await wailsApi.saveProfile({ ...created, config_path: item.config_path });
    }, '配置档案已创建');
  }

  async function saveProfile(item: ProfileDTO): Promise<void> {
    await runAction(() => wailsApi.saveProfile(item), '配置档案已保存');
  }

  async function deleteProfile(id: string): Promise<void> {
    await runAction(() => wailsApi.deleteProfile(id), '配置档案已删除');
  }

  async function restoreProfile(id: string): Promise<void> {
    await runAction(() => wailsApi.restoreProfile(id), 'Codex 配置已恢复');
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
    currentRoute,
    currentProvider,
    currentModel,
    modelsByProvider,
    busy,
    refresh,
    runAction,
    activate,
    testProvider,
    testModel,
    startCodex,
    setAutostart,
    createProfile,
    saveProfile,
    deleteProfile,
    restoreProfile,
  };
});
