package credential

import (
	"strings"
	"testing"
)

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

func TestEnvNameNormalizesReferences(t *testing.T) {
	if got, err := EnvName("env:PROVIDER_HUB_TEST_KEY"); err != nil || got != "PROVIDER_HUB_TEST_KEY" {
		t.Fatalf("environment name = %q, %v", got, err)
	}
	first, err := EnvName("credential:cpa")
	if err != nil {
		t.Fatal(err)
	}
	second, err := EnvName("credential:cpa")
	if err != nil || first != second || !strings.HasPrefix(first, "CODEX_PROVIDER_HUB_CRED_") {
		t.Fatalf("credential names = %q/%q, %v", first, second, err)
	}
	if _, err := EnvName("env:"); err == nil {
		t.Fatal("expected empty environment name error")
	}
}
