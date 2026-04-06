package main

import (
	"fmt"
	"os"

	"github.com/hssn-research/dlpbuster/internal/report"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var (
		formatFlag string
		outFlag    string
		inputFlag  string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Render a report from the last run",
		Long:  `Generate a report in the specified format. Use --input to specify a JSON results file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFlag == "" {
				return fmt.Errorf("--input is required (path to a JSON results file from a previous run)")
			}

			data, err := os.ReadFile(inputFlag)
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			_ = data // TODO: deserialise results and re-render
			fmt.Fprintf(os.Stderr, "Re-rendering %s as %s...\n", inputFlag, formatFlag)

			renderer := report.RendererFor(report.Format(formatFlag))
			_ = renderer

			return fmt.Errorf("re-render from file not yet implemented — run dlpbuster run --format %s instead", formatFlag)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "markdown", "Output format: human|json|markdown|html")
	cmd.Flags().StringVar(&outFlag, "out", "", "Write report to file")
	cmd.Flags().StringVar(&inputFlag, "input", "", "Path to a JSON results file")

	return cmd
}
