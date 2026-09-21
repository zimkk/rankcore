# Search Foundations

Verified: 2026-09-21  
Review-after: 2027-03-21  
Authority: official-docs + internal synthesis

## Purpose

Use this reference for the durable search mental model. It is not a ranking-factor checklist and does not promise outcomes.

## Crawl → index → retrieve → present

Diagnose the earliest broken stage first:

1. **Discovery** — a system learns a URL exists.
2. **Crawl access** — the URL and required resources can be fetched.
3. **Processing/rendering** — meaningful content can be interpreted.
4. **Indexing/canonicalization** — the system decides whether/how to store or consolidate the page.
5. **Retrieval/ranking** — the page is considered for a user need.
6. **Presentation** — link, title, snippet, rich treatment, or AI-supported citation is selected.
7. **User outcome** — the visit satisfies the task.

A title rewrite does not solve `noindex`. Backlinks do not repair a sitewide wrong canonical.

## Foundational doctrine

For Google Search and its generative search features, core SEO/search eligibility and quality practices remain foundational. There is no separate technical shortcut that guarantees AI Overview/AI Mode inclusion.

Prioritize:
- crawlable public content;
- intentional indexability;
- clear canonical URLs;
- useful internal links;
- important content in accessible textual form;
- helpful, reliable, people-first information;
- unique/non-commodity value;
- accurate structured data that matches visible content;
- good page experience.

## Content quality

Prefer content that:
- solves a real user job;
- contains original information, experience, analysis, examples, data, demonstrations, tools, or perspective where appropriate;
- identifies responsibility/authorship when trust depends on it;
- explains methodology when claims require it;
- is kept current when facts change;
- fits the site's actual purpose.

Do not reduce E-E-A-T to a numeric “factor.” Treat experience/expertise/authority/trust as quality and credibility concepts, especially for high-impact subjects.

## Search intent

Classify by user job, not just keyword shape:

- informational;
- navigational;
- commercial investigation;
- transactional;
- local;
- support/troubleshooting;
- comparison/alternative;
- integration/workflow;
- conversational/multi-part answer intent.

A page should have a coherent dominant job.

## Keywords

Use keywords to:
- learn user vocabulary;
- map intents to pages;
- improve titles/headings/anchors naturally;
- discover missing use-case/comparison/integration/support content;
- identify entities and subtopics.

Do not use:
- keyword density targets;
- repetitive exact-match stuffing;
- hundreds of trivial query variants;
- invented search volumes.

## Site architecture

Healthy public sites generally provide:
- stable meaningful URLs;
- crawlable HTML links to important pages;
- descriptive anchors;
- coherent navigation;
- deliberate handling of filters/facets/parameters;
- limited orphaned content;
- consistent canonicalization;
- sitemaps focused on desired canonical URLs.

## Titles and snippets

Important indexable pages should have descriptive, useful, substantially distinct titles. Meta descriptions are candidate summaries, not guaranteed snippets and not direct ranking promises.

Avoid arbitrary character-count “pass/fail” rules.

## Links and authority

Editorial links and independent references may contribute to discovery/reputation. The sustainable approach is to create things worth citing and promote them to relevant audiences.

Do not treat third-party “Domain Authority/Domain Rating” metrics as search-engine-owned scores.

## AI-generated content

The mere use of AI is not the central policy issue. Low-value scaled content created primarily to manipulate search systems is the risk.

Use automation to increase useful production capacity, not to multiply commodity pages.

## Context matters

Do not automatically flag:
- `noindex`;
- robots blocks;
- pages absent from sitemap;
- authenticated routes;
- utility routes;
- duplicate/parameter pages;
- checkout/account pages.

Intent determines correctness.

## Reject these myths

- “Meta keywords matter for Google ranking.”
- “Every page must be indexed.”
- “Every page needs exactly one H1 to rank.”
- “Every title must be exactly 60 characters.”
- “Longer content always ranks better.”
- “More backlinks is always better.”
- “Structured data automatically increases rankings.”
- “AEO/GEO requires a special Google-only schema or AI text file.”
- “llms.txt is a Google ranking factor.”
