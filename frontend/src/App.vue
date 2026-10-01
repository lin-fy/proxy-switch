<script setup lang="ts">
import { onMounted } from 'vue';
import { NConfigProvider, NDialogProvider, NMessageProvider, NNotificationProvider } from 'naive-ui';
import { darkTheme, dateZhCN, zhCN } from 'naive-ui';
import DesktopShell from './components/DesktopShell.vue';
import { createThemeOverrides } from './app/theme';
import { useWorkspaceStore } from './stores/workspace';

// setup 阶段 tokens.css 已生效,可以读取令牌计算值生成主题。
const themeOverrides = createThemeOverrides();
const workspace = useWorkspaceStore();
onMounted(() => {
  void workspace.refresh();
});
</script>

<template>
  <NConfigProvider :theme="darkTheme" :locale="zhCN" :date-locale="dateZhCN" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <DesktopShell />
        </NNotificationProvider>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>
