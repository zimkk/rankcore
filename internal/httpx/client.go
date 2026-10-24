package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrPrivateIPBlocked = errors.New("request to private/loopback IP is blocked for remote crawls")
	ErrRedirectLoop     = errors.New("redirect loop detected")
	ErrTooManyRedirects = errors.New("stopped after 10 redirects")
)

// MaxRedirects bounds how many hops Fetch follows.
const MaxRedirects = 10

// NewSafeClient creates an HTTP client with timeouts and optional SSRF protection.
// The client never follows redirects on its own; Fetch follows them so every hop
// is recorded and re-checked.
func NewSafeClient(timeout time.Duration, allowPrivate bool) *http.Client {
	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}

	transport := &http.Transport{
		Proxy: nil, // a proxy would bypass the address checks below
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses for %s", host)
			}
			if !allowPrivate {
				for _, ip := range ips {
					if IsPrivate(ip.IP) {
						return nil, ErrPrivateIPBlocked
					}
				}
			}
			// Dial the address we just validated so a second DNS answer cannot
			// redirect the connection somewhere else (DNS rebinding).
			var lastErr error
			for _, ip := range ips {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// IsPrivate reports whether an IP belongs to loopback, link-local, private,
// unspecified, or shared (CGNAT) address space.
func IsPrivate(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1]&0xc0 == 64 {
		return true // 100.64.0.0/10
	}
	return false
}

// Result is the outcome of fetching one URL, including every redirect hop.
type Result struct {
	URL           string
	FinalURL      string
	StatusCode    int
	Header        http.Header
	Body          []byte
	Truncated     bool
	RedirectChain []string // every URL that answered with a redirect, in order
	RedirectLoop  bool
	Duration      time.Duration // time to first byte of the final response
	Err           error
}

// Fetch GETs targetURL, following up to MaxRedirects redirects manually and
// retrying briefly on 429/503. The body is read up to maxBody bytes.
func Fetch(ctx context.Context, client *http.Client, targetURL, userAgent string, maxBody int64) *Result {
	res := &Result{URL: targetURL, FinalURL: targetURL}
	seen := map[string]bool{}
	current := targetURL

	for hop := 0; ; hop++ {
		if seen[current] {
			res.RedirectLoop = true
			res.Err = ErrRedirectLoop
			return res
		}
		seen[current] = true
		if hop > MaxRedirects {
			res.Err = ErrTooManyRedirects
			return res
		}

		resp, dur, err := doWithRetry(ctx, client, current, userAgent)
		if err != nil {
			res.Err = err
			return res
		}
		res.FinalURL = current
		res.StatusCode = resp.StatusCode
		res.Header = resp.Header
		res.Duration = dur

		if isRedirect(resp.StatusCode) {
			loc := resp.Header.Get("Location")
			io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
			resp.Body.Close()
			if loc == "" {
				return res
			}
			next, err := resp.Request.URL.Parse(loc)
			if err != nil {
				res.Err = fmt.Errorf("invalid redirect location %q: %w", loc, err)
				return res
			}
			next.Fragment = ""
			res.RedirectChain = append(res.RedirectChain, current)
			current = next.String()
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
		resp.Body.Close()
		if int64(len(body)) > maxBody {
			body = body[:maxBody]
			res.Truncated = true
		}
		res.Body = body
		if err != nil {
			res.Err = err
		}
		return res
	}
}

func doWithRetry(ctx context.Context, client *http.Client, target, userAgent string) (*http.Response, time.Duration, error) {
	const attempts = 3
	for i := 0; ; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			return nil, 0, err
		}
		dur := time.Since(start)

		if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) && i < attempts-1 {
			wait := retryAfter(resp.Header.Get("Retry-After"), time.Duration(i+1)*time.Second)
			io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}
		return resp, dur, nil
	}
}

// retryAfter parses a Retry-After header, capped at 10 seconds.
func retryAfter(v string, fallback time.Duration) time.Duration {
	const max = 10 * time.Second
	d := fallback
	if secs, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && secs >= 0 {
		d = time.Duration(secs) * time.Second
	} else if t, err := http.ParseTime(v); err == nil {
		d = time.Until(t)
	}
	if d < 0 {
		d = 0
	}
	if d > max {
		d = max
	}
	return d
}

func isRedirect(code int) bool {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

// IsLocalTarget reports whether a seed host was explicitly a local or private
// address, in which case private-network access is expected.
func IsLocalTarget(u *url.URL) bool {
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return IsPrivate(ip)
	}
	return false
}
