package rules

import (
	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// RC-IMG-001: Images missing alt attributes (Lighthouse: image-alt)
type ImageAltRule struct{}

func (r *ImageAltRule) ID() string   { return "RC-IMG-001" }
func (r *ImageAltRule) Version() int { return 1 }

func (r *ImageAltRule) Evaluate(snapshot *crawl.PageSnapshot) *report.Finding {
	if snapshot.StatusCode != 200 {
		return nil
	}
	if len(snapshot.Images) == 0 {
		return nil
	}

	var missing []string
	for _, img := range snapshot.Images {
		if !img.HasAlt {
			src := img.Src
			if src == "" {
				src = "(inline/empty src)"
			}
			missing = append(missing, src)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	// Cap the evidence list to avoid huge payloads
	shown := missing
	if len(shown) > 10 {
		shown = shown[:10]
	}

	return &report.Finding{
		ID:          r.ID(),
		RuleVersion: r.Version(),
		Category:    "content",
		Severity:    "medium",
		Confidence:  1.0,
		URL:         snapshot.URL,
		Summary:     "Images are missing alt attributes",
		Remediation: "Add descriptive alt text to all <img> elements.",
		Evidence: map[string]interface{}{
			"images_without_alt": shown,
			"total_missing":      len(missing),
			"total_images":       len(snapshot.Images),
		},
	}
}
