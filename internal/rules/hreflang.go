package rules

import (
	"regexp"
	"sort"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// validLangCode matches BCP 47 language tags (basic form: xx or xx-YY).
var validLangCode = regexp.MustCompile(`^(?i)[a-z]{2,3}(-[a-zA-Z]{2,4})?$`)

// RC-HREFLANG-001: Invalid hreflang annotations (Lighthouse: hreflang)
type HreflangValidRule struct{}

func (r *HreflangValidRule) ID() string   { return "RC-HREFLANG-001" }
func (r *HreflangValidRule) Version() int { return 2 }

func (r *HreflangValidRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(snapshot) {
		return nil
	}
	if len(snapshot.Hreflang) == 0 {
		return nil
	}

	var invalid []string
	for lang, href := range snapshot.Hreflang {
		// x-default is a special valid value
		if strings.EqualFold(lang, "x-default") {
			if href != "" && !isAbsoluteURL(href) {
				invalid = append(invalid, lang+" (relative href: "+href+")")
			}
			continue
		}
		if !validLangCode.MatchString(lang) {
			invalid = append(invalid, lang)
			continue
		}
		switch {
		case href == "":
			invalid = append(invalid, lang+" (empty href)")
		case !isAbsoluteURL(href):
			invalid = append(invalid, lang+" (relative href: "+href+")")
		}
	}

	if len(invalid) == 0 {
		return nil
	}

	sort.Strings(invalid)

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "international",
		Severity:    "medium",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Hreflang annotations contain invalid language codes or empty targets",
		Remediation: "Use valid BCP 47 language-region codes (e.g. 'en', 'en-US', 'x-default') and ensure all hreflang links have valid href values.",
		Evidence: map[string]interface{}{
			"invalid_entries": invalid,
		},
	}
}
