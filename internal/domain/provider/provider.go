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
	return Provider{
		ID:       id,
		Name:     name,
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Protocol: ProtocolResponses,
		AuthRef:  strings.TrimSpace(authRef),
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
	return nil
}
