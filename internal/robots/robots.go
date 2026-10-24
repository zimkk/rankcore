package robots

import (
	"bufio"
	"io"
	"net/url"
	"strings"
)

// Group represents a set of rules for a specific User-Agent.
type Group struct {
	UserAgents []string
	Allow      []string
	Disallow   []string
}

// RobotsTxt represents the parsed contents of a robots.txt file.
type RobotsTxt struct {
	Groups   []*Group
	Sitemaps []string
}

// Parse reads a robots.txt stream and returns a parsed model.
func Parse(r io.Reader) *RobotsTxt {
	rt := &RobotsTxt{}
	scanner := bufio.NewScanner(io.LimitReader(r, 500*1024))

	var currentGroup *Group

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "user-agent":
			if currentGroup == nil || len(currentGroup.Allow) > 0 || len(currentGroup.Disallow) > 0 {
				currentGroup = &Group{}
				rt.Groups = append(rt.Groups, currentGroup)
			}
			currentGroup.UserAgents = append(currentGroup.UserAgents, val)
		case "allow":
			if currentGroup != nil {
				currentGroup.Allow = append(currentGroup.Allow, val)
			}
		case "disallow":
			if currentGroup != nil {
				currentGroup.Disallow = append(currentGroup.Disallow, val)
			}
		case "sitemap":
			rt.Sitemaps = append(rt.Sitemaps, val)
		}
	}

	return rt
}

// IsAllowed checks whether a given user-agent is allowed to access the given
// URL path, implementing RFC 9309 longest-match semantics.
//
// Matching priority:
//  1. Find the most specific group whose User-Agent matches (longest prefix).
//  2. Within that group, find the longest matching Allow or Disallow pattern.
//  3. If the longest match is an Allow, the path is allowed.
//  4. If the longest match is a Disallow, the path is disallowed.
//  5. If no rule matches, the path is allowed (default).
//  6. An empty Disallow means "allow all" for that group.
func (rt *RobotsTxt) IsAllowed(userAgent, path string) bool {
	group := rt.findGroup(userAgent)
	if group == nil {
		return true // no matching group means allowed
	}
	return group.isAllowed(path)
}

// IsDisallowed is the inverse of IsAllowed for convenience.
func (rt *RobotsTxt) IsDisallowed(userAgent, path string) bool {
	return !rt.IsAllowed(userAgent, path)
}

// findGroup locates the best matching group for the given user agent.
// The most specific (longest) matching User-Agent token wins, falling back to
// the wildcard (*) group. Per RFC 9309 §2.2.1, every group naming that same
// token is merged into one rule set.
func (rt *RobotsTxt) findGroup(userAgent string) *Group {
	ua := strings.ToLower(userAgent)

	best := ""
	for _, g := range rt.Groups {
		for _, gua := range g.UserAgents {
			token := strings.ToLower(strings.TrimSpace(gua))
			if token == "" || token == "*" {
				continue
			}
			// The user-agent token is matched case-insensitively against the
			// crawler's product token / User-Agent string.
			if strings.Contains(ua, token) && len(token) > len(best) {
				best = token
			}
		}
	}
	if best == "" {
		best = "*"
	}

	var merged *Group
	for _, g := range rt.Groups {
		for _, gua := range g.UserAgents {
			if strings.ToLower(strings.TrimSpace(gua)) == best {
				if merged == nil {
					merged = &Group{}
				}
				merged.UserAgents = append(merged.UserAgents, gua)
				merged.Allow = append(merged.Allow, g.Allow...)
				merged.Disallow = append(merged.Disallow, g.Disallow...)
				break
			}
		}
	}
	return merged
}

// PathOf returns the path-plus-query string that robots.txt rules match against.
func PathOf(u *url.URL) string {
	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}
	if u.RawQuery != "" {
		p += "?" + u.RawQuery
	}
	return p
}

// isAllowed evaluates the Allow/Disallow rules within a single group using
// RFC 9309 longest-match-wins semantics.
func (g *Group) isAllowed(path string) bool {
	bestLen := -1
	allowed := true // default when no rule matches

	for _, pattern := range g.Allow {
		if n := matchLength(pattern, path); n > bestLen {
			bestLen = n
			allowed = true
		}
	}
	for _, pattern := range g.Disallow {
		if pattern == "" {
			// Empty Disallow = allow everything. Only wins if no other rule matched.
			continue
		}
		if n := matchLength(pattern, path); n > bestLen {
			bestLen = n
			allowed = false
		}
	}
	return allowed
}

// matchLength returns the length of the pattern if it matches the path, or -1.
// Supports RFC 9309 path matching:
//   - Simple prefix matching (path starts with pattern)
//   - Wildcard (*) matches any sequence of characters
//   - End-of-path anchor ($) matches end of string
func matchLength(pattern, path string) int {
	if pattern == "" {
		return -1
	}

	// Handle end-of-path anchor
	mustEnd := false
	if strings.HasSuffix(pattern, "$") {
		mustEnd = true
		pattern = pattern[:len(pattern)-1]
	}

	// If no wildcards, do simple prefix match
	if !strings.Contains(pattern, "*") {
		if strings.HasPrefix(path, pattern) {
			if mustEnd && len(path) != len(pattern) {
				return -1
			}
			return len(pattern)
		}
		return -1
	}

	// Wildcard matching: split pattern by * and match segments in order
	segments := strings.Split(pattern, "*")
	pos := 0
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		idx := strings.Index(path[pos:], seg)
		if idx == -1 {
			return -1
		}
		// First segment must match at the start
		if i == 0 && idx != 0 {
			return -1
		}
		pos += idx + len(seg)
	}
	if mustEnd && pos != len(path) {
		return -1
	}
	return len(pattern)
}

// MatchedRule returns the specific Disallow rule that blocks a path, or empty
// string if the path is allowed. Useful for evidence in findings.
func (rt *RobotsTxt) MatchedRule(userAgent, path string) string {
	group := rt.findGroup(userAgent)
	if group == nil {
		return ""
	}
	if group.isAllowed(path) {
		return ""
	}

	// Find the winning Disallow rule
	bestLen := -1
	bestRule := ""
	for _, pattern := range group.Disallow {
		if pattern == "" {
			continue
		}
		if n := matchLength(pattern, path); n > bestLen {
			bestLen = n
			bestRule = "Disallow: " + pattern
		}
	}
	return bestRule
}
