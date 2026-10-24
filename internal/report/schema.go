package report

import (
	"time"
)

// AuditReport is the top-level structure for the deterministic audit results.
type AuditReport struct {
	SchemaVersion   int               `json:"schema_version"`
	RankCoreVersion string            `json:"rankcore_version"`
	RulesetVersion  string            `json:"ruleset_version"`
	GeneratedAt     time.Time         `json:"generated_at"`
	Target          string            `json:"target"`
	Origin          string            `json:"origin,omitempty"` // scheme://host actually crawled
	Mode            Mode              `json:"mode"`
	Coverage        Coverage          `json:"coverage"`
	Findings        []Finding         `json:"findings"`
	CrawlerAccess   map[string]Policy `json:"crawler_access,omitempty"`
	Sitemaps        []SitemapInfo     `json:"sitemaps,omitempty"`
	Pages           []PageSummary     `json:"pages,omitempty"`
	Limitations     []string          `json:"limitations,omitempty"`
}

type Mode struct {
	HTTP     bool   `json:"http"`
	Rendered bool   `json:"rendered"`
	Browser  string `json:"browser,omitempty"`
}

type Coverage struct {
	PagesRequested int  `json:"pages_requested"`
	PagesAnalyzed  int  `json:"pages_analyzed"`
	PagesSkipped   int  `json:"pages_skipped"`
	LimitReached   bool `json:"limit_reached"`
}

// Policy is the robots.txt access status of one crawler at the site root.
type Policy struct {
	Status      string `json:"status"`
	Operator    string `json:"operator,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
	MatchedRule string `json:"matched_rule,omitempty"`
}

// SitemapInfo describes one sitemap file RankCore attempted to read.
type SitemapInfo struct {
	URL        string `json:"url"`
	Source     string `json:"source"` // robots.txt or default
	StatusCode int    `json:"status_code,omitempty"`
	URLCount   int    `json:"url_count"`
	IsIndex    bool   `json:"is_index,omitempty"`
	Error      string `json:"error,omitempty"`
}

// PageSummary records what happened to each URL in the crawl. Verification
// uses it to tell "fixed" apart from "not retestable".
type PageSummary struct {
	URL        string `json:"url"`
	FinalURL   string `json:"final_url,omitempty"`
	StatusCode int    `json:"status_code"`
	Depth      int    `json:"depth"`
	Source     string `json:"source,omitempty"`
	InSitemap  bool   `json:"in_sitemap,omitempty"`
	Indexable  bool   `json:"indexable"`
	DurationMS int64  `json:"response_ms,omitempty"`
	Bytes      int    `json:"html_bytes,omitempty"`
	Skipped    string `json:"skipped,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Automation safety classes (master plan §21).
const (
	AutomationSafe   = "safe_auto"       // Class A
	AutomationReview = "review_required" // Class B
	AutomationManual = "manual_only"     // Class C
)

// Finding represents a single deterministic rule evaluation result.
type Finding struct {
	ID          string                 `json:"id"`
	RuleVersion int                    `json:"rule_version"`
	Category    string                 `json:"category"`
	Severity    string                 `json:"severity"`
	Confidence  float64                `json:"confidence"`
	URL         string                 `json:"url"`
	Evidence    map[string]interface{} `json:"evidence"`
	Summary     string                 `json:"summary"`
	Remediation string                 `json:"remediation,omitempty"`
	Automation  string                 `json:"automation,omitempty"`
	Verify      *VerifyInfo            `json:"verify,omitempty"`
}

type VerifyInfo struct {
	Type     string `json:"type"`
	Expected string `json:"expected"`
}

// VerificationReport is the top-level structure for verifying a previous audit.
type VerificationReport struct {
	SchemaVersion   int            `json:"schema_version"`
	RankCoreVersion string         `json:"rankcore_version"`
	GeneratedAt     time.Time      `json:"generated_at"`
	Target          string         `json:"target"`
	Baseline        string         `json:"baseline"`
	Summary         map[string]int `json:"summary,omitempty"`
	Results         []VerifyResult `json:"results"`
}

type VerifyResult struct {
	FindingID string `json:"finding_id"`
	URL       string `json:"url"`
	Severity  string `json:"severity,omitempty"`
	Status    string `json:"status"` // fixed, still_present, regressed, not_retestable, new
}

// SeverityRank orders severities from low (1) to critical (4).
func SeverityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
