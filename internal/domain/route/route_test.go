package route

import "testing"

func TestNewDefaultsToCodexWithoutRestart(t *testing.T) {
	r, err := New("daily", "", "cpa", "gpt-5")
	if err != nil {
		t.Fatal(err)
	}
	if r.PlatformID != PlatformCodex || r.RestartOnActivate || r.Name != "daily" {
		t.Fatalf("unexpected route: %#v", r)
	}
}
