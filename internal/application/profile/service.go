package profile

import (
	"context"
	"errors"
	"path/filepath"

	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/route"
)

var (
	ErrInvalidRouteReference = errors.New("profile route reference does not exist")
	ErrPathConflict          = errors.New("profile config path is already used")
)

type Service struct {
	repo   profile.Repository
	routes route.Repository
}

func NewService(repo profile.Repository, dependencies ...route.Repository) *Service {
	service := &Service{repo: repo}
	if len(dependencies) > 0 {
		service.routes = dependencies[0]
	}
	return service
}

func (s *Service) List(ctx context.Context) ([]profile.Profile, error) { return s.repo.List(ctx) }

func (s *Service) Get(ctx context.Context, id string) (profile.Profile, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, id, name string) (profile.Profile, error) {
	item, err := profile.New(id, name)
	if err != nil {
		return profile.Profile{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return profile.Profile{}, err
	}
	return item, nil
}

func (s *Service) Save(ctx context.Context, item profile.Profile) error {
	if item.ID == "" || item.Name == "" {
		return profile.ErrInvalidName
	}
	if item.RouteID != "" && s.routes != nil {
		if _, err := s.routes.Get(ctx, item.RouteID); err != nil {
			return ErrInvalidRouteReference
		}
	}
	items, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	for _, existing := range items {
		if existing.ID == item.ID {
			continue
		}
		if item.ConfigPath != "" && existing.ConfigPath != "" && filepath.Clean(existing.ConfigPath) == filepath.Clean(item.ConfigPath) {
			return ErrPathConflict
		}
		if item.ModelCatalogPath != "" && existing.ModelCatalogPath != "" && filepath.Clean(existing.ModelCatalogPath) == filepath.Clean(item.ModelCatalogPath) {
			return ErrPathConflict
		}
	}
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
