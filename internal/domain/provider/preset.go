package provider

import "strings"

type AuthMode string

const (
	AuthModeNone       AuthMode = "none"
	AuthModeEnvRef     AuthMode = "env_ref"
	AuthModeCredential AuthMode = "credential_ref"
	AuthModeOAuth      AuthMode = "oauth"
)

// PresetMetadata is read-only catalog data. It never contains a credential or
// an auth.json payload and is safe to expose to the desktop UI.
type PresetMetadata struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	BaseURL            string   `json:"base_url"`
	AuthMode           AuthMode `json:"auth_mode"`
	RequiresCredential bool     `json:"requires_credential"`
}

var presetCatalog = []PresetMetadata{
	{ID: "codex-official", Name: "OpenAI Codex 官方登录", AuthMode: AuthModeOAuth},
	{ID: "openai", Name: "OpenAI API", BaseURL: "https://api.openai.com/v1", AuthMode: AuthModeEnvRef, RequiresCredential: true},
	{ID: "xai", Name: "xAI API", BaseURL: "https://api.x.ai/v1", AuthMode: AuthModeEnvRef, RequiresCredential: true},
	{ID: "kimi", Name: "Kimi API", BaseURL: "https://api.moonshot.cn/v1", AuthMode: AuthModeEnvRef, RequiresCredential: true},
}

func Presets() []PresetMetadata {
	result := make([]PresetMetadata, len(presetCatalog))
	copy(result, presetCatalog)
	return result
}

func MatchPreset(item Provider) (PresetMetadata, bool) {
	for _, preset := range presetCatalog {
		if strings.EqualFold(strings.TrimSpace(item.ID), preset.ID) {
			return preset, true
		}
	}
	baseURL := strings.TrimRight(strings.TrimSpace(item.BaseURL), "/")
	for _, preset := range presetCatalog {
		if preset.BaseURL != "" && strings.EqualFold(baseURL, preset.BaseURL) {
			return preset, true
		}
	}
	return PresetMetadata{}, false
}

func (p Provider) AuthMode() AuthMode {
	ref := strings.TrimSpace(p.AuthRef)
	if ref == "" {
		if preset, ok := MatchPreset(p); ok && preset.AuthMode == AuthModeOAuth {
			return AuthModeOAuth
		}
		return AuthModeNone
	}
	if strings.HasPrefix(ref, "credential:") {
		return AuthModeCredential
	}
	return AuthModeEnvRef
}
