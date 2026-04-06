package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubChannel is a test double for the Channel interface.
type stubChannel struct {
	name   string
	status channels.Status
	delay  time.Duration
}

func (s *stubChannel) Name() string        { return s.name }
func (s *stubChannel) Description() string { return "stub" }
func (s *stubChannel) Run(ctx context.Context, _ channels.ChannelConfig) channels.Result {
	select {
	case <-time.After(s.delay):
		return channels.Result{Status: s.status, BytesSent: 128}
	case <-ctx.Done():
		return channels.Result{Status: channels.StatusBlocked, Error: ctx.Err()}
	}
}

func TestRun_AllPassed(t *testing.T) {
	t.Parallel()

	chans := []channels.Channel{
		&stubChannel{name: "a", status: channels.StatusPassed},
		&stubChannel{name: "b", status: channels.StatusPassed},
		&stubChannel{name: "c", status: channels.StatusBlocked},
	}

	results := engine.Run(context.Background(), engine.RunConfig{
		Channels: chans,
		Timeout:  5 * time.Second,
	})

	require.Len(t, results, 3)
	assert.Equal(t, channels.StatusPassed, results[0].Status)
	assert.Equal(t, channels.StatusPassed, results[1].Status)
	assert.Equal(t, channels.StatusBlocked, results[2].Status)
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	chans := []channels.Channel{
		&stubChannel{name: "slow", status: channels.StatusPassed, delay: 5 * time.Second},
	}

	results := engine.Run(context.Background(), engine.RunConfig{
		Channels: chans,
		Timeout:  100 * time.Millisecond,
	})

	require.Len(t, results, 1)
	assert.Equal(t, channels.StatusBlocked, results[0].Status)
}

func TestAggregate(t *testing.T) {
	t.Parallel()

	results := []channels.Result{
		{Status: channels.StatusPassed},
		{Status: channels.StatusPassed},
		{Status: channels.StatusBlocked},
		{Status: channels.StatusPartial},
		{Status: channels.StatusError},
		{Status: channels.StatusSkipped},
	}

	s := engine.Aggregate(results)
	assert.Equal(t, 2, s.Passed)
	assert.Equal(t, 1, s.Blocked)
	assert.Equal(t, 1, s.Partial)
	assert.Equal(t, 1, s.Errors)
	assert.Equal(t, 1, s.Skipped)
	assert.Equal(t, 6, s.Total)
}
