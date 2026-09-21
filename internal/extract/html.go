package extract

import (
	"io"
	"strings"
	"golang.org/x/net/html"
	"rankcore/internal/crawl"
)

// Extract parses an HTML document and populates the PageSnapshot.
func Extract(body io.Reader, snapshot *crawl.PageSnapshot) error {
	z := html.NewTokenizer(body)

	snapshot.Hreflang = make(map[string]string)

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			switch t.Data {
			case "title":
				if tt == html.StartTagToken {
					if z.Next() == html.TextToken {
						snapshot.Title = strings.TrimSpace(z.Token().Data)
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
					snapshot.MetaDesc = content
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
					snapshot.Canonical = href
				} else if rel == "alternate" && hreflang != "" {
					snapshot.Hreflang[hreflang] = href
				}
			case "a":
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						href := strings.TrimSpace(attr.Val)
						if strings.HasPrefix(href, "http") || strings.HasPrefix(href, "/") {
							// very simplified link classification
							snapshot.InternalLinks = append(snapshot.InternalLinks, href)
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
							snapshot.JSONLD = append(snapshot.JSONLD, z.Token().Data)
						}
					}
				}
			}
		}
	}
}
