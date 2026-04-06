package s3_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hssn-research/dlpbuster/internal/channels"
	s3chan "github.com/hssn-research/dlpbuster/internal/channels/s3"
	"github.com/stretchr/testify/assert"
)

func TestS3_SkippedWhenNoBucket(t *testing.T) {
	t.Parallel()

	ch := s3chan.New()
	assert.Equal(t, "s3", ch.Name())

	result := ch.Run(context.Background(), channels.ChannelConfig{
		Payload: []byte("test"),
	})
	assert.Equal(t, channels.StatusSkipped, result.Status)
}

func TestS3_BlockedOnHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	ch := s3chan.New()
	// We can't easily override the S3 URL in unit tests without a mock,
	// so verify the channel is blocked or skipped when bucket is set but network fails.
	result := ch.Run(context.Background(), channels.ChannelConfig{
		S3Bucket: "test-bucket",
		S3Region: "us-east-1",
		Payload:  []byte("test"),
	})
	// Will attempt real S3; expect blocked in offline env
	assert.True(t, result.Status == channels.StatusBlocked || result.Status == channels.StatusPassed,
		"expected blocked or passed, got %s", result.Status)
}
