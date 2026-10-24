package crawl

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/zimkk/rankcore/internal/extract"
	"github.com/zimkk/rankcore/internal/httpx"
	"github.com/zimkk/rankcore/internal/robots"
)

// Seed is a URL the crawl starts from, with how it was discovered.
type Seed struct {
	URL    string
	Source string
}

type queued struct {
	url    string
	depth  int
	source string
}

// Crawler manages the URL frontier and concurrent execution. A crawl ends when
// the frontier is empty and no fetch is in flight.
type Crawler struct {
	Config Config
	Robots *robots.RobotsTxt // when set and RespectRobots is true, disallowed URLs are not fetched

	client  *http.Client
	base    *url.URL
	include *regexp.Regexp
	exclude *regexp.Regexp

	mu       sync.Mutex
	cond     *sync.Cond
	queue    []queued
	pending  int // queued + in flight
	admitted map[string]bool
	results  []*PageSnapshot
	limitHit bool
}

// NewCrawler initializes a Crawler.
func NewCrawler(cfg Config) *Crawler {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 5
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 500
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 5 * 1024 * 1024
	}
	c := &Crawler{Config: cfg, admitted: make(map[string]bool)}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// LimitReached reports whether URLs were dropped because of MaxPages.
func (c *Crawler) LimitReached() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.limitHit
}

// Run crawls from the seeds (the first seed defines the origin) and returns
// every snapshot sorted by URL. It blocks until the crawl completes or ctx ends.
func (c *Crawler) Run(ctx context.Context, seeds []Seed) ([]*PageSnapshot, error) {
	if len(seeds) == 0 {
		return nil, nil
	}
	first, err := NormalizeURL(seeds[0].URL, nil)
	if err != nil {
		return nil, err
	}
	c.base, _ = url.Parse(first)
	c.client = httpx.NewSafeClient(c.Config.Timeout, c.Config.AllowPrivate)

	if c.Config.IncludePattern != "" {
		if c.include, err = compilePattern(c.Config.IncludePattern); err != nil {
			return nil, err
		}
	}
	if c.Config.ExcludePattern != "" {
		if c.exclude, err = compilePattern(c.Config.ExcludePattern); err != nil {
			return nil, err
		}
	}

	c.mu.Lock()
	for i, s := range seeds {
		norm, err := NormalizeURL(s.URL, c.base)
		if err != nil {
			continue
		}
		// The first seed is always admitted, even if include/exclude would drop it.
		c.enqueueLocked(norm, 0, s.Source, i == 0)
	}
	c.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < c.Config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.worker(ctx)
		}()
	}

	// Wake idle workers if the context is cancelled.
	stop := context.AfterFunc(ctx, func() {
		c.mu.Lock()
		c.cond.Broadcast()
		c.mu.Unlock()
	})
	wg.Wait()
	stop()

	sort.Slice(c.results, func(i, j int) bool { return c.results[i].URL < c.results[j].URL })
	return c.results, ctx.Err()
}

func (c *Crawler) worker(ctx context.Context) {
	for {
		c.mu.Lock()
		for len(c.queue) == 0 && c.pending > 0 && ctx.Err() == nil {
			c.cond.Wait()
		}
		if ctx.Err() != nil || len(c.queue) == 0 {
			c.cond.Broadcast()
			c.mu.Unlock()
			return
		}
		item := c.queue[0]
		c.queue = c.queue[1:]
		c.mu.Unlock()

		snap := c.process(ctx, item)

		c.mu.Lock()
		if snap != nil && ctx.Err() == nil {
			c.results = append(c.results, snap)
			for _, next := range discovered(snap) {
				c.enqueueLocked(next.url, item.depth+1, next.source, false)
			}
		}
		c.pending--
		c.cond.Broadcast()
		c.mu.Unlock()
	}
}

// enqueueLocked admits a normalized URL into the frontier. Caller holds c.mu.
func (c *Crawler) enqueueLocked(norm string, depth int, source string, force bool) {
	if c.admitted[norm] {
		return
	}
	u, err := url.Parse(norm)
	if err != nil || !IsSameOrigin(u, c.base) {
		return
	}
	if !force {
		if depth > c.Config.MaxDepth && c.Config.MaxDepth > 0 {
			return
		}
		if c.include != nil && !c.include.MatchString(norm) {
			return
		}
		if c.exclude != nil && c.exclude.MatchString(norm) {
			return
		}
	}
	if len(c.admitted) >= c.Config.MaxPages {
		c.limitHit = true
		return
	}
	c.admitted[norm] = true
	c.queue = append(c.queue, queued{url: norm, depth: depth, source: source})
	c.pending++
	c.cond.Signal()
}

type discoveredURL struct{ url, source string }

// discovered lists same-origin URLs a page points to: links, canonicals and
// hreflang alternates (so their status can be checked).
func discovered(p *PageSnapshot) []discoveredURL {
	var out []discoveredURL
	for _, l := range p.InternalLinks {
		out = append(out, discoveredURL{l, "link"})
	}
	base, err := url.Parse(p.FinalURL)
	if err != nil {
		return out
	}
	for _, c := range p.Canonicals {
		if abs, err := NormalizeURL(c, base); err == nil {
			out = append(out, discoveredURL{abs, "canonical"})
		}
	}
	langs := make([]string, 0, len(p.Hreflang))
	for l := range p.Hreflang {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	for _, l := range langs {
		if abs, err := NormalizeURL(p.Hreflang[l], base); err == nil {
			out = append(out, discoveredURL{abs, "hreflang"})
		}
	}
	return out
}

func (c *Crawler) process(ctx context.Context, item queued) *PageSnapshot {
	snap := &PageSnapshot{URL: item.url, FinalURL: item.url, Depth: item.depth, Source: item.source}

	if c.Config.RespectRobots && c.Robots != nil {
		if u, err := url.Parse(item.url); err == nil && c.Robots.IsDisallowed(c.Config.UserAgent, robots.PathOf(u)) {
			snap.Skipped = "robots"
			return snap
		}
	}

	res := httpx.Fetch(ctx, c.client, item.url, c.Config.UserAgent, c.Config.MaxBodyBytes)
	if ctx.Err() != nil {
		return nil
	}
	snap.StatusCode = res.StatusCode
	snap.RedirectChain = res.RedirectChain
	snap.RedirectLoop = res.RedirectLoop
	snap.DurationMS = res.Duration.Milliseconds()
	snap.BodyBytes = len(res.Body)
	snap.Truncated = res.Truncated
	if final, err := NormalizeURL(res.FinalURL, nil); err == nil {
		snap.FinalURL = final
	}
	if res.Err != nil {
		snap.FetchError = res.Err.Error()
		if res.RedirectLoop || res.StatusCode == 0 {
			return snap
		}
	}
	if res.Header != nil {
		snap.Headers = res.Header
		snap.ContentType = res.Header.Get("Content-Type")
		snap.XRobotsTag = strings.Join(res.Header.Values("X-Robots-Tag"), ", ")
	}
	if snap.ContentType == "" && len(res.Body) > 0 {
		snap.ContentType = http.DetectContentType(res.Body)
	}

	// Only same-origin final destinations are analysed; a redirect off-site is
	// recorded but its content belongs to someone else.
	finalURL, _ := url.Parse(snap.FinalURL)
	if snap.StatusCode != 200 || !snap.IsHTML() || finalURL == nil || !IsSameOrigin(finalURL, c.base) {
		return snap
	}

	meta, err := extract.Extract(bytes.NewReader(res.Body))
	if err != nil || meta == nil {
		return snap
	}
	ApplyMetadata(snap, meta, finalURL, c.base)
	return snap
}

// ApplyMetadata copies extracted HTML metadata into a snapshot, resolving
// links against the page URL (or its <base href>) and the crawl origin.
func ApplyMetadata(snap *PageSnapshot, meta *extract.Metadata, pageURL, origin *url.URL) {
	snap.Title = meta.Title
	snap.MetaDesc = meta.MetaDesc
	snap.Canonical = meta.Canonical
	snap.Canonicals = meta.Canonicals
	snap.RobotsMeta = meta.RobotsMeta
	snap.Viewport = meta.Viewport
	snap.Lang = meta.Lang
	snap.Hreflang = meta.Hreflang
	snap.JSONLD = meta.JSONLD
	snap.OpenGraph = meta.OpenGraph
	snap.TwitterCard = meta.TwitterCard
	snap.H1 = meta.H1
	snap.Headings = meta.Headings
	snap.TextLength = meta.TextLength
	snap.WordCount = meta.WordCount

	for _, img := range meta.Images {
		snap.Images = append(snap.Images, ImageInfo{Src: img.Src, Alt: img.Alt, HasAlt: img.HasAlt})
	}

	linkBase := pageURL
	if meta.BaseHref != "" {
		if b, err := pageURL.Parse(meta.BaseHref); err == nil {
			linkBase = b
		}
	}

	seenInt := map[string]bool{}
	seenExt := map[string]bool{}
	for _, l := range meta.Links {
		info := LinkInfo{Raw: l.Href, Text: l.Text, Nofollow: l.Nofollow, HasHref: l.HasHref}
		if l.HasHref && l.Href != "" && !strings.HasPrefix(l.Href, "#") {
			if abs, err := NormalizeURL(l.Href, linkBase); err == nil {
				info.URL = abs
				if u, err := url.Parse(abs); err == nil && IsSameOrigin(u, origin) {
					info.Internal = true
					if !seenInt[abs] {
						seenInt[abs] = true
						snap.InternalLinks = append(snap.InternalLinks, abs)
					}
				} else if !seenExt[abs] {
					seenExt[abs] = true
					snap.ExternalLinks = append(snap.ExternalLinks, abs)
				}
			}
		}
		snap.Links = append(snap.Links, info)
	}
}

// compilePattern accepts a regular expression; anything that does not compile
// is treated as a literal substring.
func compilePattern(p string) (*regexp.Regexp, error) {
	if re, err := regexp.Compile(p); err == nil {
		return re, nil
	}
	return regexp.Compile(regexp.QuoteMeta(p))
}
