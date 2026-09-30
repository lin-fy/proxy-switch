package config

import (
	"context"
	"path/filepath"
	"testing"

	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

func TestProviderRepositoryPersistsAtomically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	repo := NewProviderRepository(store)
	p, err := provider.New("cpa", "CPA", "http://127.0.0.1:8317/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	reloaded := NewProviderRepository(NewStore(path))
	got, err := reloaded.Get(context.Background(), "cpa")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "cpa" || got.BaseURL != p.BaseURL {
		t.Fatalf("unexpected provider: %#v", got)
	}
}

func TestRouteRepositoryKeepsOneDefault(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := NewRouteRepository(store)
	first, err := route.New("first", "First", "cpa", "model-a")
	if err != nil {
		t.Fatal(err)
	}
	first.Default = true
	second, err := route.New("second", "Second", "cpa", "model-b")
	if err != nil {
		t.Fatal(err)
	}
	second.Default = true
	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Default || !items[1].Default {
		t.Fatalf("defaults = %#v", items)
	}
}
