# AI Search and Crawlers

Verified: 2026-09-21  
Review-after: 2026-10-21  
Authority: official vendor docs + official search docs + internal synthesis

## Core principle

Do not treat “AI SEO” as a separate loophole.

For Google Search generative experiences, official guidance says core Search eligibility, quality, and SEO practices remain the foundation. There are no special technical requirements for AI Overviews/AI Mode beyond normal Search eligibility, and no special AI schema is required.

## Generative-search readiness

Favor:
- indexable/crawlable public content;
- unique, non-commodity information;
- clear entities and claims;
- visible supporting evidence;
- strong internal navigation;
- accurate structured data;
- good user experience;
- original media where useful;
- concise answers embedded in genuinely useful pages.

Do not overproduce query-variation pages for “fan-out” terms.

## Crawler purpose must be separated

Different bots can represent different purposes:

1. **Search/discovery/retrieval** — may support showing links, snippets, summaries, or citations.
2. **User-triggered fetch/agent access** — retrieves content because a user asked an assistant/agent to visit.
3. **Training** — content collection for model improvement/training.

RankCore must report these categories separately and neutrally.

Blocking a training crawler must never be reported as an SEO error.

## OpenAI

Current official publisher guidance distinguishes `OAI-SearchBot` for ChatGPT search discovery from `GPTBot` for potential training controls.

RankCore must revalidate bot names and semantics against current OpenAI official docs before hard-coded presets become stale.

Do not imply that allowing GPTBot is required for ChatGPT Search visibility.

## Anthropic

Current Anthropic guidance distinguishes:
- `ClaudeBot` — training-oriented crawling;
- `Claude-SearchBot` — search indexing/discovery;
- `Claude-User` — user-initiated retrieval.

Revalidate the live Anthropic crawler documentation and bot IP feed before shipping presets. Never infer that training, search, and user-fetch crawlers have identical purposes.

## Google

Google Search generative features use normal Google Search crawling/indexing controls. Google states that sites do not need new AI text files or special markup to appear in these features.

Google-Extended and other controls may have different purposes from Googlebot; use current official documentation before advising policy changes.

## `llms.txt`

Treat `llms.txt` as an optional interoperability convention.

Do not:
- flag its absence as an SEO failure;
- claim Google ranking benefit;
- generate it by default solely for search visibility.

If a project chooses to publish it:
- keep links canonical/public;
- avoid private URLs;
- keep claims current;
- validate the chosen proposal/version.

Google's 2026 documentation clarification states `llms.txt` is not needed for Google Search and does not positively or negatively affect visibility/rankings there.

## Answer-engine citation behavior

RankCore may optimize for citation readiness, not guarantee citation.

Useful characteristics:
- stable URLs;
- precise factual statements;
- explicit source/author/company identity;
- original data;
- current dates where relevant;
- well-structured pages;
- clear relationships between claims and evidence;
- accessible content.

## Robots policy UX

When reporting AI access, present a matrix:

| Purpose | Bot/vendor | Current rule | Evidence | Recommendation |
|---|---|---|---|---|

Recommendations must respect the user's privacy/business preference, especially for training access.

Never silently change robots.txt merely because RankCore prefers broader AI access.

## Agent-friendly websites

Browser agents benefit from accessible, understandable interaction:
- semantic controls;
- labels;
- stable navigation;
- meaningful button/link names;
- predictable forms;
- accessible status/error feedback.

Treat this as user/agent interoperability, not a traditional ranking hack.
