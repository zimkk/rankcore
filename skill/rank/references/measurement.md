# Measurement and Monitoring

Verified: 2026-09-21  
Review-after: 2026-12-21  
Authority: official first-party docs + internal synthesis

## Measurement hierarchy

### 1. Deterministic technical state

RankCore can directly verify:
- HTTP behavior;
- crawl controls;
- indexability directives;
- canonical consistency;
- sitemap consistency;
- internal links;
- markup syntax/selected semantics;
- selected rendered-DOM conditions;
- change verification.

This is not evidence of actual ranking.

### 2. Search-engine state

Owned-site first-party tools can provide:
- indexing state;
- query/page impressions/clicks;
- sitemap processing;
- URL inspection;
- enhancement issues;
- country/device trends;
- generative-AI search visibility where officially reported.

Google Search Console is primary for Google-owned search evidence. Bing Webmaster Tools provides analogous Bing-side evidence.

### 3. Business outcomes

Track:
- qualified organic sessions;
- signups/leads;
- revenue;
- conversion rate;
- assisted conversions;
- activation/retention by landing page;
- branded/non-branded demand where measurable.

Traffic without business/user value is not success.

## Google generative AI measurement

Google introduced dedicated generative AI performance reporting in Search Console in 2026, with rollout later expanded broadly. Use current Search Console documentation for exact availability and semantics.

Do not backfill historical AI-feature metrics by guessing.

## Provider metrics

Commercial tools may provide:
- keyword volumes;
- difficulty estimates;
- backlink indexes;
- traffic estimates;
- SERP history;
- competitor visibility.

Rules:
- never require them for core RankCore;
- label provider/date/geo;
- call estimates estimates;
- do not merge proprietary scores as interchangeable;
- never claim provider metrics are search-engine internals.

## Experiments and change logging

Record:
- change date;
- affected URLs/templates;
- hypothesis;
- technical verification;
- important external events/releases;
- observation window.

Avoid attributing every traffic movement to the latest edit.

## Rank tracking

Positions vary by:
- query;
- location;
- device;
- personalization/context;
- time;
- search feature composition.

Use rank tracking as directional evidence, not a stable truth.

## Before/after analysis

Use comparable periods and consider:
- seasonality;
- product launches;
- migrations;
- algorithm/system changes;
- paid campaigns;
- outages;
- measurement changes.

## Indexing misconceptions

- “Live URL passes inspection” ≠ guaranteed indexing.
- “Indexed” ≠ ranks for target queries.
- 100% indexing is not inherently a goal.
- duplicate/utility/private URLs can correctly remain unindexed.

## Search Console integration

Keep optional in later RankCore versions.

Credential rules:
- user-controlled authorization;
- least privilege;
- no secrets in RankCore project state;
- no forced cloud account.

## IndexNow

Can notify participating engines about URL changes. Submission/receipt is not proof of indexing or ranking.
