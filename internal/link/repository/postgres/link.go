package postgres

import (
	"context"
	"errors"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)


type LinkRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

const ErrUniqueViolationPg string = "23505"

func (repo *LinkRepo) Get(ctx context.Context, shortCode string) (*domain.Link, error) {
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
		return nil, err
	}

	return &link, nil
}

func (repo *LinkRepo) Create(ctx context.Context, link *domain.Link) error {
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
		return err
	}
	return nil
}
