package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/verify"
)

func TestEndToEndAuditAndVerification(t *testing.T) {
	// Serve static test fixtures from testdata/site
	testdataDir, err := filepath.Abs("../testdata/site")
	if err != nil {
		t.Fatalf("failed to resolve testdata dir: %v", err)
	}

	handler := http.FileServer(http.Dir(testdataDir))
	server := httptest.NewServer(handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tmpOutDir, err := os.MkdirTemp("", "rankcore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp outDir: %v", err)
	}
	defer os.RemoveAll(tmpOutDir)

	opts := audit.DefaultOptions()
	opts.Target = server.URL + "/index.html"
	opts.OutDir = tmpOutDir
	opts.RenderMode = "off" // Use pure HTTP mode for fast deterministic integration test
	opts.MaxPages = 20

	rep, err := audit.Run(ctx, opts)
	if err != nil {
		t.Fatalf("audit.Run failed: %v", err)
	}

	if rep.Target != opts.Target {
		t.Errorf("expected target %s, got %s", opts.Target, rep.Target)
	}

	if rep.Coverage.PagesAnalyzed < 4 {
		t.Errorf("expected at least 4 pages analyzed, got %d", rep.Coverage.PagesAnalyzed)
	}

	// Verify expected rule findings were detected
	findingsByID := make(map[string]bool)
	for _, f := range rep.Findings {
		findingsByID[f.ID] = true
	}

	expectedRules := []string{
		"RC-META-001",    // Missing title on broken.html
		"RC-IMG-001",     // Image missing alt on broken.html
		"RC-CONTENT-001", // Thin content on thin.html
		"RC-INDEX-001",   // Meta noindex on noindex.html
		"RC-HTTP-002",    // 404 client error on does-not-exist-404.html
	}

	for _, ruleID := range expectedRules {
		if !findingsByID[ruleID] {
			t.Errorf("expected finding %s to be triggered in test audit, but was not found", ruleID)
		}
	}

	// Verify report saving
	if err := audit.SaveReport(rep, tmpOutDir, "both"); err != nil {
		t.Fatalf("failed to save audit report: %v", err)
	}

	jsonPath := filepath.Join(tmpOutDir, "audit.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("expected audit.json to exist at %s", jsonPath)
	}

	mdPath := filepath.Join(tmpOutDir, "audit.md")
	if _, err := os.Stat(mdPath); err != nil {
		t.Errorf("expected audit.md to exist at %s", mdPath)
	}

	// Test verification diffing
	// Compare baseline with a run where one finding was fixed
	modifiedRun := *rep
	if len(rep.Findings) > 0 {
		modifiedRun.Findings = rep.Findings[1:] // first finding resolved
	}

	vr := verify.Diff(rep, &modifiedRun)
	if len(vr.Results) == 0 {
		t.Errorf("expected verification results")
	}

	fixedCount := 0
	for _, res := range vr.Results {
		if res.Status == "fixed" {
			fixedCount++
		}
	}
	if fixedCount == 0 {
		t.Errorf("expected at least 1 fixed finding in verification diff")
	}
}
