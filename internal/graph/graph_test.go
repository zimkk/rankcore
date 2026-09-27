package graph

import (
	"sort"
	"testing"
)

func TestOrphans(t *testing.T) {
	g := New()
	g.AddPage("https://example.com", []string{"https://example.com/about", "https://example.com/contact"})
	g.AddPage("https://example.com/about", []string{"https://example.com"})
	g.AddPage("https://example.com/contact", []string{})
	g.AddPage("https://example.com/orphan", []string{}) // no inbound links

	orphans := g.Orphans("https://example.com")
	if len(orphans) != 1 || orphans[0] != "https://example.com/orphan" {
		t.Errorf("expected orphan page, got %v", orphans)
	}
}

func TestDepth(t *testing.T) {
	g := New()
	g.AddPage("https://example.com", []string{"https://example.com/a"})
	g.AddPage("https://example.com/a", []string{"https://example.com/b"})
	g.AddPage("https://example.com/b", []string{"https://example.com/c"})
	g.AddPage("https://example.com/c", []string{})

	depths := g.Depth("https://example.com")

	expected := map[string]int{
		"https://example.com":   0,
		"https://example.com/a": 1,
		"https://example.com/b": 2,
		"https://example.com/c": 3,
	}
	for url, want := range expected {
		if got, ok := depths[url]; !ok || got != want {
			t.Errorf("Depth(%q) = %d, want %d", url, got, want)
		}
	}
}

func TestInboundOutboundCount(t *testing.T) {
	g := New()
	g.AddPage("https://a.com", []string{"https://a.com/b", "https://a.com/c"})
	g.AddPage("https://a.com/b", []string{"https://a.com/c"})
	g.AddPage("https://a.com/c", []string{})

	if got := g.OutboundCount("https://a.com"); got != 2 {
		t.Errorf("outbound for root = %d, want 2", got)
	}
	if got := g.InboundCount("https://a.com/c"); got != 2 {
		t.Errorf("inbound for /c = %d, want 2", got)
	}
}

func TestPageCount(t *testing.T) {
	g := New()
	g.AddPage("https://a.com", nil)
	g.AddPage("https://a.com/b", nil)
	if g.PageCount() != 2 {
		t.Errorf("expected 2 pages, got %d", g.PageCount())
	}
}

func TestPages(t *testing.T) {
	g := New()
	g.AddPage("https://a.com", nil)
	g.AddPage("https://a.com/b", nil)

	pages := g.Pages()
	sort.Strings(pages)
	if len(pages) != 2 || pages[0] != "https://a.com" || pages[1] != "https://a.com/b" {
		t.Errorf("unexpected pages: %v", pages)
	}
}
