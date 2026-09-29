package profile

import (
	"context"

	"codex-provider-hub/internal/domain/profile"
)

type Service struct{ repo profile.Repository }

func NewService(repo profile.Repository) *Service { return &Service{repo: repo} }

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
	return s.repo.Save(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
