package smtp_test

import (
	"context"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	smtpchan "github.com/hssn-research/dlpbuster/internal/channels/smtp"
	"github.com/stretchr/testify/assert"
)

func TestSMTP_SkippedWhenNotConfigured(t *testing.T) {
	t.Parallel()

	ch := smtpchan.New()
	assert.Equal(t, "smtp", ch.Name())
	assert.NotEmpty(t, ch.Description())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}
