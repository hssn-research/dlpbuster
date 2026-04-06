package listener

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSListener listens for DNS queries on :53/udp and emits received payloads.
// Requires CAP_NET_BIND_SERVICE or root to bind to port 53.
type DNSListener struct {
	Addr     string
	received chan ReceivedPayload
}

// NewDNSListener creates a DNS listener on the given address (e.g. ":53").
func NewDNSListener(addr string) *DNSListener {
	if addr == "" {
		addr = ":53"
	}
	return &DNSListener{
		Addr:     addr,
		received: make(chan ReceivedPayload, 64),
	}
}

// Received returns the channel of received payloads.
func (l *DNSListener) Received() <-chan ReceivedPayload {
	return l.received
}

// Start begins listening for DNS UDP packets and blocks until ctx is cancelled.
func (l *DNSListener) Start(ctx context.Context) error {
	pc, err := net.ListenPacket("udp", l.Addr)
	if err != nil {
		return fmt.Errorf("dns listener: bind %s: %w", l.Addr, err)
	}
	defer pc.Close()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := pc.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			continue
		}

		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return fmt.Errorf("dns listener: read: %w", err)
		}

		// Extract subdomain from raw DNS query (offset 12 = question section)
		subdomain := extractQueryName(buf[:n])
		if subdomain == "" {
			continue
		}

		l.received <- ReceivedPayload{
			Channel:   "dns",
			Data:      []byte(subdomain),
			SourceIP:  addr.String(),
			Timestamp: time.Now().Unix(),
		}

		// Send NXDOMAIN response
		resp := buildNXDomain(buf[:n])
		pc.WriteTo(resp, addr) //nolint:errcheck
	}
}

// extractQueryName parses the QNAME from a raw DNS query packet (RFC 1035).
func extractQueryName(pkt []byte) string {
	if len(pkt) < 12 {
		return ""
	}
	pos := 12
	var labels []string
	for pos < len(pkt) {
		length := int(pkt[pos])
		if length == 0 {
			break
		}
		pos++
		if pos+length > len(pkt) {
			break
		}
		labels = append(labels, string(pkt[pos:pos+length]))
		pos += length
	}
	return strings.Join(labels, ".")
}

// buildNXDomain constructs a minimal NXDOMAIN DNS response.
func buildNXDomain(query []byte) []byte {
	if len(query) < 12 {
		return nil
	}
	resp := make([]byte, len(query))
	copy(resp, query)
	// Set QR=1 (response), RCODE=3 (NXDOMAIN)
	resp[2] = query[2] | 0x80
	resp[3] = (query[3] & 0xF0) | 0x03
	return resp
}
