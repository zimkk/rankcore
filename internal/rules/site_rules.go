package rules

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/zimkk/rankcore/internal/aiaccess"
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/graph"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/robots"
)

// linkTargetBroken reports whether a crawled link target failed outright.
func linkTargetBroken(t *crawl.PageSnapshot) bool {
	if t.Skipped != "" {
		return false
	}
	return t.StatusCode >= 400 || t.RedirectLoop || (t.StatusCode == 0 && t.FetchError != "")
}

func statusLabel(t *crawl.PageSnapshot) string {
	switch {
	case t.RedirectLoop:
		return "redirect loop"
	case t.StatusCode == 0:
		return "fetch error: " + t.FetchError
	default:
		return fmt.Sprintf("%d", t.StatusCode)
	}
}

// ── RC-LINK-001 ─────────────────────────────────────────────────────────────

// BrokenLinkRule flags pages that link to internal URLs which return 4xx/5xx
// or cannot be fetched.
type BrokenLinkRule struct{}

func (r *BrokenLinkRule) ID() string   { return "RC-LINK-001" }
func (r *BrokenLinkRule) Version() int { return 2 }

func (r *BrokenLinkRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, p := range site.Pages {
		if !isHTMLOK(p) {
			continue
		}
		var broken []string
		for _, l := range p.InternalLinks {
			if t := site.Page(l); t != nil && linkTargetBroken(t) {
				broken = append(broken, fmt.Sprintf("%s (%s)", l, statusLabel(t)))
			}
		}
		if len(broken) == 0 {
			continue
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 1.0, URL: p.URL,
			Summary:     fmt.Sprintf("Page links to %d broken internal URL(s)", len(broken)),
			Remediation: "Update or remove these links, or restore/redirect the missing destinations.",
			Evidence:    map[string]interface{}{"broken_links": capList(broken, 20), "total": len(broken)},
		})
	}
	return out
}

// ── RC-LINK-004 ─────────────────────────────────────────────────────────────

// RedirectingLinkRule flags internal links that go through a redirect.
type RedirectingLinkRule struct{}

func (r *RedirectingLinkRule) ID() string   { return "RC-LINK-004" }
func (r *RedirectingLinkRule) Version() int { return 1 }

func (r *RedirectingLinkRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, p := range site.Pages {
		if !isHTMLOK(p) {
			continue
		}
		var redirecting []string
		for _, l := range p.InternalLinks {
			t := site.Page(l)
			if t == nil || len(t.RedirectChain) == 0 || t.RedirectLoop || t.StatusCode != 200 {
				continue
			}
			redirecting = append(redirecting, l+" -> "+t.FinalURL)
		}
		if len(redirecting) == 0 {
			continue
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "low", Confidence: 1.0, URL: p.URL,
			Summary:     fmt.Sprintf("Page links to %d internal URL(s) that redirect", len(redirecting)),
			Remediation: "Point internal links straight at the final destination URL.",
			Evidence:    map[string]interface{}{"redirecting_links": capList(redirecting, 20), "total": len(redirecting)},
		})
	}
	return out
}

// ── RC-CANON-002 ────────────────────────────────────────────────────────────

// CanonicalTargetRule checks that a page's canonical target is an absolute URL
// that returns 200 directly, is indexable, and does not itself canonicalize
// elsewhere.
type CanonicalTargetRule struct{}

func (r *CanonicalTargetRule) ID() string   { return "RC-CANON-002" }
func (r *CanonicalTargetRule) Version() int { return 2 }

func (r *CanonicalTargetRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, p := range site.Pages {
		if !isHTMLOK(p) || strings.TrimSpace(p.Canonical) == "" {
			continue
		}
		raw := strings.TrimSpace(p.Canonical)
		target, ok := resolve(raw, p.FinalURL)
		if !ok {
			out = append(out, r.finding(p, "high", "Canonical URL is malformed", raw, "", nil))
			continue
		}
		if !isAbsoluteURL(raw) {
			out = append(out, r.finding(p, "low", "Canonical URL is relative", raw, target, nil))
		}
		if target == p.URL || target == p.FinalURL {
			continue // self-canonical
		}
		t := site.Page(target)
		if t == nil || t.Skipped != "" {
			continue // off-site or not crawled; nothing deterministic to say
		}
		switch {
		case linkTargetBroken(t):
			out = append(out, r.finding(p, "high", "Canonical target is broken ("+statusLabel(t)+")", raw, target, t))
		case len(t.RedirectChain) > 0:
			out = append(out, r.finding(p, "high", "Canonical target redirects to "+t.FinalURL, raw, target, t))
		case t.Noindex():
			out = append(out, r.finding(p, "high", "Canonical target is noindexed", raw, target, t))
		default:
			if tc := strings.TrimSpace(t.Canonical); tc != "" {
				if next, ok := resolve(tc, t.FinalURL); ok && next != target {
					out = append(out, r.finding(p, "medium", "Canonical target declares a different canonical (chain)", raw, target, t))
				}
			}
		}
	}
	return out
}

func (r *CanonicalTargetRule) finding(p *crawl.PageSnapshot, sev, summary, raw, target string, t *crawl.PageSnapshot) report.Finding {
	ev := map[string]interface{}{"canonical": raw}
	if target != "" {
		ev["canonical_resolved"] = target
	}
	if t != nil {
		ev["target_status"] = t.StatusCode
		ev["target_final_url"] = t.FinalURL
		if t.Canonical != "" {
			ev["target_canonical"] = t.Canonical
		}
	}
	return report.Finding{
		ID: r.ID(), RuleVersion: r.Version(), Severity: sev, Confidence: 1.0, URL: p.URL,
		Summary:     summary,
		Remediation: "Point the canonical at the absolute, indexable URL that returns 200 directly.",
		Evidence:    ev,
	}
}

// ── RC-CANON-004 ────────────────────────────────────────────────────────────

// MultipleCanonicalRule flags pages that declare conflicting canonicals.
// Search engines ignore all of them in that case.
type MultipleCanonicalRule struct{}

func (r *MultipleCanonicalRule) ID() string   { return "RC-CANON-004" }
func (r *MultipleCanonicalRule) Version() int { return 1 }

func (r *MultipleCanonicalRule) Evaluate(s *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(s) || len(s.Canonicals) < 2 {
		return nil
	}
	distinct := map[string]bool{}
	for _, c := range s.Canonicals {
		if abs, ok := resolve(c, s.FinalURL); ok {
			distinct[abs] = true
		} else {
			distinct[c] = true
		}
	}
	if len(distinct) < 2 {
		return nil
	}
	return &report.Finding{
		ID: r.ID(), RuleVersion: r.Version(), Severity: "high", Confidence: 1.0, URL: s.URL,
		Summary:     fmt.Sprintf("Page declares %d different canonical URLs", len(distinct)),
		Remediation: "Emit exactly one <link rel=\"canonical\">; check layouts, SEO plugins and head managers for duplicates.",
		Evidence:    map[string]interface{}{"canonicals": s.Canonicals},
	}
}

// ── RC-GRAPH-001 / RC-GRAPH-002 ─────────────────────────────────────────────

// inboundLinks counts, for each URL, the distinct other pages linking to it.
func inboundLinks(site *Site) map[string]int {
	in := map[string]int{}
	for _, p := range site.Pages {
		for _, l := range p.InternalLinks {
			if l != p.URL && l != p.FinalURL {
				in[l]++
			}
		}
	}
	return in
}

// OrphanPageRule flags indexable pages listed in a sitemap that no crawled
// page links to. Pages found by following links cannot be orphans.
type OrphanPageRule struct{}

func (r *OrphanPageRule) ID() string   { return "RC-GRAPH-001" }
func (r *OrphanPageRule) Version() int { return 2 }

func (r *OrphanPageRule) EvaluateSite(site *Site) []report.Finding {
	in := inboundLinks(site)
	seedFinal := site.Seed
	if s := site.Page(site.Seed); s != nil {
		seedFinal = s.FinalURL
	}
	var out []report.Finding
	for _, p := range site.Pages {
		if !p.InSitemap || !p.Indexable() || p.URL == site.Seed || p.URL == seedFinal {
			continue
		}
		if in[p.URL] > 0 || (p.FinalURL != p.URL && in[p.FinalURL] > 0) {
			continue
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 0.8, URL: p.URL,
			Summary:     "Page is listed in the sitemap but no crawled page links to it",
			Remediation: "Add internal links from related pages or navigation so crawlers and users can reach it.",
			Evidence:    map[string]interface{}{"inbound_count": 0, "sitemap": site.SitemapURLs[p.URL]},
		})
	}
	return out
}

// CrawlDepthRule flags indexable pages more than maxClickDepth clicks from the seed.
type CrawlDepthRule struct{}

const maxClickDepth = 5

func (r *CrawlDepthRule) ID() string   { return "RC-GRAPH-002" }
func (r *CrawlDepthRule) Version() int { return 2 }

func (r *CrawlDepthRule) EvaluateSite(site *Site) []report.Finding {
	g := graph.New()
	for _, p := range site.Pages {
		edges := p.InternalLinks
		if p.FinalURL != p.URL {
			edges = append(append([]string{}, edges...), p.FinalURL)
		}
		g.AddPage(p.URL, edges)
	}
	depths := g.Depth(site.Seed)
	var out []report.Finding
	for _, p := range site.Pages {
		d, ok := depths[p.URL]
		if !ok || d <= maxClickDepth || !p.Indexable() {
			continue
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "low", Confidence: 0.9, URL: p.URL,
			Summary:     fmt.Sprintf("Page is %d clicks from the start URL", d),
			Remediation: "Link to it from hub, category or navigation pages to reduce click depth.",
			Evidence:    map[string]interface{}{"depth": d, "max": maxClickDepth},
		})
	}
	return out
}

// ── RC-META-006 / RC-META-007 ───────────────────────────────────────────────

func duplicates(site *Site, field func(*crawl.PageSnapshot) string) map[string][]string {
	groups := map[string][]string{}
	for _, p := range site.Pages {
		if !p.Indexable() {
			continue
		}
		v := strings.ToLower(strings.Join(strings.Fields(field(p)), " "))
		if v == "" {
			continue
		}
		groups[v] = append(groups[v], p.URL)
	}
	for k, urls := range groups {
		if len(urls) < 2 {
			delete(groups, k)
		} else {
			sort.Strings(urls)
		}
	}
	return groups
}

// DuplicateTitleRule flags indexable pages that share an identical title.
type DuplicateTitleRule struct{}

func (r *DuplicateTitleRule) ID() string   { return "RC-META-006" }
func (r *DuplicateTitleRule) Version() int { return 1 }

func (r *DuplicateTitleRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, urls := range duplicates(site, func(p *crawl.PageSnapshot) string { return p.Title }) {
		for _, u := range urls {
			out = append(out, report.Finding{
				ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 1.0, URL: u,
				Summary:     fmt.Sprintf("Title is shared with %d other indexable page(s)", len(urls)-1),
				Remediation: "Give each indexable page a unique title, or canonicalize true duplicates to one URL.",
				Evidence:    map[string]interface{}{"title": site.ByURL[u].Title, "pages": capList(urls, 10), "total": len(urls)},
			})
		}
	}
	return out
}

// DuplicateDescriptionRule flags indexable pages that share a meta description.
type DuplicateDescriptionRule struct{}

func (r *DuplicateDescriptionRule) ID() string   { return "RC-META-007" }
func (r *DuplicateDescriptionRule) Version() int { return 1 }

func (r *DuplicateDescriptionRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, urls := range duplicates(site, func(p *crawl.PageSnapshot) string { return p.MetaDesc }) {
		for _, u := range urls {
			out = append(out, report.Finding{
				ID: r.ID(), RuleVersion: r.Version(), Severity: "low", Confidence: 1.0, URL: u,
				Summary:     fmt.Sprintf("Meta description is shared with %d other indexable page(s)", len(urls)-1),
				Remediation: "Write a description specific to each page's content.",
				Evidence:    map[string]interface{}{"pages": capList(urls, 10), "total": len(urls)},
			})
		}
	}
	return out
}

// ── RC-SITEMAP-* ────────────────────────────────────────────────────────────

// SitemapPresenceRule reports a missing sitemap, or sitemaps that cannot be read.
type SitemapPresenceRule struct{}

func (r *SitemapPresenceRule) ID() string   { return "RC-SITEMAP-001" }
func (r *SitemapPresenceRule) Version() int { return 1 }

func (r *SitemapPresenceRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	readable := 0
	for _, sm := range site.Sitemaps {
		if sm.Error == "" {
			readable++
			continue
		}
		if sm.Source == "default" {
			continue // probing /sitemap.xml is a guess; absence is reported below
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 1.0, URL: sm.URL,
			Summary:     "Sitemap declared in robots.txt cannot be read",
			Remediation: "Serve the sitemap with HTTP 200 and valid <urlset>/<sitemapindex> XML, or fix the Sitemap: line in robots.txt.",
			Evidence:    map[string]interface{}{"status_code": sm.StatusCode, "error": sm.Error},
		})
	}
	if readable == 0 && len(out) == 0 {
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "low", Confidence: 0.9, URL: site.Seed,
			Summary:     "No XML sitemap was found (robots.txt Sitemap: lines or /sitemap.xml)",
			Remediation: "Publish an XML sitemap of canonical, indexable URLs and reference it from robots.txt.",
			Evidence:    map[string]interface{}{"checked": len(site.Sitemaps)},
		})
	}
	return out
}

// SitemapEntryRule flags sitemap URLs that are not indexable canonical 200 pages.
type SitemapEntryRule struct{}

func (r *SitemapEntryRule) ID() string   { return "RC-SITEMAP-002" }
func (r *SitemapEntryRule) Version() int { return 1 }

func (r *SitemapEntryRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, p := range site.Pages {
		if !p.InSitemap || p.Skipped != "" {
			continue
		}
		var problems []string
		switch {
		case linkTargetBroken(p):
			problems = append(problems, "returns "+statusLabel(p))
		case len(p.RedirectChain) > 0:
			problems = append(problems, "redirects to "+p.FinalURL)
		case p.StatusCode == 200 && p.Noindex():
			problems = append(problems, "is noindexed")
		}
		if p.StatusCode == 200 && len(p.RedirectChain) == 0 && strings.TrimSpace(p.Canonical) != "" {
			if c, ok := resolve(p.Canonical, p.FinalURL); ok && c != p.URL {
				problems = append(problems, "canonicalizes to "+c)
			}
		}
		if p.BlockedByRobots {
			problems = append(problems, "is blocked by robots.txt")
		}
		if len(problems) == 0 {
			continue
		}
		sev := "medium"
		if linkTargetBroken(p) {
			sev = "high"
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: sev, Confidence: 1.0, URL: p.URL,
			Summary:     "Sitemap lists a URL that " + strings.Join(problems, " and "),
			Remediation: "List only canonical, indexable URLs that return 200 in the sitemap.",
			Evidence:    map[string]interface{}{"sitemap": site.SitemapURLs[p.URL], "status_code": p.StatusCode, "problems": problems},
		})
	}
	return out
}

// ── RC-HREFLANG-002 ─────────────────────────────────────────────────────────

// HreflangReturnRule checks that hreflang alternates link back and that the
// cluster includes the page itself.
type HreflangReturnRule struct{}

func (r *HreflangReturnRule) ID() string   { return "RC-HREFLANG-002" }
func (r *HreflangReturnRule) Version() int { return 1 }

func (r *HreflangReturnRule) EvaluateSite(site *Site) []report.Finding {
	var out []report.Finding
	for _, p := range site.Pages {
		if !isHTMLOK(p) || len(p.Hreflang) == 0 {
			continue
		}
		var missingReturn []string
		self := false
		for lang, href := range p.Hreflang {
			alt, ok := resolve(href, p.FinalURL)
			if !ok {
				continue
			}
			if alt == p.URL || alt == p.FinalURL {
				self = true
				continue
			}
			t := site.Page(alt)
			if t == nil || !isHTMLOK(t) {
				continue
			}
			back := false
			for _, h := range t.Hreflang {
				if b, ok := resolve(h, t.FinalURL); ok && (b == p.URL || b == p.FinalURL) {
					back = true
					break
				}
			}
			if !back {
				missingReturn = append(missingReturn, lang+" "+alt)
			}
		}
		if len(missingReturn) == 0 && self {
			continue
		}
		ev := map[string]interface{}{"self_referenced": self}
		summary := "Hreflang annotations do not include a self-reference"
		if len(missingReturn) > 0 {
			ev["missing_return_links"] = capList(missingReturn, 20)
			summary = fmt.Sprintf("%d hreflang alternate(s) do not link back to this page", len(missingReturn))
		}
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 1.0, URL: p.URL,
			Summary:     summary,
			Remediation: "Every page in an hreflang cluster must list all alternates, including itself, and each alternate must link back.",
			Evidence:    ev,
		})
	}
	return out
}

// ── RC-AI-001 ───────────────────────────────────────────────────────────────

// AISearchAccessRule flags robots.txt blocks on search and user-fetch
// crawlers. Training crawlers are a publisher choice and are never flagged.
type AISearchAccessRule struct{}

func (r *AISearchAccessRule) ID() string   { return "RC-AI-001" }
func (r *AISearchAccessRule) Version() int { return 1 }

func (r *AISearchAccessRule) EvaluateSite(site *Site) []report.Finding {
	var search, fetch []string
	for _, res := range site.CrawlerAccess {
		if res.Allowed {
			continue
		}
		switch res.Crawler.Purpose {
		case aiaccess.PurposeSearch:
			search = append(search, res.Crawler.Name)
		case aiaccess.PurposeUserFetch:
			fetch = append(fetch, res.Crawler.Name)
		}
	}
	robotsURL := site.RobotsURL
	if robotsURL == "" {
		if u, err := url.Parse(site.Seed); err == nil {
			robotsURL = u.Scheme + "://" + u.Host + "/robots.txt"
		}
	}
	var out []report.Finding
	if len(search) > 0 {
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "high", Confidence: 1.0, URL: robotsURL,
			Summary:     "robots.txt blocks search or AI-search crawlers from the site root",
			Remediation: "Allow the search crawlers you want visibility in. Training crawlers (GPTBot, ClaudeBot, Google-Extended, CCBot…) can stay blocked independently.",
			Automation:  report.AutomationReview,
			Evidence:    map[string]interface{}{"blocked_search_crawlers": capList(search, 50), "purpose": "search"},
		})
	}
	if len(fetch) > 0 {
		out = append(out, report.Finding{
			ID: r.ID(), RuleVersion: r.Version(), Severity: "low", Confidence: 1.0, URL: robotsURL,
			Summary:     "robots.txt blocks user-initiated AI fetchers from the site root",
			Remediation: "If assistants should be able to open your pages when users share links, allow these agents.",
			Automation:  report.AutomationReview,
			Evidence:    map[string]interface{}{"blocked_user_fetchers": capList(fetch, 50), "purpose": "user_fetch"},
		})
	}
	return out
}

// MarkRobotsBlocked sets BlockedByRobots on every page Googlebot may not fetch.
func MarkRobotsBlocked(site *Site) {
	if site.Robots == nil {
		return
	}
	for _, p := range site.Pages {
		u, err := url.Parse(p.URL)
		if err != nil {
			continue
		}
		path := robots.PathOf(u)
		if site.Robots.IsDisallowed("Googlebot", path) {
			p.BlockedByRobots = true
			p.RobotsRule = site.Robots.MatchedRule("Googlebot", path)
		}
	}
}
