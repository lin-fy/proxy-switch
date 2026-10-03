package provider

import "testing"

func TestNew(t *testing.T) {
	p, err := New("cpa", "CPA", "http://127.0.0.1:8317/v1/", "credential:cpa")
	if err != nil {
		t.Fatal(err)
	}
	if p.BaseURL != "http://127.0.0.1:8317/v1" || p.Protocol != ProtocolResponses {
		t.Fatalf("unexpected provider: %#v", p)
	}
	if _, err := New("bad", "Bad", "not-a-url", ""); err != ErrInvalidBaseURL {
		t.Fatalf("expected invalid URL, got %v", err)
	}
}

func TestAuthRefMustBeReferenceNotSecret(t *testing.T) {
	for _, reference := range []string{"sk-live-secret", "env:", "credential:"} {
		if _, err := New("provider", "Provider", "https://example.test/v1", reference); err != ErrInvalidAuthRef {
			t.Fatalf("New(%q) error = %v, want %v", reference, err, ErrInvalidAuthRef)
		}
	}
	for _, reference := range []string{"OPENAI_API_KEY", "env:OPENAI_API_KEY", "credential:openai"} {
		if _, err := New("provider", "Provider", "https://example.test/v1", reference); err != nil {
			t.Fatalf("New(%q) error = %v", reference, err)
		}
	}
}

func TestPresetMetadataAndAuthModeNeverContainCredentialValue(t *testing.T) {
	p, err := New("openai", "OpenAI", "https://api.openai.com/v1", "env:OPENAI_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	preset, ok := MatchPreset(p)
	if !ok || preset.ID != "openai" || preset.Name != "OpenAI API" || !preset.RequiresCredential {
		t.Fatalf("preset = %#v, matched = %v", preset, ok)
	}
	if got := p.AuthMode(); got != AuthModeEnvRef {
		t.Fatalf("auth mode = %q", got)
	}
	for _, item := range Presets() {
		if item.BaseURL == "env:OPENAI_API_KEY" || item.Name == "env:OPENAI_API_KEY" {
			t.Fatalf("preset leaked credential reference: %#v", item)
		}
	}
	oauth := Provider{ID: "codex-official"}
	if got := oauth.AuthMode(); got != AuthModeOAuth {
		t.Fatalf("OAuth preset auth mode = %q", got)
	}
}
