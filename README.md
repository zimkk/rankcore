<div align="center">
  
  
  <h1>RankCore</h1>
  <p><strong>Make your coding agent understand how your website should be discovered — then let it fix the code.</strong></p>

  <p>
    <a href="https://github.com/zimkk/rankcore/actions"><img src="https://img.shields.io/badge/build-passing-brightgreen?style=flat-square" alt="Build Status"></a>
    <a href="https://github.com/zimkk/rankcore/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
  </p>
</div>

<hr>

## 🚀 Overview

RankCore gives AI coding agents (like **Claude Code, Cursor, Codex, Gemini CLI, and Cline**) a current search/AI-search workflow paired with a deterministic local audit engine. 

Instead of relying on heuristic SEO advice from outdated data, RankCore executes a rigorous, real-time technical audit of your project and presents evidence-based fixes directly to your agent.

### Why RankCore?
- **Zero-Key:** No RankCore account or external SEO APIs required.
- **Local Native Binary:** Built in Go. No Node or Python runtime requirements.
- **Deterministic:** Only enforces strict, schema-first rules. No arbitrary SEO scores.
- **SSRF Protected:** Safe HTTP crawler with loopback protections built-in.

---

## ⚡ Features

- **Concurrent HTTP Crawler:** Respects `robots.txt`, scales efficiently, and crawls same-origin topologies automatically.
- **Streaming HTML Extraction:** Pulls out Canonicals, Hreflang maps, structured data (JSON-LD), Meta descriptions, and Titles efficiently.
- **Deterministic Rules Engine:** Audits for:
  - 🛑 `5xx` and `4xx` persistent errors
  - 🛑 Explicit `noindex` blocks in headers and Meta tags
  - 🛑 Broken internal links and redirect loops
  - 🛑 Canonical target discrepancies
- **Verification Diffing:** Compare baseline runs against new crawls to verify fixes.
- **Rendered Mode Support:** Hooks for Chromium CDP extraction for JS-heavy applications.

---

## 📦 Installation

RankCore can be installed quickly on any machine. It compiles to a single, portable binary.

### macOS / Linux
```bash
curl -fsSL https://rankcore.dev/install.sh | sh
```

### Windows (PowerShell)
```powershell
irm https://rankcore.dev/install.ps1 | iex
```

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

# Run an audit on a local or remote target
rankcore audit https://your-site.com

# Verify fixes against a baseline
rankcore verify --baseline .rankcore/runs/latest
```

---

## 🛠️ Architecture

RankCore is built on a clean, scalable Go module architecture:
- `internal/crawl`: The concurrent frontier and network layer (with SSRF protections).
- `internal/extract`: High-performance tokenization for metadata.
- `internal/rules`: The deterministic rule registry.
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
