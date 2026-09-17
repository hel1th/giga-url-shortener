package kafka

import (
	"context"

	"github.com/hel1th/giga-url-shortener/internal/broker"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
	topic  string
}

func NewProducer(client *kgo.Client, topic string) *Producer {
	return &Producer{
		client: client,
		topic:  topic,
	}
}

func (p *Producer) Publish(ctx context.Context, event broker.Envelope) error {
	headers := make([]kgo.RecordHeader, 0, 3)
	headers = append(headers,
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

	// TODO: make it produce asynchronously
	results := p.client.ProduceSync(ctx, record)

	return results.FirstErr()
}
