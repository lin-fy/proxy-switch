<script setup lang="ts">
import { computed } from 'vue';
import { NButton, NCard, NSelect, NSwitch, NTag, useDialog, useMessage } from 'naive-ui';
import { ExternalLink, FolderSync, Info, Play, RefreshCw, ShieldCheck } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';

const workspace = useWorkspaceStore();
const dialog = useDialog();
const message = useMessage();
const profileOptions = computed(() => workspace.profiles.map((item) => ({ label: item.name, value: item.id })));

async function restoreSelected(): Promise<void> {
  if (!workspace.selectedProfileID) {
    message.warning('请先选择一个配置档案');
    return;
  }
  dialog.info({
    title: '恢复最近备份',
    content: '这会覆盖当前 Codex 配置文件，确定继续吗？',
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: () => workspace.restoreProfile(workspace.selectedProfileID),
  });
}
</script>

<template>
  <div class="page-width settings-page">
    <section class="page-heading">
      <div>
        <p class="eyebrow">SYSTEM</p>
        <h2>设置</h2>
        <p>管理应用行为、Codex 启动和本地数据。</p>
      </div>
    </section>
    <div v-if="workspace.error" class="notice notice-error" role="alert">{{ workspace.message }}</div>
    <section class="settings-grid">
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><ShieldCheck :size="17" />应用</div></template
        >
        <div class="setting-row">
          <div><strong>开机自动启动</strong><span>登录 Windows 后在后台运行</span></div>
          <NSwitch
            :value="workspace.autostart"
            :loading="workspace.phase === 'saving'"
            @update:value="workspace.setAutostart"
          />
        </div>
        <div class="setting-row">
          <div><strong>关闭窗口时保留托盘</strong><span>Wails 桌面窗口行为由应用壳层控制</span></div>
          <NTag size="small" :bordered="false" type="success">已启用</NTag>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><Play :size="17" />Codex</div></template
        >
        <div class="setting-row">
          <div><strong>当前配置档案</strong><span>激活路由和恢复备份使用这个档案</span></div>
          <NSelect
            v-model:value="workspace.selectedProfileID"
            :options="profileOptions"
            size="small"
            placeholder="选择档案"
          />
        </div>
        <div class="setting-row">
          <div>
            <strong>运行状态</strong><span>{{ workspace.codexRunning ? 'Codex 当前正在运行' : 'Codex 尚未运行' }}</span>
          </div>
          <NButton size="small" secondary :loading="workspace.busy" @click="workspace.startCodex"
            ><Play :size="14" />{{ workspace.codexRunning ? '重新启动' : '启动 Codex' }}</NButton
          >
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><FolderSync :size="17" />数据</div></template
        >
        <div class="setting-row">
          <div><strong>恢复最近备份</strong><span>恢复当前档案写入前保留的 Codex 配置</span></div>
          <NButton size="small" secondary :disabled="!workspace.selectedProfileID" @click="restoreSelected"
            ><RefreshCw :size="14" />恢复备份</NButton
          >
        </div>
        <div class="setting-row">
          <div><strong>配置存储</strong><span class="mono">CODEX_HOME · 本机用户目录</span></div>
          <NTag size="small" :bordered="false">本地</NTag>
        </div>
      </NCard>
      <NCard class="resource-card" :bordered="false"
        ><template #header
          ><div class="section-title"><Info :size="17" />关于</div></template
        >
        <div class="setting-row">
          <div><strong>Codex Provider Hub</strong><span>Windows 桌面工具 · Wails 3 + Vue 3</span></div>
          <NTag size="small" :bordered="false">V1</NTag>
        </div>
        <div class="setting-row">
          <div><strong>本地运行</strong><span>所有请求在本机 Wails 进程中完成</span></div>
          <ExternalLink :size="16" class="text-muted" />
        </div>
      </NCard>
    </section>
  </div>
</template>
