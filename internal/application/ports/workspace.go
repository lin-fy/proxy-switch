package ports

import (
	"context"

	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
)

// WorkspaceImporter atomically appends validated, non-conflicting resources.
// Implementations must never touch platform auth files or credential stores.
type WorkspaceImporter interface {
	ImportWorkspace(context.Context, []provider.Provider, []model.Model, []route.Route, []profile.Profile) error
}
