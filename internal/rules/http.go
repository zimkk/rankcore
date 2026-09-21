package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
)

type HTTP5xxRule struct{}

func (r *HTTP5xxRule) ID() string {
	return "RC-HTTP-001"
}

func (r *HTTP5xxRule) Version() int {
	return 1
}

func (r *HTTP5xxRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode >= 500 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "eligibility",
			Severity:    "critical",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Intended public page returns persistent 5xx",
			Evidence: map[string]interface{}{
				"status_code": snapshot.StatusCode,
			},
		}
	}
	return nil
}

type HTTP4xxRule struct{}

func (r *HTTP4xxRule) ID() string {
	return "RC-HTTP-002"
}

func (r *HTTP4xxRule) Version() int {
	return 1
}

func (r *HTTP4xxRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode >= 400 && snapshot.StatusCode < 500 && snapshot.StatusCode != 404 { // 404 handled separately or contextual
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "hygiene",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Internal public page returns 4xx unexpectedly",
			Evidence: map[string]interface{}{
				"status_code": snapshot.StatusCode,
			},
		}
	}
	return nil
}
