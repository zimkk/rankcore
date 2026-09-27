package rules

import (
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
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
			Category:    "availability",
			Severity:    "critical",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Intended public page returns persistent 5xx",
			Evidence: map[string]interface{}{
				"status_code": snapshot.StatusCode,
			},
			Remediation: "Resolve backend application errors or server configuration faults to restore HTTP 200 responses.",
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
	if snapshot.StatusCode >= 400 && snapshot.StatusCode < 500 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "availability",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Internal public page returns 4xx client error unexpectedly",
			Evidence: map[string]interface{}{
				"status_code": snapshot.StatusCode,
			},
			Remediation: "Fix broken links, restore the missing resource, or set up a 301 redirect to the appropriate live destination.",
		}
	}
	return nil
}
