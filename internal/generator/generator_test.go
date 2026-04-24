package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitProject(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), "sample-api")

	err := InitProject(Config{
		ProjectName: "sample-api",
		ModulePath:  "github.com/acme/sample-api",
		TargetDir:   targetDir,
	})
	if err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(targetDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(goMod), "module github.com/acme/sample-api") {
		t.Fatalf("go.mod module path was not replaced")
	}

	envExample, err := os.ReadFile(filepath.Join(targetDir, ".env.example"))
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}
	if !strings.Contains(string(envExample), "POSTGRES_DB=sample_api") {
		t.Fatalf(".env.example project name placeholder was not replaced")
	}

	cmdMain, err := os.ReadFile(filepath.Join(targetDir, "cmd", "main.go"))
	if err != nil {
		t.Fatalf("read cmd/main.go: %v", err)
	}
	if !strings.Contains(string(cmdMain), `"github.com/acme/sample-api/internal/boilerplate/cli"`) {
		t.Fatalf("cmd/main.go module import was not replaced")
	}

	if _, err := os.Stat(filepath.Join(targetDir, ".gitignore")); err != nil {
		t.Fatalf(".gitignore was not generated: %v", err)
	}
}

func TestInitProjectRejectsNonEmptyDir(t *testing.T) {
	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(targetDir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := InitProject(Config{
		ProjectName: "sample-api",
		ModulePath:  "github.com/acme/sample-api",
		TargetDir:   targetDir,
	})
	if err == nil {
		t.Fatalf("expected error for non-empty directory")
	}
}
