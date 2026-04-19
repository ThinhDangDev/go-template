package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ThinhDangDev/go-template/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCIProviderPaths(t *testing.T) {
	t.Run("github actions uses hidden github directory only", func(t *testing.T) {
		projectDir := generateTestProject(t, &config.ProjectConfig{
			Name:        filepath.Join(t.TempDir(), "github-app"),
			ModulePath:  "github.com/test/github-app",
			Description: "GitHub CI app",
			APIType:     config.APITypeREST,
			AuthType:    config.AuthTypeJWT,
			Database:    "postgres",
			WithDocker:  true,
			WithCI:      "github",
			WithMonitor: true,
		})

		assert.FileExists(t, filepath.Join(projectDir, ".github", "workflows", "ci.yml"))
		assert.FileExists(t, filepath.Join(projectDir, ".github", "workflows", "release.yml"))
		assert.FileExists(t, filepath.Join(projectDir, ".goreleaser.yml"))
		assert.NoFileExists(t, filepath.Join(projectDir, "github", "workflows", "ci.yml"))
		assert.NoFileExists(t, filepath.Join(projectDir, ".gitlab-ci.yml"))
		assert.NoFileExists(t, filepath.Join(projectDir, ".gitkeep"))
	})

	t.Run("gitlab uses root pipeline file only", func(t *testing.T) {
		projectDir := generateTestProject(t, &config.ProjectConfig{
			Name:        filepath.Join(t.TempDir(), "gitlab-app"),
			ModulePath:  "github.com/test/gitlab-app",
			Description: "GitLab CI app",
			APIType:     config.APITypeREST,
			AuthType:    config.AuthTypeJWT,
			Database:    "postgres",
			WithDocker:  true,
			WithCI:      "gitlab",
			WithMonitor: true,
		})

		assert.FileExists(t, filepath.Join(projectDir, ".gitlab-ci.yml"))
		assert.NoDirExists(t, filepath.Join(projectDir, ".github"))
		assert.NoDirExists(t, filepath.Join(projectDir, "gitlab"))
	})

	t.Run("none skips ci assets entirely", func(t *testing.T) {
		projectDir := generateTestProject(t, &config.ProjectConfig{
			Name:        filepath.Join(t.TempDir(), "no-ci-app"),
			ModulePath:  "github.com/test/no-ci-app",
			Description: "No CI app",
			APIType:     config.APITypeREST,
			AuthType:    config.AuthTypeJWT,
			Database:    "postgres",
			WithDocker:  true,
			WithCI:      "none",
			WithMonitor: true,
		})

		assert.NoFileExists(t, filepath.Join(projectDir, ".gitlab-ci.yml"))
		assert.NoDirExists(t, filepath.Join(projectDir, ".github"))
	})
}

func TestGeneratedGoModIncludesPhase08Dependencies(t *testing.T) {
	projectDir := generateTestProject(t, &config.ProjectConfig{
		Name:        filepath.Join(t.TempDir(), "deps-app"),
		ModulePath:  "github.com/test/deps-app",
		Description: "Dependency app",
		APIType:     config.APITypeREST,
		AuthType:    config.AuthTypeJWT,
		Database:    "postgres",
		WithDocker:  true,
		WithCI:      "github",
		WithMonitor: true,
	})

	content, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	require.NoError(t, err)

	goMod := string(content)
	assert.Contains(t, goMod, "github.com/google/uuid v1.6.0")
	assert.Contains(t, goMod, "github.com/prometheus/client_golang v1.20.5")
	assert.Contains(t, goMod, "github.com/stretchr/testify v1.10.0")
	assert.Contains(t, goMod, "github.com/testcontainers/testcontainers-go v0.33.0")
	assert.Contains(t, goMod, "github.com/testcontainers/testcontainers-go/modules/postgres v0.33.0")
	assert.Contains(t, goMod, "go.uber.org/zap v1.27.0")
	assert.Contains(t, goMod, "golang.org/x/crypto v0.31.0")
	assert.Contains(t, goMod, "gorm.io/driver/postgres v1.5.11")
}

func TestConditionalTemplatesSkipEmptyOutputs(t *testing.T) {
	projectDir := generateTestProject(t, &config.ProjectConfig{
		Name:        filepath.Join(t.TempDir(), "conditional-app"),
		ModulePath:  "github.com/test/conditional-app",
		Description: "Conditional app",
		APIType:     config.APITypeREST,
		AuthType:    config.AuthTypeNone,
		Database:    "postgres",
		WithDocker:  true,
		WithCI:      "none",
		WithMonitor: false,
	})

	assert.NoFileExists(t, filepath.Join(projectDir, "internal", "delivery", "rest", "handler", "metrics.go"))
	assert.NoFileExists(t, filepath.Join(projectDir, "internal", "delivery", "rest", "handler", "auth.go"))
	assert.NoFileExists(t, filepath.Join(projectDir, "internal", "delivery", "rest", "middleware", "auth", "jwt.go"))
	assert.NoFileExists(t, filepath.Join(projectDir, "internal", "infrastructure", "auth", "jwt.go"))
	assert.NoFileExists(t, filepath.Join(projectDir, "internal", "usecase", "auth.go"))
}

func TestGitLabCITemplateMatchesGeneratedIntegrationStrategy(t *testing.T) {
	projectDir := generateTestProject(t, &config.ProjectConfig{
		Name:        filepath.Join(t.TempDir(), "gitlab-ci-app"),
		ModulePath:  "github.com/test/gitlab-ci-app",
		Description: "GitLab CI app",
		APIType:     config.APITypeREST,
		AuthType:    config.AuthTypeJWT,
		Database:    "postgres",
		WithDocker:  true,
		WithCI:      "gitlab",
		WithMonitor: true,
	})

	gitlabCI, err := os.ReadFile(filepath.Join(projectDir, ".gitlab-ci.yml"))
	require.NoError(t, err)

	content := string(gitlabCI)
	assert.Contains(t, content, "DATABASE_URL: postgres://test:test@postgres:5432/test?sslmode=disable")
	assert.Contains(t, content, "DOCKER_HOST: tcp://docker:2375")
	assert.Contains(t, content, "DOCKER_TLS_CERTDIR: \"\"")
	assert.NotContains(t, content, "coverage.xml")

	testutilContent, err := os.ReadFile(filepath.Join(projectDir, "testutil", "testcontainers.go"))
	require.NoError(t, err)
	assert.Contains(t, string(testutilContent), "os.Getenv(\"DATABASE_URL\")")
}

func generateTestProject(t *testing.T, cfg *config.ProjectConfig) string {
	t.Helper()

	require.NoError(t, New(cfg).Generate())

	projectDir := cfg.Name
	if !filepath.IsAbs(projectDir) {
		projectDir = filepath.Join(".", projectDir)
	}

	projectDir = filepath.Clean(projectDir)
	assert.False(t, strings.HasSuffix(projectDir, ".tmpl"))

	return projectDir
}
