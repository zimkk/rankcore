package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/zimkk/rankcore/assets"
	"github.com/zimkk/rankcore/internal/audit"
	"github.com/zimkk/rankcore/internal/render"
	"github.com/zimkk/rankcore/internal/report"
	"github.com/zimkk/rankcore/internal/rules"
	"github.com/zimkk/rankcore/internal/setup"
	"github.com/zimkk/rankcore/internal/verify"
	"github.com/zimkk/rankcore/internal/version"
)

// Exit codes (master plan §12.9).
const (
	exitOK    = 0
	exitError = 1
	exitGate  = 2
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitError)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "setup":
		runSetup(args)
	case "doctor":
		runDoctor(args)
	case "audit":
		runAudit(args)
	case "verify":
		runVerify(args)
	case "explain":
		runExplain(args)
	case "version", "--version", "-v":
		runVersion(args)
	case "uninstall":
		runUninstall(args)
	case "update":
		runUpdate()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(exitError)
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
	fmt.Println("  audit      Run deterministic SEO audit against a URL or local directory")
	fmt.Println("  verify     Re-run checks and compare against a baseline")
	fmt.Println("  explain    Print documentation for a specific rule")
	fmt.Println("  version    Show version information")
	fmt.Println("  uninstall  Remove installed RankCore skills")
	fmt.Println("  update     Show how to update RankCore")
	fmt.Println()
	fmt.Println("Run 'rankcore <command> -h' for command flags.")
}

// parseArgs parses flags that may appear before or after positional arguments
// (the standard flag package stops at the first positional).
func parseArgs(fs *flag.FlagSet, args []string) []string {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			os.Exit(exitError)
		}
		args = fs.Args()
		if len(args) == 0 {
			return positional
		}
		if args[0] == "--" {
			return append(positional, args[1:]...)
		}
		positional = append(positional, args[0])
		args = args[1:]
	}
}

func loadRegistry() *setup.Registry {
	regFile, err := assets.FS.Open("agent-registry.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read embedded agent registry: %v\n", err)
		os.Exit(exitError)
	}
	defer regFile.Close()

	reg, err := setup.LoadRegistry(regFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse agent registry: %v\n", err)
		os.Exit(exitError)
	}
	return reg
}

// ── setup ───────────────────────────────────────────────────────────────────

func runSetup(args []string) {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	agentFlag := fs.String("agent", "", "Specific agent to set up (e.g. claude-code); installs even if not detected")
	dryRunFlag := fs.Bool("dry-run", false, "Simulate setup without writing files")
	allFlag := fs.Bool("all", false, "Install for every agent in the registry, detected or not")
	parseArgs(fs, args)

	reg := loadRegistry()

	fmt.Println("RankCore setup")
	fmt.Println()

	detected := setup.DetectAgents(reg)
	detectedIDs := map[string]bool{}
	for _, a := range detected {
		detectedIDs[a.ID] = true
	}

	fmt.Println("Detected:")
	for _, a := range reg.Agents {
		mark := "○"
		if detectedIDs[a.ID] {
			mark = "✓"
		}
		fmt.Printf("  %s %s\n", mark, a.Name)
	}
	fmt.Println()

	targets := detected
	switch {
	case *agentFlag != "":
		targets = nil
		for _, a := range reg.Agents {
			if a.ID == *agentFlag {
				targets = append(targets, a)
				break
			}
		}
		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "error: agent %q not found in registry\n", *agentFlag)
			os.Exit(exitError)
		}
	case *allFlag:
		targets = reg.Agents
	}

	if len(targets) == 0 {
		fmt.Println("No supported coding agents detected. Use --agent <id> or --all to install anyway.")
		return
	}

	fmt.Println("Installing /rank:")
	failed := false
	for _, agent := range targets {
		if err := setup.InstallSkill(agent, assets.FS, *dryRunFlag); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", agent.Name, err)
			failed = true
		} else if !*dryRunFlag {
			for _, p := range agent.Paths {
				fmt.Printf("  ✓ %s\n", setup.ExpandHome(p))
			}
		}
	}
	fmt.Println()
	fmt.Printf("Engine:\n  ✓ rankcore %s\n\n", version.Version)
	if failed {
		os.Exit(exitError)
	}
	fmt.Println("Ready. Open a web project and type /rank.")
}

// ── doctor ──────────────────────────────────────────────────────────────────

type doctorReport struct {
	Status         string        `json:"status"`
	Version        string        `json:"version"`
	RulesetVersion string        `json:"ruleset_version"`
	SchemaVersion  int           `json:"schema_version"`
	RuleCount      int           `json:"rule_count"`
	Platform       string        `json:"platform"`
	Agents         []doctorAgent `json:"agents"`
	Browser        doctorBrowser `json:"browser"`
}

type doctorAgent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Detected       bool   `json:"detected"`
	SkillDir       string `json:"skill_dir"`
	SkillInstalled bool   `json:"skill_installed"`
	Writable       bool   `json:"writable"`
}

type doctorBrowser struct {
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	parseArgs(fs, args)

	reg := loadRegistry()
	dr := doctorReport{
		Status:         "ok",
		Version:        version.Version,
		RulesetVersion: version.RulesetVersion,
		SchemaVersion:  version.SchemaVersion,
		RuleCount:      rules.NewRegistry().Count(),
		Platform:       runtime.GOOS + "/" + runtime.GOARCH,
	}

	detected := map[string]bool{}
	for _, d := range setup.DetectAgents(reg) {
		detected[d.ID] = true
	}
	for _, a := range reg.Agents {
		da := doctorAgent{ID: a.ID, Name: a.Name, Detected: detected[a.ID]}
		if len(a.Paths) > 0 {
			da.SkillDir = setup.ExpandHome(a.Paths[0])
			_, err := os.Stat(da.SkillDir + string(os.PathSeparator) + "SKILL.md")
			da.SkillInstalled = err == nil
			da.Writable = setup.Writable(da.SkillDir)
		}
		dr.Agents = append(dr.Agents, da)
	}

	browserPath := render.FindChromium()
	dr.Browser.Available = browserPath != ""
	dr.Browser.Path = browserPath

	if *jsonFlag {
		data, _ := json.MarshalIndent(dr, "", "  ")
		fmt.Println(string(data))
		return
	}

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
		notes := []string{}
		if a.SkillInstalled {
			notes = append(notes, "skill installed")
		} else if a.Detected {
			notes = append(notes, "run 'rankcore setup'")
		}
		if a.Detected && !a.Writable {
			notes = append(notes, "not writable")
		}
		suffix := ""
		if len(notes) > 0 {
			suffix = " (" + strings.Join(notes, ", ") + ")"
		}
		fmt.Printf("    %s %s%s\n", mark, a.Name, suffix)
	}

	fmt.Println()
	if dr.Browser.Available {
		fmt.Printf("  Browser: ✓ %s\n", dr.Browser.Path)
	} else {
		fmt.Println("  Browser: ○ No Chromium detected (rendered mode unavailable)")
	}

	fmt.Printf("\n  Status: %s\n", dr.Status)
}

// ── audit ───────────────────────────────────────────────────────────────────

type auditFlags struct {
	renderMode, browser, include, exclude, userAgent string
	maxPages, maxDepth, maxRender, concurrency       int
	timeout                                          time.Duration
	ignoreRobots, noSitemap                          bool
}

func registerAuditFlags(fs *flag.FlagSet) *auditFlags {
	d := audit.DefaultOptions()
	f := &auditFlags{}
	fs.StringVar(&f.renderMode, "render", d.RenderMode, "Render mode: auto|off|required")
	fs.StringVar(&f.browser, "browser", "", "Path to a Chromium-family browser (default: auto-detect)")
	fs.IntVar(&f.maxPages, "max-pages", d.MaxPages, "Maximum URLs to crawl")
	fs.IntVar(&f.maxDepth, "max-depth", d.MaxDepth, "Maximum click depth from the start URL")
	fs.IntVar(&f.maxRender, "max-render", d.MaxRender, "Maximum pages to render when rendering is on")
	fs.IntVar(&f.concurrency, "concurrency", d.Concurrency, "Concurrent requests")
	fs.StringVar(&f.include, "include", "", "Only crawl URLs matching this regex/substring")
	fs.StringVar(&f.exclude, "exclude", "", "Skip URLs matching this regex/substring")
	fs.StringVar(&f.userAgent, "user-agent", "", "Custom user agent")
	fs.DurationVar(&f.timeout, "timeout", d.Timeout, "Per-request timeout")
	fs.BoolVar(&f.ignoreRobots, "ignore-robots", false, "Fetch URLs robots.txt disallows for RankCoreBot (only for sites you own)")
	fs.BoolVar(&f.noSitemap, "no-sitemap", false, "Do not read XML sitemaps")
	return f
}

func (f *auditFlags) options(target string) audit.Options {
	opts := audit.DefaultOptions()
	opts.Target = target
	opts.RenderMode = strings.ToLower(f.renderMode)
	opts.Browser = f.browser
	opts.MaxPages = f.maxPages
	opts.MaxDepth = f.maxDepth
	opts.MaxRender = f.maxRender
	opts.Concurrency = f.concurrency
	opts.Include = f.include
	opts.Exclude = f.exclude
	opts.Timeout = f.timeout
	opts.IgnoreRobots = f.ignoreRobots
	opts.NoSitemap = f.noSitemap
	if f.userAgent != "" {
		opts.UserAgent = f.userAgent
	}
	return opts
}

func validateRender(mode string) {
	switch mode {
	case "auto", "off", "required":
	default:
		fmt.Fprintf(os.Stderr, "error: --render must be auto, off or required (got %q)\n", mode)
		os.Exit(exitError)
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}

func runAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	outDir := fs.String("out", ".rankcore/runs/latest", "Output directory")
	jsonFlag := fs.Bool("json", false, "Print the audit JSON to stdout (progress goes to stderr)")
	format := fs.String("format", "both", "Artifacts to write: both (audit.json + audit.md) | json-only")
	failOn := fs.String("fail-on", "", "Exit 2 if any finding is at or above this severity: low|medium|high|critical")
	af := registerAuditFlags(fs)
	pos := parseArgs(fs, args)

	if len(pos) == 0 {
		fmt.Fprintln(os.Stderr, "error: audit requires a target URL or directory")
		fmt.Fprintln(os.Stderr, "usage: rankcore audit <url|dir> [flags]")
		os.Exit(exitError)
	}
	validateRender(strings.ToLower(af.renderMode))
	if *failOn != "" && report.SeverityRank(*failOn) == 0 {
		fmt.Fprintf(os.Stderr, "error: --fail-on must be low, medium, high or critical (got %q)\n", *failOn)
		os.Exit(exitError)
	}

	opts := af.options(pos[0])
	opts.OutDir = *outDir
	opts.Format = *format

	ctx, cancel := signalContext()
	defer cancel()

	var progress io.Writer = os.Stdout
	if *jsonFlag {
		progress = os.Stderr
	}
	fmt.Fprintf(progress, "RankCore audit: %s\n", opts.Target)
	fmt.Fprintf(progress, "  Max pages: %d | Concurrency: %d | Render: %s\n\n", opts.MaxPages, opts.Concurrency, opts.RenderMode)

	auditReport, err := audit.Run(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitError)
	}

	if err := audit.SaveReport(auditReport, opts.OutDir, opts.Format); err != nil {
		fmt.Fprintf(os.Stderr, "error saving report: %v\n", err)
		os.Exit(exitError)
	}

	if *jsonFlag {
		data, _ := json.MarshalIndent(auditReport, "", "  ")
		fmt.Println(string(data))
	} else {
		printAuditSummary(auditReport, opts.OutDir)
	}

	if *failOn != "" {
		threshold := report.SeverityRank(*failOn)
		for _, f := range auditReport.Findings {
			if report.SeverityRank(f.Severity) >= threshold {
				os.Exit(exitGate)
			}
		}
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

func runVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	baseline := fs.String("baseline", "", "Baseline audit.json to compare against")
	outDir := fs.String("out", ".rankcore/runs/verify", "Output directory")
	jsonFlag := fs.Bool("json", false, "Print the verification JSON to stdout")
	af := registerAuditFlags(fs)
	pos := parseArgs(fs, args)

	if len(pos) == 0 || *baseline == "" {
		fmt.Fprintln(os.Stderr, "error: verify requires a target and --baseline")
		fmt.Fprintln(os.Stderr, "usage: rankcore verify <url|dir> --baseline <audit.json> [flags]")
		os.Exit(exitError)
	}
	validateRender(strings.ToLower(af.renderMode))

	ctx, cancel := signalContext()
	defer cancel()

	var progress io.Writer = os.Stdout
	if *jsonFlag {
		progress = os.Stderr
	}
	fmt.Fprintf(progress, "RankCore verify: %s\n", pos[0])
	fmt.Fprintf(progress, "  Baseline: %s\n\n", *baseline)

	opts := af.options(pos[0])
	vr, newRun, err := verify.Run(ctx, verify.Options{
		Target:       pos[0],
		BaselinePath: *baseline,
		OutDir:       *outDir,
		Audit:        &opts,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitError)
	}

	if err := verify.SaveReport(vr, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "error saving report: %v\n", err)
		os.Exit(exitError)
	}
	if err := audit.SaveReport(newRun, *outDir, "both"); err != nil {
		fmt.Fprintf(os.Stderr, "error saving re-audit: %v\n", err)
		os.Exit(exitError)
	}

	if *jsonFlag {
		data, _ := json.MarshalIndent(vr, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Verification complete.")
		for _, status := range []string{"fixed", "still_present", "regressed", "not_retestable", "new"} {
			if c, ok := vr.Summary[status]; ok {
				fmt.Printf("  %s: %d\n", status, c)
			}
		}
		fmt.Printf("\n  Report: %s/verification.json\n", *outDir)
	}

	if verify.Gate(vr) {
		os.Exit(exitGate)
	}
}

// ── explain ─────────────────────────────────────────────────────────────────

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	pos := parseArgs(fs, args)

	registry := rules.NewRegistry()
	if len(pos) == 0 {
		fmt.Println("Available rules:")
		fmt.Println()
		for _, r := range registry.All() {
			fmt.Printf("  %-16s v%d  %s\n", r.ID(), r.Version(), rules.Name(r.ID()))
		}
		fmt.Printf("\nTotal: %d rules\n", registry.Count())
		fmt.Println("\nUsage: rankcore explain <rule-id>")
		return
	}

	rule := registry.Find(pos[0])
	if rule == nil {
		fmt.Fprintf(os.Stderr, "error: unknown rule %q\n", pos[0])
		fmt.Fprintln(os.Stderr, "Run 'rankcore explain' to list all rules.")
		os.Exit(exitError)
	}

	fmt.Println(rules.Documentation(rule))
}

// ── version ─────────────────────────────────────────────────────────────────

// buildInfo returns the injected build time and the VCS commit, when known.
func buildInfo() (built, commit string) {
	built = version.BuildTime
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 12 {
				commit = s.Value[:12]
			}
		}
	}
	return built, commit
}

func runVersion(args []string) {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output in JSON format")
	parseArgs(fs, args)

	registry := rules.NewRegistry()

	if *jsonFlag {
		v := map[string]interface{}{
			"version":         version.Version,
			"ruleset_version": version.RulesetVersion,
			"schema_version":  version.SchemaVersion,
			"rule_count":      registry.Count(),
			"go_version":      runtime.Version(),
			"platform":        runtime.GOOS + "/" + runtime.GOARCH,
		}
		built, commit := buildInfo()
		if built != "" {
			v["build_time"] = built
		}
		if commit != "" {
			v["commit"] = commit
		}
		data, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(data))
		return
	}
	fmt.Printf("rankcore %s\n", version.Version)
	fmt.Printf("  Ruleset:  %s (%d rules)\n", version.RulesetVersion, registry.Count())
	fmt.Printf("  Schema:   v%d\n", version.SchemaVersion)
	fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  Go:       %s\n", runtime.Version())
}

// ── uninstall ───────────────────────────────────────────────────────────────

func runUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	dryRunFlag := fs.Bool("dry-run", false, "Show what would be removed")
	parseArgs(fs, args)

	reg := loadRegistry()

	fmt.Println("RankCore uninstall")
	fmt.Println()

	removed := 0
	for _, agent := range reg.Agents {
		for _, p := range agent.Paths {
			expanded := setup.ExpandHome(p)
			// Only remove directories that hold a RankCore skill.
			if !setup.IsRankSkillDir(expanded) {
				continue
			}
			if *dryRunFlag {
				fmt.Printf("  [DRY RUN] Would remove: %s\n", expanded)
				continue
			}
			if err := os.RemoveAll(expanded); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ Failed to remove %s: %v\n", expanded, err)
			} else {
				fmt.Printf("  ✓ Removed: %s\n", expanded)
				removed++
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
	fmt.Printf("  Installed: %s\n\n", version.Version)
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
