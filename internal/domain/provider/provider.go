package provider

import (
	"errors"
	"net/url"
	"strings"
)

const ProtocolResponses = "responses"

var (
	ErrInvalidID       = errors.New("provider id is required")
	ErrInvalidName     = errors.New("provider name is required")
	ErrInvalidBaseURL  = errors.New("provider base URL is invalid")
	ErrInvalidProtocol = errors.New("provider protocol must be responses")
	ErrInvalidAuthRef  = errors.New("provider auth reference must be an environment or credential reference")
)

type Provider struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	Protocol    string            `json:"protocol"`
	AuthRef     string            `json:"auth_ref,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
}

func New(id, name, baseURL, authRef string) (Provider, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	baseURL = strings.TrimSpace(baseURL)
	if id == "" {
		return Provider{}, ErrInvalidID
	}
	if name == "" {
		return Provider{}, ErrInvalidName
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return Provider{}, ErrInvalidBaseURL
	}
	authRef = strings.TrimSpace(authRef)
	if err := ValidateAuthRef(authRef); err != nil {
		return Provider{}, err
	}
	return Provider{
		ID:       id,
		Name:     name,
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Protocol: ProtocolResponses,
		AuthRef:  authRef,
	}, nil
}

func (p Provider) Validate() error {
	if p.ID == "" {
		return ErrInvalidID
	}
	if p.Name == "" {
		return ErrInvalidName
	}
	if p.Protocol != ProtocolResponses {
		return ErrInvalidProtocol
	}
	parsed, err := url.ParseRequestURI(p.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return ErrInvalidBaseURL
	}
	return ValidateAuthRef(p.AuthRef)
}

// ValidateAuthRef keeps secrets out of state.json and the Wails DTO. Only a
// reference to an environment variable or Windows Credential Manager target
// is accepted; raw API keys are rejected before persistence.
func ValidateAuthRef(reference string) error {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return nil
	}
	if strings.HasPrefix(reference, "credential:") {
		if strings.TrimSpace(strings.TrimPrefix(reference, "credential:")) == "" {
			return ErrInvalidAuthRef
		}
		return nil
	}
	name := strings.TrimPrefix(reference, "env:")
	if name == "" || !validEnvName(name) {
		return ErrInvalidAuthRef
	}
	return nil
}

func validEnvName(name string) bool {
	for i, r := range name {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (i > 0 && r >= '0' && r <= '9') || (i > 0 && r == '_') {
			continue
		}
		return false
	}
	return true
}
