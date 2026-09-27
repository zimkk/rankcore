package verify

import (
	"testing"

	"github.com/zimkk/rankcore/internal/report"
)

func TestDiff(t *testing.T) {
	baseline := &report.AuditReport{
		RankCoreVersion: "1.0.0",
		Target:          "https://example.com",
		Coverage: report.Coverage{
			PagesAnalyzed: 5,
		},
		Findings: []report.Finding{
			{
				ID:       "RC-META-001",
				URL:      "https://example.com/about",
				Severity: "medium",
			},
			{
				ID:       "RC-META-002",
				URL:      "https://example.com/contact",
				Severity: "medium",
			},
			{
				ID:       "RC-HTTP-002",
				URL:      "https://example.com/broken",
				Severity: "medium",
			},
			{
				ID:       "RC-IMG-001",
				URL:      "https://example.com/team",
				Severity: "low",
			},
		},
	}

	newRun := &report.AuditReport{
		RankCoreVersion: "1.0.0",
		Target:          "https://example.com",
		Coverage: report.Coverage{
			PagesAnalyzed: 5,
		},
		Findings: []report.Finding{
			// RC-META-001 is fixed (absent)
			// RC-META-002 is still present (same severity)
			{
				ID:       "RC-META-002",
				URL:      "https://example.com/contact",
				Severity: "medium",
			},
			// RC-IMG-001 regressed (low -> high)
			{
				ID:       "RC-IMG-001",
				URL:      "https://example.com/team",
				Severity: "high",
			},
			// New finding
			{
				ID:       "RC-LINK-002",
				URL:      "https://example.com/new-page",
				Severity: "medium",
			},
		},
	}

	vr := Diff(baseline, newRun)

	expectedStatuses := map[string]string{
		"https://example.com/about|RC-META-001":    "fixed",
		"https://example.com/contact|RC-META-002":  "still_present",
		"https://example.com/team|RC-IMG-001":      "regressed",
		"https://example.com/broken|RC-HTTP-002":   "fixed",
		"https://example.com/new-page|RC-LINK-002": "new",
	}

	if len(vr.Results) != len(expectedStatuses) {
		t.Fatalf("expected %d results, got %d", len(expectedStatuses), len(vr.Results))
	}

	for _, res := range vr.Results {
		key := res.URL + "|" + res.FindingID
		expected, ok := expectedStatuses[key]
		if !ok {
			t.Errorf("unexpected finding key in results: %s", key)
			continue
		}
		if res.Status != expected {
			t.Errorf("for %s expected status %q, got %q", key, expected, res.Status)
		}
	}
}

func TestDiffNotRetestable(t *testing.T) {
	baseline := &report.AuditReport{
		Coverage: report.Coverage{
			PagesAnalyzed: 10,
		},
		Findings: []report.Finding{
			{
				ID:       "RC-META-001",
				URL:      "https://example.com/unreachable",
				Severity: "high",
			},
		},
	}

	// New run had crawl failure or aborted early (pages analyzed dropped from 10 to 2)
	newRun := &report.AuditReport{
		Coverage: report.Coverage{
			PagesAnalyzed: 2,
		},
		Findings: []report.Finding{},
	}

	vr := Diff(baseline, newRun)
	if len(vr.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(vr.Results))
	}
	if vr.Results[0].Status != "not_retestable" {
		t.Errorf("expected status 'not_retestable', got %q", vr.Results[0].Status)
	}
}
