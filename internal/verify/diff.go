package verify

import (
	"rankcore/internal/report"
)

// Diff compares a baseline report against a new run and categorizes findings.
func Diff(baseline, newRun *report.AuditReport) *report.VerificationReport {
	vr := &report.VerificationReport{
		SchemaVersion:   1,
		RankCoreVersion: newRun.RankCoreVersion,
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

		_, stillPresent := newFindingsMap[key]
		if stillPresent {
			vr.Results = append(vr.Results, report.VerifyResult{
				FindingID: f.ID,
				URL:       f.URL,
				Status:    "still_present",
			})
		} else {
			vr.Results = append(vr.Results, report.VerifyResult{
				FindingID: f.ID,
				URL:       f.URL,
				Status:    "fixed",
			})
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
