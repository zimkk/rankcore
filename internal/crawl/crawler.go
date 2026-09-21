package crawl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"rankcore/internal/httpx"
	"rankcore/internal/extract"
)

// Crawler manages the URL frontier and concurrent execution.
type Crawler struct {
	Config   Config
	Visited  map[string]bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	Queue    chan string
	Results  chan *PageSnapshot
	BaseHost string
}

// NewCrawler initializes a Crawler.
func NewCrawler(cfg Config) *Crawler {
	return &Crawler{
		Config:  cfg,
		Visited: make(map[string]bool),
		Queue:   make(chan string, cfg.MaxPages*2), // Buffer for some burstiness
		Results: make(chan *PageSnapshot, cfg.MaxPages),
	}
}

// Start begins crawling from the seed URL.
func (c *Crawler) Start(ctx context.Context, seed string) {
	parsedSeed, err := url.Parse(seed)
	if err != nil {
		fmt.Printf("Crawler: Invalid seed URL: %v\n", err)
		close(c.Results)
		return
	}
	
	c.BaseHost = parsedSeed.Host

	c.Queue <- seed
	c.Visited[seed] = true

	// Allow private IPs if seed is localhost
	allowPrivate := parsedSeed.Hostname() == "localhost" || parsedSeed.Hostname() == "127.0.0.1"
	client := httpx.NewSafeClient(c.Config.Timeout, c.Config.UserAgent, allowPrivate)

	concurrency := c.Config.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	for i := 0; i < concurrency; i++ {
		c.wg.Add(1)
		go c.worker(ctx, client, parsedSeed)
	}

	// This go routine will wait and then close channels when everything finishes.
	// But since the queue might drain before wg is done, we have to carefully manage idle state.
	// For simplicity in this end-to-end version, we just let it run.
	go func() {
		c.wg.Wait()
		close(c.Results)
	}()
}

func (c *Crawler) worker(ctx context.Context, client *http.Client, base *url.URL) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case targetURL, ok := <-c.Queue:
			if !ok {
				return
			}
			
			c.mu.Lock()
			if len(c.Visited) > c.Config.MaxPages {
				c.mu.Unlock()
				continue
			}
			c.mu.Unlock()

			snap := c.fetchAndExtract(ctx, client, targetURL, base)
			c.Results <- snap

			// Enqueue same-origin internal links
			if snap != nil {
				for _, link := range snap.InternalLinks {
					norm, err := NormalizeURL(link, base)
					if err == nil {
						parsed, err := url.Parse(norm)
						if err == nil && IsSameOrigin(parsed, base) {
							c.mu.Lock()
							if !c.Visited[norm] && len(c.Visited) <= c.Config.MaxPages {
								c.Visited[norm] = true
								c.Queue <- norm
							}
							c.mu.Unlock()
						}
					}
				}
			}
		}
	}
}

func (c *Crawler) fetchAndExtract(ctx context.Context, client *http.Client, targetURL string, base *url.URL) *PageSnapshot {
	resp, err := httpx.Request(ctx, client, targetURL, c.Config.UserAgent)
	
	snap := &PageSnapshot{
		URL:      targetURL,
		FinalURL: targetURL,
	}

	if err != nil {
		// Store error as status 0 or record it in some way
		return snap
	}
	defer resp.Body.Close()

	snap.FinalURL = resp.Request.URL.String()
	snap.StatusCode = resp.StatusCode
	snap.ContentType = resp.Header.Get("Content-Type")
	snap.Headers = resp.Header

	// Read up to a limit (e.g., 2MB)
	bodyReader := io.LimitReader(resp.Body, 2*1024*1024)
	
	// Use our extractor
	meta, _ := extract.Extract(bodyReader)
	if meta != nil {
		snap.Title = meta.Title
		snap.MetaDesc = meta.MetaDesc
		snap.Canonical = meta.Canonical
		snap.InternalLinks = meta.InternalLinks
		snap.Hreflang = meta.Hreflang
		snap.JSONLD = meta.JSONLD
	}

	return snap
}
