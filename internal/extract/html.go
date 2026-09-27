package extract

import (
	"encoding/json"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Metadata holds everything extracted from an HTML document.
type Metadata struct {
	Title         string
	MetaDesc      string
	Canonical     string
	RobotsMeta    string
	Viewport      string
	Lang          string
	InternalLinks []string
	ExternalLinks []string
	AnchorTexts   map[string]string // href -> visible text
	Hreflang      map[string]string
	JSONLD        []string
	JSONLDValid   []bool // true if the corresponding JSONLD entry is valid JSON
	H1            []string
	Headings      map[string][]string // "h1" through "h6"
	Images        []ImageInfo
	TextLength    int
	OpenGraph     map[string]string // property -> content (e.g. og:title, og:image)
	TwitterCard   map[string]string // name -> content (e.g. twitter:card, twitter:title)
}

// ImageInfo describes an <img> element.
type ImageInfo struct {
	Src    string
	Alt    string
	HasAlt bool
}

// headingTags is the set of heading tag names we track.
var headingTags = map[string]bool{
	"h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true,
}

// Extract parses an HTML document and returns Metadata.
func Extract(body io.Reader) (*Metadata, error) {
	z := html.NewTokenizer(body)

	meta := &Metadata{
		Hreflang:    make(map[string]string),
		AnchorTexts: make(map[string]string),
		Headings:    make(map[string][]string),
		OpenGraph:   make(map[string]string),
		TwitterCard: make(map[string]string),
	}

	// inBody loosely tracks whether we have left <head>.
	inBody := false

	// textBuf accumulates visible body text for TextLength estimation.
	var textBuf strings.Builder

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				meta.TextLength = textBuf.Len()
				return meta, nil
			}
			return meta, z.Err()

		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			switch t.Data {
			case "html":
				for _, attr := range t.Attr {
					if attr.Key == "lang" {
						meta.Lang = attr.Val
					}
				}

			case "body":
				inBody = true

			case "title":
				if tt == html.StartTagToken {
					if z.Next() == html.TextToken {
						meta.Title = strings.TrimSpace(z.Token().Data)
					}
				}

			case "meta":
				var name, content, property string
				for _, attr := range t.Attr {
					switch strings.ToLower(attr.Key) {
					case "name":
						name = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					case "property":
						property = strings.ToLower(attr.Val)
					}
				}
				if property != "" && content != "" {
					if strings.HasPrefix(property, "og:") {
						meta.OpenGraph[property] = content
					}
				}
				if name != "" && content != "" {
					if strings.HasPrefix(name, "twitter:") {
						meta.TwitterCard[name] = content
					}
				}
				switch name {
				case "description":
					meta.MetaDesc = content
				case "robots":
					meta.RobotsMeta = content
				case "viewport":
					meta.Viewport = content
				}

			case "link":
				var rel, href, hreflang string
				for _, attr := range t.Attr {
					switch attr.Key {
					case "rel":
						rel = strings.ToLower(attr.Val)
					case "href":
						href = attr.Val
					case "hreflang":
						hreflang = attr.Val
					}
				}
				if rel == "canonical" {
					meta.Canonical = href
				} else if rel == "alternate" && hreflang != "" {
					meta.Hreflang[hreflang] = href
				}

			case "a":
				href, anchorText := extractAnchor(z, t, tt)
				if href == "" {
					break
				}
				trimmed := strings.TrimSpace(href)
				if isInternalHref(trimmed) {
					meta.InternalLinks = append(meta.InternalLinks, trimmed)
				} else if strings.HasPrefix(trimmed, "http") {
					meta.ExternalLinks = append(meta.ExternalLinks, trimmed)
				}
				if trimmed != "" {
					meta.AnchorTexts[trimmed] = strings.TrimSpace(anchorText)
				}

			case "img":
				img := ImageInfo{}
				for _, attr := range t.Attr {
					switch attr.Key {
					case "src":
						img.Src = attr.Val
					case "alt":
						img.HasAlt = true
						img.Alt = attr.Val
					}
				}
				meta.Images = append(meta.Images, img)

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
							raw := z.Token().Data
							meta.JSONLD = append(meta.JSONLD, raw)
							meta.JSONLDValid = append(meta.JSONLDValid, json.Valid([]byte(raw)))
						}
					}
				}

			default:
				// Track heading tags h1-h6
				if headingTags[t.Data] {
					if tt == html.StartTagToken {
						text := collectInnerText(z, t.Data)
						trimmed := strings.TrimSpace(text)
						meta.Headings[t.Data] = append(meta.Headings[t.Data], trimmed)
						if t.Data == "h1" {
							meta.H1 = append(meta.H1, trimmed)
						}
					}
				}
			}

		case html.TextToken:
			if inBody {
				textBuf.WriteString(z.Token().Data)
			}
		}
	}
}

// extractAnchor reads the href from an <a> token and collects its inner text.
func extractAnchor(z *html.Tokenizer, t html.Token, tt html.TokenType) (href, anchorText string) {
	for _, attr := range t.Attr {
		if attr.Key == "href" {
			href = attr.Val
		}
	}
	if tt == html.StartTagToken {
		anchorText = collectInnerText(z, "a")
	}
	return
}

// collectInnerText reads tokens until the matching end tag, accumulating text.
func collectInnerText(z *html.Tokenizer, tagName string) string {
	var buf strings.Builder
	depth := 1
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return buf.String()
		case html.TextToken:
			buf.WriteString(z.Token().Data)
		case html.StartTagToken:
			t := z.Token()
			if t.Data == tagName {
				depth++
			}
		case html.EndTagToken:
			t := z.Token()
			if t.Data == tagName {
				depth--
				if depth <= 0 {
					return buf.String()
				}
			}
		}
	}
}

// isInternalHref returns true for hrefs that look like same-site navigation.
func isInternalHref(href string) bool {
	return strings.HasPrefix(href, "/") || strings.HasPrefix(href, "./") || strings.HasPrefix(href, "../")
}
