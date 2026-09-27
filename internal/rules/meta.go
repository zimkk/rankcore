package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

type MetaNoindexRule struct{}

func (r *MetaNoindexRule) ID() string {
	return "RC-INDEX-001"
}

func (r *MetaNoindexRule) Version() int {
	return 1
}

func (r *MetaNoindexRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	noindex := snapshot.Noindex()

	// Also check raw Headers map if populated
	if !noindex && snapshot.Headers != nil {
		for k, vals := range snapshot.Headers {
			if strings.EqualFold(k, "X-Robots-Tag") {
				for _, v := range vals {
					lower := strings.ToLower(v)
					if strings.Contains(lower, "noindex") || strings.Contains(lower, "none") {
						noindex = true
						break
					}
				}
			}
		}
	}

	if noindex {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "indexability",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Intended public page is explicitly noindexed",
			Evidence: map[string]interface{}{
				"robots_meta":  snapshot.RobotsMeta,
				"x_robots_tag": snapshot.XRobotsTag,
			},
			Remediation: "Remove the 'noindex' directive from meta tags or HTTP response headers if the page should be indexed.",
		}
	}
	return nil
}
