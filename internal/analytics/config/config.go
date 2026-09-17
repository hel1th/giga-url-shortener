package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"APP_ENV" env-default:"local"`
	GRPC       GRPCConfig
	HTTP       HTTPConfig
	ClickHouse ClickHouseConfig
	Kafka      KafkaConfig
	OTel       OTelConfig
}

type GRPCConfig struct {
	Port string `env:"ANALYTICS_GRPC_PORT" env-default:":50051"`
}

type HTTPConfig struct {
	Port         string        `env:"ANALYTICS_HTTP_PORT" env-default:":8082"`
	ReadTimeout  time.Duration `env:"ANALYTICS_HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"ANALYTICS_HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"ANALYTICS_HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type ClickHouseConfig struct {
	Addr          string        `env:"CLICKHOUSE_ADDR" env-default:"localhost:9000"`
	DB            string        `env:"CLICKHOUSE_DB" env-default:"analytics"`
	User          string        `env:"CLICKHOUSE_USER" env-default:"default"`
	Password      string        `env:"CLICKHOUSE_PASSWORD" env-default:"default"`
	BatchSize     int           `env:"CLICKHOUSE_BATCH_SIZE" env-default:"500"`
	FlushInterval time.Duration `env:"CLICKHOUSE_FLUSH_INTERVAL" env-default:"2s"`
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC_CLICKS" env-default:"link.clicks"`
	GroupID string   `env:"KAFKA_CONSUMER_GROUP" env-default:"analytics-group"`
}

type OTelConfig struct {
	Endpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" env-default:"localhost:4317"`
	ServiceName string `env:"OTEL_SERVICE_NAME_ANALYTICS" env-default:"analytics-service"`
}


func Load() (*Config, error) {
    var cfg Config

    _ = cleanenv.ReadConfig(".env", &cfg)

    if err := cleanenv.ReadEnv(&cfg); err != nil {
    	return nil, fmt.Errorf("failed to read analytics env config: %w", err)
    }

    return &cfg, nil
}
