package rules

import (
	"fmt"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// HTMLLangRule verifies that the <html> element defines a valid language code.
type HTMLLangRule struct{}

func (r *HTMLLangRule) ID() string {
	return "RC-LANG-001"
}

func (r *HTMLLangRule) Version() int {
	return 1
}

// Common 2-letter ISO 639-1 language codes
var validLangPrefixes = map[string]bool{
	"en": true, "es": true, "fr": true, "de": true, "it": true, "pt": true,
	"ru": true, "zh": true, "ja": true, "ko": true, "ar": true, "hi": true,
	"nl": true, "sv": true, "no": true, "da": true, "fi": true, "pl": true,
	"tr": true, "el": true, "he": true, "id": true, "th": true, "vi": true,
	"cs": true, "hu": true, "ro": true, "uk": true, "ms": true,
}

func (r *HTMLLangRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 || snapshot.Noindex() {
		return nil
	}

	lang := strings.TrimSpace(snapshot.Lang)
	if lang == "" {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "internationalization",
			Severity:    "medium",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Document does not have a [lang] attribute on the <html> element",
			Evidence: map[string]interface{}{
				"lang": "",
			},
			Remediation: "Add a valid [lang] attribute to the <html> tag, e.g. <html lang=\"en\">.",
		}
	}

	// Validate language code
	parts := strings.Split(lang, "-")
	primary := strings.ToLower(parts[0])
	if len(primary) < 2 || len(primary) > 3 || (!validLangPrefixes[primary] && len(primary) != 2) {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "internationalization",
			Severity:    "medium",
			Confidence:  0.9,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("<html> element has unrecognized or invalid lang attribute: %q", lang),
			Evidence: map[string]interface{}{
				"lang": lang,
			},
			Remediation: "Use an IETF BCP 47 compliant language tag (such as 'en' or 'en-US') on <html lang=\"...\">.",
		}
	}

	return nil
}
