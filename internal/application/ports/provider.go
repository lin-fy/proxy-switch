package ports

import (
	"context"

	"codex-provider-hub/internal/domain/provider"
)

type ProviderTester interface {
	Test(context.Context, provider.Provider) error
	TestModel(context.Context, provider.Provider, string) error
}
