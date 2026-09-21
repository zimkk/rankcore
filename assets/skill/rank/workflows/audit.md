# Audit Workflow

Use for `/rank audit`.

1. Run `rankcore doctor --json`.
2. Inspect enough repository context to distinguish public, private, duplicate, staging, and utility surfaces.
3. Load:
   - `technical-search.md`
   - `structured-data-entities.md` when markup exists
   - `ai-search-crawlers.md` for crawler policy
   - `guardrails.md`
4. Determine the safest valid audit target.
5. Run `rankcore audit <target> --out ... --json`.
6. Interpret findings in context.
7. Do not convert every recommendation into an error.
8. Do not use a synthetic “SEO score” as the main result.
9. Report:
   - critical/high-confidence failures;
   - important inconsistencies;
   - candidate improvements requiring judgment;
   - audit coverage limitations.

If the binary is unavailable, explicitly call the result a repository/manual review rather than a RankCore deterministic audit.
