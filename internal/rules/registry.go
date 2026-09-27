package rules

import (
	"fmt"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// Registry holds all loaded deterministic rules.
type Registry struct {
	rules []Rule
}

// NewRegistry initializes the registry with all standard rules.
func NewRegistry() *Registry {
	r := &Registry{}

	// Register all rules
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
		&BrokenLinkRule{},      // RC-LINK-001
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
		&CanonicalTargetRule{},   // RC-CANON-002
		&CanonicalProtocolRule{}, // RC-CANON-003

		// --- Internationalization ---
		&HreflangValidRule{}, // RC-HREFLANG-001
		&HTMLLangRule{},      // RC-LANG-001

		// --- Social Metadata ---
		&OpenGraphRule{}, // RC-SOCIAL-001

		// --- Structured Data ---
		&StructuredDataValidRule{}, // RC-STRUCT-001

		// --- Renderability ---
		&RenderDiscrepancyRule{}, // RC-RENDER-001
	)

	return r
}

// Count returns the number of registered rules.
func (r *Registry) Count() int {
	return len(r.rules)
}

// All returns a slice of all registered rules.
func (r *Registry) All() []Rule {
	return r.rules
}

// Find retrieves a rule by ID (case-insensitive).
func (r *Registry) Find(id string) Rule {
	for _, rule := range r.rules {
		if strings.EqualFold(rule.ID(), id) {
			return rule
		}
	}
	return nil
}

// EvaluateAll runs all registered rules against a page snapshot.
func (r *Registry) EvaluateAll(snapshot *crawl.PageSnapshot) []report.Finding {
	var findings []report.Finding

	for _, rule := range r.rules {
		if finding := rule.Evaluate(snapshot); finding != nil {
			findings = append(findings, *finding)
		}
	}

	return findings
}

type ruleDoc struct {
	Name        string
	Category    string
	Severity    string
	Description string
	Why         string
	Remediation string
}

var ruleDocs = map[string]ruleDoc{
	"RC-HTTP-001": {
		Name:        "HTTP 5xx Server Error",
		Category:    "availability",
		Severity:    "critical",
		Description: "The server encountered an error while processing the request and returned a 5xx HTTP response.",
		Why:         "Search engine crawlers and users cannot access the page. Persistent 5xx errors cause drops from search indices.",
		Remediation: "Check application logs, backend services, and server configuration to resolve the 5xx server error.",
	},
	"RC-HTTP-002": {
		Name:        "HTTP 4xx Client Error",
		Category:    "availability",
		Severity:    "high",
		Description: "The requested URL returned a 4xx client error (such as 404 Not Found or 403 Forbidden).",
		Why:         "Crawlers encountering 4xx errors will drop the page or fail to discover linked content, degrading SEO health.",
		Remediation: "Fix broken links, restore the missing resource, or set up a 301 redirect to the appropriate replacement URL.",
	},
	"RC-HTTP-003": {
		Name:        "Non-Successful HTTP Status Code",
		Category:    "crawlability",
		Severity:    "high",
		Description: "The page returned an HTTP status code outside the 200 OK range.",
		Why:         "Search engines require successful 200 OK responses to index and rank primary content.",
		Remediation: "Ensure all canonical URLs return HTTP 200 OK directly without unnecessary errors or redirect chains.",
	},
	"RC-HTTP-004": {
		Name:        "Page Blocked by Robots.txt",
		Category:    "crawlability",
		Severity:    "high",
		Description: "The URL path is blocked from crawling by rules in robots.txt.",
		Why:         "Search crawlers respect robots.txt disallow rules and will not fetch or analyze page content.",
		Remediation: "Update robots.txt to remove the disallow directive if this page is intended for public search indexing.",
	},
	"RC-HTTP-005": {
		Name:        "Redirect Chains and Loops",
		Category:    "crawlability",
		Severity:    "high",
		Description: "The URL triggered a circular redirect loop or exceeded recommended hop limits (> 2 hops).",
		Why:         "Redirect loops trap crawlers and users. Long chains delay page load times and dilute crawl equity.",
		Remediation: "Update redirect configurations and internal links to point directly to the final 200 OK destination.",
	},
	"RC-INDEX-001": {
		Name:        "Public Page Has Meta Noindex",
		Category:    "indexability",
		Severity:    "high",
		Description: "The page includes a 'noindex' directive in robots meta tags or X-Robots-Tag HTTP headers.",
		Why:         "Search engines will not index or rank any page with an explicit noindex directive.",
		Remediation: "Remove the 'noindex' directive from meta tags or HTTP response headers if the page should be indexed.",
	},
	"RC-CANON-001": {
		Name:        "Missing Canonical Link Tag",
		Category:    "canonical",
		Severity:    "medium",
		Description: "The document is missing a <link rel='canonical'> tag in the head section.",
		Why:         "Canonical tags establish the authoritative URL version, preventing duplicate content dilution across query parameters or protocols.",
		Remediation: "Add a `<link rel=\"canonical\" href=\"https://...\">` tag in the `<head>` specifying the authoritative URL.",
	},
	"RC-CANON-002": {
		Name:        "Canonical Target Status Issue",
		Category:    "canonical",
		Severity:    "high",
		Description: "The canonical target points to a non-existent, broken, relative, or redirecting URL.",
		Why:         "Canonical targets must be valid, absolute, canonical 200 OK URLs to prevent search engines from ignoring them.",
		Remediation: "Ensure the canonical href is an absolute URL pointing to a live 200 OK page.",
	},
	"RC-CANON-003": {
		Name:        "Canonical Protocol/Host Mismatch",
		Category:    "canonical",
		Severity:    "high",
		Description: "A secure HTTPS page designates an insecure HTTP canonical target or mismatched external domain.",
		Why:         "Specifying HTTP canonicals on HTTPS sites risks downgrading security signaling and confusing search indices.",
		Remediation: "Ensure canonical href uses https:// and points to the authoritative canonical hostname.",
	},
	"RC-LINK-001": {
		Name:        "Broken Internal Link",
		Category:    "discovery",
		Severity:    "medium",
		Description: "The page contains an internal link pointing to a broken or 404 destination.",
		Why:         "Broken links waste crawl budget, harm user experience, and disrupt PageRank flow.",
		Remediation: "Update or remove the broken href attribute to point to a valid internal destination.",
	},
	"RC-LINK-002": {
		Name:        "Non-Descriptive Link Text",
		Category:    "discovery",
		Severity:    "medium",
		Description: "Links use generic anchor text like 'click here', 'read more', or are completely empty.",
		Why:         "Descriptive anchor text provides essential contextual keywords to search engines about the destination page.",
		Remediation: "Replace generic text with clear, keyword-relevant descriptions of the linked target page.",
	},
	"RC-LINK-003": {
		Name:        "Uncrawlable Anchor Link",
		Category:    "discovery",
		Severity:    "high",
		Description: "The link uses an empty href, href='#', or javascript: pseudo-protocol instead of a real destination URL.",
		Why:         "Search crawlers do not execute JavaScript navigation onclick handlers and cannot follow non-standard anchors.",
		Remediation: "Use standard `<a href=\"/path\">` elements with valid HTTP/HTTPS URLs for all navigational links.",
	},
	"RC-META-001": {
		Name:        "Missing Document Title",
		Category:    "metadata",
		Severity:    "high",
		Description: "The document does not have a `<title>` tag inside `<head>` or the title is completely empty.",
		Why:         "Title tags are primary ranking signals and define the headline displayed in search engine results pages (SERPs).",
		Remediation: "Add a unique, descriptive `<title>` tag (recommended 30-60 characters) inside the `<head>`.",
	},
	"RC-META-002": {
		Name:        "Missing Meta Description",
		Category:    "metadata",
		Severity:    "medium",
		Description: "The page lacks a `<meta name='description'>` tag or the content attribute is empty.",
		Why:         "Meta descriptions frequently generate the SERP snippet and directly impact user click-through rate.",
		Remediation: "Add `<meta name=\"description\" content=\"...\">` summarizing page content within 120-160 characters.",
	},
	"RC-META-003": {
		Name:        "Missing Mobile Viewport Meta Tag",
		Category:    "metadata",
		Severity:    "high",
		Description: "The page lacks a `<meta name='viewport'>` tag configuring width or initial-scale.",
		Why:         "Search engines use mobile-first indexing. Pages without viewport tags fail mobile-friendliness criteria.",
		Remediation: "Include `<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">` in `<head>`.",
	},
	"RC-META-004": {
		Name:        "Title Length & Quality",
		Category:    "metadata",
		Severity:    "medium",
		Description: "The title tag is too short (<10 chars), too long (>70 chars), or uses generic placeholder text (e.g. 'Home').",
		Why:         "Clear, accurately sized titles prevent SERP clipping and give users and crawlers specific subject context.",
		Remediation: "Craft concise titles between 30 and 60 characters that highlight the page topic and brand.",
	},
	"RC-META-005": {
		Name:        "H1 Heading Structure",
		Category:    "content",
		Severity:    "medium",
		Description: "The document does not contain an <h1> heading or contains multiple conflicting <h1> headings.",
		Why:         "The <h1> heading communicates the primary topic of the document to search engines and screen readers.",
		Remediation: "Include exactly one top-level <h1> heading per page describing the main topic.",
	},
	"RC-IMG-001": {
		Name:        "Images Missing Alt Attributes",
		Category:    "accessibility",
		Severity:    "medium",
		Description: "One or more `<img>` elements do not have an `alt` attribute.",
		Why:         "Alt attributes make visual content accessible to screen readers and allow search engines to understand image context.",
		Remediation: "Provide meaningful descriptive `alt` text for images, or `alt=\"\"` for purely decorative images.",
	},
	"RC-CONTENT-001": {
		Name:        "Thin Body Content",
		Category:    "content",
		Severity:    "medium",
		Description: "The page has fewer than 200 words of extracted textual body content.",
		Why:         "Pages with sparse text may be classified as thin content by search engines and struggle to rank.",
		Remediation: "Expand page copy with informative, original text relevant to the topic and user intent.",
	},
	"RC-LANG-001": {
		Name:        "HTML Lang Attribute Missing or Invalid",
		Category:    "internationalization",
		Severity:    "medium",
		Description: "The <html> element does not specify a valid language code in its lang attribute.",
		Why:         "Language attributes ensure proper rendering in localized browsers and accurate indexing in geo-targeted search.",
		Remediation: "Add a valid BCP 47 language code to the <html> tag, e.g. <html lang=\"en\">.",
	},
	"RC-HREFLANG-001": {
		Name:        "Invalid Hreflang Tag",
		Category:    "internationalization",
		Severity:    "medium",
		Description: "The document has hreflang tags with invalid language codes or malformed URLs.",
		Why:         "Malformed hreflang tags cause search engines to ignore international targeting directives.",
		Remediation: "Use valid ISO 639-1 language codes (and optional ISO 3166-1 country codes) with absolute URLs.",
	},
	"RC-SOCIAL-001": {
		Name:        "Missing Open Graph Metadata",
		Category:    "social",
		Severity:    "low",
		Description: "The page is missing essential Open Graph tags (og:title, og:image, or og:description).",
		Why:         "Open Graph tags ensure that shared links on social networks, Slack, and messaging apps display rich preview cards.",
		Remediation: "Add <meta property=\"og:title\">, <meta property=\"og:image\">, and <meta property=\"og:description\"> tags in the <head>.",
	},
	"RC-STRUCT-001": {
		Name:        "Invalid Structured Data (JSON-LD)",
		Category:    "structured_data",
		Severity:    "medium",
		Description: "The page contains `<script type='application/ld+json'>` blocks that fail JSON parsing.",
		Why:         "Malformed JSON prevents search engines from parsing rich snippets, Schema.org entities, and structured previews.",
		Remediation: "Validate and format all JSON-LD blocks to ensure strict valid JSON syntax.",
	},
	"RC-GRAPH-001": {
		Name:        "Orphan Page",
		Category:    "discovery",
		Severity:    "medium",
		Description: "The page has no internal inbound links from any other crawled page on the site.",
		Why:         "Orphan pages are difficult for search crawlers to find and re-crawl, and receive zero internal link equity.",
		Remediation: "Add internal links to this page from relevant parent pages or site navigation.",
	},
	"RC-GRAPH-002": {
		Name:        "Excessive Crawl Depth",
		Category:    "discovery",
		Severity:    "low",
		Description: "The page is located more than 5 click hops away from the root entry point.",
		Why:         "Deep pages receive lower crawl frequency and lower internal PageRank weighting.",
		Remediation: "Reduce click depth by linking to key pages from category pages or navigation menus.",
	},
	"RC-RENDER-001": {
		Name:        "Render Discrepancy (Client-Side Rendering)",
		Category:    "renderability",
		Severity:    "high",
		Description: "A material difference was detected between raw server HTML and the client-rendered DOM (such as JavaScript injecting noindex, changing canonical, or missing SSR).",
		Why:         "Search crawlers may index the raw server HTML before running client-side JavaScript, or fail to render content if hydration errors occur.",
		Remediation: "Ensure initial server HTML contains essential metadata, canonical tags, and core text, or implement SSR/SSG.",
	},
}

// Documentation formats a comprehensive explanation of a rule for CLI display.
func Documentation(r Rule) string {
	doc, ok := ruleDocs[r.ID()]
	if !ok {
		return fmt.Sprintf("Rule: %s (v%d)\nNo detailed documentation available.\n", r.ID(), r.Version())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s (%s)\n\n", doc.Name, r.ID()))
	sb.WriteString(fmt.Sprintf("**Category:**    %s\n", doc.Category))
	sb.WriteString(fmt.Sprintf("**Severity:**    %s\n", doc.Severity))
	sb.WriteString(fmt.Sprintf("**Version:**     v%d\n\n", r.Version()))
	sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", doc.Description))
	sb.WriteString(fmt.Sprintf("## Why It Matters\n%s\n\n", doc.Why))
	sb.WriteString(fmt.Sprintf("## How to Fix\n%s\n", doc.Remediation))

	return sb.String()
}
