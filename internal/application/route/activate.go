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
	tester    ports.ProviderTester
}

func NewActivator(
	routes route.Repository,
	providers provider.Repository,
	models model.Repository,
	profiles profile.Repository,
	adapters ports.PlatformAdapterFactory,
	testers ...ports.ProviderTester,
) *Activator {
	var tester ports.ProviderTester
	if len(testers) > 0 {
		tester = testers[0]
	}
	return &Activator{
		routes: routes, providers: providers, models: models,
		profiles: profiles, adapters: adapters, tester: tester,
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
	if !m.Enabled {
		return fmt.Errorf("activate route: %w", ErrModelDisabled)
	}
	models, err := a.models.ListByProvider(ctx, r.ProviderID)
	if err != nil {
		return fmt.Errorf("list provider models: %w", err)
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
	if a.tester != nil {
		if err := a.tester.TestModel(ctx, p, m.ID); err != nil {
			return fmt.Errorf("test provider model reachability: %w", err)
		}
	}
	if err := adapter.Prepare(ctx, r, p, m, models, pr); err != nil {
		return fmt.Errorf("prepare route: %w", err)
	}
	if r.RestartOnActivate {
		if err := restart(ctx, adapter, r, p); err != nil {
			return fmt.Errorf("restart platform: %w", err)
		}
	}
	r.Default = true
	if err := a.routes.Save(ctx, r); err != nil {
		return fmt.Errorf("save active route: %w", err)
	}
	return nil
}

func restart(ctx context.Context, adapter ports.PlatformAdapter, r route.Route, p provider.Provider) error {
	return adapter.Restart(ctx, r, p)
}
