package httpx

import (
	"compress/gzip"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestFetchDecompressesGzip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") == "" {
			t.Error("transport should advertise gzip itself")
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/html")
		gz := gzip.NewWriter(w)
		gz.Write([]byte("<title>compressed</title>"))
		gz.Close()
	}))
	defer srv.Close()

	res := Fetch(context.Background(), NewSafeClient(5*time.Second, true), srv.URL, "test", 1<<20)
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if string(res.Body) != "<title>compressed</title>" {
		t.Errorf("body was not transparently decompressed: %q", res.Body)
	}
}

func TestFetchRecordsRedirectChainAndLoops(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/b", http.StatusMovedPermanently) })
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/c", http.StatusFound) })
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/loop1", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/loop2", http.StatusFound) })
	mux.HandleFunc("/loop2", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/loop1", http.StatusFound) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := NewSafeClient(5*time.Second, true)

	res := Fetch(context.Background(), client, srv.URL+"/a", "test", 1<<20)
	if res.StatusCode != 200 || len(res.RedirectChain) != 2 || res.FinalURL != srv.URL+"/c" {
		t.Errorf("chain = %v final = %s status = %d", res.RedirectChain, res.FinalURL, res.StatusCode)
	}

	loop := Fetch(context.Background(), client, srv.URL+"/loop1", "test", 1<<20)
	if !loop.RedirectLoop || !errors.Is(loop.Err, ErrRedirectLoop) {
		t.Errorf("expected redirect loop, got %+v", loop)
	}
}

func TestFetchTruncatesLargeBodies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 4096))
	}))
	defer srv.Close()
	res := Fetch(context.Background(), NewSafeClient(5*time.Second, true), srv.URL, "test", 1000)
	if !res.Truncated || len(res.Body) != 1000 {
		t.Errorf("expected truncation to 1000 bytes, got %d (truncated=%v)", len(res.Body), res.Truncated)
	}
}

func TestFetchRetriesOn429(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	res := Fetch(context.Background(), NewSafeClient(5*time.Second, true), srv.URL, "test", 1<<20)
	if res.StatusCode != 200 || calls != 2 {
		t.Errorf("expected retry then 200, got status %d after %d calls", res.StatusCode, calls)
	}
}

func TestSSRFProtectionBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	res := Fetch(context.Background(), NewSafeClient(5*time.Second, false), srv.URL, "test", 1<<20)
	if !errors.Is(res.Err, ErrPrivateIPBlocked) {
		t.Errorf("expected private IP block for remote crawl, got %v", res.Err)
	}
}

func TestIsPrivateAndLocalTarget(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.1.2.3", "192.168.0.1", "169.254.169.254", "100.64.0.1", "::1", "0.0.0.0"} {
		if !IsPrivate(net.ParseIP(ip)) {
			t.Errorf("%s should be private", ip)
		}
	}
	for _, ip := range []string{"8.8.8.8", "1.1.1.1", "2606:4700::1111"} {
		if IsPrivate(net.ParseIP(ip)) {
			t.Errorf("%s should be public", ip)
		}
	}
	for raw, want := range map[string]bool{
		"http://localhost:3000":     true,
		"http://app.localhost:5173": true,
		"http://127.0.0.1:8080":     true,
		"http://[::1]:8080":         true,
		"http://192.168.1.20:3000":  true,
		"https://example.com":       false,
	} {
		u, _ := url.Parse(raw)
		if got := IsLocalTarget(u); got != want {
			t.Errorf("IsLocalTarget(%s) = %v, want %v", raw, got, want)
		}
	}
}
