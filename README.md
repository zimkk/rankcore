<div align="center">
  <h1>RankCore</h1>
  <p><strong>Make your coding agent understand how your website should be discovered — then let it fix the code.</strong></p>

  <p>
    <a href="https://github.com/zimkk/rankcore/actions/workflows/ci.yml"><img src="https://github.com/zimkk/rankcore/actions/workflows/ci.yml/badge.svg" alt="CI Status"></a>
    <a href="https://goreportcard.com/report/github.com/zimkk/rankcore"><img src="https://goreportcard.com/badge/github.com/zimkk/rankcore?style=flat-square" alt="Go Report Card"></a>
    <a href="https://github.com/zimkk/rankcore/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
  </p>
</div>

<hr>

## 🚀 Overview

RankCore gives AI coding agents (like **Claude Code, Cursor, Codex, Gemini CLI, and Cline**) a current search/AI-search workflow paired with a deterministic local audit engine. 

Instead of relying on heuristic SEO advice from outdated data, RankCore executes a rigorous, real-time technical audit of your project and presents evidence-based fixes directly to your agent.

> **Status:** Early development. The `/rank` skill and the deterministic audit engine work end to end; see the [master plan](RANKCORE_MASTER_PLAN.md) for the roadmap.

### Why RankCore?
- **Zero-Key:** No RankCore account or external SEO APIs required.
- **Local Native Binary:** Built in Go. No Node or Python runtime requirements.
- **Deterministic:** Only enforces strict, schema-first rules. No arbitrary SEO scores.
- **SSRF Protected:** Safe HTTP crawler with loopback protections built-in.

---

## ⚡ Features

- **Bounded concurrent crawler:** Same-origin frontier with page, depth and body-size limits, include/exclude patterns, `robots.txt` etiquette for remote sites, `429`/`503` backoff, and SSRF protection (validated-IP dialing, every redirect hop re-checked).
- **Sitemap-aware discovery:** Reads `Sitemap:` lines from robots.txt (or `/sitemap.xml`), follows sitemap indexes and gzip, and crawls every listed URL alongside linked pages.
- **35 deterministic rules** across availability, indexability, canonicals, links, metadata, content, internationalization, structured data, sitemaps, renderability and AI-crawler access. Highlights:
  - 🛑 `4xx`/`5xx` pages, redirect chains and loops, `noindex` in meta or `X-Robots-Tag`
  - 🛑 Broken and redirecting internal links, uncrawlable anchors, orphan pages, deep pages
  - 🛑 Canonical targets that 404, redirect, are noindexed or chain; conflicting canonicals
  - 🛑 Sitemap entries that error, redirect, are noindexed or non-canonical
  - 🛑 Duplicate titles and descriptions, hreflang without return links
  - 🛑 JSON-LD syntax errors and missing required properties for Google rich-result types
  - 🛑 Search and AI-search crawlers blocked in robots.txt (training crawlers are reported but never flagged)
- **Verification diffing:** Re-audits and classifies every baseline finding as fixed, still present, regressed, new or not re-testable, with CI-friendly exit codes.
- **Rendered mode:** Compares server HTML with the DOM rendered by an already-installed Chrome, Edge, Brave or Chromium.
- **Local builds:** Point `rankcore audit` at a static output directory and it serves it on a loopback port.

Run `rankcore explain` for the full rule catalog.

---

## 📦 Installation

RankCore installs directly from this GitHub repository — no custom domain, account, or API key required. The installer downloads the latest precompiled release from [GitHub Releases](https://github.com/zimkk/rankcore/releases), verifies its SHA-256 checksum, installs it to a user-writable directory (no sudo/admin), and runs `rankcore setup` to register the `/rank` skill with your detected coding agents.

### macOS / Linux / WSL
```bash
curl -fsSL https://raw.githubusercontent.com/zimkk/rankcore/main/install.sh | sh
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/zimkk/rankcore/main/install.ps1 | iex
```

### Build from source
Requires Go 1.26+:
```bash
go install github.com/zimkk/rankcore/cmd/rankcore@latest
```

> **Note:** Precompiled binaries are published on tagged releases. If the installer reports a missing release for your platform, use the source build above or download manually from the [releases page](https://github.com/zimkk/rankcore/releases).

---

## 💻 Usage

### Agent Workflow
After installation, RankCore automatically registers itself with your detected agents (e.g., `~/.cursor/skills/rank`, `~/.gemini/skills/rank`). 

Open your web project in your coding agent and type:
```
/rank
```

The agent will seamlessly:
1. Understand your project topology.
2. Run a deterministic local technical SEO audit.
3. Recommend fixes based strictly on evidence.
4. Apply the fixes through your agent and verify them.

### CLI Workflow
You can also use RankCore manually from your terminal:

```bash
# Check environment health
rankcore doctor

# Run an audit on a local or remote target, or a static build directory
rankcore audit https://your-site.com
rankcore audit http://localhost:3000 --render auto --max-pages 200
rankcore audit ./dist

# Machine-readable output (stdout is pure JSON) and a CI gate
rankcore audit https://your-site.com --json > audit.json
rankcore audit https://your-site.com --fail-on high

# Verify fixes against a baseline
rankcore verify https://your-site.com --baseline .rankcore/runs/latest/audit.json

# Explain a rule
rankcore explain RC-CANON-002
```

Exit codes: `0` success, `1` operational error, `2` gate failure (`--fail-on` threshold hit, or `verify` found regressions, new findings, or unresolved critical issues).

---

## 🛠️ Architecture

RankCore is built on a clean, scalable Go module architecture:
- `internal/crawl`: The concurrent frontier, URL normalization and page snapshots.
- `internal/httpx`: SSRF-safe HTTP client with manual redirect tracking and backoff.
- `internal/extract`: HTML parsing for metadata, links, headings, images, JSON-LD and visible text.
- `internal/robots`, `internal/sitemap`, `internal/aiaccess`: RFC 9309 robots, sitemaps, and the AI crawler registry.
- `internal/rules`: Page rules and site-wide rules (link graph, canonicals, sitemaps, duplicates).
- `internal/audit`, `internal/verify`: The audit pipeline and baseline diffing.
- `internal/report`: Golden JSON schemas for `AuditReport` and `Finding`.
- `internal/setup`: Agent integration and skill installation.

---

## 🤝 Contributing

We welcome contributions! Please review our [Contributing Guidelines](CONTRIBUTING.md) and [Security Policy](SECURITY.md) before submitting pull requests.

To run the test suite locally:
```bash
go test ./...
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

<div align="center">
  <sub>Built with ❤️ by <b><a href="https://github.com/zimkk">Hassan Nazir (ZIMKK)</a></b> and Contributors</sub>
</div>
