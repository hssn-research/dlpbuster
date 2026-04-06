// Package channels defines the Channel interface and shared types used by all
// exfiltration channel implementations.
package channels

import (
	"context"
	"time"
)

// Status represents the outcome of a channel run.
type Status string

const (
	// StatusPassed means the payload was delivered and confirmed by the listener.
	StatusPassed Status = "PASSED"

	// StatusBlocked means the channel was blocked before delivery.
	StatusBlocked Status = "BLOCKED"

	// StatusPartial means partial delivery was confirmed (some packets/chunks received).
	StatusPartial Status = "PARTIAL"

	// StatusError means the channel encountered an unexpected error.
	StatusError Status = "ERROR"

	// StatusSkipped means the channel was skipped (e.g. requires root, not configured).
	StatusSkipped Status = "SKIPPED"
)

// Result holds the outcome of a single channel execution.
type Result struct {
	Channel   string
	Status    Status
	BytesSent int
	Duration  time.Duration
	Evidence  []string
	Error     error
}

// ChannelConfig holds configuration for a single channel run.
type ChannelConfig struct {
	// Shared
	ListenerAddr string
	Payload      []byte
	Timeout      time.Duration
	Verbose      bool

	// DNS
	DNSResolver    string
	DNSDomain      string
	DNSRecordTypes []string

	// HTTPS
	HTTPSUserAgent string
	HTTPSChunkSize int

	// ICMP
	ICMPTarget string

	// S3
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	// S3Endpoint overrides the default AWS endpoint for anonymous/public bucket testing
	// (e.g. "http://localhost:9000" for MinIO or a public test bucket URL).
	S3Endpoint string

	// GCS
	GCSBucket      string
	GCSCredentials string
	// GCSEndpoint overrides the default GCS endpoint for anonymous/public bucket testing.
	GCSEndpoint string

	// Azure
	AzureAccount   string
	AzureContainer string
	AzureSASToken  string
	// AzureEndpoint overrides the default Azure Blob endpoint for anonymous/public container testing.
	AzureEndpoint string

	// SMTP
	SMTPServer string
	SMTPFrom   string
	SMTPTo     string
	SMTPUser   string
	SMTPPass   string

	// Webhook
	WebhookURL string

	// Timing jitter: if > 0, sleep a random duration in [0, JitterMs] ms between chunks/packets.
	JitterMs int

	// SMTP subject-line steganography: encode first 8 bytes of payload into subject capitalisation.
	SMTPSubjectStego bool
}

// Channel is the interface every exfil channel must implement.
type Channel interface {
	// Name returns the short identifier for the channel (e.g. "dns", "https").
	Name() string

	// Description returns a one-line human-readable description.
	Description() string

	// Run executes the channel and returns a Result. It must respect ctx cancellation.
	Run(ctx context.Context, cfg ChannelConfig) Result
}
