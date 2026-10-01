import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

// 桌面端产物通过 go:embed 打进 exe：分包不减少总字节，但把框架/控件库与业务代码
// 分开可以避免业务小改动导致整套库代码 hash 翻新，也让单 chunk 回到警告线以内。
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined;
          if (
            /[\\/]node_modules[\\/](naive-ui|vueuc|vdirs|treemate|seemly|css-render|@css-render|evtd|async-validator|date-fns-tz|date-fns|vooks|@floating-ui)/.test(
              id,
            )
          ) {
            return 'naive-ui';
          }
          if (/[\\/]node_modules[\\/](vue|@vue|vue-router|pinia|@wailsio)[\\/]/.test(id)) {
            return 'framework';
          }
          if (id.includes('lucide-vue-next')) {
            return 'icons';
          }
          return 'vendor';
        },
      },
    },
  },
});
