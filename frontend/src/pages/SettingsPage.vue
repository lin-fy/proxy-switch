<script setup lang="ts">
import { computed } from 'vue';
import { NButton, NCard, NSelect, NSwitch, NTag, useDialog, useMessage } from 'naive-ui';
import { ExternalLink, FolderSync, Info, Play, RefreshCw, ShieldCheck } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { isPreviewMode } from '../services/wails-api';

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
</template>
