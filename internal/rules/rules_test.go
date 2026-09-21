package rules

import (
	"rankcore/internal/crawl"
	"testing"
)

func TestHTTP5xxRule(t *testing.T) {
	rule := &HTTP5xxRule{}
	
	snap1 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 500,
	}
	
	finding1 := rule.Evaluate(snap1)
	if finding1 == nil {
		t.Error("Expected finding for 500 status")
	} else if finding1.Severity != "critical" {
		t.Errorf("Expected critical severity, got %s", finding1.Severity)
	}
	
	snap2 := &crawl.PageSnapshot{
		URL:        "https://example.com",
		StatusCode: 200,
	}
	
	finding2 := rule.Evaluate(snap2)
	if finding2 != nil {
		t.Error("Expected no finding for 200 status")
	}
}

func TestMetaNoindexRule(t *testing.T) {
	rule := &MetaNoindexRule{}
	
	snap1 := &crawl.PageSnapshot{
		URL: "https://example.com",
		Headers: map[string][]string{
			"X-Robots-Tag": {"noindex, nofollow"},
		},
	}
	
	finding1 := rule.Evaluate(snap1)
	if finding1 == nil {
		t.Error("Expected finding for noindex header")
	}
	
	snap2 := &crawl.PageSnapshot{
		URL: "https://example.com",
		Headers: map[string][]string{
			"X-Robots-Tag": {"index, follow"},
		},
	}
	
	finding2 := rule.Evaluate(snap2)
	if finding2 != nil {
		t.Error("Expected no finding for index header")
	}
}
