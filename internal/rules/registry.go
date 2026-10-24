package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// Registry holds all loaded deterministic rules.
type Registry struct {
	rules     []Rule
	siteRules []SiteRule
}

// NewRegistry initializes the registry with all standard rules.
func NewRegistry() *Registry {
	r := &Registry{}

	r.rules = append(r.rules,
		// --- Availability & Crawlability ---
		&HTTP5xxRule{},         // RC-HTTP-001
		&HTTP4xxRule{},         // RC-HTTP-002
		&HTTPSuccessRule{},     // RC-HTTP-003
		&RobotsCrawlableRule{}, // RC-HTTP-004
		&RedirectChainRule{},   // RC-HTTP-005

		// --- Indexability ---
		&MetaNoindexRule{}, // RC-INDEX-001

		// --- Discovery & Links ---
		&LinkTextRule{},        // RC-LINK-002
		&CrawlableAnchorRule{}, // RC-LINK-003

		// --- Metadata & Headings ---
		&MissingTitleRule{},    // RC-META-001
		&MissingMetaDescRule{}, // RC-META-002
		&MissingViewportRule{}, // RC-META-003
		&TitleQualityRule{},    // RC-META-004
		&H1StructureRule{},     // RC-META-005

		// --- Content & Images ---
		&ImageAltRule{},    // RC-IMG-001
		&ThinContentRule{}, // RC-CONTENT-001

		// --- Canonical ---
		&MissingCanonicalRule{},  // RC-CANON-001
		&CanonicalProtocolRule{}, // RC-CANON-003
		&MultipleCanonicalRule{}, // RC-CANON-004

		// --- Internationalization ---
		&HreflangValidRule{}, // RC-HREFLANG-001
		&HTMLLangRule{},      // RC-LANG-001

		// --- Social Metadata ---
		&OpenGraphRule{}, // RC-SOCIAL-001

		// --- Structured Data ---
		&StructuredDataValidRule{},    // RC-STRUCT-001
		&StructuredDataRequiredRule{}, // RC-STRUCT-002

		// --- Renderability ---
		&RenderDiscrepancyRule{}, // RC-RENDER-001
	)

	r.siteRules = append(r.siteRules,
		&BrokenLinkRule{},           // RC-LINK-001
		&RedirectingLinkRule{},      // RC-LINK-004
		&CanonicalTargetRule{},      // RC-CANON-002
		&OrphanPageRule{},           // RC-GRAPH-001
		&CrawlDepthRule{},           // RC-GRAPH-002
		&DuplicateTitleRule{},       // RC-META-006
		&DuplicateDescriptionRule{}, // RC-META-007
		&SitemapPresenceRule{},      // RC-SITEMAP-001
		&SitemapEntryRule{},         // RC-SITEMAP-002
		&HreflangReturnRule{},       // RC-HREFLANG-002
		&AISearchAccessRule{},       // RC-AI-001
	)

	return r
}

// Count returns the number of registered rules.
func (r *Registry) Count() int {
	return len(r.rules) + len(r.siteRules)
}

// All returns every registered rule, page and site level, sorted by ID.
func (r *Registry) All() []Identified {
	all := make([]Identified, 0, r.Count())
	for _, x := range r.rules {
		all = append(all, x)
	}
	for _, x := range r.siteRules {
		all = append(all, x)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID() < all[j].ID() })
	return all
}

// Find retrieves a rule by ID (case-insensitive).
func (r *Registry) Find(id string) Identified {
	for _, rule := range r.All() {
		if strings.EqualFold(rule.ID(), id) {
			return rule
		}
	}
	return nil
}

// EvaluateAll runs all page rules against a page snapshot.
func (r *Registry) EvaluateAll(snapshot *crawl.PageSnapshot) []report.Finding {
	var findings []report.Finding
	for _, rule := range r.rules {
		if snapshot.Skipped != "" && rule.ID() != "RC-HTTP-004" {
			continue // nothing was fetched; only the robots rule applies
		}
		if finding := rule.Evaluate(snapshot); finding != nil {
			findings = append(findings, normalize(*finding))
		}
	}
	return findings
}

// EvaluateSite runs every page rule on every page, then every site rule, and
// returns findings in a stable order.
func (r *Registry) EvaluateSite(site *Site) []report.Finding {
	var findings []report.Finding
	for _, p := range site.Pages {
		findings = append(findings, r.EvaluateAll(p)...)
	}
	for _, rule := range r.siteRules {
		for _, f := range rule.EvaluateSite(site) {
			findings = append(findings, normalize(f))
		}
	}
	SortFindings(findings)
	return findings
}

// SortFindings orders findings by severity (worst first), then rule ID, then URL.
func SortFindings(f []report.Finding) {
	sort.SliceStable(f, func(i, j int) bool {
		si, sj := report.SeverityRank(f[i].Severity), report.SeverityRank(f[j].Severity)
		if si != sj {
			return si > sj
		}
		if f[i].ID != f[j].ID {
			return f[i].ID < f[j].ID
		}
		return f[i].URL < f[j].URL
	})
}

// normalize applies catalog metadata (category, safety class, verify contract)
// so every finding follows the master-plan schema regardless of rule author.
func normalize(f report.Finding) report.Finding {
	doc, ok := ruleDocs[f.ID]
	if !ok {
		return f
	}
	f.Category = doc.Category
	if f.Automation == "" {
		f.Automation = doc.Automation
	}
	if f.Verify == nil && doc.VerifyType != "" {
		f.Verify = &report.VerifyInfo{Type: doc.VerifyType, Expected: doc.VerifyExpected}
	}
	if f.Evidence == nil {
		f.Evidence = map[string]interface{}{}
	}
	return f
}

type ruleDoc struct {
	Name           string
	Category       string
	Severity       string
	Automation     string
	VerifyType     string
	VerifyExpected string
	Description    string
	Why            string
	Remediation    string
}

const (
	safe   = report.AutomationSafe
	review = report.AutomationReview
	manual = report.AutomationManual
)

var ruleDocs = map[string]ruleDoc{
	"RC-HTTP-001": {
		Name: "HTTP 5xx Server Error", Category: "availability", Severity: "critical", Automation: review,
		VerifyType: "status", VerifyExpected: "200",
		Description: "The server encountered an error while processing the request and returned a 5xx HTTP response.",
		Why:         "Search engine crawlers and users cannot access the page. Persistent 5xx errors cause drops from search indices.",
		Remediation: "Check application logs, backend services, and server configuration to resolve the 5xx server error.",
	},
	"RC-HTTP-002": {
		Name: "HTTP 4xx Client Error", Category: "availability", Severity: "high", Automation: review,
		VerifyType: "status", VerifyExpected: "200",
		Description: "The requested URL returned a 4xx client error (such as 404 Not Found or 403 Forbidden).",
		Why:         "Crawlers encountering 4xx errors will drop the page or fail to discover linked content, degrading SEO health.",
		Remediation: "Fix broken links, restore the missing resource, or set up a 301 redirect to the appropriate replacement URL.",
	},
	"RC-HTTP-003": {
		Name: "Non-Successful HTTP Status Code", Category: "crawlability", Severity: "high", Automation: review,
		VerifyType: "status", VerifyExpected: "200",
		Description: "The page could not be fetched, or answered with a redirect status that has no destination.",
		Why:         "Search engines require successful 200 OK responses to index and rank primary content.",
		Remediation: "Ensure all canonical URLs return HTTP 200 OK directly without unnecessary errors or redirect chains.",
	},
	"RC-HTTP-004": {
		Name: "Page Blocked by Robots.txt", Category: "crawlability", Severity: "high", Automation: review,
		VerifyType: "robots_access", VerifyExpected: "allowed",
		Description: "The URL path is blocked from crawling for Googlebot by rules in robots.txt.",
		Why:         "Search crawlers respect robots.txt disallow rules and will not fetch or analyze page content.",
		Remediation: "Update robots.txt to remove the disallow directive if this page is intended for public search indexing.",
	},
	"RC-HTTP-005": {
		Name: "Redirect Chains and Loops", Category: "crawlability", Severity: "high", Automation: safe,
		VerifyType: "redirect_hops", VerifyExpected: "<=2",
		Description: "The URL triggered a circular redirect loop or exceeded recommended hop limits (> 2 hops).",
		Why:         "Redirect loops trap crawlers and users. Long chains delay page load times and dilute crawl equity.",
		Remediation: "Update redirect configurations and internal links to point directly to the final 200 OK destination.",
	},
	"RC-INDEX-001": {
		Name: "Page Has Noindex", Category: "indexability", Severity: "high", Automation: review,
		VerifyType: "noindex", VerifyExpected: "false",
		Description: "The page includes a 'noindex' directive in robots meta tags or X-Robots-Tag HTTP headers.",
		Why:         "Search engines will not index or rank any page with an explicit noindex directive. This is correct for private or utility pages and a serious bug on public ones.",
		Remediation: "Remove the 'noindex' directive from meta tags or HTTP response headers if the page should be indexed.",
	},
	"RC-CANON-001": {
		Name: "Missing Canonical Link Tag", Category: "canonical", Severity: "medium", Automation: review,
		VerifyType: "canonical_present", VerifyExpected: "true",
		Description: "The document is missing a <link rel='canonical'> tag in the head section.",
		Why:         "Canonical tags establish the authoritative URL version, preventing duplicate content dilution across query parameters or protocols.",
		Remediation: "Add a `<link rel=\"canonical\" href=\"https://...\">` tag in the `<head>` specifying the authoritative URL.",
	},
	"RC-CANON-002": {
		Name: "Canonical Target Issue", Category: "canonical", Severity: "high", Automation: safe,
		VerifyType: "canonical_target_status", VerifyExpected: "200",
		Description: "The canonical target is malformed, relative, broken, redirecting, noindexed, or itself canonicalizes elsewhere.",
		Why:         "Canonical targets must be valid, absolute, indexable 200 OK URLs or search engines ignore the hint.",
		Remediation: "Ensure the canonical href is an absolute URL pointing to a live, indexable 200 OK page.",
	},
	"RC-CANON-003": {
		Name: "Canonical Protocol/Host Mismatch", Category: "canonical", Severity: "high", Automation: review,
		VerifyType: "canonical_scheme", VerifyExpected: "https",
		Description: "A secure HTTPS page designates an insecure HTTP canonical target or a different host.",
		Why:         "Specifying HTTP canonicals on HTTPS sites risks downgrading security signaling and confusing search indices.",
		Remediation: "Ensure canonical href uses https:// and points to the authoritative canonical hostname.",
	},
	"RC-CANON-004": {
		Name: "Multiple Conflicting Canonicals", Category: "canonical", Severity: "high", Automation: safe,
		VerifyType: "canonical_count", VerifyExpected: "1",
		Description: "The page declares more than one distinct rel=canonical URL.",
		Why:         "When canonicals conflict, search engines ignore them all and pick a canonical themselves.",
		Remediation: "Emit exactly one canonical tag per page.",
	},
	"RC-LINK-001": {
		Name: "Broken Internal Link", Category: "discovery", Severity: "medium", Automation: safe,
		VerifyType: "link_status", VerifyExpected: "200",
		Description: "The page links to internal URLs that return 4xx/5xx or cannot be fetched.",
		Why:         "Broken links waste crawl budget, harm user experience, and disrupt internal link flow.",
		Remediation: "Update or remove the broken href, or restore/redirect the destination.",
	},
	"RC-LINK-002": {
		Name: "Non-Descriptive Link Text", Category: "discovery", Severity: "medium", Automation: safe,
		VerifyType: "link_text", VerifyExpected: "descriptive",
		Description: "Links use generic anchor text like 'click here', 'read more', or are completely empty.",
		Why:         "Descriptive anchor text tells search engines and screen readers what the destination is about.",
		Remediation: "Replace generic text with clear descriptions of the linked target page.",
	},
	"RC-LINK-003": {
		Name: "Uncrawlable Anchor Link", Category: "discovery", Severity: "high", Automation: safe,
		VerifyType: "href", VerifyExpected: "valid_url",
		Description: "The link uses an empty href, href='#', a javascript: URL, or an onclick handler without href.",
		Why:         "Search crawlers follow href attributes; they do not click elements or run navigation handlers.",
		Remediation: "Use standard `<a href=\"/path\">` elements with valid HTTP/HTTPS URLs for all navigational links.",
	},
	"RC-LINK-004": {
		Name: "Internal Link to Redirect", Category: "discovery", Severity: "low", Automation: safe,
		VerifyType: "link_redirects", VerifyExpected: "0",
		Description: "The page links to internal URLs that redirect before reaching content.",
		Why:         "Each redirect costs a request for crawlers and users and weakens the signal of the link.",
		Remediation: "Link directly to the final URL.",
	},
	"RC-META-001": {
		Name: "Missing Document Title", Category: "metadata", Severity: "high", Automation: safe,
		VerifyType: "title_present", VerifyExpected: "true",
		Description: "The document does not have a `<title>` tag inside `<head>` or the title is completely empty.",
		Why:         "Title tags are primary relevance signals and usually become the headline in search results.",
		Remediation: "Add a unique, descriptive `<title>` tag (recommended 30-60 characters) inside the `<head>`.",
	},
	"RC-META-002": {
		Name: "Missing Meta Description", Category: "metadata", Severity: "medium", Automation: review,
		VerifyType: "meta_desc_present", VerifyExpected: "true",
		Description: "The page lacks a `<meta name='description'>` tag or the content attribute is empty.",
		Why:         "Meta descriptions frequently generate the search snippet and influence click-through rate.",
		Remediation: "Add `<meta name=\"description\" content=\"...\">` summarizing page content within 120-160 characters.",
	},
	"RC-META-003": {
		Name: "Missing Mobile Viewport Meta Tag", Category: "metadata", Severity: "high", Automation: safe,
		VerifyType: "viewport", VerifyExpected: "width=device-width",
		Description: "The page lacks a `<meta name='viewport'>` tag configuring width.",
		Why:         "Search engines use mobile-first indexing. Pages without viewport tags render poorly on phones.",
		Remediation: "Include `<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">` in `<head>`.",
	},
	"RC-META-004": {
		Name: "Title Length & Quality", Category: "metadata", Severity: "medium", Automation: review,
		VerifyType: "title_length", VerifyExpected: "10..70",
		Description: "The title tag is too short (<10 chars), too long (>70 chars), or uses generic placeholder text (e.g. 'Home').",
		Why:         "Clear, accurately sized titles prevent truncation and give users and crawlers specific subject context.",
		Remediation: "Craft concise titles between 30 and 60 characters that highlight the page topic and brand.",
	},
	"RC-META-005": {
		Name: "H1 Heading Structure", Category: "content", Severity: "medium", Automation: safe,
		VerifyType: "h1_count", VerifyExpected: "1",
		Description: "The document does not contain an <h1> heading or contains multiple <h1> headings.",
		Why:         "The <h1> heading communicates the primary topic of the document to search engines and screen readers.",
		Remediation: "Include exactly one top-level <h1> heading per page describing the main topic.",
	},
	"RC-META-006": {
		Name: "Duplicate Title", Category: "metadata", Severity: "medium", Automation: review,
		VerifyType: "title_unique", VerifyExpected: "true",
		Description: "Several indexable pages share exactly the same title.",
		Why:         "Duplicate titles make pages hard to tell apart in results and often signal duplicate or templated content.",
		Remediation: "Write a unique title per page, or canonicalize true duplicates.",
	},
	"RC-META-007": {
		Name: "Duplicate Meta Description", Category: "metadata", Severity: "low", Automation: review,
		VerifyType: "meta_desc_unique", VerifyExpected: "true",
		Description: "Several indexable pages share exactly the same meta description.",
		Why:         "Repeated descriptions are usually ignored and replaced by search engines, and indicate templated metadata.",
		Remediation: "Write a description specific to each page.",
	},
	"RC-IMG-001": {
		Name: "Images Missing Alt Attributes", Category: "accessibility", Severity: "medium", Automation: review,
		VerifyType: "img_alt", VerifyExpected: "present",
		Description: "One or more `<img>` elements do not have an `alt` attribute.",
		Why:         "Alt attributes make visual content accessible to screen readers and help search engines understand images.",
		Remediation: "Provide meaningful descriptive `alt` text for images, or `alt=\"\"` for purely decorative images.",
	},
	"RC-CONTENT-001": {
		Name: "Thin Body Content", Category: "content", Severity: "medium", Automation: manual,
		VerifyType: "word_count", VerifyExpected: ">=200",
		Description: "An indexable page has fewer than 200 words of visible body text.",
		Why:         "Pages with sparse text may be treated as low value. Padding text is not the fix; real substance or consolidation is.",
		Remediation: "Expand the page with genuinely useful content, or noindex/merge it if it is a utility page.",
	},
	"RC-LANG-001": {
		Name: "HTML Lang Attribute Missing or Invalid", Category: "internationalization", Severity: "medium", Automation: safe,
		VerifyType: "html_lang", VerifyExpected: "valid_bcp47",
		Description: "The <html> element does not specify a valid language code in its lang attribute.",
		Why:         "Language attributes support correct rendering, accessibility, and language targeting.",
		Remediation: "Add a valid BCP 47 language code to the <html> tag, e.g. <html lang=\"en\">.",
	},
	"RC-HREFLANG-001": {
		Name: "Invalid Hreflang Tag", Category: "internationalization", Severity: "medium", Automation: safe,
		VerifyType: "hreflang_valid", VerifyExpected: "true",
		Description: "The document has hreflang tags with invalid language codes, empty targets, or relative URLs.",
		Why:         "Malformed hreflang tags cause search engines to ignore international targeting directives.",
		Remediation: "Use valid ISO 639-1 language codes (and optional ISO 3166-1 region codes) with absolute URLs.",
	},
	"RC-HREFLANG-002": {
		Name: "Hreflang Missing Return or Self Link", Category: "internationalization", Severity: "medium", Automation: safe,
		VerifyType: "hreflang_reciprocal", VerifyExpected: "true",
		Description: "An hreflang alternate does not link back, or the page omits itself from its own hreflang set.",
		Why:         "Search engines ignore hreflang pairs that are not confirmed in both directions.",
		Remediation: "List every alternate, including the page itself, on every page in the cluster.",
	},
	"RC-SOCIAL-001": {
		Name: "Missing Open Graph Metadata", Category: "social", Severity: "low", Automation: safe,
		VerifyType: "og_present", VerifyExpected: "true",
		Description: "The page is missing essential Open Graph tags (og:title, og:image, or og:description).",
		Why:         "Open Graph tags control link previews in social networks and chat apps. They are not a ranking factor.",
		Remediation: "Add <meta property=\"og:title\">, <meta property=\"og:image\">, and <meta property=\"og:description\"> tags in the <head>.",
	},
	"RC-STRUCT-001": {
		Name: "Invalid Structured Data (JSON-LD)", Category: "structured_data", Severity: "medium", Automation: safe,
		VerifyType: "json_ld_valid", VerifyExpected: "true",
		Description: "The page contains `<script type='application/ld+json'>` blocks that fail JSON parsing.",
		Why:         "Malformed JSON prevents search engines from reading any of the structured data in the block.",
		Remediation: "Validate and format all JSON-LD blocks to ensure strict valid JSON syntax.",
	},
	"RC-STRUCT-002": {
		Name: "Structured Data Missing Required Properties", Category: "structured_data", Severity: "medium", Automation: review,
		VerifyType: "json_ld_required", VerifyExpected: "present",
		Description: "A supported rich-result type (Product, Event, Recipe, JobPosting, BreadcrumbList, FAQPage, LocalBusiness, VideoObject, SoftwareApplication) lacks a property Google documents as required, or an entity has no @type.",
		Why:         "Items missing required properties are not eligible for rich results.",
		Remediation: "Add the required properties using values that are visible on the page. Never fabricate ratings, reviews, or prices.",
	},
	"RC-SITEMAP-001": {
		Name: "Sitemap Missing or Unreadable", Category: "discovery", Severity: "low", Automation: review,
		VerifyType: "sitemap_readable", VerifyExpected: "true",
		Description: "No XML sitemap was found, or a sitemap declared in robots.txt cannot be fetched or parsed.",
		Why:         "Sitemaps help search engines discover canonical URLs, especially on large or poorly linked sites.",
		Remediation: "Publish a valid XML sitemap of canonical URLs and reference it from robots.txt.",
	},
	"RC-SITEMAP-002": {
		Name: "Non-Indexable URL in Sitemap", Category: "discovery", Severity: "medium", Automation: safe,
		VerifyType: "sitemap_entry_indexable", VerifyExpected: "true",
		Description: "A sitemap lists a URL that errors, redirects, is noindexed, canonicalizes elsewhere, or is blocked by robots.txt.",
		Why:         "Sitemaps should only contain canonical URLs you want indexed; contradictions waste crawl budget and weaken trust in the sitemap.",
		Remediation: "Remove or replace such entries with their canonical, indexable, 200 OK equivalents.",
	},
	"RC-GRAPH-001": {
		Name: "Orphan Page", Category: "discovery", Severity: "medium", Automation: review,
		VerifyType: "inbound_links", VerifyExpected: ">0",
		Description: "An indexable page listed in the sitemap has no internal links from any crawled page.",
		Why:         "Orphan pages are hard for crawlers to find and re-crawl, and receive no internal link signals.",
		Remediation: "Add internal links to this page from relevant parent pages or site navigation.",
	},
	"RC-GRAPH-002": {
		Name: "Excessive Crawl Depth", Category: "discovery", Severity: "low", Automation: review,
		VerifyType: "depth", VerifyExpected: "<=5",
		Description: "The page is more than 5 clicks away from the start URL.",
		Why:         "Deep pages tend to be crawled less often and receive fewer internal link signals.",
		Remediation: "Reduce click depth by linking to key pages from category pages or navigation menus.",
	},
	"RC-RENDER-001": {
		Name: "Render Discrepancy (Client-Side Rendering)", Category: "renderability", Severity: "high", Automation: review,
		VerifyType: "render_match", VerifyExpected: "true",
		Description: "A material difference exists between raw server HTML and the rendered DOM (JavaScript injecting noindex, changing canonical, or supplying all content).",
		Why:         "Crawlers may index the raw server HTML before rendering, or not render at all.",
		Remediation: "Ensure initial server HTML contains essential metadata, canonical tags, and core text, or implement SSR/SSG.",
	},
	"RC-AI-001": {
		Name: "Search or AI-Search Crawler Blocked", Category: "crawlability", Severity: "high", Automation: review,
		VerifyType: "crawler_access", VerifyExpected: "allowed",
		Description: "robots.txt blocks a search crawler (Googlebot, Bingbot, OAI-SearchBot, Claude-SearchBot, PerplexityBot…) or a user-initiated AI fetcher from the site root. Training crawlers are never flagged.",
		Why:         "Blocking search crawlers removes the site from that engine's results and AI answers. Training crawlers are a separate publisher choice.",
		Remediation: "Allow the search crawlers you want visibility in; keep training crawlers blocked if you prefer.",
	},
}

// Documentation formats a comprehensive explanation of a rule for CLI display.
func Documentation(r Identified) string {
	doc, ok := ruleDocs[r.ID()]
	if !ok {
		return fmt.Sprintf("Rule: %s (v%d)\nNo detailed documentation available.\n", r.ID(), r.Version())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s (%s)\n\n", doc.Name, r.ID()))
	sb.WriteString(fmt.Sprintf("**Category:**    %s\n", doc.Category))
	sb.WriteString(fmt.Sprintf("**Severity:**    %s\n", doc.Severity))
	sb.WriteString(fmt.Sprintf("**Automation:**  %s\n", doc.Automation))
	sb.WriteString(fmt.Sprintf("**Version:**     v%d\n\n", r.Version()))
	sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", doc.Description))
	sb.WriteString(fmt.Sprintf("## Why It Matters\n%s\n\n", doc.Why))
	sb.WriteString(fmt.Sprintf("## How to Fix\n%s\n", doc.Remediation))

	return sb.String()
}

// Name returns the human-readable name of a rule.
func Name(id string) string {
	return ruleDocs[id].Name
}
