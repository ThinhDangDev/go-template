package cli

import (
	"fmt"
	"path/filepath"

	"github.com/ThinhDangDev/go-template/internal/generator"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var modulePath string

	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Initialize a new project from the embedded backend template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			cfg := generator.Config{
				ProjectName: filepath.Base(targetDir),
				ModulePath:  modulePath,
				TargetDir:   targetDir,
			}

			if err := generator.InitProject(cfg); err != nil {
				return err
			}

			if cfg.ModulePath == "" {
				cfg.ModulePath = cfg.ProjectName
			}

			fmt.Printf("project created at %s\n", targetDir)
			fmt.Println("next steps:")
			fmt.Printf("  cd %s\n", targetDir)
			fmt.Println("  cp .env.example .env")
			fmt.Println("  set JWT_SECRET and ADMIN_PASSWORD in .env")
			fmt.Println("  make migrate-up")
			fmt.Println("  make seed-admin")
			fmt.Println("  make run")

			return nil
		},
	}

	cmd.Flags().StringVar(&modulePath, "module", "", "Go module path for the generated project")
	return cmd
}
