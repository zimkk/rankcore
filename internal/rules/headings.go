package rules

import (
	"fmt"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// H1StructureRule checks for missing or multiple <h1> headings on an indexable page.
type H1StructureRule struct{}

func (r *H1StructureRule) ID() string {
	return "RC-META-005"
}

func (r *H1StructureRule) Version() int {
	return 1
}

func (r *H1StructureRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 || snapshot.Noindex() {
		return nil
	}

	h1s := snapshot.H1
	if len(h1s) == 0 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "medium",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Document does not have an <h1> heading",
			Evidence: map[string]interface{}{
				"h1_count": 0,
			},
			Remediation: "Add a single, descriptive <h1> element representing the main topic of the page.",
		}
	}

	if len(h1s) > 1 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "low",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Document has multiple (count: %d) <h1> headings", len(h1s)),
			Evidence: map[string]interface{}{
				"h1_count": len(h1s),
				"headings": h1s,
			},
			Remediation: "Use exactly one primary <h1> per page to clearly convey hierarchy to search engines and assistive technology.",
		}
	}

	return nil
}
