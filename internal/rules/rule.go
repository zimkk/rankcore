package rules

import (
	"net/url"
	"sort"
	"strings"

	"github.com/zimkk/rankcore/internal/aiaccess"
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/robots"
)

// Identified is anything with a stable rule identity.
type Identified interface {
	// ID returns the stable identifier of the rule (e.g. "RC-HTTP-004").
	ID() string

	// Version returns the rule version for schema tracking.
	Version() int
}

// Rule represents a deterministic SEO check on a single page.
type Rule interface {
	Identified

	// Evaluate processes a page snapshot and optionally returns a finding.
	Evaluate(snapshot *crawl.PageSnapshot) *report.Finding
}

// SiteRule is a deterministic check that needs the whole crawl: link targets,
// canonical targets, sitemaps, duplicates and the link graph.
type SiteRule interface {
	Identified

	EvaluateSite(site *Site) []report.Finding
}

// Site is everything a SiteRule can see after the crawl.
type Site struct {
	Seed  string // normalized seed URL
	Pages []*crawl.PageSnapshot
	ByURL map[string]*crawl.PageSnapshot

	Robots        *robots.RobotsTxt
	RobotsURL     string
	CrawlerAccess []aiaccess.PolicyResult

	Sitemaps    []report.SitemapInfo
	SitemapURLs map[string]string // page URL -> sitemap file that lists it
}

// NewSite indexes crawled pages by their requested URL.
func NewSite(seed string, pages []*crawl.PageSnapshot) *Site {
	s := &Site{Seed: seed, Pages: pages, ByURL: make(map[string]*crawl.PageSnapshot, len(pages)), SitemapURLs: map[string]string{}}
	for _, p := range pages {
		s.ByURL[p.URL] = p
	}
	return s
}

// Page returns the crawled snapshot for an absolute URL, if any.
func (s *Site) Page(u string) *crawl.PageSnapshot {
	if p, ok := s.ByURL[u]; ok {
		return p
	}
	if norm, err := crawl.NormalizeURL(u, nil); err == nil {
		return s.ByURL[norm]
	}
	return nil
}

// isHTMLOK reports whether a snapshot is a successfully fetched HTML page. An
// unknown content type is treated as HTML so hand-built snapshots still apply.
func isHTMLOK(s *crawl.PageSnapshot) bool {
	return s.StatusCode == 200 && (s.ContentType == "" || s.IsHTML())
}

// resolve turns a possibly relative href into a normalized absolute URL.
func resolve(href, pageURL string) (string, bool) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", false
	}
	abs, err := crawl.NormalizeURL(href, base)
	if err != nil {
		return "", false
	}
	return abs, true
}

func isAbsoluteURL(href string) bool {
	u, err := url.Parse(strings.TrimSpace(href))
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// capList sorts a list and keeps at most n entries for evidence.
func capList(items []string, n int) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	if len(out) > n {
		out = out[:n]
	}
	return out
}
