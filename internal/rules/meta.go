package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
	"strings"
)

type MetaNoindexRule struct{}

func (r *MetaNoindexRule) ID() string {
	return "RC-INDEX-001"
}

func (r *MetaNoindexRule) Version() int {
	return 1
}

func (r *MetaNoindexRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	// Look for X-Robots-Tag or meta robots
	noindex := false
	for _, vals := range snapshot.Headers {
		for _, val := range vals {
			if strings.Contains(strings.ToLower(val), "noindex") {
				noindex = true
			}
		}
	}
	
	// Assume HTML extractor populates something or we evaluate raw headers here.
	// We'd also check snapshot metadata if extracted.
	
	if noindex {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "indexability",
			Severity:    "high", // Could be intentional, requires context
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Intended public page is explicitly noindexed",
			Evidence: map[string]interface{}{
				"directive": "noindex",
			},
		}
	}
	return nil
}
