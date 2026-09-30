package credential

import "testing"

func TestResolveEnvironmentReference(t *testing.T) {
	t.Setenv("PROVIDER_HUB_TEST_KEY", "secret")
	for _, reference := range []string{"PROVIDER_HUB_TEST_KEY", "env:PROVIDER_HUB_TEST_KEY"} {
		value, err := Resolve(reference)
		if err != nil || value != "secret" {
			t.Fatalf("Resolve(%q) = %q, %v", reference, value, err)
		}
	}
}

func TestResolveMissingEnvironmentReference(t *testing.T) {
	t.Setenv("PROVIDER_HUB_MISSING_KEY", "")
	if _, err := Resolve("PROVIDER_HUB_MISSING_KEY"); err == nil {
		t.Fatal("expected missing environment variable error")
	}
}
