package model

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"codex-provider-hub/internal/domain/model"
	domainroute "codex-provider-hub/internal/domain/route"
	"codex-provider-hub/internal/infrastructure/config"
)

func TestSyncUpsertsRemoteModelsWithoutDeletingLocalModels(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := config.NewModelRepository(store)
	service := NewService(repo)
	ctx := context.Background()

	if err := repo.Save(ctx, model.Model{ProviderID: "p", ID: "kept", Name: "Old name", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, model.Model{ProviderID: "p", ID: "local-only", Name: "Local", Enabled: true}); err != nil {
		t.Fatal(err)
	}

	items, err := service.Sync(ctx, "p", []model.Model{
		{ID: "kept", Name: "New name", Enabled: true},
		{ID: "remote-new", Name: "Remote new", Enabled: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]model.Model, len(items))
	for _, item := range items {
		got[item.ID] = item
	}
	if len(got) != 3 {
		t.Fatalf("synced models = %#v", got)
	}
	if got["kept"].Enabled || got["kept"].Name != "New name" {
		t.Fatalf("existing model was not updated while preserving enabled state: %#v", got["kept"])
	}
	if !got["remote-new"].Enabled {
		t.Fatalf("new model should default enabled: %#v", got["remote-new"])
	}
}

func TestModelDeleteBlockedByRoute(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	service := NewService(repo, routes)

	item, err := model.New("p", "m", "Model")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, item); err != nil {
		t.Fatal(err)
	}
	r, err := domainroute.New("r", "Route", "p", item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.Save(ctx, r); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, item.ProviderID, item.ID); !errors.Is(err, ErrReferenced) {
		t.Fatalf("model delete with route = %v", err)
	}
	if _, err := repo.Get(ctx, item.ProviderID, item.ID); err != nil {
		t.Fatalf("blocked delete should keep model: %v", err)
	}
}

func TestModelDeleteOnlyChecksRoutesForSameProvider(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	service := NewService(repo, routes)

	item, err := model.New("p", "m", "Model")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, item); err != nil {
		t.Fatal(err)
	}
	// 同名模型 ID 属于另一个 Provider 时不应阻止删除。
	r, err := domainroute.New("r", "Route", "other", item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.Save(ctx, r); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, item.ProviderID, item.ID); err != nil {
		t.Fatalf("model delete with unrelated provider route = %v", err)
	}
}

func TestModelDeleteSucceedsWhenUnreferenced(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := config.NewModelRepository(store)
	service := NewService(repo, config.NewRouteRepository(store))

	item, err := model.New("p", "m", "Model")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, item); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, item.ProviderID, item.ID); err != nil {
		t.Fatalf("unreferenced model delete = %v", err)
	}
	if _, err := repo.Get(ctx, item.ProviderID, item.ID); !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("model after delete = %v", err)
	}
}
