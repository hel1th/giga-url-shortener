package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	shortCodeParam = "short_code"
	apiVPrefix     = "/api/v1"
)

// NewRouter constructs and configures the HTTP router with routes and middlewares.
func NewRouter(h *LinkHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromXFF())
	r.Use(middleware.Recoverer)

	r.Get(apiVPrefix+"/r/{"+shortCodeParam+"}", h.Redirect)
	r.Post(apiVPrefix+"/links", h.CreateLink)

	r.Get(apiVPrefix+"/r/{"+shortCodeParam+"}/stats", h.Stats)

	return r
}
