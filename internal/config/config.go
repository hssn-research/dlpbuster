// Package config loads and validates dlpbuster configuration.
package config

import (
"fmt"
"os"
"path/filepath"
"strings"
"time"

"gopkg.in/yaml.v3"
)

// Config is the root configuration struct.
type Config struct {
	Listener       Listener       `yaml:"listener"`
	Payload        PayloadConfig  `yaml:"payload"`
	Timeout        time.Duration  `yaml:"-"`
	Channels       ChannelsConfig `yaml:"channels"`
	Output         OutputConfig   `yaml:"output"`
	TimeoutSeconds int            `yaml:"timeout_seconds"`
}

// Listener holds the callback server addresses.
type Listener struct {
	DNSAddress   string `yaml:"dns_address"`
	HTTPSAddress string `yaml:"https_address"`
}

// PayloadConfig controls payload generation.
type PayloadConfig struct {
	SizeBytes int    `yaml:"size_bytes"`
	Encrypt   bool   `yaml:"encrypt"`
	Compress  bool   `yaml:"compress"`
	FilePath  string `yaml:"file_path"`
}

// ChannelsConfig holds per-channel configuration blocks.
type ChannelsConfig struct {
	DNS     DNSConfig     `yaml:"dns"`
	HTTPS   HTTPSConfig   `yaml:"https"`
	ICMP    ICMPConfig    `yaml:"icmp"`
	S3      S3Config      `yaml:"s3"`
	GCS     GCSConfig     `yaml:"gcs"`
	Azure   AzureConfig   `yaml:"azure"`
	SMTP    SMTPConfig    `yaml:"smtp"`
	Webhook WebhookConfig `yaml:"webhook"`
}

// DNSConfig holds DNS channel settings.
type DNSConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Resolver    string   `yaml:"resolver"`
	Domain      string   `yaml:"domain"`
	RecordTypes []string `yaml:"record_types"`
}

// HTTPSConfig holds HTTPS channel settings.
type HTTPSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	UserAgent  string `yaml:"user_agent"`
	ChunkSize  int    `yaml:"chunk_size"`
	SkipVerify bool   `yaml:"skip_verify"`
}

// ICMPConfig holds ICMP channel settings.
type ICMPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Target  string `yaml:"target"`
}

// S3Config holds AWS S3 channel settings.
type S3Config struct {
	Enabled   bool   `yaml:"enabled"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

// GCSConfig holds GCP Cloud Storage channel settings.
type GCSConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Bucket      string `yaml:"bucket"`
	Credentials string `yaml:"credentials"`
}

// AzureConfig holds Azure Blob Storage channel settings.
type AzureConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Account   string `yaml:"account"`
	Container string `yaml:"container"`
	SASToken  string `yaml:"sas_token"`
}

// SMTPConfig holds SMTP/email channel settings.
type SMTPConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Server   string `yaml:"server"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// WebhookConfig holds SaaS webhook channel settings.
type WebhookConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// OutputConfig controls report and log output.
type OutputConfig struct {
	Format    string `yaml:"format"`
	ReportDir string `yaml:"report_dir"`
	LogDir    string `yaml:"log_dir"`
}

// Load reads configuration from ~/.dlpbuster/config.yaml and applies defaults.
// Returns the default config if no file exists.
func Load() (*Config, error) {
	cfg := defaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config: home dir: %w", err)
	}

	cfgPath := filepath.Join(home, ".dlpbuster", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return applyComputed(cfg), nil
		}
		return nil, fmt.Errorf("config: read %s: %w", cfgPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}

	// Allow env overrides for sensitive fields
	if v := os.Getenv("DLPBUSTER_S3_ACCESS_KEY"); v != "" {
		cfg.Channels.S3.AccessKey = v
	}
	if v := os.Getenv("DLPBUSTER_S3_SECRET_KEY"); v != "" {
		cfg.Channels.S3.SecretKey = v
	}
	if v := os.Getenv("DLPBUSTER_WEBHOOK_URL"); v != "" {
		cfg.Channels.Webhook.URL = v
	}

	return applyComputed(cfg), nil
}

func applyComputed(cfg *Config) *Config {
	if cfg.TimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	} else {
		cfg.Timeout = DefaultTimeout
	}
	cfg.Output.ReportDir = expandHome(cfg.Output.ReportDir)
	cfg.Output.LogDir = expandHome(cfg.Output.LogDir)
	return cfg
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
