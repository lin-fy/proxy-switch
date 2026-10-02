package codex

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
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
	providerConfig, ok := providers[p.ID].(map[string]any)
	if !ok || providerConfig["env_key"] != "CPA_API_KEY" {
		t.Fatalf("provider env_key = %#v", providerConfig["env_key"])
	}
	catalogPath, ok := config["model_catalog_json"].(string)
	if !ok || catalogPath != filepath.Join(home, "default.models.json") {
		t.Fatalf("model_catalog_json = %#v", config["model_catalog_json"])
	}
	catalogBytes, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatal(err)
	}
	var catalog modelCatalog
	if err := json.Unmarshal(catalogBytes, &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Models) != 2 || catalog.Models[0].Slug != selected.ID || catalog.Models[1].Slug != other.ID {
		t.Fatalf("catalog = %#v", catalog.Models)
	}
	if catalog.Models[0].Priority != 0 || catalog.Models[0].BaseInstructions == "" || catalog.Models[0].TruncationPolicy.Mode == "" {
		t.Fatalf("catalog required Codex fields = %#v", catalog.Models[0])
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
	if _, err := os.Stat(catalogPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("catalog after restoring a previously absent file = %v", err)
	}
}

func TestPrepareUsesConfiguredModelCatalogPath(t *testing.T) {
	home := t.TempDir()
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.New(p.ID, "model-a", "Model A")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "Route", p.ID, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	pr.ModelCatalogPath = filepath.Join("catalogs", "work.json")
	if err := adapter.Prepare(context.Background(), r, p, m, []model.Model{m}, pr); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "catalogs", "work.json")
	if got := adapter.CatalogPath(pr); got != want {
		t.Fatalf("catalog path = %q, want %q", got, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if _, err := toml.DecodeFile(filepath.Join(home, "config.toml"), &config); err != nil {
		t.Fatal(err)
	}
	if got := config["model_catalog_json"]; got != want {
		t.Fatalf("model_catalog_json = %#v, want %q", got, want)
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

func TestExecutableImageNameUsesConfiguredExecutable(t *testing.T) {
	for _, test := range []struct {
		name       string
		executable string
		want       string
	}{
		{name: "default command", executable: "codex", want: "codex.exe"},
		{name: "configured exe", executable: `C:\\Apps\\Codex Desktop\\codex.exe`, want: "codex.exe"},
		{name: "configured command", executable: "custom-codex", want: "custom-codex.exe"},
	} {
		t.Run(test.name, func(t *testing.T) {
			adapter := &Adapter{executable: test.executable}
			if got := adapter.executableImageName(); got != test.want {
				t.Fatalf("image name = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRestartResolvesCredentialBeforeStoppingCodex(t *testing.T) {
	t.Setenv("CODEX_PROVIDER_HUB_MISSING", "")
	adapter := &Adapter{homeDir: t.TempDir(), executable: "codex"}
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "CODEX_PROVIDER_HUB_MISSING")
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Restart(context.Background(), route.Route{PlatformID: route.PlatformCodex}, p); err == nil {
		t.Fatal("expected credential resolution failure before restart")
	}
}

func TestInjectCredentialAddsResolvedSecretOnlyToChildEnvironment(t *testing.T) {
	t.Setenv("CLI_API_KEY", "secret-value")
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "env:CLI_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("codex")
	if err := injectCredential(cmd, p); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range cmd.Env {
		if item == "CLI_API_KEY=secret-value" {
			found = true
		}
	}
	if !found {
		t.Fatalf("resolved credential missing from child environment")
	}
}

func TestPrepareKeepsProfilesInSeparateConfigFiles(t *testing.T) {
	home := t.TempDir()
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("cpa", "CPA", "https://cpa.example/v1", "CPA_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	first, err := model.New(p.ID, "model-a", "Model A")
	if err != nil {
		t.Fatal(err)
	}
	second, err := model.New(p.ID, "model-b", "Model B")
	if err != nil {
		t.Fatal(err)
	}
	firstRoute, err := route.New("first", "First", p.ID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	secondRoute, err := route.New("second", "Second", p.ID, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	firstProfile, err := profile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	secondProfile, err := profile.New("personal", "Personal")
	if err != nil {
		t.Fatal(err)
	}
	firstProfile.ConfigPath = filepath.Join("profiles", "work.toml")
	secondProfile.ConfigPath = filepath.Join("profiles", "personal.toml")

	models := []model.Model{first, second}
	if err := adapter.Prepare(context.Background(), firstRoute, p, first, models, firstProfile); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Prepare(context.Background(), secondRoute, p, second, models, secondProfile); err != nil {
		t.Fatal(err)
	}

	for path, want := range map[string]string{
		filepath.Join(home, "profiles", "work.toml"):     first.ID,
		filepath.Join(home, "profiles", "personal.toml"): second.ID,
		filepath.Join(home, "config.toml"):               second.ID,
	} {
		var config map[string]any
		if _, err := toml.DecodeFile(path, &config); err != nil {
			t.Fatal(err)
		}
		if got := config["model"]; got != want {
			t.Fatalf("%s model = %v, want %s", path, got, want)
		}
	}
}

func TestPrepareRollsBackActiveConfigWhenProfileWriteFails(t *testing.T) {
	home := t.TempDir()
	activePath := filepath.Join(home, "config.toml")
	original := "model = \"original\"\n"
	if err := os.WriteFile(activePath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "blocked"), []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("cpa", "CPA", "https://cpa.example/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.New(p.ID, "model-a", "Model A")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "Route", p.ID, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("broken", "Broken")
	if err != nil {
		t.Fatal(err)
	}
	pr.ConfigPath = filepath.Join("blocked", "profile.toml")

	if err := adapter.Prepare(context.Background(), r, p, m, []model.Model{m}, pr); err == nil {
		t.Fatal("expected profile write failure")
	}
	contents, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != original {
		t.Fatalf("active config was not rolled back: %q", contents)
	}
}

func TestRestoreReturnsNotFoundWithoutBackup(t *testing.T) {
	adapter, err := NewAdapter(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("default", "Default")
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Restore(pr); !errors.Is(err, ErrBackupNotFound) {
		t.Fatalf("restore error = %v, want ErrBackupNotFound", err)
	}
}

func TestRestoreRestoresActiveAndSelectedProfileFiles(t *testing.T) {
	home := t.TempDir()
	profilePath := filepath.Join(home, "profiles", "work.toml")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o700); err != nil {
		t.Fatal(err)
	}
	activePath := filepath.Join(home, "config.toml")
	activeOriginal := "model = \"active-old\"\n"
	profileOriginal := "model = \"profile-old\"\n"
	if err := os.WriteFile(activePath, []byte(activeOriginal), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, []byte(profileOriginal), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath(activePath), []byte(activeOriginal), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath(profilePath), []byte(profileOriginal), 0o600); err != nil {
		t.Fatal(err)
	}
	catalogPath := filepath.Join(home, "work.models.json")
	if err := os.WriteFile(missingBackupPath(catalogPath), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(activePath, []byte("model = \"active-new\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, []byte("model = \"profile-new\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	pr.ConfigPath = filepath.Join("profiles", "work.toml")
	if err := adapter.Restore(pr); err != nil {
		t.Fatal(err)
	}
	activeRestored, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatal(err)
	}
	profileRestored, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(activeRestored) != activeOriginal || string(profileRestored) != profileOriginal {
		t.Fatalf("restored active/profile = %q/%q", activeRestored, profileRestored)
	}
}

func TestRollbackFilesRestoresEarlierFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	originals := map[string][]byte{path: []byte("old")}
	existed := map[string]bool{path: true}
	if err := rollbackFiles(originals, existed, []string{path}); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("rolled back contents = %q", contents)
	}
}
