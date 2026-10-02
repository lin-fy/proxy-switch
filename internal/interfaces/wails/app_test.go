package wails

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	modelapp "codex-provider-hub/internal/application/model"
	"codex-provider-hub/internal/application/ports"
	providerapp "codex-provider-hub/internal/application/provider"
	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/infrastructure/config"
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
