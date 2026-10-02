package provider

import (
	"context"
	"errors"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
)

var ErrReferenced = errors.New("provider is referenced by models or routes")

type Service struct {
	repo   provider.Repository
	models model.Repository
	routes route.Repository
}

func NewService(repo provider.Repository, dependencies ...any) *Service {
	s := &Service{repo: repo}
	for _, dependency := range dependencies {
		switch item := dependency.(type) {
		case model.Repository:
			s.models = item
		case route.Repository:
			s.routes = item
		}
	}
	return s
}

func (s *Service) List(ctx context.Context) ([]provider.Provider, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (provider.Provider, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, id, name, baseURL, authRef string) (provider.Provider, error) {
	item, err := provider.New(id, name, baseURL, authRef)
	if err != nil {
		return provider.Provider{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return provider.Provider{}, err
	}
	return item, nil
}

func (s *Service) Save(ctx context.Context, item provider.Provider) error {
	if err := item.Validate(); err != nil {
		return err
	}
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s.models != nil {
		items, err := s.models.ListByProvider(ctx, id)
		if err != nil {
			return err
		}
		if len(items) > 0 {
			return ErrReferenced
		}
	}
	if s.routes != nil {
		items, err := s.routes.List(ctx)
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.ProviderID == id {
				return ErrReferenced
			}
		}
	}
	return s.repo.Delete(ctx, id)
}
