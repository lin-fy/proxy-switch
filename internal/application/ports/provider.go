package ports

import (
	"context"

	"codex-provider-hub/internal/domain/provider"
)

type RemoteModel struct {
	ID   string
	Name string
}

type ProviderTester interface {
	Test(context.Context, provider.Provider) error
	TestModel(context.Context, provider.Provider, string) error
	ListModels(context.Context, provider.Provider) ([]RemoteModel, error)
}
