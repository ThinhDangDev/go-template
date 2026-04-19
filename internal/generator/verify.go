package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// VerifyGeneratedProject checks that generation produced the expected root structure.
func VerifyGeneratedProject(projectDir string) error {
	requiredPaths := []string{
		"go.mod",
		"README.md",
		"Makefile",
		filepath.Join("cmd", "main.go"),
		"internal",
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(projectDir, path)
		if _, err := os.Stat(fullPath); err != nil {
			return fmt.Errorf("missing required generated path %q: %w", path, err)
		}
	}

	unexpectedRoots := []string{"base", "clean-arch", "ci-cd", "tests"}
	for _, path := range unexpectedRoots {
		fullPath := filepath.Join(projectDir, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("unexpected template root directory %q found in generated project", path)
		}
	}

	return nil
}
