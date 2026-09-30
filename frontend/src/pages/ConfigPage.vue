<script setup lang="ts">
import { computed, ref } from 'vue';
import {
  useDialog,
  useMessage,
  NButton,
  NCard,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSkeleton,
  NSwitch,
  NTag,
} from 'naive-ui';
import { Pencil, Plus, PlugZap, Power, Trash2 } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { wailsApi, type ModelDTO, type ProviderDTO, type RouteDTO } from '../services/wails-api';

const workspace = useWorkspaceStore();
const message = useMessage();
const dialog = useDialog();
const providerModal = ref(false);
const modelModal = ref(false);
const routeModal = ref(false);
const providerDraft = ref<ProviderDTO | null>(null);
const modelDraft = ref<ModelDTO | null>(null);
const routeDraft = ref<RouteDTO | null>(null);
const providerForm = ref({ id: '', name: '', baseURL: '', authRef: '' });
const modelForm = ref({ providerID: '', id: '', name: '' });
const routeForm = ref({ id: '', name: '', providerID: '', modelID: '', restart: false });
const routeModelKey = computed({
  get: () => `${routeForm.value.providerID}::${routeForm.value.modelID}`,
  set: (value: string) => {
    const [providerID, modelID] = value.split('::');
    routeForm.value.providerID = providerID ?? '';
    routeForm.value.modelID = modelID ?? '';
  },
});
const profileID = computed(() => workspace.selectedProfileID || workspace.profiles[0]?.id || '');
const providerOptions = computed(() => workspace.providers.map((item) => ({ label: item.name, value: item.id })));
const modelOptions = computed(() =>
  workspace.models
    .filter((item) => item.enabled)
    .map((item) => ({
      label: `${providerName(item.provider_id)} / ${item.name || item.id}`,
      value: `${item.provider_id}::${item.id}`,
    })),
);
const profileOptions = computed(() => workspace.profiles.map((item) => ({ label: item.name, value: item.id })));

function providerName(id: string): string {
  return workspace.providers.find((item) => item.id === id)?.name ?? id;
}
function modelName(providerID: string, modelID: string): string {
  return workspace.models.find((item) => item.provider_id === providerID && item.id === modelID)?.name || modelID;
}
function openProvider(item?: ProviderDTO): void {
  providerDraft.value = item ? { ...item } : null;
  providerForm.value = item
    ? { id: item.id, name: item.name, baseURL: item.base_url, authRef: item.auth_ref ?? '' }
    : { id: '', name: '', baseURL: '', authRef: '' };
  providerModal.value = true;
}
function openModel(item?: ModelDTO): void {
  modelDraft.value = item ? { ...item } : null;
  modelForm.value = item
    ? { providerID: item.provider_id, id: item.id, name: item.name }
    : { providerID: workspace.providers[0]?.id ?? '', id: '', name: '' };
  modelModal.value = true;
}
function openRoute(item?: RouteDTO): void {
  routeDraft.value = item ? { ...item } : null;
  routeForm.value = item
    ? {
        id: item.id,
        name: item.name,
        providerID: item.provider_id,
        modelID: item.model_id,
        restart: item.restart_on_activate,
      }
    : {
        id: '',
        name: '',
        providerID: workspace.providers[0]?.id ?? '',
        modelID: workspace.models[0]?.id ?? '',
        restart: false,
      };
  routeModal.value = true;
}
async function saveProvider(): Promise<void> {
  const f = providerForm.value;
  if (!f.id || !f.name || !f.baseURL) {
    message.warning('请填写标识、名称和 API 地址');
    return;
  }
  if (providerDraft.value)
    await workspace.runAction(
      () =>
        wailsApi.saveProvider({
          ...providerDraft.value!,
          id: f.id,
          name: f.name,
          base_url: f.baseURL,
          auth_ref: f.authRef,
        }),
      'Provider 已保存',
    );
  else await workspace.runAction(() => wailsApi.createProvider(f.id, f.name, f.baseURL, f.authRef), 'Provider 已添加');
  if (!workspace.error) providerModal.value = false;
}
async function saveModel(): Promise<void> {
  const f = modelForm.value;
  if (!f.providerID || !f.id) {
    message.warning('请选择 Provider 并填写模型 ID');
    return;
  }
  if (modelDraft.value)
    await workspace.runAction(
      () => wailsApi.saveModel({ ...modelDraft.value!, provider_id: f.providerID, id: f.id, name: f.name }),
      'Model 已保存',
    );
  else await workspace.runAction(() => wailsApi.createModel(f.providerID, f.id, f.name), 'Model 已添加');
  if (!workspace.error) modelModal.value = false;
}
async function saveRoute(): Promise<void> {
  const f = routeForm.value;
  if (!f.id || !f.name || !f.providerID || !f.modelID) {
    message.warning('请填写路由信息');
    return;
  }
  const value: RouteDTO = {
    ...(routeDraft.value ?? { platform_id: 'codex', default: false }),
    id: f.id,
    name: f.name,
    provider_id: f.providerID,
    model_id: f.modelID,
    restart_on_activate: f.restart,
  };
  if (routeDraft.value) await workspace.runAction(() => wailsApi.saveRoute(value), '路由已保存');
  else await workspace.runAction(() => wailsApi.createRoute(f.id, f.name, f.providerID, f.modelID), '路由已添加');
  if (!workspace.error) routeModal.value = false;
}
function remove(kind: 'provider' | 'model' | 'route', item: ProviderDTO | ModelDTO | RouteDTO): void {
  dialog.warning({
    title: '确认删除',
    content: `删除后需要重新配置“${item.name || item.id}”，确定继续吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      if (kind === 'provider')
        await workspace.runAction(() => wailsApi.deleteProvider((item as ProviderDTO).id), 'Provider 已删除');
      if (kind === 'model') {
        const m = item as ModelDTO;
        await workspace.runAction(() => wailsApi.deleteModel(m.provider_id, m.id), 'Model 已删除');
      }
      if (kind === 'route') await workspace.runAction(() => wailsApi.deleteRoute((item as RouteDTO).id), '路由已删除');
    },
  });
}
async function activate(route: RouteDTO): Promise<void> {
  await workspace.activate(route.id, profileID.value);
  if (!workspace.error) message.success('路由已激活');
}
async function testProvider(id: string): Promise<void> {
  await workspace.testProvider(id);
  if (!workspace.error) message.success('Provider 连接成功');
}
async function testModel(item: ModelDTO): Promise<void> {
  await workspace.testModel(item.provider_id, item.id);
  if (!workspace.error) message.success('模型测试成功');
}
</script>

<template>
  <div class="page-width">
    <section class="hero-grid">
      <div>
        <p class="eyebrow">LOCAL ROUTING</p>
        <h2>把当前模型配置整理成<br /><span>一个清晰的工作台</span></h2>
        <p class="hero-copy">Provider 负责连接，Model 负责能力，Route 决定 Codex 现在使用谁。</p>
      </div>
      <div class="route-summary">
        <div class="summary-label">
          <span class="status-light" :class="{ 'is-busy': workspace.busy }" />当前激活路由
        </div>
        <strong>{{ workspace.currentRoute?.name || '还没有激活路由' }}</strong
        ><span
          >{{ workspace.currentProvider?.name || '选择一个 Provider' }} <i>/</i>
          {{ workspace.currentModel?.name || '选择一个 Model' }}</span
        ><NSelect
          v-model:value="workspace.selectedProfileID"
          :options="profileOptions"
          size="small"
          placeholder="选择配置档案"
        />
      </div>
    </section>
    <div v-if="workspace.error" class="notice notice-error">{{ workspace.message }}</div>
    <div v-else-if="workspace.phase === 'loading' && !workspace.providers.length" class="loading-grid">
      <NSkeleton v-for="n in 3" :key="n" text :repeat="4" />
    </div>
    <section v-else class="resource-grid">
      <NCard class="resource-card resource-card-wide" :bordered="false">
        <div class="card-heading">
          <div>
            <p class="eyebrow">CONNECTIONS</p>
            <h3>Providers</h3>
            <p>管理可连接的 API 服务。</p>
          </div>
          <NButton size="small" type="primary" @click="openProvider()"><Plus :size="15" />添加 Provider</NButton>
        </div>
        <NEmpty v-if="!workspace.providers.length" description="还没有 Provider" />
        <div v-else class="resource-list">
          <div
            v-for="item in workspace.providers"
            :key="item.id"
            class="resource-row provider-row"
            :class="{ 'is-active': workspace.currentProvider?.id === item.id }"
          >
            <div class="resource-icon">{{ item.name.slice(0, 1).toUpperCase() }}</div>
            <div class="resource-main">
              <strong>{{ item.name }}</strong
              ><span>{{ item.base_url }}</span>
            </div>
            <NTag size="small" :bordered="false">{{ item.protocol }}</NTag
            ><NButton quaternary size="small" @click="testProvider(item.id)"><PlugZap :size="15" /></NButton
            ><NButton quaternary size="small" @click="openProvider(item)"><Pencil :size="15" /></NButton
            ><NButton quaternary size="small" type="error" @click="remove('provider', item)"
              ><Trash2 :size="15"
            /></NButton>
          </div>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false">
        <div class="card-heading">
          <div>
            <p class="eyebrow">AVAILABLE MODELS</p>
            <h3>Models</h3>
            <p>启用后可被 Route 选择。</p>
          </div>
          <NButton quaternary size="small" @click="openModel()"><Plus :size="15" /></NButton>
        </div>
        <NEmpty v-if="!workspace.models.length" description="还没有 Model" />
        <div v-else class="resource-list">
          <div
            v-for="item in workspace.models"
            :key="`${item.provider_id}:${item.id}`"
            class="resource-row compact-row"
          >
            <div class="resource-main">
              <strong>{{ item.name || item.id }}</strong
              ><span>{{ providerName(item.provider_id) }} · {{ item.id }}</span>
            </div>
            <NSwitch
              :value="item.enabled"
              size="small"
              @update:value="
                (enabled) =>
                  workspace.runAction(
                    () => wailsApi.saveModel({ ...item, enabled }),
                    enabled ? 'Model 已启用' : 'Model 已停用',
                  )
              "
            /><NButton quaternary size="small" @click="testModel(item)"><PlugZap :size="15" /></NButton
            ><NButton quaternary size="small" @click="openModel(item)"><Pencil :size="15" /></NButton>
          </div>
        </div>
      </NCard>
      <NCard class="resource-card resource-card-wide" :bordered="false">
        <div class="card-heading">
          <div>
            <p class="eyebrow">CODEX TARGETS</p>
            <h3>Routes</h3>
            <p>将 Provider 与 Model 组合成可激活的目标。</p>
          </div>
          <NButton size="small" type="primary" @click="openRoute()"><Plus :size="15" />添加 Route</NButton>
        </div>
        <NEmpty v-if="!workspace.routes.length" description="还没有 Route" />
        <div v-else class="resource-list">
          <div v-for="item in workspace.routes" :key="item.id" class="resource-row route-row">
            <div class="route-indicator" :class="{ 'is-active': item.default }"><Power :size="15" /></div>
            <div class="resource-main">
              <strong>{{ item.name }}</strong
              ><span>{{ providerName(item.provider_id) }} / {{ modelName(item.provider_id, item.model_id) }}</span>
            </div>
            <NTag v-if="item.default" type="success" size="small" :bordered="false">当前</NTag
            ><NButton size="small" secondary :loading="workspace.phase === 'activating'" @click="activate(item)"
              >激活</NButton
            ><NButton quaternary size="small" @click="openRoute(item)"><Pencil :size="15" /></NButton
            ><NButton quaternary size="small" type="error" @click="remove('route', item)"
              ><Trash2 :size="15"
            /></NButton>
          </div>
        </div>
      </NCard>
    </section>
    <p class="page-footnote">配置写入前会自动保留备份。所有请求都在本机 Wails 进程中完成。</p>
  </div>

  <NModal
    v-model:show="providerModal"
    preset="card"
    :title="providerDraft ? '编辑 Provider' : '添加 Provider'"
    class="edit-modal"
  >
    <NForm label-placement="top"
      ><NFormItem label="标识"
        ><NInput v-model:value="providerForm.id" :disabled="!!providerDraft" placeholder="例如 openai" /></NFormItem
      ><NFormItem label="显示名称"><NInput v-model:value="providerForm.name" placeholder="例如 OpenAI" /></NFormItem
      ><NFormItem label="Responses API 地址"
        ><NInput v-model:value="providerForm.baseURL" placeholder="https://api.example.com/v1" /></NFormItem
      ><NFormItem label="凭据引用"
        ><NInput v-model:value="providerForm.authRef" placeholder="环境变量名或 credential:target" /></NFormItem
    ></NForm>
    <template #footer
      ><div class="modal-footer">
        <NButton @click="providerModal = false">取消</NButton
        ><NButton type="primary" @click="saveProvider">保存 Provider</NButton>
      </div></template
    >
  </NModal>
  <NModal v-model:show="modelModal" preset="card" :title="modelDraft ? '编辑 Model' : '添加 Model'" class="edit-modal">
    <NForm label-placement="top"
      ><NFormItem label="Provider"
        ><NSelect
          v-model:value="modelForm.providerID"
          :options="providerOptions"
          placeholder="选择 Provider" /></NFormItem
      ><NFormItem label="模型 ID"
        ><NInput v-model:value="modelForm.id" :disabled="!!modelDraft" placeholder="例如 gpt-5.6-sol" /></NFormItem
      ><NFormItem label="显示名称"><NInput v-model:value="modelForm.name" placeholder="可选" /></NFormItem
    ></NForm>
    <template #footer
      ><div class="modal-footer">
        <NButton @click="modelModal = false">取消</NButton
        ><NButton type="primary" @click="saveModel">保存 Model</NButton>
      </div></template
    >
  </NModal>
  <NModal v-model:show="routeModal" preset="card" :title="routeDraft ? '编辑 Route' : '添加 Route'" class="edit-modal">
    <NForm label-placement="top"
      ><NFormItem label="路由标识"
        ><NInput v-model:value="routeForm.id" :disabled="!!routeDraft" placeholder="例如 work" /></NFormItem
      ><NFormItem label="路由名称"><NInput v-model:value="routeForm.name" placeholder="例如 工作模型" /></NFormItem
      ><NFormItem label="模型"
        ><NSelect v-model:value="routeModelKey" :options="modelOptions" placeholder="选择已启用的 Model" /></NFormItem
      ><NFormItem label="激活行为"
        ><NSwitch v-model:value="routeForm.restart" /> <span class="switch-copy">激活时尝试启动 Codex</span></NFormItem
      ></NForm
    >
    <template #footer
      ><div class="modal-footer">
        <NButton @click="routeModal = false">取消</NButton
        ><NButton type="primary" @click="saveRoute">保存 Route</NButton>
      </div></template
    >
  </NModal>
</template>
