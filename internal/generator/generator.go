package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const templateRoot = "testdata/project-template"

//go:embed all:testdata/project-template
var templateFS embed.FS

type Config struct {
	ProjectName string
	ModulePath  string
	TargetDir   string
}

func InitProject(cfg Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	if cfg.ModulePath == "" {
		cfg.ModulePath = cfg.ProjectName
	}

	if err := ensureEmptyTargetDir(cfg.TargetDir); err != nil {
		return err
	}

	replacer := strings.NewReplacer(
		"__MODULE_PATH__", cfg.ModulePath,
		"__PROJECT_NAME__", cfg.ProjectName,
		"__PROJECT_NAME_SNAKE__", toSnakeCase(cfg.ProjectName),
		"internal/boilerplate/", "internal/",
		"internal/boilerplate", "internal",
	)

	return fs.WalkDir(templateFS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel := strings.TrimPrefix(path, templateRoot)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		rel = rewriteTemplatePath(rel)
		if rel == "go.mod.tmpl" {
			rel = "go.mod"
		}

		dstPath := filepath.Join(cfg.TargetDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		content, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return err
		}
		fileInfo, err := fs.Stat(templateFS, path)
		if err != nil {
			return err
		}

		rendered := replacer.Replace(string(content))
		mode := os.FileMode(0o644)
		if fileInfo.Mode().Perm()&0o111 != 0 || strings.HasSuffix(rel, ".sh") {
			mode = 0o755
		}
		return os.WriteFile(dstPath, []byte(rendered), mode)
	})
}

func rewriteTemplatePath(rel string) string {
	if strings.HasPrefix(rel, "internal/boilerplate/") {
		return "internal/" + strings.TrimPrefix(rel, "internal/boilerplate/")
	}
	if rel == "internal/boilerplate" {
		return "internal"
	}

	return rel
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.ProjectName) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.ContainsAny(cfg.ProjectName, ` <>:"|?*`) {
		return fmt.Errorf("project name contains invalid path characters")
	}
	if strings.TrimSpace(cfg.TargetDir) == "" {
		return fmt.Errorf("target directory is required")
	}
	return nil
}

func ensureEmptyTargetDir(targetDir string) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(targetDir, 0o755)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory: %s", targetDir)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("target directory is not empty: %s", targetDir)
	}

	return nil
}

func toSnakeCase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	value = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(value, "_")
	value = regexp.MustCompile(`_+`).ReplaceAllString(value, "_")
	return strings.Trim(value, "_")
}
