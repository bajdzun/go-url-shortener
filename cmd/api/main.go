package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/bajdzun/go-url-shortener/internal/config"
	"github.com/bajdzun/go-url-shortener/internal/handler"
	custommiddleware "github.com/bajdzun/go-url-shortener/internal/middleware"
	"github.com/bajdzun/go-url-shortener/internal/repository"
	"github.com/bajdzun/go-url-shortener/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging.Level)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting URL shortener service",
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize database connection
	dbPool, err := initDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()
	logger.Info("connected to PostgreSQL")

	// Initialize Redis connection
	redisClient := initRedis(cfg.Redis)
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	logger.Info("connected to Redis")

	// Initialize repositories
	urlRepo := repository.NewPostgresURLRepository(dbPool)
	cacheRepo := repository.NewRedisCache(redisClient, cfg.Redis.TTL)
	analyticsRepo := repository.NewPostgresAnalyticsRepository(dbPool)

	// Initialize services
	urlService := service.NewURLService(urlRepo, cacheRepo, analyticsRepo, logger, cfg.App.BaseURL)

	// Initialize handlers
	urlHandler := handler.NewURLHandler(urlService, logger)
	healthHandler := handler.NewHealthHandler()

	// Initialize rate limiter
	rateLimiter := custommiddleware.NewRateLimiter(cfg.RateLimit.Requests, cfg.RateLimit.Requests/10)

	// Setup router
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(custommiddleware.LoggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(rateLimiter.Middleware)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	if cfg.Metrics.Enabled {
		r.Use(custommiddleware.MetricsMiddleware)
	}

	// Health check endpoint
	r.Get("/health", healthHandler.Health)

	// Metrics endpoint
	if cfg.Metrics.Enabled {
		r.Handle("/metrics", promhttp.Handler())
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/shorten", urlHandler.CreateShortURL)
		r.Get("/stats/{shortCode}", urlHandler.GetStats)
		r.Delete("/urls/{shortCode}", urlHandler.DeleteURL)
	})

	// Redirect route (should be last)
	r.Get("/{shortCode}", urlHandler.RedirectToOriginalURL)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		logger.Info("shutting down server...")

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal("server shutdown failed", zap.Error(err))
		}
		serverStopCtx()
	}()

	// Run the server
	logger.Info("server started", zap.String("address", srv.Addr))
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("server failed to start", zap.Error(err))
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	logger.Info("server stopped")
}

func initLogger(level string) (*zap.Logger, error) {
	var cfg zap.Config
	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	return cfg.Build()
}

func initDatabase(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
