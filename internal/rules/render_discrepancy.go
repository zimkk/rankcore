package rules

import (
	"fmt"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RenderDiscrepancyRule detects critical differences between raw HTTP HTML and client-rendered DOM.
type RenderDiscrepancyRule struct{}

func (r *RenderDiscrepancyRule) ID() string {
	return "RC-RENDER-001"
}

func (r *RenderDiscrepancyRule) Version() int {
	return 1
}

func (r *RenderDiscrepancyRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.Rendered == nil || snapshot.StatusCode != 200 {
		return nil
	}

	rend := snapshot.Rendered

	// 1. JavaScript injected a noindex directive into the rendered DOM
	if strings.Contains(strings.ToLower(rend.RobotsMeta), "noindex") && !snapshot.Noindex() {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "renderability",
			Severity:    "critical",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Client-side JavaScript dynamically injected a 'noindex' robots meta tag into the rendered DOM",
			Evidence: map[string]interface{}{
				"raw_robots":      snapshot.RobotsMeta,
				"rendered_robots": rend.RobotsMeta,
			},
			Remediation: "Remove client-side scripts that append 'noindex' directives to public indexable pages.",
		}
	}

	// 2. Client-side JS changed the canonical URL to a different URL
	if rend.Canonical != "" && snapshot.Canonical != "" && !strings.EqualFold(rend.Canonical, snapshot.Canonical) {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "renderability",
			Severity:    "high",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     "Client-side JavaScript altered the canonical tag in the rendered DOM",
			Evidence: map[string]interface{}{
				"raw_canonical":      snapshot.Canonical,
				"rendered_canonical": rend.Canonical,
			},
			Remediation: "Ensure server-rendered HTML and client-side JavaScript agree on the authoritative canonical URL.",
		}
	}

	// 3. Raw HTML has zero body text while rendered DOM has content (pure CSR without SSR)
	if snapshot.TextLength < 30 && rend.TextLength > 150 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "renderability",
			Severity:    "medium",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Page body content is rendered purely via client-side JavaScript (raw: %d chars, rendered: %d chars)", snapshot.TextLength, rend.TextLength),
			Evidence: map[string]interface{}{
				"raw_text_length":      snapshot.TextLength,
				"rendered_text_length": rend.TextLength,
			},
			Remediation: "Consider Server-Side Rendering (SSR) or Static Site Generation (SSG) so crawlers without JavaScript execution can index page content immediately.",
		}
	}

	return nil
}
