<script setup lang="ts">
import { ref } from 'vue';
import { NButton, NCard, NEmpty, NForm, NFormItem, NInput, NModal, NTag, useDialog, useMessage } from 'naive-ui';
import { ArchiveRestore, Pencil, Plus, Trash2 } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import type { ProfileDTO } from '../services/wails-api';

const workspace = useWorkspaceStore();
const dialog = useDialog();
const message = useMessage();
const modalOpen = ref(false);
const draft = ref<ProfileDTO | null>(null);
const form = ref({ id: '', name: '', configPath: '' });

function openProfile(item?: ProfileDTO): void {
  draft.value = item ? { ...item } : null;
  form.value = item
    ? { id: item.id, name: item.name, configPath: item.config_path ?? '' }
    : { id: '', name: '', configPath: '' };
  modalOpen.value = true;
}

async function saveProfile(): Promise<void> {
  const value = form.value;
  if (!value.id.trim() || !value.name.trim()) {
    message.warning('请填写档案标识和名称');
    return;
  }
  const profile: ProfileDTO = {
    ...(draft.value ?? {}),
    id: value.id.trim(),
    name: value.name.trim(),
    config_path: value.configPath.trim() || undefined,
  };
  if (draft.value) await workspace.saveProfile(profile);
  else await workspace.createProfile(profile);
  if (!workspace.error) {
    workspace.selectedProfileID = profile.id;
    modalOpen.value = false;
  }
}

function removeProfile(item: ProfileDTO): void {
  dialog.warning({
    title: '删除配置档案',
    content: `删除“${item.name}”后不能再用它激活路由，确定继续吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => workspace.deleteProfile(item.id),
  });
}

function restoreProfile(item: ProfileDTO): void {
  dialog.info({
    title: '恢复 Codex 配置',
    content: `将使用“${item.name}”档案的最近备份覆盖当前 Codex 配置，继续吗？`,
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: () => workspace.restoreProfile(item.id),
  });
}
</script>

<template>
  <div class="page-width">
    <section class="page-heading">
      <div>
        <p class="eyebrow">PROFILES</p>
        <h2>配置档案</h2>
        <p>为不同工作环境保留独立的 Codex 配置文件。</p>
      </div>
      <NButton type="primary" size="small" @click="openProfile()"><Plus :size="15" />新建档案</NButton>
    </section>
    <div v-if="workspace.error" class="notice notice-error" role="alert">{{ workspace.message }}</div>
    <NCard class="resource-card" :bordered="false">
      <NEmpty v-if="!workspace.profiles.length" description="还没有配置档案"
        ><template #extra
          ><NButton type="primary" size="small" @click="openProfile()">创建第一个档案</NButton></template
        ></NEmpty
      >
      <div v-else class="resource-list">
        <div v-for="item in workspace.profiles" :key="item.id" class="resource-row profile-row">
          <div class="resource-icon"><ArchiveRestore :size="16" /></div>
          <div class="resource-main">
            <strong>{{ item.name }}</strong
            ><span class="mono">{{ item.config_path || '默认 Codex config.toml' }}</span>
          </div>
          <NTag v-if="workspace.selectedProfileID === item.id" size="small" type="success" :bordered="false"
            >当前使用</NTag
          >
          <NButton
            size="small"
            :secondary="workspace.selectedProfileID !== item.id"
            :type="workspace.selectedProfileID === item.id ? 'primary' : 'default'"
            @click="workspace.selectedProfileID = item.id"
            >选择</NButton
          >
          <NButton quaternary size="small" aria-label="编辑档案" @click="openProfile(item)"
            ><Pencil :size="15"
          /></NButton>
          <NButton quaternary size="small" aria-label="恢复备份" @click="restoreProfile(item)"
            ><ArchiveRestore :size="15"
          /></NButton>
          <NButton quaternary size="small" type="error" aria-label="删除档案" @click="removeProfile(item)"
            ><Trash2 :size="15"
          /></NButton>
        </div>
      </div>
    </NCard>
    <p class="page-footnote">激活路由时会使用当前选中的档案。写入 Codex 前会自动保留备份。</p>
  </div>

  <NModal v-model:show="modalOpen" preset="card" :title="draft ? '编辑配置档案' : '新建配置档案'" class="edit-modal">
    <NForm label-placement="top">
      <NFormItem label="档案标识"
        ><NInput v-model:value="form.id" :disabled="!!draft" placeholder="例如 work"
      /></NFormItem>
      <NFormItem label="显示名称"><NInput v-model:value="form.name" placeholder="例如 工作环境" /></NFormItem>
      <NFormItem label="配置文件路径"
        ><NInput v-model:value="form.configPath" placeholder="相对 CODEX_HOME 的路径（可选）"
      /></NFormItem>
    </NForm>
    <template #footer
      ><div class="modal-footer">
        <NButton @click="modalOpen = false">取消</NButton
        ><NButton type="primary" :loading="workspace.phase === 'saving'" @click="saveProfile">保存档案</NButton>
      </div></template
    >
  </NModal>
</template>
