package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/dto"
)

func errorToStatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrInvalidURL):
		return http.StatusBadRequest

	case errors.Is(err, domain.ErrRetryLimitExceeded):
		return http.StatusServiceUnavailable

	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}

func (h *LinkHandler) respondWithError(w http.ResponseWriter, err error) {
	errMessage := ""
	code := errorToStatusCode(err)

	if code != http.StatusInternalServerError {
		errMessage = err.Error()
	} else {
		h.logger.Error("internal server error", "error", err)
		errMessage = "internal server error"
	}

	h.respondWithCode(w, code, errMessage)
}

func (h *LinkHandler) respondWithCode(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: message,
	})
}
