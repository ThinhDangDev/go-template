package app

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"__MODULE_PATH__/internal/boilerplate/config"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrator(cfg config.Config) (*migrate.Migrate, error) {
	dir := cfg.ResolvedMigrationsDir()
	sourceURL := (&url.URL{Scheme: "file", Path: dir}).String()

	m, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func CreateMigrationFiles(dir, name string) (string, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	safeName := sanitizeMigrationName(name)
	if safeName == "" {
		return "", "", fmt.Errorf("migration name is empty after sanitization")
	}

	version := time.Now().UTC().Format("20060102150405")
	upPath := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", version, safeName))
	downPath := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", version, safeName))

	if err := os.WriteFile(upPath, []byte("-- Write your UP migration here.\n"), 0o644); err != nil {
		return "", "", err
	}

	if err := os.WriteFile(downPath, []byte("-- Write your DOWN migration here.\n"), 0o644); err != nil {
		return "", "", err
	}

	return upPath, downPath, nil
}

func sanitizeMigrationName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	name = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	name = regexp.MustCompile(`_+`).ReplaceAllString(name, "_")
	return name
}
