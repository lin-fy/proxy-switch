package provider

import (
	"context"

	"codex-provider-hub/internal/domain/provider"
)

type Service struct{ repo provider.Repository }

func NewService(repo provider.Repository) *Service { return &Service{repo: repo} }

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

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
