package cmd

import (
	uicli "github.com/ThinhDangDev/go-template/internal/cli"
	"github.com/spf13/cobra"
)

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "go-template",
	Short: "Go backend boilerplate CLI generator",
	Long: `go-template is a CLI tool that generates production-ready
Go backend projects with Clean Architecture, authentication,
monitoring, and comprehensive testing out of the box.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		uicli.Errorf("%v", err)
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(func() {
		if noColor {
			uicli.DisableColor()
		}
	})

	// Global flags can be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
