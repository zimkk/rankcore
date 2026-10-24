package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zimkk/rankcore/internal/aiaccess"
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/httpx"
	"github.com/zimkk/rankcore/internal/render"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/robots"
	"github.com/zimkk/rankcore/internal/rules"
	"github.com/zimkk/rankcore/internal/sitemap"
	"github.com/zimkk/rankcore/internal/version"
)

// Options configures an audit run.
type Options struct {
	Target       string
	RenderMode   string
	Browser      string
	MaxPages     int
	MaxDepth     int
	MaxRender    int
	Concurrency  int
	Include      string
	Exclude      string
	OutDir       string
	Format       string
	UserAgent    string
	Timeout      time.Duration
	IgnoreRobots bool // fetch URLs even if robots.txt disallows RankCoreBot
	NoSitemap    bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		RenderMode:  "off",
		MaxPages:    500,
		MaxDepth:    10,
		MaxRender:   15,
		Concurrency: 5,
		OutDir:      ".rankcore/runs/latest",
		Format:      "json",
		UserAgent:   version.UserAgent(),
		Timeout:     20 * time.Second,
	}
}

const (
	maxSitemapFiles = 50
	maxSitemapBytes = 50 * 1024 * 1024
)

// Run executes a full deterministic audit against the target.
func Run(ctx context.Context, opts Options) (*report.AuditReport, error) {
	var limitations []string

	target, cleanup, note, err := resolveTarget(opts.Target)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	if note != "" {
		limitations = append(limitations, note)
	}

	seed, err := crawl.NormalizeURL(target, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %q: %w", opts.Target, err)
	}
	seedURL, _ := url.Parse(seed)
	local := httpx.IsLocalTarget(seedURL)
	origin := seedURL.Scheme + "://" + seedURL.Host
	client := httpx.NewSafeClient(opts.Timeout, local)

	renderer := render.NewRenderer(opts.Browser)
	if strings.EqualFold(opts.RenderMode, "required") && !renderer.Available {
		return nil, render.ErrRenderNotSupported
	}
	shouldRender := renderer.Available && !strings.EqualFold(opts.RenderMode, "off")

	// 1. robots.txt
	robotsURL := origin + "/robots.txt"
	robotsTxt, robotsNote := fetchRobotsTxt(ctx, client, robotsURL, opts.UserAgent)
	if robotsNote != "" {
		limitations = append(limitations, robotsNote)
	}

	// 2. Sitemaps
	var sitemaps []report.SitemapInfo
	sitemapURLs := map[string]string{}
	if !opts.NoSitemap {
		sitemaps, sitemapURLs = discoverSitemaps(ctx, client, origin, robotsTxt, opts.UserAgent, opts.MaxPages)
	}

	// 3. Crawl: the seed first, then every same-origin sitemap URL.
	seeds := []crawl.Seed{{URL: seed, Source: "seed"}}
	smList := make([]string, 0, len(sitemapURLs))
	for u := range sitemapURLs {
		smList = append(smList, u)
	}
	sort.Strings(smList)
	for _, u := range smList {
		seeds = append(seeds, crawl.Seed{URL: u, Source: "sitemap"})
	}

	cfg := crawl.DefaultConfig()
	cfg.RenderMode = crawl.RenderMode(opts.RenderMode)
	cfg.MaxPages = opts.MaxPages
	cfg.MaxDepth = opts.MaxDepth
	cfg.Concurrency = opts.Concurrency
	cfg.IncludePattern = opts.Include
	cfg.ExcludePattern = opts.Exclude
	cfg.UserAgent = opts.UserAgent
	cfg.Timeout = opts.Timeout
	cfg.AllowPrivate = local
	// Local dev servers belong to the user; remote sites get robots etiquette.
	cfg.RespectRobots = !local && !opts.IgnoreRobots

	crawler := crawl.NewCrawler(cfg)
	crawler.Robots = robotsTxt
	pages, err := crawler.Run(ctx, seeds)
	if err != nil {
		return nil, fmt.Errorf("crawl interrupted: %w", err)
	}
	for _, p := range pages {
		if _, ok := sitemapURLs[p.URL]; ok {
			p.InSitemap = true
		}
	}

	// 4. Optional rendering of the shallowest HTML pages.
	rendered := 0
	if shouldRender {
		byDepth := append([]*crawl.PageSnapshot(nil), pages...)
		sort.SliceStable(byDepth, func(i, j int) bool { return byDepth[i].Depth < byDepth[j].Depth })
		for _, p := range byDepth {
			if rendered >= opts.MaxRender || ctx.Err() != nil {
				break
			}
			if p.StatusCode != 200 || !p.IsHTML() || len(p.RedirectChain) > 0 {
				continue
			}
			if rs, err := renderer.Render(ctx, p.URL); err == nil {
				p.Rendered = rs
				rendered++
			}
		}
		if htmlPages := countHTML(pages); htmlPages > rendered {
			limitations = append(limitations, fmt.Sprintf("Rendered DOM was compared on %d of %d HTML pages (--max-render).", rendered, htmlPages))
		}
	}

	// 5. Rules
	site := rules.NewSite(seed, pages)
	site.Robots = robotsTxt
	site.RobotsURL = robotsURL
	site.Sitemaps = sitemaps
	site.SitemapURLs = sitemapURLs
	var access *aiaccess.Report
	if robotsTxt != nil {
		access = aiaccess.Evaluate(robotsTxt, "/")
		site.CrawlerAccess = access.Results
	}
	rules.MarkRobotsBlocked(site)
	findings := rules.NewRegistry().EvaluateSite(site)

	// 6. Report
	analyzed, skipped := 0, 0
	summaries := make([]report.PageSummary, 0, len(pages))
	for _, p := range pages {
		if p.StatusCode > 0 {
			analyzed++
		} else {
			skipped++
		}
		summaries = append(summaries, report.PageSummary{
			URL: p.URL, FinalURL: p.FinalURL, StatusCode: p.StatusCode, Depth: p.Depth, Source: p.Source,
			InSitemap: p.InSitemap, Indexable: p.Indexable(), DurationMS: p.DurationMS, Bytes: p.BodyBytes,
			Skipped: p.Skipped, Error: p.FetchError,
		})
	}

	if strings.EqualFold(opts.RenderMode, "off") {
		limitations = append(limitations, "Rendered DOM was not evaluated (render mode: off). JS-only content may be missed.")
	} else if !renderer.Available {
		limitations = append(limitations, "No Chromium-family browser found; rendered DOM was not evaluated.")
	}
	if cfg.RespectRobots && robotsTxt != nil {
		n := 0
		for _, p := range pages {
			if p.Skipped == "robots" {
				n++
			}
		}
		if n > 0 {
			limitations = append(limitations, fmt.Sprintf("%d URL(s) were not fetched because robots.txt disallows RankCoreBot (use --ignore-robots for your own site).", n))
		}
	}
	if crawler.LimitReached() {
		limitations = append(limitations, fmt.Sprintf("Crawl stopped at --max-pages=%d; some URLs were not audited.", opts.MaxPages))
	}

	return &report.AuditReport{
		SchemaVersion:   version.SchemaVersion,
		RankCoreVersion: version.Version,
		RulesetVersion:  version.RulesetVersion,
		GeneratedAt:     time.Now().UTC(),
		Target:          opts.Target,
		Origin:          origin,
		Mode: report.Mode{
			HTTP:     true,
			Rendered: shouldRender && rendered > 0,
			Browser:  renderer.BrowserPath,
		},
		Coverage: report.Coverage{
			PagesRequested: len(pages),
			PagesAnalyzed:  analyzed,
			PagesSkipped:   skipped,
			LimitReached:   crawler.LimitReached(),
		},
		Findings:      nonNil(findings),
		CrawlerAccess: crawlerAccess(access),
		Sitemaps:      sitemaps,
		Pages:         summaries,
		Limitations:   limitations,
	}, nil
}

func nonNil(f []report.Finding) []report.Finding {
	if f == nil {
		return []report.Finding{}
	}
	return f
}

func countHTML(pages []*crawl.PageSnapshot) int {
	n := 0
	for _, p := range pages {
		if p.StatusCode == 200 && p.IsHTML() {
			n++
		}
	}
	return n
}

// resolveTarget turns the user's target into a crawlable URL. A bare host gets
// https://; an existing local directory is served on a loopback port.
func resolveTarget(target string) (string, func(), string, error) {
	noop := func() {}
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		abs, _ := filepath.Abs(target)
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", noop, "", fmt.Errorf("could not serve %s: %w", abs, err)
		}
		srv := &http.Server{Handler: http.FileServer(http.Dir(abs)), ReadHeaderTimeout: 10 * time.Second}
		go srv.Serve(ln)
		return "http://" + ln.Addr().String() + "/", func() { srv.Close() },
			"Served local directory " + abs + " with a static file server; server-side redirects, headers and rewrites of the real host are not reproduced.", nil
	}
	if !strings.Contains(target, "://") {
		if strings.HasPrefix(target, "localhost") || strings.HasPrefix(target, "127.") {
			target = "http://" + target
		} else {
			target = "https://" + target
		}
	}
	u, err := url.Parse(target)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", noop, "", fmt.Errorf("invalid target %q: expected an http(s) URL or a local directory", target)
	}
	return target, noop, "", nil
}

// SaveReport writes audit.json, plus audit.md unless format is "json-only".
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

	if format != "json-only" {
		summary := buildMarkdownSummary(auditReport)
		mdPath := filepath.Join(outDir, "audit.md")
		if err := os.WriteFile(mdPath, []byte(summary), 0644); err != nil {
			return fmt.Errorf("failed to write summary: %w", err)
		}
	}

	return nil
}

func fetchRobotsTxt(ctx context.Context, client *http.Client, robotsURL, userAgent string) (*robots.RobotsTxt, string) {
	res := httpx.Fetch(ctx, client, robotsURL, userAgent, 512*1024)
	switch {
	case res.Err != nil && res.StatusCode == 0:
		return nil, "Could not fetch robots.txt (" + res.Err.Error() + "). Crawler access policies may be incomplete."
	case res.StatusCode == http.StatusOK:
		return robots.Parse(strings.NewReader(string(res.Body))), ""
	case res.StatusCode >= 500:
		// RFC 9309 §2.3.1.4: an unreachable robots.txt means "assume complete disallow".
		return robots.Parse(strings.NewReader("User-agent: *\nDisallow: /\n")),
			fmt.Sprintf("robots.txt returned %d; per RFC 9309 crawlers treat the whole site as disallowed.", res.StatusCode)
	default:
		// 4xx: no restrictions.
		return nil, fmt.Sprintf("robots.txt returned %d; crawlers treat the site as fully allowed.", res.StatusCode)
	}
}

// discoverSitemaps reads sitemaps from robots.txt (or /sitemap.xml), following
// sitemap indexes, and returns same-origin page URLs mapped to their sitemap.
func discoverSitemaps(ctx context.Context, client *http.Client, origin string, rt *robots.RobotsTxt, userAgent string, maxURLs int) ([]report.SitemapInfo, map[string]string) {
	type item struct{ url, source string }
	var queue []item
	if rt != nil {
		for _, s := range rt.Sitemaps {
			queue = append(queue, item{strings.TrimSpace(s), "robots.txt"})
		}
	}
	if len(queue) == 0 {
		queue = append(queue, item{origin + "/sitemap.xml", "default"})
	}

	originURL, _ := url.Parse(origin)
	seen := map[string]bool{}
	pages := map[string]string{}
	var infos []report.SitemapInfo

	for len(queue) > 0 && len(infos) < maxSitemapFiles && ctx.Err() == nil {
		it := queue[0]
		queue = queue[1:]
		if seen[it.url] {
			continue
		}
		seen[it.url] = true

		info := report.SitemapInfo{URL: it.url, Source: it.source}
		res := httpx.Fetch(ctx, client, it.url, userAgent, maxSitemapBytes)
		info.StatusCode = res.StatusCode
		switch {
		case res.Err != nil && res.StatusCode == 0:
			info.Error = res.Err.Error()
		case res.StatusCode != 200:
			info.Error = fmt.Sprintf("HTTP %d", res.StatusCode)
		default:
			parsed, err := sitemap.Parse(strings.NewReader(string(res.Body)))
			if err != nil {
				info.Error = err.Error()
				break
			}
			if parsed.IsIndex {
				info.IsIndex = true
				info.URLCount = len(parsed.Sitemaps)
				for _, s := range parsed.Sitemaps {
					queue = append(queue, item{strings.TrimSpace(s.Loc), "index"})
				}
				break
			}
			info.URLCount = len(parsed.URLs)
			for _, u := range parsed.URLs {
				norm, err := crawl.NormalizeURL(strings.TrimSpace(u.Loc), nil)
				if err != nil {
					continue
				}
				pu, err := url.Parse(norm)
				if err != nil || !crawl.IsSameOrigin(pu, originURL) {
					continue
				}
				if _, ok := pages[norm]; !ok && len(pages) < maxURLs {
					pages[norm] = it.url
				}
			}
		}
		infos = append(infos, info)
	}
	return infos, pages
}

// crawlerAccess converts the AI crawler evaluation into the report map.
func crawlerAccess(rep *aiaccess.Report) map[string]report.Policy {
	if rep == nil {
		return map[string]report.Policy{"note": {Status: "no_robots_txt"}}
	}
	policies := make(map[string]report.Policy, len(rep.Results))
	for _, result := range rep.Results {
		status := "allowed"
		if !result.Allowed {
			status = "blocked"
		}
		key := fmt.Sprintf("%s (%s)", result.Crawler.Name, result.Crawler.Purpose)
		policies[key] = report.Policy{
			Status:      status,
			Operator:    result.Crawler.Operator,
			Purpose:     string(result.Crawler.Purpose),
			MatchedRule: result.Rule,
		}
	}
	return policies
}

func buildMarkdownSummary(r *report.AuditReport) string {
	var b strings.Builder
	b.WriteString("# RankCore Audit Report\n\n")
	fmt.Fprintf(&b, "**Target:** %s  \n", r.Target)
	fmt.Fprintf(&b, "**Generated:** %s  \n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "**RankCore:** %s | **Ruleset:** %s\n\n", r.RankCoreVersion, r.RulesetVersion)

	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(&b, "- Pages requested: %d\n", r.Coverage.PagesRequested)
	fmt.Fprintf(&b, "- Pages analyzed: %d\n", r.Coverage.PagesAnalyzed)
	fmt.Fprintf(&b, "- Pages skipped: %d\n", r.Coverage.PagesSkipped)
	fmt.Fprintf(&b, "- Limit reached: %v\n\n", r.Coverage.LimitReached)

	counts := map[string]int{}
	byRule := map[string]int{}
	for _, f := range r.Findings {
		counts[f.Severity]++
		byRule[f.ID]++
	}
	b.WriteString("## Findings Summary\n\n")
	b.WriteString("| Severity | Count |\n|----------|-------|\n")
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		if c, ok := counts[sev]; ok {
			fmt.Fprintf(&b, "| %s | %d |\n", sev, c)
		}
	}
	fmt.Fprintf(&b, "\n**Total findings:** %d\n\n", len(r.Findings))

	if len(byRule) > 0 {
		ids := make([]string, 0, len(byRule))
		for id := range byRule {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		b.WriteString("| Rule | Name | Pages |\n|------|------|-------|\n")
		for _, id := range ids {
			fmt.Fprintf(&b, "| %s | %s | %d |\n", id, rules.Name(id), byRule[id])
		}
		b.WriteString("\n")
	}

	if len(r.Findings) > 0 {
		b.WriteString("## Findings\n\n")
		for _, f := range r.Findings {
			fmt.Fprintf(&b, "### %s — %s\n\n", f.ID, f.Summary)
			fmt.Fprintf(&b, "- **URL:** %s\n", f.URL)
			fmt.Fprintf(&b, "- **Severity:** %s · **Category:** %s · **Automation:** %s\n", f.Severity, f.Category, f.Automation)
			if f.Remediation != "" {
				fmt.Fprintf(&b, "- **Fix:** %s\n", f.Remediation)
			}
			b.WriteString("\n")
		}
	}

	if len(r.Sitemaps) > 0 {
		b.WriteString("## Sitemaps\n\n")
		for _, s := range r.Sitemaps {
			status := fmt.Sprintf("%d URLs", s.URLCount)
			if s.IsIndex {
				status = fmt.Sprintf("index of %d sitemaps", s.URLCount)
			}
			if s.Error != "" {
				status = "error: " + s.Error
			}
			fmt.Fprintf(&b, "- %s (%s) — %s\n", s.URL, s.Source, status)
		}
		b.WriteString("\n")
	}

	if len(r.CrawlerAccess) > 0 {
		b.WriteString("## Crawler Access\n\n")
		names := make([]string, 0, len(r.CrawlerAccess))
		for name := range r.CrawlerAccess {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(&b, "- **%s:** %s\n", name, r.CrawlerAccess[name].Status)
		}
		b.WriteString("\n_Blocking training crawlers is a publisher choice and is never reported as an SEO issue._\n\n")
	}

	if len(r.Limitations) > 0 {
		b.WriteString("## Limitations\n\n")
		for _, l := range r.Limitations {
			fmt.Fprintf(&b, "- %s\n", l)
		}
	}

	return b.String()
}
