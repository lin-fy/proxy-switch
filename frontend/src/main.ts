import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { router } from './router';
import { installCloseButtonA11y } from './services/a11y';
import './styles/tokens.css';
import './styles/base.css';
import './styles/desktop.css';

const app = createApp(App);
app.use(createPinia());
app.use(router);
void router.replace('/config');
app.mount('#app');
installCloseButtonA11y();
