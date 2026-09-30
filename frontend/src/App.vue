<script setup lang="ts">
import { onMounted } from 'vue';
import { NConfigProvider, NDialogProvider, NMessageProvider, NNotificationProvider } from 'naive-ui';
import { darkTheme, dateZhCN, zhCN } from 'naive-ui';
import DesktopShell from './components/DesktopShell.vue';
import { themeOverrides } from './app/theme';
import { useWorkspaceStore } from './stores/workspace';

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
