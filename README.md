[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-blue)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

# Go URL Shortener

A production-ready URL shortener service built with Go, demonstrating clean architecture principles, best practices, and modern development patterns.

## 🏗️ Architecture

This project follows **Clean Architecture** (Hexagonal/Ports & Adapters) principles:

```
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Business entities and interfaces (ports)
│   ├── service/          # Business logic (use cases)
│   ├── repository/       # Data persistence (adapters)
│   ├── handler/          # HTTP handlers
│   └── middleware/       # HTTP middlewares
├── migrations/           # Database migrations
└── docker-compose.yml    # Docker orchestration
```

## 🚀 Features

### Core Features
- ✅ URL shortening with auto-generated or custom short codes
- ✅ URL redirection with 301 permanent redirects
- ✅ URL expiration support
- ✅ Analytics tracking (clicks, IP, user agent, referer)
- ✅ Statistics API

### Technical Features
- ✅ **Clean Architecture** - Separation of concerns, dependency inversion
- ✅ **PostgreSQL** - Persistent storage with proper indexing
- ✅ **Redis** - Caching layer for performance
- ✅ **Docker & Docker Compose** - Containerization
- ✅ **Prometheus Metrics** - Application monitoring
- ✅ **Structured Logging** - JSON logging with Zap
- ✅ **Rate Limiting** - IP-based rate limiting
- ✅ **Graceful Shutdown** - Proper cleanup on termination
- ✅ **Health Checks** - Container health monitoring
- ✅ **Unit Tests** - Comprehensive test coverage
- ✅ **CORS Support** - Cross-origin resource sharing
- ✅ **Middleware Stack** - Logging, metrics, recovery, timeouts

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for convenience commands)

## 🏃 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/yourusername/go-url-shortener.git
cd go-url-shortener

# Start all services
docker-compose up --build

# Or run in background
docker-compose up -d --build
```

The API will be available at `http://localhost:8080`

### Local Development

```bash
# Install dependencies
go mod download

# Set up environment variables (IMPORTANT!)
cp .env.example .env
# Edit .env with your preferred values (optional)

# Start PostgreSQL and Redis using Docker
docker-compose up postgres redis -d

# Run the application
go run cmd/api/main.go
```

> **Note**: All configuration is managed through the `.env` file. Docker Compose automatically loads values from `.env`.

## 📚 API Documentation

### Create Short URL

**POST** `/api/v1/shorten`

```json
{
  "original_url": "https://www.example.com/very/long/url",
  "custom_code": "mycode",  // Optional
  "expires_in": 86400,      // Optional, seconds
  "metadata": {             // Optional
    "campaign": "summer-sale"
  }
}
```

**Response:**
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://www.example.com/very/long/url",
  "created_at": "2024-02-22T10:30:00Z",
  "expires_at": "2024-02-23T10:30:00Z"
}
```

### Redirect to Original URL

**GET** `/{shortCode}`

Redirects to the original URL with 301 status code.

### Get URL Statistics

**GET** `/api/v1/stats/{shortCode}`

**Response:**
```json
{
  "short_code": "abc123",
  "original_url": "https://www.example.com",
  "click_count": 42,
  "created_at": "2024-02-22T10:30:00Z",
  "last_clicked": "2024-02-22T12:45:00Z"
}
```

### Delete URL

**DELETE** `/api/v1/urls/{shortCode}`

Returns 204 No Content on success.

### Health Check

**GET** `/health`

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

### Metrics

**GET** `/metrics`

Prometheus-formatted metrics for monitoring.

## 🧪 Testing

### In Docker (Recommended)

```bash
# Run all tests
docker-compose exec app make test

# Run tests with coverage
docker-compose exec app make test-coverage

# Run specific package tests
docker-compose exec app go test -v ./internal/service

# View coverage report
open coverage/coverage.html
```

### Locally (if Go installed)

```bash
# Run all tests
make test
# or: go test ./...

# Run tests with coverage
make test-coverage
# or: go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Monitoring

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics`:

- `http_requests_total` - Total HTTP requests by method, path, and status
- `http_request_duration_seconds` - HTTP request latency

Access Prometheus UI at `http://localhost:9090`

### Example Queries

```promql
# Request rate
rate(http_requests_total[5m])

# Average request duration
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])
```

## 🔧 Configuration

Configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | `development` |
| `APP_PORT` | Server port | `8080` |
| `APP_BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `urlshortener` |
| `DB_PASSWORD` | Database password | `secret_password` |
| `DB_NAME` | Database name | `urlshortener` |
| `REDIS_HOST` | Redis host | `redis` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_TTL` | Cache TTL in seconds | `86400` |
| `RATE_LIMIT_REQUESTS` | Max requests per window | `100` |
| `RATE_LIMIT_WINDOW` | Rate limit window in seconds | `60` |
| `LOG_LEVEL` | Logging level (debug/info/error) | `info` |

## 🗄️ Database Schema

### URLs Table
```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count BIGINT DEFAULT 0,
    metadata JSONB
);
```

### Analytics Table
```sql
CREATE TABLE url_analytics (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE
);
```

## 🏛️ Design Patterns & Best Practices

### Clean Architecture
- **Domain Layer**: Business entities and repository interfaces
- **Service Layer**: Business logic and use cases
- **Repository Layer**: Data access implementations
- **Handler Layer**: HTTP transport layer

### SOLID Principles
- **Single Responsibility**: Each package has one reason to change
- **Open/Closed**: Open for extension, closed for modification
- **Liskov Substitution**: Repository interfaces are substitutable
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions

### Additional Patterns
- **Repository Pattern**: Abstract data access
- **Dependency Injection**: Constructor injection
- **Middleware Pattern**: Composable HTTP middlewares
- **Factory Pattern**: Service initialization

## 🔐 Security Features

- ✅ Rate limiting to prevent abuse
- ✅ Input validation
- ✅ SQL injection prevention (parameterized queries)
- ✅ CORS configuration
- ✅ Request timeouts
- ✅ Graceful shutdown to prevent data loss

## 📈 Performance Optimizations

- ✅ Redis caching for frequently accessed URLs
- ✅ Database connection pooling
- ✅ Efficient database indexes
- ✅ Asynchronous analytics recording
- ✅ Volume mounting for live code reload during development

## 🐳 Docker Services

The application stack includes:

- **app**: Go application
- **postgres**: PostgreSQL database
- **redis**: Redis cache
- **prometheus**: Metrics collection

## 🛠️ Development Commands

### Using Docker Compose Directly

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f app

# Rebuild specific service
docker-compose up --build app

# Run tests in Docker
docker-compose exec app go test ./...

# Run make commands in Docker
docker-compose exec app make lint
docker-compose exec app make test

# Access database
docker-compose exec postgres psql -U urlshortener -d urlshortener
```

### Local Development (without Docker)

```bash
# If you have Go installed locally
make test
make lint
make build
```

## 📝 Example Usage

```bash
# Create a short URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://www.example.com/very/long/url/that/needs/shortening"
  }'

# Response: {"short_code":"xY9zK2a","short_url":"http://localhost:8080/xY9zK2a",...}

# Visit the short URL (redirects)
curl -L http://localhost:8080/xY9zK2a

# Get statistics
curl http://localhost:8080/api/v1/stats/xY9zK2a

# Create URL with custom code and expiration
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://www.example.com",
    "custom_code": "example",
    "expires_in": 3600
  }'

# Delete a URL
curl -X DELETE http://localhost:8080/api/v1/urls/example
```

## 📄 License

MIT License
