package https_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	httpschan "github.com/hssn-research/dlpbuster/internal/channels/https"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPS_Passed(t *testing.T) {
	t.Parallel()

	var received atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		received.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch := httpschan.New()
	assert.Equal(t, "https", ch.Name())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		ListenerAddr:   srv.URL,
		Payload:        []byte("hello dlpbuster test payload"),
		HTTPSChunkSize: 8,
		Timeout:        10 * time.Second,
	})

	require.Equal(t, channels.StatusPassed, result.Status)
	assert.Greater(t, result.BytesSent, 0)
	assert.Greater(t, received.Load(), int64(0))
}

func TestHTTPS_Blocked(t *testing.T) {
	t.Parallel()

	// Point at a non-listening port
	ch := httpschan.New()
	result := ch.Run(context.Background(), channels.ChannelConfig{
		ListenerAddr: "http://127.0.0.1:19999",
		Payload:      []byte("test"),
		Timeout:      3 * time.Second,
	})

	assert.Equal(t, channels.StatusBlocked, result.Status)
}

func TestHTTPS_SkippedWhenNoAddress(t *testing.T) {
	t.Parallel()

	ch := httpschan.New()
	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}
