package robots

import (
	"strings"
	"testing"
)

func parse(s string) *RobotsTxt {
	return Parse(strings.NewReader(s))
}

func TestParseBasic(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /private
Allow: /private/public

User-agent: Googlebot
Disallow: /no-google

Sitemap: https://example.com/sitemap.xml
`)
	if len(rt.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(rt.Groups))
	}
	if len(rt.Sitemaps) != 1 || rt.Sitemaps[0] != "https://example.com/sitemap.xml" {
		t.Errorf("unexpected sitemaps: %v", rt.Sitemaps)
	}
}

func TestIsAllowed_Wildcard(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /private
Allow: /private/public
`)
	tests := []struct {
		ua   string
		path string
		want bool
	}{
		{"SomeBot", "/", true},
		{"SomeBot", "/about", true},
		{"SomeBot", "/private", false},
		{"SomeBot", "/private/secret", false},
		{"SomeBot", "/private/public", true},
		{"SomeBot", "/private/public/page", true},
	}
	for _, tt := range tests {
		got := rt.IsAllowed(tt.ua, tt.path)
		if got != tt.want {
			t.Errorf("IsAllowed(%q, %q) = %v, want %v", tt.ua, tt.path, got, tt.want)
		}
	}
}

func TestIsAllowed_SpecificAgent(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow:

User-agent: Googlebot
Disallow: /no-google
`)
	if !rt.IsAllowed("SomeBot", "/no-google") {
		t.Error("SomeBot should be allowed on /no-google")
	}
	if rt.IsAllowed("Googlebot/2.1", "/no-google") {
		t.Error("Googlebot should be blocked on /no-google")
	}
	if !rt.IsAllowed("Googlebot/2.1", "/about") {
		t.Error("Googlebot should be allowed on /about")
	}
}

func TestIsAllowed_LongestMatch(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /a
Allow: /a/b
Disallow: /a/b/c
`)
	tests := []struct {
		path string
		want bool
	}{
		{"/a", false},
		{"/a/x", false},
		{"/a/b", true},
		{"/a/b/x", true},
		{"/a/b/c", false},
		{"/a/b/c/d", false},
	}
	for _, tt := range tests {
		got := rt.IsAllowed("bot", tt.path)
		if got != tt.want {
			t.Errorf("IsAllowed(bot, %q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsAllowed_Wildcards(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /*.json
Disallow: /api/*/internal
`)
	tests := []struct {
		path string
		want bool
	}{
		{"/data.json", false},
		{"/path/to/file.json", false},
		{"/data.jsonp", false}, // starts with *.json prefix
		{"/data.xml", true},
		{"/api/v1/internal", false},
		{"/api/v2/internal", false},
		{"/api/v1/public", true},
	}
	for _, tt := range tests {
		got := rt.IsAllowed("bot", tt.path)
		if got != tt.want {
			t.Errorf("IsAllowed(bot, %q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsAllowed_EndAnchor(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /exact$
`)
	if rt.IsAllowed("bot", "/exact") {
		t.Error("should block /exact with $ anchor")
	}
	if !rt.IsAllowed("bot", "/exact/more") {
		t.Error("should allow /exact/more with $ anchor")
	}
}

func TestIsAllowed_EmptyDisallow(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow:
`)
	if !rt.IsAllowed("bot", "/anything") {
		t.Error("empty Disallow should allow everything")
	}
}

func TestIsAllowed_DisallowAll(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /
`)
	if rt.IsAllowed("bot", "/anything") {
		t.Error("Disallow: / should block everything")
	}
	if rt.IsAllowed("bot", "/") {
		t.Error("Disallow: / should block root too")
	}
}

func TestIsAllowed_NoMatchingGroup(t *testing.T) {
	rt := parse(`
User-agent: SpecificBot
Disallow: /
`)
	if !rt.IsAllowed("OtherBot", "/page") {
		t.Error("no matching group should default to allowed")
	}
}

func TestIsAllowed_CaseInsensitiveAgent(t *testing.T) {
	rt := parse(`
User-agent: GoogleBot
Disallow: /blocked
`)
	if rt.IsAllowed("googlebot", "/blocked") {
		t.Error("agent matching should be case-insensitive")
	}
}

func TestMatchedRule(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /private
Allow: /private/ok
`)
	rule := rt.MatchedRule("bot", "/private/secret")
	if rule == "" {
		t.Error("expected a matched rule")
	}
	if rule != "Disallow: /private" {
		t.Errorf("unexpected rule: %s", rule)
	}

	rule = rt.MatchedRule("bot", "/private/ok")
	if rule != "" {
		t.Errorf("expected no matched rule for allowed path, got: %s", rule)
	}
}

func TestIsDisallowed(t *testing.T) {
	rt := parse(`
User-agent: *
Disallow: /blocked
`)
	if !rt.IsDisallowed("bot", "/blocked") {
		t.Error("IsDisallowed should return true for blocked path")
	}
	if rt.IsDisallowed("bot", "/allowed") {
		t.Error("IsDisallowed should return false for allowed path")
	}
}

func TestParseInlineComments(t *testing.T) {
	rt := parse(`
User-agent: * # all bots
Disallow: /secret # keep out
Allow: /secret/public # but this is ok
`)
	if rt.IsAllowed("bot", "/secret") {
		t.Error("should block /secret")
	}
	if !rt.IsAllowed("bot", "/secret/public") {
		t.Error("should allow /secret/public")
	}
}
