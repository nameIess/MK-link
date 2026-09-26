//go:build windows

package main

import (
	"log/slog"
	"os"

	"github.com/nameIess/MK-link/internal/ui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := ui.Run(logger); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}
