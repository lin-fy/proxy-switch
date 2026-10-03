package model

import (
	"context"
	"errors"
	"strings"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
)

var (
	ErrReferenced         = errors.New("model is referenced by a route")
	ErrAlreadyExists      = errors.New("model already exists")
	ErrInvalidProviderRef = errors.New("model provider reference does not exist")
)

type Service struct {
	repo      model.Repository
	routes    route.Repository
	providers provider.Repository
}

func NewService(repo model.Repository, dependencies ...any) *Service {
	service := &Service{repo: repo}
	for _, dependency := range dependencies {
		switch item := dependency.(type) {
		case route.Repository:
			service.routes = item
		case provider.Repository:
			service.providers = item
		}
	}
	return service
}

func (s *Service) List(ctx context.Context) ([]model.Model, error) { return s.repo.List(ctx) }

func (s *Service) ListByProvider(ctx context.Context, providerID string) ([]model.Model, error) {
	return s.repo.ListByProvider(ctx, providerID)
}

func (s *Service) Get(ctx context.Context, providerID, id string) (model.Model, error) {
	return s.repo.Get(ctx, providerID, id)
}

func (s *Service) Create(ctx context.Context, providerID, id, name string) (model.Model, error) {
	item, err := model.New(providerID, id, name)
	if err != nil {
		return model.Model{}, err
	}
	if err := s.validateProvider(ctx, item.ProviderID); err != nil {
		return model.Model{}, err
	}
	if _, err := s.repo.Get(ctx, item.ProviderID, item.ID); err == nil {
		return model.Model{}, ErrAlreadyExists
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return model.Model{}, err
	}
	return item, nil
}

func (s *Service) Save(ctx context.Context, item model.Model) error {
	if item.ProviderID == "" || item.ID == "" {
		return model.ErrInvalidID
	}
	if err := s.validateProvider(ctx, item.ProviderID); err != nil {
		return err
	}
	return s.repo.Save(ctx, item)
}

func (s *Service) validateProvider(ctx context.Context, providerID string) error {
	if s.providers == nil {
		return nil
	}
	if _, err := s.providers.Get(ctx, providerID); err != nil {
		return ErrInvalidProviderRef
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, providerID, id string) error {
	if s.routes != nil {
		items, err := s.routes.List(ctx)
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.ProviderID == providerID && item.ModelID == id {
				return ErrReferenced
			}
		}
	}
	return s.repo.Delete(ctx, providerID, id)
}

func (s *Service) Sync(ctx context.Context, providerID string, remote []model.Model) ([]model.Model, error) {
	existing, err := s.repo.ListByProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]model.Model, len(existing))
	for _, item := range existing {
		byID[item.ID] = item
	}
	for _, item := range remote {
		item.ProviderID = providerID
		item.ID = strings.TrimSpace(item.ID)
		item.Name = strings.TrimSpace(item.Name)
		if item.ID == "" {
			continue
		}
		if old, ok := byID[item.ID]; ok {
			item.Enabled = old.Enabled
		} else {
			item.Enabled = true
		}
		if item.Name == "" {
			item.Name = item.ID
		}
		if err := s.repo.Save(ctx, item); err != nil {
			return nil, err
		}
	}
	return s.repo.ListByProvider(ctx, providerID)
}
