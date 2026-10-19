// Package kafka provides a Kafka-backed message producer.
package kafka

import (
	"context"
	"log/slog"

	"github.com/hel1th/giga-url-shortener/internal/broker"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer publishes broker envelopes to a designated Kafka topic.
type Producer struct {
	client *kgo.Client
	topic  string
	logger *slog.Logger
}

// NewProducer creates a new Kafka producer instance for a given topic.
func NewProducer(client *kgo.Client, topic string, logger *slog.Logger) *Producer {
	return &Producer{
		client: client,
		topic:  topic,
		logger: logger,
	}
}

// Publish publishes an envelope to Kafka.
func (p *Producer) Publish(ctx context.Context, event broker.Envelope) {
	headers := make([]kgo.RecordHeader, 0, 3)
	headers = append(
		headers,
		kgo.RecordHeader{Key: broker.HeaderEventID, Value: []byte(event.ID)},
		kgo.RecordHeader{Key: broker.HeaderEventType, Value: []byte(event.Type)},
	)

	if event.TraceParent != "" {
		headers = append(headers,
			kgo.RecordHeader{Key: broker.HeaderTraceParent, Value: []byte(event.TraceParent)})
	}

	record := &kgo.Record{
		Topic:     p.topic,
		Key:       []byte(event.Key),
		Value:     event.Payload,
		Timestamp: event.OccurredAt,
		Headers:   headers,
	}

	p.client.Produce(context.WithoutCancel(ctx), record, func(r *kgo.Record, err error) {
		if err != nil {
			p.logger.Error("can't produce record", "record key", string(r.Key), "error", err)
		}
	})
}
