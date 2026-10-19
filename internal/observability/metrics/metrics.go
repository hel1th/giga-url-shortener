// Package metrics defines Prometheus metrics and HTTP middleware for observability.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts all handled HTTP requests partitioned by method, route pattern, and status code.
	HTTPRequestsTotal *prometheus.CounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by method, path template, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures the HTTP request handling latency in seconds.
	HTTPRequestDuration *prometheus.HistogramVec = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distributions in seconds.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path", "status"},
	)

	// AnalyticsClicksIngestedTotal counts total click events consumed from Kafka by the analytics service.
	AnalyticsClicksIngestedTotal prometheus.Counter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "analytics_clicks_ingested_total",
			Help: "Total number of link click events successfully read from Kafka.",
		},
	)

	// AnalyticsBatchFlushDuration measures the duration of ClickHouse batch insert operations.
	AnalyticsBatchFlushDuration prometheus.Histogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analytics_batch_flush_duration_seconds",
			Help:    "Latency of ClickHouse batch insert flush operations in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
	)

	// AnalyticsBatchSize records distribution of batch sizes flushed to ClickHouse.
	AnalyticsBatchSize prometheus.Histogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analytics_batch_size",
			Help:    "Number of click items in each ClickHouse flush batch.",
			Buckets: []float64{1, 10, 50, 100, 250, 500, 1000},
		},
	)
)

// HTTPMiddleware measures HTTP request duration and records response status code.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unmatched"
		}

		status := rw.Status()
		if status == 0 {
			status = http.StatusOK
		}

		statusStr := strconv.Itoa(status)

		HTTPRequestsTotal.WithLabelValues(r.Method, routePattern, statusStr).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, routePattern, statusStr).
			Observe(duration.Seconds())
	})
}
