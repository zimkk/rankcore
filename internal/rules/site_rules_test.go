package rules

import (
	"strings"
	"testing"

	"github.com/zimkk/rankcore/internal/aiaccess"
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/robots"
)

func page(u string, status int) *crawl.PageSnapshot {
	return &crawl.PageSnapshot{URL: u, FinalURL: u, StatusCode: status, ContentType: "text/html"}
}

func ids(fs []report.Finding) map[string][]string {
	out := map[string][]string{}
	for _, f := range fs {
		out[f.ID] = append(out[f.ID], f.URL)
	}
	return out
}

func TestBrokenAndRedirectingLinks(t *testing.T) {
	home := page("https://s.test/", 200)
	home.InternalLinks = []string{"https://s.test/missing", "https://s.test/old", "https://s.test/ok"}
	missing := page("https://s.test/missing", 404)
	old := page("https://s.test/old", 200)
	old.RedirectChain = []string{"https://s.test/old"}
	old.FinalURL = "https://s.test/new"
	site := NewSite("https://s.test/", []*crawl.PageSnapshot{home, missing, old, page("https://s.test/ok", 200)})

	broken := (&BrokenLinkRule{}).EvaluateSite(site)
	if len(broken) != 1 || broken[0].URL != home.URL {
		t.Fatalf("expected one broken-link finding on home, got %+v", broken)
	}
	redir := (&RedirectingLinkRule{}).EvaluateSite(site)
	if len(redir) != 1 || !strings.Contains(redir[0].Evidence["redirecting_links"].([]string)[0], "/new") {
		t.Fatalf("expected one redirecting-link finding, got %+v", redir)
	}
}

func TestCanonicalTargetRule(t *testing.T) {
	a := page("https://s.test/a", 200)
	a.Canonical = "https://s.test/gone"
	b := page("https://s.test/b", 200)
	b.Canonical = "https://s.test/noidx"
	c := page("https://s.test/c", 200)
	c.Canonical = "/c" // relative self-canonical
	d := page("https://s.test/d", 200)
	d.Canonical = "https://s.test/d" // self-canonical, fine
	noidx := page("https://s.test/noidx", 200)
	noidx.RobotsMeta = "noindex"
	site := NewSite("https://s.test/a", []*crawl.PageSnapshot{a, b, c, d, noidx, page("https://s.test/gone", 404)})

	got := ids((&CanonicalTargetRule{}).EvaluateSite(site))["RC-CANON-002"]
	want := map[string]bool{"https://s.test/a": true, "https://s.test/b": true, "https://s.test/c": true}
	if len(got) != len(want) {
		t.Fatalf("got findings on %v", got)
	}
	for _, u := range got {
		if !want[u] {
			t.Errorf("unexpected canonical finding on %s", u)
		}
	}
}

func TestMultipleCanonicalRule(t *testing.T) {
	p := page("https://s.test/a", 200)
	p.Canonicals = []string{"https://s.test/a", "/a"}
	if (&MultipleCanonicalRule{}).Evaluate(p) != nil {
		t.Error("two canonicals resolving to the same URL are not a conflict")
	}
	p.Canonicals = []string{"https://s.test/a", "https://s.test/b"}
	if (&MultipleCanonicalRule{}).Evaluate(p) == nil {
		t.Error("expected conflict finding")
	}
}

func TestOrphanAndDepth(t *testing.T) {
	var pages []*crawl.PageSnapshot
	// Chain of 8 pages: depth 0..7.
	for i := 0; i < 8; i++ {
		p := page("https://s.test/"+string(rune('a'+i)), 200)
		if i < 7 {
			p.InternalLinks = []string{"https://s.test/" + string(rune('a'+i+1))}
		}
		pages = append(pages, p)
	}
	orphan := page("https://s.test/orphan", 200)
	orphan.InSitemap = true
	linked := pages[2]
	linked.InSitemap = true
	pages = append(pages, orphan)
	site := NewSite("https://s.test/a", pages)

	orphans := (&OrphanPageRule{}).EvaluateSite(site)
	if len(orphans) != 1 || orphans[0].URL != orphan.URL {
		t.Errorf("expected only the sitemap-only page to be an orphan, got %+v", orphans)
	}
	deep := (&CrawlDepthRule{}).EvaluateSite(site)
	if len(deep) != 2 { // depths 6 and 7
		t.Errorf("expected 2 deep pages, got %d", len(deep))
	}
}

func TestDuplicateTitles(t *testing.T) {
	a, b, c := page("https://s.test/a", 200), page("https://s.test/b", 200), page("https://s.test/c", 200)
	a.Title, b.Title, c.Title = "Same  Title", "same title", "Other"
	n := page("https://s.test/n", 200)
	n.Title, n.RobotsMeta = "Same Title", "noindex" // noindexed pages don't count
	site := NewSite("https://s.test/a", []*crawl.PageSnapshot{a, b, c, n})
	if got := (&DuplicateTitleRule{}).EvaluateSite(site); len(got) != 2 {
		t.Errorf("expected 2 duplicate-title findings, got %+v", got)
	}
}

func TestSitemapRules(t *testing.T) {
	ok := page("https://s.test/ok", 200)
	ok.InSitemap = true
	gone := page("https://s.test/gone", 404)
	gone.InSitemap = true
	canon := page("https://s.test/dup", 200)
	canon.InSitemap = true
	canon.Canonical = "https://s.test/ok"
	site := NewSite("https://s.test/ok", []*crawl.PageSnapshot{ok, gone, canon})

	got := ids((&SitemapEntryRule{}).EvaluateSite(site))["RC-SITEMAP-002"]
	if len(got) != 2 {
		t.Errorf("expected findings on /gone and /dup, got %v", got)
	}

	if f := (&SitemapPresenceRule{}).EvaluateSite(site); len(f) != 1 || f[0].Severity != "low" {
		t.Errorf("expected a low 'no sitemap' finding, got %+v", f)
	}
	site.Sitemaps = []report.SitemapInfo{{URL: "https://s.test/sm.xml", Source: "robots.txt", Error: "HTTP 404"}}
	if f := (&SitemapPresenceRule{}).EvaluateSite(site); len(f) != 1 || f[0].Severity != "medium" {
		t.Errorf("expected a medium 'unreadable sitemap' finding, got %+v", f)
	}
}

func TestHreflangReturnRule(t *testing.T) {
	en := page("https://s.test/en", 200)
	en.Hreflang = map[string]string{"en": "https://s.test/en", "fr": "https://s.test/fr"}
	fr := page("https://s.test/fr", 200)
	fr.Hreflang = map[string]string{"fr": "https://s.test/fr"} // no return link to /en
	site := NewSite("https://s.test/en", []*crawl.PageSnapshot{en, fr})
	got := ids((&HreflangReturnRule{}).EvaluateSite(site))["RC-HREFLANG-002"]
	if len(got) != 1 || got[0] != en.URL {
		t.Errorf("expected missing return link reported on /en, got %v", got)
	}
}

func TestAISearchAccessRuleIgnoresTraining(t *testing.T) {
	rt := robots.Parse(strings.NewReader("User-agent: GPTBot\nDisallow: /\n\nUser-agent: CCBot\nDisallow: /\n"))
	site := NewSite("https://s.test/", nil)
	site.CrawlerAccess = aiaccess.Evaluate(rt, "/").Results
	if f := (&AISearchAccessRule{}).EvaluateSite(site); len(f) != 0 {
		t.Errorf("blocking training crawlers must not be flagged: %+v", f)
	}

	rt = robots.Parse(strings.NewReader("User-agent: OAI-SearchBot\nDisallow: /\n"))
	site.CrawlerAccess = aiaccess.Evaluate(rt, "/").Results
	f := (&AISearchAccessRule{}).EvaluateSite(site)
	if len(f) != 1 || f[0].Severity != "high" {
		t.Errorf("blocking a search crawler must be flagged high, got %+v", f)
	}
}

func TestStructuredDataRequiredRule(t *testing.T) {
	r := &StructuredDataRequiredRule{}
	bad := page("https://s.test/p", 200)
	bad.JSONLD = []string{`{"@context":"https://schema.org","@graph":[{"@type":"Product","name":"W"},{"@type":"Event","name":"E"}]}`}
	f := r.Evaluate(bad)
	if f == nil {
		t.Fatal("expected missing-property finding")
	}
	problems := f.Evidence["problems"].([]string)
	if len(problems) != 3 { // Product offers/review/rating, Event startDate, Event location
		t.Errorf("problems = %v", problems)
	}

	good := page("https://s.test/q", 200)
	good.JSONLD = []string{`[{"@type":"https://schema.org/Product","name":"W","offers":{"price":"1"}},{"@type":"Organization","name":"Acme"}]`}
	if f := r.Evaluate(good); f != nil {
		t.Errorf("complete Product should pass: %+v", f)
	}
}

func TestRegistryNormalizesAndSorts(t *testing.T) {
	reg := NewRegistry()
	broken := page("https://s.test/", 500)
	findings := reg.EvaluateSite(NewSite("https://s.test/", []*crawl.PageSnapshot{broken}))
	if len(findings) == 0 || findings[0].Severity != "critical" {
		t.Fatalf("critical findings should sort first: %+v", findings)
	}
	for _, f := range findings {
		if f.Automation == "" || f.Category == "" {
			t.Errorf("%s missing catalog metadata", f.ID)
		}
		if _, ok := ruleDocs[f.ID]; !ok {
			t.Errorf("%s has no documentation", f.ID)
		}
	}
	for _, r := range reg.All() {
		if _, ok := ruleDocs[r.ID()]; !ok {
			t.Errorf("registered rule %s has no documentation", r.ID())
		}
	}
}
