package profile

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID   = errors.New("profile id is required")
	ErrInvalidName = errors.New("profile name is required")
)

type Profile struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	RouteID          string `json:"route_id,omitempty"`
	ConfigPath       string `json:"config_path,omitempty"`
	ModelCatalogPath string `json:"model_catalog_path,omitempty"`
}

func New(id, name string) (Profile, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		return Profile{}, ErrInvalidID
	}
	if name == "" {
		return Profile{}, ErrInvalidName
	}
	return Profile{ID: id, Name: name}, nil
}
