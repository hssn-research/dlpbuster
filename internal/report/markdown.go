package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// MarkdownRenderer renders a Report as a Markdown document.
type MarkdownRenderer struct{}

// Render returns the report as Markdown bytes.
func (m *MarkdownRenderer) Render(r *Report) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# dlpbuster Report — %s\n\n", r.RunAt.Format(time.RFC3339)))
	sb.WriteString("> **For authorized testing only.** Obtain written permission before use.\n\n")

	if r.Target != "" {
		sb.WriteString(fmt.Sprintf("**Target:** `%s`  \n", r.Target))
	}
	sb.WriteString(fmt.Sprintf("**Version:** %s  \n", r.Version))
	sb.WriteString(fmt.Sprintf("**Payload:** %d bytes  \n", r.PayloadSz))
	sb.WriteString(fmt.Sprintf("**Timeout:** %s  \n\n", r.Timeout))

	// Executive summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(execSummary(r))
	sb.WriteString("\n\n")

	// Results table
	sb.WriteString("## Results\n\n")
	sb.WriteString("| Channel | Status | Duration | Bytes Sent | Evidence |\n")
	sb.WriteString("|---------|--------|----------|------------|----------|\n")

	for _, res := range r.Results {
		ev := ""
		if len(res.Evidence) > 0 {
			ev = res.Evidence[len(res.Evidence)-1]
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s %s | %dms | %d | %s |\n",
			res.Channel, statusIcon(res.Status), res.Status,
			res.Duration.Milliseconds(), res.BytesSent,
			escapeMarkdown(ev)))
	}

	sb.WriteString(fmt.Sprintf("\n## Summary\n\n- **Passed:** %d\n- **Blocked:** %d\n- **Partial:** %d\n- **Errors:** %d\n- **Skipped:** %d\n",
		r.Summary.Passed, r.Summary.Blocked, r.Summary.Partial, r.Summary.Errors, r.Summary.Skipped))

	// Risk table
	sb.WriteString("\n## Risk Assessment\n\n")
	sb.WriteString(riskTable(r))

	return []byte(sb.String()), nil
}

// execSummary produces a 2-paragraph executive overview.
func execSummary(r *Report) string {
	total := len(r.Results)
	if total == 0 {
		return "No channels were tested."
	}
	var para1, para2 strings.Builder
	para1.WriteString(fmt.Sprintf(
		"This assessment tested **%d exfiltration channel(s)** against the target environment. ",
		total))
	para1.WriteString(fmt.Sprintf(
		"**%d channel(s) succeeded** in transmitting data, **%d were blocked**, "+
			"**%d were partially delivered**, and **%d encountered errors or were skipped**.",
		r.Summary.Passed, r.Summary.Blocked, r.Summary.Partial, r.Summary.Errors+r.Summary.Skipped))

	if r.Summary.Passed > 0 {
		para2.WriteString(fmt.Sprintf(
			"**Critical finding:** %d active exfiltration path(s) were confirmed. "+
				"Immediate remediation is recommended — review egress firewall rules, "+
				"DLP policy coverage, and network segmentation for the channels listed as PASSED.",
			r.Summary.Passed))
	} else {
		para2.WriteString(
			"No active exfiltration paths were confirmed during this test. " +
				"DLP and egress controls are operating as expected for the channels tested. " +
				"Continue regular assessments to maintain coverage.")
	}
	return para1.String() + "\n\n" + para2.String()
}

// riskTable produces a per-channel CVSS-adjacent risk summary.
func riskTable(r *Report) string {
	var sb strings.Builder
	sb.WriteString("| Channel | Finding | CVSS-Adjacent Severity | Recommendation |\n")
	sb.WriteString("|---------|---------|------------------------|----------------|\n")
	for _, res := range r.Results {
		finding, severity, recommendation := riskForResult(res.Status)
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			res.Channel, finding, severity, recommendation))
	}
	return sb.String()
}

func riskForResult(s channels.Status) (finding, severity, recommendation string) {
	switch s {
	case channels.StatusPassed:
		return "Data exfiltrated successfully",
			"**Critical** (CVSS ≥ 9.0)",
			"Block channel at egress firewall; update DLP rules immediately"
	case channels.StatusPartial:
		return "Partial exfiltration detected",
			"**High** (CVSS 7.0–8.9)",
			"Investigate partial bypass; tighten DLP inspection depth"
	case channels.StatusBlocked:
		return "Channel blocked by DLP/firewall",
			"**Informational**",
			"Control effective — verify rule is permanent and not bypassable"
	case channels.StatusError:
		return "Channel error (misconfiguration or connectivity)",
			"**Low**",
			"Review channel config; retest with correct credentials or targets"
	default:
		return "Skipped (not configured or requires privileges)",
			"**Informational**",
			"Configure channel and retest for complete coverage"
	}
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
