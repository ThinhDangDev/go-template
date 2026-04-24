package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeMigrationName(t *testing.T) {
	got := sanitizeMigrationName(" Add Users Table! ")
	want := "add_users_table"
	if got != want {
		t.Fatalf("sanitizeMigrationName() = %q, want %q", got, want)
	}
}

func TestCreateMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	upPath, downPath, err := CreateMigrationFiles(dir, "create users")
	if err != nil {
		t.Fatalf("CreateMigrationFiles() error = %v", err)
	}

	if !strings.HasSuffix(upPath, ".up.sql") {
		t.Fatalf("expected up migration suffix, got %s", upPath)
	}

	if !strings.HasSuffix(downPath, ".down.sql") {
		t.Fatalf("expected down migration suffix, got %s", downPath)
	}

	if _, err := os.Stat(upPath); err != nil {
		t.Fatalf("up migration was not created: %v", err)
	}

	if _, err := os.Stat(downPath); err != nil {
		t.Fatalf("down migration was not created: %v", err)
	}

	if filepath.Dir(upPath) != dir || filepath.Dir(downPath) != dir {
		t.Fatalf("expected migrations to be created in %s", dir)
	}
}
