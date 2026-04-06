package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hssn-research/dlpbuster/internal/listener"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		dnsFlag   bool
		httpsFlag bool
		dnsAddr   string
		httpsAddr string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the callback listener server",
		Long:  `Starts DNS and/or HTTPS listeners to confirm payload receipt from channel runs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !dnsFlag && !httpsFlag {
				return fmt.Errorf("specify at least one listener: --dns or --https")
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			if httpsFlag {
				l := listener.NewHTTPSListener(httpsAddr)
				fmt.Fprintf(os.Stderr, "HTTPS listener starting on %s\n", httpsAddr)
				go func() {
					for p := range l.Received() {
						fmt.Printf("[%s] received %d bytes from %s (channel: %s)\n",
							p.Channel, len(p.Data), p.SourceIP, p.Channel)
					}
				}()
				go func() {
					if err := l.Start(ctx); err != nil {
						fmt.Fprintf(os.Stderr, "https listener error: %v\n", err)
					}
				}()
			}

			if dnsFlag {
				l := listener.NewDNSListener(dnsAddr)
				fmt.Fprintf(os.Stderr, "DNS listener starting on %s\n", dnsAddr)
				go func() {
					for p := range l.Received() {
						fmt.Printf("[dns] received query: %s from %s\n", string(p.Data), p.SourceIP)
					}
				}()
				go func() {
					if err := l.Start(ctx); err != nil {
						fmt.Fprintf(os.Stderr, "dns listener error: %v\n", err)
					}
				}()
			}

			fmt.Fprintln(os.Stderr, "Listening. Press Ctrl+C to stop.")
			<-ctx.Done()
			return nil
		},
	}

	cmd.Flags().BoolVar(&dnsFlag, "dns", false, "Start DNS listener")
	cmd.Flags().BoolVar(&httpsFlag, "https", false, "Start HTTPS listener")
	cmd.Flags().StringVar(&dnsAddr, "dns-addr", ":53", "DNS listener address (requires root for :53)")
	cmd.Flags().StringVar(&httpsAddr, "https-addr", ":8443", "HTTPS listener address")

	return cmd
}
