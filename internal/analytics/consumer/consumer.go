package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/analytics/domain"
	"github.com/hel1th/giga-url-shortener/internal/analytics/repository/clickhouse"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

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

		if err := c.repo.InsertBatch(ctx, clicks); err != nil {
			c.logger.Error("failed to insert batch into clickhouse", "error", err)
			return
		}

		if err := c.client.CommitRecords(ctx, records...); err != nil {
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
			flush()
			return nil
		case record, ok := <-recordCh:
			if !ok {
				flush()
				return nil
			}

			var event analyticsv1.LinkClicked

			if err := proto.Unmarshal(record.Value, &event); err != nil {
				c.logger.Error("failed to unmarshal kafka record", "error", err)
				continue
			}

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

			fetches.EachRecord(func(record *kgo.Record) {
				select {
				case recordCh <- record:
				case <-ctx.Done():
					return
				}
			})
		}
	}()

	return recordCh
}
