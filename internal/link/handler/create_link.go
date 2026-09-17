package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/dto"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
)

func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLinkRequest
	r.Body = http.MaxBytesReader(w, r.Body, 10<<10)
	defer func() { _ = r.Body.Close() }()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode create link request body", "error", err)
		h.respondWithCode(w, http.StatusBadRequest, "invalid request body format")
		return
	}

	ctx := r.Context()
	input := toCreateLinkInput(req)
	link, err := h.uc.CreateShortLink(ctx, input)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	resp := toCreateLinkResponse(link, h.baseURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode create link response", "error", err)
	}
}

// toCreateLinkResponse converts a domain Link object into the HTTP representation (DTO).
func toCreateLinkResponse(link *domain.Link, baseURL string) dto.CreateLinkResponse {
	return dto.CreateLinkResponse{
		ShortCode:   link.ShortCode,
		ShortURL:    baseURL + "/r/" + link.ShortCode,
		OriginalURL: link.OriginalURL,
	}
}

func toCreateLinkInput(req dto.CreateLinkRequest) usecase.CreateLinkInput {
	return usecase.CreateLinkInput{OriginalURL: req.OriginalURL}
}
