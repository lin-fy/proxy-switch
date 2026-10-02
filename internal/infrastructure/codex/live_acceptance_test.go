package codex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
	"proxy-switch/internal/infrastructure/credential"
)

// Opt in explicitly: this test invokes the real Codex CLI and consumes API usage.
func TestLiveCodexGeneratedConfig(t *testing.T) {
	if os.Getenv("PROVIDER_HUB_LIVE_TEST") != "1" {
		t.Skip("set PROVIDER_HUB_LIVE_TEST=1 with provider URL, model IDs and credential reference")
	}
	url := os.Getenv("PROVIDER_HUB_LIVE_URL")
	ids := strings.Split(os.Getenv("PROVIDER_HUB_LIVE_MODELS"), ",")
	p, err := provider.New("live", "Live acceptance", url, os.Getenv("PROVIDER_HUB_LIVE_AUTH_REF"))
	if err != nil {
		t.Fatal(err)
	}
	secret, _ := credential.Resolve(p.AuthRef)
	// Codex refuses helper binaries under the system temporary directory.
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "bin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	home, err := os.MkdirTemp(root, "live-codex-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	t.Setenv("CODEX_HOME", home)
	a, err := NewAdapter(home)
	if err != nil {
		t.Fatal(err)
	}
	pr, _ := profile.New("live", "Live")
	models := make([]model.Model, 0, len(ids))
	for _, id := range ids {
		m, err := model.New(p.ID, id, id)
		if err != nil {
			t.Fatal(err)
		}
		models = append(models, m)
	}
	for _, m := range models {
		r, _ := route.New("live", "Live", p.ID, m.ID)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		if err := a.Prepare(ctx, r, p, m, models, pr); err != nil {
			cancel()
			t.Fatal(err)
		}
		cmd, err := a.command(ctx, p)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		cmd.Args = append(cmd.Args, "exec", "--ephemeral", "--skip-git-repo-check", "--sandbox", "read-only", "-C", home, "Reply with exactly OK. Do not use tools.")
		output, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			// Codex config diagnostics don't contain credentials; nevertheless redact
			// the resolved value from any output before reporting test failures.
			if secret != "" {
				output = []byte(strings.ReplaceAll(string(output), secret, "<redacted>"))
			}
			t.Fatalf("Codex model %s: %v\n%s", m.ID, err, output)
		}
		if !strings.Contains(string(output), "OK") {
			t.Fatalf("Codex model %s did not return the expected response", m.ID)
		}
		t.Logf("fresh Codex process read generated config/catalog and responded using %s", m.ID)
	}
}
