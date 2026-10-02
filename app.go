package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/wailsapp/wails/v2/pkg/runtime"

    "mklink/internal/core"
    "mklink/internal/linker"
)

type App struct{ ctx context.Context }

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) GetAppInfo() core.AppInfo {
    return core.AppInfo{Name: "MK-Link", Version: "0.1.0", Description: "A modern Windows utility for creating and understanding filesystem links."}
}

func (a *App) GetLinkTypes(target string) []core.LinkTypeInfo { return core.AvailableLinkTypes(target) }

func (a *App) ValidateTarget(path string) (core.PathInfo, error) { return core.InspectPath(path) }

func (a *App) ValidateDestination(directory, name string, linkType core.LinkType) (core.ValidationResult, error) {
    return core.ValidateDestination(directory, name, linkType)
}

func (a *App) ChooseFile() (string, error) {
    if a.ctx == nil { return "", fmt.Errorf("application is not ready") }
    return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose target file"})
}

func (a *App) ChooseDirectory(title string) (string, error) {
    if a.ctx == nil { return "", fmt.Errorf("application is not ready") }
    if strings.TrimSpace(title) == "" { title = "Choose folder" }
    return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

func (a *App) CreateLink(req core.CreateLinkRequest) core.CreateLinkResult {
    validated, err := core.ValidateCreateRequest(req)
    if err != nil { return core.CreateLinkResult{Error: err.Error()} }
    if err := linker.Create(validated); err != nil {
        return core.CreateLinkResult{Error: linker.UserFacingError(err).Error(), NeedsElevation: linker.NeedsElevation(err)}
    }
    return core.CreateLinkResult{Success: true, LinkPath: validated.LinkPath}
}

func (a *App) OpenInExplorer(path string) error {
    clean, err := core.NormalizeExistingPath(path)
    if err != nil { return err }
    return linker.OpenExplorer(clean)
}

func (a *App) FormatPath(path string) string {
    clean := filepath.Clean(strings.TrimSpace(path))
    if clean == "." { return "" }
    return clean
}

func (a *App) HealthCheck() core.HealthStatus {
    _, err := os.Getwd()
    return core.HealthStatus{Ready: err == nil}
}
