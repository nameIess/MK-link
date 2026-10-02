package main

import (
    "embed"
    "log"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()
    if err := wails.Run(&options.App{
        Title: "MK-Link",
        Width: 1280,
        Height: 820,
        MinWidth: 980,
        MinHeight: 680,
        AssetServer: &assetserver.Options{Assets: assets},
        Bind: []interface{}{app},
        OnStartup: app.startup,
        Windows: &windows.Options{
            Theme: windows.Dark,
            WebviewIsTransparent: false,
            WindowIsTranslucent: false,
        },
        DragAndDrop: &options.DragAndDrop{
            EnableFileDrop: true,
            DisableWebViewDrop: true,
        },
        SingleInstanceLock: &options.SingleInstanceLock{
            UniqueId: "9d1a4e58-4c2d-4f65-bf4b-2e6d31d9f7c2",
        },
        EnableDefaultContextMenu: false,
        EnableFraudulentWebsiteDetection: true,
    }); err != nil {
        log.Fatal(err)
    }
}
