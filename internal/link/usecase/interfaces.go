package usecase

import (
	"context"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
)

type LinkRepository interface {
	Get(ctx context.Context, shortCode string) (*domain.Link, error)
	Create(ctx context.Context, link *domain.Link) error
}

type LinkCache interface {
	Get(ctx context.Context, shortCode string) (string, error)
	Set(ctx context.Context, shortCode, originalUrl string) error
}

type EventProducer interface {
	PublishLinkClicked(ctx context.Context, event domain.LinkClicked) error
}

type AnalyticsClient interface {
	GetStats(ctx context.Context, shortCode string) (*domain.LinkStats, error)
}
