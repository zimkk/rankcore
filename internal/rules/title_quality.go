package rules

import (
	"fmt"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// TitleQualityRule checks whether the title tag is excessively long, too short, or uses generic placeholder text.
type TitleQualityRule struct{}

func (r *TitleQualityRule) ID() string {
	return "RC-META-004"
}

func (r *TitleQualityRule) Version() int {
	return 1
}

var genericTitles = map[string]bool{
	"home":       true,
	"homepage":   true,
	"untitled":   true,
	"welcome":    true,
	"index":      true,
	"page":       true,
	"document":   true,
}

func (r *TitleQualityRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 || snapshot.Noindex() {
		return nil
	}

	title := strings.TrimSpace(snapshot.Title)
	if title == "" {
		// Handled by MissingTitleRule RC-META-001
		return nil
	}

	lower := strings.ToLower(title)
	if genericTitles[lower] {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "metadata",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Title tag uses non-descriptive generic text: %q", title),
			Evidence: map[string]interface{}{
				"title":  title,
				"reason": "generic_placeholder",
			},
			Remediation: "Replace the generic placeholder title with a descriptive, keyword-rich title reflecting page content.",
		}
	}

	if len(title) > 70 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "metadata",
			Severity:    "low",
			Confidence:  0.8,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Title tag is too long (%d characters, recommended <= 60-70)", len(title)),
			Evidence: map[string]interface{}{
				"title":  title,
				"length": len(title),
				"max":    70,
			},
			Remediation: "Keep title tags concise (between 30 and 60 characters) to avoid truncation in Google SERPs.",
		}
	}

	if len(title) < 10 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "metadata",
			Severity:    "medium",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Title tag is very short (%d characters)", len(title)),
			Evidence: map[string]interface{}{
				"title":  title,
				"length": len(title),
				"min":    10,
			},
			Remediation: "Expand title with brand name and key topic phrasing (aim for 30-60 characters).",
		}
	}

	return nil
}
