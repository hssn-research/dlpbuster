package engine

import "github.com/hssn-research/dlpbuster/internal/channels"

// Summary aggregates pass/block/partial/error counts across results.
type Summary struct {
	Passed  int
	Blocked int
	Partial int
	Errors  int
	Skipped int
	Total   int
}

// Aggregate computes a Summary from a slice of Results.
func Aggregate(results []channels.Result) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case channels.StatusPassed:
			s.Passed++
		case channels.StatusBlocked:
			s.Blocked++
		case channels.StatusPartial:
			s.Partial++
		case channels.StatusError:
			s.Errors++
		case channels.StatusSkipped:
			s.Skipped++
		}
	}
	return s
}
