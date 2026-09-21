package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
)

type CanonicalTargetRule struct{}

func (r *CanonicalTargetRule) ID() string {
	return "RC-CANON-002" // Canonical target returns non-success/redirect
}

func (r *CanonicalTargetRule) Version() int {
	return 1
}

func (r *CanonicalTargetRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.Canonical == "" {
		return nil
	}

	// This is a naive check. A full implementation would fetch the Canonical URL
	// and verify its status code. Since this evaluate function works per-snapshot,
	// checking cross-url integrity usually requires a post-crawl verification step.
	// For end-to-end completeness of the rule structure, we assume we flag if canonical differs
	// significantly or is an obvious error state.
	
	// Stub check: just log if canonical is not the final URL and leave it for post-crawl.
	if snapshot.Canonical != snapshot.FinalURL {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "canonical",
			Severity:    "medium",
			Confidence:  0.5,
			URL:         snapshot.URL,
			Summary:     "Canonical target differs from final URL. Cross-check required.",
			Evidence: map[string]interface{}{
				"canonical": snapshot.Canonical,
				"final_url": snapshot.FinalURL,
			},
		}
	}

	return nil
}
