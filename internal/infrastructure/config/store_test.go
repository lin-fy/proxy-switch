package config

import (
	"context"
	"path/filepath"
	"testing"

	"codex-provider-hub/internal/domain/provider"
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
