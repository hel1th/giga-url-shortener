package handler

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
)

// Redirect handles GET requests to redirect short codes to their destination URLs.
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, shortCodeParam)

	input := usecase.RedirectInput{
		ShortCode: shortCode,
		Meta:      extractMeta(shortCode, r),
	}

	originalURL, err := h.uc.GetOriginalURL(ctx, input)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func extractMeta(shortCode string, r *http.Request) domain.ClickMeta {
	return domain.ClickMeta{
		ShortCode: shortCode,
		ClickedAt: time.Now().UTC(),
		IP:        extractIP(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
