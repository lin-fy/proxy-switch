package model

import (
	"context"

	"codex-provider-hub/internal/domain/model"
)

type Service struct{ repo model.Repository }

func NewService(repo model.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]model.Model, error) { return s.repo.List(ctx) }

func (s *Service) ListByProvider(ctx context.Context, providerID string) ([]model.Model, error) {
	return s.repo.ListByProvider(ctx, providerID)
}

func (s *Service) Create(ctx context.Context, providerID, id, name string) (model.Model, error) {
	item, err := model.New(providerID, id, name)
	if err != nil {
		return model.Model{}, err
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
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, providerID, id string) error {
	return s.repo.Delete(ctx, providerID, id)
}
