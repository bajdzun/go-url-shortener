package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bajdzun/go-url-shortener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Create(ctx context.Context, url *domain.URL) error {
	args := m.Called(ctx, url)

	return args.Error(0)
}

func (m *MockURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLRepository) Update(ctx context.Context, url *domain.URL) error {
	args := m.Called(ctx, url)

	return args.Error(0)
}

func (m *MockURLRepository) Delete(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)

	return args.Error(0)
}

func (m *MockURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)

	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)

	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)

	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)

	return args.Error(0)
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)

	return args.Bool(0), args.Error(1)
}

type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) RecordClick(ctx context.Context, analytics *domain.Analytics) error {
	args := m.Called(ctx, analytics)

	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetStats(ctx context.Context, shortCode string) (*domain.URLStats, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.URLStats), args.Error(1)
}

func TestCreateShortURL_Success(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	req := &CreateURLRequest{
		OriginalURL: "https://www.example.com",
	}

	mockURLRepo.On("GetByShortCode", mock.Anything, mock.Anything).Return(nil, domain.ErrURLNotFound)
	mockURLRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(nil)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := service.CreateShortURL(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ShortCode)
	assert.Equal(t, req.OriginalURL, resp.OriginalURL)
	assert.Contains(t, resp.ShortURL, resp.ShortCode)

	mockURLRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

func TestCreateShortURL_InvalidURL(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	req := &CreateURLRequest{
		OriginalURL: "invalid-url",
	}

	resp, err := service.CreateShortURL(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidURL, err)
	assert.Nil(t, resp)
}

func TestCreateShortURL_CustomCode(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	customCode := "custom123"
	req := &CreateURLRequest{
		OriginalURL: "https://www.example.com",
		CustomCode:  customCode,
	}

	mockURLRepo.On("GetByShortCode", mock.Anything, customCode).Return(nil, domain.ErrURLNotFound)
	mockURLRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(nil)
	mockCacheRepo.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := service.CreateShortURL(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, customCode, resp.ShortCode)

	mockURLRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

func TestGetOriginalURL_FromCache(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	shortCode := "abc123"
	expectedURL := "https://www.example.com"

	mockCacheRepo.On("Get", mock.Anything, shortCode).Return(expectedURL, nil)
	mockAnalyticsRepo.On("RecordClick", mock.Anything, mock.AnythingOfType("*domain.Analytics")).Return(nil)
	mockURLRepo.On("IncrementClickCount", mock.Anything, shortCode).Return(nil)

	analytics := &domain.Analytics{
		IPAddress: "127.0.0.1",
	}

	url, err := service.GetOriginalURL(context.Background(), shortCode, analytics)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	// Wait for async goroutine to complete
	time.Sleep(100 * time.Millisecond)

	mockCacheRepo.AssertExpectations(t)
	mockAnalyticsRepo.AssertExpectations(t)
	mockURLRepo.AssertExpectations(t)
}

func TestGetOriginalURL_FromDatabase(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	shortCode := "abc123"
	expectedURL := "https://www.example.com"

	urlEntity := &domain.URL{
		ID:          1,
		ShortCode:   shortCode,
		OriginalURL: expectedURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockCacheRepo.On("Get", mock.Anything, shortCode).Return("", errors.New("not found"))
	mockURLRepo.On("GetByShortCode", mock.Anything, shortCode).Return(urlEntity, nil)
	mockCacheRepo.On("Set", mock.Anything, shortCode, expectedURL).Return(nil)
	mockAnalyticsRepo.On("RecordClick", mock.Anything, mock.AnythingOfType("*domain.Analytics")).Return(nil)
	mockURLRepo.On("IncrementClickCount", mock.Anything, shortCode).Return(nil)

	analytics := &domain.Analytics{
		IPAddress: "127.0.0.1",
	}

	url, err := service.GetOriginalURL(context.Background(), shortCode, analytics)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	// Wait for async goroutine to complete (analytics and increment)
	time.Sleep(100 * time.Millisecond)

	mockCacheRepo.AssertExpectations(t)
	mockURLRepo.AssertExpectations(t)
	mockAnalyticsRepo.AssertExpectations(t)
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	shortCode := "notfound"

	mockCacheRepo.On("Get", mock.Anything, shortCode).Return("", errors.New("not found"))
	mockURLRepo.On("GetByShortCode", mock.Anything, shortCode).Return(nil, domain.ErrURLNotFound)

	analytics := &domain.Analytics{
		IPAddress: "127.0.0.1",
	}

	url, err := service.GetOriginalURL(context.Background(), shortCode, analytics)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrURLNotFound, err)
	assert.Empty(t, url)

	mockCacheRepo.AssertExpectations(t)
	mockURLRepo.AssertExpectations(t)
}

func TestGetOriginalURL_Expired(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	logger := zap.NewNop()

	service := NewURLService(mockURLRepo, mockCacheRepo, mockAnalyticsRepo, logger, "http://localhost:8080")

	shortCode := "expired"
	expiresAt := time.Now().Add(-1 * time.Hour)

	urlEntity := &domain.URL{
		ID:          1,
		ShortCode:   shortCode,
		OriginalURL: "https://www.example.com",
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		UpdatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   &expiresAt,
	}

	mockCacheRepo.On("Get", mock.Anything, shortCode).Return("", errors.New("not found"))
	mockURLRepo.On("GetByShortCode", mock.Anything, shortCode).Return(urlEntity, nil)

	analytics := &domain.Analytics{
		IPAddress: "127.0.0.1",
	}

	url, err := service.GetOriginalURL(context.Background(), shortCode, analytics)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrExpiredURL, err)
	assert.Empty(t, url)

	mockCacheRepo.AssertExpectations(t)
	mockURLRepo.AssertExpectations(t)
}
