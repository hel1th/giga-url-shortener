package main

import (
	"context"
	"log/slog"
	"os"

	linkapp "github.com/hel1th/giga-url-shortener/internal/link/app"
	"github.com/hel1th/giga-url-shortener/internal/link/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := linkapp.Run(context.Background(), cfg, logger); err != nil {
		logger.Error("link service stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
