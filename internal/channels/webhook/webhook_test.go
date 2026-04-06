package webhook_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	webhookchan "github.com/hssn-research/dlpbuster/internal/channels/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook_Passed(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch := webhookchan.New()
	assert.Equal(t, "webhook", ch.Name())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		WebhookURL: srv.URL,
		Payload:    []byte("dlpbuster test"),
	})

	require.Equal(t, channels.StatusPassed, result.Status)
	assert.Greater(t, result.BytesSent, 0)
}

func TestWebhook_SkippedWhenNoURL(t *testing.T) {
	t.Parallel()

	ch := webhookchan.New()
	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}

func TestWebhook_Blocked(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	ch := webhookchan.New()
	result := ch.Run(context.Background(), channels.ChannelConfig{
		WebhookURL: srv.URL,
		Payload:    []byte("test"),
	})
	assert.Equal(t, channels.StatusBlocked, result.Status)
}
