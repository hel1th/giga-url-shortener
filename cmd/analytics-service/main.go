// Package main is the entry point for the analytics service.
package main

import (
	"context"
	"log/slog"
	"os"

	analyticsapp "github.com/hel1th/giga-url-shortener/internal/analytics/app"
	"github.com/hel1th/giga-url-shortener/internal/analytics/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load analytics config", "error", err)
		os.Exit(1)
	}

	if err := analyticsapp.Run(context.Background(), cfg, logger); err != nil {
		logger.Error("analytics service stopped with error", "error", err)
		os.Exit(1)
	}
}
