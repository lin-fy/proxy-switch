package wails

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	modelapp "proxy-switch/internal/application/model"
	"proxy-switch/internal/application/ports"
	profileapp "proxy-switch/internal/application/profile"
	providerapp "proxy-switch/internal/application/provider"
	routeapp "proxy-switch/internal/application/route"
	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
	codexadapter "proxy-switch/internal/infrastructure/codex"
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
	importer  ports.WorkspaceImporter
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
	importers ...ports.WorkspaceImporter,
) *App {
	var importer ports.WorkspaceImporter
	if len(importers) > 0 {
		importer = importers[0]
	}
	return &App{providers: providers, models: models, routes: routes, profiles: profiles, activator: activator, codex: codex, autostart: autostart, tester: tester, importer: importer}
}

type ProviderDTO struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	BaseURL            string            `json:"base_url"`
	Protocol           string            `json:"protocol"`
	AuthRef            string            `json:"auth_ref,omitempty"`
	AuthMode           string            `json:"auth_mode"`
	PresetID           string            `json:"preset_id,omitempty"`
	PresetName         string            `json:"preset_name,omitempty"`
	RequiresCredential bool              `json:"requires_credential"`
	Headers            map[string]string `json:"headers,omitempty"`
	QueryParams        map[string]string `json:"query_params,omitempty"`
}

type ProviderPresetDTO struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	BaseURL            string `json:"base_url"`
	AuthMode           string `json:"auth_mode"`
	RequiresCredential bool   `json:"requires_credential"`
}

type ProviderHealthDTO struct {
	ProviderID string `json:"provider_id"`
	Status     string `json:"status"`
	CheckedAt  string `json:"checked_at,omitempty"`
	LatencyMS  int64  `json:"latency_ms,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
}

type workspaceExport struct {
	SchemaVersion int              `json:"schema_version"`
	ExportedAt    string           `json:"exported_at"`
	Providers     []providerExport `json:"providers"`
	Models        []ModelDTO       `json:"models"`
	Routes        []RouteDTO       `json:"routes"`
	Profiles      []ProfileDTO     `json:"profiles"`
}

type providerExport struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	BaseURL            string `json:"base_url"`
	Protocol           string `json:"protocol"`
	AuthRef            string `json:"auth_ref,omitempty"`
	AuthMode           string `json:"auth_mode"`
	PresetID           string `json:"preset_id,omitempty"`
	PresetName         string `json:"preset_name,omitempty"`
	RequiresCredential bool   `json:"requires_credential"`
}

type WorkspaceImportPreviewDTO struct {
	Valid             bool     `json:"valid"`
	SchemaVersion     int      `json:"schema_version"`
	ProviderCount     int      `json:"provider_count"`
	ModelCount        int      `json:"model_count"`
	RouteCount        int      `json:"route_count"`
	ProfileCount      int      `json:"profile_count"`
	ProviderConflicts []string `json:"provider_conflicts,omitempty"`
	ModelConflicts    []string `json:"model_conflicts,omitempty"`
	RouteConflicts    []string `json:"route_conflicts,omitempty"`
	ProfileConflicts  []string `json:"profile_conflicts,omitempty"`
	MissingReferences []string `json:"missing_references,omitempty"`
	Errors            []string `json:"errors,omitempty"`
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
	Priority          int    `json:"priority"`
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

type CodexConfigStatusDTO struct {
	ConfigPath        string   `json:"config_path"`
	ConfigExists      bool     `json:"config_exists"`
	ConfigValid       bool     `json:"config_valid"`
	AuthPath          string   `json:"auth_path"`
	AuthExists        bool     `json:"auth_exists"`
	AuthValid         bool     `json:"auth_valid"`
	AuthMode          string   `json:"auth_mode"`
	CredentialPresent bool     `json:"credential_present"`
	ModelProvider     string   `json:"model_provider,omitempty"`
	Model             string   `json:"model,omitempty"`
	ProviderIDs       []string `json:"provider_ids,omitempty"`
	Importable        bool     `json:"importable"`
	ImportBlocker     string   `json:"import_blocker,omitempty"`
}

func (a *App) ListProviders(ctx context.Context) ([]ProviderDTO, error) {
	items, err := a.providers.List(ctx)
	return mapProviders(items), err
}

// ListProviderPresets returns read-only templates. They contain no credential
// values and selecting one never writes state until CreateProvider is called.
func (a *App) ListProviderPresets(context.Context) []ProviderPresetDTO {
	return mapProviderPresets(provider.Presets())
}

// ExportWorkspace returns a credential-safe JSON snapshot. Header/query
// values are deliberately omitted because they may contain secrets; auth_ref
// is retained only when it is a valid environment/Credential Manager reference.
func (a *App) ExportWorkspace(ctx context.Context) (string, error) {
	providers, err := a.providers.List(ctx)
	if err != nil {
		return "", err
	}
	models, err := a.models.List(ctx)
	if err != nil {
		return "", err
	}
	routes, err := a.routes.List(ctx)
	if err != nil {
		return "", err
	}
	profiles, err := a.profiles.List(ctx)
	if err != nil {
		return "", err
	}
	export := workspaceExport{
		SchemaVersion: 1,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Providers:     exportProviders(providers),
		Models:        mapModels(models),
		Routes:        mapRoutes(routes),
		Profiles:      mapProfiles(profiles),
	}
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode workspace export: %w", err)
	}
	return string(data), nil
}

// PreviewWorkspaceImport validates an exported snapshot without mutating any
// state. Unknown fields are rejected so headers, query parameters, auth.json
// content, and future secret-bearing fields cannot sneak into an import flow.
func (a *App) PreviewWorkspaceImport(ctx context.Context, payload string) (WorkspaceImportPreviewDTO, error) {
	preview := WorkspaceImportPreviewDTO{}
	if err := ctx.Err(); err != nil {
		return preview, err
	}
	decoder := json.NewDecoder(io.LimitReader(strings.NewReader(payload), 2<<20))
	decoder.DisallowUnknownFields()
	var incoming workspaceExport
	if err := decoder.Decode(&incoming); err != nil {
		preview.Errors = []string{"导入 JSON 无效或包含不允许字段"}
		return preview, nil
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		preview.Errors = []string{"导入 JSON 必须只包含一个对象"}
		return preview, nil
	}
	preview.SchemaVersion = incoming.SchemaVersion
	preview.ProviderCount = len(incoming.Providers)
	preview.ModelCount = len(incoming.Models)
	preview.RouteCount = len(incoming.Routes)
	preview.ProfileCount = len(incoming.Profiles)
	if incoming.SchemaVersion != 1 {
		preview.Errors = append(preview.Errors, "不支持的导出 schema_version")
	}
	providerIDs := make(map[string]bool, len(incoming.Providers))
	modelIDs := make(map[string]bool, len(incoming.Models))
	routeIDs := make(map[string]bool, len(incoming.Routes))
	for _, item := range incoming.Providers {
		providerIDs[item.ID] = true
		if item.ID == "" || item.Name == "" || item.BaseURL == "" {
			preview.Errors = append(preview.Errors, "Provider 缺少必要字段")
		}
		if item.AuthRef != "" {
			if err := provider.ValidateAuthRef(item.AuthRef); err != nil {
				preview.Errors = append(preview.Errors, "Provider 包含不安全的认证引用")
			}
		}
	}
	for _, item := range incoming.Models {
		key := item.ProviderID + "\x00" + item.ID
		modelIDs[key] = true
		if item.ProviderID == "" || item.ID == "" {
			preview.Errors = append(preview.Errors, "Model 缺少必要字段")
		}
		if !providerIDs[item.ProviderID] {
			preview.MissingReferences = append(preview.MissingReferences, "Model 引用了不存在的 Provider")
		}
	}
	for _, item := range incoming.Routes {
		routeIDs[item.ID] = true
		if item.ID == "" || item.ProviderID == "" || item.ModelID == "" {
			preview.Errors = append(preview.Errors, "Route 缺少必要字段")
		}
		if !providerIDs[item.ProviderID] {
			preview.MissingReferences = append(preview.MissingReferences, "Route 引用了不存在的 Provider")
		}
		if !modelIDs[item.ProviderID+"\x00"+item.ModelID] {
			preview.MissingReferences = append(preview.MissingReferences, "Route 引用了不存在的 Model")
		}
	}
	for _, item := range incoming.Profiles {
		if item.ID == "" || item.Name == "" {
			preview.Errors = append(preview.Errors, "Profile 缺少必要字段")
		}
		if item.RouteID != "" && !routeIDs[item.RouteID] {
			preview.MissingReferences = append(preview.MissingReferences, "Profile 引用了不存在的 Route")
		}
	}
	currentProviders, err := a.providers.List(ctx)
	if err != nil {
		return preview, err
	}
	for _, item := range currentProviders {
		if providerIDs[item.ID] {
			preview.ProviderConflicts = append(preview.ProviderConflicts, item.ID)
		}
	}
	currentModels, err := a.models.List(ctx)
	if err != nil {
		return preview, err
	}
	for _, item := range currentModels {
		if modelIDs[item.ProviderID+"\x00"+item.ID] {
			preview.ModelConflicts = append(preview.ModelConflicts, item.ProviderID+"/"+item.ID)
		}
	}
	currentRoutes, err := a.routes.List(ctx)
	if err != nil {
		return preview, err
	}
	for _, item := range currentRoutes {
		if routeIDs[item.ID] {
			preview.RouteConflicts = append(preview.RouteConflicts, item.ID)
		}
	}
	currentProfiles, err := a.profiles.List(ctx)
	if err != nil {
		return preview, err
	}
	for _, item := range incoming.Profiles {
		for _, current := range currentProfiles {
			if current.ID == item.ID {
				preview.ProfileConflicts = append(preview.ProfileConflicts, item.ID)
				break
			}
		}
	}
	preview.Valid = len(preview.Errors) == 0 && len(preview.MissingReferences) == 0
	return preview, nil
}

// ImportWorkspace applies a previously previewed snapshot only when it is
// valid and conflict-free. The importer appends resources atomically and never
// overwrites existing IDs or touches auth.json.
func (a *App) ImportWorkspace(ctx context.Context, payload string) (WorkspaceImportPreviewDTO, error) {
	preview, err := a.PreviewWorkspaceImport(ctx, payload)
	conflicts := len(preview.ProviderConflicts) + len(preview.ModelConflicts) + len(preview.RouteConflicts) + len(preview.ProfileConflicts)
	if conflicts > 0 {
		preview.Valid = false
		preview.Errors = append(preview.Errors, "存在 ID 冲突，导入按安全默认被拒绝")
	}
	if err != nil || !preview.Valid {
		return preview, err
	}
	if a.importer == nil {
		preview.Valid = false
		preview.Errors = append(preview.Errors, "导入服务尚未就绪")
		return preview, nil
	}
	decoder := json.NewDecoder(io.LimitReader(strings.NewReader(payload), 2<<20))
	decoder.DisallowUnknownFields()
	var incoming workspaceExport
	if err := decoder.Decode(&incoming); err != nil {
		preview.Valid = false
		preview.Errors = append(preview.Errors, "导入 JSON 无效或包含不允许字段")
		return preview, nil
	}
	providers := make([]provider.Provider, 0, len(incoming.Providers))
	for _, item := range incoming.Providers {
		if item.Protocol != "" && item.Protocol != provider.ProtocolResponses {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Provider 协议不受支持")
			return preview, nil
		}
		created, err := provider.New(item.ID, item.Name, item.BaseURL, item.AuthRef)
		if err != nil {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Provider 字段无效")
			return preview, nil
		}
		providers = append(providers, created)
	}
	models := make([]model.Model, 0, len(incoming.Models))
	for _, item := range incoming.Models {
		created, err := model.New(item.ProviderID, item.ID, item.Name)
		if err != nil {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Model 字段无效")
			return preview, nil
		}
		created.Enabled = item.Enabled
		models = append(models, created)
	}
	routes := make([]route.Route, 0, len(incoming.Routes))
	for _, item := range incoming.Routes {
		created, err := route.New(item.ID, item.Name, item.ProviderID, item.ModelID)
		if err != nil {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Route 字段无效")
			return preview, nil
		}
		created.Priority, created.RestartOnActivate, created.Default = item.Priority, item.RestartOnActivate, item.Default
		if err := created.Validate(); err != nil {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Route 字段无效")
			return preview, nil
		}
		routes = append(routes, created)
	}
	profiles := make([]profile.Profile, 0, len(incoming.Profiles))
	for _, item := range incoming.Profiles {
		created, err := profile.New(item.ID, item.Name)
		if err != nil {
			preview.Valid = false
			preview.Errors = append(preview.Errors, "Profile 字段无效")
			return preview, nil
		}
		created.RouteID, created.ConfigPath, created.ModelCatalogPath = item.RouteID, item.ConfigPath, item.ModelCatalogPath
		profiles = append(profiles, created)
	}
	if err := a.importer.ImportWorkspace(ctx, providers, models, routes, profiles); err != nil {
		preview.Valid = false
		preview.Errors = append(preview.Errors, "导入失败，未写入任何资源")
		return preview, nil
	}
	return preview, nil
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

// CheckProviderHealth performs an explicit, read-only connectivity check. It
// is never called during list/reload and it returns only a safe error code.
func (a *App) CheckProviderHealth(ctx context.Context, id string) (ProviderHealthDTO, error) {
	status := ProviderHealthDTO{ProviderID: id, Status: "unknown", CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	if a.tester == nil {
		status.Status = "unavailable"
		return status, nil
	}
	item, err := a.providers.Get(ctx, id)
	if err != nil {
		return ProviderHealthDTO{}, err
	}
	started := time.Now()
	err = a.tester.Test(ctx, item)
	status.LatencyMS = time.Since(started).Milliseconds()
	if err == nil {
		status.Status = "healthy"
		return status, nil
	}
	status.Status = "unhealthy"
	status.ErrorCode = providerHealthErrorCode(err)
	return status, nil
}

func providerHealthErrorCode(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var timeout net.Error
	if errors.As(err, &timeout) && timeout.Timeout() {
		return "timeout"
	}
	return "request_error"
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

func (a *App) SyncProviderModels(ctx context.Context, providerID string) ([]ModelDTO, error) {
	if a.tester == nil {
		return nil, errors.New("Provider 模型同步服务尚未就绪")
	}
	item, err := a.providers.Get(ctx, providerID)
	if err != nil {
		return nil, err
	}
	remote, err := a.tester.ListModels(ctx, item)
	if err != nil {
		return nil, err
	}
	models := make([]model.Model, 0, len(remote))
	for _, item := range remote {
		models = append(models, model.Model{ProviderID: providerID, ID: item.ID, Name: item.Name, Enabled: true})
	}
	items, err := a.models.Sync(ctx, providerID, models)
	return mapModels(items), err
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
	return a.routes.Save(ctx, route.Route{ID: item.ID, Name: item.Name, PlatformID: item.PlatformID, ProviderID: item.ProviderID, ModelID: item.ModelID, Priority: item.Priority, RestartOnActivate: item.RestartOnActivate, Default: item.Default})
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

// ImportCodexConfig explicitly snapshots the current config.toml into a new
// profile. auth.json is never copied or modified by this operation.
func (a *App) ImportCodexConfig(ctx context.Context, id, name string) (ProfileDTO, error) {
	if a.codex == nil {
		return ProfileDTO{}, errors.New("Codex 配置服务尚未就绪")
	}
	item, err := profile.New(id, name)
	if err != nil {
		return ProfileDTO{}, err
	}
	items, err := a.profiles.List(ctx)
	if err != nil {
		return ProfileDTO{}, err
	}
	for _, existing := range items {
		if existing.ID == item.ID {
			return ProfileDTO{}, fmt.Errorf("配置档案 %q 已存在", item.ID)
		}
	}
	item.ConfigPath = filepath.Join("profiles", item.ID, "config.toml")
	if err := a.profiles.Save(ctx, item); err != nil {
		return ProfileDTO{}, err
	}
	if err := a.codex.ImportConfig(item); err != nil {
		if rollbackErr := a.profiles.Delete(ctx, item.ID); rollbackErr != nil {
			return ProfileDTO{}, errors.Join(err, fmt.Errorf("删除未完成的配置档案: %w", rollbackErr))
		}
		return ProfileDTO{}, err
	}
	return toProfile(item), nil
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

// InspectCodexConfig is read-only. Importing a detected config is deliberately
// a separate user action and this method never writes config.toml or auth.json.
func (a *App) InspectCodexConfig(context.Context) (CodexConfigStatusDTO, error) {
	status, err := a.codex.InspectConfig()
	if err != nil {
		return CodexConfigStatusDTO{}, err
	}
	return CodexConfigStatusDTO{
		ConfigPath: status.ConfigPath, ConfigExists: status.ConfigExists, ConfigValid: status.ConfigValid,
		AuthPath: status.AuthPath, AuthExists: status.AuthExists, AuthValid: status.AuthValid,
		AuthMode: status.AuthMode, CredentialPresent: status.CredentialPresent,
		ModelProvider: status.ModelProvider, Model: status.Model, ProviderIDs: status.ProviderIDs,
		Importable: status.Importable, ImportBlocker: status.ImportBlocker,
	}, nil
}

func (a *App) CodexRunning(ctx context.Context) (bool, error) {
	return a.codex.IsRunning(ctx)
}

func (a *App) StartCodex(ctx context.Context) error {
	items, err := a.routes.List(ctx)
	if err != nil {
		return err
	}
	var item route.Route
	for _, candidate := range items {
		if candidate.Default {
			item = candidate
			break
		}
	}
	if item.ID == "" {
		return errors.New("没有默认路由")
	}
	providerItem, err := a.providers.Get(ctx, item.ProviderID)
	if err != nil {
		return err
	}
	running, err := a.codex.IsRunning(ctx)
	if err != nil {
		return err
	}
	if running {
		return a.codex.Restart(ctx, item, providerItem)
	}
	return a.codex.Launch(ctx, item, providerItem)
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
	dto := ProviderDTO{
		ID: item.ID, Name: item.Name, BaseURL: item.BaseURL, Protocol: item.Protocol,
		AuthRef: item.AuthRef, AuthMode: string(item.AuthMode()), Headers: item.Headers, QueryParams: item.QueryParams,
	}
	if preset, ok := provider.MatchPreset(item); ok {
		dto.PresetID, dto.PresetName, dto.RequiresCredential = preset.ID, preset.Name, preset.RequiresCredential
	}
	return dto
}

func toModel(item model.Model) ModelDTO {
	return ModelDTO{ProviderID: item.ProviderID, ID: item.ID, Name: item.Name, Enabled: item.Enabled}
}

func toRoute(item route.Route) RouteDTO {
	return RouteDTO{ID: item.ID, Name: item.Name, PlatformID: item.PlatformID, ProviderID: item.ProviderID, ModelID: item.ModelID, Priority: item.Priority, RestartOnActivate: item.RestartOnActivate, Default: item.Default}
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

func exportProviders(items []provider.Provider) []providerExport {
	result := make([]providerExport, 0, len(items))
	for _, item := range items {
		dto := toProvider(item)
		exported := providerExport{
			ID: dto.ID, Name: dto.Name, BaseURL: dto.BaseURL, Protocol: dto.Protocol,
			AuthMode: dto.AuthMode, PresetID: dto.PresetID, PresetName: dto.PresetName,
			RequiresCredential: dto.RequiresCredential,
		}
		if provider.ValidateAuthRef(item.AuthRef) == nil {
			exported.AuthRef = item.AuthRef
		}
		result = append(result, exported)
	}
	return result
}

func mapProviderPresets(items []provider.PresetMetadata) []ProviderPresetDTO {
	result := make([]ProviderPresetDTO, 0, len(items))
	for _, item := range items {
		result = append(result, ProviderPresetDTO{ID: item.ID, Name: item.Name, BaseURL: item.BaseURL, AuthMode: string(item.AuthMode), RequiresCredential: item.RequiresCredential})
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
