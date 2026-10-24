package rules

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/report"
)

// requiredProps is the deliberately small registry of Google rich-result types
// and the properties Google documents as required for them. Each inner slice
// is a set of alternatives: at least one must be present.
var requiredProps = map[string][][]string{
	"Product":             {{"name"}, {"offers", "review", "aggregateRating"}},
	"BreadcrumbList":      {{"itemListElement"}},
	"FAQPage":             {{"mainEntity"}},
	"Event":               {{"name"}, {"startDate"}, {"location"}},
	"Recipe":              {{"name"}, {"image"}},
	"JobPosting":          {{"title"}, {"description"}, {"datePosted"}, {"hiringOrganization"}},
	"LocalBusiness":       {{"name"}, {"address"}},
	"VideoObject":         {{"name"}, {"thumbnailUrl"}, {"uploadDate"}},
	"SoftwareApplication": {{"name"}, {"aggregateRating", "review"}},
}

// jsonLDNodes flattens a JSON-LD document into its top-level entity objects,
// following arrays and @graph containers.
func jsonLDNodes(v interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	switch t := v.(type) {
	case []interface{}:
		for _, e := range t {
			out = append(out, jsonLDNodes(e)...)
		}
	case map[string]interface{}:
		if g, ok := t["@graph"]; ok {
			out = append(out, jsonLDNodes(g)...)
			if _, hasType := t["@type"]; !hasType {
				return out
			}
		}
		out = append(out, t)
	}
	return out
}

// jsonLDTypes returns the short schema.org type names of a node.
func jsonLDTypes(n map[string]interface{}) []string {
	var raw []string
	switch t := n["@type"].(type) {
	case string:
		raw = []string{t}
	case []interface{}:
		for _, e := range t {
			if s, ok := e.(string); ok {
				raw = append(raw, s)
			}
		}
	}
	for i, s := range raw {
		s = strings.TrimSuffix(s, "/")
		if j := strings.LastIndexAny(s, "/#:"); j >= 0 {
			s = s[j+1:]
		}
		raw[i] = s
	}
	return raw
}

func present(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	case []interface{}:
		return len(t) > 0
	case map[string]interface{}:
		return len(t) > 0
	}
	return true
}

// RC-STRUCT-002: supported rich-result types missing required properties.
type StructuredDataRequiredRule struct{}

func (r *StructuredDataRequiredRule) ID() string   { return "RC-STRUCT-002" }
func (r *StructuredDataRequiredRule) Version() int { return 1 }

func (r *StructuredDataRequiredRule) Evaluate(s *crawl.PageSnapshot) *report.Finding {
	if !isHTMLOK(s) || len(s.JSONLD) == 0 {
		return nil
	}

	var problems []string
	var types []string
	for i, raw := range s.JSONLD {
		var doc interface{}
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			continue // RC-STRUCT-001 reports syntax errors
		}
		for _, node := range jsonLDNodes(doc) {
			nodeTypes := jsonLDTypes(node)
			if len(nodeTypes) == 0 {
				problems = append(problems, fmt.Sprintf("block %d: entity has no @type", i+1))
				continue
			}
			types = append(types, nodeTypes...)
			for _, typ := range nodeTypes {
				reqs, ok := requiredProps[typ]
				if !ok {
					continue
				}
				for _, alts := range reqs {
					found := false
					for _, prop := range alts {
						if present(node[prop]) {
							found = true
							break
						}
					}
					if !found {
						problems = append(problems, fmt.Sprintf("block %d: %s missing %s", i+1, typ, strings.Join(alts, " or ")))
					}
				}
			}
		}
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(types)
	return &report.Finding{
		ID: r.ID(), RuleVersion: r.Version(), Severity: "medium", Confidence: 1.0, URL: s.URL,
		Summary:     "Structured data is missing properties required for rich results",
		Remediation: "Add the required properties with values that match what the page visibly shows. Do not invent ratings, reviews or prices.",
		Evidence:    map[string]interface{}{"problems": capList(problems, 20), "types": dedupe(types)},
	}
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
