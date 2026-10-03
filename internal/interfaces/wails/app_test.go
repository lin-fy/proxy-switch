package wails

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	modelapp "proxy-switch/internal/application/model"
	"proxy-switch/internal/application/ports"
	profileapp "proxy-switch/internal/application/profile"
	providerapp "proxy-switch/internal/application/provider"
	routeapp "proxy-switch/internal/application/route"
	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/infrastructure/codex"
	"proxy-switch/internal/infrastructure/config"
)

type fakeAutostart struct {
	enabled bool
	err     error
}

func (f *fakeAutostart) Enable() error {
	if f.err == nil {
		f.enabled = true
	}
	return f.err
}

func (f *fakeAutostart) Disable() error {
	if f.err == nil {
		f.enabled = false
	}
	return f.err
}

func (f *fakeAutostart) IsEnabled() (bool, error) { return f.enabled, f.err }

func TestAutostartFacade(t *testing.T) {
	app := &App{}
	manager := &fakeAutostart{}
	app.autostart = manager

	if err := app.SetAutostart(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	enabled, err := app.AutostartEnabled(context.Background())
	if err != nil || !enabled {
		t.Fatalf("enabled = %v, err = %v", enabled, err)
	}

	if err := app.SetAutostart(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	enabled, err = app.AutostartEnabled(context.Background())
	if err != nil || enabled {
		t.Fatalf("enabled = %v, err = %v", enabled, err)
	}
}

func TestAutostartFacadeRequiresManager(t *testing.T) {
	app := &App{}
	if _, err := app.AutostartEnabled(context.Background()); !errors.Is(err, errAutostartUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestProviderDTOExposesPresetAndAuthModeMetadataOnly(t *testing.T) {
	item, err := provider.New("openai", "OpenAI", "https://api.openai.com/v1", "credential:openai")
	if err != nil {
		t.Fatal(err)
	}
	dto := toProvider(item)
	if dto.PresetID != "openai" || dto.PresetName != "OpenAI API" || !dto.RequiresCredential {
		t.Fatalf("preset metadata = %#v", dto)
	}
	if dto.AuthMode != "credential_ref" || dto.AuthRef != "credential:openai" {
		t.Fatalf("auth metadata = %#v", dto)
	}
	if dto.Headers != nil || dto.QueryParams != nil {
		t.Fatalf("unexpected writable metadata in DTO: %#v", dto)
	}
}

func TestListProviderPresetsExposesMetadataWithoutCredentials(t *testing.T) {
	app := &App{}
	presets := app.ListProviderPresets(context.Background())
	if len(presets) < 2 {
		t.Fatalf("presets = %#v", presets)
	}
	for _, item := range presets {
		if item.ID == "" || item.Name == "" || item.AuthMode == "" {
			t.Fatalf("incomplete preset metadata: %#v", item)
		}
		if strings.Contains(item.BaseURL, "sk-") || strings.Contains(item.Name, "API_KEY") {
			t.Fatalf("preset leaked credential data: %#v", item)
		}
	}
}

func TestCreateProviderRejectsRawCredentialThroughWails(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	app := NewApp(providerapp.NewService(providers), nil, nil, nil, nil, nil, nil, nil)
	if _, err := app.CreateProvider(context.Background(), "openai", "OpenAI", "https://api.openai.com/v1", "sk-live-secret"); !errors.Is(err, provider.ErrInvalidAuthRef) {
		t.Fatalf("raw credential error = %v, want %v", err, provider.ErrInvalidAuthRef)
	}
	if items, err := providers.List(context.Background()); err != nil || len(items) != 0 {
		t.Fatalf("raw credential persisted provider: items=%#v err=%v", items, err)
	}
}

type healthTester struct{ err error }

func (h healthTester) Test(context.Context, provider.Provider) error { return h.err }
func (h healthTester) TestModel(context.Context, provider.Provider, string) error {
	return h.err
}
func (h healthTester) ListModels(context.Context, provider.Provider) ([]ports.RemoteModel, error) {
	return nil, h.err
}

func TestCheckProviderHealthReturnsSafeStatusAndErrorCode(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	p, err := provider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	app := NewApp(providerapp.NewService(providers), nil, nil, nil, nil, nil, nil, healthTester{err: context.DeadlineExceeded})
	status, err := app.CheckProviderHealth(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "unhealthy" || status.ErrorCode != "timeout" || status.ProviderID != p.ID {
		t.Fatalf("health status = %#v", status)
	}
	if status.LatencyMS < 0 || status.CheckedAt == "" {
		t.Fatalf("health timing = %#v", status)
	}
}

func TestExportWorkspaceOmitsSecretBearingProviderFields(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	profiles := config.NewProfileRepository(store)
	if err := providers.Save(context.Background(), provider.Provider{
		ID: "safe", Name: "Safe", BaseURL: "https://example.test/v1", Protocol: provider.ProtocolResponses,
		AuthRef: "credential:target", Headers: map[string]string{"Authorization": "Bearer secret-header"},
		QueryParams: map[string]string{"api_key": "secret-query"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(context.Background(), provider.Provider{
		ID: "legacy", Name: "Legacy", BaseURL: "https://legacy.test/v1", Protocol: provider.ProtocolResponses,
		AuthRef: "sk-live-secret",
	}); err != nil {
		t.Fatal(err)
	}
	app := NewApp(
		providerapp.NewService(providers), modelapp.NewService(models, routes),
		routeapp.NewService(routes, providers, models), profileapp.NewService(profiles), nil, nil, nil, nil,
	)
	payload, err := app.ExportWorkspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(payload, "secret-header") || strings.Contains(payload, "secret-query") || strings.Contains(payload, "sk-live-secret") {
		t.Fatalf("export leaked credential data: %s", payload)
	}
	if !strings.Contains(payload, `"schema_version": 1`) || !strings.Contains(payload, "credential:target") {
		t.Fatalf("export missing safe metadata: %s", payload)
	}
	preview, err := app.PreviewWorkspaceImport(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Valid || preview.ProviderCount != 2 {
		t.Fatalf("export preview = %#v", preview)
	}
}

func TestPreviewWorkspaceImportRejectsSecretFieldsWithoutEchoingPayload(t *testing.T) {
	app := &App{}
	payload := `{"schema_version":1,"providers":[{"id":"p","name":"P","base_url":"https://example.test/v1","protocol":"responses","headers":{"Authorization":"Bearer secret-token"}}]}`
	preview, err := app.PreviewWorkspaceImport(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Valid || len(preview.Errors) == 0 {
		t.Fatalf("secret field accepted: %#v", preview)
	}
	if strings.Contains(strings.Join(preview.Errors, " "), "secret-token") {
		t.Fatalf("preview echoed secret: %#v", preview)
	}
}

func TestImportWorkspaceAppendsOnlyConflictFreeResources(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	profiles := config.NewProfileRepository(store)
	app := NewApp(
		providerapp.NewService(providers), modelapp.NewService(models, routes, providers),
		routeapp.NewService(routes, providers, models, profiles), profileapp.NewService(profiles, routes), nil, nil, nil, nil, store,
	)
	payload := `{"schema_version":1,"exported_at":"2026-10-02T00:00:00Z","providers":[{"id":"p","name":"Provider","base_url":"https://example.test/v1","protocol":"responses","auth_ref":"credential:target","auth_mode":"credential_ref","requires_credential":false}],"models":[{"provider_id":"p","id":"m","name":"Model","enabled":true}],"routes":[{"id":"r","name":"Route","platform_id":"codex","provider_id":"p","model_id":"m","priority":0,"restart_on_activate":false,"default":true}],"profiles":[{"id":"work","name":"Work","route_id":"r"}]}`
	preview, err := app.ImportWorkspace(context.Background(), payload)
	if err != nil || !preview.Valid {
		t.Fatalf("first import = %#v, %v", preview, err)
	}
	items, err := providers.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("providers after import = %#v, %v", items, err)
	}
	conflict, err := app.ImportWorkspace(context.Background(), payload)
	if err != nil || conflict.Valid || len(conflict.ProviderConflicts) != 1 {
		t.Fatalf("conflicting import = %#v, %v", conflict, err)
	}
	items, err = providers.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("conflicting import changed state = %#v, %v", items, err)
	}
}

type syncTester struct{}

func (syncTester) Test(context.Context, provider.Provider) error              { return nil }
func (syncTester) TestModel(context.Context, provider.Provider, string) error { return nil }
func (syncTester) ListModels(context.Context, provider.Provider) ([]ports.RemoteModel, error) {
	return []ports.RemoteModel{{ID: "remote-a", Name: "Remote A"}}, nil
}

func TestSyncProviderModelsUpsertsThroughWailsFacade(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	p, err := provider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if err := models.Save(context.Background(), model.Model{ProviderID: p.ID, ID: "local", Name: "Local", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	app := NewApp(
		providerapp.NewService(providers),
		modelapp.NewService(models),
		nil, nil, nil, nil, nil, syncTester{},
	)
	items, err := app.SyncProviderModels(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("synced DTOs = %#v", items)
	}
	for _, item := range items {
		if item.ID == "remote-a" && !item.Enabled {
			t.Fatal("new remote model should be enabled")
		}
	}
}

func TestImportCodexConfigCreatesProfileWithoutCopyingAuth(t *testing.T) {
	home := t.TempDir()
	configText := "# preserve me\nmodel = \"official\"\n[mcp_servers.docs]\ncommand = \"docs\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(configText), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte(`{"token":"secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	profiles := config.NewProfileRepository(store)
	adapter, err := codex.NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	app := NewApp(nil, nil, nil, profileapp.NewService(profiles), nil, adapter, nil, nil)
	item, err := app.ImportCodexConfig(context.Background(), "imported", "Imported Codex")
	if err != nil {
		t.Fatal(err)
	}
	if item.ConfigPath != filepath.Join("profiles", "imported", "config.toml") {
		t.Fatalf("config path = %q", item.ConfigPath)
	}
	imported, err := os.ReadFile(filepath.Join(home, item.ConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(imported) != configText {
		t.Fatalf("imported config = %q", imported)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.Dir(item.ConfigPath), "auth.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("auth was copied under profile: %v", err)
	}
	auth, err := os.ReadFile(filepath.Join(home, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(auth) != `{"token":"secret"}` {
		t.Fatalf("auth changed = %q", auth)
	}
}

func TestImportCodexConfigRejectsDuplicateAndRollsBackProfile(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	profiles := config.NewProfileRepository(store)
	adapter, err := codex.NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	app := NewApp(nil, nil, nil, profileapp.NewService(profiles), nil, adapter, nil, nil)
	if _, err := app.ImportCodexConfig(context.Background(), "broken", "Broken"); err == nil {
		t.Fatal("expected malformed config rejection")
	}
	if _, err := profiles.Get(context.Background(), "broken"); !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("failed import left profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = \"ok\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ImportCodexConfig(context.Background(), "same", "Same"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ImportCodexConfig(context.Background(), "same", "Again"); err == nil {
		t.Fatal("expected duplicate profile rejection")
	}
}
