package extract

import (
	"encoding/json"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Metadata holds everything extracted from an HTML document.
type Metadata struct {
	Title       string
	MetaDesc    string
	Canonical   string   // first <link rel="canonical"> href
	Canonicals  []string // every canonical href, in document order
	RobotsMeta  string   // combined content of <meta name="robots"> and <meta name="googlebot">
	Viewport    string
	Lang        string
	BaseHref    string
	Links       []Link
	Hreflang    map[string]string
	JSONLD      []string
	JSONLDValid []bool // true if the corresponding JSONLD entry is valid JSON
	H1          []string
	Headings    map[string][]string // "h1" through "h6"
	Images      []ImageInfo
	TextLength  int // characters of whitespace-normalized visible body text
	WordCount   int
	OpenGraph   map[string]string // property -> content (e.g. og:title, og:image)
	TwitterCard map[string]string // name -> content (e.g. twitter:card, twitter:title)
}

// Link is an <a> element as a crawler sees it.
type Link struct {
	Href     string // raw href attribute value (unresolved)
	HasHref  bool   // false for <a> without href (e.g. onclick-only navigation)
	Text     string // accessible text: inner text, falling back to img alt / aria-label / title
	Nofollow bool
}

// ImageInfo describes an <img> element.
type ImageInfo struct {
	Src    string
	Alt    string
	HasAlt bool
}

// skippedText lists elements whose text content is never visible page copy.
var skippedText = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Noscript: true,
	atom.Template: true, atom.Svg: true, atom.Head: true,
}

// Extract parses an HTML document and returns Metadata.
func Extract(body io.Reader) (*Metadata, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	meta := &Metadata{
		Hreflang:    make(map[string]string),
		Headings:    make(map[string][]string),
		OpenGraph:   make(map[string]string),
		TwitterCard: make(map[string]string),
	}

	var text strings.Builder
	var robots []string

	var walk func(n *html.Node, inSVG bool)
	walk = func(n *html.Node, inSVG bool) {
		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Html:
				meta.Lang = strings.TrimSpace(attr(n, "lang"))

			case atom.Svg:
				inSVG = true

			case atom.Title:
				if !inSVG && meta.Title == "" {
					meta.Title = collapse(textContent(n))
				}

			case atom.Base:
				if meta.BaseHref == "" {
					meta.BaseHref = strings.TrimSpace(attr(n, "href"))
				}

			case atom.Meta:
				name := strings.ToLower(strings.TrimSpace(attr(n, "name")))
				property := strings.ToLower(strings.TrimSpace(attr(n, "property")))
				content := strings.TrimSpace(attr(n, "content"))
				if strings.HasPrefix(property, "og:") && content != "" {
					meta.OpenGraph[property] = content
				}
				if strings.HasPrefix(name, "twitter:") && content != "" {
					meta.TwitterCard[name] = content
				}
				switch name {
				case "description":
					if meta.MetaDesc == "" {
						meta.MetaDesc = content
					}
				case "robots", "googlebot":
					if content != "" {
						robots = append(robots, content)
					}
				case "viewport":
					meta.Viewport = content
				}

			case atom.Link:
				rels := strings.Fields(strings.ToLower(attr(n, "rel")))
				href := strings.TrimSpace(attr(n, "href"))
				for _, rel := range rels {
					switch rel {
					case "canonical":
						meta.Canonicals = append(meta.Canonicals, href)
					case "alternate":
						if hl := strings.TrimSpace(attr(n, "hreflang")); hl != "" {
							meta.Hreflang[hl] = href
						}
					}
				}

			case atom.A:
				href, hasHref := attrOK(n, "href")
				_, hasOnclick := attrOK(n, "onclick")
				if hasHref || hasOnclick {
					label := collapse(textContent(n))
					if label == "" {
						label = strings.TrimSpace(attr(n, "aria-label"))
					}
					if label == "" {
						label = imageAlt(n)
					}
					if label == "" {
						label = strings.TrimSpace(attr(n, "title"))
					}
					meta.Links = append(meta.Links, Link{
						Href:     strings.TrimSpace(href),
						HasHref:  hasHref,
						Text:     label,
						Nofollow: hasToken(attr(n, "rel"), "nofollow"),
					})
				}

			case atom.Img:
				img := ImageInfo{Src: attr(n, "src")}
				if img.Src == "" {
					img.Src = attr(n, "data-src")
				}
				img.Alt, img.HasAlt = attrOK(n, "alt")
				meta.Images = append(meta.Images, img)

			case atom.Script:
				if strings.EqualFold(strings.TrimSpace(attr(n, "type")), "application/ld+json") {
					raw := textContent(n)
					meta.JSONLD = append(meta.JSONLD, raw)
					meta.JSONLDValid = append(meta.JSONLDValid, json.Valid([]byte(raw)))
				}

			case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
				heading := collapse(textContent(n))
				tag := n.Data
				meta.Headings[tag] = append(meta.Headings[tag], heading)
				if n.DataAtom == atom.H1 {
					meta.H1 = append(meta.H1, heading)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inSVG)
		}
	}
	walk(doc, false)

	if len(meta.Canonicals) > 0 {
		meta.Canonical = meta.Canonicals[0]
	}
	meta.RobotsMeta = strings.Join(robots, ", ")

	if body := findBody(doc); body != nil {
		visibleText(body, &text)
	}
	visible := collapse(text.String())
	meta.TextLength = len([]rune(visible))
	meta.WordCount = len(strings.Fields(visible))

	return meta, nil
}

func attr(n *html.Node, key string) string {
	v, _ := attrOK(n, key)
	return v
}

func attrOK(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val, true
		}
	}
	return "", false
}

func hasToken(list, token string) bool {
	for _, t := range strings.Fields(strings.ToLower(list)) {
		if t == token {
			return true
		}
	}
	return false
}

// textContent returns the concatenated text of all descendant text nodes.
func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && (c.DataAtom == atom.Script || c.DataAtom == atom.Style) {
				continue
			}
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

// imageAlt returns the first non-empty alt text of an image inside n.
func imageAlt(n *html.Node) string {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom.Img {
			if alt := strings.TrimSpace(attr(c, "alt")); alt != "" {
				return alt
			}
		}
		if alt := imageAlt(c); alt != "" {
			return alt
		}
	}
	return ""
}

func visibleText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		b.WriteString(n.Data)
		b.WriteByte(' ')
		return
	}
	if n.Type == html.ElementNode && skippedText[n.DataAtom] {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visibleText(c, b)
	}
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == atom.Body {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if b := findBody(c); b != nil {
			return b
		}
	}
	return nil
}

// collapse trims and collapses runs of whitespace into single spaces.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
