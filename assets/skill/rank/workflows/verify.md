# Verify Workflow

Use for `/rank verify`.

1. Locate the baseline audit and the implementation target.
2. Ensure the target represents the changed code, not an unrelated deployment.
3. Run repository tests/build checks when appropriate.
4. Run:

```bash
rankcore verify <target> --baseline <audit.json> --out <dir> --json
```

5. Compare by rule ID, URL/entity, and verification condition.
6. Classify:
   - fixed;
   - still present;
   - changed/needs interpretation;
   - not re-testable due to coverage/environment.
7. Never report an issue as fixed solely because code was edited.
8. Flag regressions introduced by the optimization.
9. Keep the final result short: fixed, unresolved, new regressions, limitations.
