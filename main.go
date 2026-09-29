package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	modelapp "codex-provider-hub/internal/application/model"
	"codex-provider-hub/internal/application/profile"
	providerapp "codex-provider-hub/internal/application/provider"
	routeapp "codex-provider-hub/internal/application/route"
	"codex-provider-hub/internal/infrastructure/codex"
	"codex-provider-hub/internal/infrastructure/config"
	wailsui "codex-provider-hub/internal/interfaces/wails"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ui, err := newApp()
	if err != nil {
		log.Fatal(err)
	}
	app := application.New(application.Options{
		Name:        "codex-provider-hub",
		Description: "Codex Provider Hub",
		Services: []application.Service{
			application.NewService(ui),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Codex Provider Hub",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		DevToolsEnabled:  false,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func newApp() (*wailsui.App, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	store := config.NewStore(filepath.Join(configDir, "CodexProviderHub", "state.json"))
	providers := config.NewProviderRepository(store)
	models := config.NewModelRepository(store)
	routes := config.NewRouteRepository(store)
	profiles := config.NewProfileRepository(store)
	providerService := providerapp.NewService(providers)
	modelService := modelapp.NewService(models)
	routeService := routeapp.NewService(routes)
	profileService := profile.NewService(profiles)
	adapter, err := codex.NewAdapter("")
	if err != nil {
		return nil, err
	}
	factory, err := codex.NewFactory("")
	if err != nil {
		return nil, err
	}
	activator := routeapp.NewActivator(routes, providers, models, profiles, factory)
	return wailsui.NewApp(providerService, modelService, routeService, profileService, activator, adapter), nil
}
