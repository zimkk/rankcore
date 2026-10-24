package extract

import (
	"strings"
	"testing"
)

const page = `<!doctype html>
<html lang="en-GB">
<head>
  <title>  Acme &amp; Co —
     Widgets </title>
  <meta name="description" content="Widgets for everyone">
  <meta name="robots" content="index, follow">
  <meta name="googlebot" content="noindex">
  <link rel="canonical" href="https://acme.test/a">
  <link rel="canonical" href="https://acme.test/b">
  <link rel="alternate" hreflang="fr" href="https://acme.test/fr">
  <script type="application/ld+json">{"@type":"Product","name":"W"}</script>
  <style>body { color: red }</style>
</head>
<body>
  <svg><title>icon title</title></svg>
  <h1>Main <span>heading</span></h1>
  <a href="/one"><img src="/i.png" alt="Logo of Acme"></a>
  <a href="https://acme.test/two" rel="nofollow ugc">Two things</a>
  <a onclick="go()">JS nav</a>
  <a name="anchor-only">not a link</a>
  <p>Hello world from the body.</p>
  <script>var hidden = "do not count these words";</script>
  <noscript>Enable JavaScript</noscript>
</body>
</html>`

func TestExtract(t *testing.T) {
	m, err := Extract(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Acme & Co — Widgets" {
		t.Errorf("title = %q", m.Title)
	}
	if m.Lang != "en-GB" {
		t.Errorf("lang = %q", m.Lang)
	}
	if m.Canonical != "https://acme.test/a" || len(m.Canonicals) != 2 {
		t.Errorf("canonicals = %q / %v", m.Canonical, m.Canonicals)
	}
	if !strings.Contains(m.RobotsMeta, "noindex") {
		t.Errorf("googlebot meta should be merged into robots: %q", m.RobotsMeta)
	}
	if m.Hreflang["fr"] != "https://acme.test/fr" {
		t.Errorf("hreflang = %v", m.Hreflang)
	}
	if len(m.H1) != 1 || m.H1[0] != "Main heading" {
		t.Errorf("h1 = %v", m.H1)
	}
	if len(m.JSONLD) != 1 || !m.JSONLDValid[0] {
		t.Errorf("jsonld = %v", m.JSONLD)
	}
	if len(m.Images) != 1 || m.Images[0].Alt != "Logo of Acme" {
		t.Errorf("images inside links must be collected: %+v", m.Images)
	}

	if len(m.Links) != 3 {
		t.Fatalf("expected 3 links (2 href + 1 onclick), got %+v", m.Links)
	}
	if m.Links[0].Text != "Logo of Acme" {
		t.Errorf("image-only link should use alt text, got %q", m.Links[0].Text)
	}
	if !m.Links[1].Nofollow {
		t.Errorf("rel=nofollow not detected")
	}
	if m.Links[2].HasHref {
		t.Errorf("onclick link should have HasHref=false")
	}

	if strings.Contains(strings.ToLower(page), "icon title") && m.Title == "icon title" {
		t.Errorf("svg <title> must not become the document title")
	}
	// Body text excludes script, style and noscript.
	if m.WordCount != 14 {
		t.Errorf("word count = %d, want 14", m.WordCount)
	}
}

func TestExtractEmpty(t *testing.T) {
	m, err := Extract(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "" || m.WordCount != 0 {
		t.Errorf("unexpected data from empty document: %+v", m)
	}
}
