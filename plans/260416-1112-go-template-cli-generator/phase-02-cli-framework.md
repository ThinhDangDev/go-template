# Phase 02: CLI Framework Setup

---
status: completed
priority: P1
effort: 3h
dependencies: [phase-01]
completed: 2026-04-16
---

## Context Links

- [Main Plan](./plan.md)
- [Previous: Project Foundation](./phase-01-project-foundation.md)
- [Next: Template Engine](./phase-03-template-engine.md)

## Overview

Implement Cobra CLI structure with root, init, and version commands. Set up interactive prompts using Survey for project configuration.

## Key Insights

- Cobra provides automatic help generation and command completion
- Viper handles config persistence (for defaults/preferences)
- Survey v2 supports validation and conditional prompts
- Keep commands thin - delegate to internal packages

## Requirements

### Functional
- `go-template` - shows help
- `go-template init <project-name>` - creates new project
- `go-template version` - shows version info
- Interactive prompts for project configuration

### Non-Functional
- Sub-second startup time
- Clear error messages
- Colored output for better UX

## Architecture

```
cmd/
├── root.go      # Root command, global flags
├── init.go      # Init command with prompts
└── version.go   # Version command

internal/
├── config/
│   └── project.go   # ProjectConfig struct
└── prompt/
    └── prompt.go    # Survey prompt logic
```

## Related Code Files

### Files to Create
- `cmd/root.go`
- `cmd/init.go`
- `cmd/version.go`
- `internal/config/project.go`
- `internal/prompt/prompt.go`

## Implementation Steps

### Step 1: Create Root Command

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "go-template",
    Short: "Go backend boilerplate CLI generator",
    Long: `go-template is a CLI tool that generates production-ready
Go backend projects with Clean Architecture, authentication,
monitoring, and comprehensive testing out of the box.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Global flags can be added here
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
```

### Step 2: Create Version Command

```go
// cmd/version.go
package cmd

import (
    "fmt"
    "runtime"

    "github.com/spf13/cobra"
)

// Set via ldflags
var (
    version   = "dev"
    commit    = "none"
    buildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("go-template %s\n", version)
        fmt.Printf("  commit: %s\n", commit)
        fmt.Printf("  built: %s\n", buildDate)
        fmt.Printf("  go: %s\n", runtime.Version())
        fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### Step 3: Create Project Config

```go
// internal/config/project.go
package config

// APIType represents the API delivery type
type APIType string

const (
    APITypeREST APIType = "rest"
    APITypeGRPC APIType = "grpc"
    APITypeBoth APIType = "both"
)

// AuthType represents authentication method
type AuthType string

const (
    AuthTypeJWT    AuthType = "jwt"
    AuthTypeOAuth2 AuthType = "oauth2"
    AuthTypeBoth   AuthType = "both"
    AuthTypeNone   AuthType = "none"
)

// ProjectConfig holds all project configuration
type ProjectConfig struct {
    Name        string   // Project name (directory name)
    ModulePath  string   // Go module path
    Description string   // Project description
    APIType     APIType  // REST, gRPC, or both
    AuthType    AuthType // JWT, OAuth2, both, or none
    Database    string   // Database type (postgres default)
    WithDocker  bool     // Generate Docker files
    WithCI      string   // CI provider (github, gitlab, none)
    WithMonitor bool     // Generate monitoring configs
}

// NewDefaultConfig creates config with sensible defaults
func NewDefaultConfig(name string) *ProjectConfig {
    return &ProjectConfig{
        Name:        name,
        ModulePath:  "github.com/user/" + name,
        Description: "A Go backend service",
        APIType:     APITypeREST,
        AuthType:    AuthTypeJWT,
        Database:    "postgres",
        WithDocker:  true,
        WithCI:      "github",
        WithMonitor: true,
    }
}
```

### Step 4: Create Prompt Logic

```go
// internal/prompt/prompt.go
package prompt

import (
    "fmt"
    "regexp"

    "github.com/AlecAivazis/survey/v2"
    "github.com/thinhdang/go-template/internal/config"
)

var modulePathRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

// CollectConfig prompts user for project configuration
func CollectConfig(projectName string) (*config.ProjectConfig, error) {
    cfg := config.NewDefaultConfig(projectName)

    questions := []*survey.Question{
        {
            Name: "modulePath",
            Prompt: &survey.Input{
                Message: "Go module path:",
                Default: cfg.ModulePath,
            },
            Validate: func(val interface{}) error {
                str := val.(string)
                if !modulePathRegex.MatchString(str) {
                    return fmt.Errorf("invalid module path")
                }
                return nil
            },
        },
        {
            Name: "description",
            Prompt: &survey.Input{
                Message: "Project description:",
                Default: cfg.Description,
            },
        },
        {
            Name: "apiType",
            Prompt: &survey.Select{
                Message: "API type:",
                Options: []string{"REST (Gin)", "gRPC", "Both"},
                Default: "REST (Gin)",
            },
        },
        {
            Name: "authType",
            Prompt: &survey.Select{
                Message: "Authentication:",
                Options: []string{"JWT", "OAuth2", "Both", "None"},
                Default: "JWT",
            },
        },
        {
            Name: "withDocker",
            Prompt: &survey.Confirm{
                Message: "Generate Docker files?",
                Default: true,
            },
        },
        {
            Name: "ciProvider",
            Prompt: &survey.Select{
                Message: "CI/CD provider:",
                Options: []string{"GitHub Actions", "GitLab CI", "None"},
                Default: "GitHub Actions",
            },
        },
        {
            Name: "withMonitor",
            Prompt: &survey.Confirm{
                Message: "Include Prometheus + Grafana monitoring?",
                Default: true,
            },
        },
    }

    answers := struct {
        ModulePath  string
        Description string
        APIType     string
        AuthType    string
        WithDocker  bool
        CIProvider  string
        WithMonitor bool
    }{}

    if err := survey.Ask(questions, &answers); err != nil {
        return nil, err
    }

    // Map answers to config
    cfg.ModulePath = answers.ModulePath
    cfg.Description = answers.Description
    cfg.WithDocker = answers.WithDocker
    cfg.WithMonitor = answers.WithMonitor

    // Map API type
    switch answers.APIType {
    case "REST (Gin)":
        cfg.APIType = config.APITypeREST
    case "gRPC":
        cfg.APIType = config.APITypeGRPC
    case "Both":
        cfg.APIType = config.APITypeBoth
    }

    // Map auth type
    switch answers.AuthType {
    case "JWT":
        cfg.AuthType = config.AuthTypeJWT
    case "OAuth2":
        cfg.AuthType = config.AuthTypeOAuth2
    case "Both":
        cfg.AuthType = config.AuthTypeBoth
    case "None":
        cfg.AuthType = config.AuthTypeNone
    }

    // Map CI provider
    switch answers.CIProvider {
    case "GitHub Actions":
        cfg.WithCI = "github"
    case "GitLab CI":
        cfg.WithCI = "gitlab"
    case "None":
        cfg.WithCI = "none"
    }

    return cfg, nil
}
```

### Step 5: Create Init Command

```go
// cmd/init.go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"

    "github.com/spf13/cobra"
    "github.com/thinhdang/go-template/internal/prompt"
)

var projectNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

var initCmd = &cobra.Command{
    Use:   "init <project-name>",
    Short: "Initialize a new Go backend project",
    Long: `Initialize a new Go backend project with Clean Architecture,
authentication, monitoring, and testing scaffolding.

Example:
  go-template init my-service
  go-template init my-api --non-interactive`,
    Args: cobra.ExactArgs(1),
    RunE: runInit,
}

var nonInteractive bool

func init() {
    initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false,
        "use defaults without prompting")
    rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
    projectName := args[0]

    // Validate project name
    if !projectNameRegex.MatchString(projectName) {
        return fmt.Errorf("invalid project name: must start with letter, contain only letters, numbers, hyphens, underscores")
    }

    // Check if directory exists
    if _, err := os.Stat(projectName); !os.IsNotExist(err) {
        return fmt.Errorf("directory '%s' already exists", projectName)
    }

    // Get configuration
    var cfg *config.ProjectConfig
    var err error

    if nonInteractive {
        cfg = config.NewDefaultConfig(projectName)
    } else {
        cfg, err = prompt.CollectConfig(projectName)
        if err != nil {
            return fmt.Errorf("failed to collect config: %w", err)
        }
    }

    // Print summary
    fmt.Println("\n📦 Project Configuration:")
    fmt.Printf("  Name:        %s\n", cfg.Name)
    fmt.Printf("  Module:      %s\n", cfg.ModulePath)
    fmt.Printf("  API:         %s\n", cfg.APIType)
    fmt.Printf("  Auth:        %s\n", cfg.AuthType)
    fmt.Printf("  Docker:      %v\n", cfg.WithDocker)
    fmt.Printf("  CI:          %s\n", cfg.WithCI)
    fmt.Printf("  Monitoring:  %v\n", cfg.WithMonitor)

    // TODO: Call generator (Phase 03)
    fmt.Println("\n✨ Generator not yet implemented - see Phase 03")

    return nil
}
```

### Step 6: Update main.go with version ldflags

```go
// main.go
package main

import (
    "os"

    "github.com/thinhdang/go-template/cmd"
)

// These are set via ldflags at build time
var (
    version   = "dev"
    commit    = "none"
    buildDate = "unknown"
)

func main() {
    // Pass version info to cmd package
    cmd.SetVersionInfo(version, commit, buildDate)

    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Step 7: Update Makefile for ldflags

```makefile
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"
```

## Todo List

- [x] Create cmd/root.go with global flags
- [x] Create cmd/version.go with build info
- [x] Create internal/config/project.go struct
- [x] Create internal/prompt/prompt.go with Survey
- [x] Create cmd/init.go with validation
- [x] Update main.go for version ldflags
- [x] Update Makefile ldflags
- [x] Test `go-template --help`
- [x] Test `go-template version`
- [x] Test `go-template init test-project` (prompts work)

## Success Criteria

- [x] `go-template --help` shows usage
- [x] `go-template init` validates project name
- [x] Interactive prompts collect all config options
- [x] `--non-interactive` uses defaults
- [x] Version command shows build info

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Survey dependency issues | Low | Medium | Pin to v2.x |
| Terminal compatibility | Medium | Low | Test on multiple terminals |

## Security Considerations

- Validate all user inputs
- Sanitize project names for filesystem safety
- No command injection in generated paths

## Next Steps

After completing this phase:
1. Proceed to [Phase 03: Template Engine](./phase-03-template-engine.md)
2. Implement embed.FS and template rendering
