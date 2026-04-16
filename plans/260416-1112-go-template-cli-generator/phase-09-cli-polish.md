# Phase 09: CLI Polish & Validation

---
status: not-started
priority: P1
effort: 2h
dependencies: [phase-08]
blocked-by: "Phase 08 (CI/CD) not started; also needed to fix project structure issue"
critical-for-release: true
---

## CRITICAL: This Phase Must Address Project Structure Bug

**Issue:** Generated projects place files in nested directories (base/, clean-arch/, etc) instead of project root.

**Must Fix Before Release:**
- Strip first-level directory from base templates during generation
- Reorganize template structure OR add flattening logic in generator
- Verify generated projects have correct structure

## Context Links

- [Main Plan](./plan.md)
- [Previous: CI/CD & Testing](./phase-08-cicd-testing.md)
- [Next: Release & Distribution](./phase-10-release.md)

## Overview

Polish the CLI with improved error handling, validation, progress indicators, colored output, and post-generation verification. Ensure a smooth user experience.

## Key Insights

- Clear error messages reduce user frustration
- Progress indicators for longer operations
- Post-generation validation catches template errors early
- Colored output improves readability
- Shell completion increases productivity

## Requirements

### Functional
- Input validation with helpful error messages
- Progress indicator during generation
- Post-generation validation (go build, go mod tidy)
- Shell completion scripts (bash, zsh, fish)
- Colored output (disable with --no-color)

### Non-Functional
- All errors actionable
- Sub-second feedback for user actions
- Graceful degradation without color support

## Architecture

```
internal/
├── cli/
│   ├── color.go         # Colored output helpers
│   ├── progress.go      # Progress indicator
│   └── validate.go      # Post-generation validation
└── generator/
    └── verify.go        # Template verification
```

## Related Code Files

### Files to Create
- `internal/cli/color.go`
- `internal/cli/progress.go`
- `internal/cli/validate.go`
- `internal/generator/verify.go`
- `cmd/completion.go`

### Files to Modify
- `cmd/init.go` - Add progress, validation
- `cmd/root.go` - Add --no-color flag

## Implementation Steps

### Step 1: Create Color Helpers

```go
// internal/cli/color.go
package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
)

var colorEnabled = true

// DisableColor turns off colored output
func DisableColor() {
	colorEnabled = false
}

// IsColorEnabled checks if color is supported and enabled
func IsColorEnabled() bool {
	if !colorEnabled {
		return false
	}
	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

// colorize applies color if enabled
func colorize(color, text string) string {
	if !IsColorEnabled() {
		return text
	}
	return color + text + Reset
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Green, "✓ "+msg))
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Red, "✗ "+msg))
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Yellow, "⚠ "+msg))
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Cyan, "ℹ "+msg))
}

// Step prints a step indicator
func Step(step, total int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	prefix := fmt.Sprintf("[%d/%d]", step, total)
	fmt.Println(colorize(Blue, prefix) + " " + msg)
}

// Header prints a bold header
func Header(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Bold, msg))
}
```

### Step 2: Create Progress Indicator

```go
// internal/cli/progress.go
package cli

import (
	"fmt"
	"sync"
	"time"
)

// Spinner shows a progress spinner
type Spinner struct {
	message string
	frames  []string
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stop:    make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	if !IsColorEnabled() {
		fmt.Printf("%s...\n", s.message)
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Printf("\r\033[K")
				return
			default:
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r%s %s", colorize(Cyan, frame), s.message)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	close(s.stop)
	s.wg.Wait()
}

// StopWithSuccess ends spinner with success message
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	Success(message)
}

// StopWithError ends spinner with error message
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	Error(message)
}
```

### Step 3: Create Post-Generation Validation

```go
// internal/cli/validate.go
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidationResult holds validation results
type ValidationResult struct {
	Step    string
	Success bool
	Output  string
	Error   error
}

// ValidateProject runs post-generation validation
func ValidateProject(projectPath string) []ValidationResult {
	results := []ValidationResult{}

	// Change to project directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(projectPath); err != nil {
		return append(results, ValidationResult{
			Step:    "change directory",
			Success: false,
			Error:   err,
		})
	}
	defer os.Chdir(originalDir)

	// Step 1: Check go.mod exists
	if _, err := os.Stat("go.mod"); err != nil {
		results = append(results, ValidationResult{
			Step:    "check go.mod",
			Success: false,
			Error:   fmt.Errorf("go.mod not found"),
		})
		return results
	}
	results = append(results, ValidationResult{
		Step:    "check go.mod",
		Success: true,
	})

	// Step 2: Run go mod tidy
	tidyResult := runCommand("go", "mod", "tidy")
	results = append(results, tidyResult)
	if !tidyResult.Success {
		return results
	}

	// Step 3: Run go build
	buildResult := runCommand("go", "build", "./...")
	results = append(results, buildResult)
	if !buildResult.Success {
		return results
	}

	// Step 4: Check for obvious issues
	checkResult := runCommand("go", "vet", "./...")
	results = append(results, checkResult)

	return results
}

func runCommand(name string, args ...string) ValidationResult {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	step := name + " " + strings.Join(args, " ")

	if err != nil {
		return ValidationResult{
			Step:    step,
			Success: false,
			Output:  string(output),
			Error:   err,
		}
	}

	return ValidationResult{
		Step:    step,
		Success: true,
		Output:  string(output),
	}
}

// PrintValidationResults displays validation results
func PrintValidationResults(results []ValidationResult) {
	fmt.Println()
	Header("Validation Results")
	fmt.Println()

	allPassed := true
	for _, r := range results {
		if r.Success {
			Success(r.Step)
		} else {
			Error(r.Step)
			if r.Output != "" {
				fmt.Printf("  Output: %s\n", strings.TrimSpace(r.Output))
			}
			allPassed = false
		}
	}

	fmt.Println()
	if allPassed {
		Success("All validations passed!")
	} else {
		Warning("Some validations failed. Check the output above.")
	}
}
```

### Step 4: Create Generator Verification

```go
// internal/generator/verify.go
package generator

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/thinhdang/go-template/templates"
)

// VerifyTemplates checks all templates for syntax errors
func VerifyTemplates() error {
	var errors []string

	err := fs.WalkDir(templates.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tmpl") && !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read error: %v", path, err))
			return nil
		}

		// Try to parse as template
		_, err = template.New(filepath.Base(path)).
			Funcs(FuncMap()).
			Parse(string(content))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: parse error: %v", path, err))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk templates: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("template errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
```

### Step 5: Create Completion Command

```go
// cmd/completion.go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for go-template.

To load completions:

Bash:
  $ source <(go-template completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ go-template completion bash > /etc/bash_completion.d/go-template
  # macOS:
  $ go-template completion bash > $(brew --prefix)/etc/bash_completion.d/go-template

Zsh:
  # If shell completion is not already enabled, enable it:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ go-template completion zsh > "${fpath[1]}/_go-template"
  # You may need to start a new shell for this setup to take effect.

Fish:
  $ go-template completion fish | source
  # To load completions for each session, execute once:
  $ go-template completion fish > ~/.config/fish/completions/go-template.fish

PowerShell:
  PS> go-template completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> go-template completion powershell > go-template.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
```

### Step 6: Update init.go with Progress and Validation

```go
// Updated runInit function in cmd/init.go
func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if !projectNameRegex.MatchString(projectName) {
		cli.Error("Invalid project name: must start with letter, contain only letters, numbers, hyphens, underscores")
		return fmt.Errorf("invalid project name")
	}

	// Check if directory exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		cli.Error("Directory '%s' already exists", projectName)
		return fmt.Errorf("directory exists")
	}

	// Get configuration
	var cfg *config.ProjectConfig
	var err error

	if nonInteractive {
		cfg = config.NewDefaultConfig(projectName)
		cli.Info("Using default configuration (--non-interactive)")
	} else {
		cfg, err = prompt.CollectConfig(projectName)
		if err != nil {
			return fmt.Errorf("failed to collect config: %w", err)
		}
	}

	// Print configuration summary
	fmt.Println()
	cli.Header("Project Configuration")
	fmt.Printf("  Name:        %s\n", cfg.Name)
	fmt.Printf("  Module:      %s\n", cfg.ModulePath)
	fmt.Printf("  API:         %s\n", cfg.APIType)
	fmt.Printf("  Auth:        %s\n", cfg.AuthType)
	fmt.Printf("  Docker:      %v\n", cfg.WithDocker)
	fmt.Printf("  CI:          %s\n", cfg.WithCI)
	fmt.Printf("  Monitoring:  %v\n", cfg.WithMonitor)
	fmt.Println()

	// Generate project with progress
	cli.Header("Generating Project")
	fmt.Println()

	spinner := cli.NewSpinner("Creating project structure")
	spinner.Start()

	gen := generator.New(cfg.Name, verbose)
	if err := gen.Generate(cfg); err != nil {
		spinner.StopWithError("Generation failed")
		return fmt.Errorf("generation failed: %w", err)
	}

	spinner.StopWithSuccess("Project structure created")

	// Run validation if not skipped
	if !skipValidation {
		fmt.Println()
		cli.Header("Validating Generated Project")
		fmt.Println()

		validationSpinner := cli.NewSpinner("Running validation checks")
		validationSpinner.Start()

		results := cli.ValidateProject(cfg.Name)
		validationSpinner.Stop()

		cli.PrintValidationResults(results)
	}

	// Print next steps
	fmt.Println()
	cli.Header("Next Steps")
	fmt.Println()
	fmt.Printf("  cd %s\n", cfg.Name)
	fmt.Println("  go mod tidy")
	if cfg.WithDocker {
		fmt.Println("  make docker-up  # Start dependencies")
	}
	fmt.Println("  make run        # Start the service")
	fmt.Println()
	cli.Success("Project '%s' created successfully!", cfg.Name)

	return nil
}
```

### Step 7: Update root.go with Global Flags

```go
// cmd/root.go additions
var (
	verbose        bool
	noColor        bool
	skipValidation bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&skipValidation, "skip-validation", false, "skip post-generation validation")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if noColor {
		cli.DisableColor()
	}
}
```

## Todo List

- [ ] Create internal/cli/color.go with color helpers
- [ ] Create internal/cli/progress.go with spinner
- [ ] Create internal/cli/validate.go with validation
- [ ] Create internal/generator/verify.go
- [ ] Create cmd/completion.go
- [ ] Update cmd/init.go with progress and validation
- [ ] Update cmd/root.go with global flags
- [ ] Test colored output on different terminals
- [ ] Test --no-color flag works
- [ ] Test shell completion scripts
- [ ] Test validation catches broken templates

## Success Criteria

- [ ] Progress spinner shows during generation
- [ ] Colored output works in supported terminals
- [ ] --no-color disables colors
- [ ] Post-generation validation runs
- [ ] Validation catches build errors
- [ ] Shell completion works (bash, zsh, fish)
- [ ] Error messages are clear and actionable

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Terminal compatibility | Medium | Low | Fallback to no-color |
| Validation false positives | Low | Medium | Allow --skip-validation |
| Spinner performance | Low | Low | Keep animation simple |

## Security Considerations

- No sensitive data in progress output
- Shell completion doesn't expose secrets
- Validation runs in project directory only

## Next Steps

After completing this phase:
1. Proceed to [Phase 10: Release & Distribution](./phase-10-release.md)
2. Set up GoReleaser and distribution channels
