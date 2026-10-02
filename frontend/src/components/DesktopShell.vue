<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { NButton, NSelect, NTag } from 'naive-ui';
import { Activity, Boxes, CircleHelp, Cog, RefreshCw, SlidersHorizontal } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';
import { isPreviewMode } from '../services/wails-api';

const workspace = useWorkspaceStore();
const appVersion = __APP_VERSION__;
const route = useRoute();
const activeLabel = computed(
  () => ({ config: '当前配置', profiles: '配置档案', settings: '设置' })[String(route.name)] ?? '当前配置',
);
const nav = [
  { name: 'config', label: '当前配置', hint: 'Provider 与路由', icon: SlidersHorizontal },
  { name: 'profiles', label: '配置档案', hint: '切换 Codex 环境', icon: Boxes },
  { name: 'settings', label: '设置', hint: '应用与数据', icon: Cog },
];
const profileOptions = computed(() => workspace.profiles.map((item) => ({ label: item.name, value: item.id })));
const statusLightClass = computed(() => {
  if (workspace.error) return 'is-error';
  if (workspace.busy) return 'is-busy';
  if (isPreviewMode) return 'is-idle';
  return '';
});
const statusText = computed(() => {
  if (workspace.busy) return '正在同步';
  if (workspace.error) return '需要处理';
  if (isPreviewMode) return '浏览器预览';
  return '本地已连接';
});
</script>

<template>
  <div class="desktop-shell bg-surface-0">
    <aside class="desktop-rail bg-surface-1">
      <div class="rail-brand">
        <div class="brand-mark"><Activity :size="18" /></div>
        <div><strong>Codex Hub</strong><span>本地模型路由</span></div>
      </div>
      <div class="rail-section-label">工作台</div>
      <nav class="rail-nav" aria-label="主导航">
        <RouterLink
          v-for="item in nav"
          :key="item.name"
          :to="{ name: item.name }"
          class="nav-item"
          :class="{ 'is-active': route.name === item.name }"
          :aria-current="route.name === item.name ? 'page' : undefined"
        >
          <component :is="item.icon" :size="17" stroke-width="1.8" />
          <span
            ><b>{{ item.label }}</b
            ><small>{{ item.hint }}</small></span
          >
        </RouterLink>
      </nav>
      <div class="rail-spacer" />
      <div class="rail-status">
        <span class="status-light" :class="statusLightClass" />
        <div>
          <b>{{ statusText }}</b
          ><small>{{ isPreviewMode ? '演示数据,不写入真实配置' : 'Wails · Windows' }}</small>
        </div>
      </div>
      <div class="rail-version">
        <CircleHelp :size="14" /> <span>Codex Provider Hub · {{ appVersion }}</span>
      </div>
    </aside>
    <section class="desktop-content min-w-0">
      <header class="desktop-topbar">
        <div>
          <span class="topbar-kicker">Workspace</span>
          <h1>{{ activeLabel }}</h1>
        </div>
        <div class="topbar-actions flex items-center gap-2">
          <NSelect
            v-model:value="workspace.selectedProfileID"
            class="topbar-profile"
            :options="profileOptions"
            size="small"
            placeholder="当前档案"
            :disabled="!workspace.profiles.length"
            aria-label="当前 Codex 配置档案"
          />
          <NTag size="small" :bordered="false" :type="workspace.codexRunning ? 'success' : 'default'">{{
            workspace.codexRunning ? 'Codex 运行中' : 'Codex 未运行'
          }}</NTag>
          <NButton
            quaternary
            size="small"
            :loading="workspace.phase === 'loading'"
            aria-label="刷新配置"
            @click="() => workspace.refresh()"
            ><RefreshCw :size="15"
          /></NButton>
        </div>
      </header>
      <main class="desktop-main min-w-0">
        <RouterView />
      </main>
      <footer class="desktop-footer" aria-live="polite">
        <span class="status-light" :class="statusLightClass" />
        <span class="desktop-footer-message">{{ workspace.message }}</span>
        <NTag v-if="isPreviewMode" size="tiny" type="warning" :bordered="false">预览数据</NTag>
        <span class="desktop-footer-spacer" />
        <span class="desktop-footer-meta mono">{{ workspace.phase === 'ready' ? '已同步' : workspace.phase }}</span>
      </footer>
    </section>
  </div>
</template>
