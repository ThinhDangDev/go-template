package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Set via ldflags at build time or via SetVersionInfo()
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// SetVersionInfo sets version information from main package
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-template %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built: %s\n", buildDate)
		fmt.Printf("  go: %s\n", runtime.Version())
		fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
