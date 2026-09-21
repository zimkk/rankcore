# Technical Search

Verified: 2026-09-21  
Review-after: 2027-03-21  
Authority: official-docs + standards + internal synthesis

## Objective

Ensure intended public pages are discoverable, fetchable, processable, canonicalized, internally connected, and semantically understandable without exposing private or duplicate surfaces.

## Audit order

1. target/environment correctness;
2. HTTP status and availability;
3. robots/noindex;
4. canonicalization;
5. rendering/content visibility;
6. discovery/internal links/sitemaps;
7. duplicates/URL variants;
8. metadata/semantics;
9. structured data;
10. international annotations;
11. performance refinements.

## HTTP semantics

Prefer:
- `200` for real available pages;
- `301/308` for durable moves;
- `302/307` for temporary moves;
- `404/410` for absent/removed resources.

Detect:
- 5xx;
- redirect loops;
- unnecessary chains;
- soft-404 patterns;
- inconsistent protocol/host behavior.

## Robots and indexability

`robots.txt` primarily controls crawling. It is not a reliable de-indexing mechanism for an otherwise discoverable URL.

For accessible pages that should not appear in search, use an appropriate `noindex` control that crawlers can read. Truly private content should normally require authentication/authorization.

Inspect:
- robots meta;
- X-Robots-Tag;
- environment-specific staging rules;
- CDN/WAF blocks;
- bot-specific directives.

Never assume a crawler block is accidental without product intent.

## Canonicalization

Signals should align:
- redirects;
- canonical element;
- internal links;
- sitemap URLs;
- hreflang clusters;
- preferred protocol/host.

Flag contradictions, not harmless variation.

Do not treat “missing self-canonical” as universally critical; evaluate duplicate risk and framework behavior.

## Sitemaps

Sitemaps help discovery and canonical signaling; they are not ranking boosts.

Good sitemap sets generally contain:
- canonical desired URLs;
- absolute URLs;
- valid status targets;
- accurate `lastmod` only when truly maintained;
- scalable sitemap indexes where necessary.

Check for:
- redirects/404s/noindex URLs inside sitemap;
- wrong host;
- malformed XML;
- sitemap/canonical disagreement;
- missing important known indexable routes with contextual confidence.

## JavaScript and rendering

Important public content should be reachable at stable URLs and remain understandable after processing.

Prefer server/static rendering when it naturally fits, but do not rewrite an SPA solely because it uses JavaScript.

Audit both raw HTML and rendered DOM when available:
- essential text;
- titles/meta/canonical;
- links;
- structured data;
- error states.

Rendered mode should be optional and use an already-installed Chromium-family browser via CDP when possible, avoiding a mandatory Node/Playwright runtime.

## Internal linking

Audit:
- broken links;
- links to redirecting/noncanonical URLs;
- orphan candidates;
- excessive depth;
- JS navigation without crawlable anchors;
- low-information anchors when they impair navigation.

An orphan candidate is evidence for review, not proof that the page must be linked/indexed.

## URL design and duplicates

Prefer stable, human-readable URLs where practical.

Review:
- tracking/query parameters;
- filters/facets;
- case/trailing slash variants;
- print/AMP/archive variants where relevant;
- duplicate content paths;
- pagination behavior.

Use canonicalization and crawl controls intentionally; do not blanket-block useful user URLs without understanding implications.

## Internationalization

For genuine localized alternates:
- stable URL per locale;
- localized main content;
- valid `hreflang`;
- reciprocal annotations;
- self-reference;
- canonical/indexable targets;
- crawlable language switcher.

Avoid forced geo/IP redirects that prevent access to alternates.

Do not mechanically translate an English keyword map; re-research the market/language.

## Local businesses

Where relevant:
- accurate public NAP/business information;
- opening hours;
- service area/location proof;
- real location content;
- consistent public profiles;
- LocalBusiness markup only when accurate.

Never create fake offices or doorway city pages.

## Semantics and accessibility

Use semantic HTML and accessible interaction patterns because they improve usability and machine interpretation.

Do not present accessibility as merely an SEO trick.

## Performance

Performance matters to users and can relate to page experience. Treat lab data as diagnostic and field data as stronger evidence of real user experience.

Do not claim that shaving a few milliseconds guarantees ranking gains.
