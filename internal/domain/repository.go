package domain

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrInvalidURL       = errors.New("invalid url")
	ErrShortCodeExists  = errors.New("short code already exists")
	ErrExpiredURL       = errors.New("url has expired")
)

type URLRepository interface {
	Create(ctx context.Context, url *URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*URL, error)
	Update(ctx context.Context, url *URL) error
	Delete(ctx context.Context, shortCode string) error
	IncrementClickCount(ctx context.Context, shortCode string) error
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, analytics *Analytics) error
	GetStats(ctx context.Context, shortCode string) (*URLStats, error)
}
