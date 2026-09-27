package rules

import (
	"fmt"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RedirectChainRule checks for redirect loops or excessive hops (> 2 redirects).
type RedirectChainRule struct{}

func (r *RedirectChainRule) ID() string {
	return "RC-HTTP-005"
}

func (r *RedirectChainRule) Version() int {
	return 1
}

func (r *RedirectChainRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.RedirectLoop {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "crawlability",
			Severity:    "critical",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     "Redirect loop detected while requesting URL",
			Evidence: map[string]interface{}{
				"chain": snapshot.RedirectChain,
			},
			Remediation: "Resolve the cyclic redirect configuration on the server to prevent crawlers and users from being trapped.",
		}
	}

	if len(snapshot.RedirectChain) > 2 {
		return &report.Finding{
			ID:          r.ID(),
			RuleVersion: r.Version(),
			Category:    "crawlability",
			Severity:    "medium",
			Confidence:  1.0,
			URL:         snapshot.URL,
			Summary:     fmt.Sprintf("Excessive redirect chain (%d hops) detected", len(snapshot.RedirectChain)),
			Evidence: map[string]interface{}{
				"hops":  len(snapshot.RedirectChain),
				"chain": snapshot.RedirectChain,
			},
			Remediation: "Point internal links and canonical directives directly to the final 200 OK destination to eliminate redirect hops.",
		}
	}

	return nil
}
