import { createMemoryHistory, createRouter } from 'vue-router';
import ConfigPage from './pages/ConfigPage.vue';

export const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/config', name: 'config', component: ConfigPage },
    { path: '/profiles', name: 'profiles', component: () => import('./pages/ProfilesPage.vue') },
    { path: '/settings', name: 'settings', component: () => import('./pages/SettingsPage.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/config' },
  ],
});
