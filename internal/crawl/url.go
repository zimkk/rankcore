package crawl

import (
	"net/url"
	"strings"
)

// NormalizeURL cleans and normalizes a URL for deduplication in the frontier.
func NormalizeURL(rawURL string, base *url.URL) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if base != nil {
		u = base.ResolveReference(u)
	}

	// Remove fragment
	u.Fragment = ""

	// Ensure lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	return u.String(), nil
}

// IsSameOrigin checks if a target URL shares the same origin (scheme + host) as the base.
func IsSameOrigin(target, base *url.URL) bool {
	return target.Scheme == base.Scheme && target.Host == base.Host
}
