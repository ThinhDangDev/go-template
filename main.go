package main

import (
	"os"

	"github.com/thinhdang/go-template/cmd"
)

// These are set via ldflags at build time
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	// Pass version info to cmd package
	cmd.SetVersionInfo(version, commit, buildDate)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
