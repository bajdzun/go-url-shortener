package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	Logging   LoggingConfig
	Metrics   MetricsConfig
}

type AppConfig struct {
	Env     string
	Port    string
	BaseURL string
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	SSLMode        string
	MaxConnections int
	MaxIdleConns   int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	TTL      time.Duration
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

type MetricsConfig struct {
	Enabled bool
}

func Load() (*Config, error) {
	// Load .env file if it exists (optional - ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:     os.Getenv("APP_ENV"),
			Port:    os.Getenv("APP_PORT"),
			BaseURL: os.Getenv("APP_BASE_URL"),
		},
		Database: DatabaseConfig{
			Host:           os.Getenv("DB_HOST"),
			Port:           os.Getenv("DB_PORT"),
			User:           os.Getenv("DB_USER"),
			Password:       os.Getenv("DB_PASSWORD"),
			Name:           os.Getenv("DB_NAME"),
			SSLMode:        os.Getenv("DB_SSL_MODE"),
			MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS"),
			MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNECTIONS"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvAsInt("REDIS_DB"),
			TTL:      time.Duration(getEnvAsInt("REDIS_TTL")) * time.Second,
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS"),
			Window:   time.Duration(getEnvAsInt("RATE_LIMIT_WINDOW")) * time.Second,
		},
		Logging: LoggingConfig{
			Level:  os.Getenv("LOG_LEVEL"),
			Format: os.Getenv("LOG_FORMAT"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED"),
		},
	}

	return cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnvAsInt(key string) int {
	valueStr := os.Getenv(key)
	value, _ := strconv.Atoi(valueStr)

	return value
}

func getEnvAsBool(key string) bool {
	valueStr := os.Getenv(key)
	value, _ := strconv.ParseBool(valueStr)

	return value
}
