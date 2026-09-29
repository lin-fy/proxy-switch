package codex

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"

	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

func TestPrepareWritesCatalogAndRestoresBackup(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "config.toml")
	original := "model = \"old-model\"\n[other]\nvalue = \"keep\"\n"
	if err := os.WriteFile(configPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("cpa", "CPA", "https://cpa.example/v1", "CPA_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := model.New(p.ID, "model-b", "Model B")
	if err != nil {
		t.Fatal(err)
	}
	other, err := model.New(p.ID, "model-a", "Model A")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "CPA route", p.ID, selected.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("default", "Default")
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Prepare(context.Background(), r, p, selected, []model.Model{selected, other}, pr); err != nil {
		t.Fatal(err)
	}

	var config map[string]any
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		t.Fatal(err)
	}
	if got := config["model_provider"]; got != p.ID {
		t.Fatalf("model_provider = %v", got)
	}
	if got := config["model"]; got != selected.ID {
		t.Fatalf("model = %v", got)
	}
	providers, ok := config["model_providers"].(map[string]any)
	if !ok || providers[p.ID] == nil {
		t.Fatalf("model_providers missing: %#v", config["model_providers"])
	}
	var catalog modelCatalog
	if err := json.Unmarshal([]byte(config["model_catalog_json"].(string)), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Models) != 2 || catalog.Models[0].Slug != selected.ID || catalog.Models[1].Slug != other.ID {
		t.Fatalf("catalog = %#v", catalog.Models)
	}

	if _, err := os.Stat(backupPath(configPath)); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if err := adapter.Restore(pr); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != original {
		t.Fatalf("restored config = %q", restored)
	}
}

func TestFactoryRejectsUnknownPlatform(t *testing.T) {
	factory, err := NewFactory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := factory.Create("other"); err == nil {
		t.Fatal("expected unsupported platform error")
	}
}
