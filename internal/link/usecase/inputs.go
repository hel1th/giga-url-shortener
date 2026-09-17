// Package usecase contains the core application use cases and business workflows.
package usecase

import (
	"github.com/hel1th/giga-url-shortener/internal/link/domain"
)

// CreateLinkInput encapsulates data required to create a short link.
type CreateLinkInput struct {
	OriginalURL string
}

// RedirectInput encapsulates data required to process a redirect and track analytics.
type RedirectInput struct {
	ShortCode string
	Meta      domain.ClickMeta
}

// StatsInput encapsulates data required to fetch statistics for a short link.
type StatsInput struct {
	ShortCode string
}
