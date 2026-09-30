import { createMemoryHistory, createRouter } from 'vue-router';
import ConfigPage from './pages/ConfigPage.vue';
import ProfilesPage from './pages/ProfilesPage.vue';
import SettingsPage from './pages/SettingsPage.vue';

export const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/config', name: 'config', component: ConfigPage },
    { path: '/profiles', name: 'profiles', component: ProfilesPage },
    { path: '/settings', name: 'settings', component: SettingsPage },
    { path: '/:pathMatch(.*)*', redirect: '/config' },
  ],
});
