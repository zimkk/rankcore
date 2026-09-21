# Authoritative Sources

Verified: 2026-09-21  
Review-after: 2026-10-21  
Authority: official sources first

Use these URLs as the primary verification set. For fast-moving platform behavior, re-check the live official page before changing hard-coded crawler presets or search-platform assumptions.

## Google Search

- Search Essentials — https://developers.google.com/search/docs/essentials
- SEO Starter Guide — https://developers.google.com/search/docs/fundamentals/seo-starter-guide
- How Search works — https://developers.google.com/search/docs/fundamentals/how-search-works
- Developer SEO guide — https://developers.google.com/search/docs/fundamentals/get-started-developers
- Helpful, reliable, people-first content — https://developers.google.com/search/docs/fundamentals/creating-helpful-content
- Spam policies — https://developers.google.com/search/docs/essentials/spam-policies
- AI features and your website — https://developers.google.com/search/docs/appearance/ai-features
- Optimizing for generative AI features — https://developers.google.com/search/docs/fundamentals/ai-optimization-guide
- Search documentation updates — https://developers.google.com/search/updates
- JavaScript SEO basics — https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics
- Sitemap guide — https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap
- Localized versions / hreflang — https://developers.google.com/search/docs/specialty/international/localized-versions
- Structured data introduction — https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data
- Structured data policies — https://developers.google.com/search/docs/appearance/structured-data/sd-policies
- Generative AI performance reports announcement — https://developers.google.com/search/blog/2026/06/gen-ai-performance-reports

## Google Search Console

- Search Console help — https://support.google.com/webmasters/
- URL Inspection — https://support.google.com/webmasters/answer/9012289
- Page indexing — https://support.google.com/webmasters/answer/7440203
- Performance reports — https://support.google.com/webmasters/answer/10268906

## Google crawling / robots standards

- robots.txt specification guidance — https://developers.google.com/search/docs/crawling-indexing/robots/robots_txt
- Robots Exclusion Protocol RFC 9309 — https://www.rfc-editor.org/rfc/rfc9309

## OpenAI discovery/crawlers

- Publishers and Developers FAQ — https://help.openai.com/en/articles/12627856-publishers-and-developers-faq

This official guidance currently distinguishes OAI-SearchBot for ChatGPT search discovery from GPTBot training controls. Revalidate before shipping crawler presets.

## Anthropic discovery/crawlers

- Anthropic crawler controls — https://support.claude.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler
- Anthropic crawler IP feed — https://claude.com/crawling/bots.json

Current official guidance distinguishes `ClaudeBot` (training), `Claude-SearchBot` (search), and `Claude-User` (user-initiated retrieval). Revalidate before shipping crawler presets because this area is fast-moving.

## Bing / IndexNow

- Bing Webmaster Tools — https://www.bing.com/webmasters/
- IndexNow documentation — https://www.indexnow.org/documentation
- IndexNow FAQ — https://www.indexnow.org/faq
- IndexNow protocol/search engines — https://www.indexnow.org/searchengines

## Schema and structured data

- Schema.org — https://schema.org/
- Google structured data gallery/docs — https://developers.google.com/search/docs/appearance/structured-data/search-gallery

## Web performance and accessibility

- web.dev Core Web Vitals — https://web.dev/articles/vitals
- WAI-ARIA overview — https://www.w3.org/WAI/standards-guidelines/aria/

## Agent-skill / coding-agent architecture

### Codex

- Customization overview — https://developers.openai.com/docs/customization/overview
- AGENTS.md guidance — https://developers.openai.com/docs/agent-configuration/agents-md

Codex currently supports reusable Skills with progressive loading of SKILL.md, references, and helper scripts.

### Claude Code

- Claude Code documentation — https://code.claude.com/docs
- Plugins — https://code.claude.com/docs/en/plugins

Use current Claude documentation for exact skill/plugin install paths and invocation behavior at implementation time.

## llms.txt

- Community proposal — https://llmstxt.org/

Treat as an optional convention, not an official universal search-ranking requirement. Google Search documentation explicitly states AI text files such as llms.txt are not required for AI features, and 2026 Search documentation updates clarified that llms.txt has no positive or negative Google Search ranking/visibility effect.

## Source policy for RankCore

1. Prefer official search-engine/vendor docs for platform facts.
2. Prefer standards bodies for protocols.
3. Use high-quality secondary research for market interpretation, never as a replacement for official policy.
4. Use community sources for qualitative user language/experience.
5. Date fast-moving claims.
6. Never hard-code proprietary third-party SEO metrics as ground truth.
