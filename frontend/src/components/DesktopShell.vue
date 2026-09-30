<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { NButton, NSelect, NTag } from 'naive-ui';
import { Activity, Boxes, CircleHelp, Cog, RefreshCw, Rocket, SlidersHorizontal } from 'lucide-vue-next';
import { useWorkspaceStore } from '../stores/workspace';

const workspace = useWorkspaceStore();
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
</script>

<template>
  <div class="desktop-shell min-h-screen bg-surface-0">
    <aside class="desktop-rail bg-surface-1">
      <div class="rail-brand">
        <div class="brand-mark"><Activity :size="18" /></div>
        <div><strong>Codex Hub</strong><span>本地模型路由</span></div>
      </div>
      <div class="rail-section-label">工作台</div>
      <nav class="rail-nav flex flex-col gap-1" aria-label="主导航">
        <RouterLink
          v-for="item in nav"
          :key="item.name"
          :to="{ name: item.name }"
          class="nav-item"
          :class="{ 'is-active': route.name === item.name }"
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
        <span class="status-light" :class="{ 'is-busy': workspace.busy, 'is-error': workspace.error }" />
        <div>
          <b>{{ workspace.busy ? '正在同步' : workspace.error ? '需要处理' : '本地已连接' }}</b
          ><small>Wails · Windows</small>
        </div>
      </div>
      <div class="rail-version"><CircleHelp :size="14" /> <span>Codex Provider Hub · 1.0</span></div>
    </aside>
    <section class="desktop-content min-w-0">
      <header class="desktop-topbar">
        <div>
          <span class="topbar-kicker">WORKSPACE</span>
          <h1>{{ activeLabel }}</h1>
        </div>
        <div class="topbar-actions flex items-center gap-2">
          <NSelect
            v-model:value="workspace.selectedProfileID"
            class="topbar-profile"
            :options="profileOptions"
            size="small"
            placeholder="当前档案"
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
            @click="workspace.refresh"
            ><RefreshCw :size="15"
          /></NButton>
          <NButton secondary size="small" :loading="workspace.busy" @click="workspace.startCodex"
            ><Rocket :size="15" /> {{ workspace.codexRunning ? '重启 Codex' : '启动 Codex' }}</NButton
          >
        </div>
      </header>
      <main class="desktop-main min-w-0">
        <RouterView />
      </main>
    </section>
  </div>
</template>
