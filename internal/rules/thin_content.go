package rules

import (
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// thinContentWords is the minimum visible word count below which an indexable
// HTML page is flagged as potentially too thin to provide value.
const thinContentWords = 200

// RC-CONTENT-001: Very short or missing body text on indexable HTML pages
type ThinContentRule struct{}

func (r *ThinContentRule) ID() string   { return "RC-CONTENT-001" }
func (r *ThinContentRule) Version() int { return 2 }

func (r *ThinContentRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(snapshot) || snapshot.Noindex() {
		return nil
	}

	if snapshot.WordCount < thinContentWords {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "medium",
			Confidence:  0.8,
			URL:         snapshot.URL,
			Summary:     "Page has very little visible text content",
			Remediation: "Add meaningful, original text for the page's purpose, or noindex/consolidate it if it is a utility page.",
			Evidence: map[string]interface{}{
				"word_count":  snapshot.WordCount,
				"text_length": snapshot.TextLength,
				"threshold":   thinContentWords,
			},
		}
	}
	return nil
}
