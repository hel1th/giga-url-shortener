// Package base62 provides cryptographically secure Base62 random slug generation.
package base62

import (
	"crypto/rand"
	"errors"
	"fmt"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// MustGenerate generates a Base62 string of the specified length or panics on failure.
func MustGenerate(length int) string {
	res, err := Generate(length)
	if err != nil {
		panic("crypto/rand is unavailable: " + err.Error())
	}
	return res
}

// Generate generates a cryptographically random Base62 string of the requested length.
func Generate(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	for i := range bytes {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}
	return string(bytes), nil
}
