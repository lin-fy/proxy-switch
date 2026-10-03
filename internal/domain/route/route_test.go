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

func TestPriorityMustBeNonNegative(t *testing.T) {
	r, err := New("backup", "Backup", "cpa", "gpt-5")
	if err != nil {
		t.Fatal(err)
	}
	if r.Priority != 0 {
		t.Fatalf("default priority = %d", r.Priority)
	}
	r.Priority = -1
	if err := r.Validate(); err != ErrInvalidPriority {
		t.Fatalf("negative priority error = %v", err)
	}
}
