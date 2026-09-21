# Structured Data and Entities

Verified: 2026-09-21  
Review-after: 2026-12-21  
Authority: Google Search docs + schema.org + internal synthesis

## Principle

Structured data describes reality already represented by the page/site. It must not invent entities, reviews, prices, ratings, authors, locations, availability, or claims.

Prefer JSON-LD where it fits the framework and platform guidance.

## Validation layers

1. JSON parses.
2. Schema types/properties are coherent.
3. Values match visible content/source-of-truth data.
4. Entity identities do not conflict.
5. Google-specific eligibility requirements are met when targeting a supported search feature.
6. Deployment output is verified.

Passing schema syntax does not guarantee a rich result.

## Entity model

Useful entities may include:
- Organization;
- Person;
- WebSite;
- WebPage;
- Product;
- SoftwareApplication;
- LocalBusiness;
- Article/BlogPosting;
- BreadcrumbList;
- VideoObject;
- Event;
- JobPosting;
- Offer.

Only use types justified by the real page/product.

## Identity

Where useful:
- use stable `@id` values;
- connect Organization ↔ WebSite ↔ WebPage ↔ Product/SoftwareApplication;
- keep names, URLs, logos, contact information, and social identities consistent;
- distinguish publisher/author/product/brand roles.

Avoid giant graphs that add complexity without clearer meaning.

## Product / software

For software/SaaS pages, evaluate whether Product and/or SoftwareApplication accurately represent the offering and whether current Google-supported experiences make the markup useful.

Do not add fake aggregate ratings/reviews.

Pricing/Offer data must reflect actual visible/current commercial facts.

## Organization

Organization markup can clarify identity, but it is not a substitute for a trustworthy public presence.

Use accurate:
- legal/brand name;
- URL;
- logo;
- sameAs links where genuinely official;
- contact details where appropriate.

## Local business

Use only for real public business locations/service entities. Do not invent addresses for local-search capture.

## Article/authorship

Authorship should reflect actual responsibility. Do not fabricate expert author profiles to simulate E-E-A-T.

## FAQ

Do not treat FAQ schema as a universal rich-result tactic. Google deprecated the FAQ rich result feature in 2026.

FAQ-style visible content can still be useful for users when the questions are real and the answers are helpful.

## Visible-content consistency

Flag contradictions such as:
- schema price differs from page price;
- unavailable product marked available;
- review count unsupported on page/source;
- Organization identity conflicts;
- canonical URL and entity URL mismatch.

## Rich result volatility

Supported search features change. Revalidate Google-specific requirements against current Search Central documentation rather than freezing assumptions indefinitely.
