package ports

import (
	"context"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
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
