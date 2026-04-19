# Phase 01: Project Foundation

---
status: completed
priority: P1
effort: 3h
dependencies: none
completed: 2026-04-16
---

## Context Links

- [Main Plan](./plan.md)
- [Next Phase: CLI Framework](./phase-02-cli-framework.md)

## Overview

Initialize the go-template project with Go modules, core dependencies, and project structure. This phase establishes the foundation all other phases build upon.

## Key Insights

- Use Go 1.22+ for enhanced template features and embed improvements
- Keep dependencies minimal: Cobra, Viper, Survey are the core three
- Structure follows standard Go project layout conventions

## Requirements

### Functional
- Initialize Go module with proper naming
- Create directory structure for CLI and templates
- Set up Makefile for common development tasks

### Non-Functional
- Go 1.22+ compatibility
- Cross-platform support (Linux, macOS, Windows)

## Architecture

```
go-template/
├── cmd/
│   └── .gitkeep
├── internal/
│   ├── generator/
│   ├── config/
│   └── prompt/
├── templates/
│   └── .gitkeep
├── main.go
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
└── README.md
```

## Related Code Files

### Files to Create
- `go.mod` - Go module definition
- `main.go` - Entry point
- `Makefile` - Build automation
- `.gitignore` - Git ignore rules
- `README.md` - Project documentation
- Directory structure with `.gitkeep` files

## Implementation Steps

### Step 1: Initialize Go Module

```bash
cd /Users/thinhdang/go-boipleplate
go mod init github.com/ThinhDangDev/go-template
```

### Step 2: Create Directory Structure

```bash
mkdir -p cmd internal/{generator,config,prompt} templates
```

### Step 3: Create main.go

```go
// main.go
package main

import (
    "os"

    "github.com/ThinhDangDev/go-template/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Step 4: Install Core Dependencies

```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/AlecAivazis/survey/v2@latest
```

### Step 5: Create Makefile

```makefile
.PHONY: build install test lint clean

BINARY_NAME=go-template
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean

dev: build
	./bin/$(BINARY_NAME)
```

### Step 6: Create .gitignore

```
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
*.out

# Dependency directories
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Local config
.env
*.local
```

### Step 7: Create Initial README

```markdown
# go-template

Production-ready Go backend boilerplate CLI generator.

## Installation

```bash
go install github.com/ThinhDangDev/go-template@latest
```

## Usage

```bash
go-template init my-project
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

## License

MIT
```

## Todo List

- [x] Initialize Go module
- [x] Create directory structure
- [x] Create main.go entry point
- [x] Install Cobra, Viper, Survey dependencies
- [x] Create Makefile with build/test/lint targets
- [x] Create .gitignore
- [x] Create README.md
- [x] Verify `go build` succeeds
- [x] Commit initial structure

## Success Criteria

- [x] `go build` compiles without errors
- [x] `go mod tidy` shows no missing dependencies
- [x] Directory structure matches architecture diagram
- [x] Makefile targets work correctly

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Dependency version conflicts | Low | Medium | Pin versions explicitly |
| Module path issues | Low | High | Test go install early |

## Security Considerations

- No secrets in repository
- .gitignore includes .env and local configs

## Next Steps

After completing this phase:
1. Proceed to [Phase 02: CLI Framework](./phase-02-cli-framework.md)
2. Implement Cobra root command structure
