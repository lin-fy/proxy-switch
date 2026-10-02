package credential

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

func IsCredentialReference(reference string) bool {
	return strings.HasPrefix(strings.TrimSpace(reference), "credential:")
}

func EnvName(reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", nil
	}
	if IsCredentialReference(reference) {
		target := strings.TrimSpace(strings.TrimPrefix(reference, "credential:"))
		if target == "" {
			return "", fmt.Errorf("Windows Credential Manager target is empty")
		}
		sum := sha256.Sum256([]byte(target))
		return "CODEX_PROVIDER_HUB_CRED_" + fmt.Sprintf("%x", sum[:]), nil
	}
	name := strings.TrimSpace(strings.TrimPrefix(reference, "env:"))
	if name == "" {
		return "", fmt.Errorf("credential environment variable name is empty")
	}
	if strings.ContainsAny(name, "=\x00") {
		return "", fmt.Errorf("credential environment variable name is invalid")
	}
	return name, nil
}

func Resolve(reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", nil
	}
	if strings.HasPrefix(reference, "credential:") {
		return resolveGeneric(strings.TrimPrefix(reference, "credential:"))
	}
	name, err := EnvName(reference)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("credential environment variable %q is not set", name)
	}
	return value, nil
}
