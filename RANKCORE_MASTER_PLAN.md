# RankCore — Master Plan

**Research baseline:** 2026-09-21  
**Status:** Build-authoritative  
**Audience:** AI coding agents, engineers, maintainers, product owners  
**Primary product promise:** Install once, invoke `/rank`, let the user's existing coding agent understand the product, research the market when its host provides web access, audit the site deterministically, implement appropriate changes, and verify them — without requiring a RankCore account, RankCore-hosted LLM, SEO API key, MCP server, Node runtime, Python runtime, vector database, or SaaS backend.

---

# 1. Executive decision

RankCore should be built as **two cooperating layers**:

1. **One universal Agent Skill named `rank`**
   - This is the intelligence/orchestration layer.
   - It runs inside the user's existing agent: Claude Code, Codex, Cursor, Gemini CLI, OpenCode, Cline, Copilot, or any host that supports Agent Skills.
   - It understands the repository, decides what research is needed, uses the host agent's native web/search capabilities when available, interprets audit findings, edits the codebase, and verifies the result.
   - It progressively loads bundled reference knowledge only when relevant.

2. **One native local executable named `rankcore`**
   - This is the deterministic measurement layer.
   - It is written in Go and shipped as a precompiled binary.
   - It crawls, parses, measures, validates, compares, and returns structured facts.
   - It never calls an LLM.
   - It never needs a RankCore cloud service.
   - It never owns market reasoning or content strategy.

The relationship is:

```text
User
  |
  |  /rank
  v
Existing coding agent
  |
  +--> Rank Skill
  |      |
  |      +--> inspect repository
  |      +--> load relevant RankCore knowledge
  |      +--> use host web search when available
  |      +--> reason about market / intent / product
  |      +--> edit code
  |
  +--> rankcore native CLI
         |
         +--> deterministic crawl
         +--> robots / sitemap / canonical checks
         +--> HTML + rendered-DOM extraction when available
         +--> internal-link graph
         +--> structured-data checks
         +--> AI crawler access checks
         +--> before/after verification
```

The default end-user experience must be approximately:

```text
1. Install RankCore once.
2. Open any web project in your existing coding agent.
3. Type: /rank
4. RankCore does the rest.
```

That is the product.

---

# 2. Architectural principles and final decisions

RankCore is designed around a small set of constraints that should guide every implementation decision:

| Area | Final decision | Rationale |
|---|---|---|
| Primary user interface | **One universal `/rank` Agent Skill** with modes such as `/rank audit`, `/rank research`, `/rank fix`, and `/rank verify` | Keeps the command surface simple and portable across coding agents. |
| Deterministic engine | **Go** | Produces a fast, self-contained cross-platform binary with straightforward concurrency and distribution. |
| Coding-framework understanding | **Handled by the host coding agent** | RankCore should not duplicate the repository/framework intelligence that Claude Code, Codex, Cursor, and similar agents already provide. |
| Browser rendering | **Optional, using an already-installed Chromium-family browser through CDP when available** | Avoids forcing Node, Playwright, or browser downloads into the default installation. |
| MCP | **Optional integration, not a foundation** | Local CLI execution is sufficient for the primary workflow. |
| Hooks | **Not mandatory** | `/rank` should explicitly own its lifecycle and must not slow unrelated development sessions. |
| Subagents | **Optional enhancement** | The core workflow must work in a single agent session across hosts. |
| Knowledge system | **Compact progressive Markdown references** | Keeps context usage low and makes knowledge maintainable. |
| Market and competitor research | **Performed by the host agent using its available web/search capabilities** | A zero-key local binary cannot provide trustworthy live SERP, keyword-volume, or backlink-index data by itself. |
| Keyword volume and difficulty | **Never fabricated** | RankCore prioritizes using business fit, search intent, evidence strength, competitive gap, page opportunity, and implementation cost unless real provider data is available. |
| Backlink intelligence | **No global backlink claims without an actual provider** | Local crawling cannot know the web-wide backlink graph. |
| Scoring | **No synthetic 0–100 SEO score** | Findings, coverage, evidence, and verified status are more useful than false precision. |
| Core Web Vitals | **Do not claim field CWV without a valid source** | Local observations and optional browser measurements are distinct from real-user field data. |
| `llms.txt` | **Optional interoperability artifact** | It must never be presented as a ranking requirement. |
| Search Console / Bing / IndexNow | **Optional integrations** | They require accounts, OAuth, or external services and are not necessary for the zero-friction core product. |
| Cloud backend | **Not required** | RankCore's core value must work locally without accounts, hosting, or synchronization. |
| RankCore-owned LLM | **Not required** | RankCore uses the user's existing coding agent for reasoning and code changes. |

The governing principle is:

> **Do not rebuild capabilities that the host coding agent already provides well. Build only what the agent cannot reliably do by reasoning alone.**

---

# 3. What RankCore is — and is not

## 3.1 Product definition

RankCore is an **agent-native search visibility engineering system**.

It turns an existing coding agent into a disciplined search/discoverability engineer by providing:

- a repeatable workflow;
- current search/AI-search doctrine;
- deterministic local auditing;
- structured project state;
- implementation guardrails;
- verification.

It should help with:

- classic SEO foundations;
- technical SEO;
- crawlability/indexability;
- information architecture;
- search intent and content opportunities;
- structured data;
- internal linking;
- JavaScript-rendering risks;
- AI search visibility/citation readiness;
- AI crawler access;
- honest AEO/GEO practices;
- authority/citation strategy;
- measurement readiness.

## 3.2 Explicit non-goals

RankCore must **not** claim to:

- guarantee a ranking position;
- guarantee inclusion or citation in Google AI Overviews, AI Mode, ChatGPT, Claude, Copilot or any other answer engine;
- know exact search volume without a real data provider;
- know global backlinks without a real backlink index;
- know true field Core Web Vitals without an appropriate external data source;
- automatically create fake testimonials, statistics, authors, citations or reviews;
- mass-produce thin pages for rankings;
- scrape Google/Bing result pages as an undocumented core dependency;
- replace Search Console, Bing Webmaster Tools, analytics or commercial SEO datasets;
- become another general-purpose coding agent;
- become another SaaS dashboard before local value is proven.

---

# 4. Product doctrine based on current search guidance

RankCore's knowledge system must encode the following high-level rules.

## 4.1 AI search does not replace search fundamentals

Google's current guidance for generative AI features says that core Search eligibility and quality practices remain foundational and that there are no special technical requirements for AI Overviews or AI Mode. RankCore therefore must not invent a fake independent “GEO ranking stack.”

Reference: https://developers.google.com/search/docs/appearance/ai-features  
Reference: https://developers.google.com/search/docs/fundamentals/ai-optimization-guide

RankCore should model AI discoverability as:

```text
crawlability + indexability + useful/original content + entity clarity
+ machine-readable structure when appropriate + authority/evidence
+ freshness + accessible public pages
```

—not as a set of magic tags.

## 4.2 Scaled low-value generation is a risk

Google's spam policies explicitly target scaled content abuse regardless of whether content is human- or AI-generated. RankCore must require a user-value test before proposing programmatic content expansion.

Reference: https://developers.google.com/search/docs/essentials/spam-policies

## 4.3 JavaScript rendering matters, but server-visible content is safer

Google can render JavaScript, but rendering is a separate stage and not all bots render JS. Important public content should be available in robust HTML whenever practical.

Reference: https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics

## 4.4 Search crawler and training crawler controls are not the same thing

RankCore must distinguish crawler purposes.

Examples as of this research baseline:

- OpenAI `OAI-SearchBot`: search discovery / ChatGPT Search eligibility.
- OpenAI training-related crawling is a different policy surface; RankCore must not tell users that enabling training crawling is required for search visibility.
- Anthropic `Claude-SearchBot`: search indexing/quality.
- Anthropic `Claude-User`: user-initiated page retrieval.
- Anthropic `ClaudeBot`: separate crawling purpose associated with model-development data controls.

References:  
https://help.openai.com/en/articles/12627856-publishers-and-developers-faq  
https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler

**RankCore must never automatically change a site's training-crawler policy.** That is a publisher/privacy/business decision, not an SEO fix.

## 4.5 AI visibility is measurable only where platforms expose data

Google introduced dedicated generative-AI performance reporting in Search Console in 2026, and Bing Webmaster Tools introduced AI citation performance views. Those are legitimate measurement integrations for later versions.

References:  
https://developers.google.com/search/blog/2026/06/gen-ai-performance-reports  
https://blogs.bing.com/webmaster/February-2026/Introducing-AI-Performance-in-Bing-Webmaster-Tools-Public-Preview

They are **optional external integrations**, not core dependencies.

---

# 5. The zero-key core guarantee

RankCore v1 should publicly guarantee the following:

### Required

- user's existing coding agent;
- local filesystem access;
- ability for the agent to execute the `rankcore` binary;
- internet only when the user asks for remote crawling, market research, or an update.

### Not required

- RankCore account;
- RankCore API key;
- OpenAI API key;
- Anthropic API key;
- Semrush/Ahrefs API;
- Google Search Console account;
- Bing Webmaster account;
- MCP configuration;
- Node.js;
- Python;
- Docker;
- database;
- vector store;
- browser download;
- cloud backend.

Compile-time open-source dependencies are acceptable because they ship inside the binary; **runtime dependencies are the thing to minimize**.

RankCore should perform no telemetry by default.

---

# 6. User experience: the golden path

## 6.1 Installation

Recommended Unix/macOS/WSL path:

```bash
curl -fsSL https://rankcore.dev/install.sh | sh
```

Recommended Windows PowerShell path:

```powershell
irm https://rankcore.dev/install.ps1 | iex
```

The bootstrap script should do only this:

1. detect OS + architecture;
2. download the correct signed/checksummed `rankcore` release;
3. verify integrity;
4. install it to a user-writable location without `sudo`;
5. run `rankcore setup`.

All real setup logic lives in the Go binary, not in giant shell scripts.

Also provide:

- manual GitHub Release download;
- Homebrew package later;
- WinGet package later;
- `npx skills add ...` as an **optional skill-only distribution path**, not as a core runtime dependency.

## 6.2 Setup

`rankcore setup` should:

1. detect known installed coding agents;
2. print what it found;
3. install the canonical `rank` skill into each selected agent's global skill directory;
4. use copy or symlink depending on platform/support;
5. never overwrite a conflicting user skill silently;
6. support `--dry-run`;
7. record exactly what it changed for clean uninstall.

Example:

```text
RankCore setup

Detected:
  ✓ Claude Code
  ✓ Codex
  ✓ Cursor
  ○ Gemini CLI

Installing /rank:
  ✓ ~/.claude/skills/rank
  ✓ ~/.codex/skills/rank
  ✓ ~/.cursor/skills/rank

Engine:
  ✓ rankcore 1.0.0

Ready. Open a web project and type /rank.
```

## 6.3 Daily use

One public command:

```text
/rank
```

Optional modes through the same skill:

```text
/rank audit
/rank research
/rank plan
/rank fix
/rank verify
/rank status
```

No user should need to understand the underlying knowledge files, scripts or agent paths.

## 6.4 Default `/rank` behavior

When a user invokes `/rank`, the skill should:

1. verify RankCore availability (`rankcore doctor --json`);
2. understand the repository;
3. infer product, target users, product category, core jobs, markets and public surfaces with evidence/confidence;
4. research current market/search landscape if the host has web capability;
5. determine how to run the existing app using the repository's existing scripts;
6. start the app using the coding agent's normal command tools — **the RankCore binary itself must never execute arbitrary package-manager/build scripts automatically**;
7. run deterministic audit against the local URL;
8. synthesize a prioritized implementation plan;
9. apply safe code changes through the host agent;
10. rerun build/tests as appropriate;
11. run `rankcore verify`;
12. produce a concise before/after result and clearly list unresolved external work.

The skill should ask the user only when a real product decision or potentially destructive/risky action cannot be inferred safely.

---

# 7. Capability-aware orchestration

Different agent hosts have different capabilities. RankCore must not assume all hosts are Claude Code.

The Skill should reason in terms of capabilities rather than brands.

## 7.1 Capability matrix

Potential host capabilities:

- `repo_read`
- `repo_write`
- `shell`
- `web_search`
- `web_fetch`
- `browser`
- `subagents`
- `git`

Required for full auto-fix:

- `repo_read`
- `repo_write`
- `shell`

Required for market research:

- `web_search` or equivalent.

If web search is unavailable:

- continue technical audit;
- infer only from repository/site content;
- mark market/competitor research as incomplete;
- never fabricate external evidence.

If shell is unavailable:

- skill can still provide advisory analysis from repository context;
- deterministic RankCore audit cannot run;
- report that coverage limitation.

If browser rendering is unavailable:

- HTTP/raw HTML audit still runs;
- RankCore reports `rendered_dom_verified: false`;
- the agent must not claim rendered visibility was tested.

---

# 8. Technology stack decision

## 8.1 Core engine: Go

Use the current stable Go toolchain during development and pin a tested minimum version in CI.

### Why Go is the best fit

RankCore's native engine is primarily:

- HTTP networking;
- concurrent crawling;
- URL normalization;
- HTML parsing;
- XML parsing;
- graph construction;
- JSON processing;
- local filesystem operations;
- deterministic rule evaluation;
- one-file cross-platform distribution.

Go is a better fit than TypeScript/Python because it can ship as one native binary with no Node/Python runtime requirement.

Go is a better fit than Rust here because RankCore does not need Rust's additional ownership/unsafe-complexity tradeoff to achieve its goals. Engineering speed, maintainability and simple concurrency are more valuable than shaving the last few MB or microseconds.

### Stack comparison

| Stack | Advantages | Problems for RankCore | Verdict |
|---|---|---|---|
| Go | native binary, strong HTTP/concurrency, easy cross-compilation, simple codebase | GC and binary size slightly larger than Rust | **Use** |
| Rust | excellent performance, safety, native binary | slower development, more complexity for a crawler that does not need it | Do not use for v1 |
| TypeScript/Node | excellent browser tooling and agent ecosystem | runtime dependency, npm tree, Windows/version friction, slower cold start | Do not use for core |
| Python | fast scripting, excellent libraries | packaging/runtime/version friction; poor “one binary” story | Do not use for core |
| Shell | easy installer | unsuitable for parsing/crawling/state/rules | Bootstrap only |

## 8.2 Runtime dependencies

### Hard runtime dependencies

None beyond the operating system/network stack.

### Optional runtime dependency

An already-installed Chromium-family browser for rendered audits.

RankCore must **not** automatically download a browser in v1.

## 8.3 Go dependency policy

Prefer Go standard library aggressively:

- `net/http`
- `net/url`
- `encoding/json`
- `encoding/xml`
- `crypto/*`
- `os`
- `path/filepath`
- `regexp`
- `time`
- `context`
- `sync`

Use `golang.org/x/net/html` for HTML tokenization/tree parsing.

Use a proven RFC 9309-compatible robots implementation only after dependency review. Do **not** casually hand-roll robots matching without a conformance test corpus.

For optional rendered mode, use a small Chrome DevTools Protocol Go library such as `chromedp`, but keep it behind an interface so browser support can be omitted or replaced.

Do not add an ORM, web framework, dependency injection framework, SQLite, YAML stack, embedded JS runtime or plugin framework in v1.

## 8.4 CGO policy

Build core releases with `CGO_ENABLED=0` whenever possible.

This keeps cross-platform release behavior predictable and avoids native library installation requirements.

---

# 9. Repository architecture

Use a **single Go module**, not a multi-package JavaScript monorepo.

```text
rankcore/
├── README.md
├── LICENSE
├── SECURITY.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── cmd/
│   └── rankcore/
│       └── main.go
│
├── internal/
│   ├── app/                 # command orchestration
│   ├── setup/               # agent detection + skill install
│   ├── audit/               # audit coordinator
│   ├── crawl/               # URL frontier, concurrency, fetch policy
│   ├── httpx/               # safe HTTP client + redirects
│   ├── extract/             # HTML/page extraction
│   ├── robots/              # RFC 9309 evaluation
│   ├── sitemap/             # sitemap + sitemap index parser
│   ├── canonical/           # canonical normalization/checks
│   ├── graph/               # internal link graph
│   ├── schema/              # JSON-LD extraction + supported checks
│   ├── hreflang/            # language/region annotations
│   ├── aiaccess/            # AI crawler policy classification
│   ├── render/              # optional CDP rendered snapshot
│   ├── rules/               # deterministic rule registry
│   ├── report/              # JSON/terminal/markdown output
│   ├── verify/              # baseline comparison
│   ├── state/               # .rankcore state schemas
│   ├── update/              # manual binary/asset updater
│   └── safety/              # URL/network/filesystem safety helpers
│
├── assets/
│   ├── skill/
│   │   └── rank/
│   │       ├── SKILL.md
│   │       ├── workflows/
│   │       │   ├── full.md
│   │       │   ├── audit.md
│   │       │   ├── research.md
│   │       │   ├── fix.md
│   │       │   └── verify.md
│   │       └── references/
│   │           ├── search-foundations.md
│   │           ├── technical-search.md
│   │           ├── intent-content.md
│   │           ├── authority-citations.md
│   │           ├── ai-search-crawlers.md
│   │           ├── structured-data-entities.md
│   │           ├── measurement.md
│   │           ├── guardrails.md
│   │           └── sources.md
│   └── agent-registry.json
│
├── testdata/
│   ├── sites/
│   ├── html/
│   ├── robots/
│   ├── sitemaps/
│   └── reports/
│
├── tests/
│   ├── integration/
│   ├── security/
│   └── skill-evals/
│
├── install.sh
├── install.ps1
└── .github/
    └── workflows/
```

Use Go `embed` to compile the canonical `rank` skill and knowledge assets into the binary. `rankcore setup` exports those assets into the appropriate agent directories.

This creates one source of truth and prevents separate Claude/Codex copies from drifting.

---

# 10. Why there should be only one public Skill

The user wants to remember one thing:

```text
/rank
```

The internal workflow may be complex. The public interface should not be.

The skill directory should be:

```text
rank/
├── SKILL.md
├── workflows/
└── references/
```

`SKILL.md` should be short. It should describe:

- when RankCore is appropriate;
- the supported modes;
- the workflow router;
- the rule that deterministic claims come from `rankcore` output;
- the rule that external claims require evidence;
- the rule to progressively load only relevant references.

It should **not** contain the entire SEO encyclopedia.

The skill's full content should be small enough that invoking `/rank` does not unnecessarily consume a massive context budget.

---

# 11. Rank Skill workflow contract

## 11.1 `/rank` — full mode

### Step 1 — Environment check

Run:

```bash
rankcore doctor --json
```

If the binary is missing, tell the user the one-line install command. Do not substitute invented deterministic auditing.

### Step 2 — Understand the product

Use repository evidence to build a concise product profile:

- product name;
- product category;
- one-sentence value proposition;
- features;
- target users / ICP;
- primary user jobs;
- business model where evident;
- target geography/language where evident;
- primary conversion actions;
- public pages/content surfaces;
- current framework/rendering model;
- assumptions and confidence.

Do not silently invent geography, customers, pricing, certifications or market category.

Persist this as `.rankcore/profile.json`.

### Step 3 — External research

If host web search exists, research:

- category language;
- problem-aware searches;
- solution-aware searches;
- product/commercial intent;
- comparison and alternative intent;
- user questions;
- direct/indirect competitors;
- content formats competitors rank/cite with;
- communities/publications relevant to authority;
- current platform guidance when a recommendation depends on a changing rule.

Every external factual conclusion should record source URLs and date.

Do not invent search volume, keyword difficulty or backlink counts.

If no web capability exists, skip this stage and label it incomplete.

### Step 4 — Create an audit target

Prefer a locally served production build.

The host agent should inspect existing repository scripts and decide how to run the app.

Important:

- RankCore CLI itself does not execute `npm install`, `pnpm install`, builds, migrations or arbitrary project scripts.
- The host agent executes project commands under its existing user permission model.
- If an app cannot be started, audit an existing preview/production URL when available or use filesystem-limited mode.

### Step 5 — Deterministic audit

Run:

```bash
rankcore audit http://localhost:<port> \
  --render auto \
  --out .rankcore/runs/<run-id> \
  --json
```

### Step 6 — Synthesis

Combine:

- project profile;
- research evidence;
- deterministic findings;
- business importance;
- implementation cost;
- risk.

Do not convert this into a fake SEO score.

Create priorities:

- `P0 — blocked / broken`
- `P1 — high impact`
- `P2 — meaningful improvement`
- `P3 — optional / experimental`

### Step 7 — Implement

Apply safe repository changes.

Changes with factual/product implications require stronger evidence or user approval.

### Step 8 — Verify

Run the app/tests again and then:

```bash
rankcore verify http://localhost:<port> \
  --baseline .rankcore/runs/<run-id>/audit.json \
  --out .rankcore/runs/<verify-id> \
  --json
```

A finding is only reported as fixed when its verification condition passes.

### Step 9 — Final response

The final response should be compact:

```text
RankCore complete

Fixed:
- ...
- ...

Remaining:
- ...

Research limitations:
- ...

Artifacts:
- .rankcore/...
```

The point is implementation, not a 5,000-word chat report.

---

# 12. CLI surface

Keep the native CLI intentionally small.

## 12.1 `rankcore setup`

Installs/updates the bundled `rank` skill into detected agents.

```text
rankcore setup
rankcore setup --all
rankcore setup --agent claude-code
rankcore setup --agent codex
rankcore setup --dry-run
```

## 12.2 `rankcore doctor`

Checks only RankCore environment health:

- binary version;
- skill asset version;
- supported host detection;
- write access to skill directories;
- optional Chrome/Chromium availability;
- update availability only when explicitly requested with a network flag.

It must not inspect secrets.

```text
rankcore doctor
rankcore doctor --json
```

## 12.3 `rankcore audit <target>`

Target can be:

- `http://localhost:3000`
- `https://example.com`
- a filesystem directory containing static/build HTML.

Core flags:

```text
--render auto|off|required
--max-pages <n>
--concurrency <n>
--include <pattern>
--exclude <pattern>
--out <dir>
--json
--format json|text|md
--user-agent <value>          # advanced
```

Default behavior:

- local targets: higher safe concurrency;
- remote targets: conservative concurrency and same-origin crawl;
- no forms;
- no login attempts;
- no CAPTCHA bypass;
- no cross-origin crawling unless explicitly seeded;
- respect robots on remote crawls by default;
- identify itself with a transparent RankCore user-agent.

## 12.4 `rankcore verify <target>`

Re-runs relevant checks and compares against a baseline report.

```text
rankcore verify <target> --baseline <audit.json> --out <dir> --json
```

Output states:

- `fixed`
- `still_present`
- `regressed`
- `not_retestable`
- `new`

## 12.5 `rankcore explain <rule-id>`

Prints deterministic rule documentation.

Useful for humans and agents without bloating every report.

## 12.6 `rankcore update`

Manual update only.

No hidden background updater is required in v1.

## 12.7 `rankcore uninstall`

Removes:

- installed RankCore skill copies/symlinks created by RankCore;
- RankCore installation metadata;
- optionally the binary when invoked from an installer-managed installation.

Never delete unrelated agent configuration.

## 12.8 `rankcore version`

```text
rankcore version
rankcore version --json
```

Report binary, asset, rule and schema versions.

---

# 13. Deterministic engine scope

The engine should measure what can be measured reliably.

## 13.1 Crawl / HTTP

- URL normalization;
- same-origin frontier;
- HTTP status;
- redirect chain;
- redirect loops;
- content type;
- final URL;
- response headers;
- basic timing;
- canonical discovery;
- robots directives;
- link extraction;
- sitemap discovery.

## 13.2 HTML extraction

Extract:

- title;
- meta description;
- robots meta;
- canonical;
- hreflang;
- headings;
- main textual content metrics;
- internal/external links;
- image src/alt basics;
- structured data blocks;
- Open Graph / social metadata as informational appearance checks;
- HTML language;
- anchors/buttons relevant to crawlable navigation.

Do not pretend every HTML heuristic is a ranking factor.

## 13.3 Sitemap

Support:

- sitemap XML;
- sitemap indexes;
- gzip where reasonable;
- URL validation;
- duplicate sitemap URLs;
- unreachable URLs;
- sitemap/robots mismatch;
- noncanonical sitemap entries;
- noindex sitemap entries.

## 13.4 Robots

Implement RFC 9309 behavior and test it thoroughly.

Audit policy should evaluate relevant agents separately:

- general crawler `*`;
- Googlebot;
- Bingbot when useful;
- OAI-SearchBot;
- Claude-SearchBot;
- Claude-User;
- optional other AI search agents only when their current official identities are present in the bundled crawler registry.

Training-only crawler policy should be reported separately and neutrally.

## 13.5 Canonicalization

Detect:

- missing canonical when context suggests it is useful;
- malformed canonical;
- multiple canonicals;
- canonical loops/chains;
- cross-domain canonicals;
- canonical target errors;
- noncanonical URLs in sitemap;
- duplicate URL variants.

Do not automatically assume every page needs a self-canonical; report with context.

## 13.6 Internal-link graph

Build adjacency structures for crawled public pages.

Report:

- pages with no internal inbound links among discovered pages;
- excessive depth;
- broken internal links;
- redirecting internal links;
- noncrawlable navigation patterns;
- clusters with weak connectivity.

Do not convert link graph metrics into fake “authority scores.”

## 13.7 Structured data

V1 should:

- parse JSON-LD;
- detect invalid JSON;
- identify schema types;
- detect obviously missing required fields for explicitly supported Google-rich-result types;
- compare critical structured values with visible page values where feasible;
- flag impossible/contradictory markup.

V1 should **not** try to become a complete schema.org reasoning engine.

Maintain a small supported-type registry and add types deliberately.

## 13.8 Hreflang / international

Check:

- language/region syntax;
- return links;
- self references;
- canonical conflicts;
- invalid/unreachable alternatives;
- `x-default` as informational, not mandatory.

## 13.9 Rendered mode

Rendered mode is optional but valuable for JS-heavy applications.

Behavior:

1. detect Chrome, Chromium, Edge or another supported Chromium binary;
2. launch a clean headless profile;
3. navigate without extensions/user profile;
4. wait for a bounded readiness policy;
5. capture rendered HTML + final URL + relevant network failures;
6. compare raw vs rendered SEO-critical fields.

Report differences such as:

- title only appears after JS;
- canonical changes after JS;
- primary text absent in raw HTML;
- links appear only after client rendering;
- robots meta changes dynamically.

Do not auto-download Chromium in v1.

## 13.10 Performance-lite

V1 may measure:

- response latency/TTFB approximation from its own request;
- HTML transfer size;
- redirect overhead;
- number of same-origin resources in rendered mode where available.

Do not label these “Core Web Vitals.”

Real CWV integrations can be added later from appropriate sources.

---

# 14. Findings model

Every deterministic finding should be machine-readable.

Example:

```json
{
  "id": "indexability.robots_disallow",
  "rule_version": 1,
  "category": "indexability",
  "severity": "high",
  "confidence": 1.0,
  "url": "https://example.com/pricing",
  "evidence": {
    "robots_url": "https://example.com/robots.txt",
    "user_agent": "Googlebot",
    "matched_rule": "Disallow: /pricing"
  },
  "summary": "Pricing page is blocked for Googlebot",
  "remediation": "Review whether this path is intentionally blocked.",
  "automation": "review_required",
  "verify": {
    "type": "robots_access",
    "expected": "allowed"
  }
}
```

Severity describes technical risk, not business priority.

The host agent determines final business priority using project context.

---

# 15. Audit report schema

Top-level report:

```json
{
  "schema_version": 1,
  "rankcore_version": "1.0.0",
  "ruleset_version": "2026.09",
  "generated_at": "...",
  "target": "http://localhost:3000",
  "mode": {
    "http": true,
    "rendered": true,
    "browser": "Chrome ..."
  },
  "coverage": {
    "pages_requested": 92,
    "pages_analyzed": 88,
    "pages_skipped": 4,
    "limit_reached": false
  },
  "findings": [],
  "crawler_access": {},
  "limitations": []
}
```

No global “SEO score.”

A short human summary can report counts by severity, but the source of truth is the findings array.

---

# 16. Project state

Create `.rankcore/` only when `/rank` is used in a project.

Recommended layout:

```text
.rankcore/
├── config.json
├── profile.json
├── research.md
├── plan.md
├── .gitignore
└── runs/
    ├── 2026-09-21T.../
    │   ├── audit.json
    │   └── audit.md
    └── 2026-09-21T...-verify/
        ├── verification.json
        └── verification.md
```

`.rankcore/.gitignore` should ignore run artifacts by default:

```text
runs/
cache/
*.tmp
```

`profile.json`, `research.md`, `plan.md` can remain visible and may be committed if the user wants team-shared RankCore context.

Do not use SQLite in v1.

If extremely large-site support later requires a database, introduce it only after real profiling proves JSON/in-memory structures insufficient.

---

# 17. Research methodology inside the Skill

Market research is an **agent task**, not a Go CLI task.

## 17.1 Evidence classes

All research conclusions should be tagged mentally or in artifacts as:

- `observed`: directly found in repository/site;
- `official`: current first-party platform documentation;
- `external`: competitor/public web evidence;
- `inferred`: reasoned conclusion based on evidence;
- `unknown`: cannot currently establish.

## 17.2 Opportunity prioritization

Do not rely on fake numbers.

Prioritize using:

- business relevance;
- commercial/user intent;
- evidence that users care about the topic;
- competitive gap;
- fit with existing product strengths;
- ability to provide original value;
- implementation effort;
- policy/spam risk;
- confidence.

## 17.3 Competitor research

Identify competitors based on actual product/category evidence rather than only the founder's named competitors.

Separate:

- direct product competitors;
- search-result competitors;
- editorial/content competitors;
- marketplaces/directories;
- community sources;
- alternative solution categories.

## 17.4 No SERP scraping dependency

RankCore itself must not parse Google result HTML as a hidden search API.

If the host agent supplies web search, use it.

If not, research degrades gracefully.

---

# 18. Knowledge architecture

Use a compact progressive knowledge set that can be loaded only when relevant.

## 18.1 `search-foundations.md`

Contains:

- crawl/index/rank mental model;
- people-first content principles;
- search intent;
- quality/authority fundamentals;
- Search Essentials principles;
- anti-spam basics.

## 18.2 `technical-search.md`

Contains:

- status codes;
- robots;
- sitemaps;
- canonicalization;
- JavaScript rendering;
- URLs;
- duplicate surfaces;
- internal links;
- international basics.

## 18.3 `intent-content.md`

Contains:

- product understanding;
- query/problem taxonomy;
- content architecture;
- comparison/alternative pages;
- programmatic content value gates;
- original information/evidence;
- content refresh.

## 18.4 `authority-citations.md`

Contains:

- useful off-site promotion;
- digital PR;
- citations/mentions;
- communities;
- backlinks as outcomes rather than manufactured artifacts;
- manipulative-link prohibitions.

## 18.5 `ai-search-crawlers.md`

Contains:

- Google AI feature doctrine;
- answer-engine citation readiness;
- crawler-purpose distinction;
- current official AI bot identities;
- robots guidance;
- `llms.txt` caveat;
- entity/claim clarity.

## 18.6 `structured-data-entities.md`

Contains:

- JSON-LD principles;
- visible-content consistency;
- supported Google types;
- Organization/Product/SoftwareApplication/etc. decision logic;
- entity consistency.

## 18.7 `measurement.md`

Contains:

- what RankCore can verify itself;
- Search Console;
- Bing Webmaster Tools;
- analytics;
- generative AI performance reports;
- experiment/change logging.

## 18.8 `guardrails.md`

Contains:

- no fabricated claims;
- no fake E-E-A-T signals;
- no mass thin pages;
- no hidden text;
- no doorway pages;
- no schema spam;
- no manipulative backlinks;
- no silent crawler privacy changes;
- no claiming verified metrics without a real source.

## 18.9 Freshness metadata

Each reference file starts with:

```text
Verified: YYYY-MM-DD
Review-after: YYYY-MM-DD
Authority: official-docs / standard / internal synthesis
```

For fast-moving facts such as crawler user agents, platform capabilities and Search reporting behavior, the Skill should use current web search to revalidate when available and when the bundled knowledge is stale.

---

# 19. Rule philosophy

RankCore must avoid becoming a generic “SEO checklist” that flags nonsense.

Each rule must meet at least one of these conditions:

1. directly affects crawlability/indexability;
2. detects an objective technical inconsistency;
3. detects a documented eligibility/markup problem;
4. identifies a clear discoverability/navigation problem;
5. provides useful evidence for the agent to interpret.

Avoid rules like:

- “title must be exactly 60 characters”;
- “meta description must be exactly 155 characters”;
- “every page needs exactly one H1” as a ranking mandate;
- “keyword density must be X%”;
- arbitrary word-count minimums;
- “SEO score decreases because Open Graph is missing”;
- “llms.txt missing = error.”

Those are the kinds of dry, outdated checks RankCore is specifically supposed to avoid.

---

# 20. Initial deterministic rule categories

V1 should prioritize approximately 40–60 high-confidence rules rather than hundreds of weak heuristics.

## 20.1 Availability/indexability

- server errors;
- unexpected client errors;
- redirect loops;
- excessive redirect chains;
- `noindex` on public pages;
- `X-Robots-Tag` conflicts;
- robots-disallowed important pages;
- canonical target not fetchable;
- canonical target noindex/disallowed.

## 20.2 Discovery

- unreachable sitemap;
- invalid sitemap entries;
- sitemap URLs returning errors;
- sitemap URLs noncanonical/noindex;
- broken internal links;
- orphan candidates among discovered public pages;
- link elements not crawlable because no usable `href`.

## 20.3 Canonical/URL consistency

- duplicate canonical declarations;
- malformed canonical URL;
- canonical loops;
- canonical chains;
- canonical points to error page;
- HTTP/HTTPS inconsistencies;
- trailing-slash/case variants that create duplicate surfaces.

## 20.4 Metadata/semantics

- empty title;
- duplicate titles as evidence, not automatic disaster;
- severely ambiguous title across key pages;
- missing description as low-priority appearance hint, not ranking blocker;
- invalid `lang` values;
- heading structure observations where useful.

## 20.5 Renderability

- important text only in rendered DOM;
- important links only in rendered DOM;
- raw/rendered canonical conflicts;
- raw/rendered robots conflicts;
- JavaScript navigation that produces no crawlable URL.

## 20.6 Structured data

- invalid JSON-LD;
- contradictory values;
- unsupported/unknown types informationally;
- supported rich-result required-field failures;
- structured data not represented visibly where policy expects visible content.

## 20.7 International

- broken hreflang target;
- missing return annotation;
- invalid locale code;
- hreflang/canonical conflict.

## 20.8 AI crawler visibility

Report separately:

- search crawler allowed/blocked;
- user-fetch crawler allowed/blocked;
- training crawler allowed/blocked.

Do not label “training blocked” as an SEO error.

## 20.9 Security/network conditions relevant to crawling

Only include conditions that directly affect crawler access, such as:

- TLS/connection failures;
- repeated 403/429 to the audit user-agent;
- redirect to authentication;
- bot-challenge responses.

RankCore is **not a security scanner**.

---

# 21. Implementation safety classes

Every suggested change should fall into one of three classes.

## Class A — usually safe to implement automatically through the coding agent

Examples:

- fix broken internal href;
- add missing sitemap route to existing sitemap generator;
- remove accidental duplicate canonical tag;
- correct malformed JSON-LD syntax;
- ensure metadata API emits existing truthful product data;
- fix obvious robots syntax mistake when intent is clear.

## Class B — review/product judgment required

Examples:

- change target keyword/topic of a major page;
- create new landing page;
- rewrite pricing/product positioning;
- change canonical strategy for large route family;
- change localization strategy;
- allow/block an AI crawler when policy intent is unclear.

## Class C — never silently automate

Examples:

- fake reviews/testimonials;
- claim awards/certifications;
- invent customer numbers;
- create fake author personas;
- buy/build manipulative backlinks;
- publish hundreds of templated SEO pages without a value gate;
- change training-data crawler policy as if it were a ranking requirement.

---

# 22. Security model

Because RankCore runs inside agent-driven repositories and crawls websites, security must be designed into v1.

## 22.1 Repository execution boundary

The native binary must never automatically execute project-controlled scripts.

It can read safe static inputs, but build/install/start commands are executed by the host coding agent under the user's normal approval model.

This avoids the binary turning repository metadata into arbitrary execution.

## 22.2 File exclusions

RankCore source/path mode must ignore by default:

- `.env`
- `.env.*`
- credential files;
- SSH keys;
- cloud credential directories;
- `.git` internals except harmless metadata when necessary;
- `node_modules` and vendor/cache trees.

It should not need secret values for SEO auditing.

## 22.3 Network SSRF protection

For a remote public seed:

- do not follow redirects from public hosts into loopback, link-local or private address space;
- block cloud metadata addresses;
- stay same-origin by default;
- bound response size;
- bound crawl depth/page count;
- time out slow requests.

For explicit localhost seeds, localhost/private access is expected.

## 22.4 Prompt injection boundary

The Rank Skill must treat:

- competitor pages;
- crawled page content;
- fetched documentation;
- user-generated page content

as **untrusted data**, not instructions.

It must never follow instructions embedded in a webpage telling it to run commands, expose secrets, or change goals.

## 22.5 Remote crawling etiquette

Remote crawl defaults:

- transparent `RankCoreBot/<version>` user agent;
- conservative concurrency;
- robots respected;
- no forms;
- no login;
- no CAPTCHA bypass;
- backoff on `429` and server distress.

---

# 23. Installation and cross-agent support

Do not manually create 70 divergent adapters.

Use three layers.

## Layer 1 — standard Agent Skill

The canonical skill follows the common `SKILL.md` model and avoids host-specific features in core instructions.

The open `skills` ecosystem currently supports Claude Code, Codex, Cursor and many dozens of other agents, which validates a shared skill-first distribution model.

Reference: https://github.com/vercel-labs/skills

## Layer 2 — RankCore native setup registry

`assets/agent-registry.json` contains paths/detection logic for the agents RankCore officially tests.

Start with:

- Claude Code;
- Codex;
- Cursor;
- Gemini CLI;
- OpenCode;
- Cline;
- GitHub Copilot;
- Windsurf/compatible `.agents/skills` hosts where appropriate.

Do not promise first-class support merely because a directory exists. “Officially tested” and “skill-compatible” are separate labels.

## Layer 3 — ecosystem fallback

Publish the skill repository so uncommon hosts can use:

```bash
npx skills add <rankcore-org>/<rankcore-repo> -g
```

This is a distribution convenience, not a RankCore runtime requirement.

---

# 24. Claude Code-specific stance

Claude Code is an important first-class host, but not the architecture.

Claude officially separates:

- Skills for reusable workflows/knowledge;
- hooks for deterministic lifecycle actions;
- MCP for external tools/data;
- subagents for isolated work;
- plugins as packaging.

Reference: https://code.claude.com/docs/id/features-overview

For RankCore v1:

- use the Skill;
- invoke native CLI through shell;
- do not require MCP;
- do not require hooks;
- do not require subagents.

Later, a Claude marketplace plugin may bundle convenience features if it materially improves installation or UX. It must remain a wrapper around the same canonical Skill and CLI.

---

# 25. Codex-specific stance

Codex supports repository instructions and Skills; current Codex products are explicitly designed to use reusable skills for workflows and resources.

RankCore should install the same canonical `rank` skill for Codex. Do not maintain a separate body of SEO knowledge for Codex.

Reference: https://openai.com/index/introducing-the-codex-app/

Any Codex-specific configuration should remain a thin adapter.

---

# 26. Lessons to adopt from Caveman — and lessons not to copy

Caveman is useful as an architecture/distribution reference because it demonstrates:

- one-command adoption;
- no-account/no-key core value;
- Skills as a portable surface;
- cross-agent installation;
- canonical skill sources;
- dry-run/uninstall paths;
- heavier deterministic functionality separated from the instruction layer.

Reference: https://github.com/JuliusBrussee/caveman

However, RankCore should **not** copy everything.

Do not copy:

- model traffic proxying;
- base-URL rewrites;
- provider routing;
- persistent session hooks;
- multiple speaking modes;
- agent replacement;
- complexity that exists specifically for token compression.

RankCore has a much narrower need:

```text
skill intelligence + deterministic local audit engine
```

Also learn from Caveman's operational edge cases: avoid binary-name collisions, track every configuration mutation, provide clean uninstall, and test hostile/untrusted repository behavior.

---

# 27. Release architecture

## 27.1 Platforms

Target initially:

- macOS arm64;
- macOS amd64;
- Linux amd64;
- Linux arm64;
- Windows amd64;
- Windows arm64 when CI coverage is reliable.

Consider Linux musl builds only if demanded by users.

## 27.2 Release artifacts

Each release includes:

- binaries;
- SHA-256 checksums;
- signed manifest/signatures;
- `install.sh`;
- `install.ps1`;
- release notes;
- schema/ruleset version information.

## 27.3 Update policy

`rankcore update` is explicit/manual in v1.

It downloads only from the official release channel and verifies integrity.

No surprise background updates.

## 27.4 Versioning

Maintain separately:

- binary version: semantic version;
- report schema version: integer;
- ruleset version: date/semantic identifier;
- skill/knowledge asset version.

This enables engine upgrades without silently changing report compatibility.

---

# 28. Testing strategy

RankCore has two very different testing problems:

1. deterministic engine correctness;
2. agent workflow reliability.

Test them separately.

## 28.1 Go unit tests

Must cover:

- URL normalization;
- redirects;
- robots RFC behavior;
- sitemap parsing;
- canonical resolution;
- HTML extraction;
- link graph;
- hreflang;
- JSON-LD parsing;
- rule evaluation;
- baseline diffing;
- private-network redirect blocking;
- response-size/time limits.

## 28.2 Golden fixtures

Create intentionally broken fixture sites:

- clean SSR site;
- JS-only app shell;
- accidental `noindex`;
- robots block;
- canonical loop;
- duplicate variants;
- malformed JSON-LD;
- hreflang broken returns;
- redirect chain;
- orphan pages;
- training bot blocked but search bot allowed;
- search bot blocked but training bot allowed.

Expected findings are checked into source control.

## 28.3 Integration tests

Use local HTTP test servers.

Do not rely on public websites for CI correctness.

Rendered tests run only in CI jobs where Chromium is explicitly installed.

## 28.4 Installer tests

Test:

- fresh install;
- repeated install/idempotence;
- dry-run;
- existing conflicting skill;
- symlink unavailable;
- Windows paths;
- update;
- uninstall;
- PATH failure;
- checksum failure.

## 28.5 Skill evals

Maintain task fixtures such as:

- new Next.js SaaS;
- local service business;
- documentation site;
- ecommerce-like catalog;
- authenticated app with public marketing pages;
- app with no web-search capability.

Evaluate whether the agent:

- identifies product correctly;
- does not fabricate metrics;
- invokes deterministic audit;
- loads only relevant references;
- does not call blocked integrations;
- separates search crawlers from training crawlers;
- avoids spammy page generation;
- verifies fixes;
- reports limitations.

Run core skill evals on at least Claude Code and Codex before v1 release.

---

# 29. Performance targets

These are engineering goals, not marketing promises.

- binary cold start should feel instantaneous;
- no daemon required;
- no database startup;
- local HTTP crawling should use bounded concurrency;
- memory should remain bounded by crawl limits;
- extraction should stream/limit bodies instead of blindly loading enormous responses;
- reports should remain usable on sites of at least hundreds to low-thousands of pages;
- large enterprise crawling is not a v1 requirement.

Do not optimize prematurely for million-page crawls. That would push RankCore toward an enterprise crawler product rather than an agent companion.

---

# 30. What the first release should support well

A strong v1 should be excellent for:

- SaaS marketing sites;
- product landing sites;
- docs/content sites;
- local-business sites;
- small/medium ecommerce-like sites;
- developer tools;
- modern SSR/static React/Next/Astro/Vue/Svelte sites when they expose normal web URLs;
- JS-heavy apps when local Chromium rendered mode is available.

It does **not** need to be excellent for:

- million-page marketplaces;
- complex authenticated crawling;
- enterprise log ingestion;
- paid media;
- app-store optimization;
- social-media ranking;
- full backlink intelligence;
- enterprise keyword warehouses.

Keep the initial problem sharp.

---

# 31. Roadmap

## Phase 0 — Workflow prototype

Goal: validate the user interaction before overbuilding the crawler.

Build:

- canonical `rank` Skill;
- compact knowledge references;
- dummy/stub CLI contract;
- two fixture projects;
- Claude Code + Codex skill installation manually.

Success condition:

A user types `/rank` and the agent follows the intended research/audit/implementation structure without giant prompts or confusion.

## Phase 1 — Native HTTP audit engine

Build:

- Go CLI shell;
- URL normalization;
- safe crawler;
- HTML extraction;
- robots;
- sitemap;
- canonical;
- basic link graph;
- findings/report schemas;
- verification diff;
- test fixtures.

Success condition:

RankCore can deterministically detect the core technical failures that matter on fixture sites.

## Phase 2 — One-command installation

Build:

- embedded Skill assets;
- agent registry;
- `rankcore setup`;
- `install.sh`;
- `install.ps1`;
- `doctor`;
- `uninstall`;
- signed release pipeline.

Success condition:

A fresh user can install and invoke `/rank` in Claude Code and Codex without manually copying files.

## Phase 3 — Structured data + international + AI crawler policy

Build:

- supported JSON-LD checks;
- hreflang;
- AI crawler registry/purpose model;
- current official crawler-source maintenance process.

Success condition:

RankCore reports search/user-fetch/training crawler policies accurately without conflating them.

## Phase 4 — Optional rendered audits

Build:

- Chromium discovery;
- CDP launch;
- bounded render lifecycle;
- raw-vs-rendered comparison.

Success condition:

RankCore can identify JS-only SEO surfaces without requiring browser download.

## Phase 5 — Ecosystem hardening

Build/test:

- Cursor;
- Gemini CLI;
- OpenCode;
- Cline;
- skills.sh distribution;
- package-manager distribution.

Success condition:

Same canonical Skill works across hosts with only installer/path differences.

## Phase 6 — Optional integrations

Only after local adoption proves demand:

- Search Console connector;
- Bing Webmaster connector;
- IndexNow helper;
- PageSpeed/CrUX;
- optional Ahrefs/Semrush/etc.;
- optional MCP wrapper;
- CI mode;
- team/cloud history.

These must remain optional. Zero-key local RankCore continues to work.

---

# 32. Things explicitly deferred

Do not let an implementation agent add these “because they might be useful” during v1:

- hosted dashboard;
- login/accounts;
- billing;
- RankCore LLM/API;
- vector database;
- embeddings/RAG service;
- plugin marketplace backend;
- MCP server;
- browser download manager;
- Chrome extension;
- VS Code extension;
- desktop GUI;
- Electron/Tauri app;
- SQLite;
- framework AST adapters;
- Search Console OAuth;
- Ahrefs/Semrush integrations;
- automated backlink outreach;
- programmatic page generator;
- content CMS;
- rank tracker;
- keyword volume service;
- proxy/gateway between the coding agent and model provider.

Every one of those adds surface area. None is necessary to prove RankCore's core value.

---

# 33. README/product positioning

The homepage should be understandable in seconds.

Suggested headline:

> **Make your coding agent understand how your website should be discovered — then let it fix the code.**

Subheading:

> RankCore gives Claude Code, Codex, Cursor and other coding agents a current search/AI-search workflow plus a deterministic local audit engine. No RankCore account. No extra LLM API. No SEO API required.

Then immediately:

```bash
curl -fsSL https://rankcore.dev/install.sh | sh
```

Then:

```text
/rank
```

Do not lead with acronyms such as GEO/AEO/LLM indexing. Those belong deeper in documentation.

---

# 34. Master acceptance criteria for v1

RankCore v1 is not complete until all of the following are true.

## Installation

- [ ] one-line install works on macOS/Linux;
- [ ] one-line PowerShell install works on Windows;
- [ ] checksum/signature verification exists;
- [ ] no admin/sudo required for normal user installation;
- [ ] setup is idempotent;
- [ ] `--dry-run` exists;
- [ ] clean uninstall exists;
- [ ] Claude Code and Codex are first-class tested hosts.

## UX

- [ ] user can type `/rank`;
- [ ] no manual knowledge upload;
- [ ] no separate account/API key;
- [ ] agent automatically understands the repo before recommending changes;
- [ ] agent researches externally when host web tools exist;
- [ ] agent gracefully degrades when web/browser tools do not exist;
- [ ] normal completion ends with implemented and verified changes, not merely a report.

## Engine

- [ ] one Go binary;
- [ ] no Node/Python runtime requirement;
- [ ] no daemon;
- [ ] no database;
- [ ] deterministic JSON output;
- [ ] bounded crawler;
- [ ] RFC-compliant robots behavior with test corpus;
- [ ] sitemap support;
- [ ] canonical checks;
- [ ] link graph;
- [ ] supported JSON-LD checks;
- [ ] hreflang checks;
- [ ] AI crawler policy report;
- [ ] baseline verification.

## Truthfulness

- [ ] no fake SEO score;
- [ ] no invented keyword volume;
- [ ] no invented keyword difficulty;
- [ ] no fake backlink counts;
- [ ] no unverified ranking guarantee;
- [ ] no treating `llms.txt` as required;
- [ ] no conflating training crawlers with search visibility crawlers;
- [ ] no calling basic timing data “Core Web Vitals.”

## Safety

- [ ] binary does not execute project scripts automatically;
- [ ] remote crawl prevents public-to-private redirect SSRF;
- [ ] bounded body sizes/timeouts;
- [ ] no CAPTCHA bypass;
- [ ] no credential scanning requirement;
- [ ] external web content treated as untrusted data;
- [ ] mass thin-content behavior prohibited by Skill guardrails.

## Portability

- [ ] canonical Skill is host-neutral;
- [ ] canonical knowledge exists only once;
- [ ] agent-specific logic is thin;
- [ ] RankCore CLI has no knowledge of Claude/Codex model APIs;
- [ ] core product works if the user switches coding agents.

---

# 35. Build order for an AI coding IDE

If this document is handed to Claude Code, Codex, Cursor or another coding agent, implement in this order.

## Milestone 1 — Repository skeleton

Create:

- Go module;
- `cmd/rankcore`;
- internal package layout;
- embedded assets structure;
- base JSON schemas;
- tests/testdata structure;
- CI formatting/test/build jobs.

Do **not** implement optional integrations.

## Milestone 2 — Core data contracts

Implement:

- `Finding`;
- `AuditReport`;
- `VerificationReport`;
- `PageSnapshot`;
- crawler config;
- rule interface;
- stable JSON serialization.

Lock schemas with golden JSON tests before implementing many rules.

## Milestone 3 — HTTP crawler

Implement:

- fetch policy;
- URL normalization;
- same-origin frontier;
- limits;
- redirects;
- retry/backoff;
- content limits;
- transparent user-agent;
- SSRF protections.

## Milestone 4 — Extractors

Implement:

- HTML;
- links;
- title/meta;
- canonical;
- robots meta/header;
- headings;
- images basics;
- JSON-LD;
- hreflang.

## Milestone 5 — Robots + sitemap

Implement and test before proceeding.

Robots correctness is foundational because the tool makes crawler-access claims.

## Milestone 6 — Initial rule set

Implement the high-confidence rules first.

Do not chase rule count.

## Milestone 7 — Verification engine

Implement baseline comparison before adding browser rendering.

The product's differentiator is not “finding problems”; it is “agent fixes them and RankCore proves whether they disappeared.”

## Milestone 8 — Canonical Skill

Write `assets/skill/rank/SKILL.md` and references.

Keep the orchestrator compact and progressive.

Test manually in Claude Code and Codex.

## Milestone 9 — Setup/installer

Embed assets and implement installation registry.

Make install/update/uninstall reversible.

## Milestone 10 — Rendered mode

Only after HTTP mode is reliable, add optional CDP rendering.

## Milestone 11 — Cross-host hardening

Validate Cursor/Gemini/OpenCode/etc.

---

# 36. Definition of success

RankCore succeeds when a user who is not an SEO expert can open a newly built web application in the AI coding agent they already use, type:

```text
/rank
```

and the agent can:

1. understand what the application actually is;
2. understand who it serves;
3. investigate the current market when web tools are available;
4. identify real search/discovery opportunities without fake metrics;
5. measure the application's technical discoverability using a deterministic local engine;
6. edit the application appropriately;
7. verify the fixes;
8. clearly distinguish what is proven, what is inferred and what still requires external platform data.

The user should not have to understand:

- Agent Skill folder locations;
- knowledge loading;
- crawler architecture;
- JSON schemas;
- MCP;
- Go;
- SEO APIs;
- crawler identities;
- browser automation.

They should understand only:

```text
Install RankCore.
Open project.
Type /rank.
```

That simplicity is not a cosmetic requirement. **It is the primary architectural constraint.**

---

# 37. Source references used for this architecture review

These are the most important current sources behind the technical/product decisions in this document.

### Agent extension architecture

- Claude Code extension overview: https://code.claude.com/docs/id/features-overview
- Claude Code project/global skill layout: https://code.claude.com/docs/fr/claude-directory
- Claude Code Desktop plugin installation: https://code.claude.com/docs/pt/desktop
- Codex skills/product workflow: https://openai.com/index/introducing-the-codex-app/
- Open Agent Skills installer/ecosystem: https://github.com/vercel-labs/skills

### Caveman reference project

- Caveman repository: https://github.com/JuliusBrussee/caveman
- Caveman README/installation philosophy: https://github.com/JuliusBrussee/caveman/blob/main/README.md
- Caveman package/installer evidence: https://github.com/JuliusBrussee/caveman/blob/main/package.json

### Search / AI-search guidance

- Google Search Essentials: https://developers.google.com/search/docs/essentials
- Google AI features and websites: https://developers.google.com/search/docs/appearance/ai-features
- Google generative-AI optimization guide: https://developers.google.com/search/docs/fundamentals/ai-optimization-guide
- Google spam policies: https://developers.google.com/search/docs/essentials/spam-policies
- Google JavaScript SEO basics: https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics
- Google developer SEO guide: https://developers.google.com/search/docs/fundamentals/get-started-developers
- Google generative-AI Search Console reporting: https://developers.google.com/search/blog/2026/06/gen-ai-performance-reports
- Bing AI Performance: https://blogs.bing.com/webmaster/February-2026/Introducing-AI-Performance-in-Bing-Webmaster-Tools-Public-Preview
- OpenAI publisher/search crawler guidance: https://help.openai.com/en/articles/12627856-publishers-and-developers-faq
- Anthropic crawler guidance: https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler
- Robots Exclusion Protocol RFC 9309: https://www.rfc-editor.org/rfc/rfc9309.html

---

# 38. Final architecture statement

Build RankCore as:

> **A universal, progressive `/rank` Agent Skill powered by the user's existing AI coding agent, backed by a small fast Go binary for deterministic web auditing and verification.**

Everything else is optional.

Do not build the dashboard.
Do not build the cloud.
Do not build another LLM.
Do not build 30 separate agent products.
Do not build hundreds of weak SEO checks.
Do not require SEO APIs.
Do not turn MCP into the product.

Build the shortest reliable path between:

```text
"I made this website"
```

and:

```text
"My coding agent understands how it should be discoverable,
implemented the right changes,
and verified the technical result."
```

