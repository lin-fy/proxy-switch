package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
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

func TestInspectConfigReportsFilesWithoutExposingAuth(t *testing.T) {
	home := t.TempDir()
	config := "model_provider = \"custom\"\nmodel = \"model-a\"\n[model_providers.custom]\nname = \"Custom\"\nbase_url = \"https://example.test/v1\"\nwire_api = \"responses\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte(`{"OPENAI_API_KEY":"secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	status, err := adapter.InspectConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !status.ConfigExists || !status.ConfigValid || !status.AuthExists || !status.AuthValid || !status.Importable || status.AuthMode != "api_key" || !status.CredentialPresent {
		t.Fatalf("status = %#v", status)
	}
	if status.ModelProvider != "custom" || status.Model != "model-a" || len(status.ProviderIDs) != 1 || status.ProviderIDs[0] != "custom" {
		t.Fatalf("config fields = %#v", status)
	}
	if strings.Contains(fmt.Sprintf("%#v", status), "secret") {
		t.Fatalf("status leaked auth secret: %#v", status)
	}
}

func TestInspectConfigFailsClosedForMalformedConfigAndStillReportsAuth(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte(`{"token":"present"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	status, err := adapter.InspectConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !status.ConfigExists || status.ConfigValid || status.Importable || !status.AuthExists || !status.AuthValid || status.AuthMode != "oauth" || !status.CredentialPresent {
		t.Fatalf("status = %#v", status)
	}
}

func TestInspectConfigDoesNotMarkUnknownCurrentProviderImportable(t *testing.T) {
	home := t.TempDir()
	config := "model_provider = \"missing\"\n[model_providers.custom]\nname = \"Custom\"\nbase_url = \"https://example.test/v1\"\nwire_api = \"responses\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	status, err := adapter.InspectConfig()
	if err != nil {
		t.Fatal(err)
	}
	if status.Importable || status.ImportBlocker == "" {
		t.Fatalf("unknown provider should block import: %#v", status)
	}
}

func TestInspectConfigReportsMissingAuthWithoutReadingSecrets(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = \"official\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	status, err := adapter.InspectConfig()
	if err != nil {
		t.Fatal(err)
	}
	if status.AuthMode != "missing" || status.CredentialPresent || status.AuthExists {
		t.Fatalf("missing auth status = %#v", status)
	}
}

func TestImportConfigCopiesOnlyConfigAndPreservesContents(t *testing.T) {
	home := t.TempDir()
	original := "# keep comments\nmodel = \"official\"\n[mcp_servers.docs]\ncommand = \"docs\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte(`{"access_token":"secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("imported", "Imported")
	if err != nil {
		t.Fatal(err)
	}
	pr.ConfigPath = filepath.Join("profiles", pr.ID, "config.toml")
	if err := adapter.ImportConfig(pr); err != nil {
		t.Fatal(err)
	}
	imported, err := os.ReadFile(adapter.ConfigPath(pr))
	if err != nil {
		t.Fatal(err)
	}
	if string(imported) != original {
		t.Fatalf("imported config changed: %q", imported)
	}
	auth, err := os.ReadFile(filepath.Join(home, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(auth) != `{"access_token":"secret"}` {
		t.Fatalf("auth changed: %q", auth)
	}
}

func TestImportConfigRejectsMalformedOrExistingTarget(t *testing.T) {
	home := t.TempDir()
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("imported", "Imported")
	if err != nil {
		t.Fatal(err)
	}
	pr.ConfigPath = filepath.Join("profiles", pr.ID, "config.toml")
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := adapter.ImportConfig(pr); err == nil {
		t.Fatal("expected malformed config rejection")
	}
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = \"ok\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := adapter.ConfigPath(pr)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := adapter.ImportConfig(pr); err == nil {
		t.Fatal("expected existing target rejection")
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "keep" {
		t.Fatalf("existing target changed: %q", contents)
	}
}

func TestImportConfigRejectsUnsafeProfileID(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model = \"ok\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"../outside", `windows\\escape`} {
		pr := profile.Profile{ID: id, Name: "Unsafe", ConfigPath: filepath.Join("profiles", id, "config.toml")}
		if err := adapter.ImportConfig(pr); err == nil {
			t.Fatalf("expected unsafe id rejection for %q", id)
		}
	}
}

func decodeMergedCatalog(t *testing.T, existing, generated string) (map[string]json.RawMessage, []map[string]json.RawMessage) {
	t.Helper()
	merged, err := mergeCatalogBytes([]byte(existing), []byte(generated))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Metadata map[string]json.RawMessage   `json:"-"`
		Models   []map[string]json.RawMessage `json:"models"`
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(merged, &raw); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw["models"], &document.Models); err != nil {
		t.Fatal(err)
	}
	return raw, document.Models
}

func TestMergeCatalogPreservesUserPresetAndTopLevelMetadata(t *testing.T) {
	raw, models := decodeMergedCatalog(t,
		`{"schema_version":1,"preset_source":"codex","models":[{"slug":"official-model","description":"Codex preset","custom_field":"keep"}]}`,
		`{"models":[{"slug":"fresh-model","description":"Provider model"}]}`)
	if string(raw["preset_source"]) != `"codex"` || string(raw["schema_version"]) != "1" {
		t.Fatalf("top-level preset metadata lost: %#v", raw)
	}
	if len(models) != 2 {
		t.Fatalf("user preset was not retained: %#v", models)
	}
	if string(models[0]["slug"]) != `"official-model"` || string(models[0]["custom_field"]) != `"keep"` {
		t.Fatalf("official preset changed: %#v", models[0])
	}
}

func TestMergeCatalogRemovesStaleGeneratedModel(t *testing.T) {
	_, models := decodeMergedCatalog(t,
		`{"models":[{"slug":"stale-model","description":"Provider model","proxy_switch_managed":true}]}`,
		`{"models":[{"slug":"fresh-model","description":"Provider model","proxy_switch_managed":true}]}`)
	for _, model := range models {
		if string(model["slug"]) == `"stale-model"` {
			t.Fatalf("stale generated model remained: %#v", models)
		}
	}
}

func TestMergeCatalogDoesNotDeleteUnownedGenericDescription(t *testing.T) {
	_, models := decodeMergedCatalog(t,
		`{"models":[{"slug":"user-model","description":"Provider model"}]}`,
		`{"models":[{"slug":"fresh-model","description":"Provider model","proxy_switch_managed":true}]}`)
	if len(models) != 2 || string(models[0]["slug"]) != `"user-model"` {
		t.Fatalf("unowned user model was deleted: %#v", models)
	}
}

func TestMergeCatalogDoesNotReplaceUnownedSameSlug(t *testing.T) {
	_, models := decodeMergedCatalog(t,
		`{"models":[{"slug":"shared-model","description":"Provider model","custom_field":"user-value"}]}`,
		`{"models":[{"slug":"shared-model","description":"Provider model","proxy_switch_managed":true}]}`)
	if len(models) != 1 || string(models[0]["custom_field"]) != `"user-value"` {
		t.Fatalf("unowned same-slug preset was replaced: %#v", models)
	}
}

func TestMergeCatalogAppendsNewGeneratedModel(t *testing.T) {
	_, models := decodeMergedCatalog(t,
		`{"models":[{"slug":"official-model","description":"Codex preset"}]}`,
		`{"models":[{"slug":"fresh-model","description":"Provider model","proxy_switch_managed":true}]}`)
	if len(models) != 2 || string(models[1]["slug"]) != `"fresh-model"` {
		t.Fatalf("new generated model was not appended: %#v", models)
	}
}

func TestPreparePreservesUnknownCodexConfigAndComments(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "config.toml")
	original := "# keep this comment\nmodel = \"old-model\" # keep inline note\ncustom_flag = true\n\n[features]\n# keep nested comment\nweb_search = true\n\n[model_providers.custom]\nname = \"Old\"\nbase_url = \"https://old.example/v1\"\nwire_api = \"responses\"\nunknown_option = \"keep\"\nenv_key = \"OLD_KEY\"\n"
	if err := os.WriteFile(configPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("custom", "New provider", "https://new.example/v1", "NEW_KEY")
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.New(p.ID, "new-model", "New model")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "Route", p.ID, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("default", "Default")
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Prepare(context.Background(), r, p, m, []model.Model{m}, pr); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, want := range []string{
		"# keep this comment",
		"# keep inline note",
		"custom_flag = true",
		"# keep nested comment",
		"web_search = true",
		"unknown_option = \"keep\"",
		"model = \"new-model\"",
		"model_provider = \"custom\"",
		"base_url = \"https://new.example/v1\"",
		"env_key = \"NEW_KEY\"",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "env_key = \"OLD_KEY\"") {
		t.Fatalf("stale env_key remained:\n%s", text)
	}
}

func TestPrepareQuotesProviderIDsAndPreservesTableComment(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "config.toml")
	original := "model = \"old-model\"\n[model_providers.\"custom.provider\"] # keep table comment\nname = \"Old\"\nbase_url = \"https://old.example/v1\"\nwire_api = \"responses\"\n"
	if err := os.WriteFile(configPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("custom.provider", "New provider", "https://new.example/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.New(p.ID, "new-model", "New model")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "Route", p.ID, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("default", "Default")
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Prepare(context.Background(), r, p, m, []model.Model{m}, pr); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	if strings.Count(text, "[model_providers.") != 1 || !strings.Contains(text, "# keep table comment") {
		t.Fatalf("provider table was duplicated or comment lost:\n%s", text)
	}
	var config map[string]any
	if _, err := toml.Decode(string(updated), &config); err != nil {
		t.Fatalf("updated config is not valid TOML: %v\n%s", err, text)
	}
	providers, ok := config["model_providers"].(map[string]any)
	if !ok {
		t.Fatalf("model_providers missing: %#v", config["model_providers"])
	}
	if _, ok := providers[p.ID]; !ok {
		t.Fatalf("quoted provider key missing: %#v", providers)
	}
}

func TestPatchConfigRejectsMultilineOwnedValue(t *testing.T) {
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	patched, err := patchConfig([]byte("model = \"\"\"old\ncontinued\n\"\"\"\n"), p, "new", "/tmp/catalog.json", "")
	if err == nil {
		t.Fatalf("expected multiline config rejection, got %q", patched)
	}
}

func TestPatchConfigRejectsQuotedOwnedKey(t *testing.T) {
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := patchConfig([]byte(`"model" = "old"\n`), p, "new", "/tmp/catalog.json", ""); err == nil {
		t.Fatal("expected quoted owned key rejection")
	}
}

func TestPatchConfigRejectsMalformedExistingTOML(t *testing.T) {
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := patchConfig([]byte("model = [\n"), p, "new", "/tmp/catalog.json", ""); err == nil {
		t.Fatal("expected malformed TOML rejection")
	}
}

func TestPatchConfigHandlesMissingTrailingNewline(t *testing.T) {
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	patched, err := patchConfig([]byte("custom_flag = true"), p, "model", "/tmp/catalog.json", "")
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if _, err := toml.Decode(string(patched), &config); err != nil {
		t.Fatalf("patched config is invalid TOML: %v\n%s", err, patched)
	}
	if config["model"] != "model" || config["model_provider"] != "custom" {
		t.Fatalf("owned values missing after no-newline patch: %#v", config)
	}
}

func TestPrepareRejectsConfigCatalogPathCollision(t *testing.T) {
	home := t.TempDir()
	adapter, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	p, err := provider.New("custom", "Custom", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.New(p.ID, "model", "Model")
	if err != nil {
		t.Fatal(err)
	}
	r, err := route.New("route", "Route", p.ID, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := profile.New("default", "Default")
	if err != nil {
		t.Fatal(err)
	}
	pr.ModelCatalogPath = "config.toml"
	if err := adapter.Prepare(context.Background(), r, p, m, []model.Model{m}, pr); err == nil {
		t.Fatal("expected config/catalog path collision rejection")
	}
	if _, err := os.Stat(filepath.Join(home, "config.toml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("collision rejection wrote config: %v", err)
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
	for _, path := range []string{
		backupPath(activePath), missingBackupPath(activePath),
	} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("failed activation left backup artifact %q: %v", path, err)
		}
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

func TestBackupTargetsDoesNotLeaveMissingMarkerWhenStaleBackupRemovalFails(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "config.toml")
	staleBackup := backupPath(path)
	if err := os.MkdirAll(filepath.Join(staleBackup, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := backupTargets([]string{path}); err == nil {
		t.Fatal("expected stale backup removal failure")
	}
	if _, err := os.Stat(missingBackupPath(path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing marker leaked after backup failure: %v", err)
	}
}

func TestBackupTargetsRestoresEarlierArtifactsWhenLaterSnapshotFails(t *testing.T) {
	home := t.TempDir()
	first := filepath.Join(home, "first.toml")
	second := filepath.Join(home, "second.toml")
	if err := os.WriteFile(first, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath(first), []byte("old backup"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(backupPath(second), "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := backupTargets([]string{first, second}); err == nil {
		t.Fatal("expected non-regular later backup artifact failure")
	}
	contents, err := os.ReadFile(backupPath(first))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old backup" {
		t.Fatalf("earlier backup artifact changed after later failure: %q", contents)
	}
}

func TestValidateBackupTargetPathsRejectsArtifactCollision(t *testing.T) {
	home := t.TempDir()
	config := filepath.Join(home, "config.toml")
	if err := validateBackupTargetPaths([]string{config, backupPath(config)}); err == nil {
		t.Fatal("expected target/backup artifact collision rejection")
	}
}
