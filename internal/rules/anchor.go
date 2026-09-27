package rules

import (
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// genericLinkTexts is the set of anchor texts that provide no navigation value.
var genericLinkTexts = map[string]bool{
	"click here": true,
	"here":       true,
	"read more":  true,
	"learn more": true,
	"more":       true,
	"link":       true,
	"this":       true,
	"go":         true,
	"click":      true,
	"this link":  true,
	"this page":  true,
}

// RC-LINK-002: Links with empty or generic anchor text (Lighthouse: link-text)
type LinkTextRule struct{}

func (r *LinkTextRule) ID() string   { return "RC-LINK-002" }
func (r *LinkTextRule) Version() int { return 1 }

func (r *LinkTextRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 || len(snapshot.AnchorTexts) == 0 {
		return nil
	}

	var badLinks []string
	for href, text := range snapshot.AnchorTexts {
		normalized := strings.ToLower(strings.TrimSpace(text))
		if normalized == "" || genericLinkTexts[normalized] {
			badLinks = append(badLinks, href)
		}
	}

	if len(badLinks) == 0 {
		return nil
	}

	shown := badLinks
	if len(shown) > 10 {
		shown = shown[:10]
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "discovery",
		Severity:    "medium",
		Confidence:  0.9,
		URL:         snapshot.URL,
		Summary:     "Links use non-descriptive anchor text",
		Remediation: "Replace generic link text (\"click here\", \"read more\") with descriptive phrases that indicate the link destination.",
		Evidence: map[string]interface{}{
			"sample_links": shown,
			"total_bad":    len(badLinks),
		},
	}
}

// RC-LINK-003: Uncrawlable anchors (Lighthouse: crawlable-anchors)
type CrawlableAnchorRule struct{}

func (r *CrawlableAnchorRule) ID() string   { return "RC-LINK-003" }
func (r *CrawlableAnchorRule) Version() int { return 1 }

func (r *CrawlableAnchorRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}

	var uncrawlable []string
	allLinks := append(snapshot.InternalLinks, snapshot.ExternalLinks...)
	for _, href := range allLinks {
		trimmed := strings.TrimSpace(href)
		if isUncrawlable(trimmed) {
			uncrawlable = append(uncrawlable, trimmed)
		}
	}

	if len(uncrawlable) == 0 {
		return nil
	}

	shown := uncrawlable
	if len(shown) > 10 {
		shown = shown[:10]
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "discovery",
		Severity:    "medium",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Page contains links that are not crawlable",
		Remediation: "Ensure navigation links use proper <a href=\"/path\"> elements instead of javascript: URIs, empty fragments, or void hrefs.",
		Evidence: map[string]interface{}{
			"uncrawlable_links": shown,
			"total":             len(uncrawlable),
		},
	}
}

// isUncrawlable returns true if an href cannot be followed by a crawler.
func isUncrawlable(href string) bool {
	if href == "" || href == "#" {
		return true
	}
	lower := strings.ToLower(href)
	if strings.HasPrefix(lower, "javascript:") {
		return true
	}
	if strings.HasPrefix(lower, "void(") {
		return true
	}
	if strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") {
		return false // valid but not navigation — not flagged as uncrawlable
	}
	return false
}
