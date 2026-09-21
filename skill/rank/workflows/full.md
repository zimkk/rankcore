# Full Workflow

Use for `/rank` with no narrower mode.

## 1. Check environment

Run `rankcore doctor --json`.

Record:
- binary availability/version;
- supported features;
- browser/render availability;
- host limitations.

## 2. Understand the repository and product

Inspect relevant source, routes, public pages, package/build metadata, content, and existing SEO configuration.

Create/update `.rankcore/profile.json`.

Do not ask routine questions that the repository answers.

## 3. Load minimum references

Start with:
- `search-foundations.md`
- `guardrails.md`

Load additional references only when their phase becomes relevant.

## 4. Research

If the host provides web search, perform current qualitative research:
- category language;
- representative competitors;
- search intent clusters;
- comparison/alternative/use-case/integration patterns;
- credible information sources;
- AI/search-engine guidance where needed.

Persist useful conclusions to `.rankcore/research.md`.

Never invent numeric keyword or backlink metrics.

## 5. Determine audit target

Prefer a local instance started through the repository's existing workflow. The host agent may run the project; the RankCore binary must not independently execute arbitrary project scripts.

Fallbacks:
- preview URL;
- production URL;
- filesystem/static-output mode;
- repository-only advisory analysis.

## 6. Audit

Run `rankcore audit`.

Read the machine report before planning remediation.

Separate:
- critical eligibility failures;
- technical consistency failures;
- contextual recommendations;
- content/research opportunities.

## 7. Plan

Write a short prioritized plan to `.rankcore/plan.md`.

Prioritize fixes that unlock discovery/indexability and correct contradictions before adding new content.

## 8. Implement

Edit the repository using the host coding agent.

Prefer minimal changes that preserve product behavior and design.

For new content, require a real user job and meaningful unique value.

## 9. Verify

Run repository tests/builds as appropriate, then `rankcore verify`.

Do not mark a deterministic issue resolved based only on code inspection.

## 10. Report

Tell the user:
- what RankCore understood;
- what was changed;
- what verification passed/failed;
- what remains external or uncertain.

Do not promise ranking outcomes.
