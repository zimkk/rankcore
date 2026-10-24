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
	return 2
}

func (r *MetaNoindexRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	xRobots := snapshot.XRobotsTag
	if xRobots == "" && snapshot.Headers != nil {
		// Hand-built snapshots may only carry raw headers.
		for k, vals := range snapshot.Headers {
			if strings.EqualFold(k, "X-Robots-Tag") {
				xRobots = strings.Join(vals, ", ")
			}
		}
	}
	probe := crawl.PageSnapshot{RobotsMeta: snapshot.RobotsMeta, XRobotsTag: xRobots}
	if !probe.Noindex() {
		return nil
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "indexability",
		Severity:    "high",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Page carries a noindex directive; confirm it is not meant to be public",
		Evidence: map[string]interface{}{
			"robots_meta":  snapshot.RobotsMeta,
			"x_robots_tag": xRobots,
		},
		Remediation: "Remove the 'noindex' directive from meta tags or HTTP response headers if the page should be indexed.",
	}
}
