package ports

import (
	"context"

	"codex-provider-hub/internal/domain/model"
	"codex-provider-hub/internal/domain/profile"
	"codex-provider-hub/internal/domain/provider"
	"codex-provider-hub/internal/domain/route"
)

type PlatformAdapter interface {
	Validate(context.Context, route.Route, provider.Provider, model.Model) error
	Prepare(context.Context, route.Route, provider.Provider, model.Model, []model.Model, profile.Profile) error
	Launch(context.Context, route.Route, provider.Provider) error
	Restart(context.Context, route.Route, provider.Provider) error
	IsRunning(context.Context) (bool, error)
}

type PlatformAdapterFactory interface {
	Create(string) (PlatformAdapter, error)
}
