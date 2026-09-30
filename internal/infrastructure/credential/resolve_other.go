//go:build !windows

package credential

import "fmt"

func resolveGeneric(target string) (string, error) {
	return "", fmt.Errorf("Windows Credential Manager is unavailable for target %q", target)
}
