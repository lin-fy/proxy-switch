<script setup lang="ts">
import { computed, ref } from 'vue';
import {
  NButton,
  NCard,
  NDropdown,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSkeleton,
  NSwitch,
  NTag,
  NTooltip,
  useDialog,
  useMessage,
  type FormInst,
  type FormRules,
} from 'naive-ui';
import { MoreHorizontal, Plus, Power, Rocket } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { useFormModal } from '../app/form';
import type { ModelDTO, ProviderDTO, RouteDTO } from '../services/wails-api';

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
const providerFormRef = ref<FormInst | null>(null);
const modelFormRef = ref<FormInst | null>(null);
const routeFormRef = ref<FormInst | null>(null);

const {
  fieldErrors: providerErrors,
  setSummaryRef: providerSetSummaryRef,
  registerField: providerRegisterField,
  focusField: providerFocusField,
  submit: providerSubmit,
  track: providerTrack,
  guardClose: guardClose,
} = useFormModal();
const {
  fieldErrors: modelErrors,
  setSummaryRef: modelSetSummaryRef,
  registerField: modelRegisterField,
  focusField: modelFocusField,
  submit: modelSubmit,
  track: modelTrack,
} = useFormModal();
const {
  fieldErrors: routeErrors,
  setSummaryRef: routeSetSummaryRef,
  registerField: routeRegisterField,
  focusField: routeFocusField,
  submit: routeSubmit,
  track: routeTrack,
} = useFormModal();

const providerTracker = providerTrack(() => providerForm.value);
const modelTracker = modelTrack(() => modelForm.value);
const routeTracker = routeTrack(() => routeForm.value);
const providerDirty = computed(() => providerTracker.dirty());
const modelDirty = computed(() => modelTracker.dirty());
const routeDirty = computed(() => routeTracker.dirty());

const routeModelKey = computed({
  get: () => `${routeForm.value.providerID}::${routeForm.value.modelID}`,
  set: (value: string) => {
    const [providerID, modelID] = value.split('::');
    routeForm.value.providerID = providerID ?? '';
    routeForm.value.modelID = modelID ?? '';
  },
});
const profileID = computed(() => workspace.selectedProfileID || workspace.profiles[0]?.id || '');
const selectedProfileName = computed(() => workspace.profiles.find((item) => item.id === profileID.value)?.name ?? '');
const providerOptions = computed(() => workspace.providers.map((item) => ({ label: item.name, value: item.id })));
const modelOptions = computed(() =>
  workspace.models
    .filter((item) => item.enabled)
    .map((item) => ({
      label: `${providerName(item.provider_id)} / ${item.name || item.id}`,
      value: `${item.provider_id}::${item.id}`,
    })),
);

const providerRules: FormRules = {
  id: { required: true, message: '请填写 Provider 标识', trigger: ['blur', 'input'] },
  name: { required: true, message: '请填写 Provider 名称', trigger: ['blur', 'input'] },
  baseURL: { required: true, message: '请填写 Responses API 地址', trigger: ['blur', 'input'] },
};
const modelRules: FormRules = {
  providerID: { required: true, message: '请选择 Provider', trigger: 'change' },
  id: { required: true, message: '请填写模型 ID', trigger: ['blur', 'input'] },
};
const routeRules: FormRules = {
  id: { required: true, message: '请填写路由标识', trigger: ['blur', 'input'] },
  name: { required: true, message: '请填写路由名称', trigger: ['blur', 'input'] },
  modelID: { required: true, message: '请选择已启用的模型', trigger: 'change' },
};

function providerName(id: string): string {
  return workspace.providers.find((item) => item.id === id)?.name ?? id;
}

function modelName(providerID: string, modelID: string): string {
  return workspace.models.find((item) => item.provider_id === providerID && item.id === modelID)?.name || modelID;
}

function enabledModelCount(providerID: string): number {
  return (workspace.modelsByProvider.get(providerID) ?? []).filter((item) => item.enabled).length;
}

function providerHasRoute(item: ProviderDTO): boolean {
  return workspace.routes.some(
    (route) => route.provider_id === item.id && isModelEnabled(route.provider_id, route.model_id),
  );
}

function isModelEnabled(providerID: string, modelID: string): boolean {
  return workspace.models.some((item) => item.provider_id === providerID && item.id === modelID && item.enabled);
}

function openProvider(item?: ProviderDTO): void {
  providerDraft.value = item ? { ...item } : null;
  providerForm.value = item
    ? { id: item.id, name: item.name, baseURL: item.base_url, authRef: item.auth_ref ?? '' }
    : { id: '', name: '', baseURL: '', authRef: '' };
  providerErrors.value = [];
  providerModal.value = true;
  providerTracker.snapshot();
}

function openModel(item?: ModelDTO): void {
  modelDraft.value = item ? { ...item } : null;
  modelForm.value = item
    ? { providerID: item.provider_id, id: item.id, name: item.name }
    : { providerID: workspace.providers[0]?.id ?? '', id: '', name: '' };
  modelErrors.value = [];
  modelModal.value = true;
  modelTracker.snapshot();
}

function openRoute(item?: RouteDTO): void {
  routeDraft.value = item ? { ...item } : null;
  const firstProviderID = workspace.providers[0]?.id ?? '';
  const firstModel = workspace.models.find((model) => model.provider_id === firstProviderID && model.enabled);
  routeForm.value = item
    ? {
        id: item.id,
        name: item.name,
        providerID: item.provider_id,
        modelID: item.model_id,
        restart: item.restart_on_activate,
      }
    : { id: '', name: '', providerID: firstProviderID, modelID: firstModel?.id ?? '', restart: false };
  routeErrors.value = [];
  routeModal.value = true;
  routeTracker.snapshot();
}

async function saveProvider(): Promise<void> {
  await providerSubmit(providerFormRef.value, async () => {
    const f = providerForm.value;
    const ok = providerDraft.value
      ? await workspace.saveProvider({
          ...providerDraft.value,
          id: f.id,
          name: f.name,
          base_url: f.baseURL,
          auth_ref: f.authRef,
        })
      : await workspace.createProvider(f.id, f.name, f.baseURL, f.authRef);
    if (ok) providerModal.value = false;
    else message.error(workspace.message);
    return ok;
  });
}

async function saveModel(): Promise<void> {
  await modelSubmit(modelFormRef.value, async () => {
    const f = modelForm.value;
    const ok = modelDraft.value
      ? await workspace.saveModel({ ...modelDraft.value, provider_id: f.providerID, id: f.id, name: f.name })
      : await workspace.createModel(f.providerID, f.id, f.name);
    if (ok) modelModal.value = false;
    else message.error(workspace.message);
    return ok;
  });
}

async function saveRoute(): Promise<void> {
  await routeSubmit(routeFormRef.value, async () => {
    const f = routeForm.value;
    const value: RouteDTO = {
      ...(routeDraft.value ?? { platform_id: 'codex', default: false }),
      id: f.id,
      name: f.name,
      provider_id: f.providerID,
      model_id: f.modelID,
      restart_on_activate: f.restart,
    };
    const ok = routeDraft.value
      ? await workspace.saveRoute(value)
      : await workspace.createRoute({
          id: f.id,
          name: f.name,
          providerID: f.providerID,
          modelID: f.modelID,
          restart: f.restart,
        });
    if (ok) routeModal.value = false;
    else message.error(workspace.message);
    return ok;
  });
}

function confirmAction(title: string, content: string, positiveText: string, action: () => Promise<boolean>): void {
  dialog.warning({
    title,
    content,
    positiveText,
    negativeText: '取消',
    positiveButtonProps: { type: 'error', secondary: true },
    onPositiveClick: async () => {
      const ok = await action();
      if (!ok) {
        message.error(workspace.message);
        return false;
      }
      message.success(workspace.message);
      return true;
    },
  });
}

function removeProvider(item: ProviderDTO): void {
  confirmAction('删除 Provider', `删除“${item.name || item.id}”后需要重新配置,确定继续吗?`, '删除', () =>
    workspace.deleteProvider(item.id),
  );
}

function removeModel(item: ModelDTO): void {
  confirmAction('删除 Model', `删除“${item.name || item.id}”后引用它的 Route 会失效,确定继续吗?`, '删除', () =>
    workspace.deleteModel(item.provider_id, item.id),
  );
}

function removeRoute(item: RouteDTO): void {
  confirmAction('删除 Route', `删除“${item.name || item.id}”后不能再用它激活,确定继续吗?`, '删除', () =>
    workspace.deleteRoute(item.id),
  );
}

async function activateRoute(route: RouteDTO): Promise<void> {
  const ok = await workspace.activate(route.id, profileID.value);
  if (ok) message.success(workspace.message);
  else message.error(workspace.message);
}

function switchProvider(item: ProviderDTO): void {
  const route = workspace.routes.find(
    (candidate) => candidate.provider_id === item.id && isModelEnabled(candidate.provider_id, candidate.model_id),
  );
  if (!route) {
    message.warning('该 Provider 还没有可用路由,请先在 Routes 中创建');
    return;
  }
  void activateRoute(route);
}

async function testProvider(item: ProviderDTO): Promise<void> {
  const ok = await workspace.testProvider(item.id);
  if (ok) message.success(workspace.message);
  else message.error(workspace.message);
}

async function syncProviderModels(item: ProviderDTO): Promise<void> {
  const ok = await workspace.syncProviderModels(item.id);
  if (ok) message.success(workspace.message);
  else message.error(workspace.message);
}

async function testModel(item: ModelDTO): Promise<void> {
  const ok = await workspace.testModel(item.provider_id, item.id);
  if (ok) message.success(workspace.message);
  else message.error(workspace.message);
}

async function toggleModel(item: ModelDTO, enabled: boolean): Promise<void> {
  const ok = await workspace.saveModel({ ...item, enabled }, enabled ? 'Model 已启用' : 'Model 已停用');
  if (!ok) message.error(workspace.message);
}

const providerMenuOptions = [
  { label: '测试连接', key: 'test' },
  { label: '同步模型', key: 'sync' },
  { label: '编辑 Provider', key: 'edit' },
  { label: '删除 Provider', key: 'delete', props: { class: 'dropdown-danger' } },
];
const routeMenuOptions = [
  { label: '编辑 Route', key: 'edit' },
  { label: '删除 Route', key: 'delete', props: { class: 'dropdown-danger' } },
];
const modelMenuOptions = [
  { label: '测试模型', key: 'test' },
  { label: '编辑 Model', key: 'edit' },
  { label: '删除 Model', key: 'delete', props: { class: 'dropdown-danger' } },
];

function onProviderMenu(key: string, item: ProviderDTO): void {
  if (key === 'test') void testProvider(item);
  if (key === 'sync') void syncProviderModels(item);
  if (key === 'edit') openProvider(item);
  if (key === 'delete') removeProvider(item);
}

function onRouteMenu(key: string, item: RouteDTO): void {
  if (key === 'edit') openRoute(item);
  if (key === 'delete') removeRoute(item);
}

function onModelMenu(key: string, item: ModelDTO): void {
  if (key === 'test') void testModel(item);
  if (key === 'edit') openModel(item);
  if (key === 'delete') removeModel(item);
}
</script>

<template>
  <div class="page-width">
    <div v-if="workspace.error" class="notice notice-error" role="alert">
      <div>{{ workspace.message }}</div>
      <p v-if="workspace.activationStage === 'error'" class="notice-hint">
        无法确认 Codex 配置是否已恢复。如果 Codex 无法使用,请到「配置档案」页恢复最近备份。
      </p>
    </div>
    <!-- 错误横幅始终内联展示;有数据时页面内容保留,只在无数据可显示时才让错误/骨架独占页面 -->
    <template v-if="workspace.providers.length || (!workspace.error && workspace.phase !== 'loading')">
      <section class="current-card" :class="{ 'is-empty': !workspace.currentRoute }">
        <div class="current-icon">
          {{ workspace.currentProvider?.name.slice(0, 1).toUpperCase() || '·' }}
        </div>
        <div class="current-main">
          <div class="current-title">
            <strong>{{ workspace.currentRoute?.name || '还没有激活路由' }}</strong>
            <NTag v-if="workspace.currentRoute" size="small" type="success" :bordered="false">已激活</NTag>
          </div>
          <span v-if="workspace.currentRoute" class="current-sub mono">
            {{ workspace.currentModel?.name || workspace.currentRoute.model_id }}
            <i>·</i>{{ workspace.currentProvider?.name }} <i>·</i>Profile {{ selectedProfileName || '未选择' }}
          </span>
          <span v-else class="current-sub">在下方 Routes 中激活一个路由,Codex 会使用它作为当前配置</span>
        </div>
        <div class="current-action">
          <NButton
            secondary
            size="small"
            type="primary"
            :loading="workspace.phase === 'saving'"
            @click="workspace.startCodex"
          >
            <Rocket :size="14" />{{ workspace.codexRunning ? '重启 Codex' : '启动 Codex' }}
          </NButton>
        </div>
      </section>

      <section class="resource-grid">
        <NCard class="resource-card resource-card-wide" :bordered="false">
          <div class="card-heading">
            <div>
              <p class="eyebrow">Providers</p>
              <h3>连接服务</h3>
              <p>管理可连接的 Responses API 服务。</p>
            </div>
            <NButton size="small" type="primary" @click="openProvider()"><Plus :size="14" />添加 Provider</NButton>
          </div>
          <NEmpty v-if="!workspace.providers.length" description="还没有 Provider" size="small">
            <template #extra>
              <NButton size="small" type="primary" @click="openProvider()">添加第一个 Provider</NButton>
            </template>
          </NEmpty>
          <div v-else class="resource-list">
            <div
              v-for="item in workspace.providers"
              :key="item.id"
              class="resource-row provider-row"
              :class="{ 'is-active': workspace.currentProvider?.id === item.id }"
            >
              <div class="resource-icon">{{ item.name.slice(0, 1).toUpperCase() }}</div>
              <div class="resource-main">
                <strong
                  >{{ item.name }}<span class="text-muted"> · {{ item.id }}</span></strong
                >
                <span class="mono">{{ item.base_url }} · {{ enabledModelCount(item.id) }} 个启用模型</span>
              </div>
              <div class="resource-side">
                <NTag v-if="workspace.currentProvider?.id === item.id" size="small" type="success" :bordered="false"
                  >当前使用</NTag
                >
                <NTooltip v-else>
                  <template #trigger>
                    <span class="tooltip-target">
                      <NButton
                        size="small"
                        secondary
                        type="primary"
                        :disabled="!enabledModelCount(item.id) || !providerHasRoute(item)"
                        @click="switchProvider(item)"
                        >切换</NButton
                      >
                    </span>
                  </template>
                  {{
                    !enabledModelCount(item.id)
                      ? '先启用模型后才能切换'
                      : !providerHasRoute(item)
                        ? '该 Provider 还没有可用 Route'
                        : `激活 ${providerName(item.id)} 的第一条可用路由`
                  }}
                </NTooltip>
                <NDropdown
                  trigger="click"
                  :options="providerMenuOptions"
                  @select="(key: string) => onProviderMenu(key, item)"
                >
                  <NButton quaternary size="small" aria-label="Provider 更多操作"
                    ><MoreHorizontal :size="15"
                  /></NButton>
                </NDropdown>
              </div>
            </div>
          </div>
        </NCard>

        <NCard class="resource-card" :bordered="false">
          <div class="card-heading">
            <div>
              <p class="eyebrow">Models</p>
              <h3>模型</h3>
              <p>启用后可被 Route 选择。</p>
            </div>
            <NButton size="small" secondary :disabled="!workspace.providers.length" @click="openModel()"
              ><Plus :size="14" />添加</NButton
            >
          </div>
          <NEmpty v-if="!workspace.models.length" description="还没有 Model" size="small">
            <template #extra>
              <NButton size="small" type="primary" :disabled="!workspace.providers.length" @click="openModel()"
                >添加模型</NButton
              >
            </template>
          </NEmpty>
          <div v-else class="resource-list">
            <div
              v-for="item in workspace.models"
              :key="`${item.provider_id}:${item.id}`"
              class="resource-row compact-row"
            >
              <div class="resource-main">
                <strong>{{ item.name || item.id }}</strong>
                <span class="mono">{{ providerName(item.provider_id) }} · {{ item.id }}</span>
              </div>
              <div class="resource-side">
                <NSwitch
                  :value="item.enabled"
                  size="small"
                  :aria-label="`启用模型 ${item.name || item.id}`"
                  @update:value="(enabled: boolean) => toggleModel(item, enabled)"
                />
                <NDropdown
                  trigger="click"
                  :options="modelMenuOptions"
                  @select="(key: string) => onModelMenu(key, item)"
                >
                  <NButton quaternary size="small" aria-label="Model 更多操作"><MoreHorizontal :size="15" /></NButton>
                </NDropdown>
              </div>
            </div>
          </div>
        </NCard>

        <NCard class="resource-card" :bordered="false">
          <div class="card-heading">
            <div>
              <p class="eyebrow">Routes</p>
              <h3>路由目标</h3>
              <p>把 Provider 与 Model 组合成可激活的目标。</p>
            </div>
            <NButton
              size="small"
              type="primary"
              :disabled="!workspace.providers.length || !workspace.models.some((model) => model.enabled)"
              @click="openRoute()"
              ><Plus :size="14" />添加 Route</NButton
            >
          </div>
          <NEmpty v-if="!workspace.routes.length" description="还没有 Route" size="small">
            <template #extra>
              <NButton
                size="small"
                type="primary"
                :disabled="!workspace.providers.length || !workspace.models.some((model) => model.enabled)"
                @click="openRoute()"
                >添加第一个 Route</NButton
              >
            </template>
          </NEmpty>
          <div v-else class="resource-list">
            <div v-for="item in workspace.routes" :key="item.id" class="resource-row route-row">
              <div class="route-indicator" :class="{ 'is-active': item.default }"><Power :size="15" /></div>
              <div class="resource-main">
                <strong>{{ item.name }}</strong>
                <span class="mono"
                  >{{ providerName(item.provider_id) }} / {{ modelName(item.provider_id, item.model_id)
                  }}<template v-if="item.restart_on_activate"> · 激活时重启 Codex</template></span
                >
              </div>
              <div class="resource-side">
                <NTag v-if="item.default" size="small" type="success" :bordered="false">当前</NTag>
                <NButton
                  v-else
                  size="small"
                  secondary
                  type="primary"
                  :loading="workspace.activatingRouteID === item.id"
                  :disabled="workspace.phase === 'activating' || !profileID"
                  @click="activateRoute(item)"
                  >激活</NButton
                >
                <NDropdown
                  trigger="click"
                  :options="routeMenuOptions"
                  @select="(key: string) => onRouteMenu(key, item)"
                >
                  <NButton quaternary size="small" aria-label="Route 更多操作"><MoreHorizontal :size="15" /></NButton>
                </NDropdown>
              </div>
            </div>
          </div>
        </NCard>
      </section>
      <p class="page-footnote">配置写入前会自动保留备份。所有请求都在本机 Wails 进程中完成,凭据只以引用形式使用。</p>
    </template>
    <div v-else-if="workspace.phase === 'loading'" class="loading-grid">
      <NSkeleton v-for="n in 3" :key="n" text :repeat="3" />
    </div>
  </div>

  <NModal
    :show="providerModal"
    preset="card"
    :title="providerDraft ? '编辑 Provider' : '添加 Provider'"
    class="edit-modal"
    :mask-closable="false"
    @update:show="(show: boolean) => guardClose(providerDirty, (next: boolean) => (providerModal = next), show)"
  >
    <NForm
      ref="providerFormRef"
      :model="providerForm"
      :rules="providerRules"
      label-placement="top"
      @submit.prevent="saveProvider"
    >
      <div
        v-if="providerErrors.length"
        :ref="providerSetSummaryRef"
        class="form-error-summary"
        tabindex="-1"
        role="alert"
      >
        <ul>
          <li v-for="item in providerErrors" :key="`${item.field}:${item.message}`">
            <button type="button" @click="providerFocusField(item.field)">{{ item.message }}</button>
          </li>
        </ul>
      </div>
      <NFormItem label="标识" path="id">
        <NInput
          :ref="(el: unknown) => providerRegisterField('id', el)"
          v-model:value="providerForm.id"
          :disabled="!!providerDraft"
          placeholder="例如 openai"
        />
      </NFormItem>
      <NFormItem label="显示名称" path="name">
        <NInput
          :ref="(el: unknown) => providerRegisterField('name', el)"
          v-model:value="providerForm.name"
          placeholder="例如 OpenAI"
        />
      </NFormItem>
      <NFormItem label="Responses API 地址" path="baseURL">
        <NInput
          :ref="(el: unknown) => providerRegisterField('baseURL', el)"
          v-model:value="providerForm.baseURL"
          placeholder="https://api.example.com/v1"
        />
      </NFormItem>
      <NFormItem label="凭据引用" path="authRef">
        <NInput v-model:value="providerForm.authRef" placeholder="环境变量名或 credential:target" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div class="modal-footer">
        <NButton :disabled="workspace.phase === 'saving'" @click="providerModal = false">取消</NButton>
        <NButton type="primary" :loading="workspace.phase === 'saving'" @click="saveProvider">{{
          providerDraft ? '保存' : '添加'
        }}</NButton>
      </div>
    </template>
  </NModal>

  <NModal
    :show="modelModal"
    preset="card"
    :title="modelDraft ? '编辑 Model' : '添加 Model'"
    class="edit-modal"
    :mask-closable="false"
    @update:show="(show: boolean) => guardClose(modelDirty, (next: boolean) => (modelModal = next), show)"
  >
    <NForm ref="modelFormRef" :model="modelForm" :rules="modelRules" label-placement="top" @submit.prevent="saveModel">
      <div v-if="modelErrors.length" :ref="modelSetSummaryRef" class="form-error-summary" tabindex="-1" role="alert">
        <ul>
          <li v-for="item in modelErrors" :key="`${item.field}:${item.message}`">
            <button type="button" @click="modelFocusField(item.field)">{{ item.message }}</button>
          </li>
        </ul>
      </div>
      <NFormItem label="Provider" path="providerID">
        <NSelect
          :ref="(el: unknown) => modelRegisterField('providerID', el)"
          v-model:value="modelForm.providerID"
          :options="providerOptions"
          placeholder="选择 Provider"
        />
      </NFormItem>
      <NFormItem label="模型 ID" path="id">
        <NInput
          :ref="(el: unknown) => modelRegisterField('id', el)"
          v-model:value="modelForm.id"
          :disabled="!!modelDraft"
          placeholder="例如 gpt-5.6-sol"
        />
      </NFormItem>
      <NFormItem label="显示名称" path="name">
        <NInput v-model:value="modelForm.name" placeholder="可选" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div class="modal-footer">
        <NButton :disabled="workspace.phase === 'saving'" @click="modelModal = false">取消</NButton>
        <NButton type="primary" :loading="workspace.phase === 'saving'" @click="saveModel">{{
          modelDraft ? '保存' : '添加'
        }}</NButton>
      </div>
    </template>
  </NModal>

  <NModal
    :show="routeModal"
    preset="card"
    :title="routeDraft ? '编辑 Route' : '添加 Route'"
    class="edit-modal"
    :mask-closable="false"
    @update:show="(show: boolean) => guardClose(routeDirty, (next: boolean) => (routeModal = next), show)"
  >
    <NForm ref="routeFormRef" :model="routeForm" :rules="routeRules" label-placement="top" @submit.prevent="saveRoute">
      <div v-if="routeErrors.length" :ref="routeSetSummaryRef" class="form-error-summary" tabindex="-1" role="alert">
        <ul>
          <li v-for="item in routeErrors" :key="`${item.field}:${item.message}`">
            <button type="button" @click="routeFocusField(item.field)">{{ item.message }}</button>
          </li>
        </ul>
      </div>
      <NFormItem label="路由标识" path="id">
        <NInput
          :ref="(el: unknown) => routeRegisterField('id', el)"
          v-model:value="routeForm.id"
          :disabled="!!routeDraft"
          placeholder="例如 work-sol"
        />
      </NFormItem>
      <NFormItem label="路由名称" path="name">
        <NInput
          :ref="(el: unknown) => routeRegisterField('name', el)"
          v-model:value="routeForm.name"
          placeholder="例如 工作主力"
        />
      </NFormItem>
      <NFormItem label="Provider 与模型" path="modelID">
        <NSelect
          :ref="(el: unknown) => routeRegisterField('modelID', el)"
          v-model:value="routeModelKey"
          :options="modelOptions"
          placeholder="选择已启用的 Model"
        />
      </NFormItem>
      <NFormItem label="激活行为" path="restart">
        <NSwitch v-model:value="routeForm.restart" aria-label="激活时尝试启动 Codex" />
        <span class="switch-copy">激活时尝试启动 Codex</span>
      </NFormItem>
    </NForm>
    <template #footer>
      <div class="modal-footer">
        <NButton :disabled="workspace.phase === 'saving'" @click="routeModal = false">取消</NButton>
        <NButton type="primary" :loading="workspace.phase === 'saving'" @click="saveRoute">{{
          routeDraft ? '保存' : '添加'
        }}</NButton>
      </div>
    </template>
  </NModal>
</template>
