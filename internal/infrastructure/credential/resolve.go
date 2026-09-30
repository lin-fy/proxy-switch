package credential

import (
	"fmt"
	"os"
	"strings"
)

func Resolve(reference string) (string, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return "", nil
	}
	if strings.HasPrefix(reference, "credential:") {
		return resolveGeneric(strings.TrimPrefix(reference, "credential:"))
	}
	name := strings.TrimPrefix(reference, "env:")
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("credential environment variable %q is not set", name)
	}
	return value, nil
}
