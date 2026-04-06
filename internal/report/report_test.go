package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleResults() []channels.Result {
	return []channels.Result{
		{Channel: "dns", Status: channels.StatusPassed, BytesSent: 512, Duration: 300 * time.Millisecond, Evidence: []string{"3/3 queries sent"}},
		{Channel: "https", Status: channels.StatusBlocked, Duration: 1200 * time.Millisecond, Evidence: []string{"connection reset"}},
		{Channel: "icmp", Status: channels.StatusSkipped, Evidence: []string{"requires root"}},
	}
}

func TestHumanRenderer(t *testing.T) {
	t.Parallel()

	r := report.New(sampleResults(), "0.1.0", "test-target", 1024, 30*time.Second)
	b, err := report.RendererFor(report.FormatHuman).Render(r)
	require.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, "dns")
	assert.Contains(t, s, "PASSED")
	assert.Contains(t, s, "BLOCKED")
}

func TestJSONRenderer(t *testing.T) {
	t.Parallel()

	r := report.New(sampleResults(), "0.1.0", "", 1024, 30*time.Second)
	b, err := report.RendererFor(report.FormatJSON).Render(r)
	require.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, `"channel"`)
	assert.Contains(t, s, `"PASSED"`)
}

func TestMarkdownRenderer(t *testing.T) {
	t.Parallel()

	r := report.New(sampleResults(), "0.1.0", "", 1024, 30*time.Second)
	b, err := report.RendererFor(report.FormatMarkdown).Render(r)
	require.NoError(t, err)
	s := string(b)
	assert.True(t, strings.HasPrefix(s, "# dlpbuster"))
}

func TestHTMLRenderer(t *testing.T) {
	t.Parallel()

	r := report.New(sampleResults(), "0.1.0", "", 1024, 30*time.Second)
	b, err := report.RendererFor(report.FormatHTML).Render(r)
	require.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, "<!DOCTYPE html>")
	assert.Contains(t, s, "PASSED")
}
