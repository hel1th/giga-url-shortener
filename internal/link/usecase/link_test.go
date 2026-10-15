package usecase_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateShortLink_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	originalURL := "https://example.com"

	repo.EXPECT().
		Create(ctx, mock.MatchedBy(func(l *domain.Link) bool {
			return l.OriginalURL == originalURL &&
				len(l.ShortCode) == usecase.ShortCodeLen &&
				l.ID != ""
		})).Return(nil)

	cache.EXPECT().Set(ctx, mock.AnythingOfType("string"), originalURL).
		Return(nil)

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	link, err := uc.CreateShortLink(ctx, usecase.CreateLinkInput{OriginalURL: "https://example.com"})

	require.NoError(t, err)
	assert.Equal(t, originalURL, link.OriginalURL)
	assert.Len(t, link.ShortCode, usecase.ShortCodeLen)
}

func TestCreateShortLink_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	wantErr := errors.New("db down")

	repo.EXPECT().
		Create(ctx, mock.Anything).
		Return(wantErr)

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	link, err := uc.CreateShortLink(ctx, usecase.CreateLinkInput{OriginalURL: "https://example.com"})

	assert.Nil(t, link)
	assert.ErrorIs(t, err, wantErr)
	cache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateShortLink_CacheSetErrorIsSwallowed(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	repo.EXPECT().Create(ctx, mock.Anything).Return(nil)
	cache.EXPECT().
		Set(ctx, mock.Anything, mock.Anything).
		Return(errors.New("redis down"))

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	link, err := uc.CreateShortLink(ctx, usecase.CreateLinkInput{OriginalURL: "https://example.com"})

	require.NoError(t, err)
	assert.NotNil(t, link)
}

func TestStats_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	shortCode := "abc1234"
	expectedStats := &domain.LinkStats{
		ShortCode:    shortCode,
		TotalClicks:  10,
		UniqueClicks: 7,
		Browsers:     map[string]uint64{"Chrome": 10},
	}

	// Ссылка найдена в кэше
	cache.EXPECT().Get(ctx, shortCode).Return("https://example.com", nil)
	analytics.EXPECT().GetStats(ctx, shortCode).Return(expectedStats, nil)

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	stats, err := uc.Stats(ctx, usecase.StatsInput{ShortCode: shortCode})

	require.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
}

func TestStats_LinkNotFound(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	shortCode := "notfound"

	// В кэше нет, в репозитории тоже нет -> сохраняем в негативный кэш
	cache.EXPECT().Get(ctx, shortCode).Return("", domain.ErrNotFound)
	repo.EXPECT().Get(ctx, shortCode).Return(nil, domain.ErrNotFound)
	cache.EXPECT().SetUnavailable(ctx, shortCode).Return(nil)

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	stats, err := uc.Stats(ctx, usecase.StatsInput{ShortCode: shortCode})

	assert.Nil(t, stats)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	analytics.AssertNotCalled(t, "GetStats", mock.Anything, mock.Anything)
}

func TestStats_LinkNotFound_NegativeCacheHit(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewLinkRepository(t)
	cache := mocks.NewLinkCache(t)
	producer := mocks.NewEventProducer(t)
	analytics := mocks.NewAnalyticsClient(t)

	shortCode := "notfound"

	// В кэше помечен как отсутствующий (negative cache hit) -> в репозиторий не идем
	cache.EXPECT().Get(ctx, shortCode).Return("", domain.ErrNotFoundCached)

	uc := usecase.NewLinkUseCase(repo, cache, producer, analytics, slog.Default())

	stats, err := uc.Stats(ctx, usecase.StatsInput{ShortCode: shortCode})

	assert.Nil(t, stats)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	analytics.AssertNotCalled(t, "GetStats", mock.Anything, mock.Anything)
}

