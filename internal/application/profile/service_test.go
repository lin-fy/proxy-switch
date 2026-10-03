package profile

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	domainprofile "proxy-switch/internal/domain/profile"
	domainroute "proxy-switch/internal/domain/route"
	"proxy-switch/internal/infrastructure/config"
)

func TestSaveRejectsMissingRouteReference(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	profiles := config.NewProfileRepository(store)
	routes := config.NewRouteRepository(store)
	service := NewService(profiles, routes)
	item, err := domainprofile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	item.RouteID = "missing"
	if err := service.Save(context.Background(), item); !errors.Is(err, ErrInvalidRouteReference) {
		t.Fatalf("missing route error = %v", err)
	}
	if _, err := profiles.Get(context.Background(), item.ID); !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("invalid profile was persisted: %v", err)
	}
}

func TestSaveAcceptsExistingRouteReference(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	profiles := config.NewProfileRepository(store)
	routes := config.NewRouteRepository(store)
	service := NewService(profiles, routes)
	r, err := domainroute.New("route", "Route", "provider", "model")
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.Save(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	item, err := domainprofile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	item.RouteID = r.ID
	if err := service.Save(context.Background(), item); err != nil {
		t.Fatal(err)
	}
}

func TestSaveRejectsSharedProfilePaths(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	profiles := config.NewProfileRepository(store)
	service := NewService(profiles)
	first, err := domainprofile.New("first", "First")
	if err != nil {
		t.Fatal(err)
	}
	first.ConfigPath = "profiles/shared/config.toml"
	first.ModelCatalogPath = "profiles/shared/models.json"
	if err := profiles.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	second, err := domainprofile.New("second", "Second")
	if err != nil {
		t.Fatal(err)
	}
	second.ConfigPath = "profiles/shared/config.toml"
	if err := service.Save(ctx, second); !errors.Is(err, ErrPathConflict) {
		t.Fatalf("shared config path error = %v", err)
	}
	second.ConfigPath = "profiles/other/config.toml"
	second.ModelCatalogPath = "profiles/shared/models.json"
	if err := service.Save(ctx, second); !errors.Is(err, ErrPathConflict) {
		t.Fatalf("shared catalog path error = %v", err)
	}
}
