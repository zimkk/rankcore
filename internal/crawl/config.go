package crawl

import "time"

import "strings"

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
}

// DefaultConfig returns the bounded, conservative defaults used for audits.
func DefaultConfig() Config {
	return Config{
		RenderMode:   RenderModeAuto,
		MaxPages:     500,
		MaxDepth:     10,
		Concurrency:  5,
		UserAgent:    "RankCoreBot/1.0 (+https://github.com/zimkk/rankcore)",
		Timeout:      20 * time.Second,
		OutputFormat: "json",
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

// PageSnapshot represents the extracted contents of a page.
type PageSnapshot struct {
	URL           string
	FinalURL      string
	StatusCode    int
	ContentType   string
	Headers       map[string][]string
	Title         string
	MetaDesc      string
	Canonical     string
	RobotsMeta    string // content attribute of <meta name="robots">
	XRobotsTag    string // X-Robots-Tag response header
	Viewport      string // content attribute of <meta name="viewport">
	Lang          string
	H1            []string
	Headings      map[string][]string // key is "h1" through "h6"
	TextLength    int
	Hreflang      map[string]string
	InternalLinks []string
	ExternalLinks []string
	AnchorTexts   map[string]string // href -> anchor text
	Images        []ImageInfo
	JSONLD        []string
	OpenGraph     map[string]string
	TwitterCard   map[string]string
	RedirectChain []string
	RedirectLoop  bool
	FetchError    string
	BlockedByRobots bool
	RobotsRule    string // matched robots.txt rule when blocked
	Depth         int
	Rendered      *RenderedSnapshot
}

// ImageInfo tracks an image element and its alt attribute.
type ImageInfo struct {
	Src    string
	Alt    string
	HasAlt bool
}

// Noindex reports whether the page carries a noindex directive in either the
// meta robots tag or the X-Robots-Tag header.
func (p *PageSnapshot) Noindex() bool {
	return hasRobotsDirective(p.RobotsMeta, "noindex") ||
		hasRobotsDirective(p.RobotsMeta, "none") ||
		hasRobotsDirective(p.XRobotsTag, "noindex") ||
		hasRobotsDirective(p.XRobotsTag, "none")
}

func hasRobotsDirective(value, directive string) bool {
	for _, part := range strings.Split(strings.ToLower(value), ",") {
		if strings.TrimSpace(part) == directive {
			return true
		}
	}
	return false
}

