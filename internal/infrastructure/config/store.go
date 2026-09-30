package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

var ErrNotFound = errors.New("entity not found")

type snapshot struct {
	Providers []provider.Provider `json:"providers"`
	Models    []model.Model       `json:"models"`
	Routes    []route.Route       `json:"routes"`
	Profiles  []profile.Profile   `json:"profiles"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store { return &Store{path: path} }

func (s *Store) read() (snapshot, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return snapshot{}, nil
	}
	if err != nil {
		return snapshot{}, fmt.Errorf("read state: %w", err)
	}
	var state snapshot
	if err := json.Unmarshal(b, &state); err != nil {
		return snapshot{}, fmt.Errorf("decode state: %w", err)
	}
	return state, nil
}

func (s *Store) view(ctx context.Context, fn func(snapshot) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read()
	if err != nil {
		return err
	}
	return fn(state)
}

func (s *Store) update(ctx context.Context, fn func(*snapshot) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read()
	if err != nil {
		return err
	}
	if err := fn(&state); err != nil {
		return err
	}
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("create state temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("secure state temp file: %w", err)
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("write state temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close state temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	return nil
}

type ProviderRepository struct{ store *Store }

func NewProviderRepository(store *Store) *ProviderRepository {
	return &ProviderRepository{store: store}
}

func (r *ProviderRepository) List(ctx context.Context) ([]provider.Provider, error) {
	var result []provider.Provider
	err := r.store.view(ctx, func(s snapshot) error { result = append(result, s.Providers...); return nil })
	return result, err
}

func (r *ProviderRepository) Get(ctx context.Context, id string) (provider.Provider, error) {
	var result provider.Provider
	err := r.store.view(ctx, func(s snapshot) error {
		for _, item := range s.Providers {
			if item.ID == id {
				result = item
				return nil
			}
		}
		return ErrNotFound
	})
	return result, err
}

func (r *ProviderRepository) Save(ctx context.Context, item provider.Provider) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Providers {
			if s.Providers[i].ID == item.ID {
				s.Providers[i] = item
				return nil
			}
		}
		s.Providers = append(s.Providers, item)
		return nil
	})
}

func (r *ProviderRepository) Delete(ctx context.Context, id string) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Providers {
			if s.Providers[i].ID == id {
				s.Providers = append(s.Providers[:i], s.Providers[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}

type ModelRepository struct{ store *Store }

func NewModelRepository(store *Store) *ModelRepository { return &ModelRepository{store: store} }

func (r *ModelRepository) List(ctx context.Context) ([]model.Model, error) {
	var result []model.Model
	err := r.store.view(ctx, func(s snapshot) error { result = append(result, s.Models...); return nil })
	return result, err
}

func (r *ModelRepository) ListByProvider(ctx context.Context, providerID string) ([]model.Model, error) {
	var result []model.Model
	err := r.store.view(ctx, func(s snapshot) error {
		for _, item := range s.Models {
			if item.ProviderID == providerID {
				result = append(result, item)
			}
		}
		return nil
	})
	return result, err
}

func (r *ModelRepository) Get(ctx context.Context, providerID, id string) (model.Model, error) {
	var result model.Model
	err := r.store.view(ctx, func(s snapshot) error {
		for _, item := range s.Models {
			if item.ProviderID == providerID && item.ID == id {
				result = item
				return nil
			}
		}
		return ErrNotFound
	})
	return result, err
}

func (r *ModelRepository) Save(ctx context.Context, item model.Model) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Models {
			if s.Models[i].ProviderID == item.ProviderID && s.Models[i].ID == item.ID {
				s.Models[i] = item
				return nil
			}
		}
		s.Models = append(s.Models, item)
		return nil
	})
}

func (r *ModelRepository) Delete(ctx context.Context, providerID, id string) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Models {
			if s.Models[i].ProviderID == providerID && s.Models[i].ID == id {
				s.Models = append(s.Models[:i], s.Models[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}

type RouteRepository struct{ store *Store }

func NewRouteRepository(store *Store) *RouteRepository { return &RouteRepository{store: store} }

func (r *RouteRepository) List(ctx context.Context) ([]route.Route, error) {
	var result []route.Route
	err := r.store.view(ctx, func(s snapshot) error { result = append(result, s.Routes...); return nil })
	return result, err
}

func (r *RouteRepository) Get(ctx context.Context, id string) (route.Route, error) {
	var result route.Route
	err := r.store.view(ctx, func(s snapshot) error {
		for _, item := range s.Routes {
			if item.ID == id {
				result = item
				return nil
			}
		}
		return ErrNotFound
	})
	return result, err
}

func (r *RouteRepository) Save(ctx context.Context, item route.Route) error {
	return r.store.update(ctx, func(s *snapshot) error {
		if item.Default {
			for i := range s.Routes {
				s.Routes[i].Default = false
			}
		}
		for i := range s.Routes {
			if s.Routes[i].ID == item.ID {
				s.Routes[i] = item
				return nil
			}
		}
		s.Routes = append(s.Routes, item)
		return nil
	})
}

func (r *RouteRepository) Delete(ctx context.Context, id string) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Routes {
			if s.Routes[i].ID == id {
				s.Routes = append(s.Routes[:i], s.Routes[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}

type ProfileRepository struct{ store *Store }

func NewProfileRepository(store *Store) *ProfileRepository { return &ProfileRepository{store: store} }

func (r *ProfileRepository) List(ctx context.Context) ([]profile.Profile, error) {
	var result []profile.Profile
	err := r.store.view(ctx, func(s snapshot) error { result = append(result, s.Profiles...); return nil })
	return result, err
}

func (r *ProfileRepository) Get(ctx context.Context, id string) (profile.Profile, error) {
	var result profile.Profile
	err := r.store.view(ctx, func(s snapshot) error {
		for _, item := range s.Profiles {
			if item.ID == id {
				result = item
				return nil
			}
		}
		return ErrNotFound
	})
	return result, err
}

func (r *ProfileRepository) Save(ctx context.Context, item profile.Profile) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Profiles {
			if s.Profiles[i].ID == item.ID {
				s.Profiles[i] = item
				return nil
			}
		}
		s.Profiles = append(s.Profiles, item)
		return nil
	})
}

func (r *ProfileRepository) Delete(ctx context.Context, id string) error {
	return r.store.update(ctx, func(s *snapshot) error {
		for i := range s.Profiles {
			if s.Profiles[i].ID == id {
				s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}
