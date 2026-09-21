# Contributing

Thank you for considering contributing to RankCore!

## Development Setup
1. Ensure you have Go 1.23+ installed.
2. Clone the repository.
3. Run `go test ./...` to ensure all tests pass.

## Adding Rules
Please refer to `AUDIT_RULE_CATALOG.md` for the criteria of what makes a good deterministic rule. Do not add arbitrary heuristic SEO rules like "Keyword density must be X%".

## Pull Requests
1. Fork the repo and create your branch from `main`.
2. Write tests for any new code.
3. Ensure `go fmt` and `go vet` pass.
4. Submit your PR with a clear description of the issue and solution.
