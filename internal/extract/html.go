package extract

import (
	"io"
	"strings"
	"golang.org/x/net/html"
)

type Metadata struct {
	Title         string
	MetaDesc      string
	Canonical     string
	InternalLinks []string
	Hreflang      map[string]string
	JSONLD        []string
}

// Extract parses an HTML document and returns Metadata.
func Extract(body io.Reader) (*Metadata, error) {
	z := html.NewTokenizer(body)

	meta := &Metadata{
		Hreflang: make(map[string]string),
	}

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return meta, nil
			}
			return meta, z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			switch t.Data {
			case "title":
				if tt == html.StartTagToken {
					if z.Next() == html.TextToken {
						meta.Title = strings.TrimSpace(z.Token().Data)
					}
				}
			case "meta":
				var name, content string
				for _, attr := range t.Attr {
					if attr.Key == "name" {
						name = strings.ToLower(attr.Val)
					} else if attr.Key == "content" {
						content = attr.Val
					}
				}
				if name == "description" {
					meta.MetaDesc = content
				}
			case "link":
				var rel, href, hreflang string
				for _, attr := range t.Attr {
					if attr.Key == "rel" {
						rel = strings.ToLower(attr.Val)
					} else if attr.Key == "href" {
						href = attr.Val
					} else if attr.Key == "hreflang" {
						hreflang = attr.Val
					}
				}
				if rel == "canonical" {
					meta.Canonical = href
				} else if rel == "alternate" && hreflang != "" {
					meta.Hreflang[hreflang] = href
				}
			case "a":
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						href := strings.TrimSpace(attr.Val)
						if strings.HasPrefix(href, "http") || strings.HasPrefix(href, "/") {
							// very simplified link classification
							meta.InternalLinks = append(meta.InternalLinks, href)
						}
					}
				}
			case "script":
				var typ string
				for _, attr := range t.Attr {
					if attr.Key == "type" {
						typ = strings.ToLower(attr.Val)
					}
				}
				if typ == "application/ld+json" {
					if tt == html.StartTagToken {
						if z.Next() == html.TextToken {
							meta.JSONLD = append(meta.JSONLD, z.Token().Data)
						}
					}
				}
			}
		}
	}
}
