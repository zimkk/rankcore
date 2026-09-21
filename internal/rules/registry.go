package rules

import (
	"rankcore/internal/crawl"
	"rankcore/internal/report"
)

// Registry holds all loaded deterministic rules.
type Registry struct {
	rules []Rule
}

// NewRegistry initializes the registry with all standard rules.
func NewRegistry() *Registry {
	r := &Registry{}
	
	// Register all rules
	r.rules = append(r.rules,
		&HTTP5xxRule{},
		&HTTP4xxRule{},
		&MetaNoindexRule{},
		&CanonicalTargetRule{},
		&BrokenLinkRule{},
		// Additional rules would be appended here
	)
	
	return r
}

// EvaluateAll runs all registered rules against a page snapshot.
func (r *Registry) EvaluateAll(snapshot *crawl.PageSnapshot) []report.Finding {
	var findings []report.Finding
	
	for _, rule := range r.rules {
		if finding := rule.Evaluate(snapshot); finding != nil {
			findings = append(findings, *finding)
		}
	}
	
	return findings
}
