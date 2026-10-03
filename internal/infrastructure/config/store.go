package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
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

// ImportWorkspace appends a validated batch in one state.json transaction.
// It refuses any existing ID collision and never touches platform auth files.
func (s *Store) ImportWorkspace(ctx context.Context, providers []provider.Provider, models []model.Model, routes []route.Route, profiles []profile.Profile) error {
	return s.update(ctx, func(state *snapshot) error {
		providerIDs := make(map[string]bool, len(state.Providers)+len(providers))
		for _, item := range state.Providers {
			providerIDs[item.ID] = true
		}
		for _, item := range providers {
			if item.ID == "" || providerIDs[item.ID] {
				return fmt.Errorf("provider import conflict: %s", item.ID)
			}
			providerIDs[item.ID] = true
		}
		modelIDs := make(map[string]bool, len(state.Models)+len(models))
		modelEnabled := make(map[string]bool, len(state.Models)+len(models))
		for _, item := range state.Models {
			key := item.ProviderID + "\x00" + item.ID
			modelIDs[key] = true
			modelEnabled[key] = item.Enabled
		}
		for _, item := range models {
			key := item.ProviderID + "\x00" + item.ID
			if item.ProviderID == "" || item.ID == "" || modelIDs[key] || !providerIDs[item.ProviderID] {
				return fmt.Errorf("model import conflict or missing provider: %s/%s", item.ProviderID, item.ID)
			}
			modelIDs[key] = true
			modelEnabled[key] = item.Enabled
		}
		routeIDs := make(map[string]bool, len(state.Routes)+len(routes))
		for _, item := range state.Routes {
			routeIDs[item.ID] = true
		}
		defaultRoute := false
		for _, item := range state.Routes {
			defaultRoute = defaultRoute || item.Default
		}
		importedDefault := false
		for _, item := range routes {
			if item.ID == "" || routeIDs[item.ID] || !providerIDs[item.ProviderID] || !modelIDs[item.ProviderID+"\x00"+item.ModelID] || !modelEnabled[item.ProviderID+"\x00"+item.ModelID] {
				return fmt.Errorf("route import conflict or missing reference: %s", item.ID)
			}
			if item.Default && (defaultRoute || importedDefault) {
				return fmt.Errorf("route import contains conflicting default route: %s", item.ID)
			}
			importedDefault = importedDefault || item.Default
			routeIDs[item.ID] = true
		}
		profileIDs := make(map[string]bool, len(state.Profiles)+len(profiles))
		for _, item := range state.Profiles {
			profileIDs[item.ID] = true
		}
		configPaths := make(map[string]bool)
		catalogPaths := make(map[string]bool)
		for _, item := range state.Profiles {
			if item.ConfigPath != "" {
				configPaths[filepath.Clean(item.ConfigPath)] = true
			}
			if item.ModelCatalogPath != "" {
				catalogPaths[filepath.Clean(item.ModelCatalogPath)] = true
			}
		}
		for _, item := range profiles {
			configPath := filepath.Clean(item.ConfigPath)
			catalogPath := filepath.Clean(item.ModelCatalogPath)
			if item.ID == "" || profileIDs[item.ID] || (item.RouteID != "" && !routeIDs[item.RouteID]) || (item.ConfigPath != "" && configPaths[configPath]) || (item.ModelCatalogPath != "" && catalogPaths[catalogPath]) {
				return fmt.Errorf("profile import conflict or missing route: %s", item.ID)
			}
			profileIDs[item.ID] = true
			if item.ConfigPath != "" {
				configPaths[configPath] = true
			}
			if item.ModelCatalogPath != "" {
				catalogPaths[catalogPath] = true
			}
		}
		state.Providers = append(state.Providers, providers...)
		state.Models = append(state.Models, models...)
		state.Routes = append(state.Routes, routes...)
		state.Profiles = append(state.Profiles, profiles...)
		return nil
	})
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
				wasDefault := s.Routes[i].Default
				s.Routes = append(s.Routes[:i], s.Routes[i+1:]...)
				if wasDefault && len(s.Routes) > 0 {
					s.Routes[0].Default = true
				}
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
