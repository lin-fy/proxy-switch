<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { NButton, NCard, NInput, NModal, NSelect, NSwitch, NTag, useDialog, useMessage } from 'naive-ui';
import { ExternalLink, FolderSync, Info, Play, RefreshCw, ShieldCheck } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { isPreviewMode } from '../services/wails-api';
import type { WorkspaceImportPreviewDTO } from '../services/wails-api';

const workspace = useWorkspaceStore();
const dialog = useDialog();
const message = useMessage();
const profileOptions = computed(() => workspace.profiles.map((item) => ({ label: item.name, value: item.id })));
const authModeLabel = computed(() => {
  const mode = workspace.codexConfigStatus?.auth_mode;
  if (mode === 'oauth') return '官方 OAuth';
  if (mode === 'api_key') return 'API Key 引用';
  if (mode === 'present') return '已配置认证';
  if (mode === 'invalid') return 'auth.json 格式错误';
  if (mode === 'missing') return '未登录';
  return '未检查';
});
const authModeType = computed<'default' | 'success' | 'warning' | 'error'>(() => {
  const mode = workspace.codexConfigStatus?.auth_mode;
  if (mode === 'oauth' || mode === 'api_key' || mode === 'present') return 'success';
  if (mode === 'invalid') return 'error';
  return 'warning';
});
const importModal = ref(false);
const importPayload = ref('');
const importPreview = ref<WorkspaceImportPreviewDTO | null>(null);
const importCanCommit = computed(() => {
  const preview = importPreview.value;
  if (!preview?.valid) return false;
  return ![
    ...(preview.provider_conflicts ?? []),
    ...(preview.model_conflicts ?? []),
    ...(preview.route_conflicts ?? []),
    ...(preview.profile_conflicts ?? []),
  ].length;
});
const importIssues = computed(() => {
  const preview = importPreview.value;
  if (!preview) return [];
  return [
    ...(preview.errors ?? []),
    ...(preview.missing_references ?? []),
    ...(preview.provider_conflicts ?? []).map((item) => `Provider ID 冲突: ${item}`),
    ...(preview.model_conflicts ?? []).map((item) => `Model ID 冲突: ${item}`),
    ...(preview.route_conflicts ?? []).map((item) => `Route ID 冲突: ${item}`),
    ...(preview.profile_conflicts ?? []).map((item) => `Profile ID 冲突: ${item}`),
  ];
});

onMounted(() => {
  void workspace.inspectCodexConfig();
});

async function restoreSelected(): Promise<void> {
  if (!workspace.selectedProfileID) {
    message.warning('请先选择一个配置档案');
    return;
  }
  dialog.info({
    title: '恢复最近备份',
    content: '这会覆盖当前 Codex 配置文件,确定继续吗?',
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: async () => {
      const ok = await workspace.restoreProfile(workspace.selectedProfileID);
      if (!ok) {
        message.error(workspace.message);
        return false;
      }
      message.success(workspace.message);
      return true;
    },
  });
}

async function exportWorkspace(): Promise<void> {
  const payload = await workspace.exportWorkspace();
  if (!payload) {
    message.error(workspace.message);
    return;
  }
  const blob = new globalThis.Blob([payload], { type: 'application/json' });
  const url = globalThis.URL.createObjectURL(blob);
  const anchor = globalThis.document.createElement('a');
  anchor.href = url;
  anchor.download = `proxy-switch-export-${new Date().toISOString().slice(0, 10)}.json`;
  anchor.click();
  globalThis.URL.revokeObjectURL(url);
  message.success('配置已导出（不含认证内容）');
}

function openImportPreview(): void {
  importPayload.value = '';
  importPreview.value = null;
  importModal.value = true;
}

async function previewImport(): Promise<void> {
  importPreview.value = await workspace.previewWorkspaceImport(importPayload.value);
  if (!importPreview.value) message.error(workspace.message);
}

async function confirmImport(): Promise<void> {
  const result = await workspace.importWorkspace(importPayload.value);
  if (!result) {
    message.error(workspace.message);
    return;
  }
  importPreview.value = result;
  if (result.valid) {
    importModal.value = false;
    message.success(workspace.message);
  } else {
    message.error(workspace.message);
  }
}
</script>

<template>
  <div class="page-width settings-page">
    <section class="page-heading">
      <div>
        <p class="eyebrow">Settings</p>
        <h2>设置</h2>
        <p>管理应用行为、Codex 启动和本地数据。</p>
      </div>
    </section>
    <div v-if="workspace.error" class="notice notice-error" role="alert">{{ workspace.message }}</div>
    <section class="settings-grid">
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><ShieldCheck :size="16" />应用</div></template
        >
        <div class="setting-row">
          <div><strong>开机自动启动</strong><span>登录 Windows 后在后台运行</span></div>
          <NSwitch
            :value="workspace.autostart"
            :loading="workspace.phase === 'saving'"
            aria-label="开机自动启动"
            @update:value="workspace.setAutostart"
          />
        </div>
        <div class="setting-row">
          <div><strong>关闭窗口时保留托盘</strong><span>关闭窗口后从系统托盘恢复,不退出应用</span></div>
          <NTag size="small" :bordered="false" type="success">已启用</NTag>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><Play :size="16" />Codex</div></template
        >
        <div class="setting-row">
          <div><strong>当前配置档案</strong><span>激活路由和恢复备份使用这个档案</span></div>
          <NSelect
            v-model:value="workspace.selectedProfileID"
            :options="profileOptions"
            size="small"
            placeholder="选择档案"
            :disabled="!workspace.profiles.length"
            aria-label="当前配置档案"
          />
        </div>
        <div class="setting-row">
          <div>
            <strong>运行状态</strong><span>{{ workspace.codexRunning ? 'Codex 当前正在运行' : 'Codex 尚未运行' }}</span>
          </div>
          <NButton size="small" secondary :loading="workspace.busy" @click="workspace.startCodex"
            ><Play :size="13" />{{ workspace.codexRunning ? '重新启动' : '启动 Codex' }}</NButton
          >
        </div>
        <div class="setting-row">
          <div><strong>登录状态</strong><span>仅读取本机 auth.json 的存在性和类型,不会显示认证内容</span></div>
          <NTag size="small" :bordered="false" :type="authModeType">{{ authModeLabel }}</NTag>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><FolderSync :size="16" />数据</div></template
        >
        <div class="setting-row">
          <div><strong>恢复最近备份</strong><span>恢复当前档案写入前保留的 Codex 配置</span></div>
          <NButton size="small" secondary :disabled="!workspace.selectedProfileID" @click="restoreSelected"
            ><RefreshCw :size="13" />恢复备份</NButton
          >
        </div>
        <div class="setting-row">
          <div><strong>配置存储</strong><span class="mono">CODEX_HOME · 本机用户目录</span></div>
          <NTag size="small" :bordered="false">本地</NTag>
        </div>
        <div class="setting-row">
          <div>
            <strong>导出配置</strong><span>Provider/Model/Route/Profile JSON，不含请求头、查询参数或认证内容</span>
          </div>
          <NButton size="small" secondary @click="exportWorkspace">导出 JSON</NButton>
        </div>
        <div class="setting-row">
          <div><strong>导入预览</strong><span>粘贴导出 JSON，只解析冲突和错误，不会写入本地状态</span></div>
          <NButton size="small" secondary @click="openImportPreview">打开预览</NButton>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><Info :size="16" />关于</div></template
        >
        <div class="setting-row">
          <div><strong>Proxy Switch</strong><span>Windows 桌面工具 · V1</span></div>
          <NTag size="small" :bordered="false">1.0</NTag>
        </div>
        <div class="setting-row">
          <div>
            <strong>运行环境</strong><span>{{ isPreviewMode ? '浏览器预览,使用演示数据' : 'Wails 3 桌面进程' }}</span>
          </div>
          <ExternalLink :size="15" class="text-muted" />
        </div>
      </NCard>
    </section>
  </div>

  <NModal v-model:show="importModal" preset="card" title="导入配置预览" class="edit-modal">
    <NInput
      v-model:value="importPayload"
      type="textarea"
      :autosize="{ minRows: 8, maxRows: 16 }"
      placeholder="粘贴由 Proxy Switch 导出的 JSON"
      aria-label="导入配置 JSON"
    />
    <div v-if="importPreview" class="notice" :class="importPreview.valid ? '' : 'notice-error'">
      <strong>{{ importPreview.valid ? '格式和引用检查通过' : '无法导入' }}</strong>
      <span>
        Provider {{ importPreview.provider_count }} · Model {{ importPreview.model_count }} · Route
        {{ importPreview.route_count }} · Profile {{ importPreview.profile_count }}
      </span>
      <ul v-if="importIssues.length">
        <li v-for="item in importIssues" :key="item">
          {{ item }}
        </li>
      </ul>
      <p v-if="importPreview.valid" class="notice-hint">预览通过后可确认导入；仅新增资源，任何 ID 冲突都会全量拒绝。</p>
    </div>
    <template #footer>
      <div class="modal-footer">
        <NButton @click="importModal = false">关闭</NButton>
        <NButton type="primary" :disabled="!importPayload.trim()" @click="previewImport">检查 JSON</NButton>
        <NButton type="primary" :disabled="!importCanCommit" @click="confirmImport">确认导入（仅新增）</NButton>
      </div>
    </template>
  </NModal>
</template>
