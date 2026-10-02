package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
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

func TestRouteRepositoryPromotesAnotherRouteWhenDefaultIsDeleted(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "state.json"))
	repo := NewRouteRepository(store)
	first, _ := route.New("first", "First", "cpa", "model-a")
	first.Default = true
	second, _ := route.New("second", "Second", "cpa", "model-b")
	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(context.Background(), first.ID); err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || !items[0].Default || items[0].ID != second.ID {
		t.Fatalf("routes after default deletion = %#v", items)
	}
}

func TestStoreConcurrentSavesDoNotLoseUpdates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	repo := NewProviderRepository(store)
	ctx := context.Background()

	const count = 32
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item, err := provider.New(fmt.Sprintf("p-%02d", i), "Provider", "https://example.test/v1", "")
			if err != nil {
				errs <- err
				return
			}
			if err := repo.Save(ctx, item); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent save: %v", err)
	}

	// 用新 Store 从磁盘重读，确保不是内存假象。
	items, err := NewProviderRepository(NewStore(path)).List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != count {
		t.Fatalf("persisted providers = %d, want %d", len(items), count)
	}
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		seen[item.ID] = true
	}
	for i := 0; i < count; i++ {
		if id := fmt.Sprintf("p-%02d", i); !seen[id] {
			t.Fatalf("missing concurrently saved provider %q", id)
		}
	}
}

func TestStoreConcurrentSaveAndListKeepDataConsistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	providers := NewProviderRepository(store)
	models := NewModelRepository(store)
	ctx := context.Background()

	if err := providers.Save(ctx, provider.Provider{ID: "p", Name: "Provider", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := models.Save(ctx, model.Model{ProviderID: "p", ID: fmt.Sprintf("m-%d", i), Name: "Model", Enabled: true}); err != nil {
				errs <- err
			}
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := providers.List(ctx); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent access: %v", err)
	}

	items, err := models.ListByProvider(ctx, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 8 {
		t.Fatalf("models after concurrent access = %d, want 8", len(items))
	}
}

func TestStoreCorruptStateReturnsErrorAndPreservesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	corrupt := []byte("{ this is not json")
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	store := NewStore(path)
	repo := NewProviderRepository(store)
	ctx := context.Background()

	if _, err := repo.List(ctx); err == nil || !strings.Contains(err.Error(), "decode state") {
		t.Fatalf("corrupt read error = %v", err)
	}
	if err := repo.Save(ctx, provider.Provider{ID: "p", Name: "Provider", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}); err == nil || !strings.Contains(err.Error(), "decode state") {
		t.Fatalf("corrupt write error = %v", err)
	}

	// 损坏文件不得被静默覆盖，否则用户无法从磁盘恢复原始内容。
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(corrupt) {
		t.Fatalf("corrupt file was modified: %q", got)
	}
}

func TestStoreMissingStateIsEmptyAndWritable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	repo := NewProviderRepository(NewStore(path))
	ctx := context.Background()

	items, err := repo.List(ctx)
	if err != nil || len(items) != 0 {
		t.Fatalf("missing state list = %#v, %v", items, err)
	}
	if err := repo.Save(ctx, provider.Provider{ID: "p", Name: "Provider", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}); err != nil {
		t.Fatalf("save into missing directory = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state file not created: %v", err)
	}
}
