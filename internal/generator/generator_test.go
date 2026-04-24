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
	if info, err := os.Stat(filepath.Join(targetDir, "go.mod")); err != nil {
		t.Fatalf("stat go.mod: %v", err)
	} else if info.Mode().Perm()&0o200 == 0 {
		t.Fatalf("go.mod should be writable, got mode %o", info.Mode().Perm())
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

	if _, err := os.Stat(filepath.Join(targetDir, "proto", "api.proto")); err != nil {
		t.Fatalf("proto/api.proto was not generated: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "generate.sh")); err != nil {
		t.Fatalf("generate.sh was not generated: %v", err)
	}
	if info, err := os.Stat(filepath.Join(targetDir, "generate.sh")); err != nil {
		t.Fatalf("stat generate.sh: %v", err)
	} else if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("generate.sh should be executable, got mode %o", info.Mode().Perm())
	}

	if _, err := os.Stat(filepath.Join(targetDir, "protogen", "api.pb.go")); err != nil {
		t.Fatalf("protogen/api.pb.go was not generated: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "internal", "docs", "api.swagger.json")); err != nil {
		t.Fatalf("internal/docs/api.swagger.json was not generated: %v", err)
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
