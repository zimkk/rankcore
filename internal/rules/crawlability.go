package rules

import (
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RC-HTTP-003: Page doesn't return HTTP 200 for intended public pages (Lighthouse: http-status-code)
type HTTPSuccessRule struct{}

func (r *HTTPSuccessRule) ID() string  { return "RC-HTTP-003" }
func (r *HTTPSuccessRule) Version() int { return 1 }

func (r *HTTPSuccessRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	// Skip pages that already trigger 5xx or 4xx rules — this catches 3xx/redirect-only endpoints
	// and pages that return 0 (connection failure).
	if snapshot.StatusCode == 200 {
		return nil
	}
	// 5xx and 4xx are covered by existing rules, so only flag unusual non-200 states.
	if snapshot.StatusCode >= 400 {
		return nil
	}
	// Flag 0 (connection error) and 3xx that never resolved to a 200.
	if snapshot.StatusCode == 0 || (snapshot.StatusCode >= 300 && snapshot.StatusCode < 400) {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "eligibility",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Page does not return a successful HTTP status",
			Remediation: "Ensure the page serves a 200 OK response for direct access.",
			Evidence: map[string]interface{}{
				"status_code": snapshot.StatusCode,
				"final_url":   snapshot.FinalURL,
			},
		}
	}
	return nil
}

// RC-HTTP-004: Page blocked from Googlebot by robots.txt (Lighthouse: is-crawlable)
type RobotsCrawlableRule struct{}

func (r *RobotsCrawlableRule) ID() string  { return "RC-HTTP-004" }
func (r *RobotsCrawlableRule) Version() int { return 1 }

func (r *RobotsCrawlableRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if !snapshot.BlockedByRobots {
		return nil
	}
	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "eligibility",
		Severity:    "high",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Page is blocked from crawling by robots.txt",
		Remediation: "Review the robots.txt rules to ensure intended public pages are not disallowed.",
		Automation:  "review_required",
		Evidence: map[string]interface{}{
			"matched_rule": snapshot.RobotsRule,
		},
		Verify: &report.VerifyInfo{Type: "robots_access", Expected: "allowed"},
	}
}
