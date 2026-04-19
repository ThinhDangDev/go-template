# go-template CLI Generator - Implementation Plan

---
title: "go-template CLI Generator"
description: "Production-ready Go backend boilerplate CLI that generates Clean Architecture projects"
status: in-progress
progress: "100% (phases 1-10 complete)"
priority: P1
effort: 32h
branch: main
tags: [go, cli, generator, clean-architecture, boilerplate]
created: 2026-04-16
last-updated: 2026-04-16T19:36:38Z
---

## Overview

Build `go-template` - a binary CLI tool that initializes production-ready Go backend projects with Clean Architecture, flexible auth (JWT/OAuth2), monitoring (Prometheus+Grafana), and comprehensive testing.

## Architecture

```
go-template/
├── cmd/                     # Cobra CLI commands
│   ├── root.go
│   ├── init.go
│   └── version.go
├── internal/
│   ├── generator/           # Core generation engine
│   ├── config/              # Project configuration
│   └── prompt/              # Interactive prompts
├── templates/               # Embedded template files
│   ├── base/
│   ├── clean-arch/
│   ├── docker/
│   ├── monitoring/
│   ├── ci-cd/
│   └── tests/
├── main.go
├── go.mod
└── Makefile
```

## Phase Summary

| Phase | Name | Status | Effort | Dependencies |
|-------|------|--------|--------|--------------|
| 01 | Project Foundation | ✅ COMPLETE | 3h | None |
| 02 | CLI Framework Setup | ✅ COMPLETE | 3h | Phase 01 |
| 03 | Template Engine | ✅ COMPLETE | 4h | Phase 02 |
| 04 | Base Templates | ✅ COMPLETE | 4h | Phase 03 |
| 05 | Clean Architecture Templates | ✅ COMPLETE | 5h | Phase 04 |
| 06 | Authentication Templates | ✅ COMPLETE | 3h | Phase 05 |
| 07 | Docker & Monitoring | ✅ COMPLETE | 3h | Phase 05 |
| 08 | CI/CD & Testing Templates | ✅ COMPLETE | 3h | Phase 07 |
| 09 | CLI Polish & Validation | ✅ COMPLETE | 2h | Phase 08 |
| 10 | Release & Distribution | ✅ COMPLETE | 2h | Phase 09 |

## Key Decisions

1. **Template Engine**: Use `text/template` with `embed.FS` for zero-dependency templates
2. **CLI Framework**: Cobra + Viper for industry-standard CLI experience
3. **Interactive Prompts**: Survey v2 for rich terminal prompts
4. **No External Dependencies in Generated Code**: All generated imports are Go standard library or well-maintained packages

## Success Criteria

- [x] `go install github.com/user/go-template@latest` works
- [x] `go-template init my-project` generates code (with known issues)
- [x] Generated tests pass on first run
- [x] Docker compose starts all services (app, postgres, prometheus, grafana)
- [ ] Generated code passes `golangci-lint run`
- [x] README documentation complete and accurate

## Current Issues Blocking Release

1. **Code Review Critical Issues** - 3 critical issues from Phase 05 review noted for future fix
2. **Release Validation Gaps** - Hosted CI/release execution and cross-platform install confirmation still need evidence outside local development

## Risk Areas

| Risk | Mitigation |
|------|------------|
| Template complexity | Start with minimal templates, iterate |
| Cross-platform paths | Use `filepath` package consistently |
| Embed FS limitations | Test embedding early in Phase 03 |
| gRPC complexity | Make gRPC optional, REST is default |

## Phase Details

- [Phase 01: Project Foundation](./phase-01-project-foundation.md)
- [Phase 02: CLI Framework Setup](./phase-02-cli-framework.md)
- [Phase 03: Template Engine](./phase-03-template-engine.md)
- [Phase 04: Base Templates](./phase-04-base-templates.md)
- [Phase 05: Clean Architecture Templates](./phase-05-clean-architecture.md)
- [Phase 06: Authentication Templates](./phase-06-authentication.md)
- [Phase 07: Docker & Monitoring](./phase-07-docker-monitoring.md)
- [Phase 08: CI/CD & Testing Templates](./phase-08-cicd-testing.md)
- [Phase 09: CLI Polish & Validation](./phase-09-cli-polish.md)
- [Phase 10: Release & Distribution](./phase-10-release.md)

## Quick Start (After Implementation)

```bash
# Install
go install github.com/user/go-template@latest

# Create new project
go-template init my-service

# Interactive prompts:
# - Project name: my-service
# - Module path: github.com/user/my-service
# - API type: REST (default) / gRPC
# - Auth: JWT / OAuth2 / Both
# - Database: PostgreSQL (default)

# Generated project ready to run:
cd my-service
make docker-up   # Starts postgres, prometheus, grafana
make run         # Starts the service
make test        # Runs all tests
```
