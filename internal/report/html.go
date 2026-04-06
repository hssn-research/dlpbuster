package report

import (
	"fmt"
	"html"
	"strings"
	"time"
)

func execSummaryLine1(r *Report) string {
	total := len(r.Results)
	if total == 0 {
		return "No channels were tested."
	}
	return fmt.Sprintf(
		"This assessment tested %d exfiltration channel(s). %d succeeded, %d were blocked, %d partial, %d errors/skipped.",
		total, r.Summary.Passed, r.Summary.Blocked, r.Summary.Partial, r.Summary.Errors+r.Summary.Skipped)
}

func execSummaryLine2(r *Report) string {
	if r.Summary.Passed > 0 {
		return fmt.Sprintf(
			"CRITICAL: %d active exfiltration path(s) confirmed. Immediate remediation required — review egress rules and DLP policy for PASSED channels.",
			r.Summary.Passed)
	}
	return "No active exfiltration paths confirmed. DLP and egress controls effective for channels tested. Continue regular assessments."
}

// HTMLRenderer renders a Report as a self-contained HTML document.
type HTMLRenderer struct{}

// Render returns the report as HTML bytes.
func (h *HTMLRenderer) Render(r *Report) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>dlpbuster Report</title>
<style>
body{font-family:monospace;background:#0d1117;color:#c9d1d9;padding:2rem;}
h1{color:#58a6ff;}table{border-collapse:collapse;width:100%;}
th{background:#161b22;color:#8b949e;padding:.5rem 1rem;text-align:left;}
td{padding:.5rem 1rem;border-bottom:1px solid #21262d;}
.PASSED{color:#3fb950;}.BLOCKED{color:#f85149;}.PARTIAL{color:#d29922;}
.ERROR{color:#f0883e;}.SKIPPED{color:#8b949e;}
.summary{margin-top:1rem;font-size:1.1rem;}
</style>
</head>
<body>
`)

	sb.WriteString(fmt.Sprintf("<h1>dlpbuster %s — DLP Bypass Tester Report</h1>\n", html.EscapeString(r.Version)))
	sb.WriteString("<p><strong>⚠️ For authorized testing only.</strong></p>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Run at:</strong> %s</p>\n", r.RunAt.Format(time.RFC1123)))
	if r.Target != "" {
		sb.WriteString(fmt.Sprintf("<p><strong>Target:</strong> <code>%s</code></p>\n", html.EscapeString(r.Target)))
	}
	sb.WriteString(fmt.Sprintf("<p><strong>Payload:</strong> %d bytes | <strong>Timeout:</strong> %s</p>\n",
		r.PayloadSz, r.Timeout))

	sb.WriteString("<h2>Results</h2>\n<table>\n")
	sb.WriteString("<tr><th>Channel</th><th>Status</th><th>Duration</th><th>Bytes Sent</th><th>Evidence</th></tr>\n")

	for _, res := range r.Results {
		ev := ""
		if len(res.Evidence) > 0 {
			ev = res.Evidence[len(res.Evidence)-1]
		}
		sb.WriteString(fmt.Sprintf(`<tr><td><code>%s</code></td><td class="%s">%s %s</td><td>%dms</td><td>%d</td><td>%s</td></tr>`+"\n",
			html.EscapeString(res.Channel),
			html.EscapeString(string(res.Status)),
			statusIcon(res.Status),
			html.EscapeString(string(res.Status)),
			res.Duration.Milliseconds(),
			res.BytesSent,
			html.EscapeString(ev),
		))
	}
	sb.WriteString("</table>\n")

	sb.WriteString(fmt.Sprintf(`<div class="summary">
<strong>Summary:</strong>
<span class="PASSED">%d passed</span> |
<span class="BLOCKED">%d blocked</span> |
<span class="PARTIAL">%d partial</span> |
<span class="ERROR">%d errors</span> |
<span class="SKIPPED">%d skipped</span>
</div>
`, r.Summary.Passed, r.Summary.Blocked, r.Summary.Partial, r.Summary.Errors, r.Summary.Skipped))

	// Executive summary
	sb.WriteString("<h2>Executive Summary</h2>\n")
	sb.WriteString(fmt.Sprintf("<p>%s</p>\n", html.EscapeString(execSummaryLine1(r))))
	sb.WriteString(fmt.Sprintf("<p>%s</p>\n", html.EscapeString(execSummaryLine2(r))))

	// Risk table
	sb.WriteString("<h2>Risk Assessment</h2>\n<table>\n")
	sb.WriteString("<tr><th>Channel</th><th>Finding</th><th>Severity</th><th>Recommendation</th></tr>\n")
	for _, res := range r.Results {
		finding, severity, rec := riskForResult(res.Status)
		sb.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			html.EscapeString(res.Channel),
			html.EscapeString(finding),
			html.EscapeString(severity),
			html.EscapeString(rec),
		))
	}
	sb.WriteString("</table>\n")

	sb.WriteString("</body>\n</html>\n")

	return []byte(sb.String()), nil
}
