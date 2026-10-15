package domain

import (
	"errors"
)

var (
	// ErrNotFound is returned when a requested link does not exist.
	ErrNotFound = errors.New("link not found")

	// ErrNotFoundCached is returned when a requested link does not cached.
	ErrNotFoundCached = errors.New("link not found in cache")

	// ErrAlreadyExists is returned when attempting to create a link that collides with an existing one.
	ErrAlreadyExists = errors.New("link already exists")

	// ErrInvalidURL is returned when an invalid URL format or scheme is provided.
	ErrInvalidURL = errors.New("invalid url format")

	// ErrRetryLimitExceeded is returned when short code collision retry limit is exceeded.
	ErrRetryLimitExceeded = errors.New("failed to generate unique short code after max retries")
)
