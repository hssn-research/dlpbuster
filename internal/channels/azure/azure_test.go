package azure_test

import (
	"context"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	azurechan "github.com/hssn-research/dlpbuster/internal/channels/azure"
	"github.com/stretchr/testify/assert"
)

func TestAzure_SkippedWhenNotConfigured(t *testing.T) {
	t.Parallel()

	ch := azurechan.New()
	assert.Equal(t, "azure", ch.Name())
	assert.NotEmpty(t, ch.Description())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}
