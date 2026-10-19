package consumer

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/analytics/domain"
	"github.com/hel1th/giga-url-shortener/internal/analytics/repository/clickhouse"
	"github.com/hel1th/giga-url-shortener/internal/broker"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

const tracerName = "analytics.consumer"

var tracer = otel.Tracer(tracerName)

type Consumer struct {
	client        *kgo.Client
	repo          clickhouse.Repository
	batchSize     int
	flushInterval time.Duration
	logger        *slog.Logger
}

func NewConsumer(
	client *kgo.Client,
	repo clickhouse.Repository,
	batchSize int,
	flushInterval time.Duration,
	logger *slog.Logger,
) *Consumer {
	return &Consumer{
		client:        client,
		repo:          repo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		logger:        logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	clicks := make([]domain.Click, 0, c.batchSize)
	records := make([]*kgo.Record, 0, c.batchSize)

	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(clicks) == 0 {
			return
		}

		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.repo.InsertBatch(flushCtx, clicks); err != nil {
			c.logger.Error("failed to insert batch into clickhouse", "error", err)
			return
		}

		if err := c.client.CommitRecords(flushCtx, records...); err != nil {
			c.logger.Error("failed to commit kafka offsets", "error", err)
		}

		clicks = clicks[:0]
		records = records[:0]
	}

	defer flush()

	recordCh := c.startReader(ctx)

	for {
		select {
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			return nil
		case record, ok := <-recordCh:
			if !ok {
				return nil
			}

			var event analyticsv1.LinkClicked

			carrier := broker.NewCarrier(record.Headers)
			recordCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
			_, span := tracer.Start(recordCtx, "kafka.Consume")

			if err := proto.Unmarshal(record.Value, &event); err != nil {
				span.RecordError(err)
				span.End()
				c.logger.Error("failed to unmarshal kafka record", "error", err)
				continue
			}
			span.End()

			records = append(records, record)
			clicks = append(clicks, domain.NewClickFromEvent(&event))

			if len(clicks) >= c.batchSize {
				flush()
			}
		}
	}
}

func (c *Consumer) startReader(ctx context.Context) <-chan *kgo.Record {
	recordCh := make(chan *kgo.Record, c.batchSize)

	go func() {
		defer close(recordCh)
		for {
			if ctx.Err() != nil {
				return
			}

			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					if !errors.Is(err.Err, context.Canceled) {
						c.logger.Error("kafka poll error", "topic", err.Topic, "partition", err.Partition, "error", err.Err)
					}
				}
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				select {
				case recordCh <- record:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return recordCh
}
