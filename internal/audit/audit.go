package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zimkk/rankcore/internal/aiaccess"
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/graph"
	"github.com/zimkk/rankcore/internal/httpx"
	"github.com/zimkk/rankcore/internal/render"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/robots"
	"github.com/zimkk/rankcore/internal/rules"
)

// Version constants
const (
	RankCoreVersion = "1.0.0"
	RulesetVersion  = "2026.09"
	SchemaVersion   = 1
)

// Options configures an audit run.
type Options struct {
	Target      string
	RenderMode  string
	Browser     string
	MaxPages    int
	Concurrency int
	Include     string
	Exclude     string
	OutDir      string
	Format      string
	UserAgent   string
	Timeout     time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		RenderMode:  "off",
		MaxPages:    500,
		Concurrency: 5,
		OutDir:      ".rankcore/runs/latest",
		Format:      "json",
		UserAgent:   "RankCoreBot/1.0 (+https://github.com/zimkk/rankcore)",
		Timeout:     20 * time.Second,
	}
}

// Run executes a full deterministic audit against the target.
func Run(ctx context.Context, opts Options) (*report.AuditReport, error) {
	parsedTarget, err := url.Parse(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	allowPrivate := parsedTarget.Hostname() == "localhost" || parsedTarget.Hostname() == "127.0.0.1"

	// 1. Fetch and parse robots.txt
	robotsTxt := fetchRobotsTxt(ctx, parsedTarget, opts.UserAgent, opts.Timeout, allowPrivate)

	// 2. Evaluate AI crawler access
	crawlerAccess := evaluateCrawlerAccess(robotsTxt)

	// 3. Configure and run the crawler
	cfg := crawl.Config{
		RenderMode:     crawl.RenderMode(opts.RenderMode),
		MaxPages:       opts.MaxPages,
		MaxDepth:       10,
		Concurrency:    opts.Concurrency,
		IncludePattern: opts.Include,
		ExcludePattern: opts.Exclude,
		OutDir:         opts.OutDir,
		OutputFormat:   opts.Format,
		UserAgent:      opts.UserAgent,
		Timeout:        opts.Timeout,
	}

	renderer := render.NewRenderer(opts.Browser)
	if strings.EqualFold(opts.RenderMode, "required") && !renderer.Available {
		return nil, render.ErrRenderNotSupported
	}
	shouldRender := renderer.Available && !strings.EqualFold(opts.RenderMode, "off")

	crawler := crawl.NewCrawler(cfg)

	linkGraph := graph.New()
	registry := rules.NewRegistry()

	var snapshots []*crawl.PageSnapshot
	var allFindings []report.Finding

	// Start crawling
	go crawler.Start(ctx, opts.Target)

	// Collect results
	for snap := range crawler.Results {
		if snap == nil {
			continue
		}

		// Mark blocked-by-robots before rule evaluation
		if robotsTxt != nil && !allowPrivate {
			p := parsedTarget.Path
			if snapParsed, err := url.Parse(snap.URL); err == nil {
				p = snapParsed.Path
			}
			if robotsTxt.IsDisallowed("Googlebot", p) {
				snap.BlockedByRobots = true
				snap.RobotsRule = robotsTxt.MatchedRule("Googlebot", p)
			}
		}

		// Render DOM if browser is available and enabled
		if shouldRender && snap.StatusCode == 200 && len(snap.RedirectChain) == 0 && len(snapshots) < 15 {
			if rSnap, err := renderer.Render(ctx, snap.URL); err == nil {
				snap.Rendered = rSnap
			}
		}

		snapshots = append(snapshots, snap)

		// Build link graph
		linkGraph.AddPage(snap.URL, snap.InternalLinks)

		// Evaluate rules
		findings := registry.EvaluateAll(snap)
		allFindings = append(allFindings, findings...)
	}

	// Post-crawl: check for orphan pages
	orphans := linkGraph.Orphans(opts.Target)
	for _, orphanURL := range orphans {
		allFindings = append(allFindings, report.Finding{
			ID:          "RC-GRAPH-001",
			RuleVersion: 1,
			Category:    "discovery",
			Severity:    "medium",
			Confidence:  0.8,
			URL:         orphanURL,
			Summary:     "Page has no inbound internal links from other crawled pages",
			Remediation: "Add internal links from related pages to improve discoverability.",
			Evidence: map[string]interface{}{
				"inbound_count": 0,
			},
		})
	}

	// Post-crawl: check for excessive depth
	depths := linkGraph.Depth(opts.Target)
	for pageURL, depth := range depths {
		if depth > 5 {
			allFindings = append(allFindings, report.Finding{
				ID:          "RC-GRAPH-002",
				RuleVersion: 1,
				Category:    "discovery",
				Severity:    "low",
				Confidence:  0.7,
				URL:         pageURL,
				Summary:     "Page is excessively deep in the site architecture",
				Remediation: "Consider adding navigation shortcuts or restructuring to reduce click depth.",
				Evidence: map[string]interface{}{
					"depth": depth,
				},
			})
		}
	}

	// Build the audit report
	pagesAnalyzed := 0
	for _, snap := range snapshots {
		if snap.StatusCode > 0 {
			pagesAnalyzed++
		}
	}

	auditReport := &report.AuditReport{
		SchemaVersion:   SchemaVersion,
		RankCoreVersion: RankCoreVersion,
		RulesetVersion:  RulesetVersion,
		GeneratedAt:     time.Now().UTC(),
		Target:          opts.Target,
		Mode: report.Mode{
			HTTP:     true,
			Rendered: shouldRender,
			Browser:  renderer.BrowserPath,
		},
		Coverage: report.Coverage{
			PagesRequested: len(snapshots),
			PagesAnalyzed:  pagesAnalyzed,
			PagesSkipped:   len(snapshots) - pagesAnalyzed,
			LimitReached:   len(snapshots) >= opts.MaxPages,
		},
		Findings:      allFindings,
		CrawlerAccess: crawlerAccess,
		Limitations:   buildLimitations(opts, robotsTxt),
	}

	return auditReport, nil
}

// SaveReport writes the audit report to disk.
func SaveReport(auditReport *report.AuditReport, outDir, format string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(auditReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	jsonPath := filepath.Join(outDir, "audit.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Also write a human-readable summary
	if format == "md" || format == "both" || format == "all" || format == "text" || format == "json" || format == "" {
		summary := buildMarkdownSummary(auditReport)
		mdPath := filepath.Join(outDir, "audit.md")
		if err := os.WriteFile(mdPath, []byte(summary), 0644); err != nil {
			return fmt.Errorf("failed to write summary: %w", err)
		}
	}

	return nil
}

func fetchRobotsTxt(ctx context.Context, target *url.URL, userAgent string, timeout time.Duration, allowPrivate bool) *robots.RobotsTxt {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", target.Scheme, target.Host)
	client := httpx.NewSafeClient(timeout, userAgent, allowPrivate)

	resp, err := httpx.Request(ctx, client, robotsURL, userAgent)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	return robots.Parse(resp.Body)
}

func evaluateCrawlerAccess(rt *robots.RobotsTxt) map[string]report.Policy {
	if rt == nil {
		return map[string]report.Policy{
			"note": {Status: "no_robots_txt"},
		}
	}

	accessReport := aiaccess.Evaluate(rt, "/")
	policies := make(map[string]report.Policy)
	for _, result := range accessReport.Results {
		status := "allowed"
		if !result.Allowed {
			status = "blocked"
		}
		key := fmt.Sprintf("%s (%s)", result.Crawler.Name, result.Crawler.Purpose)
		policies[key] = report.Policy{Status: status}
	}
	return policies
}

func buildLimitations(opts Options, rt *robots.RobotsTxt) []string {
	var limitations []string
	if strings.EqualFold(opts.RenderMode, "off") {
		limitations = append(limitations, "Rendered DOM was not evaluated (render mode: off). JS-only content may be missed.")
	}
	if rt == nil {
		limitations = append(limitations, "Could not fetch robots.txt. Crawler access policies may be incomplete.")
	}
	return limitations
}

func buildMarkdownSummary(r *report.AuditReport) string {
	var b strings.Builder
	b.WriteString("# RankCore Audit Report\n\n")
	b.WriteString(fmt.Sprintf("**Target:** %s\n", r.Target))
	b.WriteString(fmt.Sprintf("**Generated:** %s\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("**RankCore:** %s | **Ruleset:** %s\n\n", r.RankCoreVersion, r.RulesetVersion))

	b.WriteString("## Coverage\n\n")
	b.WriteString(fmt.Sprintf("- Pages requested: %d\n", r.Coverage.PagesRequested))
	b.WriteString(fmt.Sprintf("- Pages analyzed: %d\n", r.Coverage.PagesAnalyzed))
	b.WriteString(fmt.Sprintf("- Pages skipped: %d\n", r.Coverage.PagesSkipped))
	b.WriteString(fmt.Sprintf("- Limit reached: %v\n\n", r.Coverage.LimitReached))

	// Count by severity
	counts := map[string]int{}
	for _, f := range r.Findings {
		counts[f.Severity]++
	}
	b.WriteString("## Findings Summary\n\n")
	b.WriteString(fmt.Sprintf("| Severity | Count |\n|----------|-------|\n"))
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		if c, ok := counts[sev]; ok {
			b.WriteString(fmt.Sprintf("| %s | %d |\n", sev, c))
		}
	}
	b.WriteString(fmt.Sprintf("\n**Total findings:** %d\n\n", len(r.Findings)))

	// Findings detail
	if len(r.Findings) > 0 {
		b.WriteString("## Findings\n\n")
		for _, f := range r.Findings {
			b.WriteString(fmt.Sprintf("### %s — %s\n\n", f.ID, f.Summary))
			b.WriteString(fmt.Sprintf("- **URL:** %s\n", f.URL))
			b.WriteString(fmt.Sprintf("- **Severity:** %s\n", f.Severity))
			b.WriteString(fmt.Sprintf("- **Category:** %s\n", f.Category))
			if f.Remediation != "" {
				b.WriteString(fmt.Sprintf("- **Fix:** %s\n", f.Remediation))
			}
			b.WriteString("\n")
		}
	}

	// Crawler access
	if len(r.CrawlerAccess) > 0 {
		b.WriteString("## Crawler Access\n\n")
		for name, policy := range r.CrawlerAccess {
			b.WriteString(fmt.Sprintf("- **%s:** %s\n", name, policy.Status))
		}
		b.WriteString("\n")
	}

	// Limitations
	if len(r.Limitations) > 0 {
		b.WriteString("## Limitations\n\n")
		for _, l := range r.Limitations {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
	}

	return b.String()
}
