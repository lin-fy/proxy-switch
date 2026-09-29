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

	"github.com/BurntSushi/toml"

	"codex-provider-hub/internal/application/ports"
	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrBackupNotFound      = errors.New("codex config backup not found")
)

type Adapter struct {
	homeDir    string
	executable string
}

func NewAdapter(homeDir string) (*Adapter, error) {
	if strings.TrimSpace(homeDir) == "" {
		var err error
		homeDir, err = resolveHomeDir()
		if err != nil {
			return nil, err
		}
	}
	return &Adapter{
		homeDir:    filepath.Clean(homeDir),
		executable: envOr("CODEX_DESKTOP_EXECUTABLE", "codex"),
	}, nil
}

type Factory struct{ adapter *Adapter }

func NewFactory(homeDir string) (*Factory, error) {
	adapter, err := NewAdapter(homeDir)
	if err != nil {
		return nil, err
	}
	return &Factory{adapter: adapter}, nil
}

func (f *Factory) Create(platformID string) (ports.PlatformAdapter, error) {
	if platformID != route.PlatformCodex {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platformID)
	}
	return f.adapter, nil
}

func (a *Adapter) Validate(_ context.Context, r route.Route, p provider.Provider, m model.Model) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if m.ProviderID != p.ID {
		return fmt.Errorf("model provider %q does not match provider %q", m.ProviderID, p.ID)
	}
	if strings.TrimSpace(m.ID) == "" {
		return model.ErrInvalidID
	}
	return nil
}

func (a *Adapter) Prepare(_ context.Context, r route.Route, p provider.Provider, selected model.Model, models []model.Model, pr profile.Profile) error {
	if err := a.Validate(context.Background(), r, p, selected); err != nil {
		return err
	}
	path := a.configPath(pr)
	config, err := readConfig(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		if err := copyFile(path, backupPath(path)); err != nil {
			return fmt.Errorf("backup codex config: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat codex config: %w", err)
	}

	providers, err := table(config, "model_providers")
	if err != nil {
		return err
	}
	providerConfig := map[string]any{
		"name":     p.Name,
		"base_url": p.BaseURL,
		"wire_api": p.Protocol,
	}
	if p.AuthRef != "" {
		providerConfig["env_key"] = p.AuthRef
	}
	providers[p.ID] = providerConfig
	config["model_providers"] = providers
	config["model_provider"] = p.ID
	config["model"] = selected.ID

	catalog, err := buildCatalog(selected, models)
	if err != nil {
		return fmt.Errorf("build model catalog: %w", err)
	}
	config["model_catalog_json"] = string(catalog)
	if err := writeConfig(path, config); err != nil {
		return fmt.Errorf("write codex config: %w", err)
	}
	return nil
}

func (a *Adapter) Launch(ctx context.Context, _ route.Route) error {
	cmd := exec.CommandContext(ctx, a.executable)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start Codex: %w", err)
	}
	return nil
}

func (a *Adapter) IsRunning(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq codex.exe")
	out, err := cmd.Output()
	if errors.Is(err, exec.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToLower(string(out)), "codex.exe"), nil
}

func (a *Adapter) Restore(pr profile.Profile) error {
	path := a.configPath(pr)
	backup := backupPath(path)
	if _, err := os.Stat(backup); errors.Is(err, os.ErrNotExist) {
		return ErrBackupNotFound
	} else if err != nil {
		return err
	}
	if err := copyFile(backup, path); err != nil {
		return fmt.Errorf("restore Codex config: %w", err)
	}
	return nil
}

func (a *Adapter) ConfigPath(pr profile.Profile) string { return a.configPath(pr) }

func (a *Adapter) configPath(pr profile.Profile) string {
	if strings.TrimSpace(pr.ConfigPath) == "" {
		return filepath.Join(a.homeDir, "config.toml")
	}
	if filepath.IsAbs(pr.ConfigPath) {
		return filepath.Clean(pr.ConfigPath)
	}
	return filepath.Join(a.homeDir, filepath.Clean(pr.ConfigPath))
}

func resolveHomeDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv("CODEX_HOME")); value != "" {
		return value, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, ".codex"), nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func readConfig(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	config := map[string]any{}
	if _, err := toml.Decode(string(contents), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func table(config map[string]any, key string) (map[string]any, error) {
	if value, ok := config[key]; ok {
		if result, ok := value.(map[string]any); ok {
			return result, nil
		}
		return nil, fmt.Errorf("Codex config key %q is not a table", key)
	}
	return map[string]any{}, nil
}

func writeConfig(path string, config map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".codex-provider-hub-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := toml.NewEncoder(tmp).Encode(config); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func copyFile(source, destination string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	return os.WriteFile(destination, contents, 0o600)
}

func backupPath(path string) string { return path + ".codex-provider-hub.bak" }

type modelCatalog struct {
	Models []catalogModel `json:"models"`
}

type catalogModel struct {
	Slug                      string           `json:"slug"`
	DisplayName               string           `json:"display_name"`
	Description               string           `json:"description"`
	DefaultReasoningLevel     string           `json:"default_reasoning_level"`
	SupportedReasoningLevels  []reasoningLevel `json:"supported_reasoning_levels"`
	ShellType                 string           `json:"shell_type"`
	Visibility                string           `json:"visibility"`
	SupportedInAPI            bool             `json:"supported_in_api"`
	InputModalities           []string         `json:"input_modalities"`
	SupportsParallelToolCalls bool             `json:"supports_parallel_tool_calls"`
}

type reasoningLevel struct {
	Effort      string `json:"effort"`
	Description string `json:"description"`
}

func buildCatalog(selected model.Model, models []model.Model) ([]byte, error) {
	items := make([]model.Model, 0, len(models)+1)
	seen := map[string]bool{}
	for _, item := range models {
		if item.ProviderID != selected.ProviderID || strings.TrimSpace(item.ID) == "" || !item.Enabled || seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		items = append(items, item)
	}
	if !seen[selected.ID] {
		items = append(items, selected)
	}
	levels := []reasoningLevel{
		{Effort: "low", Description: "Fast responses with lighter reasoning"},
		{Effort: "medium", Description: "Balances speed and reasoning depth"},
		{Effort: "high", Description: "Greater reasoning depth for complex tasks"},
		{Effort: "xhigh", Description: "Extra high reasoning depth for complex tasks"},
		{Effort: "max", Description: "Maximum reasoning depth"},
		{Effort: "ultra", Description: "Maximum reasoning with delegation"},
	}
	catalog := modelCatalog{Models: make([]catalogModel, 0, len(items))}
	for _, item := range items {
		name := item.Name
		if strings.TrimSpace(name) == "" {
			name = item.ID
		}
		catalog.Models = append(catalog.Models, catalogModel{
			Slug: item.ID, DisplayName: name, Description: "Provider model",
			DefaultReasoningLevel: "medium", SupportedReasoningLevels: levels,
			ShellType: "shell_command", Visibility: "list", SupportedInAPI: true,
			InputModalities: []string{"text", "image"}, SupportsParallelToolCalls: true,
		})
	}
	return json.Marshal(catalog)
}
