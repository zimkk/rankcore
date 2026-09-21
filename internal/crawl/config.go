package crawl

import "time"

// Config holds settings for a crawling session.
type Config struct {
	RenderMode     RenderMode    `json:"render_mode"`
	MaxPages       int           `json:"max_pages"`
	Concurrency    int           `json:"concurrency"`
	IncludePattern string        `json:"include_pattern,omitempty"`
	ExcludePattern string        `json:"exclude_pattern,omitempty"`
	OutDir         string        `json:"out_dir"`
	OutputFormat   string        `json:"output_format"`
	UserAgent      string        `json:"user_agent"`
	Timeout        time.Duration `json:"timeout"`
}

type RenderMode string

const (
	RenderModeAuto     RenderMode = "auto"
	RenderModeOff      RenderMode = "off"
	RenderModeRequired RenderMode = "required"
)

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
	Hreflang      map[string]string
	InternalLinks []string
	ExternalLinks []string
	JSONLD        []string
}
