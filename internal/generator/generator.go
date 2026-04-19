package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ThinhDangDev/go-template/internal/config"
	"github.com/ThinhDangDev/go-template/templates"
)

// Generator handles project generation from templates
type Generator struct {
	cfg *config.ProjectConfig
}

// New creates a new Generator
func New(cfg *config.ProjectConfig) *Generator {
	return &Generator{cfg: cfg}
}

// Generate creates the project structure
func (g *Generator) Generate() error {
	// Check if directory already exists
	if _, err := os.Stat(g.cfg.Name); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", g.cfg.Name)
	}

	// Create project directory
	if err := os.MkdirAll(g.cfg.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Template data
	data := map[string]interface{}{
		"Name":        g.cfg.Name,
		"ModulePath":  g.cfg.ModulePath,
		"Description": g.cfg.Description,
		"APIType":     string(g.cfg.APIType),
		"AuthType":    string(g.cfg.AuthType),
		"Database":    g.cfg.Database,
		"WithDocker":  g.cfg.WithDocker,
		"WithCI":      g.cfg.WithCI != "none",
		"WithMonitor": g.cfg.WithMonitor,
		"Year":        time.Now().Year(),
	}

	// Walk through embedded templates
	return fs.WalkDir(templates.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory
		if path == "." {
			return nil
		}

		switch {
		case g.cfg.WithCI == "github" && (path == "ci-cd/gitlab" || strings.HasPrefix(path, "ci-cd/gitlab/")):
			if d.IsDir() && path == "ci-cd/gitlab" {
				return fs.SkipDir
			}
			return nil
		case g.cfg.WithCI == "gitlab" && (path == "ci-cd/github" || strings.HasPrefix(path, "ci-cd/github/")):
			if d.IsDir() && path == "ci-cd/github" {
				return fs.SkipDir
			}
			return nil
		}

		// Skip CI/CD templates if not enabled
		if g.cfg.WithCI == "none" && (path == "ci-cd" || strings.HasPrefix(path, "ci-cd/")) {
			if d.IsDir() && path == "ci-cd" {
				return fs.SkipDir
			}
			return nil
		}

		// Conditionally skip docker/monitoring based on config
		if !g.cfg.WithDocker && (path == "docker" || strings.HasPrefix(path, "docker/")) {
			if d.IsDir() && path == "docker" {
				return fs.SkipDir
			}
			return nil
		}
		if !g.cfg.WithMonitor && (path == "monitoring" || strings.HasPrefix(path, "monitoring/")) {
			if d.IsDir() && path == "monitoring" {
				return fs.SkipDir
			}
			return nil
		}

		// Handle template root directories specially
		if path == "base" || path == "clean-arch" || path == "docker" || path == "monitoring" || path == "ci-cd" || path == "tests" {
			return nil // Skip the directory itself, only process its contents
		}

		// Strip template directory prefixes from paths
		relativePath := path
		switch {
		case path == "ci-cd/github":
			relativePath = ".github"
		case strings.HasPrefix(path, "ci-cd/github/"):
			relativePath = filepath.Join(".github", strings.TrimPrefix(path, "ci-cd/github/"))
		case path == "ci-cd/gitlab":
			relativePath = ""
		case strings.HasPrefix(path, "ci-cd/gitlab/"):
			relativePath = strings.TrimPrefix(path, "ci-cd/gitlab/")
		default:
			relativePath = strings.TrimPrefix(relativePath, "base/")
			relativePath = strings.TrimPrefix(relativePath, "clean-arch/")
			relativePath = strings.TrimPrefix(relativePath, "docker/")
			relativePath = strings.TrimPrefix(relativePath, "ci-cd/")
			relativePath = strings.TrimPrefix(relativePath, "tests/")
		}

		// Skip if this results in an empty path
		if relativePath == "" || relativePath == "." {
			return nil
		}

		// Calculate target path
		targetPath := filepath.Join(g.cfg.Name, relativePath)

		// Remove .tmpl suffix if present
		if strings.HasSuffix(targetPath, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read template file
		content, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Render template
		tmpl, err := template.New(path).Funcs(FuncMap()).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		var rendered bytes.Buffer
		if err := tmpl.Execute(&rendered, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", path, err)
		}

		// Skip empty outputs from conditionally-rendered templates and placeholder files.
		if strings.TrimSpace(rendered.String()) == "" {
			return nil
		}

		// Create output file
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer outFile.Close()

		if _, err := outFile.Write(rendered.Bytes()); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		return nil
	})
}
