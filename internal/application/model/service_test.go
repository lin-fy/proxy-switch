package model

import (
	"context"
	"path/filepath"
	"testing"

	"codex-provider-hub/internal/domain/model"
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
