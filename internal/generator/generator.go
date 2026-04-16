package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/thinhdang/go-template/internal/config"
	"github.com/thinhdang/go-template/templates"
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
		"WithCI":      g.cfg.WithCI,
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

		// Skip template directories that don't apply yet (BEFORE processing)
		skipDirs := []string{"clean-arch", "docker", "monitoring", "ci-cd", "tests"}
		for _, dir := range skipDirs {
			if path == dir || strings.HasPrefix(path, dir+"/") {
				if d.IsDir() && path == dir {
					return fs.SkipDir // Don't traverse into these directories
				}
				return nil
			}
		}

		// Handle base/ directory specially
		if path == "base" {
			return nil // Skip the base directory itself, only process its contents
		}

		// Strip "base/" prefix from template paths (flatten structure)
		relativePath := strings.TrimPrefix(path, "base/")

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

		// Create output file
		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer outFile.Close()

		// Execute template
		if err := tmpl.Execute(outFile, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", path, err)
		}

		return nil
	})
}
