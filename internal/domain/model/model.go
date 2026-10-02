package model

import (
	"errors"
	"strings"
)

var (
	ErrInvalidID         = errors.New("model id is required")
	ErrInvalidProviderID = errors.New("model provider id is required")
)

type Model struct {
	ProviderID string `json:"provider_id"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
}

func New(providerID, id, name string) (Model, error) {
	providerID = strings.TrimSpace(providerID)
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if providerID == "" {
		return Model{}, ErrInvalidProviderID
	}
	if id == "" {
		return Model{}, ErrInvalidID
	}
	if name == "" {
		name = id
	}
	return Model{ProviderID: providerID, ID: id, Name: name, Enabled: true}, nil
}
