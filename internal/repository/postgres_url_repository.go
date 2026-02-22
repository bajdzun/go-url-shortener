package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bajdzun/go-url-shortener/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresURLRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresURLRepository(pool *pgxpool.Pool) *PostgresURLRepository {
	return &PostgresURLRepository{pool: pool}
}

func (r *PostgresURLRepository) Create(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at, updated_at, expires_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var metadataJSON []byte
	var err error
	if url.Metadata != nil {
		metadataJSON, err = json.Marshal(url.Metadata)
		if err != nil {
			return err
		}
	}

	err = r.pool.QueryRow(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
		url.CreatedAt,
		url.UpdatedAt,
		url.ExpiresAt,
		metadataJSON,
	).Scan(&url.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, updated_at, expires_at, click_count, metadata
		FROM urls
		WHERE short_code = $1
	`

	url := &domain.URL{}
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ExpiresAt,
		&url.ClickCount,
		&metadataJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}

		return nil, err
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &url.Metadata); err != nil {
			return nil, err
		}
	}

	return url, nil
}

func (r *PostgresURLRepository) Update(ctx context.Context, url *domain.URL) error {
	query := `
		UPDATE urls
		SET original_url = $1, updated_at = $2, expires_at = $3, metadata = $4
		WHERE short_code = $5
	`

	var metadataJSON []byte
	var err error
	if url.Metadata != nil {
		metadataJSON, err = json.Marshal(url.Metadata)
		if err != nil {
			return err
		}
	}

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		url.OriginalURL,
		time.Now(),
		url.ExpiresAt,
		metadataJSON,
		url.ShortCode,
	)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (r *PostgresURLRepository) Delete(ctx context.Context, shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = $1`

	cmdTag, err := r.pool.Exec(ctx, query, shortCode)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (r *PostgresURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `
		UPDATE urls
		SET click_count = click_count + 1
		WHERE short_code = $1
	`

	_, err := r.pool.Exec(ctx, query, shortCode)
	return err
}
