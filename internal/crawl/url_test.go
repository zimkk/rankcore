package crawl

import (
	"net/url"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	
	tests := []struct {
		raw      string
		expected string
	}{
		{"https://example.com/path#fragment", "https://example.com/path"},
		{"HTTP://EXAMPLE.COM/Path", "http://example.com/Path"},
		{"/relative/path", "https://example.com/relative/path"},
	}
	
	for _, tc := range tests {
		normalized, err := NormalizeURL(tc.raw, baseURL)
		if err != nil {
			t.Errorf("NormalizeURL(%q) returned error: %v", tc.raw, err)
		}
		if normalized != tc.expected {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tc.raw, normalized, tc.expected)
		}
	}
}

func TestIsSameOrigin(t *testing.T) {
	base, _ := url.Parse("https://example.com")
	
	target1, _ := url.Parse("https://example.com/page")
	if !IsSameOrigin(target1, base) {
		t.Error("Expected same origin for target1")
	}
	
	target2, _ := url.Parse("http://example.com")
	if IsSameOrigin(target2, base) {
		t.Error("Expected different origin for target2 due to scheme")
	}
	
	target3, _ := url.Parse("https://sub.example.com")
	if IsSameOrigin(target3, base) {
		t.Error("Expected different origin for target3 due to host")
	}
}
