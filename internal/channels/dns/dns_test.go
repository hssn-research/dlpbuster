package dns_test

import (
	"context"
	"testing"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	dnschan "github.com/hssn-research/dlpbuster/internal/channels/dns"
	"github.com/stretchr/testify/assert"
)

func TestDNS_SkippedWhenNoDomain(t *testing.T) {
	t.Parallel()

	ch := dnschan.New()
	assert.Equal(t, "dns", ch.Name())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test payload"),
		Timeout: 5 * time.Second,
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}

func TestDNS_ChunkLabel(t *testing.T) {
	t.Parallel()
	// Verify the channel can be instantiated and basic fields work
	ch := dnschan.New()
	assert.Equal(t, "dns", ch.Name())
	assert.NotEmpty(t, ch.Description())
}
