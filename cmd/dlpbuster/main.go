// Command dlpbuster is the entry point for the DLP bypass testing CLI.
package main

import (
	"os"

	"github.com/hssn-research/dlpbuster/internal/ui"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	root := newRootCmd(version)
	if err := root.Execute(); err != nil {
		ui.PrintBanner(os.Stderr)
		os.Exit(1)
	}
}
