package rules

import (
	"testing"

	"github.com/zimkk/rankcore/internal/crawl"
)

func TestH1StructureRule(t *testing.T) {
	rule := &H1StructureRule{}

	// Missing H1
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com/no-h1",
		StatusCode: 200,
		H1:         nil,
	}
	f1 := rule.Evaluate(snap1)
	if f1 == nil || f1.ID != "RC-META-005" || f1.Severity != "medium" {
		t.Errorf("expected medium finding for missing H1, got %+v", f1)
	}

	// Multiple H1
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com/multi-h1",
		StatusCode: 200,
		H1:         []string{"First Title", "Second Title"},
	}
	f2 := rule.Evaluate(snap2)
	if f2 == nil || f2.Severity != "low" {
		t.Errorf("expected low finding for multiple H1, got %+v", f2)
	}

	// Exactly one H1
	snap3 := &crawl.PageSnapshot{
		URL:        "https://example.com/good",
		StatusCode: 200,
		H1:         []string{"Just One Heading"},
	}
	if f3 := rule.Evaluate(snap3); f3 != nil {
		t.Errorf("expected nil finding for valid single H1, got %+v", f3)
	}
}

func TestTitleQualityRule(t *testing.T) {
	rule := &TitleQualityRule{}

	// Generic title
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
		Title:      "Home",
	}
	f1 := rule.Evaluate(snap1)
	if f1 == nil || f1.Severity != "high" {
		t.Errorf("expected high finding for generic title, got %+v", f1)
	}

	// Too long title
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com/long",
		StatusCode: 200,
		Title:      "This is an extraordinarily long page title that exceeds seventy characters in total length for testing",
	}
	f2 := rule.Evaluate(snap2)
	if f2 == nil || f2.Severity != "low" {
		t.Errorf("expected low finding for overlong title, got %+v", f2)
	}

	// Good title
	snap3 := &crawl.PageSnapshot{
		URL:        "https://example.com/good",
		StatusCode: 200,
		Title:      "RankCore SEO Engine Documentation",
	}
	if f3 := rule.Evaluate(snap3); f3 != nil {
		t.Errorf("expected nil finding for good title, got %+v", f3)
	}
}

func TestHTMLLangRule(t *testing.T) {
	rule := &HTMLLangRule{}

	// Missing lang
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
		Lang:       "",
	}
	if f1 := rule.Evaluate(snap1); f1 == nil || f1.ID != "RC-LANG-001" {
		t.Errorf("expected finding for missing lang attribute, got %+v", f1)
	}

	// Valid lang
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
		Lang:       "en-US",
	}
	if f2 := rule.Evaluate(snap2); f2 != nil {
		t.Errorf("expected no finding for valid lang, got %+v", f2)
	}
}

func TestOpenGraphRule(t *testing.T) {
	rule := &OpenGraphRule{}

	// Missing OG
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
		OpenGraph:  nil,
	}
	if f1 := rule.Evaluate(snap1); f1 == nil || f1.ID != "RC-SOCIAL-001" {
		t.Errorf("expected finding for missing OG tags, got %+v", f1)
	}

	// Present OG
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
		OpenGraph: map[string]string{
			"og:title":       "Example Title",
			"og:image":       "https://example.com/og.jpg",
			"og:description": "Example description",
		},
	}
	if f2 := rule.Evaluate(snap2); f2 != nil {
		t.Errorf("expected no finding when OG tags present, got %+v", f2)
	}
}

func TestRedirectChainRule(t *testing.T) {
	rule := &RedirectChainRule{}

	// Loop
	snap1 := &crawl.PageSnapshot{
		URL:          "https://example.com/loop",
		RedirectLoop: true,
	}
	f1 := rule.Evaluate(snap1)
	if f1 == nil || f1.Severity != "critical" {
		t.Errorf("expected critical finding for redirect loop, got %+v", f1)
	}

	// Long chain (>2 hops)
	snap2 := &crawl.PageSnapshot{
		URL:           "https://example.com/chain",
		RedirectChain: []string{"https://example.com/a", "https://example.com/b", "https://example.com/c"},
	}
	f2 := rule.Evaluate(snap2)
	if f2 == nil || f2.Severity != "medium" {
		t.Errorf("expected medium finding for long chain, got %+v", f2)
	}
}

func TestCanonicalProtocolRule(t *testing.T) {
	rule := &CanonicalProtocolRule{}

	// Insecure canonical on secure page
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com/secure",
		StatusCode: 200,
		Canonical:  "http://example.com/secure",
	}
	f1 := rule.Evaluate(snap1)
	if f1 == nil || f1.Severity != "high" {
		t.Errorf("expected high finding for insecure canonical on https, got %+v", f1)
	}

	// Valid matching canonical
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com/secure",
		StatusCode: 200,
		Canonical:  "https://example.com/secure",
	}
	if f2 := rule.Evaluate(snap2); f2 != nil {
		t.Errorf("expected nil finding for matching canonical, got %+v", f2)
	}
}
