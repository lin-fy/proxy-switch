package route

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	modelapp "proxy-switch/internal/application/model"
	providerapp "proxy-switch/internal/application/provider"
	domainmodel "proxy-switch/internal/domain/model"
	domainprofile "proxy-switch/internal/domain/profile"
	domainprovider "proxy-switch/internal/domain/provider"
	domainroute "proxy-switch/internal/domain/route"
	"proxy-switch/internal/infrastructure/config"
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

func TestRouteRejectsDisabledModel(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
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
	m.Enabled = false
	if err := models.Save(ctx, m); err != nil {
		t.Fatal(err)
	}

	if _, err := routeService.Create(ctx, "r", "Route", p.ID, m.ID); !errors.Is(err, ErrModelDisabled) {
		t.Fatalf("disabled model create error = %v", err)
	}
	if err := routeService.Save(ctx, domainroute.Route{ID: "r", Name: "Route", PlatformID: domainroute.PlatformCodex, ProviderID: p.ID, ModelID: m.ID}); !errors.Is(err, ErrModelDisabled) {
		t.Fatalf("disabled model save error = %v", err)
	}
	items, err := routes.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("disabled model routes = %#v", items)
	}
}

func TestRouteCreateRejectsExistingID(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	service := NewService(routes, providers, models)
	p, err := domainprovider.New("p", "Provider", "https://example.test/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := providers.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err := models.Save(ctx, domainmodel.Model{ProviderID: "p", ID: "m", Name: "Model", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, "r", "Route", "p", "m"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, "r", "Replacement", "p", "m"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("duplicate route error = %v", err)
	}
}

func TestRouteDeleteBlockedByProfileReference(t *testing.T) {
	ctx := context.Background()
	store := config.NewStore(filepath.Join(t.TempDir(), "state.json"))
	routes := config.NewRouteRepository(store)
	profiles := config.NewProfileRepository(store)
	service := NewService(routes, profiles)
	item, err := domainroute.New("r", "Route", "p", "m")
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.Save(ctx, item); err != nil {
		t.Fatal(err)
	}
	profileItem, err := domainprofile.New("work", "Work")
	if err != nil {
		t.Fatal(err)
	}
	profileItem.RouteID = item.ID
	if err := profiles.Save(ctx, profileItem); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(ctx, item.ID); !errors.Is(err, ErrReferenced) {
		t.Fatalf("route delete with profile reference = %v", err)
	}
	if _, err := routes.Get(ctx, item.ID); err != nil {
		t.Fatalf("referenced route was deleted: %v", err)
	}
}
