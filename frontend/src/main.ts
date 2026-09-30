import './style.css';

import * as App from '../bindings/codex-provider-hub/internal/interfaces/wails/app';
import type { ModelDTO, ProfileDTO, ProviderDTO, RouteDTO } from '../bindings/codex-provider-hub/internal/interfaces/wails/models';

type State = { providers: ProviderDTO[]; models: ModelDTO[]; routes: RouteDTO[]; profiles: ProfileDTO[]; autostart: boolean; busy: boolean; message: string; error: boolean };
const state: State = { providers: [], models: [], routes: [], profiles: [], autostart: false, busy: false, message: '准备连接本地 Codex 配置', error: false };
const app = document.querySelector<HTMLDivElement>('#app');
if (!app) throw new Error('app root not found');

const escapeHTML = (value: string | undefined): string => (value ?? '').replace(/[&<>'"]/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[character] ?? character));
const formValue = (form: HTMLFormElement, name: string): string => { const value = new FormData(form).get(name); return typeof value === 'string' ? value.trim() : ''; };
const setMessage = (message: string, error = false): void => { state.message = message; state.error = error; render(); };

const refresh = async (): Promise<void> => {
    state.busy = true; render();
    try {
        const [providers, models, routes, profiles, autostart] = await Promise.all([App.ListProviders(), App.ListModels(), App.ListRoutes(), App.ListProfiles(), App.AutostartEnabled()]);
        state.providers = providers ?? []; state.models = models ?? []; state.routes = routes ?? []; state.profiles = profiles ?? [];
        state.autostart = autostart;
        state.message = '已同步本地配置'; state.error = false;
    } catch (error) { state.message = error instanceof Error ? error.message : '读取配置失败'; state.error = true; }
    finally { state.busy = false; render(); }
};

const providerOptions = (): string => state.providers.length > 0 ? state.providers.map((item) => `<option value="${escapeHTML(item.id)}">${escapeHTML(item.name)}</option>`).join('') : '<option value="">先添加 Provider</option>';
const modelOptions = (): string => state.models.length > 0 ? state.models.filter((item) => item.enabled).map((item) => { const provider = state.providers.find((candidate) => candidate.id === item.provider_id); return `<option value="${escapeHTML(item.provider_id)}:${escapeHTML(item.id)}">${escapeHTML(provider?.name ?? item.provider_id)} / ${escapeHTML(item.name || item.id)}</option>`; }).join('') : '<option value="">先添加 Model</option>';
const profileOptions = (): string => state.profiles.length > 0 ? state.profiles.map((item) => `<option value="${escapeHTML(item.id)}">${escapeHTML(item.name)}</option>`).join('') : '<option value="">先添加配置档案</option>';

const render = (): void => {
    const activeRoute = state.routes.find((item) => item.default) ?? state.routes[0];
    const activeProvider = activeRoute && state.providers.find((item) => item.id === activeRoute.provider_id);
    const activeModel = activeRoute && state.models.find((item) => item.provider_id === activeRoute.provider_id && item.id === activeRoute.model_id);
    app.innerHTML = `
      <main class="shell">
        <aside class="rail"><div class="brand"><span class="brand-mark">CP</span><div><strong>Codex Provider Hub</strong><small>本地路由工作台</small></div></div><div class="rail-rule"></div><p class="eyebrow">当前路由</p><div class="current-route"><span class="status-dot"></span><div><strong>${escapeHTML(activeRoute?.name ?? '尚未激活')}</strong><small>${escapeHTML(activeProvider?.name ?? '选择一个 Provider')} · ${escapeHTML(activeModel?.name ?? '选择一个 Model')}</small></div></div><div class="rail-foot"><span class="pulse"></span>${state.busy ? '同步中…' : '本地配置已就绪'}</div></aside>
        <section class="workspace"><header class="topbar"><div><p class="eyebrow">Provider routing / Windows</p><h1>把模型接入 Codex</h1><p class="lede">Provider 只需配置一次，路由负责决定 Codex 当前使用谁。</p></div><button class="ghost-button" data-action="refresh" ${state.busy ? 'disabled' : ''}>${state.busy ? '同步中…' : '刷新配置'}</button></header>
          <div class="notice ${state.error ? 'notice-error' : ''}"><span class="notice-icon">${state.error ? '!' : 'i'}</span>${escapeHTML(state.message)}</div>
          <div class="grid">
            <section class="panel panel-wide"><div class="panel-heading"><div><span class="step">01</span><h2>Providers</h2><p>CPA、OpenAI 或任意 Responses API 中转。</p></div><span class="count">${state.providers.length} 个</span></div><form class="inline-form" data-form="provider"><input name="id" placeholder="标识，如 cpa" required><input name="name" placeholder="显示名称" required><input name="baseURL" placeholder="https://api.example.com/v1" required><input name="authRef" placeholder="API Key 环境变量（可选）"><button class="primary-button" type="submit">添加 Provider</button></form><div class="item-list">${state.providers.length ? state.providers.map((item) => `<div class="item-row"><div class="item-icon">${escapeHTML(item.name.slice(0, 1).toUpperCase())}</div><div class="item-main"><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.base_url)} · ${escapeHTML(item.protocol)}</small></div><button class="icon-button" data-action="delete-provider" data-id="${escapeHTML(item.id)}" aria-label="删除 Provider">×</button></div>`).join('') : '<div class="empty">还没有 Provider。添加一个 Responses API 地址开始。</div>'}</div></section>
            <section class="panel"><div class="panel-heading"><div><span class="step">02</span><h2>Models</h2><p>这些模型会同步到 Codex 下拉列表。</p></div><span class="count">${state.models.length} 个</span></div><form class="stack-form" data-form="model"><label>Provider<select name="providerID" required>${providerOptions()}</select></label><div class="form-pair"><input name="id" placeholder="模型 ID，如 gpt-5.6-sol" required><input name="name" placeholder="显示名称（可选）"></div><button class="primary-button" type="submit">添加 Model</button></form><div class="compact-list">${state.models.length ? state.models.map((item) => `<div class="compact-row"><div><strong>${escapeHTML(item.name || item.id)}</strong><small>${escapeHTML(item.provider_id)} / ${escapeHTML(item.id)}</small></div><button class="icon-button" data-action="delete-model" data-provider="${escapeHTML(item.provider_id)}" data-id="${escapeHTML(item.id)}">×</button></div>`).join('') : '<div class="empty">添加 Provider 后，在这里登记可用模型。</div>'}</div></section>
            <section class="panel"><div class="panel-heading"><div><span class="step">03</span><h2>Routes</h2><p>激活后写入 Codex 配置。</p></div><span class="count">${state.routes.length} 个</span></div><form class="stack-form" data-form="route"><div class="form-pair"><input name="id" placeholder="路由 ID" required><input name="name" placeholder="路由名称" required></div><label>模型<select name="modelKey" required>${modelOptions()}</select></label><label>配置档案<select name="profileID" required>${profileOptions()}</select></label><label class="check"><input type="checkbox" name="restart"> 激活时尝试启动 Codex</label><button class="primary-button" type="submit">保存并激活</button></form><div class="compact-list">${state.routes.length ? state.routes.map((item) => `<div class="route-row"><div><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.provider_id)} / ${escapeHTML(item.model_id)}</small></div><button class="activate-button" data-action="activate" data-route="${escapeHTML(item.id)}">激活</button></div>`).join('') : '<div class="empty">创建一个路由，把 Provider 和 Model 组合起来。</div>'}</div></section>
	            <section class="panel panel-wide profile-panel"><div class="panel-heading"><div><span class="step">04</span><h2>配置档案</h2><p>保存不同的 Codex 配置组合，需要时一键恢复原始配置。</p></div><span class="count">${state.profiles.length} 个</span></div><form class="inline-form" data-form="profile"><input name="id" placeholder="档案 ID，如 default" required><input name="name" placeholder="显示名称" required><button class="primary-button" type="submit">添加档案</button></form><div class="profile-list">${state.profiles.length ? state.profiles.map((item) => `<div class="profile-row"><div><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.config_path || '默认 Codex config.toml')}</small></div><button class="outline-button" data-action="restore" data-profile="${escapeHTML(item.id)}">恢复备份</button><button class="icon-button" data-action="delete-profile" data-id="${escapeHTML(item.id)}">×</button></div>`).join('') : '<div class="empty">至少添加一个配置档案，路由才能被激活。</div>'}</div></section>
	            <section class="panel panel-wide"><div class="panel-heading"><div><span class="step">05</span><h2>Windows 设置</h2><p>关闭主窗口后应用会留在系统托盘中。</p></div><span class="count">托盘已开启</span></div><div class="profile-row"><div><strong>开机时自动启动</strong><small>${state.autostart ? '已开启，登录 Windows 后自动运行' : '当前关闭'}</small></div><button class="${state.autostart ? 'outline-button' : 'primary-button'}" data-action="autostart">${state.autostart ? '关闭' : '开启'}</button></div></section>
          </div><footer class="footer"><span>Codex Provider Hub · Windows V1</span><span>配置写入前自动保留备份</span></footer>
        </section>
      </main>`;
};

app.addEventListener('click', async (event) => {
    const target = event.target as HTMLElement; const action = target.dataset.action; if (!action) return;
    try {
        if (action === 'refresh') await refresh();
        if (action === 'delete-provider') await App.DeleteProvider(target.dataset.id ?? '');
        if (action === 'delete-model') await App.DeleteModel(target.dataset.provider ?? '', target.dataset.id ?? '');
        if (action === 'delete-profile') await App.DeleteProfile(target.dataset.id ?? '');
        if (action === 'activate') { const profile = state.profiles[0]; if (!profile) throw new Error('请先创建配置档案'); await App.ActivateRoute(target.dataset.route ?? '', profile.id); }
	        if (action === 'restore') await App.RestoreCodexConfig(target.dataset.profile ?? '');
	        if (action === 'autostart') await App.SetAutostart(!state.autostart);
        setMessage(action === 'restore' ? '已恢复 Codex 配置备份' : '操作完成'); await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : '操作失败', true); }
});

app.addEventListener('submit', async (event) => {
    event.preventDefault(); const form = event.target as HTMLFormElement; const kind = form.dataset.form;
    try {
        if (kind === 'provider') await App.CreateProvider(formValue(form, 'id'), formValue(form, 'name'), formValue(form, 'baseURL'), formValue(form, 'authRef'));
        if (kind === 'model') await App.CreateModel(formValue(form, 'providerID'), formValue(form, 'id'), formValue(form, 'name'));
        if (kind === 'profile') await App.CreateProfile(formValue(form, 'id'), formValue(form, 'name'));
        if (kind === 'route') { const [providerID, modelID] = formValue(form, 'modelKey').split(':'); const route = await App.CreateRoute(formValue(form, 'id'), formValue(form, 'name'), providerID, modelID); route.restart_on_activate = new FormData(form).get('restart') === 'on'; await App.SaveRoute(route); await App.ActivateRoute(route.id, formValue(form, 'profileID')); }
        form.reset(); setMessage('已保存并同步 Codex 配置'); await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : '保存失败', true); }
});

render();
void refresh();
