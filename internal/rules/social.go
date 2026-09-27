package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// OpenGraphRule checks whether indexable pages include basic Open Graph social metadata.
type OpenGraphRule struct{}

func (r *OpenGraphRule) ID() string {
	return "RC-SOCIAL-001"
}

func (r *OpenGraphRule) Version() int {
	return 1
}

func (r *OpenGraphRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 || snapshot.Noindex() {
		return nil
	}

	og := snapshot.OpenGraph
	var missing []string

	if og == nil || strings.TrimSpace(og["og:title"]) == "" {
		missing = append(missing, "og:title")
	}
	if og == nil || strings.TrimSpace(og["og:image"]) == "" {
		missing = append(missing, "og:image")
	}
	if og == nil || strings.TrimSpace(og["og:description"]) == "" {
		missing = append(missing, "og:description")
	}

	if len(missing) > 0 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "social",
			Severity:    "low",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Document is missing standard Open Graph metadata tags",
			Evidence: map[string]interface{}{
				"missing_tags": missing,
			},
			Remediation: "Add Open Graph meta tags (<meta property=\"og:title\">, <meta property=\"og:image\">, <meta property=\"og:description\">) to optimize social card previews.",
		}
	}

	return nil
}
