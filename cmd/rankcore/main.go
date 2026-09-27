package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/zimkk/rankcore/assets"
	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/rules"
	"github.com/zimkk/rankcore/internal/setup"
	"github.com/zimkk/rankcore/internal/verify"
)

const (
	version        = "1.0.0"
	rulesetVersion = "2026.09"
	schemaVersion  = 1
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "setup":
		runSetup()
	case "doctor":
		runDoctor()
	case "audit":
		runAudit()
	case "verify":
		runVerify()
	case "explain":
		runExplain()
	case "version":
		runVersion()
	case "uninstall":
		runUninstall()
	case "update":
		runUpdate()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("RankCore — Deterministic SEO Audit Engine")
	fmt.Println()
	fmt.Println("Usage: rankcore <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  setup      Install/update the rank skill into detected agents")
	fmt.Println("  doctor     Check environment health")
	fmt.Println("  audit      Run deterministic SEO audit against a target")
	fmt.Println("  verify     Re-run checks and compare against a baseline")
	fmt.Println("  explain    Print documentation for a specific rule")
	fmt.Println("  version    Show version information")
	fmt.Println("  uninstall  Remove installed RankCore skills and metadata")
	fmt.Println("  update     Check for and install updates from GitHub")
}

// ── setup ───────────────────────────────────────────────────────────────────

func runSetup() {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	agentFlag := fs.String("agent", "", "Specific agent to setup (e.g. claude-code)")
	dryRunFlag := fs.Bool("dry-run", false, "Simulate setup without writing files")
	allFlag := fs.Bool("all", false, "Install for all detected agents")
	fs.Parse(os.Args[2:])

	regFile, err := assets.FS.Open("agent-registry.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read embedded agent registry: %v\n", err)
		os.Exit(1)
	}
	defer regFile.Close()

	reg, err := setup.LoadRegistry(regFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse agent registry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("RankCore setup")
	fmt.Println()

	detected := setup.DetectAgents(reg)

	fmt.Println("Detected:")
	for _, a := range detected {
		fmt.Printf("  ✓ %s\n", a.Name)
	}
	for _, a := range reg.Agents {
		found := false
		for _, d := range detected {
			if d.ID == a.ID {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  ○ %s\n", a.Name)
		}
	}
	fmt.Println()

	targets := detected
	if *agentFlag != "" {
		targets = nil
		for _, a := range reg.Agents {
			if a.ID == *agentFlag {
				targets = append(targets, a)
				break
			}
		}
		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "error: agent %q not found in registry\n", *agentFlag)
			os.Exit(1)
		}
	}
	_ = allFlag

	fmt.Println("Installing /rank:")
	for _, agent := range targets {
		if err := setup.InstallSkill(agent, assets.FS, *dryRunFlag); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", agent.Name, err)
		} else {
			homeDir, _ := os.UserHomeDir()
			for _, p := range agent.Paths {
				expanded := strings.Replace(p, "~", homeDir, 1)
				fmt.Printf("  ✓ %s\n", expanded)
			}
		}
	}
	fmt.Println()
	fmt.Printf("Engine:\n  ✓ rankcore %s\n\n", version)
	fmt.Println("Ready. Open a web project and type /rank.")
}

// ── doctor ──────────────────────────────────────────────────────────────────

type doctorReport struct {
	Status         string            `json:"status"`
	Version        string            `json:"version"`
	RulesetVersion string            `json:"ruleset_version"`
	SchemaVersion  int               `json:"schema_version"`
	RuleCount      int               `json:"rule_count"`
	Platform       string            `json:"platform"`
	Agents         []doctorAgent     `json:"agents"`
	Browser        doctorBrowser     `json:"browser"`
}

type doctorAgent struct {
	Name     string `json:"name"`
	Detected bool   `json:"detected"`
	SkillDir string `json:"skill_dir"`
	Writable bool   `json:"writable"`
}

type doctorBrowser struct {
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
}

func runDoctor() {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(os.Args[2:])

	registry := rules.NewRegistry()

	regFile, err := assets.FS.Open("agent-registry.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer regFile.Close()

	reg, _ := setup.LoadRegistry(regFile)

	dr := doctorReport{
		Status:         "ok",
		Version:        version,
		RulesetVersion: rulesetVersion,
		SchemaVersion:  schemaVersion,
		RuleCount:      registry.Count(),
		Platform:       runtime.GOOS + "/" + runtime.GOARCH,
	}

	homeDir, _ := os.UserHomeDir()

	if reg != nil {
		detected := setup.DetectAgents(reg)
		detectedMap := map[string]bool{}
		for _, d := range detected {
			detectedMap[d.ID] = true
		}
		for _, a := range reg.Agents {
			skillDir := ""
			writable := false
			if len(a.Paths) > 0 {
				skillDir = strings.Replace(a.Paths[0], "~", homeDir, 1)
				if dir := filepath.Dir(skillDir); dir != "" {
					if f, err := os.OpenFile(filepath.Join(dir, ".rankcore_test"), os.O_CREATE|os.O_WRONLY, 0644); err == nil {
						f.Close()
						os.Remove(filepath.Join(dir, ".rankcore_test"))
						writable = true
					}
				}
			}
			dr.Agents = append(dr.Agents, doctorAgent{
				Name:     a.Name,
				Detected: detectedMap[a.ID],
				SkillDir: skillDir,
				Writable: writable,
			})
		}
	}

	// Check for Chromium
	browserPath := findChromium()
	dr.Browser.Available = browserPath != ""
	dr.Browser.Path = browserPath

	if *jsonFlag {
		data, _ := json.MarshalIndent(dr, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("RankCore Doctor\n\n")
		fmt.Printf("  Version:  %s\n", dr.Version)
		fmt.Printf("  Ruleset:  %s (%d rules)\n", dr.RulesetVersion, dr.RuleCount)
		fmt.Printf("  Platform: %s\n\n", dr.Platform)

		fmt.Println("  Agents:")
		for _, a := range dr.Agents {
			mark := "○"
			if a.Detected {
				mark = "✓"
			}
			writeStatus := ""
			if a.Detected && !a.Writable {
				writeStatus = " (not writable)"
			}
			fmt.Printf("    %s %s%s\n", mark, a.Name, writeStatus)
		}

		fmt.Println()
		if dr.Browser.Available {
			fmt.Printf("  Browser: ✓ %s\n", dr.Browser.Path)
		} else {
			fmt.Println("  Browser: ○ No Chromium detected (rendered mode unavailable)")
		}

		fmt.Printf("\n  Status: %s\n", dr.Status)
	}
}

func findChromium() string {
	candidates := []string{}
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "linux":
		candidates = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/microsoft-edge",
		}
	case "windows":
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		localAppData := os.Getenv("LOCALAPPDATA")
		candidates = []string{
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"),
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// ── audit ───────────────────────────────────────────────────────────────────

func runAudit() {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	outDir := fs.String("out", ".rankcore/runs/latest", "Output directory")
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	renderMode := fs.String("render", "off", "Render mode: auto|off|required")
	maxPages := fs.Int("max-pages", 500, "Maximum pages to crawl")
	concurrency := fs.Int("concurrency", 5, "Concurrent requests")
	include := fs.String("include", "", "URL pattern to include")
	exclude := fs.String("exclude", "", "URL pattern to exclude")
	format := fs.String("format", "json", "Output format: json|text|md")
	userAgent := fs.String("user-agent", "", "Custom user agent")
	fs.Parse(os.Args[2:])

	target := fs.Arg(0)
	if target == "" {
		fmt.Fprintln(os.Stderr, "error: audit requires a target URL")
		fmt.Fprintln(os.Stderr, "usage: rankcore audit <url> [flags]")
		os.Exit(1)
	}

	opts := audit.DefaultOptions()
	opts.Target = target
	opts.OutDir = *outDir
	opts.RenderMode = *renderMode
	opts.MaxPages = *maxPages
	opts.Concurrency = *concurrency
	opts.Include = *include
	opts.Exclude = *exclude
	opts.Format = *format
	if *userAgent != "" {
		opts.UserAgent = *userAgent
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Println("\nInterrupted. Stopping crawl...")
		cancel()
	}()

	fmt.Printf("RankCore audit: %s\n", target)
	fmt.Printf("  Max pages: %d | Concurrency: %d | Render: %s\n\n", opts.MaxPages, opts.Concurrency, opts.RenderMode)

	auditReport, err := audit.Run(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := audit.SaveReport(auditReport, opts.OutDir, opts.Format); err != nil {
		fmt.Fprintf(os.Stderr, "error saving report: %v\n", err)
		os.Exit(1)
	}

	if *jsonFlag {
		data, _ := json.MarshalIndent(auditReport, "", "  ")
		fmt.Println(string(data))
	} else {
		printAuditSummary(auditReport, opts.OutDir)
	}
}

func printAuditSummary(r *report.AuditReport, outDir string) {
	fmt.Printf("Audit complete.\n\n")
	fmt.Printf("  Pages: %d analyzed, %d skipped\n", r.Coverage.PagesAnalyzed, r.Coverage.PagesSkipped)

	counts := map[string]int{}
	for _, f := range r.Findings {
		counts[f.Severity]++
	}
	fmt.Printf("  Findings: %d total", len(r.Findings))
	parts := []string{}
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		if c, ok := counts[sev]; ok {
			parts = append(parts, fmt.Sprintf("%d %s", c, sev))
		}
	}
	if len(parts) > 0 {
		fmt.Printf(" (%s)", strings.Join(parts, ", "))
	}
	fmt.Println()

	if len(r.Limitations) > 0 {
		fmt.Println()
		for _, l := range r.Limitations {
			fmt.Printf("  ⚠ %s\n", l)
		}
	}

	fmt.Printf("\n  Report: %s/audit.json\n", outDir)
}

// ── verify ──────────────────────────────────────────────────────────────────

func runVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	baseline := fs.String("baseline", "", "Baseline audit.json to compare against")
	outDir := fs.String("out", ".rankcore/runs/verify", "Output directory")
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(os.Args[2:])

	target := fs.Arg(0)
	if target == "" || *baseline == "" {
		fmt.Fprintln(os.Stderr, "error: verify requires a target URL and --baseline")
		fmt.Fprintln(os.Stderr, "usage: rankcore verify <url> --baseline <audit.json> [flags]")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		cancel()
	}()

	fmt.Printf("RankCore verify: %s\n", target)
	fmt.Printf("  Baseline: %s\n\n", *baseline)

	vr, err := verify.Run(ctx, verify.Options{
		Target:       target,
		BaselinePath: *baseline,
		OutDir:       *outDir,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := verify.SaveReport(vr, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "error saving report: %v\n", err)
		os.Exit(1)
	}

	if *jsonFlag {
		data, _ := json.MarshalIndent(vr, "", "  ")
		fmt.Println(string(data))
	} else {
		counts := map[string]int{}
		for _, r := range vr.Results {
			counts[r.Status]++
		}
		fmt.Println("Verification complete.")
		for _, status := range []string{"fixed", "still_present", "regressed", "not_retestable", "new"} {
			if c, ok := counts[status]; ok {
				fmt.Printf("  %s: %d\n", status, c)
			}
		}
		fmt.Printf("\n  Report: %s/verification.json\n", *outDir)
	}
}

// ── explain ─────────────────────────────────────────────────────────────────

func runExplain() {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	ruleID := fs.Arg(0)
	if ruleID == "" {
		// List all rules
		registry := rules.NewRegistry()
		fmt.Println("Available rules:")
		fmt.Println()
		for _, r := range registry.All() {
			fmt.Printf("  %s (v%d)\n", r.ID(), r.Version())
		}
		fmt.Printf("\nTotal: %d rules\n", registry.Count())
		fmt.Println("\nUsage: rankcore explain <rule-id>")
		return
	}

	registry := rules.NewRegistry()
	rule := registry.Find(ruleID)
	if rule == nil {
		fmt.Fprintf(os.Stderr, "error: unknown rule %q\n", ruleID)
		fmt.Fprintln(os.Stderr, "Run 'rankcore explain' to list all rules.")
		os.Exit(1)
	}

	doc := rules.Documentation(rule)
	fmt.Println(doc)
}

// ── version ─────────────────────────────────────────────────────────────────

func runVersion() {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(os.Args[2:])

	registry := rules.NewRegistry()

	if *jsonFlag {
		v := map[string]interface{}{
			"version":         version,
			"ruleset_version": rulesetVersion,
			"schema_version":  schemaVersion,
			"rule_count":      registry.Count(),
			"go_version":      runtime.Version(),
			"platform":        runtime.GOOS + "/" + runtime.GOARCH,
			"build_time":      time.Now().Format(time.RFC3339),
		}
		data, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("rankcore %s\n", version)
		fmt.Printf("  Ruleset:  %s (%d rules)\n", rulesetVersion, registry.Count())
		fmt.Printf("  Schema:   v%d\n", schemaVersion)
		fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  Go:       %s\n", runtime.Version())
	}
}

// ── uninstall ───────────────────────────────────────────────────────────────

func runUninstall() {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	dryRunFlag := fs.Bool("dry-run", false, "Show what would be removed")
	fs.Parse(os.Args[2:])

	regFile, err := assets.FS.Open("agent-registry.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer regFile.Close()

	reg, _ := setup.LoadRegistry(regFile)
	homeDir, _ := os.UserHomeDir()

	fmt.Println("RankCore uninstall")
	fmt.Println()

	removed := 0
	if reg != nil {
		for _, agent := range reg.Agents {
			for _, p := range agent.Paths {
				expanded := strings.Replace(p, "~", homeDir, 1)
				if _, err := os.Stat(expanded); err == nil {
					if *dryRunFlag {
						fmt.Printf("  [DRY RUN] Would remove: %s\n", expanded)
					} else {
						if err := os.RemoveAll(expanded); err != nil {
							fmt.Fprintf(os.Stderr, "  ✗ Failed to remove %s: %v\n", expanded, err)
						} else {
							fmt.Printf("  ✓ Removed: %s\n", expanded)
							removed++
						}
					}
				}
			}
		}
	}

	if removed == 0 && !*dryRunFlag {
		fmt.Println("  No installed skills found to remove.")
	}
	fmt.Println()
	fmt.Println("Note: To remove the rankcore binary, delete it manually from your installation directory.")
}

// ── update ──────────────────────────────────────────────────────────────────

func runUpdate() {
	fmt.Println("RankCore update")
	fmt.Println()
	fmt.Println("  To update RankCore, re-run the installer:")
	fmt.Println()
	switch runtime.GOOS {
	case "windows":
		fmt.Println("    irm https://raw.githubusercontent.com/zimkk/rankcore/main/install.ps1 | iex")
	default:
		fmt.Println("    curl -fsSL https://raw.githubusercontent.com/zimkk/rankcore/main/install.sh | sh")
	}
	fmt.Println()
	fmt.Println("  Or build from source:")
	fmt.Println("    go install github.com/zimkk/rankcore/cmd/rankcore@latest")
	fmt.Println()
	fmt.Println("  Latest releases: https://github.com/zimkk/rankcore/releases")
}
