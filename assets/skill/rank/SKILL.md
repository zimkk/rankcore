---
name: rank
description: Understand, research, audit, improve, and verify a website or web application for search visibility, generative search visibility, crawlability, indexability, structured data, authority, and AI crawler access. Use when the user asks to rank, optimize SEO/AEO/GEO, improve AI search visibility, audit a site, research search demand/competitors, implement search improvements, or verify a previous optimization.
---

# RankCore `/rank`

RankCore is an agent-native search engineering workflow backed by a deterministic local audit engine.

The host coding agent supplies repository understanding, web research when available, reasoning, and code editing. The `rankcore` binary supplies deterministic evidence for crawl, indexability, canonicalization, links, structured data, crawler access, and verification.

## Core contract

1. Understand the product before optimizing it.
2. Prefer repository evidence over assumptions.
3. Use `rankcore` for deterministic claims whenever the binary can measure them.
4. Use current web research for external, market, platform, crawler, and search-engine claims when the host supports browsing.
5. Never invent keyword volume, traffic, backlink counts, rankings, citations, or provider metrics.
6. Load only the references relevant to the current phase.
7. Implement only changes justified by evidence and product intent.
8. Re-run verification before claiming a fix is complete.
9. Never guarantee rankings, indexing, AI citations, traffic, or conversions.
10. Do not require paid SEO platforms, a RankCore account, a RankCore-hosted LLM, MCP, Node, Python, or a cloud backend for the default workflow.

## Modes

Interpret `/rank` with no mode as the full workflow.

Supported modes:

- `/rank` — full discovery → research → audit → plan → implement → verify.
- `/rank audit` — deterministic and interpretive audit only.
- `/rank research` — product/market/search research only.
- `/rank fix` — implement prioritized fixes from current evidence.
- `/rank verify` — verify prior RankCore changes against a baseline.

Read the matching file in `workflows/`.

## Progressive references

Load only what is needed:

- fundamentals, crawl/index/rank model, search quality → `references/search-foundations.md`
- status/robots/sitemaps/canonicals/rendering/internal links/international → `references/technical-search.md`
- product understanding, intent, content architecture, comparison/use-case/programmatic content → `references/intent-content.md`
- backlinks, mentions, digital PR, citations, community visibility → `references/authority-citations.md`
- AI search, answer engines, AI crawlers, robots purpose distinctions, llms.txt → `references/ai-search-crawlers.md`
- schema, JSON-LD, entities, Organization/Product/SoftwareApplication → `references/structured-data-entities.md`
- Search Console, Bing, analytics, experiments, provider metrics → `references/measurement.md`
- spam, manipulation, privacy, safety, false claims → `references/guardrails.md`
- authoritative URLs and freshness tracking → `references/sources.md`

Always consult `guardrails.md` before scaled page generation, link acquisition, crawler-policy changes, reputation tactics, or large content automation.

## Environment check

Before relying on deterministic audits:

```bash
rankcore doctor --json
```

If the binary is unavailable, continue with repository-only analysis and explicitly mark deterministic coverage as unavailable. Do not fabricate a RankCore result.

## Product understanding

Before research or code changes, infer a compact product profile from the repository:

- product name and category;
- core user problem;
- primary capabilities;
- target users/buyers;
- business model;
- geographic/language scope;
- public versus private surfaces;
- important entities;
- framework and deployment shape;
- likely conversion events.

Persist or update `.rankcore/profile.json` when the workflow calls for it.

Ask the user only when an unresolved ambiguity materially changes:
- public/private indexing intent;
- regulated/high-risk claims;
- geographic targeting;
- destructive sitewide changes;
- credentials or external account access.

Do not turn ordinary product discovery into a questionnaire.

## Research rules

When web access exists:

- research the market and terminology;
- inspect representative competitors and search-result patterns;
- identify user jobs and intent clusters;
- identify comparison, alternative, integration, use-case, support, and educational opportunities;
- prefer first-party/official sources for platform facts;
- treat community sources as qualitative evidence;
- distinguish observation from inference.

When web access does not exist:
- continue using repository/site evidence;
- label external-market conclusions as hypotheses;
- do not invent current SERPs or market metrics.

RankCore itself must not scrape Google result HTML as a hidden SERP API.

## Audit rules

Determine an appropriate target:
- existing local development URL;
- preview/staging URL intentionally supplied for testing;
- production URL when appropriate;
- static/build output where supported.

The host agent decides how to run the application using the repository's existing tooling. The RankCore binary must not autonomously execute arbitrary repository scripts.

Typical audit:

```bash
rankcore audit http://localhost:<port> --out .rankcore/runs/<run-id> --json
```

Use the audit result as evidence, not as an SEO score.

Interpret findings in product context. A `noindex`, missing sitemap entry, or crawler block may be intentional.

## Implementation rules

Prioritize in this order unless evidence justifies otherwise:

1. accidental ineligibility or destructive technical errors;
2. canonical/discovery integrity;
3. content availability and page purpose;
4. internal linking/navigation;
5. entity and structured-data consistency;
6. page-level metadata/semantics;
7. content-market gaps with clear user value;
8. performance/UX refinements;
9. optional interoperability enhancements.

Prefer small, reviewable changes.

Do not:
- rewrite the app architecture solely for SEO when simpler fixes exist;
- create pages only to target trivial keyword variants;
- add schema not supported by visible content;
- unblock training crawlers as a search requirement;
- add `llms.txt` as a ranking requirement;
- buy/manufacture backlinks;
- alter privacy or authenticated routes for visibility.

## Verification

After implementation:

```bash
rankcore verify <target> \
  --baseline .rankcore/runs/<baseline-id>/audit.json \
  --out .rankcore/runs/<verify-id> \
  --json
```

A finding is fixed only if its verification condition passes.

Also run the repository's appropriate tests/build checks through the host agent when safe and available.

## Output style

Normal completion should be concise and implementation-oriented:

- product/search understanding;
- highest-impact evidence;
- changes made;
- verification result;
- limitations or external follow-up items.

Do not bury the user in a 5,000-word audit when the task was to improve the site.

## Privacy and safety

- Never read or expose secret values merely for SEO.
- Do not send repository contents to third-party services without explicit user authorization.
- Do not make private/authenticated pages public for discoverability.
- Do not silently change robots policy for AI training crawlers.
- Treat user-provided production targets conservatively.
- Follow the host agent's permission model for edits, shell commands, and network access.
