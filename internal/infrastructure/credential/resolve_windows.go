//go:build windows

package credential

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const credTypeGeneric = 1

var (
	advapi32 = windows.NewLazySystemDLL("advapi32.dll")
	credRead = advapi32.NewProc("CredReadW")
	credFree = advapi32.NewProc("CredFree")
)

type nativeCredential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

func resolveGeneric(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("Windows Credential Manager target is empty")
	}
	targetName, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return "", fmt.Errorf("encode credential target: %w", err)
	}
	var credentialPtr uintptr
	result, _, callErr := credRead.Call(
		uintptr(unsafe.Pointer(targetName)),
		credTypeGeneric,
		0,
		uintptr(unsafe.Pointer(&credentialPtr)),
	)
	if result == 0 {
		return "", fmt.Errorf("read Windows credential %q: %w", target, callErr)
	}
	defer credFree.Call(credentialPtr)
	credential := (*nativeCredential)(unsafe.Pointer(credentialPtr))
	if credential.CredentialBlobSize == 0 || credential.CredentialBlob == nil {
		return "", fmt.Errorf("Windows credential %q has an empty secret", target)
	}
	secret := unsafe.Slice(credential.CredentialBlob, credential.CredentialBlobSize)
	return string(secret), nil
}
