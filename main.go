package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "codex-provider-hub",
		Description: "Codex Provider Hub",
		Services: []application.Service{
			application.NewService(NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:             "Codex Provider Hub",
		Width:             1024,
		Height:            768,
		BackgroundColour:  application.NewRGB(27, 38, 54),
		URL:               "/",
		DevToolsEnabled:   false,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
