package route

import (
	"context"

	"codex-provider-hub/internal/domain/route"
)

type Service struct{ repo route.Repository }

func NewService(repo route.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]route.Route, error) { return s.repo.List(ctx) }

func (s *Service) Create(ctx context.Context, id, name, providerID, modelID string) (route.Route, error) {
	item, err := route.New(id, name, providerID, modelID)
	if err != nil {
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
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
