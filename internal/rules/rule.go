package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
)

// Rule represents a deterministic SEO check.
type Rule interface {
	// ID returns the stable identifier of the rule (e.g. "indexability.robots_disallow").
	ID() string
	
	// Version returns the rule version for schema tracking.
	Version() int
	
	// Evaluate processes a page snapshot and optionally returns a finding.
	Evaluate(snapshot *crawl.PageSnapshot) *report.Finding
}
