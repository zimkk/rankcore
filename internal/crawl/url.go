package crawl

import (
	"net/url"
	"strings"
)

// NormalizeURL cleans and normalizes a URL for deduplication in the frontier.
func NormalizeURL(rawURL string, base *url.URL) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}

	if base != nil {
		u = base.ResolveReference(u)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", &url.Error{Op: "normalize", URL: rawURL, Err: errUnsupportedScheme{scheme: u.Scheme}}
	}

	// Remove fragment
	u.Fragment = ""
	u.RawFragment = ""

	// Strip userinfo; audited pages should not embed credentials.
	u.User = nil

	// Ensure lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Drop default ports.
	if (u.Scheme == "http" && strings.HasSuffix(u.Host, ":80")) ||
		(u.Scheme == "https" && strings.HasSuffix(u.Host, ":443")) {
		u.Host = u.Host[:len(u.Host)-3]
	}

	return u.String(), nil
}

type errUnsupportedScheme struct{ scheme string }

func (e errUnsupportedScheme) Error() string {
	return "unsupported URL scheme: " + e.scheme
}

// IsSameOrigin checks if a target URL shares the same origin (scheme + host) as the base.
func IsSameOrigin(target, base *url.URL) bool {
	return target.Scheme == base.Scheme && target.Host == base.Host
}
