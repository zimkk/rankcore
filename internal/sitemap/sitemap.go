package sitemap

import (
	"encoding/xml"
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

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// ParseURLSet parses a standard sitemap XML.
func ParseURLSet(r io.Reader) (*URLSet, error) {
	var us URLSet
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&us); err != nil {
		return nil, err
	}
	return &us, nil
}

// ParseIndex parses a sitemap index XML.
func ParseIndex(r io.Reader) (*SitemapIndex, error) {
	var si SitemapIndex
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&si); err != nil {
		return nil, err
	}
	return &si, nil
}
