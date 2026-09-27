package rules

import (
	"testing"

	"github.com/zimkk/rankcore/internal/crawl"
)

func TestMissingTitleRule(t *testing.T) {
	rule := &MissingTitleRule{}

	t.Run("fires on empty title", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, Title: ""}
		f := rule.Evaluate(snap)
		if f == nil {
			t.Fatal("expected finding for empty title")
		}
		if f.ID != "RC-META-001" {
			t.Errorf("expected RC-META-001, got %s", f.ID)
		}
	})

	t.Run("skips non-200", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 404, Title: ""}
		if rule.Evaluate(snap) != nil {
			t.Error("should not fire on non-200 pages")
		}
	})

	t.Run("passes with title", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, Title: "Home Page"}
		if rule.Evaluate(snap) != nil {
			t.Error("should not fire when title is present")
		}
	})
}

func TestMissingMetaDescRule(t *testing.T) {
	rule := &MissingMetaDescRule{}

	snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, MetaDesc: ""}
	if rule.Evaluate(snap) == nil {
		t.Error("expected finding for missing meta description")
	}

	snap.MetaDesc = "A great website about things"
	if rule.Evaluate(snap) != nil {
		t.Error("should not fire when meta desc is present")
	}
}

func TestMissingViewportRule(t *testing.T) {
	rule := &MissingViewportRule{}

	t.Run("fires on missing viewport", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, Viewport: ""}
		f := rule.Evaluate(snap)
		if f == nil {
			t.Fatal("expected finding")
		}
		if f.Severity != "high" {
			t.Errorf("expected high severity, got %s", f.Severity)
		}
	})

	t.Run("fires on viewport without width", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, Viewport: "initial-scale=1"}
		f := rule.Evaluate(snap)
		if f == nil {
			t.Fatal("expected finding for viewport without width")
		}
		if f.Severity != "medium" {
			t.Errorf("expected medium severity, got %s", f.Severity)
		}
	})

	t.Run("passes with proper viewport", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200, Viewport: "width=device-width, initial-scale=1"}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with proper viewport")
		}
	})
}

func TestImageAltRule(t *testing.T) {
	rule := &ImageAltRule{}

	t.Run("fires when images lack alt", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			Images: []crawl.ImageInfo{
				{Src: "/logo.png", Alt: "Logo", HasAlt: true},
				{Src: "/hero.jpg", HasAlt: false},
				{Src: "/bg.png", HasAlt: false},
			},
		}
		f := rule.Evaluate(snap)
		if f == nil {
			t.Fatal("expected finding")
		}
		if f.Evidence["total_missing"].(int) != 2 {
			t.Errorf("expected 2 missing, got %v", f.Evidence["total_missing"])
		}
	})

	t.Run("passes when all images have alt", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			Images: []crawl.ImageInfo{
				{Src: "/logo.png", Alt: "Logo", HasAlt: true},
			},
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass when all images have alt")
		}
	})

	t.Run("skips pages with no images", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200}
		if rule.Evaluate(snap) != nil {
			t.Error("should skip pages with no images")
		}
	})
}

func TestLinkTextRule(t *testing.T) {
	rule := &LinkTextRule{}

	t.Run("fires on generic anchor text", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			AnchorTexts: map[string]string{
				"/about":   "click here",
				"/pricing": "Learn about our pricing",
			},
		}
		f := rule.Evaluate(snap)
		if f == nil {
			t.Fatal("expected finding for generic anchor text")
		}
	})

	t.Run("fires on empty anchor text", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			AnchorTexts: map[string]string{
				"/contact": "",
			},
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for empty anchor text")
		}
	})

	t.Run("passes with descriptive text", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			AnchorTexts: map[string]string{
				"/about":   "About our company",
				"/pricing": "View pricing plans",
			},
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with descriptive anchor text")
		}
	})
}

func TestCrawlableAnchorRule(t *testing.T) {
	rule := &CrawlableAnchorRule{}

	t.Run("fires on javascript: hrefs", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			InternalLinks: []string{"javascript:void(0)", "/about"},
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for javascript: href")
		}
	})

	t.Run("fires on empty fragment", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			InternalLinks: []string{"#"},
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for # href")
		}
	})

	t.Run("passes with real links", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			InternalLinks: []string{"/about", "/contact"},
			ExternalLinks: []string{"https://other.com"},
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with crawlable links")
		}
	})
}

func TestHTTPSuccessRule(t *testing.T) {
	rule := &HTTPSuccessRule{}

	t.Run("fires on status 0 (connection error)", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 0}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for status 0")
		}
	})

	t.Run("fires on 301 redirect", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com/old", StatusCode: 301, FinalURL: "https://example.com/new"}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for 301")
		}
	})

	t.Run("passes on 200", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass on 200")
		}
	})

	t.Run("skips 4xx/5xx (handled by other rules)", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 500}
		if rule.Evaluate(snap) != nil {
			t.Error("should not fire on 5xx (handled by HTTP5xxRule)")
		}
	})
}

func TestRobotsCrawlableRule(t *testing.T) {
	rule := &RobotsCrawlableRule{}

	t.Run("fires when blocked", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com/private", BlockedByRobots: true, RobotsRule: "Disallow: /private",
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for blocked page")
		}
	})

	t.Run("passes when allowed", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", BlockedByRobots: false}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass when not blocked")
		}
	})
}

func TestMissingCanonicalRule(t *testing.T) {
	rule := &MissingCanonicalRule{}

	t.Run("fires on HTML page without canonical", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			ContentType: "text/html; charset=utf-8", Canonical: "",
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for missing canonical")
		}
	})

	t.Run("passes with canonical", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			ContentType: "text/html", Canonical: "https://example.com",
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with canonical present")
		}
	})

	t.Run("skips non-HTML", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com/api", StatusCode: 200,
			ContentType: "application/json",
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should skip non-HTML content types")
		}
	})
}

func TestHreflangValidRule(t *testing.T) {
	rule := &HreflangValidRule{}

	t.Run("fires on invalid language code", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			Hreflang: map[string]string{
				"en":         "https://example.com",
				"not-a-lang": "https://example.com/oops",
			},
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for invalid hreflang")
		}
	})

	t.Run("passes with valid codes", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			Hreflang: map[string]string{
				"en":        "https://example.com",
				"fr":        "https://example.com/fr",
				"x-default": "https://example.com",
			},
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with valid language codes")
		}
	})

	t.Run("skips pages without hreflang", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200}
		if rule.Evaluate(snap) != nil {
			t.Error("should skip pages without hreflang")
		}
	})
}

func TestStructuredDataValidRule(t *testing.T) {
	rule := &StructuredDataValidRule{}

	t.Run("fires on invalid JSON-LD", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			JSONLD: []string{`{"@context": "https://schema.org", "invalid`},
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for invalid JSON")
		}
	})

	t.Run("passes with valid JSON-LD", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			JSONLD: []string{`{"@context": "https://schema.org", "@type": "WebSite"}`},
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with valid JSON-LD")
		}
	})

	t.Run("skips pages without JSON-LD", func(t *testing.T) {
		snap := &crawl.PageSnapshot{URL: "https://example.com", StatusCode: 200}
		if rule.Evaluate(snap) != nil {
			t.Error("should skip pages without JSON-LD")
		}
	})
}

func TestThinContentRule(t *testing.T) {
	rule := &ThinContentRule{}

	t.Run("fires on very short text", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			ContentType: "text/html", TextLength: 50,
		}
		if rule.Evaluate(snap) == nil {
			t.Error("expected finding for thin content")
		}
	})

	t.Run("passes with enough text", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com", StatusCode: 200,
			ContentType: "text/html", TextLength: 500,
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should pass with sufficient text")
		}
	})

	t.Run("skips non-HTML", func(t *testing.T) {
		snap := &crawl.PageSnapshot{
			URL: "https://example.com/data.json", StatusCode: 200,
			ContentType: "application/json", TextLength: 10,
		}
		if rule.Evaluate(snap) != nil {
			t.Error("should skip non-HTML")
		}
	})
}
