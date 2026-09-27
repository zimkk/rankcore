package sitemap

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func TestParseURLSet(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://example.com/</loc>
      <lastmod>2026-01-01</lastmod>
      <changefreq>daily</changefreq>
      <priority>1.0</priority>
   </url>
   <url>
      <loc>https://example.com/about</loc>
   </url>
</urlset>`

	us, err := ParseURLSet(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(us.URLs) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(us.URLs))
	}
	if us.URLs[0].Loc != "https://example.com/" {
		t.Errorf("expected first URL https://example.com/, got %s", us.URLs[0].Loc)
	}
}

func TestParseGzipSitemap(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://example.com/products</loc>
   </url>
</urlset>`

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(xmlData))
	gw.Close()

	parsed, err := Parse(&buf)
	if err != nil {
		t.Fatalf("failed to parse gzipped sitemap: %v", err)
	}
	if parsed.IsIndex {
		t.Fatalf("expected urlset, got index")
	}
	if len(parsed.URLs) != 1 || parsed.URLs[0].Loc != "https://example.com/products" {
		t.Errorf("unexpected URLs: %+v", parsed.URLs)
	}
}

func TestParseSitemapIndex(t *testing.T) {
	indexData := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <sitemap>
      <loc>https://example.com/sitemap1.xml.gz</loc>
      <lastmod>2026-01-01</lastmod>
   </sitemap>
   <sitemap>
      <loc>https://example.com/sitemap2.xml.gz</loc>
   </sitemap>
</sitemapindex>`

	parsed, err := Parse(strings.NewReader(indexData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !parsed.IsIndex {
		t.Fatalf("expected index, got urlset")
	}
	if len(parsed.Sitemaps) != 2 {
		t.Fatalf("expected 2 sitemaps, got %d", len(parsed.Sitemaps))
	}
	if parsed.Sitemaps[0].Loc != "https://example.com/sitemap1.xml.gz" {
		t.Errorf("unexpected loc: %s", parsed.Sitemaps[0].Loc)
	}
}
