package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
	"strings"
)

type BrokenLinkRule struct{}

func (r *BrokenLinkRule) ID() string {
	return "RC-LINK-001"
}

func (r *BrokenLinkRule) Version() int {
	return 1
}

func (r *BrokenLinkRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	// A true broken link rule needs to know the graph state of the crawl.
	// We'll flag obvious internal formatting errors as a proxy.
	
	for _, link := range snapshot.InternalLinks {
		if strings.Contains(link, "undefined") || strings.Contains(link, "null") {
			return &report.Finding{
				ID:          r.ID(),
				RuleVersion: r.Version(),
				Category:    "discovery",
				Severity:    "high",
				Confidence:  1.0,
				URL:         snapshot.URL,
				Summary:     "Internal link is clearly malformed",
				Evidence: map[string]interface{}{
					"malformed_link": link,
				},
			}
		}
	}
	
	return nil
}
