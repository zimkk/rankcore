package rules

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// CanonicalProtocolRule checks if an HTTPS page declares an insecure HTTP canonical target or mismatched host.
type CanonicalProtocolRule struct{}

func (r *CanonicalProtocolRule) ID() string {
	return "RC-CANON-003"
}

func (r *CanonicalProtocolRule) Version() int {
	return 1
}

func (r *CanonicalProtocolRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.Canonical == "" || snapshot.StatusCode != 200 {
		return nil
	}

	pageParsed, err := url.Parse(snapshot.URL)
	if err != nil {
		return nil
	}

	canonParsed, err := url.Parse(snapshot.Canonical)
	if err != nil {
		return nil
	}

	// 1. Insecure HTTP canonical on HTTPS page
	if strings.EqualFold(pageParsed.Scheme, "https") && strings.EqualFold(canonParsed.Scheme, "http") {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "canonical",
			Severity:    "high",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Secure HTTPS page designates an insecure HTTP canonical target",
			Evidence: map[string]interface{}{
				"page_url":      snapshot.URL,
				"canonical_url": snapshot.Canonical,
			},
			Remediation: "Update the canonical tag to use https:// to match the secure protocol.",
		}
	}

	// 2. Cross-domain canonical (different host)
	if canonParsed.Host != "" && !strings.EqualFold(canonParsed.Host, pageParsed.Host) {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "canonical",
			Severity:    "medium",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Cross-domain canonical target: points from %s to %s", pageParsed.Host, canonParsed.Host),
			Evidence: map[string]interface{}{
				"page_host":      pageParsed.Host,
				"canonical_host": canonParsed.Host,
				"canonical_url":  snapshot.Canonical,
			},
			Remediation: "Verify whether cross-domain canonicalization is intentional. For internal site pages, use the same authoritative domain.",
		}
	}

	return nil
}
