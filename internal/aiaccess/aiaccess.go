package aiaccess

import (
	"github.com/zimkk/rankcore/internal/robots"
)

// CrawlerPurpose describes why a bot crawls.
type CrawlerPurpose string

const (
	PurposeSearch   CrawlerPurpose = "search"
	PurposeUserFetch CrawlerPurpose = "user_fetch"
	PurposeTraining CrawlerPurpose = "training"
)

// CrawlerEntry describes a known AI/search crawler.
type CrawlerEntry struct {
	Name      string         `json:"name"`
	UserAgent string         `json:"user_agent"`
	Operator  string         `json:"operator"`
	Purpose   CrawlerPurpose `json:"purpose"`
}

// PolicyResult describes the access status for a specific crawler.
type PolicyResult struct {
	Crawler CrawlerEntry `json:"crawler"`
	Allowed bool         `json:"allowed"`
	Rule    string       `json:"matched_rule,omitempty"`
}

// Report holds all crawler access evaluations for a path.
type Report struct {
	Path    string         `json:"path"`
	Results []PolicyResult `json:"results"`
}

// KnownCrawlers is the registry of known AI and search crawlers,
// based on official documentation as of 2026-09.
var KnownCrawlers = []CrawlerEntry{
	// Google
	{Name: "Googlebot", UserAgent: "Googlebot", Operator: "Google", Purpose: PurposeSearch},
	{Name: "Googlebot-Image", UserAgent: "Googlebot-Image", Operator: "Google", Purpose: PurposeSearch},
	{Name: "Google-Extended", UserAgent: "Google-Extended", Operator: "Google", Purpose: PurposeTraining},

	// Bing
	{Name: "Bingbot", UserAgent: "Bingbot", Operator: "Microsoft", Purpose: PurposeSearch},

	// OpenAI
	{Name: "OAI-SearchBot", UserAgent: "OAI-SearchBot", Operator: "OpenAI", Purpose: PurposeSearch},
	{Name: "ChatGPT-User", UserAgent: "ChatGPT-User", Operator: "OpenAI", Purpose: PurposeUserFetch},
	{Name: "GPTBot", UserAgent: "GPTBot", Operator: "OpenAI", Purpose: PurposeTraining},

	// Anthropic
	{Name: "Claude-SearchBot", UserAgent: "Claude-SearchBot", Operator: "Anthropic", Purpose: PurposeSearch},
	{Name: "Claude-User", UserAgent: "Claude-User", Operator: "Anthropic", Purpose: PurposeUserFetch},
	{Name: "ClaudeBot", UserAgent: "ClaudeBot", Operator: "Anthropic", Purpose: PurposeTraining},

	// Meta
	{Name: "Meta-ExternalAgent", UserAgent: "Meta-ExternalAgent", Operator: "Meta", Purpose: PurposeTraining},
	{Name: "FacebookBot", UserAgent: "FacebookBot", Operator: "Meta", Purpose: PurposeSearch},

	// Apple
	{Name: "Applebot", UserAgent: "Applebot", Operator: "Apple", Purpose: PurposeSearch},
	{Name: "Applebot-Extended", UserAgent: "Applebot-Extended", Operator: "Apple", Purpose: PurposeTraining},

	// Perplexity
	{Name: "PerplexityBot", UserAgent: "PerplexityBot", Operator: "Perplexity", Purpose: PurposeSearch},

	// Common Crawl
	{Name: "CCBot", UserAgent: "CCBot", Operator: "Common Crawl", Purpose: PurposeTraining},
}

// Evaluate checks robot access for all known crawlers on a given path.
func Evaluate(rt *robots.RobotsTxt, path string) *Report {
	report := &Report{Path: path}
	for _, crawler := range KnownCrawlers {
		allowed := rt.IsAllowed(crawler.UserAgent, path)
		rule := ""
		if !allowed {
			rule = rt.MatchedRule(crawler.UserAgent, path)
		}
		report.Results = append(report.Results, PolicyResult{
			Crawler: crawler,
			Allowed: allowed,
			Rule:    rule,
		})
	}
	return report
}

// SummaryByPurpose returns a map of purpose → count of blocked crawlers.
func (r *Report) SummaryByPurpose() map[CrawlerPurpose]struct{ Allowed, Blocked int } {
	summary := make(map[CrawlerPurpose]struct{ Allowed, Blocked int })
	for _, result := range r.Results {
		s := summary[result.Crawler.Purpose]
		if result.Allowed {
			s.Allowed++
		} else {
			s.Blocked++
		}
		summary[result.Crawler.Purpose] = s
	}
	return summary
}
