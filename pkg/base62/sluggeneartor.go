package base62

import (
	"crypto/rand"
	"errors"
	"fmt"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func MustGenerate(length int) string {
	res, err := Generate(length)
	if err != nil {
		panic("crypto/rand is unavailable: " + err.Error())
	}
	return res
}

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
