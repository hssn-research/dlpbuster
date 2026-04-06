package main

import (
	"fmt"

	"github.com/hssn-research/dlpbuster/internal/registry"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available exfil channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%-10s  %s\n", "CHANNEL", "DESCRIPTION")
			fmt.Printf("%-10s  %s\n", "-------", "-----------")
			for _, ch := range registry.All() {
				fmt.Printf("%-10s  %s\n", ch.Name(), ch.Description())
			}
			return nil
		},
	}
}
