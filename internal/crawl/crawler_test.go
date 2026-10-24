package crawl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zimkk/rankcore/internal/robots"
)

// chainSite serves /p0 … /pN where each page links to the next, using a mix of
// relative, root-relative and absolute same-origin links.
func chainSite(t *testing.T, n int) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var i int
		if _, err := fmt.Sscanf(r.URL.Path, "/p%d", &i); err != nil || i > n {
			http.NotFound(w, r)
			return
		}
		next := ""
		switch i % 3 {
		case 0:
			next = fmt.Sprintf(`<a href="p%d">next</a>`, i+1)
		case 1:
			next = fmt.Sprintf(`<a href="/p%d">next</a>`, i+1)
		default:
			next = fmt.Sprintf(`<a href="%s/p%d">next</a>`, srv.URL, i+1)
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body>%s <a href="https://external.example/x">ext</a> <a href="//cdn.example/y">cdn</a></body></html>`, next)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestCrawler(cfg Config) *Crawler {
	cfg.AllowPrivate = true
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	return NewCrawler(cfg)
}

func TestCrawlerTerminatesAndFollowsAllLinkForms(t *testing.T) {
	srv := chainSite(t, 6)
	c := newTestCrawler(Config{MaxPages: 100, MaxDepth: 20, Concurrency: 4})

	done := make(chan []*PageSnapshot)
	go func() {
		pages, err := c.Run(context.Background(), []Seed{{URL: srv.URL + "/p0", Source: "seed"}})
		if err != nil {
			t.Error(err)
		}
		done <- pages
	}()

	select {
	case pages := <-done:
		// p0..p6 exist, p7 is linked and 404s.
		if len(pages) != 8 {
			t.Fatalf("expected 8 URLs, got %d", len(pages))
		}
		for _, p := range pages {
			if strings.Contains(p.URL, "external.example") || strings.Contains(p.URL, "cdn.example") {
				t.Errorf("crawled off-origin URL %s", p.URL)
			}
		}
		if pages[0].ExternalLinks == nil || len(pages[0].ExternalLinks) != 2 {
			t.Errorf("expected absolute and protocol-relative links to be external: %v", pages[0].ExternalLinks)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("crawler did not terminate")
	}
}

func TestCrawlerRespectsLimits(t *testing.T) {
	srv := chainSite(t, 50)

	c := newTestCrawler(Config{MaxPages: 5, MaxDepth: 20, Concurrency: 2})
	pages, _ := c.Run(context.Background(), []Seed{{URL: srv.URL + "/p0"}})
	if len(pages) != 5 || !c.LimitReached() {
		t.Errorf("max pages: got %d pages, limitReached=%v", len(pages), c.LimitReached())
	}

	c = newTestCrawler(Config{MaxPages: 100, MaxDepth: 3, Concurrency: 2})
	pages, _ = c.Run(context.Background(), []Seed{{URL: srv.URL + "/p0"}})
	if len(pages) != 4 {
		t.Errorf("max depth 3 should crawl p0..p3, got %d", len(pages))
	}

	c = newTestCrawler(Config{MaxPages: 100, MaxDepth: 20, Concurrency: 2, ExcludePattern: `/p[3-9]$`})
	pages, _ = c.Run(context.Background(), []Seed{{URL: srv.URL + "/p0"}})
	if len(pages) != 3 {
		t.Errorf("exclude should stop at p2, got %d", len(pages))
	}
}

func TestCrawlerRespectsRobotsForItsOwnAgent(t *testing.T) {
	srv := chainSite(t, 3)
	c := newTestCrawler(Config{MaxPages: 10, MaxDepth: 10, Concurrency: 1, RespectRobots: true, UserAgent: "RankCoreBot/1.0"})
	c.Robots = robots.Parse(strings.NewReader("User-agent: *\nDisallow: /p2\n"))
	pages, _ := c.Run(context.Background(), []Seed{{URL: srv.URL + "/p0"}})
	var p2 *PageSnapshot
	for _, p := range pages {
		if strings.HasSuffix(p.URL, "/p2") {
			p2 = p
		}
	}
	if p2 == nil || p2.Skipped != "robots" || p2.StatusCode != 0 {
		t.Fatalf("expected /p2 to be skipped for robots, got %+v", p2)
	}
	if len(pages) != 3 {
		t.Errorf("crawl should stop at the disallowed page, got %d pages", len(pages))
	}
}

func TestCrawlerCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	newTestCrawler(Config{MaxPages: 10, Concurrency: 2}).Run(ctx, []Seed{{URL: srv.URL}})
	if time.Since(start) > 1500*time.Millisecond {
		t.Errorf("cancellation took %s", time.Since(start))
	}
}

func TestNoindexDirectives(t *testing.T) {
	cases := map[string]bool{
		"noindex":                     true,
		"NOINDEX, nofollow":           true,
		"none":                        true,
		"googlebot: noindex":          true,
		"bingbot: noindex":            false,
		"index, follow":               false,
		"max-snippet: 50, noarchive":  false,
		"unavailable_after: 2030-1-1": false,
	}
	for v, want := range cases {
		p := PageSnapshot{XRobotsTag: v}
		if got := p.Noindex(); got != want {
			t.Errorf("Noindex(%q) = %v, want %v", v, got, want)
		}
	}
}
