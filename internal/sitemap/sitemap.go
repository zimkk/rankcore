package sitemap

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
)

// URLSet represents a standard sitemap.xml
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// SitemapIndex represents a sitemap index file
type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// URL represents a single URL entry in a sitemap.
type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap represents a sitemap reference in a sitemap index.
type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// ParsedSitemap represents the parsed result of either a URL set or a sitemap index.
type ParsedSitemap struct {
	IsIndex  bool
	URLs     []URL
	Sitemaps []Sitemap
}

// maybeDecompress checks if the reader contains gzipped content (magic bytes 0x1f, 0x8b)
// and returns a decompressed reader if so, or the original reader if not.
func maybeDecompress(r io.Reader) (io.Reader, error) {
	br := bufio.NewReader(r)
	peek, err := br.Peek(2)
	if err == nil && len(peek) >= 2 && peek[0] == 0x1f && peek[1] == 0x8b {
		return gzip.NewReader(br)
	}
	return br, nil
}

// ParseURLSet parses a standard sitemap XML, automatically decompressing gzip if needed.
func ParseURLSet(r io.Reader) (*URLSet, error) {
	reader, err := maybeDecompress(r)
	if err != nil {
		return nil, fmt.Errorf("decompress sitemap: %w", err)
	}

	var us URLSet
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&us); err != nil {
		return nil, err
	}
	return &us, nil
}

// ParseIndex parses a sitemap index XML, automatically decompressing gzip if needed.
func ParseIndex(r io.Reader) (*SitemapIndex, error) {
	reader, err := maybeDecompress(r)
	if err != nil {
		return nil, fmt.Errorf("decompress sitemap index: %w", err)
	}

	var si SitemapIndex
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&si); err != nil {
		return nil, err
	}
	return &si, nil
}

// Parse automatically detects and parses either a URL set or a sitemap index,
// supporting both plain XML and gzipped (.xml.gz) streams.
func Parse(r io.Reader) (*ParsedSitemap, error) {
	reader, err := maybeDecompress(r)
	if err != nil {
		return nil, fmt.Errorf("decompress sitemap: %w", err)
	}

	// Buffer content to allow probing both root element types
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, fmt.Errorf("read sitemap: %w", err)
	}

	// Try SitemapIndex first
	var si SitemapIndex
	dec1 := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := dec1.Decode(&si); err == nil && len(si.Sitemaps) > 0 {
		return &ParsedSitemap{
			IsIndex:  true,
			Sitemaps: si.Sitemaps,
		}, nil
	}

	// Try URLSet
	var us URLSet
	dec2 := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := dec2.Decode(&us); err == nil && (len(us.URLs) > 0 || us.XMLName.Local == "urlset") {
		return &ParsedSitemap{
			IsIndex: false,
			URLs:    us.URLs,
		}, nil
	}

	return nil, fmt.Errorf("unrecognized sitemap format: expected <urlset> or <sitemapindex>")
}
