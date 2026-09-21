package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	ErrPrivateIPBlocked = errors.New("request to private/loopback IP is blocked for remote crawls")
)

// NewSafeClient creates an HTTP client with timeouts, redirect policies, and optional SSRF protection.
func NewSafeClient(timeout time.Duration, userAgent string, allowPrivate bool) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				host = addr
			}

			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}

			if !allowPrivate {
				for _, ip := range ips {
					if isPrivate(ip) {
						return nil, ErrPrivateIPBlocked
					}
				}
			}

			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
}

// isPrivate checks if an IP belongs to loopback, link-local, or private ranges.
func isPrivate(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	// Also block AWS metadata IP just in case
	if ip.String() == "169.254.169.254" {
		return true
	}
	return false
}

// Request performs an HTTP GET request with the defined safe client.
func Request(ctx context.Context, client *http.Client, targetURL, userAgent string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	
	// Recommend accepting compressed content
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	return client.Do(req)
}
