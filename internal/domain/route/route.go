package route

import (
	"errors"
	"strings"
)

const PlatformCodex = "codex"

var (
	ErrInvalidID         = errors.New("route id is required")
	ErrInvalidPlatformID = errors.New("route platform id must be codex")
	ErrInvalidProviderID = errors.New("route provider id is required")
	ErrInvalidModelID    = errors.New("route model id is required")
	ErrInvalidPriority   = errors.New("route priority must be non-negative")
)

type Route struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PlatformID        string `json:"platform_id"`
	ProviderID        string `json:"provider_id"`
	ModelID           string `json:"model_id"`
	Priority          int    `json:"priority"`
	RestartOnActivate bool   `json:"restart_on_activate"`
	Default           bool   `json:"default"`
}

func New(id, name, providerID, modelID string) (Route, error) {
	r := Route{
		ID:         strings.TrimSpace(id),
		Name:       strings.TrimSpace(name),
		PlatformID: PlatformCodex,
		ProviderID: strings.TrimSpace(providerID),
		ModelID:    strings.TrimSpace(modelID),
	}
	if r.Name == "" {
		r.Name = r.ID
	}
	return r, r.Validate()
}

func (r Route) Validate() error {
	if r.ID == "" {
		return ErrInvalidID
	}
	if r.PlatformID != PlatformCodex {
		return ErrInvalidPlatformID
	}
	if r.ProviderID == "" {
		return ErrInvalidProviderID
	}
	if r.ModelID == "" {
		return ErrInvalidModelID
	}
	if r.Priority < 0 {
		return ErrInvalidPriority
	}
	return nil
}
