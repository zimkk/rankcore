# Fix Workflow

Use for `/rank fix`.

1. Read the latest `.rankcore` profile, plan, and audit if present.
2. If evidence is stale or absent, re-audit before broad edits.
3. Load only the references relevant to the chosen fixes.
4. Order work by:
   - eligibility;
   - canonical/discovery integrity;
   - content availability;
   - internal linking;
   - structured-data/entity consistency;
   - metadata/semantics;
   - high-value content gaps;
   - performance/UX.
5. Make small, reviewable code changes.
6. Preserve framework conventions and the application's intended behavior.
7. Do not mass-create pages without a value gate.
8. Do not manufacture testimonials, authors, expertise, locations, prices, reviews, or external citations.
9. Run appropriate repository checks.
10. Run `/rank verify` behavior before claiming completion.

If the requested fix depends on credentials or an external account, implement the local preparation and clearly identify the remaining external step instead of pretending it is complete.
