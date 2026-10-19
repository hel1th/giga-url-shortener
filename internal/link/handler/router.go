package handler

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hel1th/giga-url-shortener/internal/observability/metrics"
	"github.com/hel1th/giga-url-shortener/internal/observability/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const shortCodeParam = "short_code"

//go:embed index.html
var indexHTML []byte

// NewRouter constructs and configures the HTTP router with routes and middlewares.
func NewRouter(h *LinkHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromXFF())
	r.Use(middleware.Recoverer)
	r.Use(tracing.RouteName)
	r.Use(metrics.HTTPMiddleware)


	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/r/{"+shortCodeParam+"}", h.Redirect)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/links", h.CreateLink)
		r.Get("/links/{"+shortCodeParam+"}/stats", h.Stats)
	})

	return r
}
