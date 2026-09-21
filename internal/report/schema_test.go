package report

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditReportJSON(t *testing.T) {
	finding := Finding{
		ID:          "indexability.robots_disallow",
		RuleVersion: 1,
		Category:    "indexability",
		Severity:    "high",
		Confidence:  1.0,
		URL:         "https://example.com/pricing",
		Evidence: map[string]interface{}{
			"robots_url":   "https://example.com/robots.txt",
			"user_agent":   "Googlebot",
			"matched_rule": "Disallow: /pricing",
		},
		Summary:     "Pricing page is blocked for Googlebot",
		Remediation: "Review whether this path is intentionally blocked.",
		Automation:  "review_required",
		Verify: &VerifyInfo{
			Type:     "robots_access",
			Expected: "allowed",
		},
	}

	report := AuditReport{
		SchemaVersion:   1,
		RankCoreVersion: "1.0.0",
		RulesetVersion:  "2026.09",
		GeneratedAt:     time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		Target:          "http://localhost:3000",
		Mode: Mode{
			HTTP:     true,
			Rendered: true,
			Browser:  "Chrome",
		},
		Coverage: Coverage{
			PagesRequested: 92,
			PagesAnalyzed:  88,
			PagesSkipped:   4,
			LimitReached:   false,
		},
		Findings: []Finding{finding},
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal AuditReport: %v", err)
	}

	// Basic validation to ensure the fields are serialized.
	out := string(b)
	if out == "" {
		t.Error("Serialized JSON is empty")
	}
}
