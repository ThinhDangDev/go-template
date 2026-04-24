# __PROJECT_NAME__

`__PROJECT_NAME__` is a CLI-first Go backend template with:

- Gin HTTP server
- Protocol Buffers + grpc-gateway
- native gRPC server bootstrap
- manual SQL migrations with `up/down/status/create`
- JWT authentication
- Casbin RBAC
- structured JSON logging
- Prometheus metrics
- OpenTelemetry tracing
- OpenAPI JSON generation

## Commands

```bash
make proto
make swagger
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
- `GET /swagger.json`
- `GET /api/v1/public/ping`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/admin/ping`
- `GET /api/v1/operator/ping`
- `GET /api/v1/viewer/ping`

HTTP JSON is served by Gin + grpc-gateway on port `8080` by default. Native gRPC is served on port `9090` by default using the same service handlers and auth/RBAC rules.

## Project Layout

```text
cmd/main.go
configs/rbac_model.conf
generate.sh
internal/boilerplate/
  app/
  auth/
  cli/
  config/
  http/
  store/
  telemetry/
internal/docs/
migrations/sql/
proto/
protogen/
third_party/googleapis/
```

## Environment

Required for runtime:

- `JWT_SECRET`

Common optional variables:

- `DATABASE_URL`
- `POSTGRES_*`
- `GRPC_HOST`
- `GRPC_PORT`
- `SWAGGER_JSON_PATH`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `LOG_ENABLE_FILE`

## Docker

```bash
docker compose up -d postgres
JWT_SECRET=replace-with-a-secret go run ./cmd/main.go migrate up
JWT_SECRET=replace-with-a-secret go run ./cmd/main.go seed admin --password ChangeMe123!
docker build -t __PROJECT_NAME__ .
docker run --rm -p 8080:8080 --env-file .env __PROJECT_NAME__
```

The container starts with `serve`; migrations stay manual by design.

## Proto Tooling

Install generators once:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

Then regenerate:

```bash
make proto
```
