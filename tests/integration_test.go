package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/verify"
)

// fixtureServer serves testdata/site, rewriting the fixtures' placeholder
// origin (http://127.0.0.1/) to the test server's real origin.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir, err := filepath.Abs("../testdata/site")
	if err != nil {
		t.Fatalf("failed to resolve testdata dir: %v", err)
	}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
		data, err := os.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data = bytes.ReplaceAll(data, []byte("http://127.0.0.1/"), []byte(srv.URL+"/"))
		switch filepath.Ext(name) {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".xml":
			w.Header().Set("Content-Type", "application/xml")
		case ".txt":
			w.Header().Set("Content-Type", "text/plain")
		}
		w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func runFixtureAudit(t *testing.T, srv *httptest.Server) *report.AuditReport {
	t.Helper()
	opts := audit.DefaultOptions()
	opts.Target = srv.URL + "/index.html"
	opts.OutDir = t.TempDir()
	opts.RenderMode = "off"
	opts.MaxPages = 50

	// The crawl must finish on its own; the deadline only guards against a hang.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	start := time.Now()
	rep, err := audit.Run(ctx, opts)
	if err != nil {
		t.Fatalf("audit.Run failed: %v", err)
	}
	if time.Since(start) > 10*time.Second {
		t.Fatalf("audit took %s; crawl did not terminate promptly", time.Since(start))
	}
	return rep
}

func findingsFor(rep *report.AuditReport, id string) []report.Finding {
	var out []report.Finding
	for _, f := range rep.Findings {
		if f.ID == id {
			out = append(out, f)
		}
	}
	return out
}

func hasFindingOn(rep *report.AuditReport, id, pathSuffix string) bool {
	for _, f := range findingsFor(rep, id) {
		if strings.HasSuffix(f.URL, pathSuffix) {
			return true
		}
	}
	return false
}

func TestEndToEndAuditAndVerification(t *testing.T) {
	srv := fixtureServer(t)
	rep := runFixtureAudit(t, srv)

	if rep.Coverage.PagesAnalyzed < 6 {
		t.Errorf("expected at least 6 pages analyzed, got %d", rep.Coverage.PagesAnalyzed)
	}

	expected := []struct{ id, path string }{
		{"RC-META-001", "/broken.html"},             // missing title
		{"RC-IMG-001", "/broken.html"},              // image missing alt
		{"RC-LINK-003", "/broken.html"},             // href="#" and javascript:
		{"RC-CONTENT-001", "/thin.html"},            // thin content
		{"RC-INDEX-001", "/noindex.html"},           // meta noindex
		{"RC-HTTP-002", "/does-not-exist-404.html"}, // 404
		{"RC-HTTP-002", "/products.html"},           // 404 linked from nav
		{"RC-LINK-001", "/index.html"},              // nav links to 404s
		{"RC-LINK-001", "/broken.html"},             // links to a 404
		{"RC-SITEMAP-002", "/orphan-page.html"},     // sitemap URL returns 404
		{"RC-GRAPH-001", "/orphan.html"},            // only the sitemap knows it
		{"RC-LANG-001", "/broken.html"},             // no <html lang>
	}
	for _, e := range expected {
		if !hasFindingOn(rep, e.id, e.path) {
			t.Errorf("expected %s on %s", e.id, e.path)
		}
	}

	// Pages reachable by links must not be reported as orphans.
	for _, p := range []string{"/index.html", "/about.html", "/thin.html"} {
		if hasFindingOn(rep, "RC-GRAPH-001", p) {
			t.Errorf("%s is linked internally and must not be an orphan", p)
		}
	}
	// Self-canonicals are fine.
	if len(findingsFor(rep, "RC-CANON-002")) > 0 {
		t.Errorf("unexpected canonical target findings: %+v", findingsFor(rep, "RC-CANON-002"))
	}
	// GPTBot and CCBot are blocked, but training crawlers are never an SEO issue.
	if len(findingsFor(rep, "RC-AI-001")) > 0 {
		t.Errorf("training-crawler blocks must not produce RC-AI-001")
	}
	if p, ok := rep.CrawlerAccess["GPTBot (training)"]; !ok || p.Status != "blocked" {
		t.Errorf("expected GPTBot (training) blocked, got %+v", p)
	}
	if p := rep.CrawlerAccess["Googlebot (search)"]; p.Status != "allowed" {
		t.Errorf("expected Googlebot allowed, got %+v", p)
	}
	if len(rep.Sitemaps) != 1 || rep.Sitemaps[0].Error != "" || rep.Sitemaps[0].URLCount != 4 {
		t.Errorf("expected one readable sitemap with 4 URLs, got %+v", rep.Sitemaps)
	}

	// Every finding carries catalog metadata.
	for _, f := range rep.Findings {
		if f.Category == "" || f.Automation == "" || f.Severity == "" {
			t.Errorf("finding %s on %s is missing category/automation/severity", f.ID, f.URL)
		}
	}

	// Output must be deterministic across runs.
	again := runFixtureAudit(t, srv)
	a, _ := json.Marshal(rep.Findings)
	b, _ := json.Marshal(again.Findings)
	if !bytes.Equal(a, b) {
		t.Errorf("findings differ between two identical runs")
	}

	// Report saving.
	outDir := t.TempDir()
	if err := audit.SaveReport(rep, outDir, "both"); err != nil {
		t.Fatalf("failed to save audit report: %v", err)
	}
	for _, name := range []string{"audit.json", "audit.md"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("expected %s to be written", name)
		}
	}

	// Verification: drop one finding as if it were fixed.
	modified := *rep
	modified.Findings = rep.Findings[1:]
	vr := verify.Diff(rep, &modified)
	if vr.Summary["fixed"] != 1 {
		t.Errorf("expected exactly 1 fixed finding, got %+v", vr.Summary)
	}
	// No regressions, no new findings and no critical issues: the gate stays open.
	if verify.Gate(vr) {
		t.Errorf("gate tripped without regressions, new findings or critical issues: %+v", vr.Summary)
	}
}

func TestVerifyMarksUncrawledPagesNotRetestable(t *testing.T) {
	baseline := &report.AuditReport{
		Findings: []report.Finding{{ID: "RC-META-001", URL: "https://example.com/gone", Severity: "high"}},
		Pages:    []report.PageSummary{{URL: "https://example.com/gone", StatusCode: 200}},
	}
	newRun := &report.AuditReport{
		Pages: []report.PageSummary{{URL: "https://example.com/", StatusCode: 200}},
	}
	vr := verify.Diff(baseline, newRun)
	if len(vr.Results) != 1 || vr.Results[0].Status != "not_retestable" {
		t.Fatalf("expected not_retestable, got %+v", vr.Results)
	}
}

func TestAuditLocalDirectory(t *testing.T) {
	dir, _ := filepath.Abs("../testdata/site")
	opts := audit.DefaultOptions()
	opts.Target = dir
	opts.RenderMode = "off"
	opts.MaxPages = 20

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	rep, err := audit.Run(ctx, opts)
	if err != nil {
		t.Fatalf("audit of local directory failed: %v", err)
	}
	if rep.Coverage.PagesAnalyzed == 0 {
		t.Fatalf("expected pages to be analyzed from the directory listing")
	}
}

func TestVerifyToleratesPortChange(t *testing.T) {
	baseline := &report.AuditReport{
		Origin:   "http://127.0.0.1:5001",
		Findings: []report.Finding{{ID: "RC-META-001", URL: "http://127.0.0.1:5001/a", Severity: "high"}},
		Pages:    []report.PageSummary{{URL: "http://127.0.0.1:5001/a", StatusCode: 200}},
	}
	newRun := &report.AuditReport{
		Origin:   "http://127.0.0.1:6002",
		Findings: []report.Finding{{ID: "RC-META-001", URL: "http://127.0.0.1:6002/a", Severity: "high"}},
		Pages:    []report.PageSummary{{URL: "http://127.0.0.1:6002/a", StatusCode: 200}},
	}
	vr := verify.Diff(baseline, newRun)
	if len(vr.Results) != 1 || vr.Results[0].Status != "still_present" {
		t.Fatalf("expected the same finding to match across ports, got %+v", vr.Results)
	}
}
