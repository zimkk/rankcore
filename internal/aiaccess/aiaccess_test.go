package aiaccess

import (
	"strings"
	"testing"

	"github.com/zimkk/rankcore/internal/robots"
)

func TestEvaluate(t *testing.T) {
	robotsTxt := `
User-agent: *
Allow: /

User-agent: GPTBot
Disallow: /

User-agent: Claude-User
Allow: /api/public
Disallow: /
`
	rt := robots.Parse(strings.NewReader(robotsTxt))

	rep := Evaluate(rt, "/about")
	if rep.Path != "/about" {
		t.Errorf("expected path /about, got %s", rep.Path)
	}

	foundGPTBot := false
	foundGooglebot := false
	for _, res := range rep.Results {
		if res.Crawler.UserAgent == "GPTBot" {
			foundGPTBot = true
			if res.Allowed {
				t.Errorf("expected GPTBot to be disallowed on /about")
			}
		}
		if res.Crawler.UserAgent == "Googlebot" {
			foundGooglebot = true
			if !res.Allowed {
				t.Errorf("expected Googlebot to be allowed on /about")
			}
		}
	}

	if !foundGPTBot || !foundGooglebot {
		t.Fatalf("expected both GPTBot and Googlebot to be evaluated")
	}

	summary := rep.SummaryByPurpose()
	if summary[PurposeTraining].Blocked == 0 {
		t.Errorf("expected at least 1 blocked training crawler")
	}
}
