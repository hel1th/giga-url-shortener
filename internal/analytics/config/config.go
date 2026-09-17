// Package config manages configuration loading for analytics-service.
package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the complete application configuration for analytics-service.
type Config struct {
	Env        string `env:"APP_ENV" env-default:"local"`
	GRPC       GRPCConfig
	HTTP       HTTPConfig
	ClickHouse ClickHouseConfig
	Kafka      KafkaConfig
	OTel       OTelConfig
}

// GRPCConfig holds gRPC server configuration.
type GRPCConfig struct {
	Port string `env:"ANALYTICS_GRPC_PORT" env-default:":50051"`
}

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Port         string        `env:"ANALYTICS_HTTP_PORT" env-default:":8082"`
	ReadTimeout  time.Duration `env:"ANALYTICS_HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"ANALYTICS_HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"ANALYTICS_HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

// ClickHouseConfig holds connection and batching settings for ClickHouse.
type ClickHouseConfig struct {
	Addr          string        `env:"CLICKHOUSE_ADDR" env-default:"localhost:9000"`
	DB            string        `env:"CLICKHOUSE_DB" env-default:"analytics"`
	User          string        `env:"CLICKHOUSE_USER" env-default:"default"`
	Password      string        `env:"CLICKHOUSE_PASSWORD" env-default:"default"`
	BatchSize     int           `env:"CLICKHOUSE_BATCH_SIZE" env-default:"500"`
	FlushInterval time.Duration `env:"CLICKHOUSE_FLUSH_INTERVAL" env-default:"2s"`
}

// KafkaConfig holds Kafka consumer connection settings.
type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC_CLICKS" env-default:"link.clicks"`
	GroupID string   `env:"KAFKA_CONSUMER_GROUP" env-default:"analytics-group"`
}

// OTelConfig holds OpenTelemetry tracing configuration.
type OTelConfig struct {
	Endpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" env-default:"localhost:4317"`
	ServiceName string `env:"OTEL_SERVICE_NAME_ANALYTICS" env-default:"analytics-service"`
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() (*Config, error) {
	var cfg Config

	_ = cleanenv.ReadConfig(".env", &cfg)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read analytics env config: %w", err)
	}

	return &cfg, nil
}
