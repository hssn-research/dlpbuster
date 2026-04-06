// Package engine runs exfil channels concurrently and collects results.
package engine

import (
	"context"
	"sync"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// RunConfig controls the engine's execution behaviour.
type RunConfig struct {
	Channels    []channels.Channel
	ChannelCfg  channels.ChannelConfig
	Timeout     time.Duration
	Concurrency int
}

// Run executes all channels concurrently, respecting per-channel timeout.
// Results are returned in the same order as RunConfig.Channels.
func Run(ctx context.Context, cfg RunConfig) []channels.Result {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = len(cfg.Channels)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	sem := make(chan struct{}, cfg.Concurrency)
	results := make([]channels.Result, len(cfg.Channels))

	var wg sync.WaitGroup
	for i, ch := range cfg.Channels {
		wg.Add(1)
		go func(idx int, c channels.Channel) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			chCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()

			start := time.Now()
			r := c.Run(chCtx, cfg.ChannelCfg)
			if r.Duration == 0 {
				r.Duration = time.Since(start)
			}
			r.Channel = c.Name()
			results[idx] = r
		}(i, ch)
	}
	wg.Wait()
	return results
}
