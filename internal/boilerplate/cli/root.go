package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-template",
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
