package route

import (
	"context"
	"errors"

	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

var (
	ErrInvalidReference = errors.New("route provider or model does not exist")
	ErrModelDisabled    = errors.New("route model is disabled")
)

type Service struct {
	repo      route.Repository
	providers provider.Repository
	models    model.Repository
}

func NewService(repo route.Repository, dependencies ...any) *Service {
	s := &Service{repo: repo}
	for _, dependency := range dependencies {
		switch item := dependency.(type) {
		case provider.Repository:
			s.providers = item
		case model.Repository:
			s.models = item
		}
	}
	return s
}

func (s *Service) List(ctx context.Context) ([]route.Route, error) { return s.repo.List(ctx) }

func (s *Service) Create(ctx context.Context, id, name, providerID, modelID string) (route.Route, error) {
	item, err := route.New(id, name, providerID, modelID)
	if err != nil {
		return route.Route{}, err
	}
	if err := s.validateReferences(ctx, item); err != nil {
		return route.Route{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return route.Route{}, err
	}
	return item, nil
}

func (s *Service) Save(ctx context.Context, item route.Route) error {
	if err := item.Validate(); err != nil {
		return err
	}
	if err := s.validateReferences(ctx, item); err != nil {
		return err
	}
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

func (s *Service) validateReferences(ctx context.Context, item route.Route) error {
	if s.providers != nil {
		if _, err := s.providers.Get(ctx, item.ProviderID); err != nil {
			return ErrInvalidReference
		}
	}
	if s.models != nil {
		m, err := s.models.Get(ctx, item.ProviderID, item.ModelID)
		if err != nil {
			return ErrInvalidReference
		}
		if !m.Enabled {
			return ErrModelDisabled
		}
	}
	return nil
}
