package rules

import (
	"encoding/json"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RC-STRUCT-001: Invalid JSON-LD (Lighthouse: structured-data)
type StructuredDataValidRule struct{}

func (r *StructuredDataValidRule) ID() string   { return "RC-STRUCT-001" }
func (r *StructuredDataValidRule) Version() int { return 1 }

func (r *StructuredDataValidRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	if len(snapshot.JSONLD) == 0 {
		return nil
	}

	var invalidBlocks []int // 1-indexed positions of invalid JSON-LD blocks
	for i, raw := range snapshot.JSONLD {
		if !json.Valid([]byte(raw)) {
			invalidBlocks = append(invalidBlocks, i+1)
		}
	}

	if len(invalidBlocks) == 0 {
		return nil
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "structured-data",
		Severity:    "high",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Page contains invalid JSON-LD structured data",
		Remediation: "Fix the JSON syntax in the <script type=\"application/ld+json\"> blocks so they parse as valid JSON.",
		Evidence: map[string]interface{}{
			"invalid_block_positions": invalidBlocks,
			"total_blocks":            len(snapshot.JSONLD),
		},
	}
}
