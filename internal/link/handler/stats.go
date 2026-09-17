package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/dto"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
)

func (h *LinkHandler) Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.uc.Stats(ctx, usecase.StatsInput{
		ShortCode: chi.URLParam(r, shortCodeParam),
	})
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	resp := toLinkStatsResponse(stats)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func toLinkStatsResponse(stats *domain.LinkStats) dto.LinkStatsResponse {
	return dto.LinkStatsResponse{
		ShortCode:    stats.ShortCode,
		TotalClicks:  stats.TotalClicks,
		UniqueClicks: stats.UniqueClicks,
		Browsers:     stats.Browsers,
		OS:           stats.OS,
		Devices:      stats.Devices,
	}
}
