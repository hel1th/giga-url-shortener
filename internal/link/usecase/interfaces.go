package usecase

import (
	"context"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
)

// LinkRepository defines persistent storage operations for links.
type LinkRepository interface {
	Get(ctx context.Context, shortCode string) (*domain.Link, error)
	Create(ctx context.Context, link *domain.Link) error
}

// LinkCache defines caching operations for links.
type LinkCache interface {
	Get(ctx context.Context, shortCode string) (string, error)
	Set(ctx context.Context, shortCode, originalURL string) error
}

// EventProducer defines operations to emit domain events.
type EventProducer interface {
	PublishLinkClicked(ctx context.Context, event domain.LinkClicked) error
}

// AnalyticsClient defines the interface for communicating with the analytics service.
type AnalyticsClient interface {
	GetStats(ctx context.Context, shortCode string) (*domain.LinkStats, error)
}
