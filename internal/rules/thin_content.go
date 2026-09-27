package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// thinContentThreshold is the minimum visible text length (in bytes) below
// which a public HTML page is flagged as potentially too thin to provide value.
const thinContentThreshold = 200

// RC-CONTENT-001: Very short or missing body text on public HTML pages
type ThinContentRule struct{}

func (r *ThinContentRule) ID() string  { return "RC-CONTENT-001" }
func (r *ThinContentRule) Version() int { return 1 }

func (r *ThinContentRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	if !strings.Contains(snapshot.ContentType, "text/html") {
		return nil
	}

	if snapshot.TextLength < thinContentThreshold {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "medium",
			Confidence:  0.8,
			URL:         snapshot.URL,
			Summary:     "Page has very little visible text content",
			Remediation: "Add meaningful text content. Pages with minimal visible text may not provide enough value for search engines to index effectively.",
			Evidence: map[string]interface{}{
				"text_length": snapshot.TextLength,
				"threshold":   thinContentThreshold,
			},
		}
	}
	return nil
}
