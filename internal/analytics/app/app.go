// Package app initializes infrastructure dependencies and orchestrates the analytics service lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hel1th/giga-url-shortener/internal/analytics/config"
	"github.com/hel1th/giga-url-shortener/internal/analytics/consumer"
	chrepo "github.com/hel1th/giga-url-shortener/internal/analytics/repository/clickhouse"
	grpcserver "github.com/hel1th/giga-url-shortener/internal/analytics/server/grpc"
	"github.com/hel1th/giga-url-shortener/internal/observability/tracing"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Run initializes infrastructure dependencies and runs gRPC, HTTP and Kafka consumer until shutdown.
func Run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	tp, err := tracing.InitTracer(ctx, cfg.OTel.Endpoint, "analytics-service")
	if err != nil {
		return fmt.Errorf("init tracer: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			logger.Error("tracer shutdown failed", slog.String("error", err.Error()))
		}
	}()

	chConn, err := initClickHouse(ctx, cfg.ClickHouse)
	if err != nil {
		return fmt.Errorf("init clickhouse: %w", err)
	}
	defer func() { _ = chConn.Close() }()

	kafkaClient, err := initKafka(cfg.Kafka)
	if err != nil {
		return fmt.Errorf("init kafka: %w", err)
	}
	defer kafkaClient.Close()

	repo := chrepo.NewClicksRepository(chConn)
	c := consumer.NewConsumer(
		kafkaClient,
		repo,
		cfg.ClickHouse.BatchSize,
		cfg.ClickHouse.FlushInterval,
		logger,
	)

	grpcHandler := grpcserver.NewServer(repo)
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	analyticsv1.RegisterAnalyticsServiceServer(grpcServer, grpcHandler)

	grpcListener, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		return fmt.Errorf("listen grpc port %s: %w", cfg.GRPC.Port, err)
	}
	defer func() { _ = grpcListener.Close() }()

	httpRouter := chi.NewRouter()
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpRouter.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:         cfg.HTTP.Port,
		Handler:      httpRouter,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(sigCtx)

	g.Go(func() error {
		logger.Info(
			"starting analytics kafka consumer",
			"batch_size", cfg.ClickHouse.BatchSize,
			"flush_interval", cfg.ClickHouse.FlushInterval,
		)
		return c.Run(gCtx)
	})

	g.Go(func() error {
		logger.Info("starting analytics grpc server", "addr", cfg.GRPC.Port)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("grpc server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("starting analytics http server", "addr", cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("shutdown signal received, stopping servers gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		grpcServer.GracefulStop()
		_ = httpServer.Shutdown(shutdownCtx)

		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("service error: %w", err)
	}

	logger.Info("analytics service stopped gracefully")
	return nil
}

func initClickHouse(ctx context.Context, cfg config.ClickHouseConfig) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.DB,
			Username: cfg.User,
			Password: cfg.Password,
		},
		DialTimeout:     cfg.DialTimeout,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return conn, nil
}

func initKafka(cfg config.KafkaConfig) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topic),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer client: %w", err)
	}

	return client, nil
}
