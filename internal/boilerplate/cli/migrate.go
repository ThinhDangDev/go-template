package cli

import (
	"errors"
	"fmt"

	"github.com/ThinhDangDev/go-template/internal/boilerplate/app"
	"github.com/ThinhDangDev/go-template/internal/boilerplate/config"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	var downSteps int

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run manual SQL migrations",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "up",
			Short: "Apply all pending migrations",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}

				m, err := app.NewMigrator(cfg)
				if err != nil {
					return err
				}
				defer func() { _, _ = m.Close() }()

				if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
					return err
				}

				fmt.Println("migrations up completed")
				return nil
			},
		},
		&cobra.Command{
			Use:   "down",
			Short: "Rollback migrations",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}

				m, err := app.NewMigrator(cfg)
				if err != nil {
					return err
				}
				defer func() { _, _ = m.Close() }()

				if err := m.Steps(-downSteps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
					return err
				}

				fmt.Printf("rolled back %d migration step(s)\n", downSteps)
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show current migration version",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}

				m, err := app.NewMigrator(cfg)
				if err != nil {
					return err
				}
				defer func() { _, _ = m.Close() }()

				version, dirty, err := m.Version()
				if err != nil {
					if errors.Is(err, migrate.ErrNilVersion) {
						fmt.Println("version=0 dirty=false")
						return nil
					}
					return err
				}

				fmt.Printf("version=%d dirty=%t\n", version, dirty)
				return nil
			},
		},
		&cobra.Command{
			Use:   "create <name>",
			Short: "Create a new pair of up/down SQL migration files",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}

				upPath, downPath, err := app.CreateMigrationFiles(cfg.ResolvedMigrationsDir(), args[0])
				if err != nil {
					return err
				}

				fmt.Printf("created migration files:\n- %s\n- %s\n", upPath, downPath)
				return nil
			},
		},
	)

	cmd.PersistentFlags().IntVar(&downSteps, "steps", 1, "number of migration steps to rollback")
	return cmd
}
