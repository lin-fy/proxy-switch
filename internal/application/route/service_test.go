package route

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	modelapp "codex-provider-hub/internal/application/model"
	providerapp "codex-provider-hub/internal/application/provider"
	domainmodel "codex-provider-hub/internal/domain/model"
	domainprovider "codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/infrastructure/config"
)

func TestRouteReferencesAndDeletionGuards(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	providerService := providerapp.NewService(providers, models, routes)
	modelService := modelapp.NewService(models, routes)
	routeService := NewService(routes, providers, models)
	ctx := context.Background()

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
	if _, err := routeService.Create(ctx, "r", "Route", p.ID, m.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := routeService.Create(ctx, "bad", "Bad", p.ID, "missing"); !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("invalid route error = %v", err)
	}
	if err := modelService.Delete(ctx, p.ID, m.ID); !errors.Is(err, modelapp.ErrReferenced) {
		t.Fatalf("model deletion error = %v", err)
	}
	if err := providerService.Delete(ctx, p.ID); !errors.Is(err, providerapp.ErrReferenced) {
		t.Fatalf("provider deletion error = %v", err)
	}
}
