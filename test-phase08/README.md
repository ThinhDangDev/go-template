# test-phase08

A Go backend service

## Features

- REST API with Gin
- Authentication: jwt
- Database: POSTGRES with GORM
- Docker & Docker Compose
- Prometheus + Grafana monitoring
- CI/CD workflow templates
- Unit and integration test scaffolding
- golangci-lint configuration

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run locally
go run cmd/server/main.go

# Build
go build -o bin/test-phase08 cmd/server/main.go

# Run with Docker
docker-compose up
```

## Project Structure

```
test-phase08/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── domain/          # Business entities
│   ├── usecase/         # Business logic
│   ├── repository/      # Data access
│   └── delivery/        # HTTP/gRPC handlers
├── testutil/            # Shared test helpers and fixtures
├── pkg/                 # Public libraries
├── docker/              # Docker configurations
└── README.md
```

## Development

```bash
# Run unit tests
go test ./...

# Run integration tests (requires Docker)
go test -tags=integration ./...

# Skip integration tests in a quick pass
go test -short ./...

# Lint
golangci-lint run

# Docker build
docker build -t test-phase08 .
```

This example includes generated CI/CD workflow files.
Integration tests use testcontainers to provision PostgreSQL automatically.

## License

MIT

## Generated

This project was generated using [go-template](https://github.com/ThinhDangDev/go-template).
