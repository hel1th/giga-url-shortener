package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/pkg/base62"
)

// ShortCodeLen is the length of short codes generated for links.
const ShortCodeLen = 7

// MaxRetries is the maximum number of attempts to generate a non-colliding short code.
const MaxRetries = 5

// LinkUseCase coordinates the business logic for creating, resolving, and querying links.
type LinkUseCase struct {
	repo          LinkRepository
	cache         LinkCache
	eventProducer EventProducer
	analytics     AnalyticsClient
	logger        *slog.Logger
}

// NewLinkUseCase creates a new LinkUseCase instance with dependencies.
func NewLinkUseCase(
	repo LinkRepository,
	cache LinkCache,
	eventProducer EventProducer,
	analytics AnalyticsClient,
	logger *slog.Logger,
) *LinkUseCase {
	return &LinkUseCase{
		repo:          repo,
		cache:         cache,
		eventProducer: eventProducer,
		analytics:     analytics,
		logger:        logger,
	}
}

// CreateShortLink generates a new unique short link and persists it.
func (uc *LinkUseCase) CreateShortLink(ctx context.Context, in CreateLinkInput) (*domain.Link, error) {
	link, err := domain.NewLink(in.OriginalURL)
	if err != nil {
		return nil, err
	}

	err = uc.generateAndSaveLink(ctx, link)
	if err != nil {
		return nil, err
	}

	cacheErr := uc.cache.Set(ctx, link.ShortCode, link.OriginalURL)
	if cacheErr != nil {
		uc.logger.Error("put link in cache", "error", cacheErr)
	}

	return link, nil
}

// GetOriginalURL resolves a short code to its destination URL and emits a click event.
func (uc *LinkUseCase) GetOriginalURL(ctx context.Context, in RedirectInput) (string, error) {
	originalURL, err := uc.resolveURL(ctx, in.ShortCode)
	if err != nil {
		return "", err
	}

	uc.publishClickEvent(ctx, in.ShortCode, in.Meta)

	return originalURL, nil
}

// Stats queries aggregate click statistics for a short link.
func (uc *LinkUseCase) Stats(ctx context.Context, in StatsInput) (*domain.LinkStats, error) {
	if _, err := uc.resolveURL(ctx, in.ShortCode); err != nil {
		return nil, err
	}

	stats, err := uc.analytics.GetStats(ctx, in.ShortCode)
	if err != nil {
		uc.logger.Error("failed to get link stats", "error", err, "short_code", in.ShortCode)
		return nil, err
	}

	return stats, nil
}

func (uc *LinkUseCase) generateAndSaveLink(ctx context.Context, link *domain.Link) error {
	for retryCnt := range MaxRetries {
		link.ShortCode = base62.MustGenerate(ShortCodeLen)

		dbErr := uc.repo.Create(ctx, link)
		if dbErr == nil {
			return nil
		}

		if !errors.Is(dbErr, domain.ErrAlreadyExists) {
			uc.logger.Error("failed to put link in db", "error", dbErr)
			return dbErr
		}

		uc.logger.Warn(
			"short code collision, retrying",
			"attempt", retryCnt+1,
			"short_code", link.ShortCode,
			"error", dbErr,
		)
	}

	return domain.ErrRetryLimitExceeded
}

func (uc *LinkUseCase) resolveURL(ctx context.Context, shortCode string) (string, error) {
	url, err := uc.cache.Get(ctx, shortCode)
	if err == nil {
		return url, nil
	}
	if errors.Is(err, domain.ErrNotFoundCached) {
		return "", domain.ErrNotFound
	}
	if !errors.Is(err, domain.ErrNotFound) {
		uc.logger.Error("redis cache get error, falling back to db", "error", err, "short_code", shortCode)
	}

	link, dbErr := uc.repo.Get(ctx, shortCode)
	if dbErr != nil {
		if errors.Is(dbErr, domain.ErrNotFound) {
			if cacheErr := uc.cache.SetUnavailable(ctx, shortCode); cacheErr != nil {
				uc.logger.Error("redis cache set unavailable error", "error", cacheErr, "short_code", shortCode)
			}
			return "", domain.ErrNotFound
		}
		uc.logger.Error("failed to get link from repo", "error", dbErr, "short_code", shortCode)
		return "", dbErr
	}

	if setErr := uc.cache.Set(ctx, shortCode, link.OriginalURL); setErr != nil {
		uc.logger.Warn("failed to warm up redis cache", "error", setErr, "short_code", shortCode)
	}

	return link.OriginalURL, nil
}

func (uc *LinkUseCase) publishClickEvent(ctx context.Context, shortCode string, meta domain.ClickMeta) {
	clickEvent := domain.LinkClicked{
		ShortCode:  shortCode,
		Meta:       meta,
		OccurredAt: time.Now(),
	}

	if err := uc.eventProducer.PublishLinkClicked(ctx, clickEvent); err != nil {
		uc.logger.Error("failed to publish link clicked event to broker", "error", err, "short_code", shortCode)
	}
}
