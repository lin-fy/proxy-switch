package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"proxy-switch/internal/application/ports"
	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
	"proxy-switch/internal/infrastructure/credential"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrBackupNotFound      = errors.New("codex config backup not found")
)

type Adapter struct {
	homeDir    string
	executable string
}

// ConfigStatus describes the detected Codex files without exposing secrets.
// Import remains an explicit, separate operation; inspection never mutates
// config.toml, auth.json, profiles, or backup artifacts.
type ConfigStatus struct {
	ConfigPath        string
	ConfigExists      bool
	ConfigValid       bool
	AuthPath          string
	AuthExists        bool
	AuthValid         bool
	AuthMode          string
	CredentialPresent bool
	ModelProvider     string
	Model             string
	ProviderIDs       []string
	Importable        bool
	ImportBlocker     string
}

func (a *Adapter) InspectConfig() (ConfigStatus, error) {
	status := ConfigStatus{
		ConfigPath: filepath.Join(a.homeDir, "config.toml"),
		AuthPath:   filepath.Join(a.homeDir, "auth.json"),
	}
	configBytes, err := os.ReadFile(status.ConfigPath)
	if errors.Is(err, os.ErrNotExist) {
		return inspectAuthStatus(status)
	}
	if err != nil {
		return status, fmt.Errorf("read Codex config: %w", err)
	}
	status.ConfigExists = true
	config := map[string]any{}
	if _, err := toml.Decode(string(configBytes), &config); err != nil {
		status.ImportBlocker = "Codex config is not valid TOML"
	} else {
		status.ConfigValid = true
		status.ModelProvider, _ = config["model_provider"].(string)
		status.Model, _ = config["model"].(string)
		if providers, ok := config["model_providers"].(map[string]any); ok {
			status.ProviderIDs = make([]string, 0, len(providers))
			for id := range providers {
				status.ProviderIDs = append(status.ProviderIDs, id)
			}
			sort.Strings(status.ProviderIDs)
		}
		providerKnown := false
		for _, id := range status.ProviderIDs {
			if id == status.ModelProvider {
				providerKnown = true
				break
			}
		}
		if status.ModelProvider == "" || !providerKnown {
			status.ImportBlocker = "Codex config has no importable model provider"
		} else {
			status.Importable = true
		}
	}
	return inspectAuthStatus(status)
}

// ImportConfig copies the current Codex config into an explicit profile path.
// The source is parsed before writing so malformed configuration never becomes
// a profile. auth.json is intentionally never read or copied by this method.
func (a *Adapter) ImportConfig(pr profile.Profile) error {
	activePath := filepath.Join(a.homeDir, "config.toml")
	targetPath := a.configPath(pr)
	if filepath.Clean(activePath) == filepath.Clean(targetPath) {
		return errors.New("Codex import target must differ from the active config")
	}
	if !safeImportProfileID(pr.ID) {
		return errors.New("Codex import profile id contains unsafe path characters")
	}
	contents, err := os.ReadFile(activePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("Codex config.toml not found")
	}
	if err != nil {
		return fmt.Errorf("read Codex config for import: %w", err)
	}
	if _, err := toml.Decode(string(contents), &map[string]any{}); err != nil {
		return fmt.Errorf("Codex config is not valid TOML: %w", err)
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("Codex import target already exists: %q", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat Codex import target: %w", err)
	}
	if err := writeConfig(targetPath, contents); err != nil {
		return fmt.Errorf("write imported Codex profile: %w", err)
	}
	return nil
}

func safeImportProfileID(id string) bool {
	id = strings.TrimSpace(id)
	return id != "" && id != "." && id != ".." && !strings.ContainsAny(id, `/\\`)
}

func inspectAuthStatus(status ConfigStatus) (ConfigStatus, error) {
	authBytes, err := os.ReadFile(status.AuthPath)
	if errors.Is(err, os.ErrNotExist) {
		status.AuthMode = "missing"
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("read Codex auth: %w", err)
	}
	status.AuthExists = true
	var auth map[string]json.RawMessage
	if err := json.Unmarshal(authBytes, &auth); err != nil {
		status.AuthMode = "invalid"
		return status, nil
	}
	status.AuthValid = true
	status.AuthMode, status.CredentialPresent = classifyAuth(auth)
	return status, nil
}

func classifyAuth(auth map[string]json.RawMessage) (string, bool) {
	if len(auth) == 0 {
		return "present", false
	}
	for key := range auth {
		lower := strings.ToLower(strings.TrimSpace(key))
		if strings.Contains(lower, "token") || strings.Contains(lower, "oauth") || strings.Contains(lower, "refresh") {
			return "oauth", true
		}
	}
	for key := range auth {
		if strings.EqualFold(strings.TrimSpace(key), "OPENAI_API_KEY") {
			return "api_key", true
		}
	}
	return "present", true
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
	config, err := os.ReadFile(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		config = nil
		err = nil
	}
	if err != nil {
		return err
	}

	providerEnv := ""
	if p.AuthRef != "" {
		envName, err := credential.EnvName(p.AuthRef)
		if err != nil {
			return err
		}
		providerEnv = envName
	}

	catalog, err := buildCatalog(selected, models)
	if err != nil {
		return fmt.Errorf("build model catalog: %w", err)
	}
	catalogPath := a.catalogPath(pr)
	if catalogPath == activePath || catalogPath == profilePath {
		return fmt.Errorf("Codex model catalog path must differ from config path")
	}
	catalog, err = mergeCatalogFile(catalogPath, catalog)
	if err != nil {
		return fmt.Errorf("merge Codex model catalog: %w", err)
	}
	config, err = patchConfig(config, p, selected.ID, catalogPath, providerEnv)
	if err != nil {
		return fmt.Errorf("patch Codex config: %w", err)
	}
	targets := uniquePaths(activePath, profilePath, catalogPath)
	if err := validateBackupTargetPaths(targets); err != nil {
		return err
	}
	existed, backupState, err := backupTargets(targets)
	if err != nil {
		return err
	}
	if err := writeCatalog(catalogPath, catalog); err != nil {
		if rollbackErr := rollbackTargets(targets, existed); rollbackErr != nil {
			if artifactErr := restoreBackupArtifacts(backupState); artifactErr != nil {
				return fmt.Errorf("write Codex model catalog: %w; rollback failed: %v; backup rollback failed: %v", err, rollbackErr, artifactErr)
			}
			return fmt.Errorf("write Codex model catalog: %w; rollback failed: %v", err, rollbackErr)
		}
		if artifactErr := restoreBackupArtifacts(backupState); artifactErr != nil {
			return fmt.Errorf("write Codex model catalog: %w; backup rollback failed: %v", err, artifactErr)
		}
		return fmt.Errorf("write Codex model catalog: %w", err)
	}
	for _, path := range uniquePaths(activePath, profilePath) {
		if err := writeConfig(path, config); err != nil {
			if rollbackErr := rollbackTargets(targets, existed); rollbackErr != nil {
				if artifactErr := restoreBackupArtifacts(backupState); artifactErr != nil {
					return fmt.Errorf("write Codex config %q: %w; rollback failed: %v; backup rollback failed: %v", path, err, rollbackErr, artifactErr)
				}
				return fmt.Errorf("write Codex config %q: %w; rollback failed: %v", path, err, rollbackErr)
			}
			if artifactErr := restoreBackupArtifacts(backupState); artifactErr != nil {
				return fmt.Errorf("write Codex config %q: %w; backup rollback failed: %v", path, err, artifactErr)
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
	// filepath.Base follows the host OS separator. Normalize Windows paths
	// first so tests and configuration inspection behave consistently on the
	// Linux build host as well as Windows.
	image := filepath.Base(strings.ReplaceAll(strings.TrimSpace(a.executable), `\`, "/"))
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

func writeConfig(path string, config []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".proxy-switch-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(config); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

var topLevelKeyPattern = regexp.MustCompile(`^(\s*)([A-Za-z0-9_-]+)\s*=`)
var quotedTopLevelKeyPattern = regexp.MustCompile(`^(\s*)"([^"]+)"\s*=`)

// patchConfig updates only the fields owned by Proxy Switch. Re-encoding the
// complete TOML map loses comments and unknown Codex settings, so this narrow
// line patch leaves the user's configuration intact.
func patchConfig(contents []byte, p provider.Provider, modelID, catalogPath, envKey string) ([]byte, error) {
	if strings.TrimSpace(string(contents)) != "" {
		original := map[string]any{}
		if _, err := toml.Decode(string(contents), &original); err != nil {
			return nil, fmt.Errorf("validate existing Codex config: %w", err)
		}
	}
	text := string(contents)
	var err error
	if text, err = patchTopLevel(text, "model_provider", quoteTOML(p.ID)); err != nil {
		return nil, err
	}
	if text, err = patchTopLevel(text, "model", quoteTOML(modelID)); err != nil {
		return nil, err
	}
	if text, err = patchTopLevel(text, "model_catalog_json", quoteTOML(catalogPath)); err != nil {
		return nil, err
	}
	if text, err = patchProviderTable(text, p, envKey); err != nil {
		return nil, err
	}
	if err := validatePatchedConfig([]byte(text), p, modelID, catalogPath, envKey); err != nil {
		return nil, err
	}
	return []byte(text), nil
}

func validatePatchedConfig(contents []byte, p provider.Provider, modelID, catalogPath, envKey string) error {
	config := map[string]any{}
	if _, err := toml.Decode(string(contents), &config); err != nil {
		return fmt.Errorf("validate patched Codex config: %w", err)
	}
	if got, ok := config["model_provider"].(string); !ok || got != p.ID {
		return fmt.Errorf("validate patched Codex config: model_provider did not resolve to %q", p.ID)
	}
	if got, ok := config["model"].(string); !ok || got != modelID {
		return fmt.Errorf("validate patched Codex config: model did not resolve to %q", modelID)
	}
	if got, ok := config["model_catalog_json"].(string); !ok || got != catalogPath {
		return fmt.Errorf("validate patched Codex config: model_catalog_json did not resolve to %q", catalogPath)
	}
	providers, ok := config["model_providers"].(map[string]any)
	if !ok {
		return fmt.Errorf("validate patched Codex config: model_providers is not a table")
	}
	providerConfig, ok := providers[p.ID].(map[string]any)
	if !ok {
		return fmt.Errorf("validate patched Codex config: provider %q is not a table", p.ID)
	}
	for key, want := range map[string]string{"name": p.Name, "base_url": p.BaseURL, "wire_api": p.Protocol} {
		if got, ok := providerConfig[key].(string); !ok || got != want {
			return fmt.Errorf("validate patched Codex config: provider %q field %q did not resolve to %q", p.ID, key, want)
		}
	}
	if envKey == "" {
		if _, ok := providerConfig["env_key"]; ok {
			return fmt.Errorf("validate patched Codex config: stale env_key remained")
		}
	} else if got, ok := providerConfig["env_key"].(string); !ok || got != envKey {
		return fmt.Errorf("validate patched Codex config: env_key did not resolve to %q", envKey)
	}
	return nil
}

func quoteTOML(value string) string {
	// TOML basic strings share the common escapes used by Go. strconv.Quote is
	// not used because it also emits Go-only \a, \v and \xNN escapes.
	var builder strings.Builder
	builder.Grow(len(value) + 2)
	builder.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '\b':
			builder.WriteString(`\b`)
		case '\t':
			builder.WriteString(`\t`)
		case '\n':
			builder.WriteString(`\n`)
		case '\f':
			builder.WriteString(`\f`)
		case '\r':
			builder.WriteString(`\r`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&builder, `\u%04X`, r)
				continue
			}
			builder.WriteRune(r)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func patchTopLevel(contents, key, value string) (string, error) {
	lines := strings.SplitAfter(contents, "\n")
	var multiline byte
	for i, line := range lines {
		body, ending := splitLineEnding(line)
		if multiline != 0 {
			multiline = advanceMultilineState(body, multiline)
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(body), "[") {
			break
		}
		match := topLevelKeyPattern.FindStringSubmatch(body)
		if len(match) == 3 && match[2] == key {
			if hasMultilineDelimiter(body) {
				return "", fmt.Errorf("unsupported multiline TOML value for top-level key %q", key)
			}
			lines[i] = match[1] + key + " = " + value + commentSuffix(body) + ending
			return strings.Join(lines, ""), nil
		}
		if quoted := quotedTopLevelKeyPattern.FindStringSubmatch(body); len(quoted) == 3 && quoted[2] == key {
			return "", fmt.Errorf("unsupported quoted TOML key %q", key)
		}
		if hasMultilineDelimiter(body) {
			multiline = advanceMultilineState(body, 0)
		}
	}
	insertAt := 0
	for i, line := range lines {
		body, _ := splitLineEnding(line)
		if strings.HasPrefix(strings.TrimSpace(body), "[") {
			insertAt = i
			break
		}
		insertAt = i + 1
	}
	if insertAt > 0 {
		_, ending := splitLineEnding(lines[insertAt-1])
		if ending == "" {
			lines[insertAt-1] += "\n"
		}
	}
	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = key + " = " + value + "\n"
	return strings.Join(lines, ""), nil
}

func patchProviderTable(contents string, p provider.Provider, envKey string) (string, error) {
	lines := strings.SplitAfter(contents, "\n")
	headerIndex, endIndex := -1, len(lines)
	var multiline byte
	for i, line := range lines {
		body, _ := splitLineEnding(line)
		if multiline != 0 {
			multiline = advanceMultilineState(body, multiline)
			continue
		}
		trimmed := strings.TrimSpace(body)
		if headerIndex >= 0 && strings.HasPrefix(trimmed, "[") {
			endIndex = i
			break
		}
		if headerIndex < 0 && providerTableHeaderMatches(trimmed, p.ID) {
			headerIndex = i
		}
		if hasMultilineDelimiter(body) {
			multiline = advanceMultilineState(body, 0)
		}
	}
	entries := []struct{ key, value string }{
		{key: "name", value: quoteTOML(p.Name)},
		{key: "base_url", value: quoteTOML(p.BaseURL)},
		{key: "wire_api", value: quoteTOML(p.Protocol)},
	}
	if envKey != "" {
		entries = append(entries, struct{ key, value string }{key: "env_key", value: quoteTOML(envKey)})
	}
	if headerIndex < 0 {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			_, ending := splitLineEnding(lines[len(lines)-1])
			if ending == "" {
				lines[len(lines)-1] += "\n"
			}
			lines = append(lines, "\n")
		}
		lines = append(lines, "[model_providers."+quoteTOML(p.ID)+"]\n")
		for _, entry := range entries {
			lines = append(lines, entry.key+" = "+entry.value+"\n")
		}
		return strings.Join(lines, ""), nil
	}
	for _, entry := range entries {
		keyPattern := regexp.MustCompile(`^(\s*)` + regexp.QuoteMeta(entry.key) + `\s*=`)
		found := false
		for i := headerIndex + 1; i < endIndex; i++ {
			body, ending := splitLineEnding(lines[i])
			match := keyPattern.FindStringSubmatch(body)
			if len(match) == 2 {
				lines[i] = match[1] + entry.key + " = " + entry.value + commentSuffix(body) + ending
				found = true
				break
			}
			quotedKeyPattern := regexp.MustCompile(`^\s*"` + regexp.QuoteMeta(entry.key) + `"\s*=`)
			if quotedKeyPattern.MatchString(body) {
				return "", fmt.Errorf("unsupported quoted TOML provider key %q", entry.key)
			}
		}
		if !found {
			if endIndex > 0 {
				_, ending := splitLineEnding(lines[endIndex-1])
				if ending == "" {
					lines[endIndex-1] += "\n"
				}
			}
			lines = append(lines, "")
			copy(lines[endIndex+1:], lines[endIndex:])
			lines[endIndex] = entry.key + " = " + entry.value + "\n"
			endIndex++
		}
	}
	if envKey == "" {
		keyPattern := regexp.MustCompile(`^\s*env_key\s*=`)
		filtered := lines[:0]
		for i, line := range lines {
			body, _ := splitLineEnding(line)
			if i > headerIndex && i < endIndex && keyPattern.MatchString(body) {
				continue
			}
			filtered = append(filtered, line)
		}
		lines = filtered
	}
	return strings.Join(lines, ""), nil
}

func hasMultilineDelimiter(line string) bool {
	return strings.Contains(line, `"""`) || strings.Contains(line, `'''`)
}

func advanceMultilineState(line string, current byte) byte {
	if current != 0 {
		delimiter := strings.Repeat(string(current), 3)
		if strings.Count(line, delimiter)%2 == 1 {
			return 0
		}
		return current
	}
	if strings.Count(line, `"""`)%2 == 1 {
		return '"'
	}
	if strings.Count(line, `'''`)%2 == 1 {
		return '\''
	}
	return 0
}

func splitLineEnding(line string) (body, ending string) {
	switch {
	case strings.HasSuffix(line, "\r\n"):
		return strings.TrimSuffix(line, "\r\n"), "\r\n"
	case strings.HasSuffix(line, "\n"), strings.HasSuffix(line, "\r"):
		return line[:len(line)-1], line[len(line)-1:]
	default:
		return line, ""
	}
}

func commentSuffix(line string) string {
	inBasic, inLiteral := false, false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			if !inLiteral && (i == 0 || line[i-1] != '\\') {
				inBasic = !inBasic
			}
		case '\'':
			if !inBasic {
				inLiteral = !inLiteral
			}
		case '#':
			if !inBasic && !inLiteral {
				start := i
				for start > 0 && (line[start-1] == ' ' || line[start-1] == '\t') {
					start--
				}
				return line[start:]
			}
		}
	}
	return ""
}

func providerTableHeaderMatches(line, providerID string) bool {
	line = strings.TrimSpace(stripComment(line))
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return false
	}
	body := strings.TrimSpace(line[1 : len(line)-1])
	const prefix = "model_providers."
	if !strings.HasPrefix(body, prefix) {
		return false
	}
	segment := strings.TrimSpace(body[len(prefix):])
	if segment == providerID {
		return true
	}
	if len(segment) >= 2 && segment[0] == '"' && segment[len(segment)-1] == '"' {
		return unquoteTOML(segment) == providerID
	}
	if len(segment) >= 2 && segment[0] == '\'' && segment[len(segment)-1] == '\'' {
		return segment[1:len(segment)-1] == providerID
	}
	return false
}

func unquoteTOML(value string) string {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return ""
	}
	value = value[1 : len(value)-1]
	var builder strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' || i+1 >= len(value) {
			builder.WriteByte(value[i])
			continue
		}
		i++
		switch value[i] {
		case '\\', '"':
			builder.WriteByte(value[i])
		case 'b':
			builder.WriteByte('\b')
		case 't':
			builder.WriteByte('\t')
		case 'n':
			builder.WriteByte('\n')
		case 'f':
			builder.WriteByte('\f')
		case 'r':
			builder.WriteByte('\r')
		default:
			return ""
		}
	}
	return builder.String()
}

func stripComment(line string) string {
	inBasic, inLiteral := false, false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			if !inLiteral && (i == 0 || line[i-1] != '\\') {
				inBasic = !inBasic
			}
		case '\'':
			if !inBasic {
				inLiteral = !inLiteral
			}
		case '#':
			if !inBasic && !inLiteral {
				return line[:i]
			}
		}
	}
	return line
}

func writeCatalog(path string, catalog []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".proxy-switch-*.tmp")
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
	return writeAtomic(destination, contents, 0o600)
}

func writeAtomic(path string, contents []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".proxy-switch-backup-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(contents); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
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

func validateBackupTargetPaths(paths []string) error {
	targets := make(map[string]bool, len(paths))
	for _, path := range paths {
		targets[filepath.Clean(path)] = true
	}
	for _, path := range paths {
		path = filepath.Clean(path)
		for _, artifact := range []string{backupPath(path), missingBackupPath(path)} {
			if targets[filepath.Clean(artifact)] {
				return fmt.Errorf("Codex config path %q collides with backup artifact %q", path, artifact)
			}
		}
	}
	return nil
}

type backupArtifact struct {
	exists bool
	mode   os.FileMode
	data   []byte
}

type backupArtifactState struct {
	backup  backupArtifact
	missing backupArtifact
}

func backupTargets(paths []string) (map[string]bool, map[string]backupArtifactState, error) {
	existed := make(map[string]bool, len(paths))
	state := make(map[string]backupArtifactState, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existed[path] = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("stat Codex config %q: %w", path, err)
		}
		backup, err := snapshotBackupArtifact(backupPath(path))
		if err != nil {
			return nil, nil, err
		}
		missing, err := snapshotBackupArtifact(missingBackupPath(path))
		if err != nil {
			return nil, nil, err
		}
		state[path] = backupArtifactState{backup: backup, missing: missing}
	}
	for _, path := range paths {
		if existed[path] {
			if err := os.Remove(missingBackupPath(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return rollbackBackupPreparation(state, fmt.Errorf("remove stale missing Codex marker %q: %w", path, err))
			}
			if err := copyFile(path, backupPath(path)); err != nil {
				return rollbackBackupPreparation(state, fmt.Errorf("backup Codex config %q: %w", path, err))
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return rollbackBackupPreparation(state, fmt.Errorf("create Codex config directory %q: %w", path, err))
		}
		if err := os.Remove(backupPath(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return rollbackBackupPreparation(state, fmt.Errorf("remove stale Codex backup %q: %w", path, err))
		}
		if err := os.WriteFile(missingBackupPath(path), nil, 0o600); err != nil {
			return rollbackBackupPreparation(state, fmt.Errorf("mark missing Codex config %q: %w", path, err))
		}
	}
	return existed, state, nil
}

func snapshotBackupArtifact(path string) (backupArtifact, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return backupArtifact{}, nil
	}
	if err != nil {
		return backupArtifact{}, fmt.Errorf("stat backup artifact %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return backupArtifact{}, fmt.Errorf("backup artifact %q is not a regular file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return backupArtifact{}, fmt.Errorf("read backup artifact %q: %w", path, err)
	}
	return backupArtifact{exists: true, mode: info.Mode().Perm(), data: data}, nil
}

func restoreBackupArtifacts(state map[string]backupArtifactState) error {
	var result error
	for path, artifacts := range state {
		result = errors.Join(result, restoreBackupArtifact(backupPath(path), artifacts.backup))
		result = errors.Join(result, restoreBackupArtifact(missingBackupPath(path), artifacts.missing))
	}
	return result
}

func rollbackBackupPreparation(state map[string]backupArtifactState, cause error) (map[string]bool, map[string]backupArtifactState, error) {
	if err := restoreBackupArtifacts(state); err != nil {
		return nil, nil, fmt.Errorf("%w; backup rollback failed: %v", cause, err)
	}
	return nil, nil, cause
}

func restoreBackupArtifact(path string, artifact backupArtifact) error {
	if !artifact.exists {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return writeAtomic(path, artifact.data, artifact.mode)
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

func backupPath(path string) string { return path + ".proxy-switch.bak" }

func missingBackupPath(path string) string { return backupPath(path) + ".missing" }

type modelCatalog struct {
	Models []catalogModel `json:"models"`
}

func mergeCatalogFile(path string, generated []byte) ([]byte, error) {
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return generated, nil
	}
	if err != nil {
		return nil, err
	}
	return mergeCatalogBytes(existing, generated)
}

// mergeCatalogBytes keeps Codex's bundled/user-defined presets while replacing
// entries previously generated by Proxy Switch. Generated entries are marked
// by their stable description so disabled local models do not linger forever,
// while unrelated catalog metadata and top-level keys remain intact.
func mergeCatalogBytes(existing, generated []byte) ([]byte, error) {
	var existingDoc, generatedDoc map[string]json.RawMessage
	if err := json.Unmarshal(existing, &existingDoc); err != nil {
		return nil, fmt.Errorf("decode existing catalog: %w", err)
	}
	if err := json.Unmarshal(generated, &generatedDoc); err != nil {
		return nil, fmt.Errorf("decode generated catalog: %w", err)
	}
	existingModels, err := catalogModelEntries(existingDoc)
	if err != nil {
		return nil, fmt.Errorf("decode existing catalog models: %w", err)
	}
	generatedModels, err := catalogModelEntries(generatedDoc)
	if err != nil {
		return nil, fmt.Errorf("decode generated catalog models: %w", err)
	}
	generatedBySlug := make(map[string]json.RawMessage, len(generatedModels))
	generatedOrder := make([]string, 0, len(generatedModels))
	for _, raw := range generatedModels {
		slug, err := catalogSlug(raw)
		if err != nil || slug == "" {
			continue
		}
		generatedBySlug[slug] = raw
		generatedOrder = append(generatedOrder, slug)
	}
	merged := make([]json.RawMessage, 0, len(existingModels)+len(generatedModels))
	consumed := make(map[string]bool, len(generatedBySlug))
	for _, raw := range existingModels {
		slug, _ := catalogSlug(raw)
		owned, _ := catalogManaged(raw)
		if owned {
			if replacement, ok := generatedBySlug[slug]; ok {
				merged = append(merged, replacement)
				consumed[slug] = true
			}
			continue
		}
		if _, ok := generatedBySlug[slug]; ok && slug != "" {
			// An unowned preset with the same slug wins. Keeping its raw entry
			// avoids silently replacing user metadata with generated defaults.
			merged = append(merged, raw)
			consumed[slug] = true
			continue
		}
		merged = append(merged, raw)
	}
	for _, slug := range generatedOrder {
		if !consumed[slug] {
			merged = append(merged, generatedBySlug[slug])
		}
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("encode merged catalog models: %w", err)
	}
	if existingDoc == nil {
		existingDoc = map[string]json.RawMessage{}
	}
	existingDoc["models"] = encoded
	result, err := json.Marshal(existingDoc)
	if err != nil {
		return nil, fmt.Errorf("encode merged catalog: %w", err)
	}
	return result, nil
}

func catalogModelEntries(doc map[string]json.RawMessage) ([]json.RawMessage, error) {
	raw := doc["models"]
	if len(raw) == 0 {
		return nil, nil
	}
	var models []json.RawMessage
	if err := json.Unmarshal(raw, &models); err != nil {
		return nil, err
	}
	return models, nil
}

func catalogSlug(raw json.RawMessage) (string, error) {
	var item struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return "", err
	}
	return strings.TrimSpace(item.Slug), nil
}

func catalogManaged(raw json.RawMessage) (bool, error) {
	var item struct {
		Managed bool `json:"proxy_switch_managed"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return false, err
	}
	return item.Managed, nil
}

type catalogModel struct {
	Managed                    bool             `json:"proxy_switch_managed,omitempty"`
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
			Managed:                    true,
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
