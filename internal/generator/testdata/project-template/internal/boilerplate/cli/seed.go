package cli

import (
	"context"
	"fmt"

	"__MODULE_PATH__/internal/boilerplate/app"
	"__MODULE_PATH__/internal/boilerplate/auth"

	"github.com/spf13/cobra"
)

func newSeedCmd() *cobra.Command {
	var email string
	var password string
	var role string

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed boilerplate data",
	}

	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Create or update the initial admin user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			runtime, err := app.Bootstrap(ctx)
			if err != nil {
				return err
			}
			defer func() {
				closeCtx, cancel := context.WithTimeout(context.Background(), runtime.Config.ShutdownTimeout)
				defer cancel()
				_ = runtime.Close(closeCtx)
			}()

			if email == "" {
				email = runtime.Config.AdminEmail
			}
			if password == "" {
				password = runtime.Config.AdminPassword
			}
			if role == "" {
				role = runtime.Config.AdminRole
			}
			if password == "" {
				return fmt.Errorf("admin password is required; set ADMIN_PASSWORD or pass --password")
			}

			hashedPassword, err := auth.HashPassword(password)
			if err != nil {
				return err
			}

			user, err := runtime.Users.UpsertAdmin(ctx, email, hashedPassword, role)
			if err != nil {
				return err
			}

			fmt.Printf("admin user ready: id=%s email=%s role=%s\n", user.ID, user.Email, user.Role)
			return nil
		},
	}

	adminCmd.Flags().StringVar(&email, "email", "", "admin email")
	adminCmd.Flags().StringVar(&password, "password", "", "admin password")
	adminCmd.Flags().StringVar(&role, "role", "", "admin role")
	cmd.AddCommand(adminCmd)

	return cmd
}
