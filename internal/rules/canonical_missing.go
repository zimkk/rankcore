package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RC-CANON-001: Missing canonical on pages that should have one (Lighthouse: canonical)
type MissingCanonicalRule struct{}

func (r *MissingCanonicalRule) ID() string  { return "RC-CANON-001" }
func (r *MissingCanonicalRule) Version() int { return 1 }

func (r *MissingCanonicalRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	// Only flag HTML pages — non-HTML resources don't need canonicals.
	if !strings.Contains(snapshot.ContentType, "text/html") {
		return nil
	}
	if strings.TrimSpace(snapshot.Canonical) == "" {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "canonical",
			Severity:    "medium",
			Confidence:  0.8,
			URL:         snapshot.URL,
			Summary:     "Page is missing a canonical URL",
			Remediation: "Add a <link rel=\"canonical\" href=\"...\"> pointing to the preferred version of this page.",
			Automation:  "review_required",
			Verify:      &report.VerifyInfo{Type: "canonical_present", Expected: "non_empty"},
		}
	}
	return nil
}
