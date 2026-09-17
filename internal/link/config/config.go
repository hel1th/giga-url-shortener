// Package config manages configuration loading for link-service.
package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the complete application configuration for link-service.
type Config struct {
	Env           string `env:"APP_ENV" env-default:"local"`
	LinkHTTP      HTTPConfig
	Postgres      PostgresConfig
	Redis         RedisConfig
	Kafka         KafkaConfig
	AnalyticsGRPC AnalyticsGRPCConfig
	OTel          OTelConfig
}

// HTTPConfig holds HTTP server configuration for the link service.
type HTTPConfig struct {
	BaseURL      string        `env:"BASE_URL" env-default:"http://localhost:8080"`
	Port         string        `env:"LINK_HTTP_PORT" env-default:":8080"`
	ReadTimeout  time.Duration `env:"LINK_HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"LINK_HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"LINK_HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

// PostgresConfig holds PostgreSQL connection pool configuration.
type PostgresConfig struct {
	DSN      string `env:"POSTGRES_DSN" env-default:"postgres://postgres:postgres@localhost:5432/shortener?sslmode=disable"`
	MaxConns int32  `env:"POSTGRES_MAX_CONNS" env-default:"20"`
	MinConns int32  `env:"POSTGRES_MIN_CONNS" env-default:"5"`
}

// RedisConfig holds Redis cache connection and cache TTL configuration.
type RedisConfig struct {
	Addr     string        `env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string        `env:"REDIS_PASSWORD" env-default:""`
	LinkTTL  time.Duration `env:"REDIS_LINK_TTL" env-default:"24h"`
}

// KafkaConfig holds Kafka producer settings for publishing events.
type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC_CLICKS" env-default:"link.clicks"`
}

// AnalyticsGRPCConfig holds gRPC connection settings for communicating with analytics-service.
type AnalyticsGRPCConfig struct {
	Addr string `env:"ANALYTICS_GRPC_ADDR" env-default:"localhost:50051"`
}

// OTelConfig holds OpenTelemetry tracing configuration.
type OTelConfig struct {
	Endpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" env-default:"localhost:4317"`
	ServiceName string `env:"OTEL_SERVICE_NAME_LINK" env-default:"link-service"`
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() (*Config, error) {
	var cfg Config

	_ = cleanenv.ReadConfig(".env", &cfg)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	return &cfg, nil
}
