// Package handler implements HTTP presentation and routing for the link service.
package handler

import (
	"context"
	"log/slog"
	"strings"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
)

// LinkService defines the application usecase methods required by HTTP handlers.
type LinkService interface {
	CreateShortLink(ctx context.Context, in usecase.CreateLinkInput) (*domain.Link, error)
	GetOriginalURL(ctx context.Context, in usecase.RedirectInput) (string, error)
	Stats(ctx context.Context, in usecase.StatsInput) (*domain.LinkStats, error)
}

// LinkHandler handles incoming HTTP requests for short links.
type LinkHandler struct {
	uc      LinkService
	logger  *slog.Logger
	baseURL string
}

// NewLinkHandler creates a new LinkHandler instance with dependencies.
func NewLinkHandler(
	useCase LinkService,
	logger *slog.Logger,
	baseURL string,
) *LinkHandler {
	return &LinkHandler{
		uc:      useCase,
		logger:  logger,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}
