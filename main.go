package main

import (
	"embed"
	"log"

	"mklink/internal/linker"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp(linker.NewService())
	if err := wails.Run(&options.App{
		Title:            "MK-Link",
		Width:            960,
		Height:          680,
		MinWidth:         820,
		MinHeight:        600,
		AssetServer:      options.AssetServer{Assets: assets},
		BackgroundColour: &options.RGBA{R: 247, G: 248, B: 250, A: 1},
		OnStartup:        app.startup,
		Windows:          &windows.Options{DisableWindowIcon: false},
		Bind:             []interface{}{app},
	}); err != nil {
		log.Fatal(err)
	}
}