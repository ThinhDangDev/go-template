# Phase 03: Template Engine

---
status: completed
priority: P1
effort: 4h
dependencies: [phase-02]
completed: 2026-04-16
---

## Context Links

- [Main Plan](./plan.md)
- [Previous: CLI Framework](./phase-02-cli-framework.md)
- [Next: Base Templates](./phase-04-base-templates.md)

## Overview

Implement the core template engine using Go's `text/template` and `embed.FS`. This engine renders templates with project configuration and writes them to the target directory.

## Key Insights

- `embed.FS` bakes templates into the binary - no external files needed
- `text/template` is sufficient; `html/template` only needed for HTML
- Template functions extend capabilities (e.g., `lower`, `title`, `contains`)
- File path templates allow dynamic naming (e.g., `{{.Name}}_handler.go`)

## Requirements

### Functional
- Embed all template files at compile time
- Parse and render templates with ProjectConfig data
- Create directory structure from templates
- Handle conditional file generation
- Support dynamic file naming

### Non-Functional
- Fast rendering (< 100ms for typical project)
- Clear error messages with file/line info
- No external runtime dependencies

## Architecture

```
internal/generator/
├── generator.go     # Main Generator struct and Generate method
├── renderer.go      # Template rendering logic
├── funcs.go         # Custom template functions
└── writer.go        # File writing utilities

templates/
├── embed.go         # embed.FS declaration
└── ... (template files)
```

## Related Code Files

### Files to Create
- `internal/generator/generator.go`
- `internal/generator/renderer.go`
- `internal/generator/funcs.go`
- `internal/generator/writer.go`
- `templates/embed.go`

## Implementation Steps

### Step 1: Create Embed Declaration

```go
// templates/embed.go
package templates

import "embed"

//go:embed all:base all:clean-arch all:docker all:monitoring all:ci-cd all:tests
var FS embed.FS
```

### Step 2: Create Template Functions

```go
// internal/generator/funcs.go
package generator

import (
    "strings"
    "text/template"
    "unicode"
)

// FuncMap returns custom template functions
func FuncMap() template.FuncMap {
    return template.FuncMap{
        // String transformations
        "lower":      strings.ToLower,
        "upper":      strings.ToUpper,
        "title":      strings.Title,
        "snake":      toSnakeCase,
        "camel":      toCamelCase,
        "pascal":     toPascalCase,
        "kebab":      toKebabCase,

        // Conditionals
        "contains":   strings.Contains,
        "hasPrefix":  strings.HasPrefix,
        "hasSuffix":  strings.HasSuffix,

        // API type checks
        "hasREST": func(apiType string) bool {
            return apiType == "rest" || apiType == "both"
        },
        "hasGRPC": func(apiType string) bool {
            return apiType == "grpc" || apiType == "both"
        },

        // Auth type checks
        "hasJWT": func(authType string) bool {
            return authType == "jwt" || authType == "both"
        },
        "hasOAuth2": func(authType string) bool {
            return authType == "oauth2" || authType == "both"
        },
        "hasAuth": func(authType string) bool {
            return authType != "none"
        },
    }
}

func toSnakeCase(s string) string {
    var result strings.Builder
    for i, r := range s {
        if unicode.IsUpper(r) {
            if i > 0 {
                result.WriteRune('_')
            }
            result.WriteRune(unicode.ToLower(r))
        } else {
            result.WriteRune(r)
        }
    }
    return result.String()
}

func toCamelCase(s string) string {
    parts := strings.FieldsFunc(s, func(r rune) bool {
        return r == '_' || r == '-' || r == ' '
    })
    for i := range parts {
        if i == 0 {
            parts[i] = strings.ToLower(parts[i])
        } else {
            parts[i] = strings.Title(parts[i])
        }
    }
    return strings.Join(parts, "")
}

func toPascalCase(s string) string {
    parts := strings.FieldsFunc(s, func(r rune) bool {
        return r == '_' || r == '-' || r == ' '
    })
    for i := range parts {
        parts[i] = strings.Title(parts[i])
    }
    return strings.Join(parts, "")
}

func toKebabCase(s string) string {
    return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}
```

### Step 3: Create Renderer

```go
// internal/generator/renderer.go
package generator

import (
    "bytes"
    "fmt"
    "io/fs"
    "path/filepath"
    "strings"
    "text/template"

    "github.com/thinhdang/go-template/internal/config"
)

// TemplateData wraps ProjectConfig with additional template helpers
type TemplateData struct {
    *config.ProjectConfig
    Year int // Current year for licenses
}

// Renderer handles template parsing and rendering
type Renderer struct {
    fs       fs.FS
    funcMap  template.FuncMap
}

// NewRenderer creates a new template renderer
func NewRenderer(fsys fs.FS) *Renderer {
    return &Renderer{
        fs:      fsys,
        funcMap: FuncMap(),
    }
}

// RenderFile renders a single template file
func (r *Renderer) RenderFile(path string, data *TemplateData) ([]byte, error) {
    content, err := fs.ReadFile(r.fs, path)
    if err != nil {
        return nil, fmt.Errorf("read template %s: %w", path, err)
    }

    tmpl, err := template.New(filepath.Base(path)).
        Funcs(r.funcMap).
        Parse(string(content))
    if err != nil {
        return nil, fmt.Errorf("parse template %s: %w", path, err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return nil, fmt.Errorf("execute template %s: %w", path, err)
    }

    return buf.Bytes(), nil
}

// RenderFileName renders template variables in file names
func (r *Renderer) RenderFileName(name string, data *TemplateData) (string, error) {
    // Remove .tmpl suffix if present
    name = strings.TrimSuffix(name, ".tmpl")

    // Check if filename contains template syntax
    if !strings.Contains(name, "{{") {
        return name, nil
    }

    tmpl, err := template.New("filename").
        Funcs(r.funcMap).
        Parse(name)
    if err != nil {
        return "", fmt.Errorf("parse filename template: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("execute filename template: %w", err)
    }

    return buf.String(), nil
}
```

### Step 4: Create File Writer

```go
// internal/generator/writer.go
package generator

import (
    "fmt"
    "os"
    "path/filepath"
)

// Writer handles file and directory creation
type Writer struct {
    baseDir string
    verbose bool
}

// NewWriter creates a new file writer
func NewWriter(baseDir string, verbose bool) *Writer {
    return &Writer{
        baseDir: baseDir,
        verbose: verbose,
    }
}

// WriteFile writes content to a file, creating directories as needed
func (w *Writer) WriteFile(relPath string, content []byte, mode os.FileMode) error {
    fullPath := filepath.Join(w.baseDir, relPath)
    dir := filepath.Dir(fullPath)

    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create directory %s: %w", dir, err)
    }

    if w.verbose {
        fmt.Printf("  create: %s\n", relPath)
    }

    if err := os.WriteFile(fullPath, content, mode); err != nil {
        return fmt.Errorf("write file %s: %w", fullPath, err)
    }

    return nil
}

// MkdirAll creates a directory and all parents
func (w *Writer) MkdirAll(relPath string) error {
    fullPath := filepath.Join(w.baseDir, relPath)

    if w.verbose {
        fmt.Printf("  mkdir: %s\n", relPath)
    }

    return os.MkdirAll(fullPath, 0755)
}

// Exists checks if a path exists
func (w *Writer) Exists(relPath string) bool {
    fullPath := filepath.Join(w.baseDir, relPath)
    _, err := os.Stat(fullPath)
    return err == nil
}
```

### Step 5: Create Generator

```go
// internal/generator/generator.go
package generator

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/thinhdang/go-template/internal/config"
    "github.com/thinhdang/go-template/templates"
)

// Generator orchestrates project generation
type Generator struct {
    renderer *Renderer
    writer   *Writer
    verbose  bool
}

// New creates a new Generator
func New(targetDir string, verbose bool) *Generator {
    return &Generator{
        renderer: NewRenderer(templates.FS),
        writer:   NewWriter(targetDir, verbose),
        verbose:  verbose,
    }
}

// Generate creates the project from templates
func (g *Generator) Generate(cfg *config.ProjectConfig) error {
    data := &TemplateData{
        ProjectConfig: cfg,
        Year:          time.Now().Year(),
    }

    // Create base directory
    if err := g.writer.MkdirAll("."); err != nil {
        return err
    }

    // Process template directories based on config
    dirs := []string{"base", "clean-arch"}

    if cfg.WithDocker {
        dirs = append(dirs, "docker")
    }
    if cfg.WithMonitor {
        dirs = append(dirs, "monitoring")
    }
    if cfg.WithCI != "none" {
        dirs = append(dirs, "ci-cd")
    }
    dirs = append(dirs, "tests")

    for _, dir := range dirs {
        if err := g.processDir(dir, data); err != nil {
            return fmt.Errorf("process %s: %w", dir, err)
        }
    }

    return nil
}

// processDir walks a template directory and renders all files
func (g *Generator) processDir(templateDir string, data *TemplateData) error {
    return fs.WalkDir(templates.FS, templateDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Skip the root template directory itself
        if path == templateDir {
            return nil
        }

        // Calculate relative path (remove template dir prefix)
        relPath := strings.TrimPrefix(path, templateDir+"/")

        // Skip based on conditions
        if g.shouldSkip(path, data) {
            if d.IsDir() {
                return fs.SkipDir
            }
            return nil
        }

        if d.IsDir() {
            return g.writer.MkdirAll(relPath)
        }

        // Render file name (may contain template variables)
        outputName, err := g.renderer.RenderFileName(relPath, data)
        if err != nil {
            return err
        }

        // Render file content
        content, err := g.renderer.RenderFile(path, data)
        if err != nil {
            return err
        }

        // Determine file mode (executable for scripts)
        mode := os.FileMode(0644)
        if strings.HasSuffix(outputName, ".sh") {
            mode = 0755
        }

        return g.writer.WriteFile(outputName, content, mode)
    })
}

// shouldSkip determines if a file/dir should be skipped based on config
func (g *Generator) shouldSkip(path string, data *TemplateData) bool {
    // Skip gRPC files if not using gRPC
    if strings.Contains(path, "grpc") && data.APIType != config.APITypeGRPC && data.APIType != config.APITypeBoth {
        return true
    }

    // Skip REST files if only using gRPC
    if strings.Contains(path, "/rest/") && data.APIType == config.APITypeGRPC {
        return true
    }

    // Skip auth files if no auth
    if strings.Contains(path, "/auth/") && data.AuthType == config.AuthTypeNone {
        return true
    }

    // Skip JWT files if not using JWT
    if strings.Contains(path, "jwt") && data.AuthType != config.AuthTypeJWT && data.AuthType != config.AuthTypeBoth {
        return true
    }

    // Skip OAuth2 files if not using OAuth2
    if strings.Contains(path, "oauth") && data.AuthType != config.AuthTypeOAuth2 && data.AuthType != config.AuthTypeBoth {
        return true
    }

    // Skip GitHub CI if using GitLab
    if strings.Contains(path, "github") && data.WithCI == "gitlab" {
        return true
    }

    // Skip GitLab CI if using GitHub
    if strings.Contains(path, "gitlab") && data.WithCI == "github" {
        return true
    }

    return false
}
```

### Step 6: Update init.go to use Generator

```go
// In cmd/init.go, update runInit function:
func runInit(cmd *cobra.Command, args []string) error {
    // ... (existing validation code)

    fmt.Printf("\n🚀 Creating project '%s'...\n", cfg.Name)

    gen := generator.New(cfg.Name, verbose)
    if err := gen.Generate(cfg); err != nil {
        return fmt.Errorf("generation failed: %w", err)
    }

    fmt.Printf("\n✅ Project '%s' created successfully!\n\n", cfg.Name)
    fmt.Println("Next steps:")
    fmt.Printf("  cd %s\n", cfg.Name)
    fmt.Println("  go mod tidy")
    fmt.Println("  make docker-up  # Start dependencies")
    fmt.Println("  make run        # Start the service")

    return nil
}
```

## Todo List

- [x] Create templates/embed.go with FS declaration
- [x] Create internal/generator/funcs.go with template functions
- [x] Create internal/generator/renderer.go
- [x] Create internal/generator/writer.go
- [x] Create internal/generator/generator.go
- [x] Update cmd/init.go to use generator
- [x] Create placeholder template files for testing
- [x] Test generation with minimal templates
- [x] Verify embed works in built binary

## Success Criteria

- [x] Templates compile into binary (no external files needed)
- [x] Template functions work (snake_case, camelCase, etc.)
- [x] Conditional generation works (skip gRPC if REST only)
- [x] Dynamic file naming works
- [x] Clear error messages for template errors

## Known Issues

- Files placed in nested directories (base/, clean-arch/, etc) instead of project root
  - Root cause: Generator preserves template directory structure
  - Impact: Generated projects need manual file reorganization
  - Status: Identified, needs fix in Phase 09 (CLI Polish)

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Embed path issues | Medium | High | Test early, use relative paths |
| Template syntax errors | Medium | Medium | Validate templates at build |
| Large binary size | Low | Low | Templates are text, minimal impact |

## Security Considerations

- Validate template output doesn't escape target directory
- Sanitize file names from templates
- No code execution in templates (text/template is safe)

## Next Steps

After completing this phase:
1. Proceed to [Phase 04: Base Templates](./phase-04-base-templates.md)
2. Create the foundational template files
