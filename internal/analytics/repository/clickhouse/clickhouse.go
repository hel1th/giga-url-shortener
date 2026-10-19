// Package clickhouse provides ClickHouse repository implementation for analytics data.
package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/hel1th/giga-url-shortener/internal/analytics/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "repo.clickhouse"

var tracer = otel.Tracer(tracerName)

type ClicksRepository struct {
	client clickhouse.Conn
}

func NewClicksRepository(client clickhouse.Conn) *ClicksRepository {
	return &ClicksRepository{client: client}
}

type Repository interface {
	GetStats(ctx context.Context, shortCode string) (*domain.LinkSummary, error)
	InsertBatch(ctx context.Context, clicks []domain.Click) error
}

func (r *ClicksRepository) InsertBatch(ctx context.Context, clicks []domain.Click) error {
	ctx, span := tracer.Start(ctx, "clickhouse.InsertBatch")
	defer span.End()

	const q = `
	INSERT INTO analytics.link_clicks (
		short_code,
		clicked_at,
		ip,
		user_agent,
		browser,
		os,
		device,
		referer
	) VALUES`

	batch, err := r.client.PrepareBatch(ctx, q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	for _, click := range clicks {
		err := batch.Append(
			click.ShortCode,
			click.ClickedAt,
			click.IP,
			click.UserAgent,
			click.Browser,
			click.OS,
			click.Device,
			click.Referer,
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	if err := batch.Send(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (r *ClicksRepository) GetStats(ctx context.Context, shortCode string) (*domain.LinkSummary, error) {
	ctx, span := tracer.Start(ctx, "clickhouse.GetStats")
	defer span.End()

	const qCount = `
	SELECT count(*), uniqExact(ip)
	FROM analytics.link_clicks
	WHERE short_code = ?
	`
	var summary domain.LinkSummary

	err := r.client.QueryRow(ctx, qCount, shortCode).Scan(&summary.TotalClicks, &summary.UniqueClicks)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if summary.TotalClicks == 0 {
		return &summary, nil
	}

	summary.Browser, err = r.fetchDistribution(ctx, "browser", shortCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	summary.OS, err = r.fetchDistribution(ctx, "os", shortCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	summary.Device, err = r.fetchDistribution(ctx, "device", shortCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &summary, nil
}

func (r *ClicksRepository) fetchDistribution(
	ctx context.Context,
	column string,
	shortCode string,
) (map[string]uint64, error) {
	switch column {
	case "browser", "os", "device":

	default:
		return nil, fmt.Errorf("invalid column: %s", column)
	}

	q := fmt.Sprintf(
		`
    SELECT %s, count(*) as cnt
    FROM analytics.link_clicks
    WHERE short_code = ?
    GROUP BY %s
    ORDER BY cnt DESC`,
		column, column,
	)

	rows, err := r.client.Query(ctx, q, shortCode)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	distr := make(map[string]uint64)

	var key string
	var cnt uint64

	for rows.Next() {
		err := rows.Scan(&key, &cnt)
		if err != nil {
			return nil, err
		}
		distr[key] = cnt
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return distr, nil
}
