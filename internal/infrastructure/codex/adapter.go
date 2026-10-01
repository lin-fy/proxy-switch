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
	"codex-provider-hub/internal/infrastructure/credential"
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
	activePath := filepath.Join(a.homeDir, "config.toml")
	profilePath := a.configPath(pr)
	sourcePath := profilePath
	if profilePath != activePath {
		if _, err := os.Stat(profilePath); errors.Is(err, os.ErrNotExist) {
			sourcePath = activePath
		} else if err != nil {
			return fmt.Errorf("stat Codex profile: %w", err)
		}
	}
	config, err := readConfig(sourcePath)
	if err != nil {
		return err
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
		envName, err := credential.EnvName(p.AuthRef)
		if err != nil {
			return err
		}
		providerConfig["env_key"] = envName
	}
	providers[p.ID] = providerConfig
	config["model_providers"] = providers
	config["model_provider"] = p.ID
	config["model"] = selected.ID

	catalog, err := buildCatalog(selected, models)
	if err != nil {
		return fmt.Errorf("build model catalog: %w", err)
	}
	catalogPath := a.catalogPath(pr)
	config["model_catalog_json"] = catalogPath
	targets := uniquePaths(activePath, profilePath, catalogPath)
	existed, err := backupTargets(targets)
	if err != nil {
		return err
	}
	if err := writeCatalog(catalogPath, catalog); err != nil {
		if rollbackErr := rollbackTargets(targets, existed); rollbackErr != nil {
			return fmt.Errorf("write Codex model catalog: %w; rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("write Codex model catalog: %w", err)
	}
	for _, path := range uniquePaths(activePath, profilePath) {
		if err := writeConfig(path, config); err != nil {
			if rollbackErr := rollbackTargets(targets, existed); rollbackErr != nil {
				return fmt.Errorf("write Codex config %q: %w; rollback failed: %v", path, err, rollbackErr)
			}
			return fmt.Errorf("write Codex config %q: %w", path, err)
		}
	}
	return nil
}

func (a *Adapter) Launch(ctx context.Context, _ route.Route, p provider.Provider) error {
	cmd, err := a.command(ctx, p)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start Codex: %w", err)
	}
	return nil
}

func (a *Adapter) Restart(ctx context.Context, r route.Route, p provider.Provider) error {
	cmd, err := a.command(ctx, p)
	if err != nil {
		return err
	}
	running, err := a.IsRunning(ctx)
	if err != nil {
		return err
	}
	if running {
		image := a.executableImageName()
		cmd := exec.CommandContext(ctx, "taskkill", "/IM", image, "/T", "/F")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("stop Codex %q: %w: %s", image, err, strings.TrimSpace(string(output)))
		}
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start Codex: %w", err)
	}
	return nil
}

func (a *Adapter) command(ctx context.Context, p provider.Provider) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, a.executable)
	if err := injectCredential(cmd, p); err != nil {
		return nil, err
	}
	return cmd, nil
}

func injectCredential(cmd *exec.Cmd, p provider.Provider) error {
	if strings.TrimSpace(p.AuthRef) == "" {
		return nil
	}
	value, err := credential.Resolve(p.AuthRef)
	if err != nil {
		return fmt.Errorf("resolve provider credential: %w", err)
	}
	name, err := credential.EnvName(p.AuthRef)
	if err != nil {
		return err
	}
	env := os.Environ()
	for i := range env {
		key, _, _ := strings.Cut(env[i], "=")
		if strings.EqualFold(key, name) {
			env[i] = name + "=" + value
			cmd.Env = env
			return nil
		}
	}
	cmd.Env = append(env, name+"="+value)
	return nil
}

func (a *Adapter) IsRunning(ctx context.Context) (bool, error) {
	image := a.executableImageName()
	cmd := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq "+image)
	out, err := cmd.Output()
	if errors.Is(err, exec.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToLower(string(out)), strings.ToLower(image)), nil
}

func (a *Adapter) executableImageName() string {
	image := filepath.Base(strings.TrimSpace(a.executable))
	if filepath.Ext(image) == "" {
		image += ".exe"
	}
	return image
}

func (a *Adapter) Restore(pr profile.Profile) error {
	targets := uniquePaths(filepath.Join(a.homeDir, "config.toml"), a.configPath(pr), a.catalogPath(pr))
	backups := make(map[string][]byte, len(targets))
	missing := make(map[string]bool, len(targets))
	for _, path := range targets {
		backup := backupPath(path)
		if marker, err := os.Stat(missingBackupPath(path)); err == nil {
			if !marker.Mode().IsRegular() {
				return fmt.Errorf("invalid missing-file backup marker %q", path)
			}
			missing[path] = true
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		contents, err := os.ReadFile(backup)
		if errors.Is(err, os.ErrNotExist) {
			return ErrBackupNotFound
		}
		if err != nil {
			return fmt.Errorf("read Codex backup %q: %w", backup, err)
		}
		backups[path] = contents
	}
	originals := make(map[string][]byte, len(targets))
	existed := make(map[string]bool, len(targets))
	for _, path := range targets {
		if contents, err := os.ReadFile(path); err == nil {
			originals[path], existed[path] = contents, true
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	for i, path := range targets {
		var err error
		if missing[path] {
			err = os.Remove(path)
			if errors.Is(err, os.ErrNotExist) {
				err = nil
			}
		} else {
			err = writeCatalog(path, backups[path])
		}
		if err != nil {
			rollbackErr := rollbackFiles(originals, existed, targets[:i])
			return errors.Join(fmt.Errorf("restore Codex file %q: %w", path, err), rollbackErr)
		}
	}
	return nil
}

func (a *Adapter) ConfigPath(pr profile.Profile) string { return a.configPath(pr) }

func (a *Adapter) CatalogPath(pr profile.Profile) string { return a.catalogPath(pr) }

func (a *Adapter) configPath(pr profile.Profile) string {
	if strings.TrimSpace(pr.ConfigPath) == "" {
		return filepath.Join(a.homeDir, "config.toml")
	}
	if filepath.IsAbs(pr.ConfigPath) {
		return filepath.Clean(pr.ConfigPath)
	}
	return filepath.Join(a.homeDir, filepath.Clean(pr.ConfigPath))
}

func (a *Adapter) catalogPath(pr profile.Profile) string {
	if strings.TrimSpace(pr.ModelCatalogPath) == "" {
		return filepath.Join(a.homeDir, pr.ID+".models.json")
	}
	if filepath.IsAbs(pr.ModelCatalogPath) {
		return filepath.Clean(pr.ModelCatalogPath)
	}
	return filepath.Join(a.homeDir, filepath.Clean(pr.ModelCatalogPath))
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

func writeCatalog(path string, catalog []byte) error {
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
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(catalog); err != nil {
		_ = tmp.Close()
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

func uniquePaths(paths ...string) []string {
	result := make([]string, 0, len(paths))
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		path = filepath.Clean(path)
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	return result
}

func backupTargets(paths []string) (map[string]bool, error) {
	existed := make(map[string]bool, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existed[path] = true
			_ = os.Remove(missingBackupPath(path))
			if err := copyFile(path, backupPath(path)); err != nil {
				return nil, fmt.Errorf("backup Codex config %q: %w", path, err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat Codex config %q: %w", path, err)
		} else {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return nil, fmt.Errorf("create Codex config directory %q: %w", path, err)
			}
			if err := os.WriteFile(missingBackupPath(path), nil, 0o600); err != nil {
				return nil, fmt.Errorf("mark missing Codex config %q: %w", path, err)
			}
			if err := os.Remove(backupPath(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("remove stale Codex backup %q: %w", path, err)
			}
		}
	}
	return existed, nil
}

func rollbackTargets(paths []string, existed map[string]bool) error {
	var result error
	for _, path := range paths {
		if existed[path] {
			result = errors.Join(result, copyFile(backupPath(path), path))
			continue
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, err)
		}
	}
	return result
}

func rollbackFiles(originals map[string][]byte, existed map[string]bool, paths []string) error {
	var result error
	for _, path := range paths {
		if existed[path] {
			result = errors.Join(result, writeCatalog(path, originals[path]))
		} else if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, err)
		}
	}
	return result
}

func backupPath(path string) string { return path + ".codex-provider-hub.bak" }

func missingBackupPath(path string) string { return backupPath(path) + ".missing" }

type modelCatalog struct {
	Models []catalogModel `json:"models"`
}

type catalogModel struct {
	Priority                   int              `json:"priority"`
	SupportVerbosity           bool             `json:"support_verbosity"`
	BaseInstructions           string           `json:"base_instructions"`
	TruncationPolicy           truncationPolicy `json:"truncation_policy"`
	ExperimentalSupportedTools []string         `json:"experimental_supported_tools"`
	Slug                       string           `json:"slug"`
	DisplayName                string           `json:"display_name"`
	Description                string           `json:"description"`
	DefaultReasoningLevel      string           `json:"default_reasoning_level"`
	SupportedReasoningLevels   []reasoningLevel `json:"supported_reasoning_levels"`
	ShellType                  string           `json:"shell_type"`
	Visibility                 string           `json:"visibility"`
	SupportedInAPI             bool             `json:"supported_in_api"`
	InputModalities            []string         `json:"input_modalities"`
	SupportsParallelToolCalls  bool             `json:"supports_parallel_tool_calls"`
}

type reasoningLevel struct {
	Effort      string `json:"effort"`
	Description string `json:"description"`
}

type truncationPolicy struct {
	Mode  string `json:"mode"`
	Limit int    `json:"limit"`
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
			BaseInstructions:           "You are Codex, an AI coding assistant. Follow the user's instructions and use the available tools to complete tasks.",
			Priority:                   len(catalog.Models),
			TruncationPolicy:           truncationPolicy{Mode: "bytes", Limit: 10000},
			ExperimentalSupportedTools: []string{},
			Slug:                       item.ID, DisplayName: name, Description: "Provider model",
			DefaultReasoningLevel: "medium", SupportedReasoningLevels: levels,
			ShellType: "shell_command", Visibility: "list", SupportedInAPI: true,
			InputModalities: []string{"text", "image"}, SupportsParallelToolCalls: true,
		})
	}
	return json.Marshal(catalog)
}
