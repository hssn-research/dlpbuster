package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// PrintTable writes a results table to w using ANSI color codes.
func PrintTable(w io.Writer, results []channels.Result) {
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  %-14s %-8s %8s   %s\n", "CHANNEL", "STATUS", "DURATION", "EVIDENCE")
	fmt.Fprintln(w, "  "+strings.Repeat("─", 70))

	for _, r := range results {
		color := colorFor(r.Status)
		reset := "\033[0m"
		ev := ""
		if len(r.Evidence) > 0 {
			ev = r.Evidence[len(r.Evidence)-1]
		}
		if len(ev) > 50 {
			ev = ev[:47] + "..."
		}
		fmt.Fprintf(w, "  %-14s %s%s %-7s%s  %6dms   %s\n",
			r.Channel, color, statusIcon(r.Status), r.Status, reset,
			r.Duration.Milliseconds(), ev)
	}
	fmt.Fprintln(w, "")
}

func statusIcon(s channels.Status) string {
	switch s {
	case channels.StatusPassed:
		return "✓"
	case channels.StatusBlocked:
		return "✗"
	case channels.StatusPartial:
		return "~"
	case channels.StatusError:
		return "!"
	default:
		return "-"
	}
}

func colorFor(s channels.Status) string {
	switch s {
	case channels.StatusPassed:
		return "\033[32m" // green
	case channels.StatusBlocked:
		return "\033[31m" // red
	case channels.StatusPartial:
		return "\033[33m" // yellow
	case channels.StatusError:
		return "\033[33m" // yellow
	default:
		return "\033[90m" // gray
	}
}
