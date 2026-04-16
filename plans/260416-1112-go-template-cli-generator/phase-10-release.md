# Phase 10: Release & Distribution

---
status: not-started
priority: P2
effort: 2h
dependencies: [phase-09]
blocked-by: "Phase 09 (CLI Polish) not started"
---

## Context Links

- [Main Plan](./plan.md)
- [Previous: CLI Polish](./phase-09-cli-polish.md)

## Overview

Set up release automation with GoReleaser, GitHub releases, and documentation for users to install go-template via multiple methods.

## Key Insights

- GoReleaser handles cross-compilation and release artifacts
- GitHub releases provide binaries for non-Go users
- `go install` is the simplest method for Go developers
- Checksums verify download integrity
- CHANGELOG.md documents releases

## Requirements

### Functional
- GoReleaser configuration for multi-platform builds
- GitHub releases with binaries and checksums
- Installation documentation
- CHANGELOG.md template
- Version command shows build info

### Non-Functional
- Release process automated via GitHub Actions
- Reproducible builds
- Small binary size (< 15MB)

## Architecture

```
go-template/
├── .goreleaser.yml          # GoReleaser config
├── CHANGELOG.md             # Release history
├── LICENSE                  # MIT license
├── scripts/
│   └── install.sh           # One-liner install script
└── .github/
    └── workflows/
        └── release.yml      # Release automation (already created)
```

## Related Code Files

### Files to Create
- `.goreleaser.yml`
- `CHANGELOG.md`
- `LICENSE`
- `scripts/install.sh`
- Update `README.md` with installation instructions

## Implementation Steps

### Step 1: Create GoReleaser Config

```yaml
# .goreleaser.yml
project_name: go-template

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: go-template
    main: ./main.go
    binary: go-template
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.buildDate={{.Date}}

archives:
  - id: default
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

snapshot:
  name_template: "{{ .Tag }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
      - Merge pull request
      - Merge branch

release:
  github:
    owner: thinhdang
    name: go-template
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"
  header: |
    ## What's Changed

    See the [CHANGELOG](CHANGELOG.md) for details.

    ## Installation

    ### Go Install (recommended)
    ```bash
    go install github.com/thinhdang/go-template@{{.Tag}}
    ```

    ### Download Binary
    Download the appropriate binary for your platform from the assets below.

brews:
  - name: go-template
    repository:
      owner: thinhdang
      name: homebrew-tap
    folder: Formula
    homepage: https://github.com/thinhdang/go-template
    description: Go backend boilerplate CLI generator
    license: MIT
    test: |
      system "#{bin}/go-template", "version"
    install: |
      bin.install "go-template"
    # Skip if no homebrew tap repo exists
    skip_upload: auto

# Future: Chocolatey, Scoop, etc.
```

### Step 2: Create CHANGELOG

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of go-template CLI
- Interactive project initialization
- Clean Architecture template structure
- REST API (Gin) support
- gRPC support (optional)
- JWT authentication templates
- OAuth2 authentication templates (Google provider)
- PostgreSQL with GORM
- Database migrations with golang-migrate
- Docker and Docker Compose configuration
- Prometheus metrics integration
- Grafana dashboard provisioning
- GitHub Actions CI/CD workflows
- GitLab CI configuration
- Comprehensive test scaffolding with testify
- Integration tests with testcontainers
- Shell completion (bash, zsh, fish, powershell)

## [0.1.0] - YYYY-MM-DD

### Added
- First public release

[Unreleased]: https://github.com/thinhdang/go-template/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/thinhdang/go-template/releases/tag/v0.1.0
```

### Step 3: Create LICENSE

```
MIT License

Copyright (c) 2026 Thinh Dang

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Step 4: Create Install Script

```bash
#!/bin/bash
# scripts/install.sh
# One-liner installation script for go-template

set -e

REPO="thinhdang/go-template"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="go-template"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest release
if [ -z "$VERSION" ]; then
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi

if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

# Download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"

echo "Installing ${BINARY_NAME} ${VERSION}..."
echo "  OS: ${OS}"
echo "  Arch: ${ARCH}"
echo "  URL: ${DOWNLOAD_URL}"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
curl -sL "$DOWNLOAD_URL" | tar xz -C "$TMP_DIR"

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo ""
echo "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "Run 'go-template --help' to get started"
```

### Step 5: Update README with Installation

```markdown
# go-template

Production-ready Go backend boilerplate CLI generator.

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/thinhdang/go-template)](https://github.com/thinhdang/go-template/releases)

## Installation

### Using Go (recommended)

```bash
go install github.com/thinhdang/go-template@latest
```

### Using Homebrew (macOS/Linux)

```bash
brew install thinhdang/tap/go-template
```

### Using Script

```bash
curl -sSL https://raw.githubusercontent.com/thinhdang/go-template/main/scripts/install.sh | bash
```

### Manual Download

Download the appropriate binary for your platform from the [releases page](https://github.com/thinhdang/go-template/releases).

## Quick Start

```bash
# Create a new project
go-template init my-service

# Follow the interactive prompts to configure:
# - Module path
# - API type (REST/gRPC)
# - Authentication (JWT/OAuth2)
# - Docker & monitoring options

# Start your project
cd my-service
make docker-up    # Start PostgreSQL, Prometheus, Grafana
make run          # Start the service
```

## Generated Project Structure

```
my-service/
├── cmd/my-service/       # Application entry point
├── internal/
│   ├── config/           # Configuration
│   ├── domain/           # Business entities & interfaces
│   ├── usecase/          # Business logic
│   ├── delivery/         # HTTP/gRPC handlers
│   └── infrastructure/   # External implementations
├── migrations/           # Database migrations
├── docker-compose.yml    # Docker services
└── Makefile             # Build & dev commands
```

## Features

- **Clean Architecture** - Domain-driven, testable structure
- **REST API** - Gin framework with middleware
- **gRPC** - Optional gRPC support
- **Authentication** - JWT and/or OAuth2 (Google)
- **Database** - PostgreSQL with GORM ORM
- **Migrations** - golang-migrate for schema management
- **Monitoring** - Prometheus metrics + Grafana dashboards
- **Docker** - Multi-stage Dockerfile + docker-compose
- **CI/CD** - GitHub Actions or GitLab CI
- **Testing** - Unit tests (testify) + integration (testcontainers)
- **Linting** - golangci-lint configuration

## Commands

```bash
go-template init <name>     # Create new project
go-template version         # Show version info
go-template completion      # Generate shell completions
```

### Flags

```bash
--non-interactive    # Use defaults without prompts
--no-color          # Disable colored output
--skip-validation   # Skip post-generation validation
-v, --verbose       # Verbose output
```

## Shell Completion

```bash
# Bash
source <(go-template completion bash)

# Zsh
go-template completion zsh > "${fpath[1]}/_go-template"

# Fish
go-template completion fish > ~/.config/fish/completions/go-template.fish
```

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.
```

### Step 6: Create Version Embedding

Update main.go to properly embed version info:

```go
// main.go
package main

import (
	"os"

	"github.com/thinhdang/go-template/cmd"
)

// These variables are set via ldflags during build
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, buildDate)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

```go
// cmd/version.go
package cmd

var (
	Version   string
	Commit    string
	BuildDate string
)

func SetVersionInfo(version, commit, buildDate string) {
	Version = version
	Commit = commit
	BuildDate = buildDate
}
```

## Todo List

- [ ] Create .goreleaser.yml
- [ ] Create CHANGELOG.md
- [ ] Create LICENSE (MIT)
- [ ] Create scripts/install.sh
- [ ] Update README.md with installation instructions
- [ ] Update main.go for version ldflags
- [ ] Test `goreleaser check`
- [ ] Test `goreleaser release --snapshot --clean`
- [ ] Tag first release (v0.1.0)
- [ ] Verify GitHub release created
- [ ] Verify binaries downloadable
- [ ] Test install script

## Success Criteria

- [ ] `goreleaser check` passes
- [ ] Snapshot release builds all platforms
- [ ] GitHub release contains binaries for linux/darwin/windows
- [ ] Checksums file generated
- [ ] `go install` works from GitHub
- [ ] Install script downloads and installs correctly
- [ ] README installation instructions accurate

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Release automation fails | Low | High | Test with --snapshot first |
| Binary size too large | Low | Low | Use -ldflags="-s -w" |
| Platform compatibility | Low | Medium | Test on each platform |

## Security Considerations

- Checksums for download verification
- HTTPS for all downloads
- No secrets in release artifacts
- Install script verifies checksums (optional enhancement)

## Release Checklist

When releasing a new version:

1. Update CHANGELOG.md with release notes
2. Commit changes
3. Create and push tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
4. Push tag: `git push origin v0.1.0`
5. GitHub Actions automatically creates release
6. Verify release on GitHub
7. Test installation methods

## Future Enhancements

- Homebrew tap repository
- Chocolatey package for Windows
- Scoop manifest
- Docker image for CI usage
- Automated update notifications
