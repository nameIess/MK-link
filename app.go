package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"mklink/internal/linker"
)

type App struct {
	ctx     context.Context
	service *linker.Service
}

func NewApp(service *linker.Service) *App {
	return &App{service: service}
}

type LinkRequest struct {
	Type      string `json:"type"`
	Target    string `json:"target"`
	Link      string `json:"link"`
	Overwrite bool   `json:"overwrite"`
}

type LinkResult struct {
	LinkType string `json:"linkType"`
	Target   string `json:"target"`
	Link     string `json:"link"`
	Message  string `json:"message"`
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CreateLink(request LinkRequest) (LinkResult, error) {
	if a.service == nil {
		return LinkResult{}, errors.New("link service is unavailable")
	}
	if request.Overwrite {
		link, err := filepath.Abs(filepath.Clean(request.Link))
		if err != nil {
			return LinkResult{}, fmt.Errorf("resolve link path: %w", err)
		}
		if err := a.service.RemoveExistingLink(link); err != nil {
			return LinkResult{}, err
		}
	}

	if err := a.service.Create(linker.Request{
		Type:   linker.Type(request.Type),
		Target: request.Target,
		Link:   request.Link,
	}); err != nil {
		return LinkResult{}, err
	}

	target, err := filepath.Abs(filepath.Clean(request.Target))
	if err != nil {
		return LinkResult{}, fmt.Errorf("resolve target path: %w", err)
	}
	link, err := filepath.Abs(filepath.Clean(request.Link))
	if err != nil {
		return LinkResult{}, fmt.Errorf("resolve link path: %w", err)
	}

	return LinkResult{
		LinkType: request.Type,
		Target:   target,
		Link:     link,
		Message:  "Link created successfully.",
	}, nil
}

func (a *App) Validate(request LinkRequest) error {
	return linker.Validate(linker.Request{
		Type:   linker.Type(request.Type),
		Target: request.Target,
		Link:   request.Link,
	})
}
