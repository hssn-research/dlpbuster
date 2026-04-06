package config

import (
	"time"
)

// DefaultTimeout is used when no timeout is configured.
const DefaultTimeout = 30 * time.Second

// defaultConfig returns a Config with all sane defaults applied.
func defaultConfig() *Config {
	return &Config{
		TimeoutSeconds: 30,
		Payload: PayloadConfig{
			SizeBytes: 1024,
			Encrypt:   false,
			Compress:  false,
		},
		Channels: ChannelsConfig{
			DNS: DNSConfig{
				Enabled:     false,
				Resolver:    "8.8.8.8:53",
				RecordTypes: []string{"TXT"},
			},
			HTTPS: HTTPSConfig{
				Enabled:   false,
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				ChunkSize: 256,
			},
			ICMP:    ICMPConfig{Enabled: false},
			S3:      S3Config{Enabled: false, Region: "us-east-1"},
			GCS:     GCSConfig{Enabled: false},
			Azure:   AzureConfig{Enabled: false},
			SMTP:    SMTPConfig{Enabled: false},
			Webhook: WebhookConfig{Enabled: false},
		},
		Output: OutputConfig{
			Format:    "human",
			ReportDir: "~/.dlpbuster/reports",
			LogDir:    "~/.dlpbuster/logs",
		},
	}
}
