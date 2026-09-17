package domain

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("link not found")
	ErrAlreadyExists = errors.New("link already exists")
	ErrInvalidURL    = errors.New("invalid url format")

	ErrCacheMiss          = errors.New("not found in cache")
	ErrRetryLimitExceeded = errors.New("failed to generate unique short code after max retries")
)
