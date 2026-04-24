package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "__PROJECT_NAME__",
	Short: "CLI-first Go backend template",
}

func Execute() {
	rootCmd.AddCommand(
		newServeCmd(),
		newMigrateCmd(),
		newSeedCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
