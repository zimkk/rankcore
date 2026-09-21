# Guardrails

Verified: 2026-09-21  
Review-after: 2026-10-21  
Authority: official search spam policies + security/privacy principles

## Purpose

RankCore optimizes legitimate discoverability. It must not become a manipulation engine, fabrication engine, privacy bypass, or mass-content spammer.

## Never recommend or implement as ranking tactics

- keyword stuffing;
- hidden text/links;
- cloaking;
- doorway pages;
- deceptive redirects;
- paid links that pass ranking credit;
- automated bulk link creation;
- excessive reciprocal linking schemes;
- low-quality directory/comment/forum spam;
- expired-domain abuse;
- site-reputation abuse;
- mass low-value AI pages;
- scraped/spun/stiched commodity content;
- fake functionality;
- misleading structured data;
- browser-navigation manipulation;
- fake locations/reviews/testimonials/authors/credentials;
- manufactured citations or media mentions.

## Scaled content

Automation is acceptable when it produces real differentiated value.

Require a value gate before large indexable page generation.

Escalate for human review when:
- dozens/hundreds of new template pages are proposed;
- pages mostly differ by keyword/location/name substitutions;
- content uses regulated or high-impact claims;
- evidence is unavailable;
- UGC will become indexable at scale.

## External factual claims

For claims about:
- competitors;
- prices;
- laws/regulations;
- market statistics;
- product capabilities;
- crawler behavior;
- search platform features;
- search-engine policy;

use current evidence when available and date it where relevant.

Do not invent a source because a claim would be useful.

## Privacy

Do not:
- read `.env` values unless explicitly necessary and authorized;
- expose private routes;
- publish internal docs;
- submit proprietary repository content to third-party SEO services without consent;
- change analytics/advertising tracking without permission;
- silently opt the site into AI training crawling.

## Repository execution safety

The RankCore binary should not execute arbitrary repository scripts based on parsed project metadata.

The host coding agent may run project commands under its normal permission model after inspecting them.

RankCore should confine itself to its documented CLI behavior and safe filesystem/network operations.

## Network safety

Protect against:
- SSRF to local metadata endpoints;
- unrestricted file URLs;
- unexpected private-network crawling;
- redirect escape to unsafe targets;
- credential-bearing URLs;
- unlimited crawl scope.

Use same-origin/default scope and explicit opt-in for broader crawling.

## Evidence levels

Use:
- `observed` — directly measured/read;
- `sourced` — supported by named external source;
- `inferred` — reasoned from evidence;
- `hypothesis` — plausible but unverified.

Do not present hypotheses as audit failures.

## No ranking guarantees

Never promise:
- position #1;
- indexation;
- rich result eligibility;
- AI citation;
- traffic growth;
- conversion growth.

RankCore improves readiness, evidence quality, and implementation discipline.

## AI crawler policy

Search/retrieval, user-triggered fetching, and training are different purposes.

A business may reasonably allow one and block another.

Do not frame privacy-preserving crawler choices as “bad SEO” unless there is direct, current evidence about the specific discovery surface being targeted.

## High-impact/YMYL content

Require stronger evidence and human review for medical, legal, financial, safety-critical, or similarly high-impact claims.

RankCore may improve technical discoverability, but it should not fabricate professional expertise or unsafe guidance.
