package domain

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Link represents the core domain model for a shortened URL.
type Link struct {
	ID          string
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}

// NewLink is a factory function that ensures business invariants are met before creating a Link.
func NewLink(originalURL string) (*Link, error) {
	if !isValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	return &Link{
		ID:          uuid.New().String(),
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func isValidURL(u string) bool {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	if parsedURL.Host == "" {
		return false
	}
	return true
}
