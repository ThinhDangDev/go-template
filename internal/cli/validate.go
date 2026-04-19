package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ValidationResult captures post-generation checks that were run.
type ValidationResult struct {
	Checks   []string
	Warnings []string
}

type validationCommand struct {
	Name string
	Args []string
}

// ValidateGeneratedProject runs lightweight validation against a generated project.
func ValidateGeneratedProject(projectDir string) (*ValidationResult, error) {
	result := &ValidationResult{}

	if _, err := exec.LookPath("go"); err != nil {
		result.Warnings = append(result.Warnings, "Go toolchain not found; skipped go mod tidy/build/test validation")
		return result, nil
	}

	commands := []validationCommand{
		{Name: "go", Args: []string{"mod", "tidy"}},
		{Name: "go", Args: []string{"build", "./..."}},
		{Name: "go", Args: []string{"test", "./..."}},
	}

	for _, command := range commands {
		if err := runValidationCommand(projectDir, command.Name, command.Args...); err != nil {
			return result, err
		}
		result.Checks = append(result.Checks, command.Name+" "+fmt.Sprint(command.Args))
	}

	return result, nil
}

func runValidationCommand(projectDir, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = projectDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("%s %v timed out after 2m", name, args)
		}
		if stderr.Len() > 0 {
			return fmt.Errorf("%s %v failed: %s", name, args, stderr.String())
		}
		return fmt.Errorf("%s %v failed: %w", name, args, err)
	}

	return nil
}
