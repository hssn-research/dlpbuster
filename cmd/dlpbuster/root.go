package main

import (
	"os"

	"github.com/hssn-research/dlpbuster/internal/config"
	"github.com/hssn-research/dlpbuster/internal/ui"
	"github.com/spf13/cobra"
)

func newRootCmd(ver string) *cobra.Command {
	root := &cobra.Command{
		Use:   "dlpbuster",
		Short: "Automated DLP control bypass tester",
		Long: `dlpbuster systematically tests whether your DLP controls block data
leaving your environment across DNS, HTTPS, ICMP, cloud, email, and SaaS channels.

[!] For authorized testing only. Obtain written permission before use.`,
		Version:       ver,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Print banner unless --silent or running list/config/version
			silent, _ := cmd.Flags().GetBool("silent")
			if !silent {
				ui.PrintBanner(os.Stderr)
			}
			return nil
		},
	}

	root.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	root.PersistentFlags().Bool("silent", false, "Suppress all output except results")

	root.AddCommand(
		newRunCmd(),
		newServeCmd(),
		newListCmd(),
		newReportCmd(),
		newConfigCmd(),
	)

	_ = config.DefaultTimeout // ensure config package is linked

	return root
}
