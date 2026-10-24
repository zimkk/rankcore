package crawl

import (
	"strings"
	"time"
)

// Config holds settings for a crawling session.
type Config struct {
	RenderMode     RenderMode    `json:"render_mode"`
	MaxPages       int           `json:"max_pages"`
	MaxDepth       int           `json:"max_depth"`
	Concurrency    int           `json:"concurrency"`
	IncludePattern string        `json:"include_pattern,omitempty"`
	ExcludePattern string        `json:"exclude_pattern,omitempty"`
	OutDir         string        `json:"out_dir"`
	OutputFormat   string        `json:"output_format"`
	UserAgent      string        `json:"user_agent"`
	Timeout        time.Duration `json:"timeout"`
	MaxBodyBytes   int64         `json:"max_body_bytes"`
	AllowPrivate   bool          `json:"allow_private"`
	RespectRobots  bool          `json:"respect_robots"`
}

// DefaultConfig returns the bounded, conservative defaults used for audits.
func DefaultConfig() Config {
	return Config{
		RenderMode:    RenderModeOff,
		MaxPages:      500,
		MaxDepth:      10,
		Concurrency:   5,
		UserAgent:     "RankCoreBot/1.0 (+https://github.com/zimkk/rankcore)",
		Timeout:       20 * time.Second,
		OutputFormat:  "json",
		MaxBodyBytes:  5 * 1024 * 1024,
		RespectRobots: true,
	}
}

type RenderMode string

const (
	RenderModeAuto     RenderMode = "auto"
	RenderModeOff      RenderMode = "off"
	RenderModeRequired RenderMode = "required"
)

// RenderedSnapshot holds the rendered-DOM observation for a page, captured
// through an already-installed Chromium-family browser when available.
type RenderedSnapshot struct {
	Title         string   `json:"title,omitempty"`
	Canonical     string   `json:"canonical,omitempty"`
	RobotsMeta    string   `json:"robots_meta,omitempty"`
	InternalLinks []string `json:"internal_links,omitempty"`
	TextLength    int      `json:"text_length"`
}

// LinkInfo is an anchor found on a page, resolved against the page URL.
type LinkInfo struct {
	Raw      string // href exactly as written
	URL      string // absolute, normalized URL; empty when not an http(s) target
	Text     string
	Internal bool // same origin as the crawl seed
	Nofollow bool
	HasHref  bool
}

// PageSnapshot represents the extracted contents of a page.
type PageSnapshot struct {
	URL             string
	FinalURL        string
	StatusCode      int
	ContentType     string
	Headers         map[string][]string
	Title           string
	MetaDesc        string
	Canonical       string   // raw href of the first canonical
	Canonicals      []string // every canonical href
	RobotsMeta      string   // content of <meta name="robots"> / <meta name="googlebot">
	XRobotsTag      string   // all X-Robots-Tag response header values, comma-joined
	Viewport        string
	Lang            string
	H1              []string
	Headings        map[string][]string // key is "h1" through "h6"
	TextLength      int
	WordCount       int
	Hreflang        map[string]string
	Links           []LinkInfo
	InternalLinks   []string // absolute, normalized, de-duplicated same-origin URLs
	ExternalLinks   []string // absolute external URLs
	Images          []ImageInfo
	JSONLD          []string
	OpenGraph       map[string]string
	TwitterCard     map[string]string
	RedirectChain   []string
	RedirectLoop    bool
	FetchError      string
	BlockedByRobots bool   // disallowed for Googlebot
	RobotsRule      string // matched robots.txt rule when blocked
	Skipped         string // reason the page was not fetched (e.g. "robots")
	Depth           int
	Source          string // how the URL was discovered: seed, link, sitemap, canonical, hreflang
	InSitemap       bool
	DurationMS      int64
	BodyBytes       int
	Truncated       bool
	Rendered        *RenderedSnapshot
}

// ImageInfo tracks an image element and its alt attribute.
type ImageInfo struct {
	Src    string
	Alt    string
	HasAlt bool
}

// IsHTML reports whether the response was an HTML document.
func (p *PageSnapshot) IsHTML() bool {
	ct := strings.ToLower(p.ContentType)
	return strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml")
}

// Indexable reports whether the page was fetched successfully, is HTML, and
// carries no noindex directive.
func (p *PageSnapshot) Indexable() bool {
	return p.StatusCode == 200 && p.IsHTML() && !p.Noindex()
}

// Noindex reports whether the page carries a noindex directive in either the
// meta robots tag or the X-Robots-Tag header.
func (p *PageSnapshot) Noindex() bool {
	return hasRobotsDirective(p.RobotsMeta, "noindex") ||
		hasRobotsDirective(p.RobotsMeta, "none") ||
		hasRobotsDirective(p.XRobotsTag, "noindex") ||
		hasRobotsDirective(p.XRobotsTag, "none")
}

// hasRobotsDirective matches a directive in a robots value. It understands the
// user-agent-scoped header form ("googlebot: noindex") and ignores scopes for
// other crawlers.
func hasRobotsDirective(value, directive string) bool {
	for _, part := range strings.Split(strings.ToLower(value), ",") {
		part = strings.TrimSpace(part)
		if i := strings.Index(part, ":"); i >= 0 {
			if scope := strings.TrimSpace(part[:i]); !isRobotsDirective(scope) {
				// "<crawler>: <directive>" — only Googlebot-scoped values count.
				if scope != "googlebot" {
					continue
				}
				part = strings.TrimSpace(part[i+1:])
			}
		}
		if part == directive {
			return true
		}
	}
	return false
}

func isRobotsDirective(s string) bool {
	switch s {
	case "all", "noindex", "nofollow", "none", "noarchive", "nosnippet", "notranslate",
		"noimageindex", "unavailable_after", "max-snippet", "max-image-preview", "max-video-preview", "indexifembedded":
		return true
	}
	return false
}
