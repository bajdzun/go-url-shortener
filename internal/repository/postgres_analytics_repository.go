package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bajdzun/go-url-shortener/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAnalyticsRepository(pool *pgxpool.Pool) *PostgresAnalyticsRepository {
	return &PostgresAnalyticsRepository{pool: pool}
}

func (r *PostgresAnalyticsRepository) RecordClick(ctx context.Context, analytics *domain.Analytics) error {
	query := `
		INSERT INTO url_analytics (short_code, clicked_at, ip_address, user_agent, referer, country)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		analytics.ShortCode,
		time.Now(),
		analytics.IPAddress,
		analytics.UserAgent,
		analytics.Referer,
		analytics.Country,
	)

	return err
}

func (r *PostgresAnalyticsRepository) GetStats(ctx context.Context, shortCode string) (*domain.URLStats, error) {
	query := `
		SELECT
			u.short_code,
			u.original_url,
			u.click_count,
			u.created_at,
			MAX(a.clicked_at) as last_clicked
		FROM urls u
		LEFT JOIN url_analytics a ON u.short_code = a.short_code
		WHERE u.short_code = $1
		GROUP BY u.short_code, u.original_url, u.click_count, u.created_at
	`

	stats := &domain.URLStats{}
	var lastClicked *time.Time

	err := r.pool.QueryRow(ctx, query, shortCode).Scan(
		&stats.ShortCode,
		&stats.OriginalURL,
		&stats.ClickCount,
		&stats.CreatedAt,
		&lastClicked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}

		return nil, err
	}

	stats.LastClicked = lastClicked

	return stats, nil
}
