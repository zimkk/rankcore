package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/report"
)

// Options configures a verification run.
type Options struct {
	Target       string
	BaselinePath string
	OutDir       string
	Format       string
}

// Run executes a verification by re-auditing and comparing against a baseline.
func Run(ctx context.Context, opts Options) (*report.VerificationReport, error) {
	// Load baseline
	baselineData, err := os.ReadFile(opts.BaselinePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline: %w", err)
	}

	var baseline report.AuditReport
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Re-run audit
	auditOpts := audit.DefaultOptions()
	auditOpts.Target = opts.Target
	auditOpts.OutDir = opts.OutDir

	newRun, err := audit.Run(ctx, auditOpts)
	if err != nil {
		return nil, fmt.Errorf("verification audit failed: %w", err)
	}

	// Compare
	vr := Diff(&baseline, newRun)
	vr.GeneratedAt = time.Now().UTC()
	vr.Baseline = opts.BaselinePath

	return vr, nil
}

// Diff compares a baseline report against a new run and categorizes findings.
// Statuses: fixed, still_present, regressed, not_retestable, new
func Diff(baseline, newRun *report.AuditReport) *report.VerificationReport {
	vr := &report.VerificationReport{
		SchemaVersion:   1,
		RankCoreVersion: newRun.RankCoreVersion,
		GeneratedAt:     time.Now().UTC(),
		Target:          newRun.Target,
	}

	newFindingsMap := make(map[string]report.Finding)
	for _, f := range newRun.Findings {
		key := f.URL + "|" + f.ID
		newFindingsMap[key] = f
	}

	baselineFindingsMap := make(map[string]report.Finding)
	for _, f := range baseline.Findings {
		key := f.URL + "|" + f.ID
		baselineFindingsMap[key] = f

		newFinding, stillPresent := newFindingsMap[key]
		if stillPresent {
			// Check for regression: same rule fires but severity worsened
			if severityRank(newFinding.Severity) > severityRank(f.Severity) {
				vr.Results = append(vr.Results, report.VerifyResult{
					FindingID: f.ID,
					URL:       f.URL,
					Status:    "regressed",
				})
			} else {
				vr.Results = append(vr.Results, report.VerifyResult{
					FindingID: f.ID,
					URL:       f.URL,
					Status:    "still_present",
				})
			}
		} else {
			// Check if the page was even reachable in the new run
			pageReached := false
			for _, nf := range newRun.Findings {
				if nf.URL == f.URL {
					pageReached = true
					break
				}
			}
			// Also check coverage
			if !pageReached && newRun.Coverage.PagesAnalyzed < baseline.Coverage.PagesAnalyzed {
				vr.Results = append(vr.Results, report.VerifyResult{
					FindingID: f.ID,
					URL:       f.URL,
					Status:    "not_retestable",
				})
			} else {
				vr.Results = append(vr.Results, report.VerifyResult{
					FindingID: f.ID,
					URL:       f.URL,
					Status:    "fixed",
				})
			}
		}
	}

	for key, f := range newFindingsMap {
		if _, exists := baselineFindingsMap[key]; !exists {
			vr.Results = append(vr.Results, report.VerifyResult{
				FindingID: f.ID,
				URL:       f.URL,
				Status:    "new",
			})
		}
	}

	return vr
}

// SaveReport writes the verification report to disk.
func SaveReport(vr *report.VerificationReport, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(vr, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal verification report: %w", err)
	}

	path := filepath.Join(outDir, "verification.json")
	return os.WriteFile(path, data, 0644)
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
