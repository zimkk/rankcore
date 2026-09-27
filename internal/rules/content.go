package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RC-META-001: Page missing <title>
type MissingTitleRule struct{}

func (r *MissingTitleRule) ID() string  { return "RC-META-001" }
func (r *MissingTitleRule) Version() int { return 1 }

func (r *MissingTitleRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	if strings.TrimSpace(snapshot.Title) == "" {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Page is missing a <title> element",
			Remediation: "Add a descriptive <title> tag within <head>.",
			Automation:  "safe",
			Verify:      &report.VerifyInfo{Type: "title_present", Expected: "non_empty"},
		}
	}
	return nil
}

// RC-META-002: Page missing <meta name="description">
type MissingMetaDescRule struct{}

func (r *MissingMetaDescRule) ID() string  { return "RC-META-002" }
func (r *MissingMetaDescRule) Version() int { return 1 }

func (r *MissingMetaDescRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	if strings.TrimSpace(snapshot.MetaDesc) == "" {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "medium",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Page is missing a meta description",
			Remediation: "Add a <meta name=\"description\" content=\"...\"> tag within <head>.",
			Automation:  "review_required",
			Verify:      &report.VerifyInfo{Type: "meta_desc_present", Expected: "non_empty"},
		}
	}
	return nil
}

// RC-META-003: Missing or invalid viewport meta tag
type MissingViewportRule struct{}

func (r *MissingViewportRule) ID() string  { return "RC-META-003" }
func (r *MissingViewportRule) Version() int { return 1 }

func (r *MissingViewportRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	vp := strings.TrimSpace(snapshot.Viewport)
	if vp == "" {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Page is missing a viewport meta tag",
			Remediation: "Add <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"> within <head>.",
			Automation:  "safe",
			Verify:      &report.VerifyInfo{Type: "viewport_present", Expected: "non_empty"},
		}
	}
	// Check that it contains width= at minimum
	if !strings.Contains(vp, "width=") {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "content",
			Severity:    "medium",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     "Viewport meta tag does not set width",
			Remediation: "Ensure the viewport meta tag includes 'width=device-width'.",
			Evidence: map[string]interface{}{
				"viewport": vp,
			},
		}
	}
	return nil
}
