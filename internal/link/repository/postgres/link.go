// Package postgres provides PostgreSQL persistence for short links.
package postgres

import (
	"context"
	"errors"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const tracerName = "repo.postgres"

var tracer = otel.Tracer(tracerName)

// LinkRepo implements usecase.LinkRepository backed by PostgreSQL.
type LinkRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRepo creates a new LinkRepo with the provided pgx connection pool.
func NewPostgresRepo(pool *pgxpool.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

// ErrUniqueViolationPg is the PostgreSQL error code for unique key violation (23505).
const ErrUniqueViolationPg string = "23505"

// Get retrieves a link entity by its unique short code.
func (repo *LinkRepo) Get(ctx context.Context, shortCode string) (*domain.Link, error) {
	ctx, span := tracer.Start(ctx, "Get")
	defer span.End()

	const q = `
	SELECT id, short_code, original_url, created_at
	FROM links
	WHERE short_code = $1
	`

	var link domain.Link
	err := repo.pool.QueryRow(ctx, q, shortCode).Scan(
		&link.ID,
		&link.ShortCode,
		&link.OriginalURL,
		&link.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &link, nil
}

// Create inserts a new link record into the links table.
func (repo *LinkRepo) Create(ctx context.Context, link *domain.Link) error {
	ctx, span := tracer.Start(ctx, "Create")
	defer span.End()

	const q = `
	INSERT INTO links
		(id, short_code, original_url, created_at)
	VALUES ($1, $2, $3, $4)
	`
	_, err := repo.pool.Exec(ctx, q, link.ID, link.ShortCode, link.OriginalURL, link.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrUniqueViolationPg {
			return domain.ErrAlreadyExists
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
