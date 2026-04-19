# go-template

Production-ready Go backend boilerplate CLI generator.

## Installation

```bash
go install github.com/ThinhDangDev/go-template@latest
```

Binary releases are also published through GitHub Releases.

## Quick Start

```bash
go-template init my-project
cd my-project
make run
```

## Usage

```bash
go-template init my-project
go-template init my-project --skip-validate
go-template version
go-template completion zsh > "${fpath[1]}/_go-template"
```

## Features

- Clean Architecture structure
- REST API (Gin) or gRPC
- PostgreSQL with GORM + migrations
- JWT and/or OAuth2 authentication
- Prometheus + Grafana monitoring
- Docker + Docker Compose
- CI/CD templates (GitHub Actions, GitLab CI)
- Comprehensive testing (testify + testcontainers)
- Shell completion generation
- Post-generation validation (`go mod tidy`, `go build`, `go test`)

## Release

Release assets for this CLI are defined in:

- `.goreleaser.yml`
- `.github/workflows/release.yml`
- `scripts/install.sh`
- `CHANGELOG.md`

## License

MIT
