package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cfg := &cobra.Command{
		Use:   "config",
		Short: "Manage dlpbuster configuration",
	}
	cfg.AddCommand(newConfigInitCmd())
	return cfg
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive config setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			cfgDir := filepath.Join(home, ".dlpbuster")
			cfgFile := filepath.Join(cfgDir, "config.yaml")

			if _, err := os.Stat(cfgFile); err == nil {
				fmt.Printf("Config already exists at %s\n", cfgFile)
				fmt.Print("Overwrite? [y/N]: ")
				var answer string
				fmt.Scanln(&answer)
				if strings.ToLower(answer) != "y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			scanner := bufio.NewScanner(os.Stdin)

			ask := func(prompt, defaultVal string) string {
				if defaultVal != "" {
					fmt.Printf("%s [%s]: ", prompt, defaultVal)
				} else {
					fmt.Printf("%s: ", prompt)
				}
				scanner.Scan()
				val := strings.TrimSpace(scanner.Text())
				if val == "" {
					return defaultVal
				}
				return val
			}

				fmt.Println("\n--- dlpbuster config init ---")
			dnsAddr := ask("Callback DNS address (your VPS domain)", "exfil.example.com")
			httpsAddr := ask("Callback HTTPS address", "https://your-vps.example.com:8443")
			payloadSize := ask("Payload size in bytes", "1024")
			dnsDomain := ask("DNS exfil domain (subdomain of DNS address)", "exfil.example.com")
			webhookURL := ask("Webhook URL (Slack/Discord, or leave blank)", "")

			if err := os.MkdirAll(cfgDir, 0o750); err != nil {
				return fmt.Errorf("mkdir %s: %w", cfgDir, err)
			}
			if err := os.MkdirAll(filepath.Join(cfgDir, "reports"), 0o750); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(cfgDir, "logs"), 0o750); err != nil {
				return err
			}

			var sb strings.Builder
			sb.WriteString("# dlpbuster configuration\n\n")
			sb.WriteString("listener:\n")
			sb.WriteString(fmt.Sprintf("  dns_address: %q\n", dnsAddr))
			sb.WriteString(fmt.Sprintf("  https_address: %q\n\n", httpsAddr))
			sb.WriteString("payload:\n")
			sb.WriteString(fmt.Sprintf("  size_bytes: %s\n", payloadSize))
			sb.WriteString("  encrypt: false\n  compress: false\n\n")
			sb.WriteString("timeout_seconds: 30\n\n")
			sb.WriteString("channels:\n")
			sb.WriteString("  dns:\n    enabled: true\n    resolver: \"8.8.8.8:53\"\n")
			sb.WriteString(fmt.Sprintf("    domain: %q\n    record_types: [\"TXT\"]\n", dnsDomain))
			sb.WriteString("  https:\n    enabled: true\n    user_agent: \"Mozilla/5.0\"\n    chunk_size: 256\n")
			sb.WriteString("  icmp:\n    enabled: false\n")
			sb.WriteString("  s3:\n    enabled: false\n    bucket: \"\"\n    region: \"us-east-1\"\n")
			sb.WriteString("  gcs:\n    enabled: false\n    bucket: \"\"\n")
			sb.WriteString("  azure:\n    enabled: false\n    account: \"\"\n    container: \"\"\n")
			sb.WriteString("  smtp:\n    enabled: false\n    server: \"\"\n    from: \"\"\n    to: \"\"\n")
			if webhookURL != "" {
				sb.WriteString(fmt.Sprintf("  webhook:\n    enabled: true\n    url: %q\n", webhookURL))
			} else {
				sb.WriteString("  webhook:\n    enabled: false\n    url: \"\"\n")
			}
			sb.WriteString("\noutput:\n  format: \"human\"\n")
			sb.WriteString(fmt.Sprintf("  report_dir: %q\n", filepath.Join(cfgDir, "reports")))
			sb.WriteString(fmt.Sprintf("  log_dir: %q\n", filepath.Join(cfgDir, "logs")))

			if err := os.WriteFile(cfgFile, []byte(sb.String()), 0o600); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("\nConfig written to: %s\n", cfgFile)
			fmt.Println("Run 'dlpbuster run' to start testing.")
			return nil
		},
	}
}
