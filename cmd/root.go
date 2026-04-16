package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-template",
	Short: "Go backend boilerplate CLI generator",
	Long: `go-template is a CLI tool that generates production-ready
Go backend projects with Clean Architecture, authentication,
monitoring, and comprehensive testing out of the box.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
