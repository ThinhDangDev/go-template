# go-template

`go-template` is a generator CLI for bootstrapping Go backend services.

## Usage

```bash
go-template init my-service
go-template init my-service --module github.com/acme/my-service
```

The generated project includes:

- CLI-first app commands: `serve`, `migrate`, `seed`
- manual SQL migrations with `up/down/status/create`
- JWT authentication
- Casbin RBAC
- Prometheus metrics
- OpenTelemetry tracing
- structured JSON logging
- Docker + docker-compose

## Local Development

```bash
go build ./cmd/main.go
go test ./...
go run ./cmd/main.go init demo-api
```

## Generated Project Flow

After running `go-template init <project-name>`:

```bash
cd <project-name>
cp .env.example .env
```

Set at least:

```bash
JWT_SECRET=replace-with-a-long-random-secret
ADMIN_PASSWORD=ChangeMe123!
```

Then run:

```bash
make migrate-up
make seed-admin
make run
```
