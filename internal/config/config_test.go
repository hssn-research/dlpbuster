package config_test

import (
	"testing"

	"github.com/hssn-research/dlpbuster/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Parallel()

	// No config file present → pure defaults
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, config.DefaultTimeout, cfg.Timeout)
	assert.Equal(t, 1024, cfg.Payload.SizeBytes)
	assert.Equal(t, "human", cfg.Output.Format)
}
