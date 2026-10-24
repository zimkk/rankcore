package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/version"
)

// Options configures a verification run.
type Options struct {
	Target       string
	BaselinePath string
	OutDir       string
	Format       string
	Audit        *audit.Options // optional overrides for the re-audit
}

// Run executes a verification by re-auditing and comparing against a baseline.
// It returns the verification report and the fresh audit.
func Run(ctx context.Context, opts Options) (*report.VerificationReport, *report.AuditReport, error) {
	baselineData, err := os.ReadFile(opts.BaselinePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read baseline: %w", err)
	}

	var baseline report.AuditReport
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return nil, nil, fmt.Errorf("failed to parse baseline: %w", err)
	}

	auditOpts := audit.DefaultOptions()
	if opts.Audit != nil {
		auditOpts = *opts.Audit
	}
	auditOpts.Target = opts.Target
	auditOpts.OutDir = opts.OutDir
	// Re-test at least as many pages as the baseline saw.
	if baseline.Coverage.PagesRequested > auditOpts.MaxPages {
		auditOpts.MaxPages = baseline.Coverage.PagesRequested
	}
	if baseline.Mode.Rendered && auditOpts.RenderMode == "off" {
		auditOpts.RenderMode = "auto"
	}

	newRun, err := audit.Run(ctx, auditOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("verification audit failed: %w", err)
	}

	vr := Diff(&baseline, newRun)
	vr.Baseline = opts.BaselinePath
	return vr, newRun, nil
}

// relURL strips a report's own origin from a URL, so runs against the same
// site on a different port (a restarted dev server, a served directory)
// still line up. URLs on other origins are kept whole.
func relURL(u, origin string) string {
	if origin != "" && strings.HasPrefix(u, origin+"/") {
		return u[len(origin):]
	}
	return u
}

func keyIn(f report.Finding, origin string) string { return relURL(f.URL, origin) + "|" + f.ID }

// Diff compares a baseline report against a new run and categorizes findings.
// Statuses: fixed, still_present, regressed, not_retestable, new
func Diff(baseline, newRun *report.AuditReport) *report.VerificationReport {
	vr := &report.VerificationReport{
		SchemaVersion:   version.SchemaVersion,
		RankCoreVersion: newRun.RankCoreVersion,
		GeneratedAt:     time.Now().UTC(),
		Target:          newRun.Target,
		Summary:         map[string]int{},
	}

	newFindings := make(map[string]report.Finding, len(newRun.Findings))
	for _, f := range newRun.Findings {
		newFindings[keyIn(f, newRun.Origin)] = f
	}

	// Which URLs did the new run actually fetch? Older reports without a page
	// list fall back to "any finding on that URL" plus coverage comparison.
	retested := map[string]bool{}
	for _, p := range newRun.Pages {
		if p.StatusCode > 0 && p.Skipped == "" {
			retested[relURL(p.URL, newRun.Origin)] = true
		}
	}
	for _, f := range newRun.Findings {
		retested[relURL(f.URL, newRun.Origin)] = true
	}
	hasPageList := len(newRun.Pages) > 0

	add := func(f report.Finding, status string) {
		vr.Results = append(vr.Results, report.VerifyResult{FindingID: f.ID, URL: f.URL, Severity: f.Severity, Status: status})
		vr.Summary[status]++
	}

	baseKeys := make(map[string]bool, len(baseline.Findings))
	for _, f := range baseline.Findings {
		k := keyIn(f, baseline.Origin)
		if baseKeys[k] {
			continue
		}
		baseKeys[k] = true

		if nf, ok := newFindings[k]; ok {
			if report.SeverityRank(nf.Severity) > report.SeverityRank(f.Severity) {
				add(nf, "regressed")
			} else {
				add(nf, "still_present")
			}
			continue
		}

		switch {
		case retested[relURL(f.URL, baseline.Origin)]:
			add(f, "fixed")
		case !hasPageList && newRun.Coverage.PagesAnalyzed >= baseline.Coverage.PagesAnalyzed:
			add(f, "fixed")
		default:
			add(f, "not_retestable")
		}
	}

	for k, f := range newFindings {
		if !baseKeys[k] {
			add(f, "new")
		}
	}

	order := map[string]int{"regressed": 0, "new": 1, "still_present": 2, "not_retestable": 3, "fixed": 4}
	sort.SliceStable(vr.Results, func(i, j int) bool {
		a, b := vr.Results[i], vr.Results[j]
		if order[a.Status] != order[b.Status] {
			return order[a.Status] < order[b.Status]
		}
		if a.FindingID != b.FindingID {
			return a.FindingID < b.FindingID
		}
		return a.URL < b.URL
	})
	return vr
}

// Gate reports whether verification should fail a CI gate: anything regressed,
// any new finding, or a critical finding still present.
func Gate(vr *report.VerificationReport) bool {
	for _, r := range vr.Results {
		switch r.Status {
		case "regressed", "new":
			return true
		case "still_present":
			if r.Severity == "critical" {
				return true
			}
		}
	}
	return false
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
