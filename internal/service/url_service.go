package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"time"

	"github.com/bajdzun/go-url-shortener/internal/domain"
	"go.uber.org/zap"
)

type URLService struct {
	urlRepo       domain.URLRepository
	cacheRepo     domain.CacheRepository
	analyticsRepo domain.AnalyticsRepository
	logger        *zap.Logger
	baseURL       string
}

func NewURLService(
	urlRepo domain.URLRepository,
	cacheRepo domain.CacheRepository,
	analyticsRepo domain.AnalyticsRepository,
	logger *zap.Logger,
	baseURL string,
) *URLService {
	return &URLService{
		urlRepo:       urlRepo,
		cacheRepo:     cacheRepo,
		analyticsRepo: analyticsRepo,
		logger:        logger,
		baseURL:       baseURL,
	}
}

type CreateURLRequest struct {
	OriginalURL string                 `json:"original_url"`
	CustomCode  string                 `json:"custom_code,omitempty"`
	ExpiresIn   *int64                 `json:"expires_in,omitempty"` // seconds
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreateURLResponse struct {
	ShortCode   string                 `json:"short_code"`
	ShortURL    string                 `json:"short_url"`
	OriginalURL string                 `json:"original_url"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (s *URLService) CreateShortURL(ctx context.Context, req *CreateURLRequest) (*CreateURLResponse, error) {
	if !s.isValidURL(req.OriginalURL) {
		return nil, domain.ErrInvalidURL
	}

	var shortCode string
	var err error

	if req.CustomCode != "" {
		shortCode = req.CustomCode
		existing, _ := s.urlRepo.GetByShortCode(ctx, shortCode)
		if existing != nil {
			return nil, domain.ErrShortCodeExists
		}
	} else {
		shortCode = s.generateShortCode(req.OriginalURL)
		for {
			existing, _ := s.urlRepo.GetByShortCode(ctx, shortCode)
			if existing == nil {
				break
			}

			shortCode = s.generateShortCode(req.OriginalURL + time.Now().String())
		}
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expTime := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Second)
		expiresAt = &expTime
	}

	urlEntity := &domain.URL{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		ClickCount:  0,
		Metadata:    req.Metadata,
	}

	if err = s.urlRepo.Create(ctx, urlEntity); err != nil {
		s.logger.Error("failed to create URL", zap.Error(err))

		return nil, err
	}

	if err = s.cacheRepo.Set(ctx, shortCode, req.OriginalURL); err != nil {
		s.logger.Warn("failed to cache URL", zap.Error(err))
	}

	return &CreateURLResponse{
		ShortCode:   shortCode,
		ShortURL:    s.baseURL + "/" + shortCode,
		OriginalURL: req.OriginalURL,
		CreatedAt:   urlEntity.CreatedAt,
		ExpiresAt:   expiresAt,
		Metadata:    req.Metadata,
	}, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string, analytics *domain.Analytics) (string, error) {
	cachedURL, err := s.cacheRepo.Get(ctx, shortCode)
	if err == nil && cachedURL != "" {
		go func() {
			analytics.ShortCode = shortCode
			if err := s.analyticsRepo.RecordClick(context.Background(), analytics); err != nil {
				s.logger.Error("failed to record analytics", zap.Error(err))
			}
			if err := s.urlRepo.IncrementClickCount(context.Background(), shortCode); err != nil {
				s.logger.Error("failed to increment click count", zap.Error(err))
			}
		}()

		return cachedURL, nil
	}

	urlEntity, err := s.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			return "", domain.ErrURLNotFound
		}

    s.logger.Error("failed to get URL", zap.Error(err))

		return "", err
	}

	if urlEntity.ExpiresAt != nil && time.Now().After(*urlEntity.ExpiresAt) {
		return "", domain.ErrExpiredURL
	}

	if err = s.cacheRepo.Set(ctx, shortCode, urlEntity.OriginalURL); err != nil {
		s.logger.Warn("failed to update cache", zap.Error(err))
	}

	go func() {
		analytics.ShortCode = shortCode
		if err := s.analyticsRepo.RecordClick(context.Background(), analytics); err != nil {
			s.logger.Error("failed to record analytics", zap.Error(err))
		}
		if err := s.urlRepo.IncrementClickCount(context.Background(), shortCode); err != nil {
			s.logger.Error("failed to increment click count", zap.Error(err))
		}
	}()

	return urlEntity.OriginalURL, nil
}

func (s *URLService) GetStats(ctx context.Context, shortCode string) (*domain.URLStats, error) {
	stats, err := s.analyticsRepo.GetStats(ctx, shortCode)
	if err != nil {
		s.logger.Error("failed to get stats", zap.Error(err))

		return nil, err
	}

	return stats, nil
}

func (s *URLService) DeleteURL(ctx context.Context, shortCode string) error {
	if err := s.urlRepo.Delete(ctx, shortCode); err != nil {
		s.logger.Error("failed to delete URL", zap.Error(err))

		return err
	}

	if err := s.cacheRepo.Delete(ctx, shortCode); err != nil {
		s.logger.Warn("failed to delete from cache", zap.Error(err))
	}

	return nil
}

func (s *URLService) generateShortCode(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL + time.Now().String()))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return encoded[:7]
}

func (s *URLService) isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}
