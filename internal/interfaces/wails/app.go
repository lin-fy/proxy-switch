package wails

import (
	"context"
	"errors"

	modelapp "codex-provider-hub/internal/application/model"
	"codex-provider-hub/internal/application/ports"
	profileapp "codex-provider-hub/internal/application/profile"
	providerapp "codex-provider-hub/internal/application/provider"
	routeapp "codex-provider-hub/internal/application/route"
	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
	codexadapter "codex-provider-hub/internal/infrastructure/codex"
)

type App struct {
	providers *providerapp.Service
	models    *modelapp.Service
	routes    *routeapp.Service
	profiles  *profileapp.Service
	activator *routeapp.Activator
	codex     *codexadapter.Adapter
	autostart AutostartManager
	tester    ports.ProviderTester
}

type AutostartManager interface {
	Enable() error
	Disable() error
	IsEnabled() (bool, error)
}

var errAutostartUnavailable = errors.New("开机启动服务尚未就绪")

func NewApp(
	providers *providerapp.Service,
	models *modelapp.Service,
	routes *routeapp.Service,
	profiles *profileapp.Service,
	activator *routeapp.Activator,
	codex *codexadapter.Adapter,
	autostart AutostartManager,
	tester ports.ProviderTester,
) *App {
	return &App{providers: providers, models: models, routes: routes, profiles: profiles, activator: activator, codex: codex, autostart: autostart, tester: tester}
}

type ProviderDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	Protocol    string            `json:"protocol"`
	AuthRef     string            `json:"auth_ref,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
}

type ModelDTO struct {
	ProviderID string `json:"provider_id"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
}

type RouteDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PlatformID        string `json:"platform_id"`
	ProviderID        string `json:"provider_id"`
	ModelID           string `json:"model_id"`
	RestartOnActivate bool   `json:"restart_on_activate"`
	Default           bool   `json:"default"`
}

type ProfileDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	RouteID          string `json:"route_id,omitempty"`
	ConfigPath       string `json:"config_path,omitempty"`
	ModelCatalogPath string `json:"model_catalog_path,omitempty"`
}

func (a *App) ListProviders(ctx context.Context) ([]ProviderDTO, error) {
	items, err := a.providers.List(ctx)
	return mapProviders(items), err
}

func (a *App) CreateProvider(ctx context.Context, id, name, baseURL, authRef string) (ProviderDTO, error) {
	item, err := a.providers.Create(ctx, id, name, baseURL, authRef)
	return toProvider(item), err
}

func (a *App) SaveProvider(ctx context.Context, item ProviderDTO) error {
	return a.providers.Save(ctx, provider.Provider{ID: item.ID, Name: item.Name, BaseURL: item.BaseURL, Protocol: item.Protocol, AuthRef: item.AuthRef, Headers: item.Headers, QueryParams: item.QueryParams})
}

func (a *App) DeleteProvider(ctx context.Context, id string) error {
	return a.providers.Delete(ctx, id)
}

func (a *App) TestProvider(ctx context.Context, id string) error {
	if a.tester == nil {
		return errors.New("Provider 连接测试服务尚未就绪")
	}
	item, err := a.providers.Get(ctx, id)
	if err != nil {
		return err
	}
	return a.tester.Test(ctx, item)
}

func (a *App) TestProviderModel(ctx context.Context, providerID, modelID string) error {
	if a.tester == nil {
		return errors.New("Provider 模型测试服务尚未就绪")
	}
	item, err := a.providers.Get(ctx, providerID)
	if err != nil {
		return err
	}
	return a.tester.TestModel(ctx, item, modelID)
}

func (a *App) ListModels(ctx context.Context) ([]ModelDTO, error) {
	items, err := a.models.List(ctx)
	return mapModels(items), err
}

func (a *App) ListModelsByProvider(ctx context.Context, providerID string) ([]ModelDTO, error) {
	items, err := a.models.ListByProvider(ctx, providerID)
	return mapModels(items), err
}

func (a *App) CreateModel(ctx context.Context, providerID, id, name string) (ModelDTO, error) {
	item, err := a.models.Create(ctx, providerID, id, name)
	return toModel(item), err
}

func (a *App) SaveModel(ctx context.Context, item ModelDTO) error {
	return a.models.Save(ctx, model.Model{ProviderID: item.ProviderID, ID: item.ID, Name: item.Name, Enabled: item.Enabled})
}

func (a *App) DeleteModel(ctx context.Context, providerID, id string) error {
	return a.models.Delete(ctx, providerID, id)
}

func (a *App) ListRoutes(ctx context.Context) ([]RouteDTO, error) {
	items, err := a.routes.List(ctx)
	return mapRoutes(items), err
}

func (a *App) CreateRoute(ctx context.Context, id, name, providerID, modelID string) (RouteDTO, error) {
	item, err := a.routes.Create(ctx, id, name, providerID, modelID)
	return toRoute(item), err
}

func (a *App) SaveRoute(ctx context.Context, item RouteDTO) error {
	return a.routes.Save(ctx, route.Route{ID: item.ID, Name: item.Name, PlatformID: item.PlatformID, ProviderID: item.ProviderID, ModelID: item.ModelID, RestartOnActivate: item.RestartOnActivate, Default: item.Default})
}

func (a *App) DeleteRoute(ctx context.Context, id string) error { return a.routes.Delete(ctx, id) }

func (a *App) ActivateRoute(ctx context.Context, routeID, profileID string) error {
	return a.activator.Activate(ctx, routeID, profileID)
}

func (a *App) ListProfiles(ctx context.Context) ([]ProfileDTO, error) {
	items, err := a.profiles.List(ctx)
	return mapProfiles(items), err
}

func (a *App) CreateProfile(ctx context.Context, id, name string) (ProfileDTO, error) {
	item, err := a.profiles.Create(ctx, id, name)
	return toProfile(item), err
}

func (a *App) SaveProfile(ctx context.Context, item ProfileDTO) error {
	return a.profiles.Save(ctx, profile.Profile{ID: item.ID, Name: item.Name, RouteID: item.RouteID, ConfigPath: item.ConfigPath, ModelCatalogPath: item.ModelCatalogPath})
}

func (a *App) DeleteProfile(ctx context.Context, id string) error { return a.profiles.Delete(ctx, id) }

func (a *App) RestoreCodexConfig(ctx context.Context, profileID string) error {
	item, err := a.profiles.Get(ctx, profileID)
	if err != nil {
		return err
	}
	return a.codex.Restore(item)
}

func (a *App) AutostartEnabled(context.Context) (bool, error) {
	if a.autostart == nil {
		return false, errAutostartUnavailable
	}
	return a.autostart.IsEnabled()
}

func (a *App) SetAutostart(_ context.Context, enabled bool) error {
	if a.autostart == nil {
		return errAutostartUnavailable
	}
	if enabled {
		return a.autostart.Enable()
	}
	return a.autostart.Disable()
}

func toProvider(item provider.Provider) ProviderDTO {
	return ProviderDTO{ID: item.ID, Name: item.Name, BaseURL: item.BaseURL, Protocol: item.Protocol, AuthRef: item.AuthRef, Headers: item.Headers, QueryParams: item.QueryParams}
}

func toModel(item model.Model) ModelDTO {
	return ModelDTO{ProviderID: item.ProviderID, ID: item.ID, Name: item.Name, Enabled: item.Enabled}
}

func toRoute(item route.Route) RouteDTO {
	return RouteDTO{ID: item.ID, Name: item.Name, PlatformID: item.PlatformID, ProviderID: item.ProviderID, ModelID: item.ModelID, RestartOnActivate: item.RestartOnActivate, Default: item.Default}
}

func toProfile(item profile.Profile) ProfileDTO {
	return ProfileDTO{ID: item.ID, Name: item.Name, RouteID: item.RouteID, ConfigPath: item.ConfigPath, ModelCatalogPath: item.ModelCatalogPath}
}

func mapProviders(items []provider.Provider) []ProviderDTO {
	result := make([]ProviderDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toProvider(item))
	}
	return result
}

func mapModels(items []model.Model) []ModelDTO {
	result := make([]ModelDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toModel(item))
	}
	return result
}

func mapRoutes(items []route.Route) []RouteDTO {
	result := make([]RouteDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toRoute(item))
	}
	return result
}

func mapProfiles(items []profile.Profile) []ProfileDTO {
	result := make([]ProfileDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toProfile(item))
	}
	return result
}
