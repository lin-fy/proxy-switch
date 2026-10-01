package provider

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	domainmodel "codex-provider-hub/internal/domain/model"
	domainprovider "codex-provider-hub/internal/domain/provider"
	domainroute "codex-provider-hub/internal/domain/route"
	"codex-provider-hub/internal/infrastructure/config"
)

func newProviderService(t *testing.T) (*Service, *config.ProviderRepository, *config.ModelRepository, *config.RouteRepository) {
	t.Helper()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	return NewService(providers, models, routes), providers, models, routes
}

func TestProviderDeleteBlockedByModels(t *testing.T) {
	ctx := context.Background()
	service, providers, models, _ := newProviderService(t)

	p, err := domainprovider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	m, err := domainmodel.New(p.ID, "m", "Model")
	if err != nil {
		t.Fatal(err)
	}
	if err := models.Save(ctx, m); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, p.ID); !errors.Is(err, ErrReferenced) {
		t.Fatalf("provider delete with models = %v", err)
	}
	if _, err := providers.Get(ctx, p.ID); err != nil {
		t.Fatalf("blocked delete should keep provider: %v", err)
	}
}

func TestProviderDeleteBlockedByRoutes(t *testing.T) {
	ctx := context.Background()
	service, providers, _, routes := newProviderService(t)

	p, err := domainprovider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	r, err := domainroute.New("r", "Route", p.ID, "m")
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.Save(ctx, r); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, p.ID); !errors.Is(err, ErrReferenced) {
		t.Fatalf("provider delete with routes = %v", err)
	}
	if _, err := providers.Get(ctx, p.ID); err != nil {
		t.Fatalf("blocked delete should keep provider: %v", err)
	}
}

func TestProviderDeleteSucceedsWhenUnreferenced(t *testing.T) {
	ctx := context.Background()
	service, providers, _, _ := newProviderService(t)

	p, err := domainprovider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(ctx, p); err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(ctx, p.ID); err != nil {
		t.Fatalf("unreferenced provider delete = %v", err)
	}
	if _, err := providers.Get(ctx, p.ID); !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("provider after delete = %v", err)
	}
}
