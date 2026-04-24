# go-template

`go-template` is a CLI-first Go backend template with:

- HTTP server bootstrap
- manual SQL migrations with `up/down/status/create`
- JWT authentication
- Casbin RBAC
- structured JSON logging
- Prometheus metrics
- OpenTelemetry tracing

## Commands

```bash
go run ./cmd/main.go serve
go run ./cmd/main.go migrate up
go run ./cmd/main.go migrate down --steps 1
go run ./cmd/main.go migrate status
go run ./cmd/main.go migrate create add_audit_log
go run ./cmd/main.go seed admin --email admin@example.com --password ChangeMe123!
```

## Quick Start

1. Copy `.env.example` to `.env`.
2. Start PostgreSQL.
3. Apply migrations:

```bash
make migrate-up
```

4. Seed the first admin:

```bash
make seed-admin
```

5. Start the API:

```bash
make run
```

`serve` fails fast if the required schema is missing, so migrations must run first.

## Routes

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/admin/ping`
- `GET /api/v1/operator/ping`
- `GET /api/v1/viewer/ping`

## Project Layout

```text
cmd/main.go
configs/rbac_model.conf
internal/boilerplate/
  app/
  auth/
  cli/
  config/
  http/
  store/
  telemetry/
migrations/sql/
```

## Environment

Required for runtime:

- `JWT_SECRET`

Common optional variables:

- `DATABASE_URL`
- `POSTGRES_*`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `LOG_ENABLE_FILE`

## Docker

```bash
docker compose up -d postgres
JWT_SECRET=replace-with-a-secret go run ./cmd/main.go migrate up
JWT_SECRET=replace-with-a-secret go run ./cmd/main.go seed admin --password ChangeMe123!
docker build -t go-template .
docker run --rm -p 8080:8080 --env-file .env go-template
```

The container starts with `serve`; migrations stay manual by design.
