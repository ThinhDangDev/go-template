package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	uicli "github.com/ThinhDangDev/go-template/internal/cli"
	"github.com/ThinhDangDev/go-template/internal/config"
	"github.com/ThinhDangDev/go-template/internal/generator"
	"github.com/ThinhDangDev/go-template/internal/prompt"
	"github.com/spf13/cobra"
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
var skipValidate bool

func init() {
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false,
		"use defaults without prompting")
	initCmd.Flags().BoolVar(&skipValidate, "skip-validate", false,
		"skip post-generation validation (go mod tidy/build/test)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if !projectNameRegex.MatchString(projectName) {
		return fmt.Errorf("invalid project name %q: use a leading letter followed by letters, numbers, hyphens, or underscores", projectName)
	}

	// Check if directory exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("directory %q already exists; choose a different project name or remove the existing directory", projectName)
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

	uicli.Header("Project Configuration")
	uicli.Info("  Name:        %s", cfg.Name)
	uicli.Info("  Module:      %s", cfg.ModulePath)
	uicli.Info("  API:         %s", cfg.APIType)
	uicli.Info("  Auth:        %s", cfg.AuthType)
	uicli.Info("  Docker:      %v", cfg.WithDocker)
	uicli.Info("  CI:          %s", cfg.WithCI)
	uicli.Info("  Monitoring:  %v", cfg.WithMonitor)

	uicli.Step(1, 3, "Generating project files")
	gen := generator.New(cfg)
	spinner := uicli.NewSpinner("Generating project")
	spinner.Start()
	if err := gen.Generate(); err != nil {
		spinner.StopWithError("Generation failed")
		return fmt.Errorf("failed to generate project: %w", err)
	}
	spinner.StopWithSuccess(fmt.Sprintf("Project %q created", cfg.Name))

	uicli.Step(2, 3, "Verifying generated structure")
	if err := generator.VerifyGeneratedProject(cfg.Name); err != nil {
		return fmt.Errorf("project generated but structure verification failed: %w", err)
	}
	uicli.Success("Generated structure looks correct")

	if skipValidate {
		uicli.Warning("Skipped post-generation validation")
	} else {
		uicli.Step(3, 3, "Running post-generation validation")
		spinner = uicli.NewSpinner("Validating generated project")
		spinner.Start()
		result, err := uicli.ValidateGeneratedProject(cfg.Name)
		if err != nil {
			spinner.StopWithError("Validation failed")
			uicli.Warning("Project files were generated in %s", cfg.Name)
			return fmt.Errorf("post-generation validation failed: %w", err)
		}
		spinner.StopWithSuccess("Validation passed")
		for _, warning := range result.Warnings {
			uicli.Warning("%s", warning)
		}
	}

	uicli.Header("Next Steps")
	uicli.Info("  cd %s", filepath.Base(cfg.Name))
	uicli.Info("  make run")
	uicli.Info("  go-template completion zsh > \"${fpath[1]}/_go-template\"")

	return nil
}
