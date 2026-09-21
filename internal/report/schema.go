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
	Mode            Mode              `json:"mode"`
	Coverage        Coverage          `json:"coverage"`
	Findings        []Finding         `json:"findings"`
	CrawlerAccess   map[string]Policy `json:"crawler_access,omitempty"`
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

type Policy struct {
	Status string `json:"status"`
}

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
	SchemaVersion   int               `json:"schema_version"`
	RankCoreVersion string            `json:"rankcore_version"`
	GeneratedAt     time.Time         `json:"generated_at"`
	Target          string            `json:"target"`
	Baseline        string            `json:"baseline"`
	Results         []VerifyResult    `json:"results"`
}

type VerifyResult struct {
	FindingID string `json:"finding_id"`
	URL       string `json:"url"`
	Status    string `json:"status"` // fixed, still_present, regressed, not_retestable, new
}
