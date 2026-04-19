package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyGeneratedProject(t *testing.T) {
	t.Run("accepts expected structure", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "cmd"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "internal"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module test\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("# test\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "Makefile"), []byte("test:\n\tgo test ./...\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "cmd", "main.go"), []byte("package main\nfunc main() {}\n"), 0644))

		assert.NoError(t, VerifyGeneratedProject(projectDir))
	})

	t.Run("rejects leaked template roots", func(t *testing.T) {
		projectDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "cmd"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "internal"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "base"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module test\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("# test\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "Makefile"), []byte("test:\n\tgo test ./...\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "cmd", "main.go"), []byte("package main\nfunc main() {}\n"), 0644))

		err := VerifyGeneratedProject(projectDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected template root directory")
	})
}
