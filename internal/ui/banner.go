// Package ui provides terminal output helpers: banner, table, and progress.
package ui

import (
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

// PrintBanner writes the dlpbuster startup banner to w.
func PrintBanner(w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, "\ndlpbuster %s — DLP Bypass Tester\n", version)
	fmt.Fprintln(w, "[!] For authorized testing only. Obtain written permission before use.")
}
