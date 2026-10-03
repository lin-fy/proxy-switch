<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
  NButton,
  NCard,
  NDropdown,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NTag,
  useDialog,
  useMessage,
  type FormInst,
  type FormRules,
} from 'naive-ui';
import { ArchiveRestore, MoreHorizontal, Plus } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { useFormModal } from '../app/form';
import type { ProfileDTO } from '../services/wails-api';

const workspace = useWorkspaceStore();
const dialog = useDialog();
const message = useMessage();

const modalOpen = ref(false);
const draft = ref<ProfileDTO | null>(null);
const importing = ref(false);
const form = ref({ id: '', name: '', configPath: '' });
const formRef = ref<FormInst | null>(null);

const { fieldErrors, setSummaryRef, registerField, focusField, submit, track, guardClose } = useFormModal();
const tracker = track(() => form.value);
const dirty = computed(() => tracker.dirty());
const importReady = computed(
  () => workspace.codexConfigStatus?.config_exists && workspace.codexConfigStatus.config_valid,
);

const formRules: FormRules = {
  id: { required: true, message: '请填写档案标识', trigger: ['blur', 'input'] },
  name: { required: true, message: '请填写档案名称', trigger: ['blur', 'input'] },
};

const menuOptions = [
  { label: '编辑档案', key: 'edit' },
  { label: '恢复备份', key: 'restore' },
  { label: '删除档案', key: 'delete', props: { class: 'dropdown-danger' } },
];

function openProfile(item?: ProfileDTO): void {
  importing.value = false;
  draft.value = item ? { ...item } : null;
  form.value = item
    ? { id: item.id, name: item.name, configPath: item.config_path ?? '' }
    : { id: '', name: '', configPath: '' };
  fieldErrors.value = [];
  modalOpen.value = true;
  tracker.snapshot();
}

function openImport(): void {
  importing.value = true;
  draft.value = null;
  form.value = { id: 'codex-current', name: '当前 Codex 配置', configPath: '' };
  fieldErrors.value = [];
  modalOpen.value = true;
  tracker.snapshot();
}

async function saveProfile(): Promise<void> {
  await submit(formRef.value, async () => {
    const value = form.value;
    if (importing.value) {
      const ok = await workspace.importCodexConfig(value.id.trim(), value.name.trim());
      if (ok) {
        workspace.selectedProfileID = value.id.trim();
        modalOpen.value = false;
        message.success(workspace.message);
      } else {
        message.error(workspace.message);
      }
      return ok;
    }
    const profile: ProfileDTO = {
      ...(draft.value ?? {}),
      id: value.id.trim(),
      name: value.name.trim(),
      config_path: value.configPath.trim() || undefined,
    };
    const ok = draft.value ? await workspace.saveProfile(profile) : await workspace.createProfile(profile);
    if (ok) {
      workspace.selectedProfileID = profile.id;
      modalOpen.value = false;
      message.success(workspace.message);
    } else {
      message.error(workspace.message);
    }
    return ok;
  });
}

onMounted(() => {
  void workspace.inspectCodexConfig();
});

function removeProfile(item: ProfileDTO): void {
  const isCurrent = workspace.selectedProfileID === item.id;
  dialog.warning({
    title: '删除配置档案',
    content: isCurrent
      ? `“${item.name}”是当前档案,删除后激活路由前需要重新选择档案。确定继续吗?`
      : `删除“${item.name}”后不能再用它激活路由,确定继续吗?`,
    positiveText: '删除',
    negativeText: '取消',
    positiveButtonProps: { type: 'error', secondary: true },
    onPositiveClick: async () => {
      const ok = await workspace.deleteProfile(item.id);
      if (!ok) {
        message.error(workspace.message);
        return false;
      }
      message.success(workspace.message);
      return true;
    },
  });
}

function restoreProfile(item: ProfileDTO): void {
  dialog.info({
    title: '恢复 Codex 配置',
    content: `将使用“${item.name}”档案的最近备份覆盖当前 Codex 配置,继续吗?`,
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: async () => {
      const ok = await workspace.restoreProfile(item.id);
      if (!ok) {
        message.error(workspace.message);
        return false;
      }
      message.success(workspace.message);
      return true;
    },
  });
}

function onMenu(key: string, item: ProfileDTO): void {
  if (key === 'edit') openProfile(item);
  if (key === 'restore') restoreProfile(item);
  if (key === 'delete') removeProfile(item);
}
</script>

<template>
  <div class="page-width">
    <section class="page-heading">
      <div>
        <p class="eyebrow">Profiles</p>
        <h2>配置档案</h2>
        <p>为不同工作环境保留独立的 Codex 配置文件,激活路由时使用当前选中的档案。</p>
      </div>
      <div class="heading-actions">
        <NButton secondary size="small" :disabled="!importReady" @click="openImport">导入当前 Codex</NButton>
        <NButton type="primary" size="small" @click="openProfile()"><Plus :size="14" />新建档案</NButton>
      </div>
    </section>
    <div v-if="workspace.error" class="notice notice-error" role="alert">{{ workspace.message }}</div>
    <NCard class="resource-card" :bordered="false">
      <NEmpty v-if="!workspace.profiles.length" description="还没有配置档案" size="small">
        <template #extra>
          <NButton type="primary" size="small" @click="openProfile()">创建第一个档案</NButton>
        </template>
      </NEmpty>
      <div v-else class="resource-list">
        <div v-for="item in workspace.profiles" :key="item.id" class="resource-row">
          <div class="resource-icon"><ArchiveRestore :size="15" /></div>
          <div class="resource-main">
            <strong
              >{{ item.name }}<span class="text-muted"> · {{ item.id }}</span></strong
            >
            <span class="mono">{{ item.config_path || '默认 Codex config.toml' }}</span>
          </div>
          <div class="resource-side">
            <NTag v-if="workspace.selectedProfileID === item.id" size="small" type="success" :bordered="false"
              >当前使用</NTag
            >
            <NButton v-else size="small" secondary type="primary" @click="workspace.selectedProfileID = item.id"
              >选择</NButton
            >
            <NDropdown trigger="click" :options="menuOptions" @select="(key: string) => onMenu(key, item)">
              <NButton quaternary size="small" aria-label="档案更多操作"><MoreHorizontal :size="15" /></NButton>
            </NDropdown>
          </div>
        </div>
      </div>
    </NCard>
    <p class="page-footnote">写入 Codex 前会自动保留备份,恢复操作使用对应档案的最近备份。</p>
  </div>

  <NModal
    :show="modalOpen"
    preset="card"
    :title="importing ? '导入当前 Codex 配置' : draft ? '编辑配置档案' : '新建配置档案'"
    class="edit-modal"
    :mask-closable="false"
    @update:show="(show: boolean) => guardClose(dirty, (next: boolean) => (modalOpen = next), show)"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top" @submit.prevent="saveProfile">
      <div v-if="fieldErrors.length" :ref="setSummaryRef" class="form-error-summary" tabindex="-1" role="alert">
        <ul>
          <li v-for="item in fieldErrors" :key="`${item.field}:${item.message}`">
            <button type="button" @click="focusField(item.field)">{{ item.message }}</button>
          </li>
        </ul>
      </div>
      <div v-if="importing" class="notice notice-info">
        只复制当前 config.toml 到新档案,不会复制或修改 auth.json。未知配置、MCP 与登录语义会原样保留。
      </div>
      <div v-if="!draft && !importing && workspace.pendingProfile?.id === form.id" class="notice notice-error">
        档案已创建,但保存配置路径失败。再次保存将只补存路径,不会重复创建档案。
      </div>
      <NFormItem label="档案标识" path="id">
        <NInput
          :ref="(el: unknown) => registerField('id', el)"
          v-model:value="form.id"
          :disabled="!!draft"
          placeholder="例如 work"
        />
      </NFormItem>
      <NFormItem label="显示名称" path="name">
        <NInput
          :ref="(el: unknown) => registerField('name', el)"
          v-model:value="form.name"
          placeholder="例如 工作环境"
        />
      </NFormItem>
      <NFormItem v-if="!importing" label="配置文件路径" path="configPath">
        <NInput v-model:value="form.configPath" placeholder="相对 CODEX_HOME 的路径(可选)" />
      </NFormItem>
    </NForm>
    <template #footer>
      <div class="modal-footer">
        <NButton :disabled="workspace.phase === 'saving'" @click="modalOpen = false">取消</NButton>
        <NButton type="primary" :loading="workspace.phase === 'saving'" @click="saveProfile">{{
          importing ? '导入配置' : draft ? '保存档案' : '创建档案'
        }}</NButton>
      </div>
    </template>
  </NModal>
</template>
