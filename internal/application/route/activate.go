package route

import (
	"context"
	"fmt"

	"codex-provider-hub/internal/application/ports"
	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

type Activator struct {
	routes    route.Repository
	providers provider.Repository
	models    model.Repository
	profiles  profile.Repository
	adapters  ports.PlatformAdapterFactory
}

func NewActivator(
	routes route.Repository,
	providers provider.Repository,
	models model.Repository,
	profiles profile.Repository,
	adapters ports.PlatformAdapterFactory,
) *Activator {
	return &Activator{
		routes: routes, providers: providers, models: models,
		profiles: profiles, adapters: adapters,
	}
}

func (a *Activator) Activate(ctx context.Context, routeID, profileID string) error {
	r, err := a.routes.Get(ctx, routeID)
	if err != nil {
		return fmt.Errorf("get route: %w", err)
	}
	p, err := a.providers.Get(ctx, r.ProviderID)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}
	m, err := a.models.Get(ctx, r.ProviderID, r.ModelID)
	if err != nil {
		return fmt.Errorf("get model: %w", err)
	}
	pr, err := a.profiles.Get(ctx, profileID)
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}
	adapter, err := a.adapters.Create(r.PlatformID)
	if err != nil {
		return fmt.Errorf("create platform adapter: %w", err)
	}
	if err := adapter.Validate(ctx, r, p, m); err != nil {
		return fmt.Errorf("validate route: %w", err)
	}
	if err := adapter.Prepare(ctx, r, p, m, pr); err != nil {
		return fmt.Errorf("prepare route: %w", err)
	}
	if r.RestartOnActivate {
		if err := adapter.Launch(ctx, r); err != nil {
			return fmt.Errorf("launch platform: %w", err)
		}
	}
	return nil
}
