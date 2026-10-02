package route

import (
	"context"
	"errors"
	"testing"

	"proxy-switch/internal/application/ports"
	"proxy-switch/internal/domain/model"
	"proxy-switch/internal/domain/profile"
	"proxy-switch/internal/domain/provider"
	"proxy-switch/internal/domain/route"
)

type activationRoutes struct {
	item  route.Route
	saved []route.Route
}

func (r *activationRoutes) List(context.Context) ([]route.Route, error) {
	return []route.Route{r.item}, nil
}
func (r *activationRoutes) Get(context.Context, string) (route.Route, error) { return r.item, nil }
func (r *activationRoutes) Save(_ context.Context, item route.Route) error {
	r.item = item
	r.saved = append(r.saved, item)
	return nil
}
func (r *activationRoutes) Delete(context.Context, string) error { return nil }

type activationProviders struct{ item provider.Provider }

func (r activationProviders) List(context.Context) ([]provider.Provider, error) {
	return []provider.Provider{r.item}, nil
}
func (r activationProviders) Get(context.Context, string) (provider.Provider, error) {
	return r.item, nil
}
func (r activationProviders) Save(context.Context, provider.Provider) error { return nil }
func (r activationProviders) Delete(context.Context, string) error          { return nil }

type activationModels struct {
	selected model.Model
	items    []model.Model
	disabled bool
}

func (r activationModels) List(context.Context) ([]model.Model, error) { return r.items, nil }
func (r activationModels) ListByProvider(context.Context, string) ([]model.Model, error) {
	return r.items, nil
}
func (r activationModels) Get(context.Context, string, string) (model.Model, error) {
	item := r.selected
	item.Enabled = !r.disabled
	return item, nil
}
func (r activationModels) Save(context.Context, model.Model) error      { return nil }
func (r activationModels) Delete(context.Context, string, string) error { return nil }

type activationProfiles struct{ item profile.Profile }

func (r activationProfiles) List(context.Context) ([]profile.Profile, error) {
	return []profile.Profile{r.item}, nil
}
func (r activationProfiles) Get(context.Context, string) (profile.Profile, error) { return r.item, nil }
func (r activationProfiles) Save(context.Context, profile.Profile) error          { return nil }
func (r activationProfiles) Delete(context.Context, string) error                 { return nil }

type activationAdapter struct {
	validated         bool
	prepared          bool
	launched          bool
	restarted         bool
	restartProviderID string
	restartErr        error
}

func (a *activationAdapter) Validate(context.Context, route.Route, provider.Provider, model.Model) error {
	a.validated = true
	return nil
}
func (a *activationAdapter) Prepare(_ context.Context, _ route.Route, _ provider.Provider, selected model.Model, models []model.Model, _ profile.Profile) error {
	if selected.ID == "" || len(models) != 2 {
		return errors.New("unexpected model selection")
	}
	a.prepared = true
	return nil
}
func (a *activationAdapter) Launch(context.Context, route.Route, provider.Provider) error {
	a.launched = true
	return nil
}
func (a *activationAdapter) Restart(_ context.Context, _ route.Route, p provider.Provider) error {
	a.restarted = true
	a.restartProviderID = p.ID
	return a.restartErr
}
func (a *activationAdapter) IsRunning(context.Context) (bool, error) { return false, nil }

type activationFactory struct{ adapter ports.PlatformAdapter }

func (f activationFactory) Create(string) (ports.PlatformAdapter, error) { return f.adapter, nil }

type activationTester struct {
	err   error
	calls int
}

func (t *activationTester) Test(context.Context, provider.Provider) error { return nil }
func (t *activationTester) TestModel(context.Context, provider.Provider, string) error {
	t.calls++
	return t.err
}
func (t *activationTester) ListModels(context.Context, provider.Provider) ([]ports.RemoteModel, error) {
	return nil, nil
}

func TestActivatorRejectsDisabledModel(t *testing.T) {
	routes := &activationRoutes{item: route.Route{
		ID: "route-a", Name: "Route A", PlatformID: route.PlatformCodex,
		ProviderID: "provider-a", ModelID: "model-a",
	}}
	adapter := &activationAdapter{}
	tester := &activationTester{}
	activator := NewActivator(
		routes,
		activationProviders{item: provider.Provider{ID: "provider-a", Name: "Provider A", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}},
		activationModels{selected: model.Model{ProviderID: "provider-a", ID: "model-a"}, disabled: true, items: []model.Model{{ProviderID: "provider-a", ID: "model-a"}}},
		activationProfiles{item: profile.Profile{ID: "profile-a", Name: "Profile A"}},
		activationFactory{adapter: adapter}, tester,
	)

	err := activator.Activate(context.Background(), "route-a", "profile-a")
	if !errors.Is(err, ErrModelDisabled) {
		t.Fatalf("disabled model error = %v", err)
	}
	if tester.calls != 0 || adapter.validated || adapter.prepared || len(routes.saved) != 0 || routes.item.Default {
		t.Fatalf("disabled model calls = tester:%d validated:%v prepared:%v saved:%d default:%v", tester.calls, adapter.validated, adapter.prepared, len(routes.saved), routes.item.Default)
	}
}

func TestActivatorPreparesAndMarksRouteDefault(t *testing.T) {
	routes := &activationRoutes{item: route.Route{
		ID: "route-a", Name: "Route A", PlatformID: route.PlatformCodex,
		ProviderID: "provider-a", ModelID: "model-a",
	}}
	adapter := &activationAdapter{}
	activator := NewActivator(
		routes,
		activationProviders{item: provider.Provider{ID: "provider-a", Name: "Provider A", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}},
		activationModels{selected: model.Model{ProviderID: "provider-a", ID: "model-a"}, items: []model.Model{{ProviderID: "provider-a", ID: "model-a"}, {ProviderID: "provider-a", ID: "model-b"}}},
		activationProfiles{item: profile.Profile{ID: "profile-a", Name: "Profile A"}},
		activationFactory{adapter: adapter},
	)

	if err := activator.Activate(context.Background(), "route-a", "profile-a"); err != nil {
		t.Fatal(err)
	}
	if !adapter.validated || !adapter.prepared {
		t.Fatalf("adapter calls = validated:%v prepared:%v", adapter.validated, adapter.prepared)
	}
	if adapter.launched {
		t.Fatal("route without restart flag should not launch Codex")
	}
	if !routes.item.Default || len(routes.saved) != 1 {
		t.Fatalf("route default/save = %v/%d", routes.item.Default, len(routes.saved))
	}
}

func TestActivatorLaunchesWhenRestartIsEnabled(t *testing.T) {
	routes := &activationRoutes{item: route.Route{
		ID: "route-a", Name: "Route A", PlatformID: route.PlatformCodex,
		ProviderID: "provider-a", ModelID: "model-a", RestartOnActivate: true,
	}}
	adapter := &activationAdapter{}
	activator := NewActivator(
		routes,
		activationProviders{item: provider.Provider{ID: "provider-a", Name: "Provider A", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}},
		activationModels{selected: model.Model{ProviderID: "provider-a", ID: "model-a"}, items: []model.Model{{ProviderID: "provider-a", ID: "model-a"}, {ProviderID: "provider-a", ID: "model-b"}}},
		activationProfiles{item: profile.Profile{ID: "profile-a", Name: "Profile A"}},
		activationFactory{adapter: adapter},
	)

	if err := activator.Activate(context.Background(), "route-a", "profile-a"); err != nil {
		t.Fatal(err)
	}
	if !adapter.restarted || adapter.launched {
		t.Fatalf("restart-enabled route calls = restarted:%v launched:%v", adapter.restarted, adapter.launched)
	}
	if adapter.restartProviderID != "provider-a" {
		t.Fatalf("restart provider = %q", adapter.restartProviderID)
	}
}

func TestActivatorDoesNotDefaultWhenRestartFails(t *testing.T) {
	routes := &activationRoutes{item: route.Route{ID: "route-a", Name: "Route A", PlatformID: route.PlatformCodex, ProviderID: "provider-a", ModelID: "model-a", RestartOnActivate: true}}
	adapter := &activationAdapter{restartErr: errors.New("restart failed")}
	activator := NewActivator(
		routes,
		activationProviders{item: provider.Provider{ID: "provider-a", Name: "Provider A", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}},
		activationModels{selected: model.Model{ProviderID: "provider-a", ID: "model-a"}, items: []model.Model{{ProviderID: "provider-a", ID: "model-a"}, {ProviderID: "provider-a", ID: "model-b"}}},
		activationProfiles{item: profile.Profile{ID: "profile-a", Name: "Profile A"}},
		activationFactory{adapter: adapter},
	)
	if err := activator.Activate(context.Background(), routes.item.ID, "profile-a"); err == nil {
		t.Fatal("expected restart failure")
	}
	if len(routes.saved) != 0 || routes.item.Default {
		t.Fatalf("route saved/default after restart failure = %d/%v", len(routes.saved), routes.item.Default)
	}
}

func TestActivatorDoesNotPrepareOrDefaultWhenReachabilityFails(t *testing.T) {
	routes := &activationRoutes{item: route.Route{
		ID: "route-a", Name: "Route A", PlatformID: route.PlatformCodex,
		ProviderID: "provider-a", ModelID: "model-a",
	}}
	adapter := &activationAdapter{}
	tester := &activationTester{err: errors.New("unreachable")}
	activator := NewActivator(
		routes,
		activationProviders{item: provider.Provider{ID: "provider-a", Name: "Provider A", BaseURL: "https://example.test", Protocol: provider.ProtocolResponses}},
		activationModels{selected: model.Model{ProviderID: "provider-a", ID: "model-a"}, items: []model.Model{{ProviderID: "provider-a", ID: "model-a"}, {ProviderID: "provider-a", ID: "model-b"}}},
		activationProfiles{item: profile.Profile{ID: "profile-a", Name: "Profile A"}},
		activationFactory{adapter: adapter}, tester,
	)

	if err := activator.Activate(context.Background(), "route-a", "profile-a"); err == nil {
		t.Fatal("expected reachability failure")
	}
	if tester.calls != 1 || adapter.prepared || len(routes.saved) != 0 || routes.item.Default {
		t.Fatalf("reachability failure calls = tester:%d prepared:%v saved:%d default:%v", tester.calls, adapter.prepared, len(routes.saved), routes.item.Default)
	}
}
