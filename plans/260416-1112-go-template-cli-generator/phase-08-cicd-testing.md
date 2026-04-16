# Phase 08: CI/CD & Testing Templates

---
status: not-started
priority: P1
effort: 3h
dependencies: [phase-07]
blocked-by: "Phase 07 (Docker & Monitoring) not started"
---

## Context Links

- [Main Plan](./plan.md)
- [Previous: Docker & Monitoring](./phase-07-docker-monitoring.md)
- [Next: CLI Polish](./phase-09-cli-polish.md)

## Overview

Create CI/CD workflow templates (GitHub Actions, GitLab CI) and comprehensive test scaffolding using testify and testcontainers for integration tests.

## Key Insights

- GitHub Actions for most users, GitLab CI for enterprise
- Test structure: unit tests alongside code, integration in _test package
- Testcontainers for real database tests (no mocks for integration)
- golangci-lint for consistent code quality
- Parallel test execution for speed

## Requirements

### Functional
- GitHub Actions: lint, test, build, docker push
- GitLab CI: similar stages
- Unit test examples with testify
- Integration tests with testcontainers
- Test utilities and fixtures

### Non-Functional
- Fast CI pipeline (< 5 minutes)
- Clear test output
- Reproducible tests

## Architecture

```
templates/ci-cd/
├── github/
│   └── workflows/
│       ├── ci.yml.tmpl
│       └── release.yml.tmpl
└── gitlab/
    └── .gitlab-ci.yml.tmpl

templates/tests/
├── internal/
│   ├── usecase/
│   │   └── user_test.go.tmpl
│   └── infrastructure/
│       └── repository/
│           └── user_integration_test.go.tmpl
├── testutil/
│   ├── fixtures.go.tmpl
│   ├── database.go.tmpl
│   └── testcontainers.go.tmpl
└── .golangci.yml.tmpl
```

## Related Code Files

### Files to Create
- `templates/ci-cd/github/workflows/ci.yml.tmpl`
- `templates/ci-cd/github/workflows/release.yml.tmpl`
- `templates/ci-cd/gitlab/.gitlab-ci.yml.tmpl`
- `templates/tests/internal/usecase/user_test.go.tmpl`
- `templates/tests/internal/infrastructure/repository/user_integration_test.go.tmpl`
- `templates/tests/testutil/fixtures.go.tmpl`
- `templates/tests/testutil/database.go.tmpl`
- `templates/tests/testutil/testcontainers.go.tmpl`
- `templates/tests/.golangci.yml.tmpl`

## Implementation Steps

### Step 1: Create GitHub Actions CI Workflow

```yaml
{{/* templates/ci-cd/github/workflows/ci.yml.tmpl */}}
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.22'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{"{{"}} env.GO_VERSION {{"}}"}}
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{"{{"}} env.GO_VERSION {{"}}"}}
          cache: true

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Run integration tests
        run: go test -v -race -tags=integration ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test?sslmode=disable

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          fail_ci_if_error: false

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{"{{"}} env.GO_VERSION {{"}}"}}
          cache: true

      - name: Build binary
        run: |
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-w -s -X main.version=${{"{{"}} github.sha {{"}}"}}" \
            -o bin/{{.Name}} \
            ./cmd/{{.Name}}

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: binary
          path: bin/{{.Name}}

  docker:
    name: Docker Build
    runs-on: ubuntu-latest
    needs: [build]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: {{.Name}}:${{"{{"}} github.sha {{"}}"}}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Step 2: Create GitHub Actions Release Workflow

```yaml
{{/* templates/ci-cd/github/workflows/release.yml.tmpl */}}
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

env:
  GO_VERSION: '1.22'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{"{{"}} github.repository {{"}}"}}

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: ${{"{{"}} env.GO_VERSION {{"}}"}}
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}

  docker-release:
    name: Docker Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{"{{"}} env.REGISTRY {{"}}"}}
          username: ${{"{{"}} github.actor {{"}}"}}
          password: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{"{{"}} env.REGISTRY {{"}}"}}/${{"{{"}} env.IMAGE_NAME {{"}}"}}
          tags: |
            type=semver,pattern={{`{{version}}`}}
            type=semver,pattern={{`{{major}}`}}.{{`{{minor}}`}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{"{{"}} steps.meta.outputs.tags {{"}}"}}
          labels: ${{"{{"}} steps.meta.outputs.labels {{"}}"}}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Step 3: Create GitLab CI

```yaml
{{/* templates/ci-cd/gitlab/.gitlab-ci.yml.tmpl */}}
stages:
  - lint
  - test
  - build
  - deploy

variables:
  GO_VERSION: "1.22"
  POSTGRES_USER: test
  POSTGRES_PASSWORD: test
  POSTGRES_DB: test

.go-cache:
  variables:
    GOPATH: $CI_PROJECT_DIR/.go
  before_script:
    - mkdir -p .go
  cache:
    paths:
      - .go/pkg/mod/

lint:
  stage: lint
  image: golangci/golangci-lint:latest
  extends: .go-cache
  script:
    - golangci-lint run --timeout=5m
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

test:unit:
  stage: test
  image: golang:${GO_VERSION}
  extends: .go-cache
  script:
    - go test -v -race -coverprofile=coverage.out ./...
  coverage: '/coverage: \d+.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

test:integration:
  stage: test
  image: golang:${GO_VERSION}
  extends: .go-cache
  services:
    - name: postgres:16-alpine
      alias: postgres
  variables:
    DATABASE_URL: postgres://test:test@postgres:5432/test?sslmode=disable
  script:
    - go test -v -race -tags=integration ./...
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

build:
  stage: build
  image: golang:${GO_VERSION}
  extends: .go-cache
  script:
    - CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/{{.Name}} ./cmd/{{.Name}}
  artifacts:
    paths:
      - bin/{{.Name}}
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_COMMIT_TAG

docker:build:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_COMMIT_TAG
```

### Step 4: Create golangci-lint Config

```yaml
{{/* templates/tests/.golangci.yml.tmpl */}}
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - bodyclose
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - staticcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - whitespace

linters-settings:
  dupl:
    threshold: 100
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 2
    min-occurrences: 2
  goimports:
    local-prefixes: {{.ModulePath}}
  govet:
    check-shadowing: true
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gosec
    - path: internal/mocks/
      linters:
        - all
```

### Step 5: Create Unit Test Example

```go
{{/* templates/tests/internal/usecase/user_test.go.tmpl */}}
package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"{{.ModulePath}}/internal/domain/entity"
	"{{.ModulePath}}/internal/usecase"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func TestUserUsecase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		input := usecase.CreateUserInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		// Email doesn't exist
		mockRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, nil)
		// Create succeeds
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

		user, err := uc.Create(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Email, user.Email)
		assert.Equal(t, input.Name, user.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		existingUser := &entity.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		input := usecase.CreateUserInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		mockRepo.On("GetByEmail", mock.Anything, input.Email).Return(existingUser, nil)

		user, err := uc.Create(context.Background(), input)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, usecase.ErrEmailExists)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		userID := uuid.New()
		expectedUser := &entity.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

		user, err := uc.GetByID(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		userID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)

		user, err := uc.GetByID(context.Background(), userID)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, usecase.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_List(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	users := []*entity.User{
		{ID: uuid.New(), Email: "user1@example.com"},
		{ID: uuid.New(), Email: "user2@example.com"},
	}
	var total int64 = 2

	mockRepo.On("List", mock.Anything, 0, 20).Return(users, total, nil)

	input := usecase.ListInput{Page: 1, PageSize: 20}
	result, err := uc.List(context.Background(), input)

	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, total, result.Total)
	assert.Equal(t, 1, result.TotalPages)
	mockRepo.AssertExpectations(t)
}
```

### Step 6: Create Integration Test with Testcontainers

```go
{{/* templates/tests/internal/infrastructure/repository/user_integration_test.go.tmpl */}}
//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"{{.ModulePath}}/internal/domain/entity"
	"{{.ModulePath}}/internal/infrastructure/repository"
	"{{.ModulePath}}/testutil"
)

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.SetupTestDatabase(t, ctx)
	repo := repository.NewUserRepository(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		user := &entity.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Password: "hashed_password",
			Name:     "Test User",
			Role:     "user",
			Active:   true,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Name, found.Name)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		user := &entity.User{
			ID:       uuid.New(),
			Email:    "email-test@example.com",
			Password: "hashed_password",
			Name:     "Email Test",
			Role:     "user",
			Active:   true,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		user := &entity.User{
			ID:       uuid.New(),
			Email:    "update-test@example.com",
			Password: "hashed_password",
			Name:     "Before Update",
			Role:     "user",
			Active:   true,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		user.Name = "After Update"
		err = repo.Update(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "After Update", found.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		user := &entity.User{
			ID:       uuid.New(),
			Email:    "delete-test@example.com",
			Password: "hashed_password",
			Name:     "Delete Test",
			Role:     "user",
			Active:   true,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("List with pagination", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 5; i++ {
			user := &entity.User{
				ID:       uuid.New(),
				Email:    fmt.Sprintf("list-test-%d@example.com", i),
				Password: "hashed_password",
				Name:     fmt.Sprintf("List User %d", i),
				Role:     "user",
				Active:   true,
			}
			err := repo.Create(ctx, user)
			require.NoError(t, err)
		}

		users, total, err := repo.List(ctx, 0, 3)
		require.NoError(t, err)
		assert.Len(t, users, 3)
		assert.GreaterOrEqual(t, total, int64(5))
	})
}
```

### Step 7: Create Test Utilities

```go
{{/* templates/tests/testutil/testcontainers.go.tmpl */}}
//go:build integration

package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"{{.ModulePath}}/internal/domain/entity"
)

// SetupTestDatabase creates a PostgreSQL container and returns a GORM connection
func SetupTestDatabase(t *testing.T, ctx context.Context) *gorm.DB {
	t.Helper()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}
```

```go
{{/* templates/tests/testutil/fixtures.go.tmpl */}}
package testutil

import (
	"time"

	"github.com/google/uuid"
	"{{.ModulePath}}/internal/domain/entity"
)

// NewTestUser creates a test user with default values
func NewTestUser(overrides ...func(*entity.User)) *entity.User {
	user := &entity.User{
		ID:        uuid.New(),
		Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Password:  "$2a$10$abcdefghijklmnopqrstuv", // bcrypt hash placeholder
		Name:      "Test User",
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// WithEmail sets the email for a test user
func WithEmail(email string) func(*entity.User) {
	return func(u *entity.User) {
		u.Email = email
	}
}

// WithRole sets the role for a test user
func WithRole(role string) func(*entity.User) {
	return func(u *entity.User) {
		u.Role = role
	}
}

// WithInactive makes the test user inactive
func WithInactive() func(*entity.User) {
	return func(u *entity.User) {
		u.Active = false
	}
}
```

## Todo List

- [ ] Create .github/workflows/ci.yml.tmpl
- [ ] Create .github/workflows/release.yml.tmpl
- [ ] Create .gitlab-ci.yml.tmpl
- [ ] Create .golangci.yml.tmpl
- [ ] Create usecase/user_test.go.tmpl with mocks
- [ ] Create repository/user_integration_test.go.tmpl
- [ ] Create testutil/testcontainers.go.tmpl
- [ ] Create testutil/fixtures.go.tmpl
- [ ] Test CI workflow runs in generated project
- [ ] Verify golangci-lint passes
- [ ] Verify unit tests pass
- [ ] Verify integration tests pass with testcontainers

## Success Criteria

- [ ] GitHub Actions workflow runs successfully
- [ ] GitLab CI configuration valid
- [ ] Unit tests run with `go test ./...`
- [ ] Integration tests run with `-tags=integration`
- [ ] golangci-lint passes with no issues
- [ ] Test coverage reports generated

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| CI service limits | Low | Medium | Use caching, parallel jobs |
| Testcontainers slow | Medium | Low | Use in integration only |
| Flaky tests | Medium | Medium | Proper test isolation |

## Security Considerations

- No secrets in CI config (use GitHub/GitLab secrets)
- Test database isolated per test
- No production credentials in tests

## Next Steps

After completing this phase:
1. Proceed to [Phase 09: CLI Polish](./phase-09-cli-polish.md)
2. Add validation, error handling, help text
