package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/config"
	"github.com/hssn-research/dlpbuster/internal/engine"
	"github.com/hssn-research/dlpbuster/internal/payload"
	"github.com/hssn-research/dlpbuster/internal/registry"
	"github.com/hssn-research/dlpbuster/internal/report"
	"github.com/hssn-research/dlpbuster/internal/ui"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var (
		channelFlag string
		payloadSize int
		timeout     int
		formatFlag  string
		outFlag     string
		payloadFile string
		encrypt     bool
		compress    bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run exfil channel tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}

			// CLI flags override config
			if payloadSize > 0 {
				cfg.Payload.SizeBytes = payloadSize
			}
			if timeout > 0 {
				cfg.Timeout = time.Duration(timeout) * time.Second
			}
			if encrypt {
				cfg.Payload.Encrypt = true
			}
			if compress {
				cfg.Payload.Compress = true
			}

			// Build channel list
			selected, err := selectChannels(channelFlag, cfg)
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no channels enabled — run 'dlpbuster config init' or use --channel flag")
			}

			// Generate payload
			gen := &payload.Generator{
				SizeBytes: cfg.Payload.SizeBytes,
				FilePath:  payloadFile,
			}
			rawPayload, err := gen.Generate()
			if err != nil {
				return fmt.Errorf("payload: %w", err)
			}
			if cfg.Payload.Compress {
				rawPayload, err = payload.Compress(rawPayload)
				if err != nil {
					return fmt.Errorf("compress: %w", err)
				}
			}
			if cfg.Payload.Encrypt {
				key, err := payload.RandomKey()
				if err != nil {
					return fmt.Errorf("encrypt key: %w", err)
				}
				rawPayload, err = payload.Encrypt(key, rawPayload)
				if err != nil {
					return fmt.Errorf("encrypt: %w", err)
				}
			}

			// Build channel config
			chCfg := buildChannelConfig(cfg, rawPayload)

			silent, _ := cmd.Flags().GetBool("silent")
			if !silent {
				fmt.Fprintf(os.Stderr, "Running %d channels | Payload: %d bytes | Timeout: %s\n\n",
					len(selected), len(rawPayload), cfg.Timeout)
			}

			// Trap signals
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			// Run
			results := engine.Run(ctx, engine.RunConfig{
				Channels:   selected,
				ChannelCfg: chCfg,
				Timeout:    cfg.Timeout,
			})

			// Render
			format := report.Format(formatFlag)
			if format == "" {
				format = report.Format(cfg.Output.Format)
			}

			r := report.New(results, version, cfg.Listener.DNSAddress, len(rawPayload), cfg.Timeout)
			renderer := report.RendererFor(format)
			b, err := renderer.Render(r)
			if err != nil {
				return fmt.Errorf("render: %w", err)
			}

			if outFlag != "" {
				if err := os.MkdirAll(filepath.Dir(outFlag), 0o750); err != nil {
					return fmt.Errorf("mkdir: %w", err)
				}
				if err := os.WriteFile(outFlag, b, 0o600); err != nil {
					return fmt.Errorf("write report: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Report written to: %s\n", outFlag)
			} else {
				if format != report.FormatHuman && !silent {
					ui.PrintTable(os.Stderr, results)
				}
				fmt.Println(string(b))
			}

			// Auto-save default report
			if outFlag == "" && format == report.FormatHuman {
				autoSave(cfg, results, r)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&channelFlag, "channel", "", "Comma-separated channels to run (default: all enabled)")
	cmd.Flags().IntVar(&payloadSize, "payload-size", 0, "Synthetic payload size in bytes (default: from config)")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "Per-channel timeout in seconds (default: from config)")
	cmd.Flags().StringVar(&formatFlag, "format", "human", "Output format: human|json|markdown|html")
	cmd.Flags().StringVar(&outFlag, "out", "", "Write report to file")
	cmd.Flags().StringVar(&payloadFile, "payload-file", "", "Use this file as payload instead of random bytes")
	cmd.Flags().BoolVar(&encrypt, "encrypt", false, "Encrypt payload before exfil")
	cmd.Flags().BoolVar(&compress, "compress", false, "Compress payload before exfil")

	return cmd
}

func selectChannels(channelFlag string, cfg *config.Config) ([]channels.Channel, error) {
	all := registry.All()

	if channelFlag != "" {
		names := strings.Split(channelFlag, ",")
		var out []channels.Channel
		for _, name := range names {
			name = strings.TrimSpace(name)
			ch, err := registry.Lookup(name)
			if err != nil {
				return nil, err
			}
			out = append(out, ch)
		}
		return out, nil
	}

	// Use enabled channels from config
	enabled := map[string]bool{
		"dns":     cfg.Channels.DNS.Enabled,
		"https":   cfg.Channels.HTTPS.Enabled,
		"icmp":    cfg.Channels.ICMP.Enabled,
		"s3":      cfg.Channels.S3.Enabled,
		"gcs":     cfg.Channels.GCS.Enabled,
		"azure":   cfg.Channels.Azure.Enabled,
		"smtp":    cfg.Channels.SMTP.Enabled,
		"webhook": cfg.Channels.Webhook.Enabled,
	}

	var out []channels.Channel
	for _, ch := range all {
		if enabled[ch.Name()] {
			out = append(out, ch)
		}
	}
	return out, nil
}

func buildChannelConfig(cfg *config.Config, p []byte) channels.ChannelConfig {
	return channels.ChannelConfig{
		ListenerAddr:   cfg.Listener.HTTPSAddress,
		Payload:        p,
		Timeout:        cfg.Timeout,
		DNSResolver:    cfg.Channels.DNS.Resolver,
		DNSDomain:      cfg.Channels.DNS.Domain,
		DNSRecordTypes: cfg.Channels.DNS.RecordTypes,
		HTTPSUserAgent: cfg.Channels.HTTPS.UserAgent,
		HTTPSChunkSize: cfg.Channels.HTTPS.ChunkSize,
		ICMPTarget:     cfg.Channels.ICMP.Target,
		S3Bucket:       cfg.Channels.S3.Bucket,
		S3Region:       cfg.Channels.S3.Region,
		S3AccessKey:    cfg.Channels.S3.AccessKey,
		S3SecretKey:    cfg.Channels.S3.SecretKey,
		GCSBucket:      cfg.Channels.GCS.Bucket,
		GCSCredentials: cfg.Channels.GCS.Credentials,
		AzureAccount:   cfg.Channels.Azure.Account,
		AzureContainer: cfg.Channels.Azure.Container,
		AzureSASToken:  cfg.Channels.Azure.SASToken,
		SMTPServer:     cfg.Channels.SMTP.Server,
		SMTPFrom:       cfg.Channels.SMTP.From,
		SMTPTo:         cfg.Channels.SMTP.To,
		SMTPUser:       cfg.Channels.SMTP.Username,
		SMTPPass:       cfg.Channels.SMTP.Password,
		WebhookURL:     cfg.Channels.Webhook.URL,
	}
}

func autoSave(cfg *config.Config, results []channels.Result, r *report.Report) {
	dir := cfg.Output.ReportDir
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return
	}
	fname := filepath.Join(dir, fmt.Sprintf("dlpbuster-report-%s.md", time.Now().Format("2006-01-02T15-04-05")))
	b, err := report.RendererFor(report.FormatMarkdown).Render(r)
	if err != nil {
		return
	}
	_ = os.WriteFile(fname, b, 0o600)
	fmt.Fprintf(os.Stderr, "Report: %s\n", fname)

	_ = results // already in r
}
