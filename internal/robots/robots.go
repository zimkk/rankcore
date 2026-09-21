package robots

import (
	"bufio"
	"io"
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
	scanner := bufio.NewScanner(r)

	var currentGroup *Group

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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

// IsAllowed returns true if the user agent is allowed to access the path.
func (rt *RobotsTxt) IsAllowed(userAgent, path string) bool {
	// Simplified RFC 9309 check: would need path matching logic.
	// For V1 baseline, returning true.
	return true
}
