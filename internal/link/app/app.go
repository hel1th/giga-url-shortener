package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/link/client/analytics"
	"github.com/hel1th/giga-url-shortener/internal/link/config"
	"github.com/hel1th/giga-url-shortener/internal/link/producer"
	"github.com/hel1th/giga-url-shortener/internal/link/repository/postgres"
	redisrepo "github.com/hel1th/giga-url-shortener/internal/link/repository/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hel1th/giga-url-shortener/internal/broker/kafka"

	"github.com/hel1th/giga-url-shortener/internal/link/handler"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
)

func Run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	pgPool, err := initPostgres(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("init pg: %w", err)
	}
	defer pgPool.Close()

	rdb, err := initRedis(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer rdb.Close()

	kafkaClient, err := initKafka(cfg.Kafka)
	if err != nil {
		return fmt.Errorf("init kafka: %w", err)
	}
	defer kafkaClient.Close()

	linkPgRepo := postgres.NewPostgresRepo(pgPool)
	linkRedisRepo := redisrepo.NewRedisRepo(rdb, cfg.Redis.LinkTTL)

	baseKafkaProducer := kafka.NewProducer(kafkaClient, cfg.Kafka.Topic)
	eventProducer := producer.NewLinkEventProducer(baseKafkaProducer)

	analyticsConn, err := grpc.NewClient(cfg.AnalyticsGRPC.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("connect to analytics grpc: %w", err)
	}
	defer analyticsConn.Close()

	analyticsClient := analytics.NewClient(analyticsConn)

	linkUseCase := usecase.NewLinkUseCase(linkPgRepo, linkRedisRepo, eventProducer, analyticsClient, logger)

	linkHandler := handler.NewLinkHandler(linkUseCase, logger, cfg.LinkHTTP.BaseURL)
	router := handler.NewRouter(linkHandler)

	server := &http.Server{
		Addr:         cfg.LinkHTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.LinkHTTP.ReadTimeout,
		WriteTimeout: cfg.LinkHTTP.WriteTimeout,
		IdleTimeout:  cfg.LinkHTTP.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("starting http server", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		logger.Info("shutdown started", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		logger.Info("shutdown completed", "signal", sig)
	}

	return nil
}

func initPostgres(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}

	return pool, nil
}

func initRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}

func initKafka(cfg config.KafkaConfig) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return client, nil
}
