package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/thinhdang/go-template/internal/config"
	"github.com/thinhdang/go-template/internal/generator"
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

	// Generate project
	fmt.Println("\n🚀 Generating project...")
	gen := generator.New(cfg)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Printf("\n✅ Project '%s' created successfully!\n", cfg.Name)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", cfg.Name)
	fmt.Println("  go mod tidy")
	fmt.Println("  make run")

	return nil
}
