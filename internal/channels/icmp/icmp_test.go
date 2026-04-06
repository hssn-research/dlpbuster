package icmp_test

import (
	"context"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	icmpchan "github.com/hssn-research/dlpbuster/internal/channels/icmp"
	"github.com/stretchr/testify/assert"
)

func TestICMP_SkippedWithoutRoot(t *testing.T) {
	t.Parallel()

	// In CI and typical test runs we are not root, so this should be skipped.
	ch := icmpchan.New()
	assert.Equal(t, "icmp", ch.Name())
	assert.NotEmpty(t, ch.Description())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		ICMPTarget: "127.0.0.1",
		Payload:    []byte("test"),
	})
	// Either SKIPPED (not root) or ERROR (unexpected) — not PASSED/BLOCKED in unit tests
	assert.True(t, result.Status == channels.StatusSkipped || result.Status == channels.StatusError,
		"expected skipped or error, got %s", result.Status)
}

func TestICMP_SkippedWhenNoTarget(t *testing.T) {
	t.Parallel()

	ch := icmpchan.New()
	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}
