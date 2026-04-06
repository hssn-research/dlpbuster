package gcs_test

import (
	"context"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	gcschan "github.com/hssn-research/dlpbuster/internal/channels/gcs"
	"github.com/stretchr/testify/assert"
)

func TestGCS_SkippedWhenNoBucket(t *testing.T) {
	t.Parallel()

	ch := gcschan.New()
	assert.Equal(t, "gcs", ch.Name())
	assert.NotEmpty(t, ch.Description())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}
