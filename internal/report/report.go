// Package report renders run results to various output formats.
package report

import (
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/engine"
)

// Report contains all data needed to render a run summary.
type Report struct {
	Version   string
	RunAt     time.Time
	Results   []channels.Result
	Summary   engine.Summary
	Target    string
	PayloadSz int
	Timeout   time.Duration
}

// Format represents a supported output format.
type Format string

const (
	FormatHuman    Format = "human"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
)

// Renderer renders a Report to bytes.
type Renderer interface {
	Render(r *Report) ([]byte, error)
}

// New creates a Report from engine results.
func New(results []channels.Result, version, target string, payloadSz int, timeout time.Duration) *Report {
	return &Report{
		Version:   version,
		RunAt:     time.Now(),
		Results:   results,
		Summary:   engine.Aggregate(results),
		Target:    target,
		PayloadSz: payloadSz,
		Timeout:   timeout,
	}
}

// RendererFor returns the appropriate Renderer for the given format.
func RendererFor(f Format) Renderer {
	switch f {
	case FormatJSON:
		return &JSONRenderer{}
	case FormatMarkdown:
		return &MarkdownRenderer{}
	case FormatHTML:
		return &HTMLRenderer{}
	default:
		return &HumanRenderer{}
	}
}
