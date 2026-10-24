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
	"more info":  true,
	"details":    true,
}

// RC-LINK-002: Links with empty or generic anchor text (Lighthouse: link-text)
type LinkTextRule struct{}

func (r *LinkTextRule) ID() string   { return "RC-LINK-002" }
func (r *LinkTextRule) Version() int { return 2 }

func (r *LinkTextRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(snapshot) || len(snapshot.Links) == 0 {
		return nil
	}

	var badLinks []string
	seen := map[string]bool{}
	for _, l := range snapshot.Links {
		if !l.HasHref || isUncrawlable(l.Raw) {
			continue // reported by RC-LINK-003
		}
		normalized := strings.ToLower(strings.Trim(strings.TrimSpace(l.Text), ".…»›→ "))
		if normalized == "" || genericLinkTexts[normalized] {
			if !seen[l.Raw] {
				seen[l.Raw] = true
				badLinks = append(badLinks, l.Raw)
			}
		}
	}

	if len(badLinks) == 0 {
		return nil
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
			"sample_links": capList(badLinks, 10),
			"total_bad":    len(badLinks),
		},
	}
}

// RC-LINK-003: Uncrawlable anchors (Lighthouse: crawlable-anchors)
type CrawlableAnchorRule struct{}

func (r *CrawlableAnchorRule) ID() string   { return "RC-LINK-003" }
func (r *CrawlableAnchorRule) Version() int { return 2 }

func (r *CrawlableAnchorRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(snapshot) {
		return nil
	}

	var uncrawlable []string
	for _, l := range snapshot.Links {
		switch {
		case !l.HasHref:
			uncrawlable = append(uncrawlable, "(no href, onclick handler)")
		case isUncrawlable(l.Raw):
			uncrawlable = append(uncrawlable, l.Raw)
		}
	}
	// Older snapshots (and hand-built ones) only carry link lists.
	if len(snapshot.Links) == 0 {
		for _, href := range append(append([]string{}, snapshot.InternalLinks...), snapshot.ExternalLinks...) {
			if isUncrawlable(strings.TrimSpace(href)) {
				uncrawlable = append(uncrawlable, href)
			}
		}
	}

	if len(uncrawlable) == 0 {
		return nil
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "discovery",
		Severity:    "medium",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Page contains links that are not crawlable",
		Remediation: "Ensure navigation links use proper <a href=\"/path\"> elements instead of javascript: URIs, empty fragments, or click handlers.",
		Evidence: map[string]interface{}{
			"uncrawlable_links": capList(uncrawlable, 10),
			"total":             len(uncrawlable),
		},
	}
}

// isUncrawlable returns true if an href cannot be followed by a crawler.
func isUncrawlable(href string) bool {
	href = strings.TrimSpace(href)
	if href == "" || href == "#" || href == "#!" {
		return true
	}
	lower := strings.ToLower(href)
	return strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "void(")
}
