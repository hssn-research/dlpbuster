package report

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// jsonResult is the JSON representation of a single channel result.
type jsonResult struct {
	Channel    string   `json:"channel"`
	Status     string   `json:"status"`
	BytesSent  int      `json:"bytes_sent"`
	DurationMs int64    `json:"duration_ms"`
	Evidence   []string `json:"evidence"`
	Error      string   `json:"error,omitempty"`
}

// jsonReport is the top-level JSON output structure.
type jsonReport struct {
	Version   string       `json:"version"`
	RunAt     time.Time    `json:"run_at"`
	Target    string       `json:"target,omitempty"`
	PayloadSz int          `json:"payload_bytes"`
	Results   []jsonResult `json:"results"`
	Summary   struct {
		Passed  int `json:"passed"`
		Blocked int `json:"blocked"`
		Partial int `json:"partial"`
		Errors  int `json:"errors"`
		Skipped int `json:"skipped"`
		Total   int `json:"total"`
	} `json:"summary"`
}

// JSONRenderer renders a Report as a single JSON object.
type JSONRenderer struct{}

// Render returns the report as JSON bytes.
func (j *JSONRenderer) Render(r *Report) ([]byte, error) {
	out := jsonReport{
		Version:   r.Version,
		RunAt:     r.RunAt,
		Target:    r.Target,
		PayloadSz: r.PayloadSz,
	}
	out.Summary.Passed = r.Summary.Passed
	out.Summary.Blocked = r.Summary.Blocked
	out.Summary.Partial = r.Summary.Partial
	out.Summary.Errors = r.Summary.Errors
	out.Summary.Skipped = r.Summary.Skipped
	out.Summary.Total = r.Summary.Total

	for _, res := range r.Results {
		jr := jsonResult{
			Channel:    res.Channel,
			Status:     string(res.Status),
			BytesSent:  res.BytesSent,
			DurationMs: res.Duration.Milliseconds(),
			Evidence:   res.Evidence,
		}
		if res.Error != nil {
			jr.Error = res.Error.Error()
		}
		out.Results = append(out.Results, jr)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("report: json marshal: %w", err)
	}
	return b, nil
}

// HumanRenderer renders a human-readable table to stdout (no TUI library dep).
type HumanRenderer struct{}

// Render returns the report as a human-readable byte slice.
func (h *HumanRenderer) Render(r *Report) ([]byte, error) {
	var sb []byte
	line := func(s string) { sb = append(sb, (s + "\n")...) }

	line(fmt.Sprintf("dlpbuster %s — DLP Bypass Tester", r.Version))
	line("[!] For authorized testing only. Obtain written permission before use.")
	line("")
	if r.Target != "" {
		line(fmt.Sprintf("Target: %s", r.Target))
	}
	line(fmt.Sprintf("Payload: %d bytes  |  Timeout: %s  |  Run at: %s",
		r.PayloadSz, r.Timeout, r.RunAt.Format("2006-01-02 15:04:05")))
	line("")

	for _, res := range r.Results {
		icon := statusIcon(res.Status)
		ev := ""
		if len(res.Evidence) > 0 {
			ev = res.Evidence[len(res.Evidence)-1]
		}
		line(fmt.Sprintf("  %-14s %s %-8s  %6dms   %s",
			res.Channel, icon, res.Status, res.Duration.Milliseconds(), ev))
	}

	line("")
	line(fmt.Sprintf("Summary: %d passed  |  %d blocked  |  %d partial  |  %d errors  |  %d skipped",
		r.Summary.Passed, r.Summary.Blocked, r.Summary.Partial, r.Summary.Errors, r.Summary.Skipped))

	return sb, nil
}

func statusIcon(s channels.Status) string {
	switch s {
	case channels.StatusPassed:
		return "✓"
	case channels.StatusBlocked:
		return "✗"
	case channels.StatusPartial:
		return "~"
	case channels.StatusError:
		return "!"
	default:
		return "-"
	}
}
